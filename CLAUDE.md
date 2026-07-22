# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

**ALWAYS** start replies with all "STARTER_CHARACTER" found in markdown files you adding to your context (default: 🍀) + space at the end. Stack emojis when requested, don't replace. No spaces between the emojis. Example: "🍀🌐 Bla bla bla..."

## Project Overview

A pedagogical **second-hand marketplace** using **Domain-Driven Design (DDD)** and **Hexagonal Architecture** (Ports & Adapters) in Go.

## DDD Implementation Guide

For detailed implementation patterns, **use the tactical-ddd skill** at `.claude/skills/tactical-ddd`. It covers:
- Domain entities, aggregates, events, repositories
- Application commands and queries (CQRS)
- Driven and driving adapters
- Testing patterns and examples
- Step-by-step workflows for new features

**Always invoke the skill when implementing new features or refactoring.**

## Development Commands

```bash
# Testing
go test ./...                                    # Run all tests
go test -v ./internal/order/domain/...           # Verbose, specific package
go test -run TestName ./path/to/package/        # Specific test
go test -cover ./...                            # With coverage

# Building
go build -o bin/api ./cmd/api
go mod tidy

# Code quality
go fmt ./...
go vet ./...
``

**Server**: `./bin/api` runs on http://localhost:8080`

## Architecture Overview

**Dependency Rule**: `Adapter → Application → Domain` (never reverse)

```
internal/{bounded-context}/
├── domain/           # Pure business logic, repository interfaces (ports)
├── application/
│   ├── command/     # Write operations (CQRS)
│   └── query/       # Read operations (CQRS)
└── adapter/
    ├── driven/      # Outbound: repositories, external services
    └── driving/     # Inbound: HTTP handlers, event consumers

pkg/                  # Shared infrastructure (eventbus, mailer)
```

## Bounded Contexts

> **Note**: Each module represents a **subdomain** (problem space) implemented as a **bounded context** (solution space). Subdomains are business segments (Core, Supporting, Generic), while bounded contexts are the technical boundaries where a ubiquitous language applies. In this project, there's a 1:1 mapping between them.

**IMPORTANT**: Each bounded context has an `agent.md` file (`internal/{context}/agent.md`) describing the module's concepts, entities, commands, events, and code structure. **Always read the agent.md before modifying a bounded context.**


## Naming Conventions

| Concept | Convention | Example |
|---------|------------|---------|
| Aggregate Root | Singular noun | `Order` |
| Internal Entity | Singular noun | `OrderItem`, `PaymentDetails` |
| Constructor | `New` prefix | `NewOrder()` |
| Error | `Err` prefix | `ErrEmptyCart` |
| Command | Verb + Noun + "Command" | `PlaceOrderCommand` |
| Event struct | Past tense + "Event" | `OrderPlacedEvent` |
| Event type (string) | Past tense | `"OrderPlaced"` |
| Repository | AggregateRoot + "Repository" | `OrderRepository` |
| Files | snake_case | `order.go` |

## Key Guidelines

1. **Always ask questions** - clarify requirements before coding
2. **Start with domain** - aggregates, value objects, business rules first
3. **Domain stays pure** - no external dependencies in domain layer
4. **Events after persistence** - emit domain events AFTER successful save
5. **Test with in-memory** - use `adapter/driven/inmemory/` and `pkg/eventbus/testing/EventCollector`
6. **One use case per file** - commands in `command/`, queries in `query/`
7. **Validate in constructors** - business rules in `New*()` or `Validate()` methods

## Domain Concepts

See each bounded context's `agent.md` for detailed ubiquitous language and user journeys.
