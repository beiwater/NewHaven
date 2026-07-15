# New Haven Backend

This is the canonical Go API for New Haven. It runs with in-memory storage by default, so no database is required for local development.

## Guiding Principles

- **Single backend**: New backend work belongs here; there is no parallel legacy tree
- **Contract first**: Define endpoints in `openapi/openapi-draft.yaml`
- **Typed everywhere**: No `map[string]any` returns
- **Test-first**: Domain logic is pure and independently testable

## Architecture

```
cmd/simapi/    Entry point (chi router)
internal/
  app/              Application use cases (orchestration)
  httpapi/          HTTP layer (router, middleware, DTOs)
  domain/           Domain entities + rules (no IO)
  storage/          Storage interfaces + implementations
  bridge/           Optional HTTP comparison helpers
  config/           Configuration
  platform/         Cross-cutting utilities (clock, idgen, logger)
```

## Getting Started

```bash
cd backend
go run ./cmd/simapi/
```

The server starts on `:8088` with health endpoints at `/healthz` and `/readyz`.

Development mode creates a `dev` user. The default password is `123`; override it with `SIM_API_DEV_PASSWORD`.

## Tech Stack

| Layer | Choice |
|-------|--------|
| Router | go-chi/chi v5 |
| API Contract | OpenAPI 3.1 + oapi-codegen |
| DB | pgx/v5 |
| Persistence | Memory by default; optional PostgreSQL snapshots |
