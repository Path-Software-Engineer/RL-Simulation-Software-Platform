SELECT CASE
  WHEN EXISTS (SELECT 1 FROM schema_migrations WHERE version = '0002') THEN 'false'
  ELSE 'true'
END AS apply_sprint_02 \gset

\if :apply_sprint_02

BEGIN;

ALTER TABLE agent_policies DROP CONSTRAINT IF EXISTS agent_policies_algorithm_check;
ALTER TABLE agent_policies
  ADD CONSTRAINT agent_policies_algorithm_check
  CHECK (algorithm IN ('q-learning', 'sarsa', 'dqn'));

ALTER TABLE metric_samples DROP CONSTRAINT IF EXISTS metric_samples_metric_check;
ALTER TABLE metric_samples
  ADD CONSTRAINT metric_samples_metric_check
  CHECK (metric IN (
    'episode_reward', 'moving_average_reward', 'epsilon', 'loss', 'episode_steps',
    'collisions', 'success_rate', 'action_up', 'action_right', 'action_down', 'action_left'
  ));

UPDATE environment_versions
SET manifest_sha256 = '1b818490c0e1ab1b17753f0d089d547686035efb4c52f4e6bbbf0dfac23cbc43'
WHERE id = '11111111-1111-4111-8111-111111111101';

UPDATE agent_policies
SET artifact_sha256 = CASE id
  WHEN '22222222-2222-4222-8222-222222222201'::uuid
    THEN '22d5faf9a94fcd05fdf31d2a1429a1b8f02d4f394f6cb61b95cc26351060927f'
  WHEN '22222222-2222-4222-8222-222222222202'::uuid
    THEN 'b7043771657ec37f235103485a7375a395431db0b2dc0e39d2070ca4e87aadcb'
  ELSE artifact_sha256
END
WHERE id IN (
  '22222222-2222-4222-8222-222222222201',
  '22222222-2222-4222-8222-222222222202'
);

INSERT INTO agent_policies (
  id, environment_version_id, algorithm, version, artifact_uri, artifact_sha256,
  observation_space, action_space, action_by_state, provenance, enabled
)
VALUES (
  '22222222-2222-4222-8222-222222222203',
  '11111111-1111-4111-8111-111111111101',
  'dqn', '1.0.0', 'artifact://policies/dqn-gridworld-training-v1.json',
  '99902302d7119e631c23d982eba7a06b0135718a81780caa55e68d0732dccac7',
  'Discrete(36)', '["up","right","down","left"]'::jsonb, '{}'::jsonb,
  'Repository-authored bounded DQN training profile with deterministic seeded initialization.',
  TRUE
)
ON CONFLICT DO NOTHING;

INSERT INTO schema_migrations (version, description)
VALUES ('0002', 'Sprint 2 DQN Training Dashboard');

COMMIT;

\else

\echo 'Migration 0002 already applied; no database changes required.'

\endif
