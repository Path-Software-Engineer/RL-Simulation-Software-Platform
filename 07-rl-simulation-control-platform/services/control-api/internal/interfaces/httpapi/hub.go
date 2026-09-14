package httpapi

import (
	"encoding/json"
	"sync"
	"time"
)

type Hub struct {
	mu          sync.RWMutex
	subscribers map[string]map[chan []byte]struct{}
}

func NewHub() *Hub {
	return &Hub{subscribers: make(map[string]map[chan []byte]struct{})}
}

func (hub *Hub) Subscribe(runID string) (<-chan []byte, func()) {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	channel := make(chan []byte, 8)
	if hub.subscribers[runID] == nil {
		hub.subscribers[runID] = make(map[chan []byte]struct{})
	}
	hub.subscribers[runID][channel] = struct{}{}
	return channel, func() {
		hub.mu.Lock()
		defer hub.mu.Unlock()
		delete(hub.subscribers[runID], channel)
		close(channel)
		if len(hub.subscribers[runID]) == 0 {
			delete(hub.subscribers, runID)
		}
	}
}

func (hub *Hub) NotifyRun(runID string, occurredAt time.Time) {
	payload, _ := json.Marshal(map[string]any{
		"type":       "run.updated",
		"runId":      runID,
		"occurredAt": occurredAt.UTC().Format(time.RFC3339Nano),
		"resync":     "/api/v1/training-runs/" + runID,
	})
	hub.mu.RLock()
	defer hub.mu.RUnlock()
	for subscriber := range hub.subscribers[runID] {
		select {
		case subscriber <- payload:
		default:
		}
	}
}
