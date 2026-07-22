# Codebase Organization

## Project Structure

```
project-root/
├── cmd/
│   └── api/
│       └── main.go              # Application entry point, DI wiring
├── internal/                    # Private application code
│   └── {bounded-context}/       # One directory per bounded context
│       ├── domain/              # Pure business logic
│       │   ├── {entity}.go      # Aggregates and entities
│       │   ├── repository.go    # Repository interfaces (ports)
│       │   └── event.go         # Domain events
│       ├── application/         # Use cases
│       │   ├── command/         # Write operations
│       │   │   ├── place_order.go
│       │   │   ├── place_order_test.go
│       │   │   └── ...
│       │   └── query/           # Read operations
│       │       ├── get_order.go
│       │       ├── get_order_test.go
│       │       └── ...
│       └── adapter/             # Infrastructure implementations
│           ├── driven/          # Outbound adapters
│           │   └── inmemory/    # In-memory implementations
│           │       └── repository.go
│           └── driving/         # Inbound adapters
│               ├── http/        # HTTP handlers
│               │   ├── handler.go
│               │   ├── dto.go
│               │   └── handler_test.go
│               └── consumer/    # Event consumers
│                   └── confirmation_email_consumer.go
├── pkg/                         # Public shared libraries
│   ├── eventbus/               # Event bus infrastructure
│   │   ├── eventbus.go         # Interface
│   │   ├── inmemory_eventbus.go
│   │   └── testing/
│   │       └── event_collector.go
│   └── mailer/                 # Email infrastructure
│       ├── mailer.go           # Interface
│       ├── fake_mailer.go
│       └── testing/
│           └── mailer_spy.go
├── go.mod
├── go.sum
└── CLAUDE.md                   # Project documentation
```

> **Note**: Each module represents a **subdomain** (problem space) implemented as a **bounded context** (solution space). Subdomains are business segments (Core, Supporting, Generic), while bounded contexts are the technical boundaries where a ubiquitous language applies. In this project, there's a 1:1 mapping between them.

## Layer Responsibilities

### `cmd/` - Entry Points

Application startup, dependency injection, configuration.

```go
// cmd/api/main.go

func main() {
    // 1. Create infrastructure (driven adapters)
    repo := inmemory.NewInMemoryOrderRepository()
    eventBus := eventbus.NewInMemoryEventBus()
    mailer := mailer.NewFakeMailer()

    // 2. Build application layer (commands and queries)
    placeOrderCommand := command.BuildPlaceOrderCommand(repo, eventBus)
    cancelOrderCommand := command.BuildCancelOrderCommand(repo, eventBus)
    getOrderQuery := query.BuildGetOrderQuery(repo)
    findOrdersQuery := query.BuildFindOrdersQuery(repo)

    // 3. Register event consumers (driving adapters)
    consumer.NewLoggingConsumer(eventBus)
    consumer.NewConfirmationEmailConsumer(eventBus, notifyCustomerCommand)

    // 4. Create HTTP handler (driving adapter)
    handler := http.NewHandler(placeOrderCommand, cancelOrderCommand, ...)

    // 5. Start server
    log.Fatal(http.ListenAndServe(":8080", handler))
}
```

### `internal/` - Private Application Code

Code that should not be imported by other projects.

### `internal/{context}/domain/` - Business Logic

Pure domain logic with no external dependencies.

| File | Content |
|------|---------|
| `{entity}.go` | Aggregate root, entities, value objects |
| `repository.go` | Repository interfaces (ports) |
| `event.go` | Domain events |

### `internal/{context}/application/` - Use Cases

Orchestration layer using CQRS pattern.

| Directory | Purpose |
|-----------|---------|
| `command/` | Write operations (modify state) |
| `query/` | Read operations (read state) |

### `internal/{context}/adapter/` - Infrastructure

Implementations of ports and external interfaces.

| Directory | Type | Examples |
|-----------|------|----------|
| `driven/` | Outbound | Repositories, external services |
| `driving/` | Inbound | HTTP handlers, consumers |

### `pkg/` - Shared Infrastructure

Public packages that can be used across bounded contexts.

See [pkg-infrastructure.md](pkg-infrastructure.md) for details.

## Naming Conventions

### Files

