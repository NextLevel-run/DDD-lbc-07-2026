# Domain Aggregates

## Concept

An **Aggregate** is a cluster of entities and value objects treated as a single unit for data changes. Each aggregate has an **Aggregate Root** — the only entity through which external code can interact with the aggregate.

Key distinctions:
- **Aggregate Root**: The entry point entity with a globally unique identity. Only aggregate roots have repositories and public constructors. Example: `Order`
- **Internal Entity**: An entity that exists only within an aggregate, with a locally unique identity. Accessed only through the root. Example: `OrderItem`
- **Value Object**: Immutable, identity-less concept defined by its attributes. Constructed via factory functions that validate invariants. Example: `Money`, `Email`, `Address`

In this example, `Order` is an aggregate root containing `OrderItem` internal entities and several value objects.

### Aggregate Versioning

Every aggregate root maintains a **version** number (starting at 1) that increments on each state change. This serves two purposes:

1. **Event idempotency**: Domain events include `AggregateID + Version`, allowing consumers to detect and skip duplicate/late events
2. **Optimistic locking**: Repositories can use the version to detect concurrent modifications and reject stale updates

## Implementation Pattern

### File Location
```
internal/{bounded-context}/domain/{aggregate}.go
```

### Structure

```go
// internal/order/domain/order.go

package domain

import (
    "errors"
    "net/mail"
    "time"

    "github.com/google/uuid"
)

// ============================================
// ERRORS
// ============================================

var (
    ErrInvalidEmail        = errors.New("invalid email format")
    ErrInvalidCurrency     = errors.New("invalid currency")
    ErrNegativeAmount      = errors.New("amount cannot be negative")
    ErrEmptyAddress        = errors.New("address fields cannot be empty")
    ErrEmptyOrderItems     = errors.New("order must have at least one item")
    ErrInvalidQuantity     = errors.New("quantity must be greater than zero")
    ErrOrderAlreadyShipped = errors.New("cannot modify shipped order")
    ErrInvalidTransition   = errors.New("invalid status transition")
)

// ============================================
// AGGREGATE ROOT: Order
// ============================================

type Order struct {
    id              string
    version         int
    customerEmail   Email
    items           []OrderItem
    shippingAddress Address
    status          OrderStatus
    placedAt        time.Time
}

func NewOrder(customerEmail string, street string, city string, postalCode string, country string, items []OrderItem) (*Order, error) {
    email, err := NewEmail(customerEmail)
    if err != nil {
        return nil, err
    }
    address, err := NewAddress(street, city, postalCode, country)
    if err != nil {
        return nil, err
    }
    if len(items) == 0 {
        return nil, ErrEmptyOrderItems
    }
    return &Order{
        id:              uuid.New().String(),
        version:         1,
        customerEmail:   email,
        items:           items,
        shippingAddress: address,
        status:          OrderPending,
        placedAt:        time.Now(),
    }, nil
}

func (o *Order) ID() string               { return o.id }
func (o *Order) Version() int             { return o.version }
func (o *Order) CustomerEmail() Email     { return o.customerEmail }
func (o *Order) Items() []OrderItem       { return o.items }
func (o *Order) ShippingAddress() Address { return o.shippingAddress }
func (o *Order) Status() OrderStatus      { return o.status }
func (o *Order) PlacedAt() time.Time      { return o.placedAt }

func (o *Order) TotalAmount() (Money, error) {
    if len(o.items) == 0 {
        return Money{}, ErrEmptyOrderItems
    }

    total, _ := NewMoney(0, o.items[0].unitPrice.Currency())
    for _, item := range o.items {
        itemTotal, err := item.Total()
        if err != nil {
            return Money{}, err
        }
        total, err = total.Add(itemTotal)
        if err != nil {
            return Money{}, err
        }
    }
    return total, nil
}

func (o *Order) Confirm() error {
    if o.status != OrderPending {
        return ErrInvalidTransition
    }
    o.status = OrderConfirmed
    o.version++
    return nil
}

func (o *Order) Ship() error {
    if o.status != OrderConfirmed {
        return ErrInvalidTransition
    }
    o.status = OrderShipped
    o.version++
    return nil
}

func (o *Order) Deliver() error {
    if o.status != OrderShipped {
        return ErrInvalidTransition
    }
    o.status = OrderDelivered
    o.version++
    return nil
}

func (o *Order) Cancel() error {
    if o.status == OrderShipped || o.status == OrderDelivered {
        return ErrOrderAlreadyShipped
    }
    o.status = OrderCancelled
    o.version++
    return nil
}

// ============================================
// ENUM: OrderStatus
// ============================================

type OrderStatus string

const (
    OrderPending   OrderStatus = "Pending"
    OrderConfirmed OrderStatus = "Confirmed"
    OrderShipped   OrderStatus = "Shipped"
    OrderDelivered OrderStatus = "Delivered"
    OrderCancelled OrderStatus = "Cancelled"
)

// ============================================
// INTERNAL ENTITY: OrderItem
// ============================================

type OrderItem struct {
    id        string
    productID string
    quantity  int
    unitPrice Money
}

func NewOrderItem(productID string, quantity int, amount int64, currency Currency) (OrderItem, error) {
    if quantity <= 0 {
        return OrderItem{}, ErrInvalidQuantity
    }
    unitPrice, err := NewMoney(amount, currency)
    if err != nil {
        return OrderItem{}, err
    }
    return OrderItem{
        id:        uuid.New().String(),
        productID: productID,
        quantity:  quantity,
        unitPrice: unitPrice,
    }, nil
}

func (i OrderItem) ID() string        { return i.id }
func (i OrderItem) ProductID() string { return i.productID }
func (i OrderItem) Quantity() int     { return i.quantity }
func (i OrderItem) UnitPrice() Money  { return i.unitPrice }

func (i OrderItem) Total() (Money, error) {
    return i.unitPrice.Multiply(i.quantity)
}

// ============================================
// VALUE OBJECT: Money
// ============================================

type Money struct {
    amount   int64 // in cents
    currency Currency
}

func NewMoney(amount int64, currency Currency) (Money, error) {
    if amount < 0 {
        return Money{}, ErrNegativeAmount
    }
    if !currency.IsValid() {
        return Money{}, ErrInvalidCurrency
    }
    return Money{amount: amount, currency: currency}, nil
}

func (m Money) Amount() int64      { return m.amount }
func (m Money) Currency() Currency { return m.currency }

func (m Money) Add(other Money) (Money, error) {
    if m.currency != other.currency {
        return Money{}, ErrInvalidCurrency
    }
    return NewMoney(m.amount+other.amount, m.currency)
}

func (m Money) Multiply(quantity int) (Money, error) {
    return NewMoney(m.amount*int64(quantity), m.currency)
}

// ============================================
// VALUE OBJECT: Currency (enum)
// ============================================

type Currency string

const (
    EUR Currency = "EUR"
    USD Currency = "USD"
)

var validCurrencies = map[Currency]struct{}{
    EUR: {},
    USD: {},
}

func (c Currency) IsValid() bool {
    _, exists := validCurrencies[c]
    return exists
}

// ============================================
// VALUE OBJECT: Email
// ============================================

type Email struct {
    value string
}

func NewEmail(value string) (Email, error) {
    addr, err := mail.ParseAddress(value)
    if err != nil {
        return Email{}, ErrInvalidEmail
    }
    return Email{value: addr.Address}, nil
}

func (e Email) String() string { return e.value }

// ============================================
// VALUE OBJECT: Address
// ============================================

type Address struct {
    street     string
    city       string
    postalCode string
    country    string
}

func NewAddress(street, city, postalCode, country string) (Address, error) {
    if street == "" || city == "" || postalCode == "" || country == "" {
        return Address{}, ErrEmptyAddress
    }
    return Address{
        street:     street,
        city:       city,
        postalCode: postalCode,
        country:    country,
    }, nil
}

func (a Address) Street() string     { return a.street }
func (a Address) City() string       { return a.city }
func (a Address) PostalCode() string { return a.postalCode }
func (a Address) Country() string    { return a.country }
```

