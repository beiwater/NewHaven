# New Haven Backend Refactor Constitution

Last updated: 2026-06-06

This document is the planning constitution for the New Haven backend refactor cycle. It defines the refactor goals, boundaries, technical standards, allowed libraries, recommended versions, and the phased plan. It is intentionally a plan only: no runtime behavior should change until a specific implementation task is approved.

Supporting planning packet: `docs/backend-refactor/README.md`

Migration strategy note: the preferred long-form path is prototype first, then build `backend-next`, then migrate domains through bridge routing with `old`, `next`, and `shadow` modes.

## 1. Refactor Mission

New Haven is entering a backend-first refactor cycle. The target is not to rewrite the game; the target is to make the current Go backend easier to extend, test, persist, and operate while preserving the existing gameplay contract.

Primary goals:

- Keep the browser game playable throughout the refactor.
- Preserve API compatibility unless a versioned replacement is explicitly documented.
- Make business logic testable without HTTP, schedulers, or storage.
- Make persistence optional but cleanly isolated.
- Keep economic formulas pure and deterministic.
- Prepare the backend for larger multiplayer state, market ticks, bots, anti-cheat, and future production systems.

Non-goals for the first backend cycle:

- No full framework migration.
- No frontend redesign.
- No database-only runtime requirement.
- No speculative microservices split.
- No replacement of all in-memory state in one pass.

## 2. Current Baseline

Backend:

- Go module: `go-sim-api`
- Go version in `backend/go.mod`: `1.25.0`
- HTTP stack: standard library `net/http`
- Storage: memory-first, optional PostgreSQL
- PostgreSQL driver: `github.com/jackc/pgx/v5 v5.9.2`
- Auth: JWT-style bearer token handled by local middleware
- Password hashing: `golang.org/x/crypto`
- State model: `model.GameState` guarded by `sync.Mutex` through `service.Service`

Frontend compatibility baseline:

- React: `^19.2.6`
- TypeScript: `~6.0.2`
- Vite: `^8.0.12`
- TanStack Query: `^5.100.14`
- PixiJS: `^8.18.1`
- Zustand: `^5.0.14`

The refactor must respect the current client API expectations until replacement contracts exist in `docs/api-contract.md`.

## 3. Project Constitution

### 3.1 Compatibility First

Every backend refactor must answer: "Can the current frontend still log in, load company state, render the map, start production, trade, and refresh market data?"

Breaking API changes require:

- A documented replacement endpoint.
- A migration note in `docs/api-contract.md`.
- A frontend compatibility plan.
- Tests covering the old and new behavior when both remain supported.

### 3.2 State Has One Owner

Mutable game state must remain owned by the service layer. Handlers, schedulers, storage adapters, and formula packages must not mutate shared state directly.

Rules:

- Public service methods own locking.
- Internal helpers that require a lock should keep a clear `Locked` suffix.
- Formula packages must remain side-effect free.
- Schedulers call service methods; they do not reach into `GameState`.
- Storage receives snapshots or explicit persistence DTOs; it does not own gameplay decisions.

### 3.3 Domain Boundaries Beat File Count

The backend should move toward domain packages where each domain has a clear API, tests, and ownership.

Initial domains:

- `auth`: login, password verification, token issuing, dev account bootstrapping.
- `company`: company profile, inventory, buildings, level unlocks.
- `production`: production jobs, claims, recipes, resource conversion.
- `market`: orders, matching, depth, trades, bot liquidity, ticker updates.
- `finance`: bonds, debt, accounting, government hooks.
- `research`: research projects, buffs, unlocks, executive effects.
- `social`: messages, chat, leaderboard-facing identity.
- `system`: health, config, save/load, snapshots, clock.

The first cycle may keep package paths stable while improving internal shape. Physical package moves should be done only when tests and API callers make the migration safe.

### 3.4 Tests Are the Refactor Harness

Refactoring without tests is not a refactor for this project. Each meaningful backend change should either preserve an existing test or add a focused test around the behavior being moved.

Required test layers:

- Formula tests for pure economic functions.
- Service tests for gameplay behavior.
- Handler tests for request/response compatibility.
- Storage tests for PostgreSQL adapter behavior where practical.
- Regression tests for bugs found during the refactor.

Preferred command before merging backend changes:

```bash
cd backend
go test ./...
go vet ./...
```

### 3.5 Determinism Over Convenience

Economic simulation code should be deterministic by default.

Rules:

- Time should be injectable in business logic that depends on `time.Now()`.
- Randomness should be injectable or isolated behind a small interface.
- IDs should be generated through one helper or interface per domain.
- Tests should not depend on wall-clock timing unless testing time parsing or scheduler wiring.

### 3.6 Persistence Is an Adapter, Not the Model

PostgreSQL should support the game model; it should not dictate gameplay structs.

Rules:

