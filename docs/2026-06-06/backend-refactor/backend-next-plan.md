# Backend Next Plan

Last updated: 2026-06-06

This document defines the planned `backend-next` path. It is a planning document only. Do not create or switch runtime traffic to `backend-next` until a specific implementation task is approved.

## 1. Why Backend Next Exists

The current backend is playable and valuable, but it mixes several concerns in one runtime shape:

- HTTP compatibility routes.
- Mutable in-memory game state.
- Business use cases.
- Economic formulas.
- Optional persistence.
- Scheduler-owned background work.
- Compatibility stubs for future game systems.

`backend-next` exists to design a cleaner backend beside the current one without breaking the playable game. The old backend remains the source of truth until individual domains are proven safe to migrate.

## 2. Migration Philosophy

The migration path is:

```txt
Current backend A
  -> prototypes and contracts
  -> backend-next B
  -> bridge routing
  -> domain-by-domain migration
```

Rules:

- Prototype before coding.
- Build new domains behind explicit interfaces.
- Route traffic through a bridge, not direct frontend rewrites.
- Prefer read-only migration before command/write migration.
- Keep old backend rollback available until a domain is fully migrated.
- Use shadow mode for risky domains before switching live behavior.

## 3. Proposed Repository Shape

When implementation begins, create:

```txt
backend-next/
  cmd/
    simapi-next/
      main.go
  internal/
    app/
      app.go
      lifecycle.go
    httpapi/
      router.go
      middleware.go
      response.go
      dto/
    domain/
      auth/
      company/
      production/
      market/
      finance/
      research/
      social/
      system/
    service/
      usecase.go
    storage/
      memory/
      postgres/
      migration/
    bridge/
      client.go
      compare.go
    config/
      config.go
    platform/
      clock.go
      ids.go
      logger.go
```

This shape is a starting target, not a final law. If a package has no real code yet, do not create it just to satisfy the tree.

## 4. Boundary Principles

### `httpapi`

Owns:

- Request parsing.
- Response DTOs.
- Error response shape.
- Auth middleware wiring.
- Route registration.

Does not own:

- Gameplay mutation.
- Database queries.
- Economic formulas.

### `domain`

Owns:

- Domain entities.
- Domain services.
- Domain validation.
- Domain errors.

Does not own:

- HTTP request objects.
- SQL rows.
- Frontend-specific response maps.

### `service`

Owns:

- Application use cases.
- Transaction boundaries.
- Cross-domain orchestration.

Does not own:

- Low-level SQL.
- Route parsing.
- Background loop timing.

### `storage`

Owns:

- Persistence interfaces.
- Memory adapter.
- PostgreSQL adapter.
- Snapshot/version conversion.

Does not own:

- Gameplay rules.
- Economic calculation decisions.

### `platform`

Owns:

- Clock abstraction.
- ID generation.
- Logging adapters.
- Runtime helpers.

## 5. First Implementation Order

Do not start with market matching. Start with low-risk routes and infrastructure.

Recommended order:

1. `system`: health, ready, version, config echo.
2. `auth`: token validation and dev login compatibility.
3. `catalog`: static resources/buildings/economy data.
4. `company`: read-only company profile and inventory.
5. `production`: read-only jobs and options.
6. `production`: start/cancel/claim commands.
7. `finance`: read-only statements and bonds.
8. `market`: read-only ticker, resource info, orderbook, depth.
9. `market`: order create/cancel/take.
10. `scheduler`: ticks, bots, cleanup.
11. `storage`: durable save/load and migration.

## 6. Shared Contract Strategy

During migration, the old frontend should not need to know whether a route is served by A or B.

Contracts should be captured in:

```txt
docs/api-contract.md
docs/backend-refactor/api-route-inventory.md
docs/backend-refactor/prototypes/*.pseudo.md
```

For any route migrated to `backend-next`, record:

- Route path.
- Request DTO.
- Response DTO.
- Error shape.
- Auth requirement.
- Old backend behavior.
- New backend behavior.
- Compatibility gaps.

## 7. Data Strategy

Early `backend-next` should support memory mode first.

Persistence should come later because it depends on settled domain boundaries.

Initial data sources:

- Static data from `decompiled/data/*.json`.
- Config from `backend/configs/game.json` or a copied compatible config.
- Dev bootstrap data compatible with old `dev` / `dev` flow.

Later data sources:

- Versioned snapshots.
- PostgreSQL adapter.
- Migration scripts.

## 8. Success Criteria

`backend-next` is useful when it can:

- Start independently on a separate port.
- Answer health/readiness routes.
- Load the same static resource/building data as the old backend.
- Validate old tokens or issue compatible dev tokens.
- Serve one read-only domain behind bridge routing.
- Run comparison tests against old backend responses.

`backend-next` is ready for real migration when it can:

- Pass domain tests.
- Pass handler contract tests.
- Run in shadow mode without user-visible effects.
- Produce comparable responses for a chosen domain.
- Roll back to old backend with one route flag change.

