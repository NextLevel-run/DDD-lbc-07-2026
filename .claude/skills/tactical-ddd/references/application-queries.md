# Application Queries

## Concept

**Queries** represent read operations in CQRS. They:

- Do NOT modify system state
- Return data tailored for specific use cases (presenters/views)
- Filter out sensitive data (e.g., internal IDs, implementation details)
- Can have different data shapes than domain entities

Queries translate domain entities into **view models** appropriate for the consumer.

## Implementation Pattern

### File Location
```
internal/{bounded-context}/application/query/{query_name}.go
```

### Single Item Query

```go
// internal/order/application/query/get_order.go

package query

import (
    "ddd-second-hand-marketplace/internal/order/domain"
    "time"
)

// OrderView is the presenter for a single order detail view
type OrderView struct {
    Id              string
    Status          string
    ShippingAddress AddressView
    Items           []OrderItemView
    TotalAmount     int64
    TotalCurrency   string
    PlacedAt        time.Time
}

type AddressView struct {
    Street     string
    City       string
    PostalCode string
    Country    string
}

type OrderItemView struct {
    ProductID string
    Quantity  int
    UnitPrice int64
    Currency  string
}

// GetOrderQuery is a function that retrieves a single order by ID
type GetOrderQuery func(id string) (*OrderView, error)

// BuildGetOrderQuery builds a query with dependencies injected
func BuildGetOrderQuery(repo domain.OrderRepository) GetOrderQuery {
    return func(id string) (*OrderView, error) {
        // Retrieve from repository
        order, err := repo.GetById(id)
        if err != nil {
            return nil, err
        }

        // Map domain entity to view
        items := make([]OrderItemView, len(order.Items()))
        for i, item := range order.Items() {
            items[i] = OrderItemView{
                ProductID: item.ProductID(),
                Quantity:  item.Quantity(),
                UnitPrice: item.UnitPrice().Amount(),
                Currency:  string(item.UnitPrice().Currency()),
            }
        }

        total, _ := order.TotalAmount()
        addr := order.ShippingAddress()

        view := &OrderView{
            Id:     order.ID(),
            Status: string(order.Status()),
            ShippingAddress: AddressView{
                Street:     addr.Street(),
                City:       addr.City(),
                PostalCode: addr.PostalCode(),
                Country:    addr.Country(),
            },
            Items:         items,
            TotalAmount:   total.Amount(),
            TotalCurrency: string(total.Currency()),
            PlacedAt:      order.PlacedAt(),
        }

        return view, nil
    }
}
```

### List Query with Filters

