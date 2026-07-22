# PKG Infrastructure

## Concept

The `pkg/` directory contains **shared infrastructure packages** that:

- Are domain-agnostic (not specific to any bounded context)
- Define interfaces AND provide implementations
- Can be used by multiple bounded contexts
- Include testing utilities

## Package Structure

```
pkg/
├── eventbus/                    # Event bus infrastructure
│   ├── eventbus.go             # Interface definitions
│   ├── inmemory_eventbus.go    # In-memory implementation
│   ├── inmemory_eventbus_test.go
│   └── testing/
│       └── event_collector.go  # Test utility
└── mailer/                      # Email infrastructure
    ├── mailer.go               # Interface definition
    ├── fake_mailer.go          # Fake implementation
    └── testing/
        └── mailer_spy.go       # Test spy
```

## Event Bus Package

### Interface Definition

```go
// pkg/eventbus/eventbus.go

package eventbus

// DomainEvent is implemented by all domain events
type DomainEvent interface {
    EventType() string
}

// EventHandler processes a domain event
type EventHandler func(event DomainEvent) error

// Bus defines publishing and subscribing to events
type Bus interface {
    // Publish sends an event to all subscribed handlers
    Publish(event DomainEvent) error

    // Subscribe registers a handler for an event type
    Subscribe(eventType string, handler EventHandler) error
}
```

### In-Memory Implementation

```go
// pkg/eventbus/inmemory_eventbus.go

package eventbus

import (
    "log"
    "sync"
)

// InMemoryEventBus is an asynchronous in-memory event bus
type InMemoryEventBus struct {
    handlers map[string][]EventHandler
    mu       sync.RWMutex
}

func NewInMemoryEventBus() *InMemoryEventBus {
    return &InMemoryEventBus{
        handlers: make(map[string][]EventHandler),
    }
}

func (b *InMemoryEventBus) Publish(event DomainEvent) error {
    b.mu.RLock()
    handlers := b.handlers[event.EventType()]
    b.mu.RUnlock()

    for _, handler := range handlers {
        // Async: fire-and-forget
        go func(h EventHandler) {
            if err := h(event); err != nil {
                log.Printf("Event handler error: %v", err)
            }
        }(handler)
    }
    return nil
}

func (b *InMemoryEventBus) Subscribe(eventType string, handler EventHandler) error {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.handlers[eventType] = append(b.handlers[eventType], handler)
    return nil
}
```

### Synchronous Implementation (for Testing)

```go
// pkg/eventbus/inmemory_eventbus.go

// SyncInMemoryEventBus is a synchronous event bus for deterministic testing
type SyncInMemoryEventBus struct {
    handlers map[string][]EventHandler
    mu       sync.RWMutex
}

func NewSyncInMemoryEventBus() *SyncInMemoryEventBus {
    return &SyncInMemoryEventBus{
        handlers: make(map[string][]EventHandler),
    }
}

func (b *SyncInMemoryEventBus) Publish(event DomainEvent) error {
    b.mu.RLock()
    handlers := b.handlers[event.EventType()]
    b.mu.RUnlock()

    // Sync: execute handlers in order, propagate errors
    for _, handler := range handlers {
        if err := handler(event); err != nil {
            return err
        }
    }
    return nil
}

func (b *SyncInMemoryEventBus) Subscribe(eventType string, handler EventHandler) error {
    b.mu.Lock()
    defer b.mu.Unlock()
    b.handlers[eventType] = append(b.handlers[eventType], handler)
    return nil
}
```

### Event Collector (Testing Utility)

```go
// pkg/eventbus/testing/event_collector.go

package testing

import "ddd-second-hand-marketplace/pkg/eventbus"

// EventCollector captures events for test assertions
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
```

### Usage in Tests

```go
import (
    "ddd-second-hand-marketplace/pkg/eventbus"
    eventbustesting "ddd-second-hand-marketplace/pkg/eventbus/testing"
)

func setupTest(t *testing.T) *testSetup {
    eventBus := eventbus.NewSyncInMemoryEventBus()
    collector := eventbustesting.NewEventCollector()

    // Subscribe collector before command execution
    eventBus.Subscribe("OrderPlaced", collector.EventHandler())

    command := BuildPlaceOrderCommand(repo, eventBus)

    return &testSetup{
        eventBus:       eventBus,
        eventCollector: collector,
        command:        command,
    }
}

func TestCommand_EmitsEvent(t *testing.T) {
    setup := setupTest(t)

    _, err := setup.command(args)
    require.NoError(t, err)

    events := setup.eventCollector.GetEvents()
    require.Len(t, events, 1)
    assert.Equal(t, "OrderPlaced", events[0].EventType())
}
```

