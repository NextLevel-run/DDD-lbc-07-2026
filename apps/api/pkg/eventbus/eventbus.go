package eventbus

// DomainEvent represents any domain event that can be published
// This is a generic interface that all domain events must implement
type DomainEvent interface {
	EventType() string
}

// EventHandler is a function that handles a domain event
type EventHandler func(event DomainEvent) error

// Bus defines the interface for publishing and subscribing to domain events
// This is a generic event bus that can be used across all bounded contexts
type Bus interface {
	// Publish publishes a domain event to all registered handlers
	Publish(event DomainEvent) error

	// Subscribe registers an event handler for a specific event type
	Subscribe(eventType string, handler EventHandler) error
}
