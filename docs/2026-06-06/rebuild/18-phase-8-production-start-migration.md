# Phase 8: Production Start Write - Migration Summary

Date: 2026-06-06

## Scope

Implemented the first production-write endpoint: `POST /api/v2/production/start/`.
This allows an authenticated company to start a production job for a building they own,
validating the resource recipe, deducting input inventory, calculating duration, and
creating a running production job.

## Non-Goals

- Production claim, cancel, queue, or slot management
- Market, finance, schedule, or PostgreSQL changes
- Any changes to `backend/`, `client/`, or `decompiled/` data files
- Staging, committing, or branch changes

## Changes

| File | Description |
|---|---|
| `backend-next/openapi/openapi-draft.yaml` | Added `POST /api/v2/production/start/` endpoint, `BuildingProductionStatus` and `StartProductionResponse` schemas |
| `backend-next/internal/generated/openapi/types.gen.go` | Regenerated via `scripts/generate-openapi.sh` |
| `backend-next/internal/catalog/catalog.go` | NEW - small static-data loader for `resources.json` and `buildings.json` |
| `backend-next/internal/config/config.go` | Exported `FindProjectRoot()` for use by catalog loader |
| `backend-next/internal/storage/memory/memory.go` | Implemented `UpdateInventory` to actually modify company inventory and reject negative final amounts |
| `backend-next/internal/app/production/service.go` | Added `StartProduction` method: building/resource validation, recipe lookup, inventory deduction, duration calculation, job creation |
| `backend-next/internal/httpapi/production_handler.go` | Added `handleStartProduction` handler with JSON decoding, basic validation, and error mapping |
| `backend-next/internal/httpapi/router.go` | Registered `POST /api/v2/production/start/` route |
| `backend-next/cmd/simapi-next/main.go` | Wired catalog loading and new production service dependencies |
| `backend-next/internal/app/production/service_test.go` | Updated `NewService` calls; added 4 StartProduction tests |
| `backend-next/internal/httpapi/production_handler_test.go` | Updated `NewService` calls; added 3 StartProduction handler tests |
| `docs/2026-06-06/rebuild/18-phase-8-production-start-migration.md` | This document |

## Acceptance Criteria

1. `POST /api/v2/production/start/` with valid token, valid building, producible resource,
   sufficient inventory -> 200 with `{job: ProductionJobDTO, building: BuildingProductionStatus}`
2. Missing or invalid token -> 401
3. Invalid request body -> 400
4. Building not owned / resource not producible -> 404 or 400
5. Insufficient input inventory -> 400 with `INSUFFICIENT_INVENTORY`
6. Duration exceeding 48h cap -> 400
7. Created job is observable via `GET /api/v2/production/jobs/`
8. Inventory is deducted correctly after job creation
9. On failure (inventory insufficient), no job is created and inventory is unchanged

## Implementation Details

### Duration formula
```
duration_seconds = max(30, ceil(quantity / (producedPerHourRaw * max(level,1) * ProductionMod) * 3600))
```
`ProductionMod` from `game.json` (currently 1.02) is applied as a divisor:
higher mod = faster production. Capped at 172800 seconds (48h); exceeding returns 400 validation error.

### Recipe lookup
`resources.json` entries contain `producedFrom` mapping `resourceId -> amountPerUnit`.
Required input = `ceil(amountPerUnit * quantity)`. If `producedFrom` is empty (raw resources),
no inputs are deducted.

### Building validation
Company's `buildings[]` must contain the requested `building_id`. The building's `BuildingID` (type ID)
is looked up in the buildings catalog; the `produces` array must include the requested `resource_id`.

### Inventory deduction
Uses `CompanyStorage.UpdateInventory(ctx, companyID, resourceID, delta)` with negative delta.
The memory implementation rejects negative final amounts. All inputs are pre-checked before
deduction to avoid partial changes.

### Job fields
- `ID`: generated via `platform.IDGen.Next("prod")` (nanosecond timestamp + sequence)
- `Status`: `running`
- `StartedAt`: current clock time
- `TargetQuantity` = `Quantity` (no partial claiming yet)

## Verification

```bash
cd backend-next
go test ./...
go vet ./...
go build ./cmd/simapi-next
```

All 3 commands pass.
