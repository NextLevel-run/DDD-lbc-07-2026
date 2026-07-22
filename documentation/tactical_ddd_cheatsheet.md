# DDD Tactical Patterns Cheat Sheet

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          DRIVING ADAPTERS                                   │
│                    (HTTP Handlers, CLI, Consumers)                          │
│                              ↓ calls                                        │
├─────────────────────────────────────────────────────────────────────────────┤
│                         APPLICATION LAYER                                   │
│              Commands (write)    ←→    Queries (read)                       │
│                              ↓ uses                                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                           DOMAIN LAYER                                      │
│     Entities | Aggregates | Value Objects | Events | Repository Ports       │
│                              ↑ implements                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                          DRIVEN ADAPTERS                                    │
│              (In-Memory, PostgreSQL, Kafka, External APIs)                  │
└─────────────────────────────────────────────────────────────────────────────┘

DEPENDENCY RULE: Adapter → Application → Domain (NEVER reverse)
```

**Hexagonal Architecture** (Ports & Adapters) isolates pure business logic in the domain layer, making it framework-independent and easy to test. The **Dependency Inversion Principle** ensures high-level modules (domain, application) never depend on low-level ones (infrastructure)—instead, the domain defines interfaces (ports) that adapters implement. This allows swapping databases, message brokers, or HTTP frameworks without touching business code. **CQRS** (Command Query Responsibility Segregation) separates write operations (commands that modify state and emit events) from read operations (queries optimized for specific views), enabling independent scaling and optimization of each path.

## Folder Structure

```
internal/{bounded-context}/
├── domain/           # Entities, events, repository interfaces (ports)
├── application/
│   ├── command/      # Write operations (CQRS)
│   └── query/        # Read operations (CQRS)
└── adapter/
    ├── driven/       # Outbound: inmemory/, postgres/, kafka/
    └── driving/      # Inbound: http/, consumer/
