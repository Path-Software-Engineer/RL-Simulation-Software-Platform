package redisstream

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/Path-Software-Engineer/rl-simulation-control-platform/services/control-api/internal/application"
	"github.com/Path-Software-Engineer/rl-simulation-control-platform/services/control-api/internal/domain"
	"github.com/redis/go-redis/v9"
)

type Client struct {
	client *redis.Client
}

func Open(redisURL string) (*Client, error) {
	options, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("parse redis configuration: %w", err)
	}
	return &Client{client: redis.NewClient(options)}, nil
}

func (client *Client) Close() error { return client.client.Close() }

func (client *Client) Ping(ctx context.Context) error { return client.client.Ping(ctx).Err() }

func (client *Client) Publish(ctx context.Context, stream string, data []byte) error {
	return client.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		MaxLen: 10000,
		Approx: true,
		Values: map[string]any{"data": string(data)},
	}).Err()
}

func (client *Client) SetControl(ctx context.Context, runID, state string) error {
	return client.client.Set(ctx, "rl.control:"+runID, state, 15*time.Minute).Err()
}

func (client *Client) EnsureGroup(ctx context.Context, stream, group string) error {
	err := client.client.XGroupCreateMkStream(ctx, stream, group, "0").Err()
	if err != nil && !errors.Is(err, redis.Nil) && !containsBusyGroup(err.Error()) {
		return err
	}
	return nil
}

func (client *Client) ReadGroup(ctx context.Context, stream, group, consumer string) ([]redis.XMessage, error) {
	results, err := client.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group: group, Consumer: consumer, Streams: []string{stream, ">"},
		Count: 10, Block: 2 * time.Second,
	}).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	messages := make([]redis.XMessage, 0)
	for _, result := range results {
		messages = append(messages, result.Messages...)
	}
	return messages, nil
}

func (client *Client) Acknowledge(ctx context.Context, stream, group, id string) error {
	return client.client.XAck(ctx, stream, group, id).Err()
}

func (client *Client) DeadLetter(ctx context.Context, stream, sourceID, reason, payload string) error {
	if len(reason) > 500 {
		reason = reason[:500]
	}
	if len(payload) > 4000 {
		payload = payload[:4000]
	}
	return client.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream, MaxLen: 1000, Approx: true,
		Values: map[string]any{"source_id": sourceID, "error": reason, "payload": payload},
	}).Err()
}

func containsBusyGroup(message string) bool {
	for index := 0; index+9 <= len(message); index++ {
		if message[index:index+9] == "BUSYGROUP" {
			return true
		}
	}
	return false
}

type OutboxDispatcher struct {
	store     application.Store
	transport *Client
}

func NewOutboxDispatcher(store application.Store, transport *Client) *OutboxDispatcher {
	return &OutboxDispatcher{store: store, transport: transport}
}

func (dispatcher *OutboxDispatcher) Run(ctx context.Context) {
	ticker := time.NewTicker(250 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			dispatcher.flush(ctx)
		}
	}
}

func (dispatcher *OutboxDispatcher) flush(ctx context.Context) {
	messages, err := dispatcher.store.PendingOutbox(ctx, 25)
	if err != nil {
		slog.Error("outbox query failed", "error", err)
		return
	}
	for _, message := range messages {
		if err := dispatcher.transport.Publish(ctx, message.Stream, message.Payload); err != nil {
			_ = dispatcher.store.MarkOutboxFailure(ctx, message.ID, err.Error())
			continue
		}
		_ = dispatcher.store.MarkOutboxPublished(ctx, message.ID)
	}
}

type RunNotifier interface {
	NotifyRun(runID string, occurredAt time.Time)
}

type EventProjector struct {
	store      application.Store
	transport  *Client
	notifier   RunNotifier
	stream     string
	group      string
	consumer   string
	deadLetter string
}

func NewEventProjector(store application.Store, transport *Client, notifier RunNotifier, stream, group, consumer, deadLetter string) *EventProjector {
	return &EventProjector{
		store: store, transport: transport, notifier: notifier,
		stream: stream, group: group, consumer: consumer, deadLetter: deadLetter,
	}
}

func (projector *EventProjector) Run(ctx context.Context) {
	if err := projector.transport.EnsureGroup(ctx, projector.stream, projector.group); err != nil {
		slog.Error("projector group creation failed", "error", err)
		return
	}
	for ctx.Err() == nil {
		messages, err := projector.transport.ReadGroup(ctx, projector.stream, projector.group, projector.consumer)
		if err != nil {
			slog.Error("event stream read failed", "error", err)
			time.Sleep(time.Second)
			continue
		}
		for _, event := range messages {
			projector.project(ctx, event.ID, event.Values)
		}
	}
}

func (projector *EventProjector) project(ctx context.Context, id string, values map[string]any) {
	raw, ok := values["data"].(string)
	if !ok {
		_ = projector.transport.DeadLetter(ctx, projector.deadLetter, id, "missing data field", fmt.Sprint(values))
		_ = projector.transport.Acknowledge(ctx, projector.stream, projector.group, id)
		return
	}
	var envelope domain.MessageEnvelope
	if err := json.Unmarshal([]byte(raw), &envelope); err != nil {
		_ = projector.transport.DeadLetter(ctx, projector.deadLetter, id, err.Error(), raw)
		_ = projector.transport.Acknowledge(ctx, projector.stream, projector.group, id)
		return
	}
	applied, err := projector.store.ApplyEvent(ctx, projector.group, envelope)
	if err != nil {
		_ = projector.transport.DeadLetter(ctx, projector.deadLetter, id, err.Error(), raw)
		_ = projector.transport.Acknowledge(ctx, projector.stream, projector.group, id)
		return
	}
	if applied {
		projector.notifier.NotifyRun(envelope.RunID, envelope.OccurredAt)
	}
	_ = projector.transport.Acknowledge(ctx, projector.stream, projector.group, id)
}
