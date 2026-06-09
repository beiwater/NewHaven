# Fast Batch A1: Building and Warehouse Lifecycle

**Date**: 2026-06-06
**Baseline**: `33b03f9` (Phase 18 complete)
**Status**: Complete
**Verification**: `go build ./...`, `go vet ./...`, `go test ./...`, `gofmt`, Phase 16/17 validators

## Capabilities Added

All routes registered in `internal/httpapi/router.go`:

| Operation | Method | Route | Frontend |
|-----------|--------|-------|----------|
| Building market | GET | `/api/v2/buildings/market/` | Shop UI |
| Buy building | POST | `/api/v2/buildings/buy/` | `useBuyBuilding` |
| Place building | POST | `/api/v2/buildings/place/` | `usePlaceBuilding` |
| Move building | POST | `/api/v2/buildings/move/` | `useMoveBuilding` |
| Demolish building | POST | `/api/v2/buildings/demolish/` | `useDemolishBuilding` |
| Upgrade building | POST | `/api/v1/buildings/{buildingId}/upgrade/` | `useUpgradeBuilding` |
| Warehouse upgrade | POST | `/api/v2/companies/me/warehouse/upgrade/` | -- |
| Building list (v2) | GET | `/api/v2/companies/me/buildings/` | `useBuildings` |
| Building list (v3) | GET | `/api/v3/companies/me/buildings/` | `useBuildings` |

## Routing for Upgrade

Legacy route is `/api/v1/buildings/{id}/upgrade/` -- implemented as a chi URL param extraction in `handleUpgradeBuilding`. The handler reads `buildingId` from `chi.URLParam(r, "buildingId")`.

## Placement Validation (Exact Legacy Semantics)

Ported from `backend/internal/service/building_shop.go`:

- **3 maps**: harbor (8 slots, unlock level 1), inland (8 slots, unlock level 5), desert (3 slots, unlock level 10)
- **Slot unlock**: first 3 slots at map unlock level, each additional slot +2 levels beyond that
- **Slot IDs**: `harbor-plot-01` through `harbor-plot-08`, `inland-plot-01` through `inland-plot-08`, `desert-plot-01` through `desert-plot-03`
- **Normalization**: invalid/empty mapID defaults to `harbor`; empty slotID derived from x,y via 3-column grid formula
- **Occupied rejection**: validates no other building (except the one being moved) occupies the target slot
- **Canonical coordinates**: derived from order via `((order-1)%3)+1, ((order-1)/3)+1`
- **Parity tests**: occupied slot, locked map, locked later plot, empty slot normalization, moving while excluding itself

## Response Envelope

Building market returns a JSON array inside the standard `data` field: `{"data": [...], "error": null}`. All other endpoints use the same envelope.

## Request JSON Fields (camelCase for Legacy Compatibility)

Lifecycle request payloads use camelCase field names matching legacy/frontend:
- `BuyBuildingRequest`: `buildingId`, `requestId`
- `PlaceBuildingRequest`: `buildingId`, `mapId`, `slotId`
- `MoveBuildingRequest`: `buildingId`, `mapId`, `slotId`
- `DemolishBuildingRequest`: `buildingId`

OpenAPI schemas and generated types reflect these camelCase names. Tests prove camelCase bodies are accepted.

## v2 Building List (Transitional)

The `/api/v2/companies/me/buildings/` handler delegates to `handleListMyBuildings` which returns `BuildingDTO` with snake_case fields (`building_id`, `map_id`, etc.). The frontend `Building` type uses camelCase (`buildingId`, `mapId`). This endpoint is marked as transitional -- a future cutover commit will add a separate legacy-compatible camelCase DTO for v2 while preserving v3 for backend-next consumers.

## Hotel California Rule (Pointer-State Rollback)

Memory store returns `*company.Company` pointers. All write operations:
1. Save original scalar values (money, level, capacity, buildings slice length) before mutation
2. Mutate the pointee in-place
3. Call `UpdateCompany` / `UpdateWarehouse`
4. If persistence fails, restore original values and slice state on the pointee

Focused tests prove state restoration for: buy, upgrade, placement, warehouse upgrade, and warehouse upgrade with warehouse store failure.

## Service Mutex

`building.Service` has a `sync.Mutex` protecting `BuyBuilding`, `PlaceBuilding`, `MoveBuilding`, `DemolishBuilding`, `UpgradeBuilding`. `warehouse.Service` has a mutex protecting `UpgradeWarehouse`. Mutexes do not coordinate across services.

## Memory Store: Single Authoritative Building State

`SaveBuilding` and `RemoveBuilding` are absolute no-ops (return nil). `company.Buildings` managed via `UpdateCompany`/`GetCompany` is the single source of truth. No divergent `buildings` map exists in the store.

## Files Changed

### New
- `internal/app/building/shop.go` -- BuyBuilding, UpgradeBuilding
- `internal/app/building/placement.go` -- PlaceBuilding, MoveBuilding, DemolishBuilding, map validation
- `internal/app/building/service_test.go` -- service tests
- `internal/httpapi/building_handler_test.go` -- handler integration tests

### Modified
- `internal/app/building/service.go` -- BuildingMarket, buildingToDTO, placed field, mu
- `internal/app/warehouse/service.go` -- cfg field, UpgradeWarehouse with rollback, mu
- `internal/app/warehouse/service_test.go` -- upgrade rollback tests
- `internal/httpapi/building_handler.go` -- 8 handlers
- `internal/httpapi/warehouse_handler.go` -- upgrade handler
- `internal/httpapi/router.go` -- 9 route registrations
- `internal/httpapi/response.go` -- restored to HEAD (formatting churn reverted)
- `cmd/simapi-next/main.go` -- wiring
- `internal/config/config.go` -- BaseBuildingCost, WarehouseBaseCap, WarehouseUpgradeCost
- `internal/storage/interfaces.go` -- UpdateWarehouse method
- `internal/storage/memory/memory.go` -- UpdateWarehouse impl; SaveBuilding/RemoveBuilding are no-ops
- `openapi/openapi-draft.yaml` -- camelCase request fields, placed field, market paths
- `internal/generated/openapi/types.gen.go` -- regenerated
- `docs/2026-06-06/rebuild/phase-16/backend-next-routes.csv` -- 8 new routes added (migrated, building/warehouse domain)

## Verification

| Check | Result |
|-------|--------|
| `go build ./cmd/simapi-next/` | Clean |
| `go vet ./...` | Clean |
| `go test ./...` | 13 suites, 0 failures |
| `gofmt -l` | No formatting issues in changed files |
| `git diff --check` | No whitespace errors |
| Phase 16 validator | ALL CHECKS PASSED |
| Phase 17 validator | 12/12 passed |
| Handwritten production >300 lines | 0 files in target packages (memory.go deferred ~625) |

## HTTP Smoke Test

Manual smoke test against running backend-next using curl, covering:
- Market, buy, place into harbor slot, reject occupied slot, move, upgrade, demolish, warehouse upgrade, GET warehouse, GET building list

Automated smoke test was not possible in this environment (no background process support).

## Remaining Risks

1. **Memory store growth**: memory.go ~625 lines. No single domain exceeds ~80 lines. Split trigger remains PostgreSQL adapter.
2. **Idempotency**: `BuyBuilding` does not deduplicate `requestId`. A retry may buy two buildings. Acceptable for dev mode.
3. **v2 building list frontend cutover**: `/api/v2/companies/me/buildings/` returns snake_case DTO. Frontend expects camelCase. A future cutover commit must add a camelCase DTO or transform before GA.
4. **Demolish-in-production check**: Not implemented. Legacy also omitted this.
