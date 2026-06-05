# Phase 14: Finance Read (Ledger Cashflow + Basic Statements)

Date: 2026-06-06

## Scope

Implemented finance read-only endpoints in backend-next:
- `GET /api/v2/companies/me/cashflow/recent/` -- recent ledger entries with signed moneyDelta
- `GET /api/v2/companies/me/income-statement/` -- revenue, expenses, net income
- `GET /api/v2/companies/me/balance-sheet/` -- assets (money), liabilities (0), equity
- `GET /api/v2/companies/me/cashflow-statement/` -- operating, investing, financing categories
- `GET /api/v3/companies/me/past-finances/` -- daily net series from ledger, with fallback

## Non-Goals

- No finance write endpoints (bonds, interest, etc.)
- No bond issue/buy/call/interest settlement
- No frontend or legacy backend changes
- No PostgreSQL/sqlc
- No scheduler/background jobs

## Changes

| File | Change |
|---|---|
| `backend-next/openapi/openapi-draft.yaml` | Added 5 finance GET paths + 7 schemas (`CashflowEntry`, `RecentCashflowResponse`, `IncomeStatementResponse`, `BalanceSheetResponse`, `CashflowStatementResponse`, `PastFinancePoint`, `PastFinancesResponse`); fixed register response indentation |
| `backend-next/internal/generated/openapi/types.gen.go` | Regenerated via `scripts/generate-openapi.sh` |
| `backend-next/internal/app/finance/service.go` | New service with `GetRecentCashflow`, `GetIncomeStatement`, `GetBalanceSheet`, `GetCashflowStatement`, `GetPastFinances` |
| `backend-next/internal/app/finance/service_test.go` | 7 tests: recentCashflow direction, empty ledger, income aggregation, balance sheet, cashflow categorization, past-finances fallback, past-finances from timestamps |
| `backend-next/internal/httpapi/finance_handler.go` | New handler with 5 auth endpoints |
| `backend-next/internal/httpapi/finance_handler_test.go` | 6 handler tests: 401, recentCashflow 200, income 200, balance 200, cashflow 200, past-finances 200 |
| `backend-next/internal/httpapi/router.go` | Added `*FinanceHandler` param to `NewRouter`; registered 4 v2 + 1 v3 routes |
| `backend-next/cmd/simapi-next/main.go` | Wired `finance.NewService(st, st, clock)` and `NewFinanceHandler` |
| `backend-next/internal/httpapi/*_test.go` | All 5 handler test files updated to pass 8th `nil` financeHandler arg |
| `docs/.../24-phase-14-finance-read-migration.md` | This document |

## Behavior

### Endpoint details

| Endpoint | Computation |
|---|---|
| `recent-cashflow` | Recent 100 ledger entries; `moneyDelta = abs(amount)`, negative if `direction == out`; `oldestPulled` = timestamp of oldest returned entry |
| `income-statement` | Sum `abs(amount)` by direction across last 1000 ledger entries |
| `balance-sheet` | `company.Money` as assets and equity; liabilities = 0 |
| `cashflow-statement` | `bond_*` -> financing; `buy_building/building_upgrade/demolish_building/warehouse_upgrade/research_*/slot_upgrade` -> investing; rest -> operating |
| `past-finances` | Daily net from `CreatedAt` date; fallback: `[{date:2026-05-28, net:890.2}, {date:2026-05-29, net:1022.4}]` |

## Verification

```bash
cd backend-next
go test ./... -count=1    # all pass (112 tests)
go vet ./...              # clean
go build ./cmd/simapi-next # clean
gofmt -w                  # applied
git diff --check          # no whitespace errors
```
