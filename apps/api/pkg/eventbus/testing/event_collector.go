package testing

import (
	"ddd-second-hand-marketplace/pkg/eventbus"
)

// EventCollector collects events published to the event bus
// Simplified for use with sync event bus - no need for WaitGroup or mutex
type EventCollector struct {
	events []eventbus.DomainEvent
}

func NewEventCollector() *EventCollector {
	return &EventCollector{
		events: make([]eventbus.DomainEvent, 0),
	}
}

// EventHandler returns a handler that collects events
func (c *EventCollector) EventHandler() eventbus.EventHandler {
	return func(evt eventbus.DomainEvent) error {
		c.events = append(c.events, evt)
		return nil
	}
}

// GetEvents returns all collected events
func (c *EventCollector) GetEvents() []eventbus.DomainEvent {
	return c.events
}

// Clear removes all collected events
func (c *EventCollector) Clear() {
	c.events = make([]eventbus.DomainEvent, 0)
}
