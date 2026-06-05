# Phase 11: Market Order Write Migration

Date: 2026-06-06

## Scope

Implemented market order create and cancel endpoints in backend-next:
- `POST /api/v2/market-order/` -- create a buy or sell order
- `DELETE /api/v2/market-order/cancel/{orderId}/` -- cancel an open order

## Non-Goals

- No market matching / trade execution
- No take-order behavior
- No bot replenishment, scheduler, or websocket
- No XP updates on order creation or cancellation
- No frontend/legacy backend changes
- No quality-aware inventory (non-zero quality rejected with validation error)
- No market tick-size formula validation yet

## Changes

| File | Description |
|---|---|
| `backend-next/openapi/openapi-draft.yaml` | Added `POST /api/v2/market-order/` and `DELETE /api/v2/market-order/cancel/{orderId}/` paths; added `CreateOrderRequestFrontend`, `CreateOrderResponse`, `CancelOrderResponse` schemas |
| `backend-next/internal/generated/openapi/types.gen.go` | Regenerated via `scripts/generate-openapi.sh` |
| `backend-next/internal/app/market/service.go` | Added `companies storage.CompanyStorage`, `finance storage.FinanceStorage`, `idgen *platform.IDGen` to Service; implemented `CreateOrder` (buy reserves cash, sell reserves inventory) and `CancelOrder` (buy refunds cash+ledger, sell returns inventory) |
| `backend-next/internal/app/market/service_test.go` | Added `newTestCompany` helper; 8 new tests: buy reserves cash+creates order+ledger, sell reserves inventory, invalid payloads (5 cases), insufficient funds, insufficient inventory, cancel buy refunds cash+ledger, cancel sell returns inventory, wrong company cannot cancel, already cancelled cannot cancel again |
| `backend-next/internal/httpapi/market_handler.go` | Added `handleCreateOrder` and `handleCancelOrder` handlers with validation error mapping |
| `backend-next/internal/httpapi/market_handler_test.go` | Updated `newMarketSvc` for new constructor; added 6 handler tests: create 401, create 400 invalid JSON, create 200 success, cancel 401, cancel 200 success, cancel 404 missing |
| `backend-next/internal/httpapi/router.go` | Registered `POST /market-order/` and `DELETE /market-order/cancel/{orderId}/` under `/api/v2` |
| `backend-next/cmd/simapi-next/main.go` | Wired `st` (as company+finance), `application.IDGen` into `market.NewService` |
| `docs/2026-06-06/rebuild/21-phase-11-market-order-write-migration.md` | This document |

## Behavior

### Create Order
- Validates: resource exists in catalog, kind in {0,1}, quantity > 0, price > 0, quality == 0
- Buy order: deducts `quantity * price` from company.Money, creates open order, appends market_buy_reserve ledger entry
- Sell order: calls `UpdateInventory(ctx, companyID, resourceID, -quantity)`, creates open order
- Best-effort rollback if CreateOrder storage call fails

### Cancel Order
- Validates order exists, belongs to authenticated company, not already filled/cancelled
- Buy cancel: refunds `remaining * price` to company.Money, appends market_buy_refund ledger entry
- Sell cancel: returns remaining inventory via `UpdateInventory(ctx, companyID, resourceID, +remaining)`
- Sets order Status=cancelled, preserves FilledQuantity as actual executed quantity, persists via UpdateOrder

### Error Codes
| Condition | HTTP | Error Code |
|---|---|---|
| No auth | 401 | UNAUTHORIZED |
| Invalid JSON body | 400 | BAD_REQUEST |
| Resource not found | 404 | NOT_FOUND |
| Insufficient funds (buy) | 400 | INSUFFICIENT_FUNDS |
| Insufficient inventory (sell) | 400 | INSUFFICIENT_INVENTORY |
| Non-zero quality | 400 | BAD_REQUEST |
| Validation errors | 400 | BAD_REQUEST |
| Cancel already settled | 400 | CONFLICT |
| Cancel missing/wrong company | 404 | NOT_FOUND |

## Verification

```bash
cd backend-next
go test ./... -count=1    # 88 tests pass
go vet ./...              # clean
go build ./cmd/simapi-next # clean
gofmt -w                  # applied
git diff --check          # no whitespace errors
```
