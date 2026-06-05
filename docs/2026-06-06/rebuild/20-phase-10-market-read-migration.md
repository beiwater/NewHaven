# Phase 10: Market Read Migration

Date: 2026-06-06

## Scope

Implemented market read endpoints in backend-next (all GET, no mutations):
- `GET /api/v3/resources/` -- list market-tradable resources
- `GET /api/v3/market-ticker/{resourceId}/` -- ticker with fallback series
- `GET /api/v3/market-depth/{resourceId}/{quality}/` -- aggregated buy/sell depth, top 5
- `GET /api/v3/market/{resourceId}/{quality}/` -- order list filtered by resource+quality

## Non-Goals

- No POST/PUT/DELETE market mutations (create/cancel/take)
- No market matching engine, bot cycle, or scheduler
- No websocket or streaming
- No finance settlement, inventory/money mutation
- No changes to legacy `backend/`, `client/`, or `decompiled/data/`

## Changes

| File | Description |
|---|---|
| `backend-next/openapi/openapi-draft.yaml` | Added 4 market GET paths with typed request/response schemas; added `ResourceDefinition`, `ResourcesResponse`, `MarketTickerPoint`, `MarketTickerResponse`, `MarketDepthLevel`, `MarketDepthResponse`, `MarketOrderDTO`, `MarketOrderListResponse` |
| `backend-next/internal/generated/openapi/types.gen.go` | Regenerated via `scripts/generate-openapi.sh` |
| `backend-next/internal/catalog/catalog.go` | Added `IsResearch bool` field to `ResourceEntry` |
| `backend-next/internal/domain/market/types.go` | Added `Quality int` to `MarketOrder`; added `Remaining()` method |
| `backend-next/internal/storage/interfaces.go` | Added `GetOrdersByResource(ctx, resourceID)` to `MarketStorage` |
| `backend-next/internal/storage/memory/memory.go` | Implemented `GetOrdersByResource` |
| `backend-next/internal/app/market/service.go` | New service: `ListResources`, `GetMarketTicker` (storage-backed + fallback with injected clock), `GetMarketDepth`, `ListMarketOrders` |
| `backend-next/internal/app/market/service_test.go` | 8 tests: resources filtering, depth sort/top5, depth aggregation/partial remaining, depth empty, ticker fallback, orders quality filter, orders empty |
| `backend-next/internal/httpapi/market_handler.go` | New handler with 4 auth+path-param handlers |
| `backend-next/internal/httpapi/market_handler_test.go` | 8 tests covering unauthorized access, successful GETs, and invalid path parameters |
| `backend-next/internal/httpapi/router.go` | Added `marketHandler *MarketHandler` to `NewRouter`; registered 4 routes under `/api/v3` |
| `backend-next/cmd/simapi-next/main.go` | Wired `market.NewService(st, resources)` and `httpapi.NewMarketHandler(marketSvc)` |
| `backend-next/internal/httpapi/*_test.go` | Updated all `NewRouter` calls to pass 7th `nil` marketHandler arg |
| `backend-next/internal/app/production/service.go` | Stabilized production job list ordering by job ID after full-suite tests exposed map iteration flakiness |
| `docs/2026-06-06/rebuild/20-phase-10-market-read-migration.md` | This document |

## Endpoint Behaviors

### `GET /api/v3/resources/`
- Filters catalog for `DbLetter > 0`, `IsExchangeTradable true`, `IsResearch false`
- Returns `{resources: [{resourceId, name, producedFrom, producedPerHourRaw, unitsSoldAnHour, hasEconomyModel}]}`
- Sorted by `resourceId` for deterministic UI/test output

### `GET /api/v3/market-ticker/{resourceId}/`
- Reads stored `Ticker` from `MarketStorage.GetTicker` if available (flat price series)
- Falls back to synthetic 48-point series using `BasePrice` from catalog + deterministic sine wave
- Uses injected `platform.Clock` for deterministic tests and repeatable hour buckets
- No bot cycle or active matching invoked
- Returns `{resource, series: [{price, time}]}`

### `GET /api/v3/market-depth/{resourceId}/{quality}/`
- Reads orders from `GetOrdersByResource`, filters by quality and open/partial status
- Aggregates remaining quantity by price level
- Returns top 5 buys (descending price) and top 5 sells (ascending price)
- Each level has `{price, quantity, qty}` (qty mirrors quantity for frontend compatibility)
- Empty: returns `{buys: [], sells: []}` not null

### `GET /api/v3/market/{resourceId}/{quality}/`
- Returns all orders (any status) for the resource+quality combination
- DTO fields: `id, resourceId, kind (0=sell, 1=buy), price, quality, quantity, remaining, companyId, createdAt, status`
- `remaining` is computed as `quantity - filledQuantity`
- Sorted deterministically: sell orders by ascending price, buy orders by descending price, then ID

## Verification

```bash
cd backend-next
go test ./... -count=1    # all package tests pass
go vet ./...              # clean
go build ./cmd/simapi-next # clean
gofmt -w on changed files  # clean
git diff --check           # no whitespace errors
```