## Mailer Package

### Interface Definition

```go
// pkg/mailer/mailer.go

package mailer

// Email represents an email to send
type Email struct {
    To      string
    Subject string
    Body    string
}

// Mailer sends emails
type Mailer interface {
    Send(email Email) error
}
```

### Fake Implementation

```go
// pkg/mailer/fake_mailer.go

package mailer

import "fmt"

// FakeMailer logs emails instead of sending them
type FakeMailer struct{}

func NewFakeMailer() *FakeMailer {
    return &FakeMailer{}
}

func (m *FakeMailer) Send(email Email) error {
    fmt.Printf("FAKE EMAIL: To=%s Subject=%s\n", email.To, email.Subject)
    return nil
}
```

### Mailer Spy (Testing Utility)

```go
// pkg/mailer/testing/mailer_spy.go

package testing

import "ddd-second-hand-marketplace/pkg/mailer"

// MailerSpy captures sent emails for test assertions
type MailerSpy struct {
    sentEmails []mailer.Email
}

func NewMailerSpy() *MailerSpy {
    return &MailerSpy{
        sentEmails: make([]mailer.Email, 0),
    }
}

func (m *MailerSpy) Send(email mailer.Email) error {
    m.sentEmails = append(m.sentEmails, email)
    return nil
}

func (m *MailerSpy) GetSentEmails() []mailer.Email {
    return m.sentEmails
}
```

### Usage in Tests

```go
import mailertesting "ddd-second-hand-marketplace/pkg/mailer/testing"

func TestNotifyCustomer_SendsEmail(t *testing.T) {
    mailerSpy := mailertesting.NewMailerSpy()
    command := BuildNotifyCustomerCommand(repo, mailerSpy, eventBus)

    err := command(args)
    require.NoError(t, err)

    emails := mailerSpy.GetSentEmails()
    require.Len(t, emails, 1)
    assert.Equal(t, "customer@example.com", emails[0].To)
    assert.Contains(t, emails[0].Subject, "Order Confirmation")
}
```

## Adding a New PKG Package

### 1. Create Package Structure

```
pkg/
└── newservice/
    ├── newservice.go         # Interface
    ├── inmemory_impl.go      # In-memory implementation
    └── testing/
        └── spy.go            # Test spy
```

### 2. Define Interface

```go
// pkg/newservice/newservice.go

package newservice

type Service interface {
    DoSomething(input string) (string, error)
}
```

### 3. Create Implementation

```go
// pkg/newservice/inmemory_impl.go

package newservice

type InMemoryService struct{}

func NewInMemoryService() *InMemoryService {
    return &InMemoryService{}
}

func (s *InMemoryService) DoSomething(input string) (string, error) {
    return "result", nil
}
```

### 4. Create Test Utility

```go
// pkg/newservice/testing/spy.go

package testing

import "ddd-second-hand-marketplace/pkg/newservice"

type ServiceSpy struct {
    calls []string
}

func NewServiceSpy() *ServiceSpy {
    return &ServiceSpy{calls: make([]string, 0)}
}

func (s *ServiceSpy) DoSomething(input string) (string, error) {
    s.calls = append(s.calls, input)
    return "spy-result", nil
}

func (s *ServiceSpy) GetCalls() []string {
    return s.calls
}
```

## Key Principles

1. **Interface + Implementation** - pkg contains both
2. **Domain-agnostic** - no business logic
3. **Testing utilities** - provide spies/collectors in `testing/` subdirectory
4. **Sync for tests** - provide synchronous implementations for deterministic testing
5. **No internal imports** - pkg cannot import from `internal/`

## When to Use PKG vs Domain

| Put in `pkg/` | Put in `domain/` |
|---------------|------------------|
| Event bus | Domain events |
| Mailer interface | Business validation |
| Logger | Entity behavior |
| HTTP client | Repository interface |
| Metrics | Value objects |

The key distinction: **pkg is infrastructure**, **domain is business logic**.
