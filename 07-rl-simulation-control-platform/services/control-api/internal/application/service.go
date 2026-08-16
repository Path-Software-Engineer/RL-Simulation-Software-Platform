package application

import (
	"context"
	"fmt"
	"strings"

	"github.com/Path-Software-Engineer/rl-simulation-control-platform/services/control-api/internal/domain"
	"github.com/google/uuid"
)

type Store interface {
	Ping(ctx context.Context) error
	ListEnvironments(ctx context.Context) ([]domain.Environment, error)
	GetEnvironment(ctx context.Context, id string) (domain.Environment, error)
	ListPolicies(ctx context.Context) ([]domain.Policy, error)
	GetPolicy(ctx context.Context, id string) (domain.Policy, error)
	CreateRun(ctx context.Context, command domain.CreateRunCommand, idempotencyKey, actor string) (domain.TrainingRun, error)
	ListRuns(ctx context.Context, query domain.RunQuery) ([]domain.TrainingRun, error)
	GetRun(ctx context.Context, id string) (domain.TrainingRun, error)
	RequestRunControl(ctx context.Context, id, action, idempotencyKey, actor string) (domain.TrainingRun, error)
	ListRunMetrics(ctx context.Context, runID string, query domain.MetricQuery) ([]domain.MetricSample, error)
	ListRunEpisodes(ctx context.Context, runID string, query domain.EpisodeQuery) ([]domain.Episode, error)
	GetEpisode(ctx context.Context, id string) (domain.Episode, error)
	ListTransitions(ctx context.Context, episodeID string, query domain.TransitionQuery) ([]domain.Transition, error)
	CreateFeedback(ctx context.Context, episodeID, idempotencyKey, actor string, command domain.FeedbackCommand) error
	PendingOutbox(ctx context.Context, limit int) ([]OutboxMessage, error)
	MarkOutboxPublished(ctx context.Context, id string) error
	MarkOutboxFailure(ctx context.Context, id, message string) error
	ApplyEvent(ctx context.Context, consumer string, message domain.MessageEnvelope) (bool, error)
}

type CommandTransport interface {
	Ping(ctx context.Context) error
	Publish(ctx context.Context, stream string, data []byte) error
	SetControl(ctx context.Context, runID, state string) error
}

type OutboxMessage struct {
	ID      string
	Stream  string
	Payload []byte
}

type Service struct {
	store     Store
	transport CommandTransport
}

func NewService(store Store, transport CommandTransport) *Service {
	return &Service{store: store, transport: transport}
}

func (service *Service) Ready(ctx context.Context) error {
	if err := service.store.Ping(ctx); err != nil {
		return fmt.Errorf("database: %w", err)
	}
	if err := service.transport.Ping(ctx); err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	return nil
}

func (service *Service) ListEnvironments(ctx context.Context) ([]domain.Environment, error) {
	return service.store.ListEnvironments(ctx)
}

func (service *Service) GetEnvironment(ctx context.Context, id string) (domain.Environment, error) {
	if _, err := uuid.Parse(id); err != nil {
		return domain.Environment{}, fmt.Errorf("%w: environment id", domain.ErrInvalidArgument)
	}
	return service.store.GetEnvironment(ctx, id)
}

func (service *Service) ListPolicies(ctx context.Context) ([]domain.Policy, error) {
	return service.store.ListPolicies(ctx)
}

func (service *Service) GetPolicy(ctx context.Context, id string) (domain.Policy, error) {
	if _, err := uuid.Parse(id); err != nil {
		return domain.Policy{}, fmt.Errorf("%w: policy id", domain.ErrInvalidArgument)
	}
	return service.store.GetPolicy(ctx, id)
}

func (service *Service) CreateRun(ctx context.Context, command domain.CreateRunCommand, key, actor string) (domain.TrainingRun, error) {
	if err := validateKey(key); err != nil {
		return domain.TrainingRun{}, err
	}
	return service.store.CreateRun(ctx, command, key, actor)
}

