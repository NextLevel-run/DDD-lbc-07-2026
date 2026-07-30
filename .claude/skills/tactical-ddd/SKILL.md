---
name: tactical-ddd
description: Guide for implementing Domain-Driven Design patterns with Hexagonal Architecture in Go. This skill should be used when users request to create entities, aggregates, domain events, commands, queries, repositories, or HTTP handlers. Use when adding new features, implementing use cases, or setting up bounded contexts in the ddd-second-hand-marketplace codebase.
---

# Tactical DDD Skill

A pedagogical guide for implementing Domain-Driven Design patterns with Hexagonal Architecture in Go.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                     DRIVING ADAPTERS                            │
│  (HTTP Handlers, CLI, Consumers)                                │
│         ↓ calls                                                 │
├─────────────────────────────────────────────────────────────────┤
│                   APPLICATION LAYER                             │
│  Commands (write) ←→ Queries (read)                             │
│         ↓ uses                                                  │
├─────────────────────────────────────────────────────────────────┤
│                     DOMAIN LAYER                                │
│  Entities, Aggregates, Value Objects, Events, Repository Ports  │
│         ↑ implements                                            │
├─────────────────────────────────────────────────────────────────┤
│                     DRIVEN ADAPTERS                             │
│  (In-Memory, Database, External Services)                       │
└─────────────────────────────────────────────────────────────────┘
```

**Dependency Rule**: Adapter → Application → Domain (never the reverse)

## Reference Documentation

| Topic | File | Description |
|-------|------|-------------|
| Aggregates | [references/domain-aggregates.md](references/domain-aggregates.md) | Aggregate roots, internal entities, and value objects |
| Domain Events | [references/domain-events.md](references/domain-events.md) | Facts that happened in the domain |
| Repositories | [references/domain-repositories.md](references/domain-repositories.md) | Persistence ports and driven adapters |
| Commands | [references/application-commands.md](references/application-commands.md) | Write operations (CQRS) |
| Queries | [references/application-queries.md](references/application-queries.md) | Read operations (CQRS) |
| Event Consumers | [references/adapter-consumers.md](references/adapter-consumers.md) | Event handlers as driving adapters |
| HTTP Handlers | [references/adapter-http.md](references/adapter-http.md) | REST API driving adapters |
| Codebase Organization | [references/codebase-organization.md](references/codebase-organization.md) | Project structure and conventions |
| PKG Infrastructure | [references/pkg-infrastructure.md](references/pkg-infrastructure.md) | Shared infrastructure packages |

## Quick Search Patterns

To locate specific patterns in reference files:

| Pattern | Command |
|---------|---------|
| Command testing setup | `grep -n "testSetup\|setupTest" references/application-commands.md` |
| Event emission | `grep -rn "eventBus.Publish\|Emit" references/` |
| HTTP validation | `grep -n "BadRequest\|validation" references/adapter-http.md` |
| Repository methods | `grep -n "func.*Repository" references/domain-repositories.md` |
| Consumer subscription | `grep -n "Subscribe\|EventHandler" references/adapter-consumers.md` |

## Key Principles

1. **Domain First**: Start with entities, value objects, and business rules
2. **Ports & Adapters**: Domain defines interfaces (ports), infrastructure implements them
3. **CQRS**: Separate read (Query) and write (Command) operations
4. **Event-Driven**: Commands emit domain events after successful persistence
5. **Pure Domain**: No external dependencies in the domain layer

## Bounded Context Structure

```
internal/
└── {bounded-context}/
    ├── domain/               # Pure business logic
    │   ├── {aggregate}.go   # Aggregate roots
    │   ├── repository.go    # Repository interfaces (ports)
    │   └── event.go         # Domain events
    ├── application/          # Use cases
    │   ├── command/         # Write operations
    │   └── query/           # Read operations
    └── adapter/             # Infrastructure
        ├── driven/          # Outbound (repositories)
        │   └── inmemory/
        └── driving/         # Inbound (HTTP, consumers)
            ├── http/
            └── consumer/
```

> **Note**: Each module represents a **subdomain** (problem space) implemented as a **bounded context** (solution space). Subdomains are business segments (Core, Supporting, Generic), while bounded contexts are the technical boundaries where a ubiquitous language applies. In this project, there's a 1:1 mapping between them.

## Naming Conventions Summary

| Concept | Convention | Example |
|---------|------------|---------|
| Aggregate Root | Singular noun | `Order` |
| Internal Entity | Singular noun | `OrderItem`, `BidHistory` |
| Constructor | `New` prefix | `NewOrder()` |
| Error | `Err` prefix | `ErrEmptyOrderItems` |
| Command | Verb + Noun + "Command" | `PlaceOrderCommand` |
| Event struct | Past tense + "Event" | `OrderPlacedEvent` |
| Event type | Past tense string | `"OrderPlaced"` |
| Repository | AggregateRoot + "Repository" | `OrderRepository` |
| View/Presenter | AggregateRoot + "View" | `OrderView` |

## Workflow: Adding a New Feature

Follow these steps in order when implementing a new feature:

### Step 1: Define Domain Changes
- Read [domain-aggregates.md](references/domain-aggregates.md)
- Add aggregate behavior methods or value objects to `domain/{aggregate}.go`
- Define domain errors as `Err*` variables
- Write domain unit tests in `domain/{aggregate}_test.go`

### Step 2: Create Domain Event (if needed)
- Read [domain-events.md](references/domain-events.md)
- Define event struct in `domain/event.go` with past-tense naming
- Implement `EventType() string` method
- Create convenience constructor `New{Event}From{Entity}()`

### Step 3: Build Command or Query
- For writes: Read [application-commands.md](references/application-commands.md)
  - Create `application/command/{action}.go`
  - Follow pattern: Create → Validate → Persist → Emit
- For reads: Read [application-queries.md](references/application-queries.md)
  - Create `application/query/{query_name}.go`
  - Return view/presenter structs, not domain entities

### Step 4: Write Application Tests
- Use in-memory repository from `adapter/driven/inmemory/`
- Use `EventCollector` from `pkg/eventbus/testing/`
- Verify: return value, persistence, event emission
- Test error cases verify no side effects

### Step 5: Add Driving Adapter
- For HTTP: Read [adapter-http.md](references/adapter-http.md)
  - Create DTO structs in `adapter/driving/http/dto.go`
  - Add handler method to `adapter/driving/http/handler.go`
  - Register route
- For consumers: Read [adapter-consumers.md](references/adapter-consumers.md)
  - Create consumer in `adapter/driving/consumer/`

### Step 6: Wire Dependencies
- Open `cmd/api/main.go`
- Instantiate command/query with `Build*` function
- Pass to HTTP handler or register consumer with event bus