```

---

## Tactical Patterns Quick Reference

| Pattern | Layer | Purpose | Implementation Details |
|---------|-------|---------|------------------------|
| **Aggregate Root** | Domain | Entry point entity with global identity & repository | [domain-aggregates.md](./.claude/skills/tactical-ddd/references/domain-aggregates.md) |
| **Internal Entity** | Domain | Entity within an aggregate, local identity | [domain-aggregates.md](./.claude/skills/tactical-ddd/references/domain-aggregates.md) |
| **Value Object** | Domain | Immutable, identity-less concept | [domain-aggregates.md](./.claude/skills/tactical-ddd/references/domain-aggregates.md) |
| **Domain Event** | Domain | Fact that happened (past tense) | [domain-events.md](./.claude/skills/tactical-ddd/references/domain-events.md) |
| **Repository Port** | Domain | Persistence interface | [domain-repositories.md](./.claude/skills/tactical-ddd/references/domain-repositories.md) |
| **Repository Adapter** | Adapter | Persistence implementation | [domain-repositories.md](./.claude/skills/tactical-ddd/references/domain-repositories.md) |
| **Command** | Application | Write operation (returns ID or error) | [application-commands.md](./.claude/skills/tactical-ddd/references/application-commands.md) |
| **Query** | Application | Read operation (returns View) | [application-queries.md](./.claude/skills/tactical-ddd/references/application-queries.md) |
| **HTTP Handler** | Adapter | REST endpoint (driving) | [adapter-http.md](./.claude/skills/tactical-ddd/references/adapter-http.md) |
| **Event Consumer** | Adapter | Reacts to events (driving) | [adapter-consumers.md](./.claude/skills/tactical-ddd/references/adapter-consumers.md) |

---

## Workflow: Adding a New Feature

1. **Domain** → Aggregate behavior + event + repository method (if needed)
2. **Application** → Command or Query
3. **Adapter** → HTTP handler or Consumer
4. **Wire** → Register in `cmd/api/main.go`

> **Claude Prompt**: `Use tactical-ddd skill. Implement feature {description} following the workflow`

---

## Pattern Details

### Aggregate

| Aspect | Description |
|--------|-------------|
| **Definition** | An **Aggregate** is a cluster of related entities and value objects treated as a single transactional unit. The **Aggregate Root** is the entry point entity with a globally unique identity—only aggregate roots have repositories and public constructors. **Internal Entities** exist within the aggregate with locally unique IDs, accessed only through the root. |
| **Rules** | Constructor generates ID, `Validate()` enforces rules, behavior methods modify state, no external deps |
| **Advantages** | Pure business logic is easy to reason about, test in isolation without mocks, remains completely independent of frameworks, databases, or HTTP concerns. |
| **Naming** | Aggregate Root: `Order` (singular noun) · Internal Entity: `OrderItem` · Constructor: `NewOrder()` · Errors: `ErrEmptyCart` · File: `order.go` |
| **Testing** | Unit tests with no mocks. Use table-driven tests for validation scenarios. Test behavior methods for both success and error paths. |
| **Claude Prompt** | `Use tactical-ddd skill. Create aggregate {Name} with fields {...} and behavior {action}` |
| **Reference** | [domain-aggregates.md](./.claude/skills/tactical-ddd/references/domain-aggregates.md) |

```go
// Pseudo-code skeleton
var ErrInvalidState = errors.New("invalid state")
type Entity struct { Id string; Field1 string; Status string }
func NewEntity(field1 string) *Entity { return &Entity{Id: uuid(), Field1: field1, Status: "active"} }
func (e *Entity) Validate() error { if e.Field1 == "" { return ErrNoField1 }; return nil }
func (e *Entity) DoAction(reason string) error { if e.Status != "active" { return ErrInvalidState }; e.Status = "done"; return nil }
```

---

### Domain Event

| Aspect | Description |
|--------|-------------|
| **Definition** | A **Domain Event** is an immutable record of something significant that happened in the domain—a fact that domain experts care about. Events are always named in **past tense** because they represent completed actions, not intentions. They enable loose coupling between components, provide an audit trail, and allow other bounded contexts to react without direct dependencies. |
| **Rules** | Struct ends with `Event`, type string is past tense, emit AFTER persistence |
| **Advantages** | Decoupling, audit trail, cross-context integration |
| **Naming** | Struct: `EntityActionEvent` · Type string: `"EntityAction"` · Constructor: `NewEntityActionEventFrom()` |
| **Testing** | Events are tested via command tests using `EventCollector`. Verify event type, payload content, and that events are emitted only after successful persistence. |
| **Claude Prompt** | `Use tactical-ddd skill. Create domain event for when {entity} is {action}` |
| **Reference** | [domain-events.md](./.claude/skills/tactical-ddd/references/domain-events.md) |

```go
type EntityActionEvent struct { id, eventType string; emitedAt time.Time; Entity *Entity }
func NewEntityActionEventFrom(e *Entity) *EntityActionEvent {
    return &EntityActionEvent{id: uuid(), eventType: "EntityAction", emitedAt: time.Now(), Entity: e}
}
func (e *EntityActionEvent) EventType() string { return e.eventType }
```

---

### Repository (Port + Adapter)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              DOMAIN LAYER                                   │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  EntityRepository (Port/Interface)                                   │   │
│  │  ─────────────────────────────────                                   │   │
│  │  Save(e *Entity) error                                               │   │
│  │  GetById(id string) (*Entity, error)                                 │   │
│  │  FindAll(filters) ([]*Entity, error)                                 │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
└─────────────────────────────────────────────────────────────────────────────┘
                                     ▲ implements
┌─────────────────────────────────────────────────────────────────────────────┐
│                             ADAPTER LAYER                                   │
│  ┌──────────────────────┐              ┌──────────────────────┐            │
│  │  InMemoryRepository  │              │ PostgresRepository   │            │
│  │  (for tests)         │              │ (for production)     │            │
│  └──────────────────────┘              └──────────────────────┘            │
└─────────────────────────────────────────────────────────────────────────────┘
```

