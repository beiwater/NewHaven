# Backend Inventory

Last updated: 2026-06-06

This inventory captures the backend structure at the start of the backend refactor cycle. It is meant to guide planning, not prescribe immediate file moves.

## Current Shape

| Area | Current path | Notes |
|------|--------------|-------|
| Entry point | `backend/cmd/simapi` | Small server wiring and graceful shutdown. |
| HTTP handlers | `backend/internal/handler` | Standard library `http.ServeMux`, one rough file per feature/domain. |
| Middleware | `backend/internal/middleware` | Auth/context helpers and shared middleware. |
| Service layer | `backend/internal/service` | Main mutable gameplay owner; public methods usually lock through `Service`. |
| Model | `backend/internal/model` | Shared gameplay structs and `GameState`. |
| Formula | `backend/internal/formula` | Pure-ish economic and math helpers. |
| Scheduler | `backend/internal/scheduler` | Background market/bot/periodic tasks. |
| Storage | `backend/internal/storage` | Optional PostgreSQL persistence and storage interfaces. |
| Config | `backend/internal/config` | Env config and game tuning loader. |
| Data loader | `backend/internal/data` | Static JSON loading for game data. |
| Anti-cheat | `backend/internal/anticheat` | Detection logic and tests. |
| AML | `backend/internal/aml` | Financial/market abuse detection logic. |

## File Counts

| Area | Go files | Test files |
|------|----------|------------|
| `internal/service` | 41 | 12 |
| `internal/handler` | 22 | 1 |

The service layer is already the largest refactor pressure point. The handler layer has broad route coverage but comparatively little contract testing.

## Current Domain Clusters

| Domain | Main files |
|--------|------------|
| Auth/player bootstrap | `service_player.go`, `auth.go`, `auth_test.go`, `handler/auth.go` |
| Company/profile/inventory | `company.go`, `state_snapshot.go`, `level_unlocks.go`, `handler/company.go` |
| Buildings/shop/map placement | `building.go`, `building_shop.go`, `handler/building_shop.go` |
| Production/recipes/queue | `production.go`, `production_claim.go`, `recipe.go`, `handler/production.go`, `handler/production_queue.go`, `handler/recipe.go` |
| Market/orders/depth/ticker | `market_trade.go`, `market_match.go`, `market_info.go`, `market_depth.go`, `market_competition.go`, `order.go`, `handler/market.go` |
| Finance/bonds/statements | `bond.go`, `handler/bond.go`, `handler/financial.go`, `formula/bonds.go`, `formula/costs.go` |
| Research/levels/executives | `research.go`, `research_level.go`, `service_stubs.go`, `handler/dev.go`, `handler/executive.go` |
| Government/contracts/auctions | `government.go`, `auction.go`, `handler/government.go`, `handler/auction.go` |
| Social/messages/news | `state_snapshot.go`, `handler/message.go` |
| Operations/save/load | `service_save.go`, `storage`, `config`, `scheduler` |

## Early Refactor Hotspots

These areas should be planned carefully before implementation:

- `service.Service` mixes use cases, runtime helpers, persistence hooks, time, ID generation, and inventory helpers.
- Handler files expose mixed API versions (`/api`, `/api/v1`, `/api/v2`, `/api/v3`, `/api/v4`) that need compatibility mapping before cleanup.
- `handler/dev.go` currently owns several production-facing-looking v2/v3/v4 routes, so it should be split by contract, not just by filename.
- Production code has timing, resource validation, inventory mutation, building slot checks, and XP rewards in one domain flow.
- Market code has order creation, matching, bot liquidity, depth, ticker, and competition logic spread across multiple files.
- Persistence is optional and should stay adapter-shaped until snapshot/version rules are clear.

## Suggested First Read Order

1. `backend/internal/service/service.go`
2. `backend/internal/model/types.go`
3. `backend/internal/handler/handler.go`
4. `backend/internal/service/company.go`
5. `backend/internal/service/production.go`
6. `backend/internal/service/market_trade.go`
7. `backend/internal/service/market_match.go`
8. `backend/internal/storage/storage.go`
9. `backend/internal/scheduler/scheduler.go`

