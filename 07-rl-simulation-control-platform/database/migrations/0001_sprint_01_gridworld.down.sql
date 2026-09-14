BEGIN;

DROP VIEW IF EXISTS run_metric_hourly;
DROP TABLE IF EXISTS idempotency_records;
DROP TABLE IF EXISTS inbox_messages;
DROP TABLE IF EXISTS outbox_messages;
DROP TABLE IF EXISTS audit_entries;
DROP TABLE IF EXISTS feedback_annotations;
DROP TABLE IF EXISTS metric_samples;
DROP FUNCTION IF EXISTS prune_expired_metric_samples();
DROP TABLE IF EXISTS step_transitions;
DROP TABLE IF EXISTS episodes;
DROP TABLE IF EXISTS training_runs;
DROP TABLE IF EXISTS agent_policies;
DROP TABLE IF EXISTS environment_versions;
DROP TABLE IF EXISTS environment_definitions;
DROP TYPE IF EXISTS run_status;
DROP TABLE IF EXISTS schema_migrations;

COMMIT;