| Aspect | Description |
|--------|-------------|
| **Definition** | A **Repository** abstracts persistence operations behind an interface. The **Port** (interface) lives in the domain layer and defines *what* operations are needed using only domain types. The **Adapter** (implementation) lives in the adapter layer and defines *how* those operations are performed (in-memory, PostgreSQL, MongoDB, etc.). This separation allows the domain to remain pure while infrastructure can be swapped freely. |
| **Rules** | The interface uses only domain types—no SQL, no ORM structs. |
| **Advantages** | Domain stays pure and testable. Swap implementations without changing business logic (in-memory for tests, PostgreSQL for production). Clear contract between domain and infrastructure. |
| **Naming** | Interface: `EntityRepository` · Implementation: `InMemoryEntityRepository`, `PostgresEntityRepository` · File: `repository.go` |
| **Testing** | Commands and queries are tested with in-memory implementations. Test helpers: `createAndSaveOrder()`, `assertNoOrdersInRepository()`. Verify persistence after commands, verify no persistence on validation errors. |
| **Claude Prompt** | `Use tactical-ddd skill. Add repository method {FindBy...} for {Entity}` |
| **Reference** | [domain-repositories.md](./.claude/skills/tactical-ddd/references/domain-repositories.md) |

```go
// Port (domain/)
type FindAllFilters struct { Category *string; IsActive *bool }  // nil = no filter
type EntityRepository interface {
    Save(e *Entity) error
    GetById(id string) (*Entity, error)
    FindAll(filters FindAllFilters) ([]*Entity, error)
}
// Adapter (adapter/driven/inmemory/)
type InMemoryRepo struct { data map[string]*Entity; mu sync.RWMutex }
func (r *InMemoryRepo) Save(e *Entity) error { r.mu.Lock(); defer r.mu.Unlock(); r.data[e.Id] = e; return nil }
```

---

### Command (CQRS Write)

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│    Args     │───▶│   Entity    │───▶│ Repository  │───▶│  EventBus   │
│  (input)    │    │  (create +  │    │   (save)    │    │  (publish)  │
│             │    │  validate)  │    │             │    │             │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                         │                   │
                         ▼                   ▼
                   [Err on fail]      [Err on fail]
```

| Aspect | Description |
|--------|-------------|
| **Definition** | A **Command** represents a write operation that modifies system state. Commands follow a strict flow: **Create** domain object → **Validate** business rules → **Persist** to repository → **Emit** domain event. They are named with imperative verbs (`Register`, `Delete`, `Publish`) and return minimal data—typically just the created ID or an error. Commands orchestrate domain objects but contain no business logic themselves. |
| **Rules** | One command per file (`create.go`, `delete.go`). Functional style: `BuildXCommand(deps) XCommand` returns a function. Args struct encapsulates all inputs. Emit events **only after** successful persistence. Return ID for creates, error-only for updates/deletes. |
| **Advantages** | Clear, auditable write path. Separation from queries allows independent optimization. Event emission enables reactive patterns. Easy to test with in-memory repo and event collector. |
| **Naming** | Type: `CreateEntityCommand` · Args: `CreateEntityCommandArgs` · Builder: `BuildCreateEntityCommand()` · File: `create.go` |
| **Testing** | Use in-memory repo + `EventCollector`. Verify three things: return value, persistence (via repo), event emission. On error: verify no persistence and no events. Use table-driven tests for validation scenarios. |
| **Claude Prompt** | `Use tactical-ddd skill. Create command {Action}{Entity} that {does what}` |
| **Reference** | [application-commands.md](./.claude/skills/tactical-ddd/references/application-commands.md) |

```go
type ActionCommandArgs struct { Field1 string; Field2 uint }
type ActionCommand func(args ActionCommandArgs) (string, error)
func BuildActionCommand(repo Repository, eventBus Bus) ActionCommand {
    return func(args ActionCommandArgs) (string, error) {
        entity := NewEntity(args.Field1, args.Field2)         // 1. Create
        if err := entity.Validate(); err != nil { return "", err }  // 2. Validate
        if err := repo.Save(entity); err != nil { return "", err }  // 3. Persist
        eventBus.Publish(NewEntityCreatedEventFrom(entity))         // 4. Emit
        return entity.Id, nil
    }
}
```

---

### Query (CQRS Read)

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│    Args     │───▶│ Repository  │───▶│   Entity    │───▶│    View     │
│  (filters)  │    │   (fetch)   │    │  (filter +  │    │  (safe DTO) │
│             │    │             │    │   check)    │    │             │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                         │                   │
                         ▼                   ▼
                   [ErrNotFound]      [Hide sensitive]
```

