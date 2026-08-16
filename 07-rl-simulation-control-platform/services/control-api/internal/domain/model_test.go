package domain

import (
	"errors"
	"testing"
)

func TestRunControlStateMachine(t *testing.T) {
	tests := []struct {
		name   string
		status RunStatus
		action string
		want   RunStatus
	}{
		{name: "running can pause", status: StatusRunning, action: "pause", want: StatusPausing},
		{name: "paused can resume", status: StatusPaused, action: "resume", want: StatusQueued},
		{name: "queued can cancel", status: StatusQueued, action: "cancel", want: StatusCancelling},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := (TrainingRun{Status: test.status}).CanRequest(test.action)
			if err != nil || got != test.want {
				t.Fatalf("CanRequest() = %q, %v; want %q, nil", got, err, test.want)
			}
		})
	}
}

func TestTerminalRunRejectsControl(t *testing.T) {
	_, err := (TrainingRun{Status: StatusSucceeded}).CanRequest("pause")
	if !errors.Is(err, ErrInvalidState) {
		t.Fatalf("expected ErrInvalidState, got %v", err)
	}
}

func TestEnvelopeBoundaries(t *testing.T) {
	message := MessageEnvelope{SchemaVersion: "2.0", EventID: "event", RunID: "run", CorrelationID: "trace", Attempt: 1}
	if !errors.Is(message.Validate(), ErrInvalidArgument) {
		t.Fatal("unsupported message schema was accepted")
	}
}
