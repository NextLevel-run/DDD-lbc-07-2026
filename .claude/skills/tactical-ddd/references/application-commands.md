# Application Commands

## Concept

**Commands** represent write operations in CQRS (Command Query Responsibility Segregation). They:

- Modify system state
- Are named with imperative verbs (Place, Cancel, Update)
- Return minimal data (usually just ID or success/error)
- Emit domain events after successful persistence

Commands orchestrate domain objects without containing business logic themselves.

## Implementation Pattern

### File Location
```
internal/{bounded-context}/application/command/{action}.go
```

### Structure

```go
// internal/order/application/command/place_order.go

package command

import (
    "ddd-second-hand-marketplace/internal/order/domain"
    "ddd-second-hand-marketplace/pkg/eventbus"
)

// OrderItemArgs contains input data for a single order item
type OrderItemArgs struct {
    ProductID string
    Quantity  int
    UnitPrice int64  // in cents
    Currency  string
}

// PlaceOrderCommandArgs contains input data for the command
type PlaceOrderCommandArgs struct {
    CustomerEmail string
    Street        string
    City          string
    PostalCode    string
    Country       string
    Items         []OrderItemArgs
}

// PlaceOrderCommand is the command function type
type PlaceOrderCommand func(args PlaceOrderCommandArgs) (string, error)

// BuildPlaceOrderCommand builds a command with dependencies injected
func BuildPlaceOrderCommand(repo domain.OrderRepository, eventBus eventbus.Bus) PlaceOrderCommand {
    return func(args PlaceOrderCommandArgs) (string, error) {
        // 1. Create value objects
        email, err := domain.NewEmail(args.CustomerEmail)
        if err != nil {
            return "", err
        }

        address, err := domain.NewAddress(args.Street, args.City, args.PostalCode, args.Country)
        if err != nil {
            return "", err
        }

        // 2. Create order items
        items := make([]domain.OrderItem, 0, len(args.Items))
        for _, itemArg := range args.Items {
            price, err := domain.NewMoney(itemArg.UnitPrice, domain.Currency(itemArg.Currency))
            if err != nil {
                return "", err
            }
            item, err := domain.NewOrderItem(itemArg.ProductID, itemArg.Quantity, price)
            if err != nil {
                return "", err
            }
            items = append(items, item)
        }

        // 3. Create domain object
        order, err := domain.NewOrder(email, address, items)
        if err != nil {
            return "", err
        }

        // 4. Persist
        if err := repo.Save(order); err != nil {
            return "", err
        }

        // 5. Emit domain event (AFTER successful persistence)
        event := domain.NewOrderPlacedEventFrom(order)
        if err := eventBus.Publish(event); err != nil {
            return "", err
        }

        return order.ID(), nil
    }
}
```

### Command with Aggregate Modification

```go
// internal/order/application/command/cancel_order.go

package command

type CancelOrderCommandArgs struct {
    OrderId            string
    CancellationReason string
}

type CancelOrderCommand func(args CancelOrderCommandArgs) error

func BuildCancelOrderCommand(repo domain.OrderRepository, eventBus eventbus.Bus) CancelOrderCommand {
    return func(args CancelOrderCommandArgs) error {
        // 1. Retrieve existing aggregate
        order, err := repo.GetById(args.OrderId)
        if err != nil {
            return domain.ErrOrderNotFound
        }

        // 2. Execute domain behavior
        if err := order.Cancel(); err != nil {
            return err
        }

        // 3. Persist changes
        if err := repo.Save(order); err != nil {
            return err
        }

        // 4. Emit event
        event := domain.NewOrderCancelledEventFrom(order)
        return eventBus.Publish(event)
    }
}
```

### Command with External Service

```go
// internal/order/application/command/notify_customer.go

package command

import "ddd-second-hand-marketplace/pkg/mailer"

type NotifyCustomerCommandArgs struct {
    OrderId string
    Subject string
    Message string
}

type NotifyCustomerCommand func(args NotifyCustomerCommandArgs) error

func BuildNotifyCustomerCommand(
    repo domain.OrderRepository,
    mailer mailer.Mailer,
    eventBus eventbus.Bus,
) NotifyCustomerCommand {
    return func(args NotifyCustomerCommandArgs) error {
        // Get order to find customer
        order, err := repo.GetById(args.OrderId)
        if err != nil {
            return domain.ErrOrderNotFound
        }

        // Use external service
        err = mailer.Send(mailer.Email{
            To:      order.CustomerEmail().String(),
            Subject: args.Subject,
            Body:    args.Message,
        })
        if err != nil {
            return err
        }

        // Emit event
        event := domain.NewCustomerNotifiedEventFrom(
            args.OrderId,
            order.CustomerEmail().String(),
            args.Subject,
            args.Message,
        )
        return eventBus.Publish(event)
    }
}
```

## Key Principles

1. **One command per file** - `place_order.go`, `cancel_order.go`, etc.
2. **Functional approach** - command is a function, not a struct with `Execute()`
3. **Builder pattern** - `Build{Action}Command()` injects dependencies
4. **Args struct** - encapsulates all input parameters
5. **Domain validation** - create value objects with constructors that validate
6. **Emit events AFTER persistence** - ensures consistency
7. **Return minimal data** - ID for creates, nothing for updates/deletes

## Testing

### Test File Location
```
internal/{bounded-context}/application/command/{action}_test.go
```

### Test Setup Pattern