- Keep `model` free from database tags unless there is a strong reason.
- Storage adapters convert between rows and model/snapshot types.
- Save/load formats should be versioned.
- Introduce migrations before relying on durable PostgreSQL schemas.

### 3.7 Keep the Backend Boring

The backend should remain simple, explicit, and debuggable. Prefer standard library solutions unless a library clearly removes risk or repeated infrastructure code.

Avoid:

- Heavy web frameworks before endpoint shape stabilizes.
- Global registries.
- Hidden goroutine ownership.
- Reflection-heavy validation in core gameplay paths.
- Large dependency additions without a written reason.

## 4. Recommended Versions

These are planning recommendations, not an instruction to upgrade immediately.

| Area | Recommended version | Reason |
|------|---------------------|--------|
| Go | Keep `1.25.x` for this cycle | Matches current `go.mod`; avoids toolchain churn during structural refactor. |
| PostgreSQL | `17.x` preferred, `16.x` acceptable | Modern enough for JSONB, generated columns, and good local tooling. |
| pgx | Keep `github.com/jackc/pgx/v5 v5.9.x` | Already adopted; excellent low-level PostgreSQL driver. |
| Node | `22.x LTS` | Matches project tooling preference and Vite/React generation. |
| React | Keep `19.x` | Current client baseline; backend API should not force client churn. |
| TypeScript | Keep `6.0.x` | Current client strict-mode baseline. |
| Vite | Keep `8.x` | Current dev server and proxy baseline. |

Version policy:

- Pin backend libraries in `go.mod`.
- Do not upgrade language/toolchain versions in the same PR as domain refactors.
- Upgrade dependencies in small, named PRs with test output attached.
- Keep one dependency family per upgrade PR where possible.

## 5. Allowed Backend Libraries

### 5.1 Recommended Now

Use these when they solve an immediate backend refactor need.

| Library | Recommended use | Version guidance |
|---------|-----------------|------------------|
| `github.com/jackc/pgx/v5` | PostgreSQL driver, connection pool, transactions | Keep current `v5.9.x`. |
| `golang.org/x/crypto` | Bcrypt/password hashing | Keep current module range from `go.mod`. |
| `golang.org/x/sync` | Errgroup or concurrency helpers when standard library is awkward | Keep current module range from `go.mod`. |
| Standard library `log/slog` | Structured server logs | Prefer over adding logging frameworks. |
| Standard library `net/http` | HTTP routing and middleware | Keep unless routing complexity becomes a real bottleneck. |
| Standard library `testing`, `httptest` | Unit and handler tests | Default test stack. |

### 5.2 Good Candidates When Needed

Add only when there is a concrete task that benefits from the dependency.

| Library | Use case | Recommendation |
|---------|----------|----------------|
| `github.com/golang-jwt/jwt/v5` | JWT creation and validation | Consider replacing local token parsing if auth grows. |
| `github.com/go-chi/chi/v5` | Lightweight routing and route groups | Consider only after route registration becomes hard to maintain. |
| `github.com/pressly/goose/v3` | SQL migrations | Good first migration tool if PostgreSQL persistence becomes first-class. |
| `github.com/stretchr/testify` | Assertions and require helpers | Acceptable for tests, but not required. |
| `github.com/google/uuid` | UUIDs for stable external IDs | Use if timestamp IDs become collision-prone or externally exposed. |
| `github.com/caarlos0/env/v11` | Environment config parsing | Use if config parsing expands beyond the current simple loader. |
| `github.com/rs/cors` | CORS middleware | Use only if standard middleware becomes noisy. |

### 5.3 Avoid for Now

Do not introduce these during the first backend refactor cycle without a separate architecture decision.

| Category | Examples | Why avoid now |
|----------|----------|---------------|
| Full web frameworks | Gin, Fiber, Echo | Adds migration cost before endpoint design has stabilized. |
| ORMs | GORM, Ent | Persistence model is still evolving; row adapters are clearer for now. |
| Message brokers | NATS, Kafka, RabbitMQ | The game is not yet split into services. |
| Distributed cache | Redis | In-memory state ownership is still being clarified. |
| Code-generation-first APIs | OpenAPI server generation, GraphQL generation | Useful later, but premature before contracts settle. |

## 6. Target Backend Shape

Desired dependency direction:

```txt
cmd/simapi
  -> handler
  -> service
  -> domain logic / formula
  -> model

scheduler
  -> service

storage
  -> model snapshots / persistence DTOs

formula
  -> no project side effects
```

The service layer may initially remain a single `Service` type, but its internals should become smaller and domain-oriented.

Preferred long-term shape:

