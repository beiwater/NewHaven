# API Route Inventory

Last updated: 2026-06-06

This document groups the current backend HTTP routes for refactor planning. It is not a full API contract; keep `docs/api-contract.md` as the canonical contract document when endpoint behavior is formalized.

## Public Health

| Route | Notes |
|-------|-------|
| `GET /healthz` | Health check. |
| `GET /readyz` | Readiness check. |

## Auth

| Route | Notes |
|-------|-------|
| `/api/register` | Registered in auth handler; currently allowed through auth wrapper exception. |
| `/api/login` | Login and token issuing; dev flow depends on `dev` / `dev`. |
| `/api/csrf/` | Compatibility endpoint. |

## Company And Player

| Route group | Notes |
|-------------|-------|
| `/api/v2/players/me/companies/` | Current player companies. |
| `/api/v2/players/` | Player lookup by ID path. |
| `/api/v2/companies/me/buildings/` | Current company buildings. |
| `/api/v2/companies/me/warehouse/` | Warehouse state. |
| `/api/v2/companies/me/warehouse/upgrade/` | Warehouse upgrade. |
| `/api/v2/companies/me/tutorial/` | Tutorial completion. |
| `/api/v2/companies/me/administration-overhead/` | Admin overhead. |
| `/api/v2/companies/` | Company compatibility group. |
| `/api/v3/companies/` | Newer company compatibility group. |
| `/api/v2/companies/me/achievements/` | Achievements. |
| `/api/v2/no-cache/companies/me/achievements/` | No-cache achievements. |
| `/api/v2/no-cache/companies/achievements/` | Achievement deletion/admin-style endpoint. |

## Buildings And Production

| Route group | Notes |
|-------------|-------|
| `/api/v1/buildings/` | Legacy building route group. |
| `/api/v2/buildings/` | Building production/upgrade/busy path group. |
| `/api/v2/buildings/market/` | Building market. |
| `/api/v2/buildings/buy/` | Buy building. |
| `/api/v2/buildings/place/` | Place building on map. |
| `/api/v2/buildings/move/` | Move placed building. |
| `/api/v2/buildings/demolish/` | Demolish building. |
| `/api/v2/production/jobs/` | Production jobs. |
| `/api/v2/production/queue/` | Production queue. |
| `/api/v2/production/slots/add/` | Add production slot. |
| `/api/v2/production/cancel/` | Cancel production. |
| `/api/v2/production/claim/` | Claim one production job. |
| `/api/v2/production/claimable/` | Claimable production state. |
| `/api/v2/production/claim-all/` | Claim all production. |
| `/api/v2/recipes/` | Recipe list/detail. |
| `/api/v2/production-modifiers/` | Production modifiers. |

## Market And Resources

| Route group | Notes |
|-------------|-------|
| `/api/v3/market-ticker/` | Market ticker. |
| `/api/v3/market/` | Market orderbook by resource/quality path. |
| `/api/v3/market-depth/` | Depth by resource/quality path. |
| `/api/v2/market-order/` | Create market order. |
| `/api/v2/market-order/cancel/` | Cancel order. |
| `/api/v2/market-order/take/` | Take order. |
| `/api/market/buy/orders/` | Legacy buy order path. |
| `/api/v3/resources/` | Resource list/detail. |
| `/api/v3/resources-info/` | Resource info. |
| `/api/v2/weather/` | Weather state. |

## Finance

| Route group | Notes |
|-------------|-------|
| `/api/bonds/` | Bond list/detail/create/buy/call path group. |
| `/api/bonds/settle-interest/` | Settle bond interest. |
| `/api/v2/companies/me/bonds/owned/` | Owned bonds. |
| `/api/v2/companies/me/bonds/sold/` | Sold bonds. |
| `/api/v2/companies/me/income-statement/` | Income statement. |
| `/api/v2/companies/me/balance-sheet/` | Balance sheet. |
| `/api/v2/companies/me/cashflow-statement/` | Cashflow statement. |
| `/api/v2/companies/me/cashflow/recent/` | Recent cashflow. |
| `/api/v2/companies/me/past-finances-overview/` | Past finance overview. |
| `/api/v3/companies/me/past-finances/` | Past finance detail. |

