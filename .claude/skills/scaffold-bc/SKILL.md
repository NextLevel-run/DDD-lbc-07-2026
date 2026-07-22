---
name: scaffold-bc
description: This skill should be used when users want to create a new bounded context directory structure following DDD and Hexagonal Architecture patterns. Use when users ask to "create a bounded context", "scaffold a domain", or "setup a new module structure".
---

# Scaffold Bounded Context

This skill creates the standard directory structure for a bounded context following Domain-Driven Design (DDD) and Hexagonal Architecture patterns.

## When to Use

Use this skill when:
- User asks to create a new bounded context
- User wants to scaffold a new domain module
- User requests to setup the structure for a new feature domain
- User mentions creating a new DDD module

## How to Use

Execute the bundled shell script to create the complete directory structure:

```bash
bash .claude/skills/scaffold-bc/scripts/scaffold-bc.sh <bounded-context-name>
```

The script automatically generates a minimal `agent.md` with complete template structure.

**Example:**
```bash
bash .claude/skills/scaffold-bc/scripts/scaffold-bc.sh user-management
```

This will create:

```
internal/<bounded-context-name>/
├── agent.md                    # Auto-generated minimal documentation with template
├── domain/                     # Pure business logic (empty)
├── application/
│   ├── command/
│   │   └── .gitkeep
│   └── query/
│       └── .gitkeep
└── adapter/
    ├── driven/                 # Outbound adapters
    │   ├── inmemory/           # In-memory implementations (empty)
    │   └── postgres/           # PostgreSQL implementations (empty)
    └── driving/                # Inbound adapters
        ├── http/               # HTTP handlers (empty)
        └── consumer/           # Event consumers
            └── .gitkeep
```

The generated `agent.md` includes:
- ✅ Title with emoji and bounded context name
- ✅ Complete template structure with all sections (Ubiquitous Language, User Journeys, Architecture, Domain, Application, Adapters, Dependencies)
- ✅ "To be implemented" placeholders ready to fill during development
- ✅ Folder structure tree with correct paths
- ✅ Table templates for Aggregates, Events, Commands, Queries, and Adapters

Simply edit the generated `agent.md` as you implement the bounded context, replacing "To be implemented" sections with actual content.

## Naming Conventions

- Bounded context directory: `kebab-case` (e.g., `user-management`, `classified-ad`)
- Files: `snake_case` (e.g., `user_account.go`)
- Go packages: `lowercase` (e.g., `package domain`)

## Notes

- All `.go` files should be created manually based on domain requirements
- The `agent.md` file provides a minimal structure overview
- `.gitkeep` files ensure empty directories are tracked in git
- The structure is framework-agnostic and ready for Go implementation
