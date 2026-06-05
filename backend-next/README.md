# Backend Next

This is the next-generation backend for NewHaven, built alongside the current `backend/`.

## Guiding Principles

- **Read-only first**: Each domain starts with read-only routes, then write routes, then scheduler
- **Bridge coexists**: Old backend remains the source of truth until a domain is proven safe to migrate
- **Eventual OpenAPI**: All endpoints defined in `openapi/openapi-draft.yaml` before implementation
- **Typed everywhere**: No `map[string]any` returns
- **Test-first**: Domain logic is pure and independently testable

## Architecture

```
cmd/simapi-next/    Entry point (chi router)
internal/
  app/              Application use cases (orchestration)
  httpapi/          HTTP layer (router, middleware, DTOs)
  domain/           Domain entities + rules (no IO)
  storage/          Storage interfaces + implementations
  bridge/           Old-backend proxy for shadow comparison
  config/           Configuration
  platform/         Cross-cutting utilities (clock, idgen, logger)
```

## Getting Started

```bash
cd backend-next
go mod tidy
go run ./cmd/simapi-next/
```

The server starts on `:8088` with health endpoints at `/healthz` and `/readyz`.

## Status

This is a scaffold. Each domain will be populated incrementally:
1. Health → Auth → Catalog → Company (read) → Production (read) → Market (read) → ...

## Tech Stack

| Layer | Choice |
|-------|--------|
| Router | go-chi/chi v5 |
| API Contract | OpenAPI 3.1 + oapi-codegen |
| DB | pgx/v5 |
| SQL | sqlc |
| Migration | goose |
| Realtime | coder/websocket |
| Validation | go-playground/validator |
