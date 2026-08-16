BEGIN;

DELETE FROM training_runs
WHERE policy_id = '22222222-2222-4222-8222-222222222203';

DELETE FROM agent_policies
WHERE id = '22222222-2222-4222-8222-222222222203';

ALTER TABLE metric_samples DROP CONSTRAINT IF EXISTS metric_samples_metric_check;
ALTER TABLE metric_samples
  ADD CONSTRAINT metric_samples_metric_check
  CHECK (metric IN ('episode_reward', 'episode_steps', 'collisions'));

ALTER TABLE agent_policies DROP CONSTRAINT IF EXISTS agent_policies_algorithm_check;
ALTER TABLE agent_policies
  ADD CONSTRAINT agent_policies_algorithm_check
  CHECK (algorithm IN ('q-learning', 'sarsa'));

DELETE FROM schema_migrations WHERE version = '0002';

COMMIT;
