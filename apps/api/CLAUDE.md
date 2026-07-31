# CLAUDE.md

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
go test -v ./internal/classified-ad/domain/...   # Verbose, specific package
go test -run TestName ./path/to/package/        # Specific test
go test -cover ./...                            # With coverage

# Building
go build -o bin/api ./cmd/api
go mod tidy

# Code quality
go fmt ./...
go vet ./...
```

**Server**: `./bin/api` runs on http://localhost:8080

## Bounded Contexts

**IMPORTANT**: Each bounded context has an `agent.md` file (`internal/{context}/agent.md`) describing the module's concepts, entities, commands, events, and code structure. **Always read the agent.md before modifying a bounded context.**

## Key Guidelines

1. **Always ask questions** - clarify requirements before coding
2. **Test with in-memory** - use `adapter/driven/inmemory/` and `pkg/eventbus/testing/EventCollector`
3. **One use case per file** - commands in `command/`, queries in `query/`
4. **Validate in constructors** - business rules in `New*()` or `Validate()` methods

## Domain Concepts

See each bounded context's `agent.md` for detailed ubiquitous language and user journeys.
