BEGIN;

DELETE FROM training_runs
WHERE policy_id = '22222222-2222-4222-8222-222222222204';

DELETE FROM agent_policies
WHERE id = '22222222-2222-4222-8222-222222222204';

ALTER TABLE step_transitions
  DROP CONSTRAINT IF EXISTS step_transitions_world_model_fields_check,
  DROP COLUMN IF EXISTS predicted_state_row,
  DROP COLUMN IF EXISTS predicted_state_column,
  DROP COLUMN IF EXISTS predicted_next_state_row,
  DROP COLUMN IF EXISTS predicted_next_state_column,
  DROP COLUMN IF EXISTS step_error,
  DROP COLUMN IF EXISTS accumulated_error,
  DROP COLUMN IF EXISTS model_version;

ALTER TABLE metric_samples DROP CONSTRAINT IF EXISTS metric_samples_metric_check;
ALTER TABLE metric_samples
  ADD CONSTRAINT metric_samples_metric_check
  CHECK (metric IN (
    'episode_reward', 'moving_average_reward', 'epsilon', 'loss', 'episode_steps',
    'collisions', 'success_rate', 'action_up', 'action_right', 'action_down', 'action_left'
  ));

ALTER TABLE agent_policies DROP CONSTRAINT IF EXISTS agent_policies_algorithm_check;
ALTER TABLE agent_policies
  ADD CONSTRAINT agent_policies_algorithm_check
  CHECK (algorithm IN ('q-learning', 'sarsa', 'dqn'));

DELETE FROM schema_migrations WHERE version = '0003';

COMMIT;
