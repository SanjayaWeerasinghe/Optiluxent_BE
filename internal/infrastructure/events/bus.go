package events

import "context"

// EventBus defines the interface for publishing and subscribing to events.
type EventBus interface {
	// Publish sends an event to the specified stream.
	Publish(ctx context.Context, stream string, event Event) error

	// Subscribe registers a handler for events on stream within the consumer group.
	Subscribe(stream, group, consumer string, handler EventHandler)

	// Start begins consuming events from all subscribed streams.
	Start(ctx context.Context) error

	// Stop gracefully shuts down the consumer.
	Stop()
}
