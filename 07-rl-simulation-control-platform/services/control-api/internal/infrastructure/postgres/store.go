package postgres

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/Path-Software-Engineer/rl-simulation-control-platform/services/control-api/internal/application"
	"github.com/Path-Software-Engineer/rl-simulation-control-platform/services/control-api/internal/domain"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Store struct {
	pool *pgxpool.Pool
}

func Open(ctx context.Context, databaseURL string) (*Store, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database configuration: %w", err)
	}
	config.MaxConns = 8
	config.MinConns = 1
	config.MaxConnLifetime = 30 * time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("open database pool: %w", err)
	}
	return &Store{pool: pool}, nil
}

func (store *Store) Close() { store.pool.Close() }

func (store *Store) Ping(ctx context.Context) error { return store.pool.Ping(ctx) }

func (store *Store) ListEnvironments(ctx context.Context) ([]domain.Environment, error) {
	rows, err := store.pool.Query(ctx, `
		SELECT ev.id, ed.slug, ed.name, ev.version, ev.definition
		FROM environment_versions ev
		JOIN environment_definitions ed ON ed.id = ev.environment_id
		WHERE ev.enabled = TRUE
		ORDER BY ed.slug, ev.version`)
	if err != nil {
		return nil, fmt.Errorf("list environments: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Environment, 0)
	for rows.Next() {
		environment, scanErr := scanEnvironment(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, environment)
	}
	return items, rows.Err()
}

func (store *Store) GetEnvironment(ctx context.Context, id string) (domain.Environment, error) {
	row := store.pool.QueryRow(ctx, `
		SELECT ev.id, ed.slug, ed.name, ev.version, ev.definition
		FROM environment_versions ev
		JOIN environment_definitions ed ON ed.id = ev.environment_id
		WHERE ev.id = $1 AND ev.enabled = TRUE`, id)
	environment, err := scanEnvironment(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Environment{}, domain.ErrNotFound
	}
	return environment, err
}

func (store *Store) ListPolicies(ctx context.Context) ([]domain.Policy, error) {
	rows, err := store.pool.Query(ctx, `
		SELECT id, algorithm, version, artifact_uri, artifact_sha256,
		       provenance, observation_space, action_space, action_by_state, enabled
		FROM agent_policies WHERE enabled = TRUE ORDER BY algorithm`)
	if err != nil {
		return nil, fmt.Errorf("list policies: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Policy, 0)
	for rows.Next() {
		var item domain.Policy
		var actionSpace, actionByState []byte
		if err := rows.Scan(&item.ID, &item.Algorithm, &item.Version, &item.ArtifactURI, &item.SHA256, &item.Provenance, &item.ObservationSpace, &actionSpace, &actionByState, &item.Enabled); err != nil {
			return nil, fmt.Errorf("scan policy: %w", err)
		}
		if err := json.Unmarshal(actionSpace, &item.ActionSpace); err != nil {
			return nil, fmt.Errorf("decode action space: %w", err)
		}
		if err := json.Unmarshal(actionByState, &item.ActionByState); err != nil {
			return nil, fmt.Errorf("decode policy overlay: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (store *Store) GetPolicy(ctx context.Context, id string) (domain.Policy, error) {
	var item domain.Policy
	var actionSpace, actionByState []byte
	err := store.pool.QueryRow(ctx, `
		SELECT id, algorithm, version, artifact_uri, artifact_sha256,
		       provenance, observation_space, action_space, action_by_state, enabled
		FROM agent_policies WHERE id=$1 AND enabled=TRUE`, id).Scan(
		&item.ID, &item.Algorithm, &item.Version, &item.ArtifactURI, &item.SHA256,
		&item.Provenance, &item.ObservationSpace, &actionSpace, &actionByState, &item.Enabled,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Policy{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.Policy{}, fmt.Errorf("get policy: %w", err)
	}
	if err := json.Unmarshal(actionSpace, &item.ActionSpace); err != nil {
		return domain.Policy{}, fmt.Errorf("decode action space: %w", err)
	}
	if err := json.Unmarshal(actionByState, &item.ActionByState); err != nil {
		return domain.Policy{}, fmt.Errorf("decode policy overlay: %w", err)
	}
	return item, nil
}

func (store *Store) CreateRun(ctx context.Context, command domain.CreateRunCommand, key, actor string) (domain.TrainingRun, error) {
	tx, err := store.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return domain.TrainingRun{}, fmt.Errorf("begin create run: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	requestHash := hashJSON(command)
	var existingID, existingHash string
	err = tx.QueryRow(ctx, `SELECT aggregate_id, request_hash FROM idempotency_records WHERE idempotency_key = $1`, key).Scan(&existingID, &existingHash)
	if err == nil {
		if existingHash != requestHash {
			return domain.TrainingRun{}, domain.ErrConflict
		}
		return getRun(ctx, tx, existingID)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.TrainingRun{}, fmt.Errorf("read idempotency record: %w", err)
	}

	var environmentVersion, algorithm, policySHA string
	err = tx.QueryRow(ctx, `
		SELECT ev.version, p.algorithm, p.artifact_sha256
		FROM environment_versions ev
		JOIN agent_policies p ON p.environment_version_id = ev.id
		WHERE ev.id = $1 AND p.id = $2 AND ev.enabled = TRUE AND p.enabled = TRUE`,
		command.EnvironmentID, command.PolicyID,
	).Scan(&environmentVersion, &algorithm, &policySHA)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.TrainingRun{}, fmt.Errorf("%w: environment/policy pair", domain.ErrInvalidArgument)
	}
	if err != nil {
		return domain.TrainingRun{}, fmt.Errorf("validate run references: %w", err)
	}

	runID := uuid.NewString()
	correlationID := uuid.NewString()
	eventID := uuid.NewString()
	now := time.Now().UTC()
	run := domain.TrainingRun{
		ID: runID, EnvironmentID: command.EnvironmentID, PolicyID: command.PolicyID,
		Status: domain.StatusQueued, Attempt: 1, Seed: command.Seed, MaxSteps: command.MaxSteps,
		CreatedAt: now, UpdatedAt: now, CorrelationID: correlationID, IdempotencyKey: key,
		EnvironmentVers: environmentVersion, PolicyAlgorithm: algorithm, PolicySHA256: policySHA,
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO training_runs (
		  id, environment_version_id, policy_id, status, attempt, seed, max_steps,
		  idempotency_key, correlation_id, requested_by, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$11)`,
		run.ID, run.EnvironmentID, run.PolicyID, run.Status, run.Attempt, run.Seed, run.MaxSteps,
		key, correlationID, actor, now,
	)
	if err != nil {
		return domain.TrainingRun{}, fmt.Errorf("insert run: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO idempotency_records (idempotency_key, aggregate_id, command_name, request_hash) VALUES ($1,$2,'create-run',$3)`, key, run.ID, requestHash)
	if err != nil {
		return domain.TrainingRun{}, fmt.Errorf("insert idempotency record: %w", err)
	}

	envelope := domain.MessageEnvelope{
		EventID: eventID, MessageType: "rl.run.requested.v1", SchemaVersion: "1.0",
		OccurredAt: now, CorrelationID: correlationID, CausationID: eventID,
		RunID: run.ID, Attempt: 1,
		Payload: map[string]any{
			"environment_id":      command.EnvironmentID,
			"environment_version": environmentVersion,
			"policy_id":           command.PolicyID,
			"policy_algorithm":    algorithm,
			"policy_sha256":       policySHA,
			"seed":                command.Seed,
			"max_steps":           command.MaxSteps,
		},
	}
	if err := insertOutbox(ctx, tx, envelope, "rl.commands.v1"); err != nil {
		return domain.TrainingRun{}, err
	}
	if err := insertAudit(ctx, tx, "training-run", run.ID, "run.queued", actor, correlationID, map[string]any{"idempotency_key": key}); err != nil {
		return domain.TrainingRun{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.TrainingRun{}, fmt.Errorf("commit create run: %w", err)
	}
	return run, nil
}

func (store *Store) ListRuns(ctx context.Context, query domain.RunQuery) ([]domain.TrainingRun, error) {
	rows, err := store.pool.Query(ctx, runSelect+`
		WHERE ($2 = '' OR r.status::text = $2)
		  AND ($3::timestamptz IS NULL OR r.created_at < $3)
		ORDER BY r.created_at DESC, r.id DESC
		LIMIT $1`, query.Limit, string(query.Status), query.Before)
	if err != nil {
		return nil, fmt.Errorf("list runs: %w", err)
	}
	defer rows.Close()
	items := make([]domain.TrainingRun, 0)
	for rows.Next() {
		item, scanErr := scanRun(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (store *Store) GetRun(ctx context.Context, id string) (domain.TrainingRun, error) {
	run, err := getRun(ctx, store.pool, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.TrainingRun{}, domain.ErrNotFound
	}
	return run, err
}

func (store *Store) RequestRunControl(ctx context.Context, id, action, key, actor string) (domain.TrainingRun, error) {
	tx, err := store.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return domain.TrainingRun{}, fmt.Errorf("begin control command: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	requestHash := hashJSON(map[string]string{"run_id": id, "action": action})
	var existingID, existingHash string
	err = tx.QueryRow(ctx, `SELECT aggregate_id, request_hash FROM idempotency_records WHERE idempotency_key=$1`, key).Scan(&existingID, &existingHash)
	if err == nil {
		if existingID != id || existingHash != requestHash {
			return domain.TrainingRun{}, domain.ErrConflict
		}
		return getRun(ctx, tx, id)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.TrainingRun{}, fmt.Errorf("read idempotency record: %w", err)
	}
	run, err := getRunForUpdate(ctx, tx, id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.TrainingRun{}, domain.ErrNotFound
	}
	if err != nil {
		return domain.TrainingRun{}, err
	}
	nextStatus, err := run.CanRequest(action)
	if err != nil {
		return domain.TrainingRun{}, err
	}
	now := time.Now().UTC()
	_, err = tx.Exec(ctx, `UPDATE training_runs SET status=$2, updated_at=$3 WHERE id=$1`, id, nextStatus, now)
	if err != nil {
		return domain.TrainingRun{}, fmt.Errorf("update requested state: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO idempotency_records (idempotency_key, aggregate_id, command_name, request_hash) VALUES ($1,$2,$3,$4)`, key, id, action+"-run", requestHash)
	if err != nil {
		return domain.TrainingRun{}, fmt.Errorf("insert idempotency record: %w", err)
	}
	eventID := uuid.NewString()
	envelope := domain.MessageEnvelope{
		EventID: eventID, MessageType: "rl.run." + action + "-requested.v1", SchemaVersion: "1.0",
		OccurredAt: now, CorrelationID: run.CorrelationID, CausationID: eventID,
		RunID: run.ID, Attempt: run.Attempt, Payload: map[string]any{"requested_by": actor},
	}
	if err := insertOutbox(ctx, tx, envelope, "rl.commands.v1"); err != nil {
		return domain.TrainingRun{}, err
	}
	if err := insertAudit(ctx, tx, "training-run", id, "run."+action+"-requested", actor, run.CorrelationID, map[string]any{"idempotency_key": key}); err != nil {
		return domain.TrainingRun{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return domain.TrainingRun{}, fmt.Errorf("commit control command: %w", err)
	}
	run.Status, run.UpdatedAt = nextStatus, now
	return run, nil
}

func (store *Store) ListRunMetrics(ctx context.Context, runID string, query domain.MetricQuery) ([]domain.MetricSample, error) {
	rows, err := store.pool.Query(ctx, `
		SELECT run_id, metric, value, unit, step, sampled_at
		FROM metric_samples
		WHERE run_id=$1
		  AND ($3 = '' OR metric=$3)
		  AND ($4::timestamptz IS NULL OR sampled_at > $4)
		ORDER BY sampled_at, metric
		LIMIT $2`, runID, query.Limit, query.Metric, query.After)
	if err != nil {
		return nil, fmt.Errorf("list metrics: %w", err)
	}
	defer rows.Close()
	items := make([]domain.MetricSample, 0)
	for rows.Next() {
		var item domain.MetricSample
		if err := rows.Scan(&item.RunID, &item.Metric, &item.Value, &item.Unit, &item.Step, &item.SampledAt); err != nil {
			return nil, fmt.Errorf("scan metric: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (store *Store) ListRunEpisodes(ctx context.Context, runID string, query domain.EpisodeQuery) ([]domain.Episode, error) {
	return store.listEpisodes(
		ctx,
		`WHERE run_id=$1 AND episode_number > $2 ORDER BY episode_number LIMIT $3`,
		runID,
		query.AfterEpisode,
		query.Limit,
	)
}

func (store *Store) GetEpisode(ctx context.Context, id string) (domain.Episode, error) {
	items, err := store.listEpisodes(ctx, `WHERE id=$1`, id)
	if err != nil {
		return domain.Episode{}, err
	}
	if len(items) == 0 {
		return domain.Episode{}, domain.ErrNotFound
	}
	return items[0], nil
}

func (store *Store) listEpisodes(ctx context.Context, suffix string, args ...any) ([]domain.Episode, error) {
	rows, err := store.pool.Query(ctx, `SELECT id, run_id, episode_number, status, total_reward, step_count, collisions, terminal_reason, started_at, completed_at FROM episodes `+suffix, args...)
	if err != nil {
		return nil, fmt.Errorf("query episodes: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Episode, 0)
	for rows.Next() {
		var item domain.Episode
		if err := rows.Scan(&item.ID, &item.RunID, &item.EpisodeNumber, &item.Status, &item.TotalReward, &item.StepCount, &item.Collisions, &item.TerminalReason, &item.StartedAt, &item.CompletedAt); err != nil {
			return nil, fmt.Errorf("scan episode: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (store *Store) ListTransitions(ctx context.Context, episodeID string, query domain.TransitionQuery) ([]domain.Transition, error) {
	rows, err := store.pool.Query(ctx, `
		SELECT id, episode_id, step_index, state_row, state_column, action,
		       next_state_row, next_state_column, reward, terminated, truncated, sampled_at,
		       predicted_state_row, predicted_state_column,
		       predicted_next_state_row, predicted_next_state_column,
		       step_error, accumulated_error, model_version
		FROM step_transitions
		WHERE episode_id=$1 AND step_index > $2
		ORDER BY step_index
		LIMIT $3`, episodeID, query.AfterStep, query.Limit)
	if err != nil {
		return nil, fmt.Errorf("list transitions: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Transition, 0)
	for rows.Next() {
		var item domain.Transition
		var predictedStateRow, predictedStateColumn sql.NullInt64
		var predictedNextStateRow, predictedNextStateColumn sql.NullInt64
		var stepError, accumulatedError sql.NullFloat64
		var modelVersion sql.NullString
		if err := rows.Scan(
			&item.ID, &item.EpisodeID, &item.StepIndex, &item.State.Row, &item.State.Column,
			&item.Action, &item.NextState.Row, &item.NextState.Column, &item.Reward,
			&item.Terminated, &item.Truncated, &item.SampledAt,
			&predictedStateRow, &predictedStateColumn,
			&predictedNextStateRow, &predictedNextStateColumn,
			&stepError, &accumulatedError, &modelVersion,
		); err != nil {
			return nil, fmt.Errorf("scan transition: %w", err)
		}
		if modelVersion.Valid {
			predictedState := domain.Coordinate{
				Row: int(predictedStateRow.Int64), Column: int(predictedStateColumn.Int64),
			}
			predictedNextState := domain.Coordinate{
				Row: int(predictedNextStateRow.Int64), Column: int(predictedNextStateColumn.Int64),
			}
			item.PredictedState = &predictedState
			item.PredictedNextState = &predictedNextState
			item.StepError = &stepError.Float64
			item.AccumulatedError = &accumulatedError.Float64
			item.ModelVersion = &modelVersion.String
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (store *Store) CreateFeedback(ctx context.Context, episodeID, key, actor string, command domain.FeedbackCommand) error {
	tx, err := store.pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.Serializable})
	if err != nil {
		return fmt.Errorf("begin feedback command: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	requestHash := hashJSON(map[string]any{
		"episode_id": episodeID,
		"category":   command.Category,
		"note":       command.Note,
	})
	var existingID, existingHash string
	err = tx.QueryRow(ctx, `
		SELECT aggregate_id, request_hash
		FROM idempotency_records
		WHERE idempotency_key=$1`, key).Scan(&existingID, &existingHash)
	if err == nil {
		if existingID != episodeID || existingHash != requestHash {
			return domain.ErrConflict
		}
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("read feedback idempotency record: %w", err)
	}
	var correlationID string
	err = tx.QueryRow(ctx, `
		SELECT tr.correlation_id
		FROM episodes e
		JOIN training_runs tr ON tr.id = e.run_id
		WHERE e.id=$1`, episodeID).Scan(&correlationID)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("verify feedback episode: %w", err)
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO feedback_annotations (
		  id, episode_id, category, note, actor, idempotency_key
		) VALUES ($1,$2,$3,$4,$5,$6)`,
		uuid.NewString(), episodeID, command.Category, command.Note, actor, key,
	)
	if err != nil {
		return fmt.Errorf("insert feedback: %w", err)
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO idempotency_records (
		  idempotency_key, aggregate_id, command_name, request_hash
		) VALUES ($1,$2,'create-feedback',$3)`, key, episodeID, requestHash)
	if err != nil {
		return fmt.Errorf("insert feedback idempotency record: %w", err)
	}
	if err := insertAudit(
		ctx,
		tx,
		"episode",
		episodeID,
		"episode.feedback-created",
		actor,
		correlationID,
		map[string]any{"category": command.Category, "idempotency_key": key},
	); err != nil {
		return fmt.Errorf("insert feedback audit: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit feedback command: %w", err)
	}
	return nil
}

func (store *Store) PendingOutbox(ctx context.Context, limit int) ([]application.OutboxMessage, error) {
	rows, err := store.pool.Query(ctx, `SELECT id, stream, payload FROM outbox_messages WHERE published_at IS NULL AND publish_attempts < 10 ORDER BY created_at LIMIT $1`, limit)
	if err != nil {
		return nil, fmt.Errorf("query outbox: %w", err)
	}
	defer rows.Close()
	items := make([]application.OutboxMessage, 0)
	for rows.Next() {
		var item application.OutboxMessage
		if err := rows.Scan(&item.ID, &item.Stream, &item.Payload); err != nil {
			return nil, fmt.Errorf("scan outbox: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func (store *Store) MarkOutboxPublished(ctx context.Context, id string) error {
	_, err := store.pool.Exec(ctx, `UPDATE outbox_messages SET published_at=NOW(), publish_attempts=publish_attempts+1, last_error=NULL WHERE id=$1`, id)
	return err
}

func (store *Store) MarkOutboxFailure(ctx context.Context, id, message string) error {
	_, err := store.pool.Exec(ctx, `UPDATE outbox_messages SET publish_attempts=publish_attempts+1, last_error=$2 WHERE id=$1`, id, message)
	return err
}

func (store *Store) ApplyEvent(ctx context.Context, consumer string, message domain.MessageEnvelope) (bool, error) {
	if err := message.Validate(); err != nil {
		return false, err
	}
	tx, err := store.pool.Begin(ctx)
	if err != nil {
		return false, fmt.Errorf("begin projection: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	commandTag, err := tx.Exec(ctx, `INSERT INTO inbox_messages (consumer_name,event_id,message_type) VALUES ($1,$2,$3) ON CONFLICT DO NOTHING`, consumer, message.EventID, message.MessageType)
	if err != nil {
		return false, fmt.Errorf("insert inbox: %w", err)
	}
	if commandTag.RowsAffected() == 0 {
		return false, nil
	}
	switch message.MessageType {
	case "rl.run.started.v1":
		_, err = tx.Exec(ctx, `UPDATE training_runs SET status='running', started_at=COALESCE(started_at,$2), updated_at=$2 WHERE id=$1 AND status IN ('queued','pausing')`, message.RunID, message.OccurredAt)
	case "rl.run.episode-completed.v1":
		err = projectEpisode(ctx, tx, message)
	case "rl.run.metric-sampled.v1":
		err = projectMetric(ctx, tx, message)
	case "rl.run.paused.v1":
		_, err = tx.Exec(ctx, `UPDATE training_runs SET status='paused', updated_at=$2 WHERE id=$1 AND status='pausing'`, message.RunID, message.OccurredAt)
	case "rl.run.completed.v1":
		_, err = tx.Exec(ctx, `UPDATE training_runs SET status='succeeded', terminal_reason=$2, completed_at=$3, updated_at=$3 WHERE id=$1 AND status NOT IN ('cancelled','failed')`, message.RunID, stringValue(message.Payload, "terminal_reason"), message.OccurredAt)
	case "rl.run.cancelled.v1":
		_, err = tx.Exec(ctx, `UPDATE training_runs SET status='cancelled', terminal_reason=$2, completed_at=$3, updated_at=$3 WHERE id=$1`, message.RunID, stringValue(message.Payload, "terminal_reason"), message.OccurredAt)
	case "rl.run.failed.v1":
		_, err = tx.Exec(ctx, `UPDATE training_runs SET status='failed', terminal_reason=$2, completed_at=$3, updated_at=$3 WHERE id=$1`, message.RunID, stringValue(message.Payload, "terminal_reason"), message.OccurredAt)
	default:
		return false, fmt.Errorf("%w: event type %s", domain.ErrInvalidArgument, message.MessageType)
	}
	if err != nil {
		return false, fmt.Errorf("project event: %w", err)
	}
	if err := insertAudit(ctx, tx, "training-run", message.RunID, message.MessageType, "rl-runner", message.CorrelationID, map[string]any{"event_id": message.EventID}); err != nil {
		return false, err
	}
	if err := tx.Commit(ctx); err != nil {
		return false, fmt.Errorf("commit projection: %w", err)
	}
	return true, nil
}

type rowScanner interface{ Scan(...any) error }
type queryer interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}

const runSelect = `
	SELECT r.id, r.environment_version_id, r.policy_id, r.status, r.attempt, r.seed,
	       r.max_steps, r.terminal_reason, r.created_at, r.updated_at, r.started_at,
	       r.completed_at, r.correlation_id, r.idempotency_key, ev.version,
	       p.algorithm, p.artifact_sha256
	FROM training_runs r
	JOIN environment_versions ev ON ev.id=r.environment_version_id
	JOIN agent_policies p ON p.id=r.policy_id`

func getRun(ctx context.Context, query queryer, id string) (domain.TrainingRun, error) {
	return scanRun(query.QueryRow(ctx, runSelect+` WHERE r.id=$1`, id))
}

func getRunForUpdate(ctx context.Context, tx pgx.Tx, id string) (domain.TrainingRun, error) {
	return scanRun(tx.QueryRow(ctx, runSelect+` WHERE r.id=$1 FOR UPDATE`, id))
}

func scanRun(row rowScanner) (domain.TrainingRun, error) {
	var run domain.TrainingRun
	err := row.Scan(&run.ID, &run.EnvironmentID, &run.PolicyID, &run.Status, &run.Attempt, &run.Seed, &run.MaxSteps, &run.TerminalReason, &run.CreatedAt, &run.UpdatedAt, &run.StartedAt, &run.CompletedAt, &run.CorrelationID, &run.IdempotencyKey, &run.EnvironmentVers, &run.PolicyAlgorithm, &run.PolicySHA256)
	return run, err
}

func scanEnvironment(row rowScanner) (domain.Environment, error) {
	var environment domain.Environment
	var definition []byte
	if err := row.Scan(&environment.ID, &environment.Slug, &environment.Name, &environment.Version, &definition); err != nil {
		return domain.Environment{}, err
	}
	var values struct {
		Rows             int                 `json:"rows"`
		Columns          int                 `json:"columns"`
		Start            domain.Coordinate   `json:"start"`
		Goal             domain.Coordinate   `json:"goal"`
		Obstacles        []domain.Coordinate `json:"obstacles"`
		RewardMap        domain.RewardMap    `json:"rewardMap"`
		ObservationSpace string              `json:"observationSpace"`
		ActionSpace      []string            `json:"actionSpace"`
	}
	if err := json.Unmarshal(definition, &values); err != nil {
		return domain.Environment{}, fmt.Errorf("decode environment: %w", err)
	}
	environment.Rows, environment.Columns = values.Rows, values.Columns
	environment.Start, environment.Goal = values.Start, values.Goal
	environment.Obstacles, environment.RewardMap = values.Obstacles, values.RewardMap
	environment.ObservationSpace, environment.ActionSpace = values.ObservationSpace, values.ActionSpace
	return environment, nil
}

func insertOutbox(ctx context.Context, tx pgx.Tx, envelope domain.MessageEnvelope, stream string) error {
	payload, err := json.Marshal(envelope)
	if err != nil {
		return fmt.Errorf("encode outbox envelope: %w", err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO outbox_messages (id,stream,message_type,payload,correlation_id) VALUES ($1,$2,$3,$4,$5)`, envelope.EventID, stream, envelope.MessageType, payload, envelope.CorrelationID)
	if err != nil {
		return fmt.Errorf("insert outbox: %w", err)
	}
	return nil
}

func insertAudit(ctx context.Context, tx pgx.Tx, aggregateType, aggregateID, action, actor, correlationID string, metadata map[string]any) error {
	encoded, _ := json.Marshal(metadata)
	_, err := tx.Exec(ctx, `INSERT INTO audit_entries (id,aggregate_type,aggregate_id,action,actor,correlation_id,metadata) VALUES ($1,$2,$3,$4,$5,$6,$7)`, uuid.NewString(), aggregateType, aggregateID, action, actor, correlationID, encoded)
	return err
}

func projectEpisode(ctx context.Context, tx pgx.Tx, message domain.MessageEnvelope) error {
	encoded, _ := json.Marshal(message.Payload)
	var payload struct {
		EpisodeID      string    `json:"episode_id"`
		EpisodeNumber  int       `json:"episode_number"`
		Status         string    `json:"status"`
		TotalReward    float64   `json:"total_reward"`
		StepCount      int       `json:"step_count"`
		Collisions     int       `json:"collisions"`
		TerminalReason string    `json:"terminal_reason"`
		StartedAt      time.Time `json:"started_at"`
		CompletedAt    time.Time `json:"completed_at"`
		Transitions    []struct {
			ID                 string             `json:"id"`
			StepIndex          int                `json:"step_index"`
			State              domain.Coordinate  `json:"state"`
			Action             string             `json:"action"`
			NextState          domain.Coordinate  `json:"next_state"`
			Reward             float64            `json:"reward"`
			Terminated         bool               `json:"terminated"`
			Truncated          bool               `json:"truncated"`
			SampledAt          time.Time          `json:"sampled_at"`
			PredictedState     *domain.Coordinate `json:"predicted_state"`
			PredictedNextState *domain.Coordinate `json:"predicted_next_state"`
			StepError          *float64           `json:"step_error"`
			AccumulatedError   *float64           `json:"accumulated_error"`
			ModelVersion       *string            `json:"model_version"`
		} `json:"transitions"`
	}
	if err := json.Unmarshal(encoded, &payload); err != nil {
		return err
	}
	if payload.StepCount != len(payload.Transitions) || payload.StepCount > 200 {
		return fmt.Errorf("%w: transition count", domain.ErrInvalidArgument)
	}
	_, err := tx.Exec(ctx, `INSERT INTO episodes (id,run_id,episode_number,status,total_reward,step_count,collisions,terminal_reason,started_at,completed_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10) ON CONFLICT (run_id,episode_number) DO NOTHING`, payload.EpisodeID, message.RunID, payload.EpisodeNumber, payload.Status, payload.TotalReward, payload.StepCount, payload.Collisions, payload.TerminalReason, payload.StartedAt, payload.CompletedAt)
	if err != nil {
		return err
	}
	for _, transition := range payload.Transitions {
		worldModelFields := transition.ModelVersion != nil
		completeWorldModelFields := transition.PredictedState != nil &&
			transition.PredictedNextState != nil && transition.StepError != nil &&
			transition.AccumulatedError != nil
		if worldModelFields != completeWorldModelFields {
			return fmt.Errorf("%w: incomplete world-model transition", domain.ErrInvalidArgument)
		}
		var predictedStateRow, predictedStateColumn any
		var predictedNextStateRow, predictedNextStateColumn any
		if worldModelFields {
			predictedStateRow, predictedStateColumn = transition.PredictedState.Row, transition.PredictedState.Column
			predictedNextStateRow, predictedNextStateColumn = transition.PredictedNextState.Row, transition.PredictedNextState.Column
		}
		_, err = tx.Exec(ctx, `
			INSERT INTO step_transitions (
			  id,episode_id,step_index,state_row,state_column,action,
			  next_state_row,next_state_column,reward,terminated,truncated,sampled_at,
			  predicted_state_row,predicted_state_column,
			  predicted_next_state_row,predicted_next_state_column,
			  step_error,accumulated_error,model_version
			) VALUES (
			  $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18,$19
			) ON CONFLICT (episode_id,step_index) DO NOTHING`,
			transition.ID, payload.EpisodeID, transition.StepIndex,
			transition.State.Row, transition.State.Column, transition.Action,
			transition.NextState.Row, transition.NextState.Column, transition.Reward,
			transition.Terminated, transition.Truncated, transition.SampledAt,
			predictedStateRow, predictedStateColumn,
			predictedNextStateRow, predictedNextStateColumn,
			transition.StepError, transition.AccumulatedError, transition.ModelVersion,
		)
		if err != nil {
			return err
		}
	}
	return nil
}

func projectMetric(ctx context.Context, tx pgx.Tx, message domain.MessageEnvelope) error {
	metric := stringValue(message.Payload, "metric")
	unit := stringValue(message.Payload, "unit")
	value, ok := message.Payload["value"].(float64)
	if !ok {
		return fmt.Errorf("%w: metric value", domain.ErrInvalidArgument)
	}
	step, ok := message.Payload["step"].(float64)
	if !ok {
		return fmt.Errorf("%w: metric step", domain.ErrInvalidArgument)
	}
	_, err := tx.Exec(ctx, `INSERT INTO metric_samples (id,run_id,metric,value,unit,step,sampled_at) VALUES ($1,$2,$3,$4,$5,$6,$7) ON CONFLICT DO NOTHING`, uuid.NewString(), message.RunID, metric, value, unit, int(step), message.OccurredAt)
	return err
}

func stringValue(payload map[string]any, key string) string {
	value, _ := payload[key].(string)
	return value
}

func hashJSON(value any) string {
	encoded, _ := json.Marshal(value)
	digest := sha256.Sum256(encoded)
	return hex.EncodeToString(digest[:])
}
