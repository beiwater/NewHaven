# Post-Phase 15 Review and Abstract Roadmap

Date: 2026-06-06
Status: Review checkpoint; no new implementation phase started
Baseline commit: `96210ba` (`next-phase 15`)

## 1. Executive Conclusion

The backend rewrite has successfully validated its target delivery model:

- chi routing and middleware
- OpenAPI contract-first DTO generation
- app service orchestration
- domain-specific storage interfaces
- in-memory storage for development and tests
- injectable clock and ID generation
- focused service and HTTP tests
- one Git checkpoint per completed phase

The migrated core economy loop is functional:

```text
auth
-> company/assets read
-> production start and claim
-> inventory mutation
-> market order, take, and matching
-> finance views and ledger effects
-> bond issue
```

This proves that `backend-next` is a viable replacement architecture.

It does not yet prove that the legacy backend can be removed. The remaining
work contains the most operationally sensitive areas: scheduler behavior,
economic formulas, persistence, frontend cutover, and several gameplay
domains.

### Deferred extension plugins

The following capability groups are optional extension plugins and are not
required implementations for the core backend rewrite:

- anti-cheat and AML
- aerospace
- executives
- research
- government contracts

The core rewrite must reserve stable integration ports, optional API
namespaces, configuration and registration boundaries, and failure isolation
for these plugins. A missing or disabled plugin must not prevent core startup
or operation.

## 2. Stability Review

### Stable foundations

The following foundations are stable enough to continue building on:

| Foundation | Evidence | Verdict |
|---|---|---|
| HTTP architecture | Consistent chi router, auth middleware, response envelope, and handler pattern | Stable |
| Contract workflow | OpenAPI draft generates server/types successfully and is used by migrated endpoints | Stable |
| Application boundaries | Services receive storage, clock, ID generator, config, and catalogs through constructors | Stable per service; composition is split between `app.App` and `main.go` |
| Storage abstraction | Eight domain-oriented storage interfaces have a complete memory implementation | Stable for development |
| Core economic writes | Production, market, finance, and bond issue paths have focused rollback and state tests | Stable within current scope |
| Verification baseline | `go test ./...`, `go vet ./...`, and build pass at the Phase 15 checkpoint | Stable |
| Delivery discipline | Phase 1 through Phase 15 are independently committed and documented | Stable |

### Structural gaps

| Gap | Evidence | Risk |
|---|---|---|
| Domain packages mostly contain data types rather than reusable business rules | `backend-next/internal/domain/*/types.go`; large app services such as market, production, and finance | High growth risk |
| backend-next has no governed formula package; economic formulas are currently re-derived inside services | Legacy `backend/internal/formula/`; local calculations in backend-next production, market, and finance services | High parity risk |
| Scheduler and background orchestration do not exist in backend-next | Legacy `backend/internal/scheduler/scheduler.go` owns a nine-step tick | Critical |
| PostgreSQL implementation and migrations are not implemented | `backend-next/internal/storage/postgres/` is empty; migration folder has only a README | Critical |
| Shadow comparison exists only as a utility, not a trusted validation system | `backend-next/internal/bridge/`; shape comparison is one-directional and semantic comparison currently falls through as an unconditional pass | High |
| Runtime composition has two assembly points | `app.App` wires auth while `cmd/simapi-next/main.go` constructs the other services | Medium consistency risk |
| Race verification is not currently runnable in this environment | `go test -race ./...` requires CGO and could not run | Medium verification gap |
| Error classification remains handler-specific | Services return general errors and handlers manually map status codes | Medium consistency risk |

## 3. Current Coverage

The route counts are useful only as a rough indicator because several legacy
registrations multiplex multiple methods and path shapes.

```text
legacy production route registrations: 106
backend-next route registrations:     28
frontend API call sites observed:      58
```

The migrated routes cover a disproportionately important part of the game:
authentication, core reads, production writes, market writes and matching,
finance reads, and bond issue. However, route parity and operational parity
remain incomplete.

## 4. Responsibilities Still Owned by Legacy

### Gameplay responsibilities

- building market, purchase, placement, movement, demolition, and upgrades
- production queue management, cancellation, slot management, and claim-all
- player progression, achievements, boosts, and offline income
- remaining company, auction, leaderboard, social, newspaper, and daily order
  behavior
- remaining market, bond, and finance lifecycle behavior
- optional extension-plugin behavior: anti-cheat/AML, aerospace, executives,
  research, and government contracts

### Operational responsibilities

- scheduler tick and its ordering contract
- bond settlement and default processing
- optional government-contract awards and defaults
- bot market cycle and market lock behavior
- order cleanup and daily refresh
- persistent save and restore
- current PostgreSQL snapshot persistence
- runtime behavior still expected by the frontend

## 5. Abstract Future Migration Batches

