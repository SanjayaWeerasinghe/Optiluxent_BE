package events

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"erp-system/pkg/logger"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	maxRetries     = 3
	pollInterval   = 100 * time.Millisecond
	claimMinIdle   = 30 * time.Second
)

type subscription struct {
	stream   string
	group    string
	consumer string
	handler  EventHandler
}

// RedisStreamBus implements EventBus using Redis Streams.
type RedisStreamBus struct {
	client *redis.Client
	subs   []subscription
	mu     sync.RWMutex
	stopCh chan struct{}
	wg     sync.WaitGroup
}

func NewRedisStreamBus(client *redis.Client) *RedisStreamBus {
	return &RedisStreamBus{
		client: client,
		stopCh: make(chan struct{}),
	}
}

func (b *RedisStreamBus) Publish(ctx context.Context, stream string, event Event) error {
	if event.ID == "" {
		event.ID = uuid.New().String()
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now()
	}

	data, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("events: marshal event: %w", err)
	}

	return b.client.XAdd(ctx, &redis.XAddArgs{
		Stream: stream,
		Values: map[string]interface{}{"data": string(data)},
	}).Err()
}

func (b *RedisStreamBus) Subscribe(stream, group, consumer string, handler EventHandler) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Ensure the consumer group exists.
	_ = b.client.XGroupCreateMkStream(context.Background(), stream, group, "0").Err()

	b.subs = append(b.subs, subscription{
		stream:   stream,
		group:    group,
		consumer: consumer,
		handler:  handler,
	})
}

func (b *RedisStreamBus) Start(ctx context.Context) error {
	b.mu.RLock()
	subs := make([]subscription, len(b.subs))
	copy(subs, b.subs)
	b.mu.RUnlock()

	for _, sub := range subs {
		s := sub
		b.wg.Add(1)
		go b.consume(ctx, s)
	}
	return nil
}

func (b *RedisStreamBus) Stop() {
	close(b.stopCh)
	b.wg.Wait()
}

func (b *RedisStreamBus) consume(ctx context.Context, sub subscription) {
	defer b.wg.Done()

	for {
		select {
		case <-b.stopCh:
			return
		default:
		}

		msgs, err := b.client.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    sub.group,
			Consumer: sub.consumer,
			Streams:  []string{sub.stream, ">"},
			Count:    10,
			Block:    pollInterval,
		}).Result()

		if err != nil && err != redis.Nil {
			logger.Error("events: read group error",
				logger.String("stream", sub.stream),
				logger.Err(err),
			)
			time.Sleep(time.Second)
			continue
		}

		for _, msg := range flattenMessages(msgs) {
			b.processMessage(ctx, sub, msg)
		}
	}
}

func (b *RedisStreamBus) processMessage(ctx context.Context, sub subscription, msg redis.XMessage) {
	dataStr, ok := msg.Values["data"].(string)
	if !ok {
		b.ack(ctx, sub, msg.ID)
		return
	}

	var event Event
	if err := json.Unmarshal([]byte(dataStr), &event); err != nil {
		logger.Error("events: unmarshal error", logger.Err(err))
		b.ack(ctx, sub, msg.ID)
		return
	}

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if lastErr = sub.handler(event); lastErr == nil {
			b.ack(ctx, sub, msg.ID)
			return
		}
		logger.Warn("events: handler error, retrying",
			logger.String("event_id", event.ID),
			logger.Int("attempt", attempt+1),
			logger.Err(lastErr),
		)
	}

	// Move to DLQ after max retries.
	logger.Error("events: max retries exceeded, sending to DLQ",
		logger.String("event_id", event.ID),
		logger.String("stream", sub.stream),
	)
	_ = b.Publish(ctx, StreamDLQ, event)
	b.ack(ctx, sub, msg.ID)
}

func (b *RedisStreamBus) ack(ctx context.Context, sub subscription, msgID string) {
	if err := b.client.XAck(ctx, sub.stream, sub.group, msgID).Err(); err != nil {
		logger.Error("events: ack error",
			logger.String("msg_id", msgID),
			logger.Err(err),
		)
	}
}

func flattenMessages(streams []redis.XStream) []redis.XMessage {
	var msgs []redis.XMessage
	for _, s := range streams {
		msgs = append(msgs, s.Messages...)
	}
	return msgs
}