```go
// internal/order/application/query/find_orders.go

package query

import (
    "ddd-second-hand-marketplace/internal/order/domain"
    "sort"
    "time"
)

// OrderListItem is the presenter for listing view
// Note: Excludes items details and full address
type OrderListItem struct {
    Id            string
    Status        string
    City          string
    TotalAmount   int64
    TotalCurrency string
    PlacedAt      time.Time
}

// FindOrdersArgs contains optional filter arguments
type FindOrdersArgs struct {
    Status    *string  // nil = no status filter
    MinTotal  *int64   // nil = no min total filter (in cents)
    MaxTotal  *int64   // nil = no max total filter (in cents)
    SortBy    string   // "placedAt" or "total", defaults to "placedAt"
    SortOrder string   // "asc" or "desc", defaults to "desc"
}

// FindOrdersQuery retrieves a list of orders with optional filters
type FindOrdersQuery func(args FindOrdersArgs) ([]OrderListItem, error)

// BuildFindOrdersQuery builds a query with dependencies injected
func BuildFindOrdersQuery(repo domain.OrderRepository) FindOrdersQuery {
    return func(args FindOrdersArgs) ([]OrderListItem, error) {
        // Apply defaults
        sortBy := args.SortBy
        if sortBy == "" {
            sortBy = "placedAt"
        }

        sortOrder := args.SortOrder
        if sortOrder == "" {
            sortOrder = "desc"
        }

        // Build repository filters
        var status *domain.OrderStatus
        if args.Status != nil {
            s := domain.OrderStatus(*args.Status)
            status = &s
        }

        filters := domain.FindAllFilters{
            Status:   status,
            MinTotal: args.MinTotal,
            MaxTotal: args.MaxTotal,
        }

        // Retrieve from repository
        orders, err := repo.FindAll(filters)
        if err != nil {
            return nil, err
        }

        // Sort in application layer
        sort.Slice(orders, func(i, j int) bool {
            var less bool
            if sortBy == "total" {
                totalI, _ := orders[i].TotalAmount()
                totalJ, _ := orders[j].TotalAmount()
                less = totalI.Amount() < totalJ.Amount()
            } else {
                less = orders[i].PlacedAt().Before(orders[j].PlacedAt())
            }

            if sortOrder == "desc" {
                return !less
            }
            return less
        })

        // Map to view models
        listItems := make([]OrderListItem, len(orders))
        for i, order := range orders {
            total, _ := order.TotalAmount()
            listItems[i] = OrderListItem{
                Id:            order.ID(),
                Status:        string(order.Status()),
                City:          order.ShippingAddress().City(),
                TotalAmount:   total.Amount(),
                TotalCurrency: string(total.Currency()),
                PlacedAt:      order.PlacedAt(),
            }
        }

        return listItems, nil
    }
}
```

## Key Principles

1. **View models, not entities** - return purpose-specific structs
2. **Filter sensitive data** - exclude internal details
3. **Different shapes for different needs** - list view vs detail view
4. **Business rules in queries** - e.g., filter by status
5. **Sorting in application layer** - keep repository simple
6. **No events** - queries don't modify state, so no events

## View Model Design

### Detail View vs List View

```go
// Detail view - more fields
type OrderView struct {
    Id              string
    Status          string
    ShippingAddress AddressView  // Full address
    Items           []OrderItemView  // All items
    TotalAmount     int64
    TotalCurrency   string
    PlacedAt        time.Time
}

// List view - fewer fields (performance + relevance)
type OrderListItem struct {
    Id            string
    Status        string
    City          string      // Only city, not full address
    TotalAmount   int64
    TotalCurrency string
    PlacedAt      time.Time
    // Items excluded - not needed in listings
}
```

### What to Exclude

- **Sensitive data**: Internal aggregate version, customer email in public lists
- **Large data in lists**: Full item details, complete addresses
- **Internal state**: `CancellationReason` (unless needed for the view)

## Testing

### Test File Location
```
internal/{bounded-context}/application/query/{query_name}_test.go
```

### Test Setup

```go
// internal/order/application/query/get_order_test.go

package query

import (
    "ddd-second-hand-marketplace/internal/order/adapter/driven/inmemory"
    "ddd-second-hand-marketplace/internal/order/domain"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

type getOrderTestSetup struct {
    repo  domain.OrderRepository
    query GetOrderQuery
}

func setupGetOrderTest(t *testing.T) *getOrderTestSetup {
    t.Helper()

    repo := inmemory.NewInMemoryOrderRepository()
    query := BuildGetOrderQuery(repo)

    return &getOrderTestSetup{
        repo:  repo,
        query: query,
    }
}

// Helper to create test data
func createAndSaveOrder(t *testing.T, repo domain.OrderRepository, customerEmail string) *domain.Order {
    t.Helper()

    email, _ := domain.NewEmail(customerEmail)
    address, _ := domain.NewAddress("123 Main St", "Paris", "75001", "France")
    price, _ := domain.NewMoney(2999, domain.EUR)
    item, _ := domain.NewOrderItem("PROD-123", 2, price)
    order, err := domain.NewOrder(email, address, []domain.OrderItem{item})
    require.NoError(t, err)

    err = repo.Save(order)
    require.NoError(t, err, "Failed to save order")

    return order
}
```