These are capability batches, not concrete implementation phases. A batch
should begin only after its entry gate is satisfied and should not be treated
as complete until its exit gate is met.

### Batch A: Architecture Consolidation

Purpose:

- establish a governed backend-next formula boundary and one source of truth for
  economic formulas and static data
- move reusable business rules out of oversized app services where valuable
- standardize typed application errors and transaction expectations
- repair and strengthen compatibility comparison tooling

Entry gate:

- Phase 15 remains green and frozen as the behavioral baseline

Exit gate:

- formula ownership, domain/app boundaries, runtime composition, error
  classification, and comparison levels are documented and enforceable by tests

### Batch B: Bounded Gameplay Completion

Purpose:

- migrate remaining gameplay domains that can operate without the global
  scheduler
- prioritize domains already represented by domain types or storage interfaces
- keep each domain independently testable

Entry gate:

- Batch A decisions are available for new services to follow

Exit gate:

- all player-triggered production routes required by the frontend have a
  backend-next owner, or are explicitly retired

### Batch C: High-Risk Economic Writes

Purpose:

- complete remaining building, market, finance, bond, and progression write
  paths
- define atomicity and rollback behavior for cross-domain mutations

Entry gate:

- formula parity and error/transaction policy are established

Exit gate:

- every money, inventory, XP, ledger, and ownership mutation has invariant,
  rollback, and concurrency coverage

### Batch D: Scheduler and Background Runtime

Purpose:

- define typed background jobs and preserve the legacy tick ordering contract
- migrate settlement, bot behavior, cleanup, refresh, and default processing
- make jobs replayable and observable

Entry gate:

- all core domains called by the scheduler are owned by backend-next; optional
  plugin jobs are isolated behind extension ports

Exit gate:

- deterministic replay from the same starting state produces approved economic
  results, and job failures are observable and recoverable

### Batch E: Persistence and Data Migration

Purpose:

- implement PostgreSQL-backed storage and explicit transaction boundaries
- introduce versioned schema migrations
- migrate legacy snapshot data without silent loss

Entry gate:

- domain ownership and mutation contracts are stable enough to define schema

Exit gate:

- memory and PostgreSQL implementations pass the same contract tests
- restart, migration, rollback, and backup/restore tests pass

### Batch F: Frontend and Operational Cutover

Purpose:

- prove response and behavior compatibility
- move frontend workflows to backend-next in controlled groups
- observe production behavior before retiring legacy

Entry gate:

- backend-next owns all required runtime responsibilities and persistent data

Exit gate:

- all supported frontend workflows use backend-next
- shadow comparison and operational metrics show approved parity
- rollback procedure has been exercised
- legacy runtime can be disabled without loss of functionality

## 6. Dependency Order

```text
Architecture consolidation
        |
        v
Bounded gameplay completion
        |
        v
High-risk economic writes
        |
        v
Scheduler/background runtime
        |
        v
Persistence and data migration
        |
        v
Frontend and operational cutover
```

Limited read-only frontend cutover experiments may happen earlier, but final
cutover must remain after scheduler and persistence parity.

## 7. Definition of Rewrite Complete

The rewrite is complete only when all of the following are true:

### Ownership

- every supported production route and frontend workflow has an explicit
  backend-next owner
- legacy-only routes are either migrated or formally retired
- no required gameplay or operational behavior depends on the legacy service

### Contract and behavior parity

- OpenAPI describes every supported endpoint
- status codes, response shapes, and error semantics have approved parity
- critical write paths pass happy-path, edge-case, failure, and rollback tests
- economic formulas and scheduler outcomes have deterministic parity fixtures

### Data safety

- PostgreSQL implements all required storage contracts
- migrations are versioned, repeatable, and reversible where required
- restart, backup/restore, legacy import, and rollback preserve verified state
- money, inventory, XP, ownership, orders, trades, ledger, bonds, messages, and
  progression survive restart
- enabled plugins declare and verify their own persistence requirements

### Runtime safety

- scheduler jobs are observable, replayable, and failure-aware
- concurrency-sensitive paths pass race testing in a CGO-enabled environment
- health and readiness checks reflect real dependencies
- core rate limiting and security responsibilities are implemented
- anti-cheat and AML are represented by documented optional plugin ports and do
  not block core startup when absent

### Cutover

- the frontend builds and all supported workflows pass against backend-next
- shadow comparison and operational monitoring meet agreed thresholds
- rollback has been tested
- the legacy backend can be disabled and later removed

## 8. Current Verdict

```text
Architecture viability: confirmed
Core economy migration: confirmed
Production readiness: not confirmed
Legacy removal readiness: not confirmed
```

The next implementation work should not begin by choosing another endpoint.
It should begin by approving the architecture consolidation decisions and the
completion gates in this document.
