# Phase 15: Bonds Read + Issue

Date: 2026-06-06

## Scope

Implemented bond read-only and issue endpoints in backend-next:
- `GET /api/bonds/` -- list all active bonds on market
- `GET /api/bonds/{bondId}/` -- single bond detail
- `POST /api/bonds/` -- issue a new bond
- `GET /api/v2/companies/me/bonds/owned/` -- bonds held by the company (via BondHoldings)
- `GET /api/v2/companies/me/bonds/sold/` -- bonds issued by the company

## Non-Goals

- No buy bond (`PATCH /api/bonds/{bondId}/`)
- No call bond (`PUT /api/bonds/{bondId}/`)
- No settle interest (`POST /api/bonds/settle-interest/`)
- No scheduler/background jobs for interest accrual
- No frontend or legacy backend changes
- No PostgreSQL/sqlc
- No unrelated refactors

## Changes

| File | Change |
|---|---|
| `backend-next/openapi/openapi-draft.yaml` | Added 5 bond paths + 5 schemas (`BondDTO`, `BondListResponse`, `GetBondResponse`, `CreateBondRequest`, `CreateBondResponse`) |
| `backend-next/internal/generated/openapi/types.gen.go` | Regenerated via `scripts/generate-openapi.sh` |
| `backend-next/internal/app/finance/service.go` | Extended `Service` with `idgen *platform.IDGen`, `gameCfg *config.GameConfig`; added `NewService` params; added bond helpers and 5 bond methods |
| `backend-next/internal/app/finance/service_test.go` | Updated `newTestSvc` constructor; added 10 bond tests |
| `backend-next/internal/httpapi/bond_handler.go` | New handler with 5 auth endpoints |
| `backend-next/internal/httpapi/bond_handler_test.go` | 9 handler tests: 401, 200 list empty, 200 list with token, invalid JSON 400, invalid payload 400, create success, get 404, get success, owned/sold 200 |
| `backend-next/internal/httpapi/router.go` | Added `*BondHandler` param to `NewRouter`; registered `/api/bonds/` (3 routes) + `/api/v2/companies/me/bonds/owned/` + `sold/` (2 routes) |
| `backend-next/cmd/simapi-next/main.go` | Wired `finance.NewService(..., idgen, gameCfg)` and `NewBondHandler` |
| `backend-next/internal/httpapi/*_test.go` | Existing handler test files updated to pass 9th `nil` bondHandler arg |
| `docs/.../25-phase-15-bonds-read-issue-migration.md` | This document |

## Behavior

### Interest field protocol

The legacy API accepts `interest` as a percent input (e.g. `1.5` means 1.5%). The stored `Bond.InterestRate` is a decimal (1.5% -> 0.015). The response `BondDTO.interest` returns the stored decimal (0.015).

### Daily interest formula

```
dailyInterest = floor(amount * bondFaceValue * interestRatePct / 100)
```

Where `interestRatePct = storedRate * 100` (e.g. 0.015 * 100 = 1.5). At `bondFaceValue = 5000` and 1.2%, one bond unit yields `floor(1 * 5000 * 1.2 / 100) = 60/day`.

### Endpoint details

| Endpoint | Behavior |
|---|---|
| `GET /api/bonds/` | Returns active bonds (`status == "active"`) sorted by `CreatedAt` desc then `ID` asc; optional `rating` query filters by the legacy rating bucket |
| `GET /api/bonds/{bondId}/` | Single bond lookup by ID; 404 if missing |
| `POST /api/bonds/` | Validates amount > 0 and interest in [min, max]; creates bond with `bond-{id}` prefix; credits `amount * faceValue` to company money; appends `bond_issue` ledger entry; rollbacks company money if storage fails |
| `GET /api/v2/companies/me/bonds/owned/` | Joins `CompanyBondHoldings` with `GetBond`; shows holding quantity as `amount`; sorts by `PurchasedAt` desc then `BondID` asc; skips orphaned holdings |
| `GET /api/v2/companies/me/bonds/sold/` | Returns bonds by `IssuerCompanyID`; sorted by `CreatedAt` desc then `ID` asc |

### Default config values

- `BondFaceValue` defaults to 5000 if nil/zero in GameConfig
- `BondMinInterest` defaults to 0.5
- `BondMaxInterest` defaults to 2.0

## Verification

```bash
cd backend-next
go test ./... -count=1    # all pass (9 packages)
go vet ./...              # clean
go build ./cmd/simapi-next # clean
gofmt -l .                # no formatting issues
git diff --check          # no whitespace errors
```