### Success Test

```go
func TestGetOrderQuery_Success(t *testing.T) {
    // Given
    setup := setupGetOrderTest(t)
    order := createAndSaveOrder(t, setup.repo, "customer@example.com")

    // When
    view, err := setup.query(order.ID())

    // Then
    require.NoError(t, err)
    require.NotNil(t, view)

    assert.Equal(t, order.ID(), view.Id)
    assert.Equal(t, "Pending", view.Status)
    assert.Equal(t, "Paris", view.ShippingAddress.City)
    assert.Len(t, view.Items, 1)
    assert.Equal(t, int64(5998), view.TotalAmount) // 2999 * 2
}
```

### Not Found Test

```go
func TestGetOrderQuery_NotFound(t *testing.T) {
    // Given
    setup := setupGetOrderTest(t)

    // When
    view, err := setup.query("non-existent-id")

    // Then
    require.Error(t, err)
    assert.Nil(t, view)
}
```

### Cancelled Order Test

```go
func TestGetOrderQuery_CancelledOrder(t *testing.T) {
    // Given
    setup := setupGetOrderTest(t)
    order := createAndSaveOrder(t, setup.repo, "customer@example.com")

    err := order.Cancel()
    require.NoError(t, err)
    require.NoError(t, setup.repo.Save(order))

    // When
    view, err := setup.query(order.ID())

    // Then
    require.NoError(t, err)
    assert.Equal(t, "Cancelled", view.Status)
}
```

### List Query Tests

```go
func TestFindOrdersQuery_FilterByStatus(t *testing.T) {
    // Given
    setup := setupFindOrdersTest(t)
    order1 := createAndSaveOrder(t, setup.repo, "a@b.com")
    order2 := createAndSaveOrder(t, setup.repo, "c@d.com")

    // Confirm order2
    _ = order2.Confirm()
    _ = setup.repo.Save(order2)

    status := "Pending"

    // When
    results, err := setup.query(FindOrdersArgs{Status: &status})

    // Then
    require.NoError(t, err)
    assert.Len(t, results, 1)
    assert.Equal(t, order1.ID(), results[0].Id)
}

func TestFindOrdersQuery_SortByTotal(t *testing.T) {
    // Given
    setup := setupFindOrdersTest(t)

    // Create orders with different totals
    createOrderWithTotal(t, setup.repo, 10000) // 100.00
    createOrderWithTotal(t, setup.repo, 5000)  // 50.00
    createOrderWithTotal(t, setup.repo, 15000) // 150.00

    // When
    results, err := setup.query(FindOrdersArgs{SortBy: "total", SortOrder: "asc"})

    // Then
    require.NoError(t, err)
    require.Len(t, results, 3)
    assert.Equal(t, int64(5000), results[0].TotalAmount)
    assert.Equal(t, int64(10000), results[1].TotalAmount)
    assert.Equal(t, int64(15000), results[2].TotalAmount)
}
```

### Assertion Helpers for Lists

```go
func assertResultCount(t *testing.T, results []OrderListItem, expected int) {
    t.Helper()
    assert.Len(t, results, expected)
}

func assertOrderInResults(t *testing.T, results []OrderListItem, orderId string) {
    t.Helper()
    for _, item := range results {
        if item.Id == orderId {
            return
        }
    }
    t.Errorf("Order %s not found in results", orderId)
}

func assertOrderNotInResults(t *testing.T, results []OrderListItem, orderId string) {
    t.Helper()
    for _, item := range results {
        if item.Id == orderId {
            t.Errorf("Order %s should not be in results", orderId)
        }
    }
}
```

## Testing Principles

1. **No event verification** - queries don't emit events
2. **Test view model shape** - ensure correct fields included/excluded
3. **Test filtering logic** - status, price range
4. **Test sorting** - verify order of results
5. **Compile-time privacy checks** - verify excluded fields don't exist
