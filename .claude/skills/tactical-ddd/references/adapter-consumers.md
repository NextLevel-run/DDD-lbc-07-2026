# Event Consumers

## Concept

**Consumers** (or Event Handlers) are **driving adapters** that react to domain events. They:

- Subscribe to specific event types
- Execute side effects (send emails, update read models, trigger other commands)
- Are decoupled from the event producer
- Live in the adapter layer (they "drive" the application)

Consumers bridge the gap between events and reactions.

## Implementation Pattern

### File Location
```
internal/{bounded-context}/adapter/driving/consumer/{consumer_name}.go
```

### Simple Logging Consumer

```go
// internal/order/adapter/driving/consumer/logging_consumer.go

package consumer

import (
    "ddd-second-hand-marketplace/pkg/eventbus"
    "fmt"
)

// NewLoggingConsumer creates and registers a consumer that logs events
func NewLoggingConsumer(eventBus eventbus.Bus) error {
    return eventBus.Subscribe("OrderPlaced", func(evt eventbus.DomainEvent) error {
        fmt.Printf("Event received: %s\n", evt.EventType())
        return nil
    })
}
```

### Consumer with Command Execution

```go
// internal/order/adapter/driving/consumer/confirmation_email_consumer.go

package consumer

import (
    "ddd-second-hand-marketplace/internal/order/application/command"
    "ddd-second-hand-marketplace/internal/order/domain"
    "ddd-second-hand-marketplace/pkg/eventbus"
    "fmt"
)

// NewConfirmationEmailConsumer sends confirmation emails when orders are placed
func NewConfirmationEmailConsumer(
    eventBus eventbus.Bus,
    notifyCustomerCommand command.NotifyCustomerCommand,
) error {
    return eventBus.Subscribe("OrderPlaced", func(evt eventbus.DomainEvent) error {
        // Type assertion to access event payload
        orderEvent, ok := evt.(*domain.OrderPlacedEvent)
        if !ok {
            fmt.Printf("Error: expected OrderPlacedEvent, got %T\n", evt)
            return nil  // Don't fail on type mismatch
        }

        // Get data from event
        orderId := orderEvent.Order.ID()

        // Execute command
        err := notifyCustomerCommand(command.NotifyCustomerCommandArgs{
            OrderId: orderId,
            Subject: "Order Confirmation",
            Message: "Thank you for your order!",
        })
        if err != nil {
            fmt.Printf("Error sending confirmation email: %v\n", err)
            return err
        }

        return nil
    })
}
```

### Consumer Triggering Another Bounded Context

```go
// internal/order/adapter/driving/consumer/inventory_consumer.go

package consumer

import (
    "ddd-second-hand-marketplace/internal/order/domain"
    "ddd-second-hand-marketplace/pkg/eventbus"
)

// InventoryService interface (could be from another bounded context)
type InventoryService interface {
    ReserveStock(productID string, quantity int) error
}

func NewInventoryConsumer(eventBus eventbus.Bus, inventory InventoryService) error {
    return eventBus.Subscribe("OrderPlaced", func(evt eventbus.DomainEvent) error {
        orderEvent, ok := evt.(*domain.OrderPlacedEvent)
        if !ok {
            return nil
        }

        // Reserve stock for each item
        for _, item := range orderEvent.Order.Items() {
            if err := inventory.ReserveStock(item.ProductID(), item.Quantity()); err != nil {
                return err
            }
        }

        return nil
    })
}
```

## Key Principles

1. **One event type per consumer** - or group related reactions
2. **Constructor returns error** - subscription can fail
3. **Type assertion with safety** - don't fail on wrong type
4. **Log errors, don't crash** - event handlers should be resilient
5. **Keep handlers simple** - delegate to commands/services
6. **Register at startup** - in `main.go`

## Registration at Startup

```go
// cmd/api/main.go

func main() {
    // Create dependencies
    repo := inmemory.NewInMemoryOrderRepository()
    eventBus := eventbus.NewInMemoryEventBus()
    mailer := mailer.NewFakeMailer()

    // Build commands
    notifyCustomerCommand := command.BuildNotifyCustomerCommand(repo, mailer, eventBus)

    // Register consumers (driving adapters)
    if err := consumer.NewLoggingConsumer(eventBus); err != nil {
        log.Fatal(err)
    }

    if err := consumer.NewConfirmationEmailConsumer(eventBus, notifyCustomerCommand); err != nil {
        log.Fatal(err)
    }

    // Build other commands that emit events
    placeOrderCommand := command.BuildPlaceOrderCommand(repo, eventBus)

    // Start HTTP server...
}
```

## Testing Consumers

Consumers are tested by:
1. Publishing events directly to the event bus
2. Verifying the expected side effects

### Test Setup

