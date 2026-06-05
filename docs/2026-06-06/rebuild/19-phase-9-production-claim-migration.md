# Phase 9: Production Claim Migration

Date: 2026-06-06

## Scope

Implemented production claim read and write endpoints in backend-next:
- `POST /api/v2/production/claim/{jobId}/` -- claim produced resources from a single job
- `GET /api/v2/production/claimable/` -- list jobs with claimable amounts

Also updated `GET /api/v2/production/jobs/` to return refreshed status/claimable fields
computed from elapsed time.

## Non-Goals

- Cancel, claim-all, production queue, slots, scheduler
- Quality/input/output maps on ProductionJob (deferred)
- Market, finance APIs, bonds, research, or PostgreSQL changes
- No finance migration beyond the minimal `production_output` ledger audit entry for claims
- Any changes to `backend/`, `client/`, or `decompiled/data/`

## Changes

| File | Description |
|---|---|
| `backend-next/openapi/openapi-draft.yaml` | Added `POST /api/v2/production/claim/{jobId}/` and `GET /api/v2/production/claimable/` paths, `ClaimProductionResponse`, `ClaimableJobDTO`, `ClaimableJobListResponse` schemas; added `claimed_amount`, `claimable_amount` to `ProductionJobDTO` |
| `backend-next/internal/generated/openapi/types.gen.go` | Regenerated via `scripts/generate-openapi.sh` |
| `backend-next/internal/domain/production/types.go` | Added `ClaimedAmount`, `ClaimableAmount`, `XPAwarded` fields to `ProductionJob` |
| `backend-next/internal/app/production/service.go` | Added `FinanceStorage` dependency; `refreshJobStatuses` for time-based claimable computation; `ClaimProduction` with inventory add, XP, ledger entry; `ListClaimableJobs`; updated `ListProductionJobs` to refresh before returning |
| `backend-next/internal/app/production/service_test.go` | Updated `newTestService` for new constructor; added claim/claimable service tests |
| `backend-next/internal/httpapi/production_handler.go` | Added `handleClaimProduction` and `handleListClaimableJobs` handlers |
| `backend-next/internal/httpapi/production_handler_test.go` | Updated `newProductionSvc` for new constructor; added claim/claimable handler tests |
| `backend-next/internal/httpapi/router.go` | Registered two new routes under `/api/v2` |
| `backend-next/cmd/simapi-next/main.go` | Wired `FinanceStorage` into production service |
| `backend-next/go.mod` / `go.sum` | `go mod tidy` added `oapi-codegen/runtime` dependency |
| `docs/2026-06-06/rebuild/19-phase-9-production-claim-migration.md` | This document |

## Behavior

### Claim computation
- `claimableAmount = floor(elapsed / durationSeconds * targetQuantity) - claimedAmount`
- If `elapsed >= durationSeconds`: claimable = remaining = `targetQuantity - claimedAmount`
- Partial claiming allowed: job stays `running` until fully claimed
- Completed jobs are marked `ready` when elapsed >= duration and claimable > 0

### Claim success effects
1. Produced resource added to inventory via `CompanyStorage.UpdateInventory(ctx, companyID, resourceID, +claimAmount)`
2. Job `ClaimedAmount` incremented; if fully claimed, status -> `claimed`
3. Incremental XP awarded: total 10 per job, proportional to claimed fraction
4. Finance ledger entry appended with kind `production_output`, metadata `resourceId/jobId/partial`
5. Company level returned in response

### Error handling
| Condition | Code | Error |
|---|---|---|
| No auth token | 401 | UNAUTHORIZED |
| Job not found / wrong company | 404 | NOT_FOUND (no leak) |
| Already claimed | 400 | CONFLICT |
| Nothing claimable yet | 400 | VALIDATION_ERROR |

### List refresh
`GET /api/v2/production/jobs/` now calls `refreshJobStatuses()` before returning,
so ready/claimable amounts are always current relative to server clock.

## Acceptance Criteria

1. `POST /api/v2/production/claim/{jobId}/` with valid token and claimable job -> 200 with job_id, status, output, claimed_amount, remaining, xp
2. Claiming a completed job marks it `claimed` and credits inventory
3. Partial claim leaves job `running` with reduced remaining
4. Claiming too early (no elapsed) -> 400 "nothing to claim yet"
5. Claiming already claimed job -> 400 "already claimed"
6. Claiming another company's job -> 404 "not found" (no info leak)
7. `GET /api/v2/production/claimable/` -> 200 with typed response; empty = `[]` not null
8. Existing production start/list/write endpoints still work
9. `go test ./...`, `go vet ./...`, `go build ./cmd/simapi-next` all pass

## Verification

```bash
cd backend-next
go test ./... -count=1
go vet ./...
go build ./cmd/simapi-next
```

All 3 commands pass.
