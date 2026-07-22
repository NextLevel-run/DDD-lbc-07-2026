# Domain Repositories

## Concept

**Repositories** are the bridge between the domain model and data persistence. In Hexagonal Architecture:

- **Port (Interface)**: Defined in the domain layer - what operations are needed
- **Adapter (Implementation)**: Lives in the adapter layer - how operations are performed

The domain defines WHAT it needs; adapters decide HOW to provide it.

## Port Definition (Domain Layer)

### File Location
```
internal/{bounded-context}/domain/repository.go
```

### Interface Pattern

```go
// internal/order/domain/repository.go

package domain

// FindAllFilters contains optional filters for querying
type FindAllFilters struct {
    Status      *OrderStatus  // nil = no filter
    CustomerID  *string       // nil = no customer filter
    MinTotal    *int64        // nil = no min total filter (in cents)
    MaxTotal    *int64        // nil = no max total filter (in cents)
}

// OrderRepository defines persistence operations for Order aggregate
type OrderRepository interface {
    // Save persists an order (create or update)
    Save(order *Order) error

    // GetById retrieves a single order by its unique identifier
    GetById(id string) (*Order, error)

    // FindAll retrieves orders with optional filtering
    FindAll(filters FindAllFilters) ([]*Order, error)
}
```

### Key Principles for Ports

1. **Interface in domain package** - no infrastructure dependencies
2. **Named after the aggregate** - `{Aggregate}Repository`
3. **Uses domain types only** - parameters and returns are domain objects
4. **Filter structs** - encapsulate optional query parameters
5. **Pointer for optional filters** - `nil` means "no filter"

## Adapter Implementation (Driven Adapter)

### File Location
```
internal/{bounded-context}/adapter/driven/inmemory/repository.go
```

### In-Memory Implementation

```go
// internal/order/adapter/driven/inmemory/repository.go

package inmemory

import (
    "ddd-second-hand-marketplace/internal/order/domain"
    "errors"
    "sync"
)

// InMemoryOrderRepository is a thread-safe in-memory implementation
type InMemoryOrderRepository struct {
    orders map[string]*domain.Order
    mu     sync.RWMutex
}

// NewInMemoryOrderRepository creates a new in-memory repository
func NewInMemoryOrderRepository() *InMemoryOrderRepository {
    return &InMemoryOrderRepository{
        orders: make(map[string]*domain.Order),
    }
}

// Save stores an order in memory
func (r *InMemoryOrderRepository) Save(order *domain.Order) error {
    r.mu.Lock()
    defer r.mu.Unlock()
    r.orders[order.ID()] = order
    return nil
}

// GetById retrieves an order by ID
func (r *InMemoryOrderRepository) GetById(id string) (*domain.Order, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()
    order, exists := r.orders[id]
    if !exists {
        return nil, errors.New("order not found")
    }
    return order, nil
}

// FindAll retrieves orders with optional filters
func (r *InMemoryOrderRepository) FindAll(filters domain.FindAllFilters) ([]*domain.Order, error) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    result := make([]*domain.Order, 0, len(r.orders))
    for _, order := range r.orders {
        // Apply status filter
        if filters.Status != nil && order.Status() != *filters.Status {
            continue
        }

        // Apply customer filter
        if filters.CustomerID != nil && order.CustomerEmail().String() != *filters.CustomerID {
            continue
        }

        // Apply min total filter
        total, _ := order.TotalAmount()
        if filters.MinTotal != nil && total.Amount() < *filters.MinTotal {
            continue
        }

        // Apply max total filter
        if filters.MaxTotal != nil && total.Amount() > *filters.MaxTotal {
            continue
        }

        result = append(result, order)
    }

    return result, nil
}
```

### Implementation Principles

1. **Thread-safe** - use `sync.RWMutex` for concurrent access
2. **Implements domain interface** - verified at compile time
3. **Returns domain types** - no infrastructure leakage
4. **Constructor pattern** - `New{Implementation}()`
5. **Package naming** - `inmemory`, `postgres`, `mongodb`, etc.

## Database Implementation (Example Pattern)

```go
// internal/order/adapter/driven/postgres/repository.go

package postgres

import (
    "database/sql"
    "ddd-second-hand-marketplace/internal/order/domain"
)

type PostgresOrderRepository struct {
    db *sql.DB
}

func NewPostgresOrderRepository(db *sql.DB) *PostgresOrderRepository {
    return &PostgresOrderRepository{db: db}
}

func (r *PostgresOrderRepository) Save(order *domain.Order) error {
    query := `
        INSERT INTO orders (id, version, customer_email, status, placed_at)
        VALUES ($1, $2, $3, $4, $5)
        ON CONFLICT (id) DO UPDATE SET
            version = EXCLUDED.version,
            status = EXCLUDED.status
    `
    _, err := r.db.Exec(query,
        order.ID(), order.Version(), order.CustomerEmail().String(),
        order.Status(), order.PlacedAt(),
    )
    return err
}

// ... other methods
```

## Dependency Injection

Repositories are injected into commands/queries at application startup:

```go
// cmd/api/main.go

func main() {
    // Create repository (choose implementation)
    repo := inmemory.NewInMemoryOrderRepository()
    // OR: repo := postgres.NewPostgresOrderRepository(db)

    // Inject into commands
    placeOrderCommand := command.BuildPlaceOrderCommand(repo, eventBus)

    // Inject into queries
    getOrderQuery := query.BuildGetOrderQuery(repo)
}
```

## Testing

### In Tests: Use In-Memory Implementation

```go
func setupTest(t *testing.T) *testSetup {
    t.Helper()

    repo := inmemory.NewInMemoryOrderRepository()
    command := BuildPlaceOrderCommand(repo, eventBus)

    return &testSetup{
        repo:    repo,
        command: command,
    }
}

func TestPlaceOrderCommand_Success(t *testing.T) {
    // Given
    setup := setupTest(t)

    // When
    orderId, err := setup.command(args)

    // Then - verify persistence via repository
    require.NoError(t, err)
    retrievedOrder, err := setup.repo.GetById(orderId)
    require.NoError(t, err)
    assert.Equal(t, OrderPending, retrievedOrder.Status())
}
```

### Helper for Creating Test Data

```go
func createAndSaveOrder(t *testing.T, repo domain.OrderRepository, customerEmail string, items []domain.OrderItem) *domain.Order {
    t.Helper()

    email, _ := domain.NewEmail(customerEmail)
    address, _ := domain.NewAddress("123 Main St", "Paris", "75001", "France")
    order, err := domain.NewOrder(email, address, items)
    require.NoError(t, err)

    err = repo.Save(order)
    require.NoError(t, err, "Failed to save order to repository")

    return order
}
```

### Testing No Persistence on Error

```go
func assertNoOrdersInRepository(t *testing.T, repo domain.OrderRepository) {
    t.Helper()
    orders, err := repo.FindAll(domain.FindAllFilters{})
    require.NoError(t, err)
    assert.Empty(t, orders, "Expected no orders in repository")
}

func TestPlaceOrderCommand_ValidationError_NoPersistence(t *testing.T) {
    // Given
    setup := setupTest(t)
    invalidArgs := PlaceOrderCommandArgs{Items: []OrderItemArgs{}} // Invalid - empty items

    // When
    _, err := setup.command(invalidArgs)

    // Then
    require.Error(t, err)
    assertNoOrdersInRepository(t, setup.repo) // Nothing was saved
}
```
