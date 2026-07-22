# Domain Events

## Concept

**Domain Events** are facts that happened in the domain. They capture something significant that domain experts care about. Events are named in past tense because they represent something that already occurred.

Events enable:
- **Decoupling**: Components react to events without direct dependencies
- **Audit trail**: Record of what happened and when
- **Integration**: Other bounded contexts can subscribe to events

## Implementation Pattern

### File Location
```
internal/{bounded-context}/domain/event.go
```

### Interface (from pkg/eventbus)

```go
// pkg/eventbus/eventbus.go

type DomainEvent interface {
    EventType() string
}
```

### Event Structure

```go
// internal/order/domain/event.go

package domain

import (
    "time"
    "github.com/google/uuid"
)

// OrderPlacedEvent is emitted when a new order is placed
type OrderPlacedEvent struct {
    id        string    // Event ID (private)
    eventType string    // Event type (private)
    emitedAt  time.Time // When event was created
    Order     *Order    // The aggregate (public for consumers)
}

// Full constructor with all parameters
func NewOrderPlacedEvent(emitedAt time.Time, order *Order) *OrderPlacedEvent {
    id := uuid.New().String()
    eventType := "OrderPlaced"  // Past tense, no suffix
    return &OrderPlacedEvent{
        id:        id,
        eventType: eventType,
        emitedAt:  emitedAt,
        Order:     order,
    }
}

// Convenience constructor using current time
func NewOrderPlacedEventFrom(order *Order) *OrderPlacedEvent {
    return NewOrderPlacedEvent(time.Now(), order)
}

// EventType implements eventbus.DomainEvent interface
func (e *OrderPlacedEvent) EventType() string {
    return e.eventType
}
```

### Event with Additional Data

```go
// OrderCancelledEvent includes cancellation reason
type OrderCancelledEvent struct {
    id                 string
    eventType          string
    emitedAt           time.Time
    Order              *Order
    CancellationReason *string  // Additional context
}

func NewOrderCancelledEventFrom(order *Order) *OrderCancelledEvent {
    reason := ""
    if order.CancellationReason != nil {
        reason = *order.CancellationReason
    }
    return &OrderCancelledEvent{
        id:                 uuid.New().String(),
        eventType:          "OrderCancelled",
        emitedAt:           time.Now(),
        Order:              order,
        CancellationReason: &reason,
    }
}

func (e *OrderCancelledEvent) EventType() string {
    return e.eventType
}
```

### Event Without Aggregate Reference

```go
// CustomerNotifiedEvent - references IDs instead of full entities
type CustomerNotifiedEvent struct {
    id            string
    eventType     string
    emitedAt      time.Time
    OrderId       string   // Reference by ID
    CustomerEmail string
    Subject       string
    Message       string
}

func NewCustomerNotifiedEventFrom(orderId, customerEmail, subject, message string) *CustomerNotifiedEvent {
    return &CustomerNotifiedEvent{
        id:            uuid.New().String(),
        eventType:     "CustomerNotified",
        emitedAt:      time.Now(),
        OrderId:       orderId,
        CustomerEmail: customerEmail,
        Subject:       subject,
        Message:       message,
    }
}

func (e *CustomerNotifiedEvent) EventType() string {
    return e.eventType
}
```

## Naming Conventions

| Concept | Convention | Example |
|---------|------------|---------|
| Event struct | `{Entity}{Action}Event` | `OrderPlacedEvent` |
| Event type string | Past tense, no suffix | `"OrderPlaced"` |
| Constructor | `New{Event}From{Entity}` | `NewOrderPlacedEventFrom` |

## Key Principles

1. **Past tense**: Events describe what happened, not what should happen
2. **Immutable**: Events should not be modified after creation
3. **Self-contained**: Include all data needed by consumers
4. **Private metadata**: `id`, `eventType`, `emitedAt` are implementation details
5. **Public payload**: Data needed by handlers is public

## Publishing Events

Events are published by Commands after successful persistence:

```go
// In application/command/place_order.go

// After successful save
event := domain.NewOrderPlacedEventFrom(order)
if err := eventBus.Publish(event); err != nil {
    return "", err
}
```

## Testing Events

Events are tested via command tests using `EventCollector`:

```go
import eventbustesting "ddd-second-hand-marketplace/pkg/eventbus/testing"

func TestPlaceOrderCommand_EmitsCorrectEvent(t *testing.T) {
    // Given
    eventBus := eventbus.NewSyncInMemoryEventBus()
    collector := eventbustesting.NewEventCollector()
    eventBus.Subscribe("OrderPlaced", collector.EventHandler())

    command := BuildPlaceOrderCommand(repo, eventBus)

    // When
    orderId, err := command(args)

    // Then
    require.NoError(t, err)

    events := collector.GetEvents()
    require.Len(t, events, 1)
    assert.Equal(t, "OrderPlaced", events[0].EventType())

    // Type assertion to verify payload
    event, ok := events[0].(*domain.OrderPlacedEvent)
    require.True(t, ok)
    assert.Equal(t, orderId, event.Order.ID())
}
```

See [application-commands.md](application-commands.md) for full testing patterns.