| Aspect | Description |
|--------|-------------|
| **Definition** | A **Query** represents a read operation that retrieves data without modifying state. Queries return **View models** (also called presenters or DTOs) tailored for specific use cases—not domain entities directly. This allows filtering sensitive data, shaping responses for different consumers (list view vs. detail view), and optimizing read paths independently from writes. |
| **Rules** | Return purpose-specific view structs, not entities. Filter sensitive data (emails, internal IDs). Different shapes for different needs: `EntityListItem` (compact) vs. `EntityView` (detailed). Sorting/pagination in application layer. No Events |
| **Advantages** | Privacy by design—sensitive fields never leak. Different optimizations for read vs. write paths. View models can evolve independently from domain. Simpler, faster reads without aggregate loading overhead. |
| **Naming** | Query: `GetEntityQuery`, `FindEntitiesQuery` · View: `EntityView`, `EntityListItem` · Args: `FindEntitiesArgs` · File: `get_entity.go` |
| **Testing** | Use in-memory repo. Verify view shape (correct fields included, sensitive fields excluded). Test filters and sorting. Test that deleted/inactive items are excluded. Compile-time check: accessing excluded field causes error. |
| **Claude Prompt** | `Use tactical-ddd skill. Create query Get{Entity} returning view without {sensitive fields}` |
| **Reference** | [application-queries.md](./.claude/skills/tactical-ddd/references/application-queries.md) |

```go
type EntityView struct { Id, Title string; Price uint }  // No SellerEmail!
type EntityListItem struct { Id, Title string }          // Even less for lists
type GetEntityQuery func(id string) (*EntityView, error)
func BuildGetEntityQuery(repo Repository) GetEntityQuery {
    return func(id string) (*EntityView, error) {
        entity, err := repo.GetById(id)
        if err != nil { return nil, err }
        if !entity.IsActive { return nil, ErrNotFound }  // Business rule
        return &EntityView{Id: entity.Id, Title: entity.Title, Price: entity.Price}, nil
    }
}
```

---

### HTTP Handler (Driving Adapter)

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   HTTP      │───▶│   Handler   │───▶│  Command /  │───▶│    HTTP     │
│  Request    │    │  (parse +   │    │   Query     │    │  Response   │
│   + DTO     │    │   map err)  │    │             │    │   + DTO     │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                         │                   │
                         ▼                   ▼
                   [400/405/415]      [Domain Err → HTTP Code]
```

| Aspect | Description |
|--------|-------------|
| **Definition** | An **HTTP Handler** is a driving adapter that translates HTTP requests into command/query calls and domain responses back into HTTP responses. Handlers are thin translation layers with **no business logic**—they parse requests, delegate to application layer, map domain errors to HTTP status codes, and serialize responses. DTOs (Data Transfer Objects) define the API contract separately from domain types. |
| **Rules** | DTOs separate from domain, map domain errors to status codes, no business logic |
| **Advantages** | Clean API contract, testable with httptest |
| **Naming** | Handler: `Handler` struct with methods · DTOs: `EntityRequest`, `EntityResponse`, `ErrorResponse` · File: `handler.go`, `dto.go` |
| **Testing** | Use `httptest.NewRequest` and `httptest.NewRecorder`. Mock commands/queries to isolate HTTP concerns. Test: status codes, JSON encoding, error mapping, wrong method handling, invalid JSON handling. |
| **Claude Prompt** | `Use tactical-ddd skill. Create HTTP handler for {POST/GET} /path calling {command/query}` |
| **Reference** | [adapter-http.md](./.claude/skills/tactical-ddd/references/adapter-http.md) |

```go
func (h *Handler) RegisterEntity(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost { h.error(w, "Method not allowed", 405); return }
    var req RegisterRequest; json.NewDecoder(r.Body).Decode(&req)      // 1. Parse
    id, err := h.registerCommand(CommandArgs{Field1: req.Field1})       // 2. Execute
    if err != nil { h.handleDomainError(w, err); return }               // 3. Map errors
    h.success(w, RegisterResponse{ID: id}, 201)                         // 4. Respond
}
func (h *Handler) handleDomainError(w http.ResponseWriter, err error) {
    switch { case errors.Is(err, ErrNoTitle): h.error(w, "Title required", 400)
             case errors.Is(err, ErrNotFound): h.error(w, "Not found", 404)
             default: h.error(w, err.Error(), 500) }
}
```

---

### Event Consumer (Driving Adapter)

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  EventBus   │───▶│  Consumer   │───▶│  Command /  │───▶│ Side Effect │
│ (subscribe) │    │  (type      │    │  Service    │    │ (email, DB, │
│             │    │   assert)   │    │             │    │  external)  │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                         │
                         ▼
                   [Ignore wrong type]
```

