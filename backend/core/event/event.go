package event

import "context"

// Event represents a domain or application event that can be published to the local bus.
type Event interface {
	Topic() string
}

// Handler processes a published event.
type Handler func(context.Context, Event) error

// Bus publishes events and registers handlers for topics.
type Bus interface {
	Subscribe(topic string, handler Handler) error
	Publish(context.Context, Event) error
}