| Type | Convention | Example |
|------|------------|---------|
| Entity | `{entity}.go` | `order.go` |
| Repository interface | `repository.go` | `repository.go` |
| Events | `event.go` | `event.go` |
| Command | `{action}.go` | `place_order.go`, `cancel_order.go` |
| Query | `{query_name}.go` | `get_order.go` |
| Test | `{file}_test.go` | `place_order_test.go` |
| DTO | `dto.go` | `dto.go` |

### Packages

| Type | Convention | Example |
|------|------------|---------|
| Bounded context | lowercase, singular | `order` |
| Layer | lowercase | `domain`, `application` |
| Sub-layer | lowercase | `command`, `query` |
| Adapter type | lowercase | `inmemory`, `postgres` |

### Types

| Concept | Convention | Example |
|---------|------------|---------|
| Entity | PascalCase, singular | `Order` |
| Constructor | `New` prefix | `NewOrder()` |
| Error | `Err` prefix | `ErrEmptyOrderItems` |
| Command func | `{Action}Command` | `PlaceOrderCommand` |
| Command builder | `Build{Action}Command` | `BuildPlaceOrderCommand` |
| Query func | `{Action}Query` | `GetOrderQuery` |
| Query builder | `Build{Action}Query` | `BuildGetOrderQuery` |
| Event | `{Entity}{Action}Event` | `OrderPlacedEvent` |
| Event type string | Past tense | `"OrderPlaced"` |
| Repository | `{Entity}Repository` | `OrderRepository` |
| View/Presenter | `{Entity}View` | `OrderView` |

## Dependency Rules

```
┌──────────────────────────────────────────────────────┐
│                     ADAPTERS                         │
│  (can import application, domain, pkg)               │
├──────────────────────────────────────────────────────┤
│                   APPLICATION                        │
│  (can import domain, pkg)                            │
├──────────────────────────────────────────────────────┤
│                      DOMAIN                          │
│  (can import pkg interfaces only)                    │
├──────────────────────────────────────────────────────┤
│                       PKG                            │
│  (no internal imports)                               │
└──────────────────────────────────────────────────────┘
```

### Valid Imports

```go
// adapter → application ✓
import "project/internal/order/application/command"

// adapter → domain ✓
import "project/internal/order/domain"

// application → domain ✓
import "project/internal/order/domain"

// application → pkg ✓
import "project/pkg/eventbus"

// domain → pkg (interface only) ✓
import "project/pkg/eventbus"  // Only for DomainEvent interface
```

### Invalid Imports

```go
// domain → adapter ✗
import "project/internal/order/adapter/driven/inmemory"

// domain → application ✗
import "project/internal/order/application/command"

// application → adapter ✗
import "project/internal/order/adapter/driven/inmemory"

// pkg → internal ✗
import "project/internal/order/domain"
```

## Adding a New Feature

### 1. Start with Domain

```go
// internal/order/domain/order.go
// Add new behavior or validation
func (o *Order) Ship() error {
    if o.status != OrderConfirmed {
        return ErrInvalidTransition
    }
    o.status = OrderShipped
    o.version++
    return nil
}
```

### 2. Add Domain Event (if needed)

```go
// internal/order/domain/event.go
type OrderShippedEvent struct { ... }
```

### 3. Create Command

```go
// internal/order/application/command/ship_order.go
func BuildShipOrderCommand(...) ShipOrderCommand { ... }
```

### 4. Add Tests

```go
// internal/order/application/command/ship_order_test.go
func TestShipOrderCommand_Success(t *testing.T) { ... }
```

### 5. Add HTTP Handler

```go
// internal/order/adapter/driving/http/handler.go
func (h *Handler) ShipOrder(w http.ResponseWriter, r *http.Request) { ... }
```

### 6. Wire in main.go

```go
// cmd/api/main.go
shipOrderCommand := command.BuildShipOrderCommand(repo, eventBus)
```

## Testing Organization

```
{file}.go           # Implementation
{file}_test.go      # Tests in same package

# Test utilities
internal/{context}/adapter/driven/inmemory/  # Reusable test doubles
pkg/{concept}/testing/                        # Infrastructure test utilities
```

### Test Helpers Location

| Helper Type | Location |
|-------------|----------|
| In-memory repos | `adapter/driven/inmemory/` |
| Event collector | `pkg/eventbus/testing/` |
| Mailer spy | `pkg/mailer/testing/` |
| Command setup | Same `_test.go` file |
