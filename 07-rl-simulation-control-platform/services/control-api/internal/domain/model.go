package domain

import (
	"errors"
	"fmt"
	"time"
)

type RunStatus string

const (
	StatusDraft      RunStatus = "draft"
	StatusQueued     RunStatus = "queued"
	StatusRunning    RunStatus = "running"
	StatusPausing    RunStatus = "pausing"
	StatusPaused     RunStatus = "paused"
	StatusCancelling RunStatus = "cancelling"
	StatusCancelled  RunStatus = "cancelled"
	StatusSucceeded  RunStatus = "succeeded"
	StatusFailed     RunStatus = "failed"
)

var (
	ErrNotFound        = errors.New("resource not found")
	ErrInvalidState    = errors.New("invalid run state transition")
	ErrInvalidArgument = errors.New("invalid argument")
	ErrConflict        = errors.New("resource conflict")
)

type Coordinate struct {
	Row    int `json:"row"`
	Column int `json:"column"`
}

type RewardMap struct {
	Step      float64 `json:"step"`
	Collision float64 `json:"collision"`
	Goal      float64 `json:"goal"`
}

type Environment struct {
	ID               string       `json:"id"`
	Slug             string       `json:"slug"`
	Name             string       `json:"name"`
	Version          string       `json:"version"`
	Rows             int          `json:"rows"`
	Columns          int          `json:"columns"`
	Start            Coordinate   `json:"start"`
	Goal             Coordinate   `json:"goal"`
	Obstacles        []Coordinate `json:"obstacles"`
	RewardMap        RewardMap    `json:"rewardMap"`
	ObservationSpace string       `json:"observationSpace"`
	ActionSpace      []string     `json:"actionSpace"`
}

type Policy struct {
	ID               string            `json:"id"`
	Algorithm        string            `json:"algorithm"`
	Version          string            `json:"version"`
	ArtifactURI      string            `json:"artifactUri"`
	SHA256           string            `json:"sha256"`
	Provenance       string            `json:"provenance"`
	ObservationSpace string            `json:"observationSpace"`
	ActionSpace      []string          `json:"actionSpace"`
	ActionByState    map[string]string `json:"actionByState"`
	Enabled          bool              `json:"enabled"`
}

type TrainingRun struct {
	ID              string     `json:"id"`
	EnvironmentID   string     `json:"environmentId"`
	PolicyID        string     `json:"policyId"`
	Status          RunStatus  `json:"status"`
	Attempt         int        `json:"attempt"`
	Seed            int        `json:"seed"`
	MaxSteps        int        `json:"maxSteps"`
	TerminalReason  *string    `json:"terminalReason"`
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`
	StartedAt       *time.Time `json:"startedAt,omitempty"`
	CompletedAt     *time.Time `json:"completedAt,omitempty"`
	CorrelationID   string     `json:"-"`
	IdempotencyKey  string     `json:"-"`
	EnvironmentVers string     `json:"-"`
	PolicyAlgorithm string     `json:"-"`
	PolicySHA256    string     `json:"-"`
}

func (run TrainingRun) CanRequest(action string) (RunStatus, error) {
	switch action {
	case "pause":
		if run.Status == StatusRunning {
			return StatusPausing, nil
		}
	case "resume":
		if run.Status == StatusPaused {
			return StatusQueued, nil
		}
	case "cancel":
		if run.Status == StatusQueued || run.Status == StatusRunning || run.Status == StatusPaused || run.Status == StatusPausing {
			return StatusCancelling, nil
		}
	}
	return "", fmt.Errorf("%w: cannot %s from %s", ErrInvalidState, action, run.Status)
}

type Episode struct {
	ID             string    `json:"id"`
	RunID          string    `json:"runId"`
	EpisodeNumber  int       `json:"episodeNumber"`
	Status         string    `json:"status"`
	TotalReward    float64   `json:"totalReward"`
	StepCount      int       `json:"stepCount"`
	Collisions     int       `json:"collisions"`
	TerminalReason string    `json:"terminalReason"`
	StartedAt      time.Time `json:"startedAt"`
	CompletedAt    time.Time `json:"completedAt"`
}

type Transition struct {
	ID                 string      `json:"id"`
	EpisodeID          string      `json:"episodeId"`
	StepIndex          int         `json:"stepIndex"`
	State              Coordinate  `json:"state"`
	Action             string      `json:"action"`
	NextState          Coordinate  `json:"nextState"`
	Reward             float64     `json:"reward"`
	Terminated         bool        `json:"terminated"`
	Truncated          bool        `json:"truncated"`
	SampledAt          time.Time   `json:"sampledAt"`
	PredictedState     *Coordinate `json:"predictedState,omitempty"`
	PredictedNextState *Coordinate `json:"predictedNextState,omitempty"`
	StepError          *float64    `json:"stepError,omitempty"`
	AccumulatedError   *float64    `json:"accumulatedError,omitempty"`
	ModelVersion       *string     `json:"modelVersion,omitempty"`
}

type MetricSample struct {
	RunID     string    `json:"runId"`
	Metric    string    `json:"metric"`
	Value     float64   `json:"value"`
	Unit      string    `json:"unit"`
	Step      int       `json:"step"`
	SampledAt time.Time `json:"sampledAt"`
}

type RunQuery struct {
	Limit  int
	Status RunStatus
	Before *time.Time
}

type MetricQuery struct {
	Limit  int
	Metric string
	After  *time.Time
}

type EpisodeQuery struct {
	Limit        int
	AfterEpisode int
}

type TransitionQuery struct {
	Limit     int
	AfterStep int
}

type CreateRunCommand struct {
	EnvironmentID string `json:"environmentId" binding:"required,uuid"`
	PolicyID      string `json:"policyId" binding:"required,uuid"`
	Seed          int    `json:"seed" binding:"min=0,max=2147483647"`
	MaxSteps      int    `json:"maxSteps" binding:"required,min=1,max=200"`
}

type FeedbackCommand struct {
	Category string `json:"category" binding:"required,oneof=clear unexpected loop collision other"`
	Note     string `json:"note" binding:"required,min=1,max=500"`
}

type MessageEnvelope struct {
	EventID       string         `json:"event_id"`
	MessageType   string         `json:"message_type"`
	SchemaVersion string         `json:"schema_version"`
	OccurredAt    time.Time      `json:"occurred_at"`
	CorrelationID string         `json:"correlation_id"`
	CausationID   string         `json:"causation_id"`
	RunID         string         `json:"run_id"`
	Attempt       int            `json:"attempt"`
	Payload       map[string]any `json:"payload"`
}

func (message MessageEnvelope) Validate() error {
	if message.SchemaVersion != "1.0" || message.EventID == "" || message.RunID == "" || message.CorrelationID == "" {
		return fmt.Errorf("%w: invalid message envelope", ErrInvalidArgument)
	}
	if message.Attempt < 1 || message.Attempt > 5 || len(message.Payload) > 20 {
		return fmt.Errorf("%w: retry or payload boundary exceeded", ErrInvalidArgument)
	}
	return nil
}
