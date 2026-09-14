SELECT CASE
  WHEN EXISTS (SELECT 1 FROM schema_migrations WHERE version = '0003') THEN 'false'
  ELSE 'true'
END AS apply_sprint_03 \gset

\if :apply_sprint_03

BEGIN;

ALTER TABLE agent_policies DROP CONSTRAINT IF EXISTS agent_policies_algorithm_check;
ALTER TABLE agent_policies
  ADD CONSTRAINT agent_policies_algorithm_check
  CHECK (algorithm IN ('q-learning', 'sarsa', 'dqn', 'world-model'));

ALTER TABLE metric_samples DROP CONSTRAINT IF EXISTS metric_samples_metric_check;
ALTER TABLE metric_samples
  ADD CONSTRAINT metric_samples_metric_check
  CHECK (metric IN (
    'episode_reward', 'moving_average_reward', 'epsilon', 'loss', 'episode_steps',
    'collisions', 'success_rate', 'action_up', 'action_right', 'action_down', 'action_left',
    'training_examples', 'prediction_error', 'accumulated_error', 'rollout_risk'
  ));

ALTER TABLE step_transitions
  ADD COLUMN predicted_state_row SMALLINT,
  ADD COLUMN predicted_state_column SMALLINT,
  ADD COLUMN predicted_next_state_row SMALLINT,
  ADD COLUMN predicted_next_state_column SMALLINT,
  ADD COLUMN step_error DOUBLE PRECISION,
  ADD COLUMN accumulated_error DOUBLE PRECISION,
  ADD COLUMN model_version TEXT;

ALTER TABLE step_transitions
  ADD CONSTRAINT step_transitions_world_model_fields_check CHECK (
    (
      model_version IS NULL AND predicted_state_row IS NULL AND predicted_state_column IS NULL
      AND predicted_next_state_row IS NULL AND predicted_next_state_column IS NULL
      AND step_error IS NULL AND accumulated_error IS NULL
    ) OR (
      model_version IS NOT NULL
      AND predicted_state_row BETWEEN 0 AND 5 AND predicted_state_column BETWEEN 0 AND 5
      AND predicted_next_state_row BETWEEN 0 AND 5 AND predicted_next_state_column BETWEEN 0 AND 5
      AND step_error >= 0 AND accumulated_error >= step_error
    )
  );

INSERT INTO agent_policies (
  id, environment_version_id, algorithm, version, artifact_uri, artifact_sha256,
  observation_space, action_space, action_by_state, provenance, enabled
)
VALUES (
  '22222222-2222-4222-8222-222222222204',
  '11111111-1111-4111-8111-111111111101',
  'world-model', '1.0.0', 'artifact://policies/world-model-gridworld-v1.json',
  '7b962a91ea1b9a833e1935413a9df2795b40c0533505a63d8487d16a193be8b2',
  'Discrete(36)', '["up","right","down","left"]'::jsonb, '{}'::jsonb,
  'Repository-authored empirical action-delta model fitted from 12 versioned Gridworld transitions.',
  TRUE
)
ON CONFLICT DO NOTHING;

INSERT INTO schema_migrations (version, description)
VALUES ('0003', 'Sprint 3 World Model Rollout Viewer');

COMMIT;

\else

\echo 'Migration 0003 already applied; no database changes required.'

\endif
