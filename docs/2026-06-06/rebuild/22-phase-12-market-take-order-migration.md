# Phase 12: Market Take-Order (Direct Buy)

Date: 2026-06-06

## Scope

Implemented market take-order endpoint in backend-next:
- `POST /api/v2/market-order/take/` -- buy directly from sell orders (market order fill)

Completes the core market order lifecycle: create -> read -> take -> cancel.

## Non-Goals

- No create-order automatic matching (Phase 13+)
- No buy-order matching loop / limit order matching
- No bot replenishment, scheduler, websocket
- No XP, AML, anticheat, national team, weather, production modifiers
- No price tick validation / formula package migration
- No frontend or legacy backend changes
- No quality-aware inventory migration

## Changes

| File | Change |
|---|---|
| `backend-next/openapi/openapi-draft.yaml` | Added `POST /api/v2/market-order/take/` path, `TakeOrderRequest`, `TakeOrderResponse`, `TradeDTO` schemas |
| `backend-next/internal/generated/openapi/types.gen.go` | Regenerated via `scripts/generate-openapi.sh` |
| `backend-next/internal/domain/market/types.go` | Added `Quality int` field to `Trade` struct |
| `backend-next/internal/app/market/service.go` | Added `sync.Mutex`, `cfg *config.GameConfig` to Service; implemented `TakeOrder` with price-ascending fill, fee calculation, trade recording, ticker update, ledger entry; guarded `CreateOrder`/`CancelOrder`/`GetMarketTicker`/`GetMarketDepth`/`ListMarketOrders` with mutex |
| `backend-next/internal/app/market/service_test.go` | Updated `newTestSvc` with config; added 7 service tests (buy from best orders, partial fill when sells run out, zero returns, stop on affordability, trades+ticker, ledger+seller credit, bad payloads) |
| `backend-next/internal/httpapi/market_handler.go` | Added `handleTakeOrder` handler |
| `backend-next/internal/httpapi/market_handler_test.go` | Updated `newMarketSvc` with config; added 4 handler tests (401, 400 invalid JSON, 400 invalid payload, 200 success) |
| `backend-next/internal/httpapi/router.go` | Registered `POST /api/v2/market-order/take/` under `/api/v2` |
| `backend-next/cmd/simapi-next/main.go` | Passed `cfg.Game` to `market.NewService` |
| `docs/.../22-phase-12-market-take-order-migration.md` | This document |

## Behavior

### Take Order algorithm
1. Validate: company exists, resource in catalog, quantity>0, maxPrice>0, quality==0
2. Get sell orders from storage for the resource
3. Filter: `IsBuy==false`, status open/partial, matching quality, `Price <= maxPrice`, `Remaining() > 0`
4. Sort by price asc, CreatedAt asc, ID asc
5. For each candidate seller:
   - `fill = min(need, sell.Remaining())`
   - `fee = ceil(fill * price * ExchangeFeePct)` (default 4% from game.json)
   - `cost = fill * price + fee`
   - If taker cannot afford full cost, stop (no partial fill per-order)
   - Deduct cost from taker money, add fill to taker inventory
   - Credit seller money by `fill * price` (if different company)
   - Update sell order status (filled/partial) and FilledQuantity
   - Record trade, update ticker, append `market_take_buy` ledger entry
6. Persist taker money deduction per fill before downstream mutations, with best-effort rollback on later fill failures
7. Return `{amountBought, trades[], moneyDelta}`

### Error Codes
| Condition | HTTP | Error Code |
|---|---|---|
| No auth | 401 | UNAUTHORIZED |
| Invalid JSON body | 400 | BAD_REQUEST |
| Resource not found | 404 | NOT_FOUND |
| Validation errors | 400 | BAD_REQUEST |

### Concurrency
Added `sync.Mutex` to Service. All read/write market methods acquire the mutex, preventing races on take-order fills and order state mutations.

## Verification

```bash
cd backend-next
go test ./... -count=1    # all pass
go vet ./...              # clean
go build ./cmd/simapi-next # clean
gofmt -w                  # applied
git diff --check          # no whitespace errors
```
