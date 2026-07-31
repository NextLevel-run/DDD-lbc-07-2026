# CLAUDE.md

## Monorepo Structure

- `apps/api/` — Go backend (DDD + Hexagonal Architecture). See `apps/api/CLAUDE.md` for full guide.
- `apps/web/` — Frontend (placeholder).

## Development

```bash
# Backend
cd apps/api && go test ./...
cd apps/api && go build -o bin/api ./cmd/api

# Frontend
cd apps/web && ...
```
