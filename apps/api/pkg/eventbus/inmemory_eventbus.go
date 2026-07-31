package eventbus

import (
	"fmt"
	"sync"
)

// executionMode defines how event handlers are executed
type executionMode int

const (
	asyncMode executionMode = iota
	syncMode
)

// InMemoryEventBus is a thread-safe in-memory implementation of Bus
type InMemoryEventBus struct {
	handlers map[string][]EventHandler
	mu       sync.RWMutex
	mode     executionMode
}

// NewAsyncInMemoryEventBus creates a new asynchronous in-memory event bus
// Handlers are executed in separate goroutines (fire-and-forget)
func NewAsyncInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]EventHandler),
		mode:     asyncMode,
	}
}

// NewSyncInMemoryEventBus creates a new synchronous in-memory event bus
// Handlers are executed sequentially, blocking until all complete
func NewSyncInMemoryEventBus() *InMemoryEventBus {
	return &InMemoryEventBus{
		handlers: make(map[string][]EventHandler),
		mode:     syncMode,
	}
}

// Subscribe registers an event handler for a specific event type
func (bus *InMemoryEventBus) Subscribe(eventType string, handler EventHandler) error {
	bus.mu.Lock()
	defer bus.mu.Unlock()

	bus.handlers[eventType] = append(bus.handlers[eventType], handler)
	return nil
}

// Publish publishes a domain event to all registered handlers
// Execution is async or sync depending on the bus configuration
func (bus *InMemoryEventBus) Publish(evt DomainEvent) error {
	bus.mu.RLock()
	handlers := bus.handlers[evt.EventType()]
	mode := bus.mode
	bus.mu.RUnlock()

	if mode == asyncMode {
		// Fire handlers asynchronously (goroutines)
		for _, handler := range handlers {
			go func(h EventHandler) {
				if err := h(evt); err != nil {
					// Log error and continue with other handlers
					fmt.Printf("Error handling event %s: %v\n", evt.EventType(), err)
				}
			}(handler)
		}
	} else {
		// Execute handlers synchronously
		for _, handler := range handlers {
			if err := handler(evt); err != nil {
				// Log error and continue with other handlers
				fmt.Printf("Error handling event %s: %v\n", evt.EventType(), err)
			}
		}
	}

	return nil
}
