CREATE TABLE IF NOT EXISTS schema_migrations (
  version TEXT PRIMARY KEY,
  description TEXT NOT NULL,
  applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

SELECT CASE
  WHEN EXISTS (SELECT 1 FROM schema_migrations WHERE version = '0001') THEN 'false'
  ELSE 'true'
END AS apply_sprint_01 \gset

\if :apply_sprint_01

BEGIN;

CREATE EXTENSION IF NOT EXISTS timescaledb;

CREATE TYPE run_status AS ENUM (
  'draft', 'queued', 'running', 'pausing', 'paused',
  'cancelling', 'cancelled', 'succeeded', 'failed'
);

CREATE TABLE environment_definitions (
  id UUID PRIMARY KEY,
  slug TEXT NOT NULL UNIQUE,
  name TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE environment_versions (
  id UUID PRIMARY KEY,
  environment_id UUID NOT NULL REFERENCES environment_definitions(id),
  version TEXT NOT NULL,
  definition JSONB NOT NULL,
  manifest_uri TEXT NOT NULL,
  manifest_sha256 CHAR(64) NOT NULL CHECK (manifest_sha256 ~ '^[a-f0-9]{64}$'),
  enabled BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (environment_id, version)
);

CREATE TABLE agent_policies (
  id UUID PRIMARY KEY,
  environment_version_id UUID NOT NULL REFERENCES environment_versions(id),
  algorithm TEXT NOT NULL CHECK (algorithm IN ('q-learning', 'sarsa')),
  version TEXT NOT NULL,
  artifact_uri TEXT NOT NULL CHECK (artifact_uri LIKE 'artifact://%'),
  artifact_sha256 CHAR(64) NOT NULL CHECK (artifact_sha256 ~ '^[a-f0-9]{64}$'),
  observation_space TEXT NOT NULL,
  action_space JSONB NOT NULL,
  action_by_state JSONB NOT NULL,
  provenance TEXT NOT NULL,
  enabled BOOLEAN NOT NULL DEFAULT FALSE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  UNIQUE (algorithm, version, environment_version_id)
);

CREATE TABLE training_runs (
  id UUID PRIMARY KEY,
  environment_version_id UUID NOT NULL REFERENCES environment_versions(id),
  policy_id UUID NOT NULL REFERENCES agent_policies(id),
  status run_status NOT NULL,
  attempt INTEGER NOT NULL DEFAULT 1 CHECK (attempt BETWEEN 1 AND 5),
  seed INTEGER NOT NULL CHECK (seed >= 0),
  max_steps INTEGER NOT NULL CHECK (max_steps BETWEEN 1 AND 200),
  idempotency_key TEXT NOT NULL UNIQUE,
  correlation_id UUID NOT NULL,
  terminal_reason TEXT,
  requested_by TEXT NOT NULL DEFAULT 'local-operator',
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  started_at TIMESTAMPTZ,
  completed_at TIMESTAMPTZ
);

CREATE TABLE episodes (
  id UUID PRIMARY KEY,
  run_id UUID NOT NULL REFERENCES training_runs(id) ON DELETE CASCADE,
  episode_number INTEGER NOT NULL CHECK (episode_number > 0),
  status TEXT NOT NULL CHECK (status IN ('succeeded', 'truncated', 'cancelled', 'failed')),
  total_reward DOUBLE PRECISION NOT NULL,
  step_count INTEGER NOT NULL CHECK (step_count BETWEEN 0 AND 200),
  collisions INTEGER NOT NULL CHECK (collisions >= 0),
  terminal_reason TEXT NOT NULL,
  started_at TIMESTAMPTZ NOT NULL,
  completed_at TIMESTAMPTZ NOT NULL,
  UNIQUE (run_id, episode_number)
);

CREATE TABLE step_transitions (
  id UUID PRIMARY KEY,
  episode_id UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
  step_index INTEGER NOT NULL CHECK (step_index >= 0),
  state_row SMALLINT NOT NULL,
  state_column SMALLINT NOT NULL,
  action TEXT NOT NULL CHECK (action IN ('up', 'right', 'down', 'left')),
  next_state_row SMALLINT NOT NULL,
  next_state_column SMALLINT NOT NULL,
  reward DOUBLE PRECISION NOT NULL,
  terminated BOOLEAN NOT NULL,
  truncated BOOLEAN NOT NULL,
  sampled_at TIMESTAMPTZ NOT NULL,
  UNIQUE (episode_id, step_index)
);

CREATE TABLE metric_samples (
  id UUID NOT NULL,
  run_id UUID NOT NULL REFERENCES training_runs(id) ON DELETE CASCADE,
  metric TEXT NOT NULL CHECK (metric IN ('episode_reward', 'episode_steps', 'collisions')),
  value DOUBLE PRECISION NOT NULL,
  unit TEXT NOT NULL,
  step INTEGER NOT NULL CHECK (step >= 0),
  sampled_at TIMESTAMPTZ NOT NULL,
  PRIMARY KEY (id, sampled_at),
  UNIQUE (run_id, metric, step, sampled_at)
);

SELECT create_hypertable('metric_samples', by_range('sampled_at'), if_not_exists => TRUE);

CREATE MATERIALIZED VIEW run_metric_hourly
WITH (timescaledb.continuous) AS
SELECT
  time_bucket(INTERVAL '1 hour', sampled_at) AS bucket,
  run_id,
  metric,
  AVG(value) AS average_value,
  MIN(value) AS minimum_value,
  MAX(value) AS maximum_value,
  COUNT(*) AS sample_count
FROM metric_samples
GROUP BY bucket, run_id, metric
WITH NO DATA;

SELECT add_continuous_aggregate_policy(
  'run_metric_hourly',
  start_offset => INTERVAL '7 days',
  end_offset => INTERVAL '10 minutes',
  schedule_interval => INTERVAL '1 hour',
  if_not_exists => TRUE
);

SELECT add_retention_policy('metric_samples', INTERVAL '30 days', if_not_exists => TRUE);

CREATE TABLE feedback_annotations (
  id UUID PRIMARY KEY,
  episode_id UUID NOT NULL REFERENCES episodes(id) ON DELETE CASCADE,
  category TEXT NOT NULL CHECK (category IN ('clear', 'unexpected', 'loop', 'collision', 'other')),
  note TEXT NOT NULL CHECK (char_length(note) BETWEEN 1 AND 500),
  actor TEXT NOT NULL,
  idempotency_key TEXT NOT NULL UNIQUE,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE audit_entries (
  id UUID PRIMARY KEY,
  aggregate_type TEXT NOT NULL,
  aggregate_id UUID NOT NULL,
  action TEXT NOT NULL,
  actor TEXT NOT NULL,
  correlation_id UUID NOT NULL,
  metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
  occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX audit_entries_aggregate_idx
  ON audit_entries (aggregate_type, aggregate_id, occurred_at DESC);

CREATE TABLE outbox_messages (
  id UUID PRIMARY KEY,
  stream TEXT NOT NULL,
  message_type TEXT NOT NULL,
  payload JSONB NOT NULL,
  correlation_id UUID NOT NULL,
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  published_at TIMESTAMPTZ,
  publish_attempts INTEGER NOT NULL DEFAULT 0 CHECK (publish_attempts BETWEEN 0 AND 10),
  last_error TEXT
);

CREATE INDEX outbox_pending_idx ON outbox_messages (created_at) WHERE published_at IS NULL;

CREATE TABLE inbox_messages (
  consumer_name TEXT NOT NULL,
  event_id UUID NOT NULL,
  message_type TEXT NOT NULL,
  processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  PRIMARY KEY (consumer_name, event_id)
);

CREATE TABLE idempotency_records (
  idempotency_key TEXT PRIMARY KEY,
  aggregate_id UUID NOT NULL,
  command_name TEXT NOT NULL,
  request_hash CHAR(64) NOT NULL CHECK (request_hash ~ '^[a-f0-9]{64}$'),
  created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

INSERT INTO environment_definitions (id, slug, name)
VALUES ('11111111-1111-4111-8111-111111111100', 'pathfinder-grid-6x6', 'Pathfinder Grid 6x6')
ON CONFLICT DO NOTHING;

INSERT INTO environment_versions (
  id, environment_id, version, definition, manifest_uri, manifest_sha256, enabled
)
VALUES (
  '11111111-1111-4111-8111-111111111101',
  '11111111-1111-4111-8111-111111111100',
  '1.0.0',
  '{"rows":6,"columns":6,"start":{"row":5,"column":0},"goal":{"row":0,"column":5},"obstacles":[{"row":4,"column":1},{"row":3,"column":1},{"row":2,"column":3},{"row":1,"column":3},{"row":1,"column":4}],"rewardMap":{"step":-0.04,"collision":-1.0,"goal":10.0},"observationSpace":"Discrete(36)","actionSpace":["up","right","down","left"]}'::jsonb,
  'artifact://manifests/gridworld-environment-v1.json',
  '1b818490c0e1ab1b17753f0d089d547686035efb4c52f4e6bbbf0dfac23cbc43',
  TRUE
)
ON CONFLICT DO NOTHING;

INSERT INTO agent_policies (
  id, environment_version_id, algorithm, version, artifact_uri, artifact_sha256,
  observation_space, action_space, action_by_state, provenance, enabled
)
VALUES
  (
    '22222222-2222-4222-8222-222222222201',
    '11111111-1111-4111-8111-111111111101',
    'q-learning', '1.0.0', 'artifact://policies/q-learning-gridworld-v1.json',
    '22d5faf9a94fcd05fdf31d2a1429a1b8f02d4f394f6cb61b95cc26351060927f',
    'Discrete(36)', '["up","right","down","left"]'::jsonb,
    '{"0,0":"right","0,1":"right","0,2":"right","0,3":"right","0,4":"right","0,5":"up","1,0":"up","1,1":"up","1,2":"up","1,3":"blocked","1,4":"blocked","1,5":"up","2,0":"up","2,1":"up","2,2":"up","2,3":"blocked","2,4":"right","2,5":"up","3,0":"up","3,1":"blocked","3,2":"up","3,3":"right","3,4":"up","3,5":"up","4,0":"up","4,1":"blocked","4,2":"up","4,3":"up","4,4":"up","4,5":"up","5,0":"up","5,1":"right","5,2":"up","5,3":"up","5,4":"up","5,5":"up"}'::jsonb,
    'Repository-authored deterministic teaching policy; not imported from a notebook.', TRUE
  ),
  (
    '22222222-2222-4222-8222-222222222202',
    '11111111-1111-4111-8111-111111111101',
    'sarsa', '1.0.0', 'artifact://policies/sarsa-gridworld-v1.json',
    'b7043771657ec37f235103485a7375a395431db0b2dc0e39d2070ca4e87aadcb',
    'Discrete(36)', '["up","right","down","left"]'::jsonb,
    '{"0,0":"right","0,1":"right","0,2":"right","0,3":"right","0,4":"right","0,5":"up","1,0":"up","1,1":"left","1,2":"up","1,3":"blocked","1,4":"blocked","1,5":"up","2,0":"up","2,1":"left","2,2":"left","2,3":"blocked","2,4":"right","2,5":"up","3,0":"up","3,1":"blocked","3,2":"down","3,3":"right","3,4":"right","3,5":"up","4,0":"down","4,1":"blocked","4,2":"down","4,3":"down","4,4":"right","4,5":"up","5,0":"right","5,1":"right","5,2":"right","5,3":"right","5,4":"right","5,5":"up"}'::jsonb,
    'Repository-authored conservative teaching policy; not imported from a notebook.', TRUE
  )
ON CONFLICT DO NOTHING;

INSERT INTO schema_migrations (version, description)
VALUES ('0001', 'Sprint 1 Gridworld Agent Visualizer');

COMMIT;

\else

\echo 'Migration 0001 already applied; no database changes required.'

\endif