| Aspect | Description |
|--------|-------------|
| **Definition** | An **Event Consumer** is a driving adapter that subscribes to domain events and triggers side effects—sending emails, updating read models, calling external services, or executing commands in other bounded contexts. Consumers are decoupled from event producers: they react to facts without the producer knowing who's listening. Multiple consumers can subscribe to the same event type (fan-out pattern). |
| **Rules** | Type assertion for payload, don't crash on wrong type, register at startup |
| **Advantages** | Decoupled reactions, fan-out support
| **Naming** | Constructor: `NewAdminEmailConsumer()`, `NewAnalyticsConsumer()` · File: `admin_email_consumer.go` |
| **Testing** | Use sync event bus for deterministic tests. Publish event directly, verify side effect (email sent, command called). Test type safety: wrong event type shouldn't crash. Integration test: command → event → consumer → side effect. |
| **Claude Prompt** | `Use tactical-ddd skill. Create consumer that reacts to {Event} by {action}` |
| **Reference** | [adapter-consumers.md](./.claude/skills/tactical-ddd/references/adapter-consumers.md) |

```go
func NewNotificationConsumer(eventBus Bus, notifyCommand NotifyCommand) error {
    return eventBus.Subscribe("EntityCreated", func(evt DomainEvent) error {
        event, ok := evt.(*EntityCreatedEvent)
        if !ok { return nil }  // Type safety: ignore wrong types
        return notifyCommand(NotifyArgs{EntityId: event.Entity.Id})
    })
}
// Registration in main.go
func main() {
    eventBus := eventbus.NewInMemoryEventBus()
    notifyCmd := command.BuildNotifyCommand(mailer)
    consumer.NewNotificationConsumer(eventBus, notifyCmd)  // Subscribe before commands emit
    registerCmd := command.BuildRegisterCommand(repo, eventBus)
}
```

---

## Skill Reference

For detailed implementation patterns with full code examples, see the tactical-ddd skill:

`.claude/skills/tactical-ddd/SKILL.md`

---

## Complete Prompt Template: Domain to Wire

Use this template when implementing a new feature end-to-end. Copy, customize the `{placeholders}`, and paste into Claude Code.

```markdown
# Feature: {FeatureName}

## Objective

Implement {brief description of what the feature does} in the {bounded-context} bounded context.

## Context Files

- Bounded context: `internal/{bounded-context}/agent.md`
- Use your `tactical-ddd` skill

## Domain Layer

### Aggregate

Add a new `{Aggregate}` aggregate root:
- Fields: {field1} {type}, {field2} {type}
- Business rules: {rule1}, {rule2}
- Behavior methods: {method1}, {method2}
- Errors: `Err{Condition1}`, `Err{Condition2}`
- Test

### Domain Event

Create `{Aggregate}{Action}Event`:
- Payload:

### Repository
Create a new `{Aggregate}Repository`
- Test

## Application Layer

### Command
Create `{Action}{Entity}Command`:
- Args: `{Field1} {type}`, `{Field2} {type}`
- Returns: `{string|error}` (ID for creates, error-only for updates)
- Emits: {Entity}{Action}Event
- Errors to handle: {list domain errors and what causes them}
- Test


## Adapter Layer

### HTTP Handler
Add a new Endpoint: `{METHOD} /{path}`
- Wire to the `{Action}{Entity}Command`
- Convert Command Errors to HTTP status codes
- Test
- Wire in the main.go file.


### Event Consumer
Create a new Event Consumer that reacts to the `{Aggregate}{Action}Event` by executing the `{Action}{Aggregate}Command`
- Test
- Wire in the main.go file.


**Instructions:**
1. Make a plan before implementing.
2. Ask questions if requirements are unclear or if you miss information.
3. Follow TDD: write tests first, then implementation.
4. Implement layer by layer: Domain → Application → Adapter → Wire.
```