```go
// internal/order/adapter/driving/consumer/confirmation_email_consumer_test.go

package consumer

import (
    "ddd-second-hand-marketplace/internal/order/adapter/driven/inmemory"
    "ddd-second-hand-marketplace/internal/order/application/command"
    "ddd-second-hand-marketplace/internal/order/domain"
    "ddd-second-hand-marketplace/pkg/eventbus"
    mailertesting "ddd-second-hand-marketplace/pkg/mailer/testing"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

type consumerTestSetup struct {
    eventBus   eventbus.Bus
    mailerSpy  *mailertesting.MailerSpy
    repo       domain.OrderRepository
}

func setupConsumerTest(t *testing.T) *consumerTestSetup {
    t.Helper()

    eventBus := eventbus.NewSyncInMemoryEventBus()
    mailerSpy := mailertesting.NewMailerSpy()
    repo := inmemory.NewInMemoryOrderRepository()

    // Build the command the consumer will use
    notifyCustomerCommand := command.BuildNotifyCustomerCommand(repo, mailerSpy, eventBus)

    // Register the consumer
    err := NewConfirmationEmailConsumer(eventBus, notifyCustomerCommand)
    require.NoError(t, err)

    return &consumerTestSetup{
        eventBus:  eventBus,
        mailerSpy: mailerSpy,
        repo:      repo,
    }
}
```

### Consumer Test

```go
func TestConfirmationEmailConsumer_SendsEmailOnOrderPlaced(t *testing.T) {
    // Given
    setup := setupConsumerTest(t)

    // Create and save an order (so the command can find it)
    email, _ := domain.NewEmail("customer@example.com")
    address, _ := domain.NewAddress("123 Main St", "Paris", "75001", "France")
    price, _ := domain.NewMoney(2999, domain.EUR)
    item, _ := domain.NewOrderItem("PROD-123", 1, price)
    order, _ := domain.NewOrder(email, address, []domain.OrderItem{item})
    require.NoError(t, setup.repo.Save(order))

    // Create the event
    event := domain.NewOrderPlacedEventFrom(order)

    // When - publish the event
    err := setup.eventBus.Publish(event)

    // Then
    require.NoError(t, err)

    // Verify email was sent
    emails := setup.mailerSpy.GetSentEmails()
    require.Len(t, emails, 1)
    assert.Contains(t, emails[0].Subject, "Order Confirmation")
}

func TestConfirmationEmailConsumer_HandlesWrongEventType(t *testing.T) {
    // Given
    setup := setupConsumerTest(t)

    // Create a different event type
    wrongEvent := &mockEvent{eventType: "SomeOtherEvent"}

    // When - publish wrong event type
    err := setup.eventBus.Publish(wrongEvent)

    // Then - should not fail
    require.NoError(t, err)

    // No emails sent
    emails := setup.mailerSpy.GetSentEmails()
    assert.Empty(t, emails)
}

// Mock event for testing type safety
type mockEvent struct {
    eventType string
}

func (e *mockEvent) EventType() string {
    return e.eventType
}
```

### Integration Test with Full Flow

```go
func TestConfirmationEmailConsumer_Integration(t *testing.T) {
    // Given - full setup with command that emits events
    repo := inmemory.NewInMemoryOrderRepository()
    eventBus := eventbus.NewSyncInMemoryEventBus()
    mailerSpy := mailertesting.NewMailerSpy()

    // Register consumer BEFORE creating command
    notifyCustomerCommand := command.BuildNotifyCustomerCommand(repo, mailerSpy, eventBus)
    err := NewConfirmationEmailConsumer(eventBus, notifyCustomerCommand)
    require.NoError(t, err)

    // Create the command that emits events
    placeOrderCommand := command.BuildPlaceOrderCommand(repo, eventBus)

    // When - place an order (which emits event, which triggers consumer)
    orderId, err := placeOrderCommand(command.PlaceOrderCommandArgs{
        CustomerEmail: "buyer@example.com",
        Street:        "456 Oak Ave",
        City:          "Lyon",
        PostalCode:    "69001",
        Country:       "France",
        Items: []command.OrderItemArgs{
            {ProductID: "PROD-1", Quantity: 1, UnitPrice: 4999, Currency: "EUR"},
        },
    })

    // Then
    require.NoError(t, err)
    require.NotEmpty(t, orderId)

    // Verify the consumer's side effect
    emails := mailerSpy.GetSentEmails()
    require.Len(t, emails, 1)
}
```

## Testing Principles

1. **Use sync event bus** - ensures handlers complete before assertions
2. **Test side effects** - emails sent, commands executed, etc.
3. **Test type safety** - verify wrong event types don't cause failures
4. **Integration tests** - verify full flow from command → event → consumer
5. **Use spies/fakes** - `MailerSpy`, mock services

## Consumer Patterns

### Fan-out (Multiple Consumers per Event)

```go
// Multiple consumers for same event type
eventBus.Subscribe("OrderPlaced", loggingHandler)
eventBus.Subscribe("OrderPlaced", emailHandler)
eventBus.Subscribe("OrderPlaced", inventoryHandler)
```

### Selective Handling

```go
func NewHighValueOrderConsumer(eventBus eventbus.Bus) error {
    return eventBus.Subscribe("OrderPlaced", func(evt eventbus.DomainEvent) error {
        orderEvent, ok := evt.(*domain.OrderPlacedEvent)
        if !ok {
            return nil
        }

        // Only handle high-value orders
        total, _ := orderEvent.Order.TotalAmount()
        if total.Amount() < 100000 {  // Less than 1000.00
            return nil
        }

        // Handle high-value order logic
        return nil
    })
}
```