## Research, Executives, And Boosts

| Route group | Notes |
|-------------|-------|
| `/api/v2/research/` | Research projects. |
| `/api/v2/research/start/` | Start research. |
| `/api/v2/research/progress/` | Research progress. |
| `/api/v2/research/complete/` | Complete research. |
| `/api/v2/executives/search/` | Search executives. |
| `/api/v2/executives/recruit/` | Recruit executive. |
| `/api/v2/executives/train/` | Train executive. |
| `/api/v3/executives/poach/` | Poach executive. |
| `/api/v3/executives/offers/` | Executive offers. |
| `/api/v3/executives/` | Executive detail path group. |
| `/api/v2/players/simboosts-use/` | Use simboost. |
| `/api/v2/players/simboosts/` | Simboost types. |
| `/api/v2/players/unlocked-hqs/` | Unlocked HQs. |
| `/api/v2/players/devices/` | Player devices. |
| `/api/v2/players/me/level/` | Level info. |
| `/api/v2/players/me/xp/` | Add XP. |
| `/api/v2/players/me/level-rewards/` | Level rewards. |
| `/api/v2/players/me/offline-income/` | Offline income. |

## Orders, Government, Auctions, And Contracts

| Route group | Notes |
|-------------|-------|
| `/api/v2/orders/daily/` | Daily orders. |
| `/api/v2/orders/daily/complete/` | Complete daily order. |
| `/api/v2/orders/daily/claim/` | Claim daily order reward. |
| `/api/v3/government-orders/` | Government orders. |
| `/api/v3/government-orders/bid/` | Bid on government order. |
| `/api/v3/government-orders/award/` | Award government orders. |
| `/api/v3/government-orders/deliver/` | Deliver government order. |
| `/api/v3/government-orders/resolve-defaults/` | Resolve defaults. |
| `/api/v2/auctions/` | Auction list/detail/bid path group. |
| `/api/v2/companies/me/auctions/` | Company auctions. |
| `/api/v3/contracts-incoming/` | Incoming contracts. |
| `/api/v3/contracts-outgoing/me/` | Outgoing contracts. |
| `/api/v2/contracts-history-incoming/` | Incoming contract history. |
| `/api/v2/contracts-history-outgoing/` | Outgoing contract history. |
| `/api/v2/warehouse-contracts-summary/` | Warehouse contract summary. |

## Social, Newspaper, And Leaderboard

| Route group | Notes |
|-------------|-------|
| `/api/messages/` | Messages compatibility endpoint. |
| `/api/messages_by_company/` | Messages by company. |
| `/api/v2/message/` | Send/read v2 message group. |
| `/api/v2/chatroom/` | Chatroom. |
| `/api/v2/contacts/` | Contacts. |
| `/api/v2/newspaper/articles-by-author/` | Newspaper articles by author. |
| `/api/v2/newspaper/articles/` | Newspaper article list/detail. |
| `/api/v2/newspaper/publishing-costs/` | Publishing costs. |
| `/api/v2/leaderboard/` | Leaderboard. |

## Dev And Misc

| Route group | Notes |
|-------------|-------|
| `/api/dev/ledger/` | Dev ledger. |
| `/api/dev/formulas/production/` | Production formula inspection. |
| `/api/dev/formulas/retail/` | Retail formula inspection. |
| `/api/dev/formulas/retail-season-weather/` | Retail formula inspection with season/weather. |
| `/api/dev/time/` | Simulated time. |
| `/api/v4/` | v4 compatibility group, including payment package paths. |
| `/api/v2/aerospace/projects/` | Aerospace projects. |
| `/api/v2/aerospace/projects/create/` | Create aerospace project. |
| `/api/v2/aerospace/launches/` | Launch history. |
| `/api/v2/aerospace/launch/` | Launch action. |
| `/api/v2/aerospace/components/` | Aerospace components. |

## Refactor Notes

- Route versions are mixed by feature and file. Preserve all active route groups until frontend usage is mapped.
- `handler/dev.go` contains routes that may be product-facing compatibility endpoints, so do not delete or hide them based on filename alone.
- Handler contract tests should be expanded before moving route registration.
- This inventory should be regenerated after each route consolidation phase.