### Key Principles

1. **Aggregate Root first** - the main entry point for all operations
2. **Constructors take primitives** - aggregate and entity constructors accept primitive types (strings, ints, etc.) and build value objects internally. This keeps the public API simple and encapsulates validation within the aggregate
3. **Value Objects are immutable** - no setters, created via constructors that validate
4. **Constructors validate invariants** - `NewMoney()` rejects negative amounts, `NewEmail()` validates format
5. **Enums as typed strings** - `Currency`, `OrderStatus` with validation methods
6. **Internal entities accessed through root** - `OrderItem` has no repository
7. **Behavior methods** (`Confirm`, `Ship`, `Cancel`) enforce state transitions
8. **Domain errors** as exported package-level variables
9. **No external dependencies** - pure Go only (except uuid for IDs)
10. **Version tracking** - aggregates start at version 1 and increment on each mutation. Include version in domain events for consumer idempotency

## Testing

### File Location
```
internal/{bounded-context}/domain/{aggregate}_test.go
```

### Test Pattern

```go
// internal/order/domain/order_test.go

package domain

import (
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/require"
)

// ============================================
// VALUE OBJECT TESTS
// ============================================

func TestNewMoney_RejectsNegativeAmount(t *testing.T) {
    // When
    _, err := NewMoney(-100, EUR)

    // Then
    require.ErrorIs(t, err, ErrNegativeAmount)
}

func TestNewMoney_RejectsInvalidCurrency(t *testing.T) {
    // When
    _, err := NewMoney(100, Currency("INVALID"))

    // Then
    require.ErrorIs(t, err, ErrInvalidCurrency)
}

func TestMoney_Add_WorksWithSameCurrency(t *testing.T) {
    // Given
    m1, _ := NewMoney(1000, EUR)
    m2, _ := NewMoney(500, EUR)

    // When
    result, err := m1.Add(m2)

    // Then
    require.NoError(t, err)
    assert.Equal(t, int64(1500), result.Amount())
    assert.Equal(t, EUR, result.Currency())
}

func TestMoney_Add_RejectsDifferentCurrencies(t *testing.T) {
    // Given
    m1, _ := NewMoney(1000, EUR)
    m2, _ := NewMoney(500, USD)

    // When
    _, err := m1.Add(m2)

    // Then
    require.ErrorIs(t, err, ErrInvalidCurrency)
}

func TestNewEmail_ValidatesFormat(t *testing.T) {
    tests := []struct {
        name    string
        input   string
        wantErr bool
    }{
        {"ValidEmail", "user@example.com", false},
        {"ValidWithName", "John Doe <john@example.com>", false},
        {"InvalidNoAt", "invalid-email", true},
        {"InvalidEmpty", "", true},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            _, err := NewEmail(tt.input)
            if tt.wantErr {
                require.ErrorIs(t, err, ErrInvalidEmail)
            } else {
                require.NoError(t, err)
            }
        })
    }
}

// ============================================
// AGGREGATE ROOT TESTS
// ============================================

func TestNewOrder_CreatesWithPendingStatusAndVersion1(t *testing.T) {
    // Given
    item, _ := NewOrderItem("PROD-123", 2, 2999, EUR)

    // When
    order, err := NewOrder("customer@example.com", "123 Main St", "Paris", "75001", "France", []OrderItem{item})

    // Then
    require.NoError(t, err)
    assert.NotEmpty(t, order.ID())
    assert.Equal(t, 1, order.Version())
    assert.Equal(t, OrderPending, order.Status())
    assert.Len(t, order.Items(), 1)
}

func TestNewOrder_RejectsEmptyItems(t *testing.T) {
    // When
    _, err := NewOrder("customer@example.com", "123 Main St", "Paris", "75001", "France", []OrderItem{})

    // Then
    require.ErrorIs(t, err, ErrEmptyOrderItems)
}

func TestOrder_TotalAmount_CalculatesCorrectly(t *testing.T) {
    // Given
    item1, _ := NewOrderItem("PROD-1", 2, 1000, EUR) // 10.00 EUR
    item2, _ := NewOrderItem("PROD-2", 3, 500, EUR)  // 5.00 EUR
    order, _ := NewOrder("customer@example.com", "123 Main St", "Paris", "75001", "France", []OrderItem{item1, item2})

    // When
    total, err := order.TotalAmount()

    // Then
    require.NoError(t, err)
    assert.Equal(t, int64(3500), total.Amount()) // 2*1000 + 3*500
    assert.Equal(t, EUR, total.Currency())
}

// ============================================
// STATE TRANSITION TESTS
// ============================================

func TestOrder_Confirm_TransitionsFromPending(t *testing.T) {
    // Given
    order := createTestOrder(t)
    assert.Equal(t, OrderPending, order.Status())
    assert.Equal(t, 1, order.Version())

    // When
    err := order.Confirm()

    // Then
    require.NoError(t, err)
    assert.Equal(t, OrderConfirmed, order.Status())
    assert.Equal(t, 2, order.Version())
}

func TestOrder_Ship_RequiresConfirmedStatus(t *testing.T) {
    // Given
    order := createTestOrder(t)

    // When
    err := order.Ship()

    // Then
    require.ErrorIs(t, err, ErrInvalidTransition)
}

func TestOrder_Cancel_NotAllowedAfterShipped(t *testing.T) {
    // Given
    order := createTestOrder(t)
    _ = order.Confirm()
    _ = order.Ship()

    // When
    err := order.Cancel()

    // Then
    require.ErrorIs(t, err, ErrOrderAlreadyShipped)
}

func TestOrder_FullLifecycle(t *testing.T) {
    // Given
    order := createTestOrder(t)
    assert.Equal(t, 1, order.Version())

    // When/Then - full happy path with version tracking
    require.NoError(t, order.Confirm())
    assert.Equal(t, OrderConfirmed, order.Status())
    assert.Equal(t, 2, order.Version())

    require.NoError(t, order.Ship())
    assert.Equal(t, OrderShipped, order.Status())
    assert.Equal(t, 3, order.Version())

    require.NoError(t, order.Deliver())
    assert.Equal(t, OrderDelivered, order.Status())
    assert.Equal(t, 4, order.Version())
}

// ============================================
// TEST HELPERS
// ============================================

func createTestOrder(t *testing.T) *Order {
    t.Helper()
    item, _ := NewOrderItem("PROD-123", 1, 2999, EUR)
    order, err := NewOrder("customer@example.com", "123 Main St", "Paris", "75001", "France", []OrderItem{item})
    require.NoError(t, err)
    return order
}
```

### Testing Principles

- **No mocks needed** - aggregates are pure domain objects
- **Test Value Object validation** - constructors should reject invalid data
- **Test state transitions** - verify allowed/forbidden transitions
- **Table-driven tests** for validation scenarios
- **Test helpers** for creating valid test objects
- Use `require.*` for critical assertions, `assert.*` for non-critical