func (service *Service) ListRuns(ctx context.Context, query domain.RunQuery) ([]domain.TrainingRun, error) {
	query.Limit = boundedLimit(query.Limit, 50, 200)
	if query.Status != "" && !knownRunStatus(query.Status) {
		return nil, fmt.Errorf("%w: run status filter", domain.ErrInvalidArgument)
	}
	return service.store.ListRuns(ctx, query)
}

func (service *Service) GetRun(ctx context.Context, id string) (domain.TrainingRun, error) {
	if err := validateResourceID(id, "run"); err != nil {
		return domain.TrainingRun{}, err
	}
	return service.store.GetRun(ctx, id)
}

func (service *Service) RequestControl(ctx context.Context, id, action, key, actor string) (domain.TrainingRun, error) {
	if err := validateResourceID(id, "run"); err != nil {
		return domain.TrainingRun{}, err
	}
	if err := validateKey(key); err != nil {
		return domain.TrainingRun{}, err
	}
	run, err := service.store.RequestRunControl(ctx, id, action, key, actor)
	if err != nil {
		return domain.TrainingRun{}, err
	}
	controlState := map[string]string{"pause": "pause", "resume": "run", "cancel": "cancel"}[action]
	if err := service.transport.SetControl(ctx, id, controlState); err != nil {
		return domain.TrainingRun{}, fmt.Errorf("set control signal: %w", err)
	}
	return run, nil
}

func (service *Service) ListRunMetrics(ctx context.Context, id string, query domain.MetricQuery) ([]domain.MetricSample, error) {
	if err := validateResourceID(id, "run"); err != nil {
		return nil, err
	}
	query.Limit = boundedLimit(query.Limit, 50, 200)
	if query.Metric != "" && query.Metric != "episode_reward" && query.Metric != "episode_steps" && query.Metric != "collisions" {
		return nil, fmt.Errorf("%w: metric filter", domain.ErrInvalidArgument)
	}
	return service.store.ListRunMetrics(ctx, id, query)
}

func (service *Service) ListRunEpisodes(ctx context.Context, id string, query domain.EpisodeQuery) ([]domain.Episode, error) {
	if err := validateResourceID(id, "run"); err != nil {
		return nil, err
	}
	query.Limit = boundedLimit(query.Limit, 50, 100)
	return service.store.ListRunEpisodes(ctx, id, query)
}

func (service *Service) GetEpisode(ctx context.Context, id string) (domain.Episode, error) {
	if err := validateResourceID(id, "episode"); err != nil {
		return domain.Episode{}, err
	}
	return service.store.GetEpisode(ctx, id)
}

func (service *Service) ListTransitions(ctx context.Context, id string, query domain.TransitionQuery) ([]domain.Transition, error) {
	if err := validateResourceID(id, "episode"); err != nil {
		return nil, err
	}
	query.Limit = boundedLimit(query.Limit, 100, 200)
	return service.store.ListTransitions(ctx, id, query)
}

func (service *Service) CreateFeedback(ctx context.Context, id, key, actor string, command domain.FeedbackCommand) error {
	if err := validateResourceID(id, "episode"); err != nil {
		return err
	}
	if err := validateKey(key); err != nil {
		return err
	}
	return service.store.CreateFeedback(ctx, id, key, actor, command)
}

func boundedLimit(value, fallback, maximum int) int {
	if value <= 0 {
		return fallback
	}
	if value > maximum {
		return maximum
	}
	return value
}

func validateKey(key string) error {
	if len(key) < 8 || len(key) > 100 || strings.ContainsAny(key, " \t\r\n") {
		return fmt.Errorf("%w: idempotency key", domain.ErrInvalidArgument)
	}
	return nil
}

func validateResourceID(value, resource string) error {
	if _, err := uuid.Parse(value); err != nil {
		return fmt.Errorf("%w: %s id", domain.ErrInvalidArgument, resource)
	}
	return nil
}

func knownRunStatus(status domain.RunStatus) bool {
	switch status {
	case domain.StatusDraft, domain.StatusQueued, domain.StatusRunning, domain.StatusPausing,
		domain.StatusPaused, domain.StatusCancelling, domain.StatusCancelled,
		domain.StatusSucceeded, domain.StatusFailed:
		return true
	default:
		return false
	}
}
