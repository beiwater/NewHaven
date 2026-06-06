# Phase 16: Compatibility and Completion Baseline (v3)

**Date**: 2026-06-06
**Status**: Complete
**Baseline commit**: `eabb962`
**Phase 15 commit**: `96210ba`
**Validator**: `backend-next/scripts/phase16-inventory-validator.ps1`
**CSV location**: `docs/2026-06-06/rebuild/phase-16/`

## 1. Executive Summary

Phase 16 produced a trusted inventory of every legacy production HTTP route registration, backend-next route, frontend API callsite, scheduler responsibility, and persistent data group. All routes are classified with approved dispositions. The inventory is machine-readable (6 CSV files) with enriched schemas (legacy_owner, backend_next_owner, verification_requirement, source_ref) and validated by a strengthened PowerShell script that performs actual source-drift comparisons.

**Scope note**: This inventory counts **route registrations** (explicit `mux.HandleFunc` calls in Go 1.25's `http.ServeMux`). Several legacy handlers use manual path dispatching (`strings.TrimPrefix`, `strings.Split`) to serve multiple logical sub-routes from a single registration. The registration count (106) undercounts the effective logical endpoint surface. The CSV `method_ambiguity` column records `internal-switch` for these patterns. The frontend callsites CSV (58 callsites) provides complementary workflow-level coverage.

### Key Counts

| Metric | Value | Level |
|--------|-------|-------|
| Legacy production route registrations | 106 | Registration-level |
| Backend-next route registrations | 28 | Registration-level |
| Frontend API callsites (all src/) | 58 | Callsite-level |
| Scheduler tick stages/calls | 9 stages / 10 calls | Stage-level |
| Persistent data groups | 37 | Group-level |
| Deferred extension plugins | 5 | Plugin-level |

### Disposition Summary (Registration-Level)

| Disposition | Count | Meaning |
|-------------|-------|---------|
| migrated | 24 | Backend-next ownership exists; parity verification remains required |
| remaining-core | 54 | Legacy-only core gameplay; not yet migrated |
| dev-only | 6 | Dev/debug endpoints; excluded from production target |
| retire-candidate | 2 | Redundant or unused; recommended for retirement |
| deferred-plugin | 20 | Optional plugins; excluded from core completion gate |

Total: 106 classified. Zero unclassified.

## 2. Inventory Methodology

### Extraction Method

All inventories were extracted by reading source code directly. Route registration counts are source-verified:

- **Legacy routes**: Counted all `mux.HandleFunc` calls in `backend/internal/handler/*.go` -- 17 handler files, 106 registrations.
- **Backend-next routes**: Counted all `r.Get()`, `r.Post()`, `r.Delete()` etc. calls in `backend-next/internal/httpapi/router.go` -- 28 registrations.
- **Frontend callsites**: Counted all `api.get()`, `api.post()`, `api.delete()` calls across ALL `.ts`/`.tsx` files under `client/atlas-foods-client/src/` -- 58 callsites in 14 files (13 in `src/api/`, 1 in `src/features/buildings/BuildView.tsx`).
- **Scheduler**: Read `backend/internal/scheduler/scheduler.go` -- the `tick()` function has 9 ordered stages (10 method calls, with stage 5 having two calls: 5a ResourcesWithMarket and 5b CheckMarketLock).
- **Data groups**: Read `backend/internal/model/types.go` (GameState struct), `backend-next/internal/storage/memory/memory.go` (storage interfaces), and `backend-next/internal/domain/*/types.go` (domain types). Count: 37 groups.

### Verification

The validator (`backend-next/scripts/phase16-inventory-validator.ps1`) performs:

1. **File existence**: All 6 CSVs must exist.
2. **Non-ASCII check**: All CSV files must be pure ASCII.
3. **Source drift**: Current legacy route paths, backend-next handler registrations, and frontend file-and-line callsite references are compared as exact sets against the CSV inventories.
4. **Enum validation**: Disposition, method_ambiguity, verification_requirement checked against approved lists.
5. **Empty field checks**: Every production-related row must have disposition, dependency_group, verification_requirement. A value of `none` for verification_requirement requires explanatory notes.
6. **Cross-file references**: Every `classification.csv` path must exist in `legacy-routes.csv`.
7. **Source_ref validation**: Every `frontend-callsites.csv` source_ref points to a real file.
8. **Unique keys**: No duplicate `legacy_path` or `source_ref` values.
9. **Dashboard computed from CSV**: Counts are recomputed directly from CSV data, not hard-coded.

### Known Limitations

1. **Registration-level vs. endpoint-level coverage**: This inventory operates at the Go `http.ServeMux` **registration** level. Several legacy handlers (e.g., `handleV1Buildings`, `handleV2Buildings`, `handleV2Companies`, `handleV4`, `handleAuctions`) use manual path parsing to dispatch 2-5 logical sub-endpoints from a single `mux.HandleFunc` call. The registration count of 106 undercounts the effective API surface by an estimated 40-70 logical sub-routes. The CSV `method_ambiguity` column marks these as `internal-switch`. Frontend workflow coverage (58 callsites in ~30 workflows) provides the complementary workflow-level view.

2. **Side-effect coupling**: Legacy market read routes (`/api/v3/market-ticker/`, `/api/v3/market/`, `/api/v3/market-depth/`) call `s.svc.RunBotMarketCycle()` as a side effect before returning data. Backend-next equivalents do not. This is flagged as `shape+semantics+side-effects` and parity verification MUST account for this.

3. **Method ambiguity**: Go 1.25 `http.ServeMux` does path-prefix matching. Legacy routes fall into three categories: `exact` (handler checks method explicitly), `prefix` (any method accepted; path identifies sub-resource), `internal-switch` (handler switches on method or sub-path internally).

4. **Dead interface method**: `ResetDailyMarket` is declared in the `GameService` interface but never called by the tick or any handler.

5. **Bot market cycle called from handlers**: `RunBotMarketCycle` is called from 3 market handler routes in addition to scheduler tick stage 4, creating read-vs-write coupling.

## 3. Scheduler Tick Ordering

The legacy scheduler tick executes 9 ordered stages (10 method calls) every 60 seconds:

| Stage | Order | Method | Domain | Calls | Responsible Owner |
|-------|-------|--------|--------|-------|-------------------|
| 1 | 1 | `SettleAllBonds` | Bond | 1 | legacy-only |
| 2 | 2 | `AwardGovernmentContracts` | Government | 1 | deferred-plugin |
| 3 | 3 | `ResolveGovernmentDefaults` | Government | 1 | deferred-plugin |
| 4 | 4 | `RunBotMarketCycle` | Market | 1 | migrated-to-bn |
| 5 | 5a,5b | `ResourcesWithMarket` + `CheckMarketLock` | Market | 2 | migrated-to-bn |
| 6 | 6 | `CleanupOrders` | Market | 1 | migrated-to-bn |
| 7 | 7 | `RunAllProductionJobs` | Production | 1 | legacy-only |
| 8 | 8 | `RefreshDailyOrders` | Daily Orders | 1 | legacy-only |
| 9 | 9 | `SaveAll` | Storage | 1 | legacy-only |

Stage 5 is a single logical stage with two method calls (prerequisite + loop execution).

**Dead code**: `ResetDailyMarket` in service interface, never invoked.

**Handler-side calls**: `RunBotMarketCycle` called from 3 market handler routes; `AwardGovernmentContracts` and `ResolveGovernmentDefaults` called from government admin routes.

## 4. Deferred Extension Plugins

These 5 capability groups are optional and **excluded from the core completion gate**. They are fully inventoried but do not block core route coverage:

| Plugin | Legacy Routes | Frontend Callsites | Backend-Next Status |
|--------|--------------|-------------------|---------------------|
| Research | 4 | 4 (`research.api.ts`) | ResearchStorage interface exists, no HTTP handlers |
| Executives | 6 | 5 (`executives.api.ts`) | No storage or HTTP handlers |
| Government Contracts | 5 | 2 (`contracts.api.ts`) | No storage or HTTP handlers |
| Aerospace | 5 | 0 | No storage or HTTP handlers |
| Anti-cheat / AML | 0 (in-memory only) | 0 | Not planned for scope |

Core completion excludes deferred plugins. The core dashboard reports core-only numbers alongside full inventory for transparency.

## 5. Compatibility Level Definitions

Each migrated route has a `verification_requirement` indicating the depth of compatibility testing needed. "Migrated" means backend-next ownership exists; parity verification remains required unless explicit evidence is produced.

| Level | Meaning | Example |
|-------|---------|---------|
| `status-200` | Returns HTTP 200 on success | `/healthz` |
| `shape` | Response JSON structure matches | `/api/v3/resources/` |
| `shape+semantics` | Shape + field semantics match | `/api/v2/market-order/` |
| `shape+semantics+side-effects` | Behavior includes side effects | `/api/v3/market-ticker/` |

## 6. Frontend Callsite Coverage

58 frontend API callsites across 14 files, grouped into ~30 workflows:

| Workflow | Callsites | Status |
|----------|-----------|--------|
| Auth (login/register) | 2 | migrated |
| Company view | 2 | remaining-core |
| Building lifecycle (list/buy/place/move/upgrade/demolish/shop) | 8 | remaining-core |
| Production (jobs/queue/claim/options/cancel) | 8 | partly migrated |
| Market (ticker/depth/orderbook/orders) | 7 | migrated (side-effect parity) |
| Financial reporting | 5 | migrated |
| Chat / social | 5 | remaining-core |
| Daily orders | 3 | remaining-core |
| Leaderboard | 1 | remaining-core |
| Powerups / progression | 3 | remaining-core |
| Warehouse | 1 | migrated |
| **Deferred plugins** | **11** | excluded (research 4, executives 5, government 2) |

### Frontend Disposition

| Disposition | Count |
|-------------|-------|
| remaining-core | 29 |
| migrated | 18 |
| deferred-plugin | 11 |
| Total | 58 |

## 7. Core Completion Dashboard

All counts are computed directly from CSV data.

```
Phase 16 Completion Dashboard v3
==================================
Generated: 2026-06-06

--- Legacy Route Counts by Disposition ---
  deferred-plugin:           20
  dev-only:                   6
  migrated:                  24
  remaining-core:            54
  retire-candidate:           2
  TOTAL:                    106

  Core total (excl. deferred-plugin + dev-only): 80
  Deferred-plugin excluded from core: 20

--- Frontend Callsites by Disposition ---
  deferred-plugin:           11
  migrated:                  18
  remaining-core:            29
  TOTAL:                     58

--- Validation ---
  PASS  ALL_CSVS_EXIST:              6/6 files present
  PASS  NO_NON_ASCII:                0 files with non-ASCII
  PASS  SOURCE_LEGACY_PATHS:         exact source path set matches CSV
  PASS  SOURCE_BN_HANDLERS:          exact registered-handler set matches CSV
  PASS  SOURCE_FE_REFS:              exact source file-and-line set matches CSV
  PASS  LEGACY_COUNT:                106 rows (expected 106)
  PASS  LEGACY_VALID:                0 issues
  PASS  LEGACY_UNCLASSIFIED:         0 unclassified
  PASS  CLASSIFICATION_COUNT:        106 rows (expected 106)
  PASS  CLASSIFICATION_VALID:        0 issues, 0 unclassified
  PASS  CROSS_FILE:                  cross-file OK
  PASS  BN_COUNT:                    28 rows (expected 28)
  PASS  BN_VALID:                    0 issues
  PASS  FE_COUNT:                    58 rows (expected 58)
  PASS  FE_VALID:                    0 issues
  PASS  FE_SOURCE_REFS:              all 58 have source_ref
  PASS  SCHEDULER_STAGES:            9 stages (10 calls)
  PASS  SCHEDULER_VALID:             0 issues
  PASS  DG_COUNT:                    37 rows (expected >= 30)
  PASS  DG_VALID:                    0 issues

================== ALL CHECKS PASSED ==================
```

## 8. Deliverable Files

All files under `docs/2026-06-06/rebuild/phase-16/`:

| File | Rows | Schema Highlights |
|------|------|-------------------|
| `legacy-routes.csv` | 106 | method_ambiguity, backend_next_owner, verification_requirement |
| `classification.csv` | 106 | disposition, backend_next_owner, cross-file with legacy-routes |
| `backend-next-routes.csv` | 28 | handler_domain, disposition, verification_requirement |
| `frontend-callsites.csv` | 58 | legacy_owner, backend_next_owner, verification_requirement, source_ref |
| `scheduler-inventory.csv` | 16 | responsible_owner, verification_requirement (9 stages, 10 calls) |
| `data-groups.csv` | 37 | verification_requirement (owners already exist) |

Tooling: `backend-next/scripts/phase16-inventory-validator.ps1` -- v3 with source-drift checks.

## 9. Next Steps (Phase 17+)

1. **Phase 17 (Formula Governance)**: Establish governed formula boundary across all 37 data groups
2. **Phase 18 (Domain/Error/Transaction Boundaries)**: Architectural rules for 54 remaining-core routes
3. **Phase 19 (Trusted Parity)**: Side-effect comparison for 3 market reads with RunBotMarketCycle
4. **Phase 20+ (Migration by priority)**: remaining-core groups: buildings -> production -> market -> social -> progression

Deferred plugins (20 routes, 5 plugins) are excluded from core completion targets.