```txt
backend/internal/
  app/          Wiring, lifecycle, server dependencies
  handler/      HTTP DTOs, request parsing, response writing
  service/      Use cases and transaction boundaries
  domain/       Domain rules that are not HTTP/storage specific
  formula/      Pure economy math
  model/        Shared domain data structures
  storage/      Memory/PostgreSQL adapters
  scheduler/    Background loops calling service APIs
  config/       Env and game tuning config
```

This shape is a destination, not a required first PR.

## 7. Refactor Phases

### Phase 0: Safety Snapshot

Purpose: make sure the project can be refactored without guessing.

Deliverables:

- Confirm `go test ./...` baseline.
- Document failing tests if any.
- Add a short backend smoke checklist.
- Identify API endpoints used by the frontend.
- Record the current dev login flow.

No production behavior changes.

### Phase 1: Service Boundary Cleanup

Purpose: reduce the risk of the large `service.Service` becoming the permanent dumping ground.

Deliverables:

- Group service methods by domain.
- Normalize public method naming.
- Keep lock ownership consistent.
- Extract repeated lookup, validation, and ID helpers.
- Add tests before moving complex behavior.

Suggested order:

1. Company and inventory.
2. Production and recipe.
3. Market and order matching.
4. Research and executives.
5. Finance and bonds.
6. Auth and player bootstrap.

### Phase 2: Time, IDs, and Determinism

Purpose: make game behavior easier to test and replay.

Deliverables:

- Introduce a small clock abstraction for service logic.
- Centralize ID generation.
- Remove direct `time.Now()` calls from core use cases where practical.
- Add tests for production timing, market matching, research completion, and bond maturity.

### Phase 3: Handler Contract Hardening

Purpose: make HTTP behavior explicit and stable.

Deliverables:

- Define request/response DTOs near handlers.
- Make error responses consistent.
- Add handler tests for auth, company, production, market, and research flows.
- Update `docs/api-contract.md` when contracts are clarified.

### Phase 4: Persistence Boundary

Purpose: make memory and PostgreSQL modes share the same gameplay semantics.

Deliverables:

- Version save snapshots.
- Clarify what is persisted and what is recomputed.
- Add migration tooling if PostgreSQL becomes durable project infrastructure.
- Add storage tests for load/save round trips.
- Keep memory-only dev mode working.

### Phase 5: Scheduler and Background Work

Purpose: keep background loops observable and safe.

Deliverables:

- Make scheduler lifecycle explicit.
- Ensure goroutines stop on shutdown.
- Keep tick intervals configurable.
- Add structured logs around market matching and bot replenishment.
- Add tests for scheduler-owned timing where feasible.

### Phase 6: Observability and Operations

Purpose: make local debugging and future deployment less painful.

Deliverables:

- Use `log/slog` consistently.
- Add request IDs if needed.
- Expand health endpoint to report build/config basics.
- Add lightweight metrics only after the runtime shape stabilizes.

## 8. Backend Refactor Checklist

Before each backend refactor PR:

- State the domain being changed.
- State whether API behavior changes.
- Run `go test ./...`.
- Run `go vet ./...` when practical.
- Include new tests for moved or clarified behavior.
- Avoid dependency upgrades unless the PR is specifically for that.
- Avoid frontend changes unless needed to preserve compatibility.

Before merging each phase:

- Dev login `dev` / `dev` still works.
- Company state loads.
- Existing buildings render from API data.
- Production can start and claim.
- Market buy/sell flow still works.
- No scheduler goroutine leaks during shutdown in normal dev runs.

## 9. Architecture Decision Record Template

Use this short format for decisions that affect long-term backend shape.

```md
# ADR: <decision title>

Date: YYYY-MM-DD
Status: Proposed | Accepted | Rejected | Superseded

## Context

What pressure or problem forced this decision?

## Decision

What are we choosing?

## Consequences

What gets easier, what gets harder, and what must be watched?

## Alternatives Considered

What did we reject, and why?
```

Recommended location:

```txt
docs/adr/YYYY-MM-DD-short-title.md
```

## 10. First Planning Backlog

This is the suggested non-execution backlog for the beginning of the backend refactor cycle.

1. Capture the current backend test baseline.
2. Map frontend API usage to backend handlers.
3. Create a service method inventory by domain.
4. Identify all direct time and ID generation points.
5. Identify all direct `GameState` mutation points.
6. Decide whether to introduce a clock interface in Phase 2.
7. Decide whether SQL migrations are needed before persistence work.
8. Update `docs/api-contract.md` only after contract decisions are made.

## 11. Definition of Done for the Refactor Cycle

The backend refactor cycle is done when:

- Core gameplay behavior is covered by service tests.
- Handler contracts for active frontend flows are documented and tested.
- Time and ID generation are deterministic in core use cases.
- Persistence boundaries are explicit.
- Background schedulers have clear lifecycle ownership.
- The current frontend can run against the refactored backend without special compatibility hacks.
- Future gameplay features can be added inside a domain without touching unrelated domains.
