# Phase 13: Auto Matching on Limit-Order Creation

Date: 2026-06-06

## Scope

Added automatic order matching when a limit order is created in backend-next:

- When a **buy** limit order is created (`POST /api/v2/market-order/`, kind=1), it now matches against existing sell orders with price <= buy price.
- When a **sell** limit order is created (kind=0), it now matches against existing buy orders with price >= sell price.
- Single-pass matching only: the new order matches against the existing book. No cascading or cross-book matching.
- No new HTTP endpoint. The existing `POST /api/v2/market-order/` endpoint behavior is enhanced.

## Non-Goals

- No bot replacement, market pressure, or scheduler
- No XP, AML, anticheat, or price tick validation
- No daily high/low stats (beyond simple ticker LastPrice/Volume24h)
- No quality-aware inventory
- No OpenAPI, storage interface, handler, router, or main.go changes

## Changes

| File | Change |
|---|---|
| `backend-next/internal/app/market/service.go` | Added `matchNewBuyOrder`, `matchNewSellOrder`, `executeMatchFill` private methods; inserted matching call in `CreateOrder` after reserve ledger |
| `backend-next/internal/app/market/service_test.go` | 5 new tests: buy auto-matches sell, sell auto-matches buy, partial match, no match when price does not cross, no match same-company |
| `docs/.../23-phase-13-market-auto-match-migration.md` | This document |

## Behavior

### Match flow (buy order created)
1. `CreateOrder` reserves cash and creates the buy order as before
2. `matchNewBuyOrder` finds sell orders: `IsBuy=false`, status open/partial, same resource+quality, `Price <= order.Price`, different company
3. Sells sorted by price ascending, CreatedAt ascending, ID ascending
4. For each sell, `executeMatchFill` fills at sell.Price (execution price)

### Match flow (sell order created)
1. `CreateOrder` reserves inventory and creates the sell order as before
2. `matchNewSellOrder` finds buy orders: `IsBuy=true`, status open/partial, same resource+quality, `Price >= order.Price`, different company
3. Buys sorted by price descending, CreatedAt ascending, ID ascending
4. For each buy, `executeMatchFill` fills at sell.Price (new order's price = execution price)

### Per-fill accounting

| Action | Details |
|---|---|
| **Execution price** | `sellOrder.Price` |
| **Fee** | `ceil(fill * execPrice * ExchangeFeePct)` (deducted from seller's proceeds) |
| **Buyer inventory** | `UpdateInventory(buyer, resource, +fill)` |
| **Buyer refund** | If `buyOrder.Price > execPrice`: refund `fill * (buyOrder.Price - execPrice)` via `UpdateCompany`, append `market_buy_refund` ledger |
| **Seller revenue** | `seller.Money += fill * execPrice - fee`, append `market_trade` (in) and `market_fee` (out) ledger entries |
| **Order updates** | Both orders' `FilledQuantity += fill`; status set to `filled` (if remaining==0) or `partial` |
| **Trade** | `market.Trade` recorded with buy/sell order IDs, price, quantity, quality, buyerFee=0, sellerFee=fee |
| **Ticker** | `LastPrice`, `Volume24h += fill*execPrice`, `UpdatedAt` |
| **Mutual exclusion** | All under existing `s.mu.Lock()` from Phase 12 |

### Same-company exclusion
Orders from the same company are never matched against each other, preventing wash trading.

## Verification

```bash
cd backend-next
go test ./... -count=1    # all pass (101 tests, 5 new)
go vet ./...              # clean
go build ./cmd/simapi-next # clean
gofmt -w                  # applied
git diff --check          # no whitespace errors
```