```go
// internal/order/application/command/place_order_test.go

package command

import (
    "ddd-second-hand-marketplace/internal/order/adapter/driven/inmemory"
    "ddd-second-hand-marketplace/internal/order/domain"
    "ddd-second-hand-marketplace/pkg/eventbus"
    eventbustesting "ddd-second-hand-marketplace/pkg/eventbus/testing"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// testSetup contains all dependencies needed for testing
type testSetup struct {
    repo           domain.OrderRepository
    eventBus       eventbus.Bus
    eventCollector *eventbustesting.EventCollector
    command        PlaceOrderCommand
}

// setupTest creates a fresh test setup with all dependencies
func setupTest(t *testing.T) *testSetup {
    t.Helper()

    repo := inmemory.NewInMemoryOrderRepository()
    eventBus := eventbus.NewSyncInMemoryEventBus()  // Sync for deterministic tests
    collector := eventbustesting.NewEventCollector()

    // Subscribe collector to events
    err := eventBus.Subscribe("OrderPlaced", collector.EventHandler())
    require.NoError(t, err, "Failed to subscribe to events")

    command := BuildPlaceOrderCommand(repo, eventBus)

    return &testSetup{
        repo:           repo,
        eventBus:       eventBus,
        eventCollector: collector,
        command:        command,
    }
}
```

### Assertion Helpers

```go
func assertNoOrdersInRepository(t *testing.T, repo domain.OrderRepository) {
    t.Helper()
    orders, err := repo.FindAll(domain.FindAllFilters{})
    require.NoError(t, err)
    assert.Empty(t, orders, "Expected no orders in repository")
}

func assertNoEventsEmitted(t *testing.T, collector *eventbustesting.EventCollector) {
    t.Helper()
    events := collector.GetEvents()
    assert.Empty(t, events, "Expected no events to be emitted")
}
```

### Success Test

```go
func TestPlaceOrderCommand_Success(t *testing.T) {
    // Given
    setup := setupTest(t)
    args := PlaceOrderCommandArgs{
        CustomerEmail: "customer@example.com",
        Street:        "123 Main St",
        City:          "Paris",
        PostalCode:    "75001",
        Country:       "France",
        Items: []OrderItemArgs{
            {ProductID: "PROD-123", Quantity: 2, UnitPrice: 2999, Currency: "EUR"},
        },
    }

    // When
    orderId, err := setup.command(args)

    // Then
    require.NoError(t, err, "Expected no error when placing valid order")
    assert.NotEmpty(t, orderId, "Expected order ID to be returned")

    // Verify persistence
    retrievedOrder, err := setup.repo.GetById(orderId)
    require.NoError(t, err)
    assert.Equal(t, domain.OrderPending, retrievedOrder.Status())

    // Verify event emission
    events := setup.eventCollector.GetEvents()
    require.Len(t, events, 1)
    assert.Equal(t, "OrderPlaced", events[0].EventType())
}
```

### Validation Error Tests (Table-Driven)

```go
func TestPlaceOrderCommand_ValidationErrors(t *testing.T) {
    tests := []struct {
        name          string
        args          PlaceOrderCommandArgs
        expectedError error
    }{
        {
            name: "InvalidEmail",
            args: PlaceOrderCommandArgs{
                CustomerEmail: "invalid-email",
                Street: "123 Main St", City: "Paris", PostalCode: "75001", Country: "France",
                Items: []OrderItemArgs{{ProductID: "P1", Quantity: 1, UnitPrice: 100, Currency: "EUR"}},
            },
            expectedError: domain.ErrInvalidEmail,
        },
        {
            name: "EmptyItems",
            args: PlaceOrderCommandArgs{
                CustomerEmail: "customer@example.com",
                Street: "123 Main St", City: "Paris", PostalCode: "75001", Country: "France",
                Items: []OrderItemArgs{},
            },
            expectedError: domain.ErrEmptyOrderItems,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Given
            setup := setupTest(t)

            // When
            orderId, err := setup.command(tt.args)

            // Then
            require.ErrorIs(t, err, tt.expectedError)
            assert.Empty(t, orderId)
            assertNoOrdersInRepository(t, setup.repo)
            assertNoEventsEmitted(t, setup.eventCollector)
        })
    }
}
```

### Event Payload Verification

```go
func TestPlaceOrderCommand_EmitsCorrectEvent(t *testing.T) {
    // Given
    setup := setupTest(t)
    args := PlaceOrderCommandArgs{
        CustomerEmail: "buyer@example.com",
        Street:        "456 Oak Ave",
        City:          "Lyon",
        PostalCode:    "69001",
        Country:       "France",
        Items: []OrderItemArgs{
            {ProductID: "GAMING-001", Quantity: 1, UnitPrice: 49900, Currency: "EUR"},
        },
    }

    // When
    orderId, err := setup.command(args)

    // Then
    require.NoError(t, err)

    events := setup.eventCollector.GetEvents()
    require.Len(t, events, 1)

    // Type assertion to verify payload
    event, ok := events[0].(*domain.OrderPlacedEvent)
    require.True(t, ok, "Expected event to be *OrderPlacedEvent")

    assert.Equal(t, orderId, event.Order.ID())
    assert.Equal(t, domain.OrderPending, event.Order.Status())
}
```

## Testing Principles

1. **Use sync event bus** - `NewSyncInMemoryEventBus()` for deterministic tests
2. **Verify three things**: return value, persistence, event emission
3. **Test error cases verify no side effects** - no persistence, no events
4. **Table-driven for validation scenarios**
5. **Type assertions for event payload verification**
