# NewHaven API 速查表

生成日期: 2026-06-12 | 共计 85 个端点 | 来源: `heritage/httpapi/`

## 快捷键
- `Ctrl+F` / `Cmd+F` 搜路径、方法名、模块名
- 搜 `POST` 看所有写操作
- 搜 `GET` 看所有读操作

---

## Auth

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| POST   | `/api/register` | `Auth.handleRegister` | Register new account |
| POST   | `/api/login` | `Auth.handleLogin` | Login |

## Health

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| GET | `/healthz` | `handleHealthz` | Health check |
| GET | `/readyz` | `handleReadyz` | Readiness check |

## Bonds

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| GET   | `/api/bonds/` | `Bond.handleListBonds` | List bonds |
| POST   | `/api/bonds/` | `Bond.handleCreateBond` | Create bond |
| GET   | `/api/bonds/{bondId}/` | `Bond.handleGetBond` | Get bond details |
| POST   | `/api/bonds/{bondId}/call/` | `Bond.handleCallBond` | Call bond |
| POST   | `/api/bonds/settle-interest/` | `Bond.handleSettleBondInterest` | Settle bond interest |
| GET   | `/api/v2/companies/me/bonds/owned/` | `Bond.handleGetOwnedBonds` | Get owned bonds |
| GET   | `/api/v2/companies/me/bonds/sold/` | `Bond.handleGetSoldBonds` | Get sold bonds |
| POST   | `/api/v2/bonds/{bondId}/buy/` | `Bond.handleBuyBond` | Buy bond |

## Building

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| GET   | `/api/v2/buildings/market/` | `Building.handleBuildingMarket` | Building market |
| POST   | `/api/v2/buildings/buy/` | `Building.handleBuyBuilding` | Buy building |
| POST   | `/api/v2/buildings/place/` | `Building.handlePlaceBuilding` | Place building |
| POST   | `/api/v2/buildings/move/` | `Building.handleMoveBuilding` | Move building |
| POST   | `/api/v2/buildings/demolish/` | `Building.handleDemolishBuilding` | Demolish building |
| POST   | `/api/v2/buildings/stash/` | `Building.handleStashBuilding` | Stash building |
| POST   | `/api/v2/buildings/{buildingId}/stock/` | `Building.handleStockShelf` | Stock shelf |
| POST   | `/api/v2/buildings/{buildingId}/unstock/` | `Building.handleUnstockShelf` | Unstock shelf |
| POST   | `/api/v2/buildings/{buildingId}/shelf-price/` | `Building.handleSetShelfPrice` | Set shelf price |
| GET   | `/api/v2/companies/me/buildings/` | `Building.handleListMyBuildingsV2` | List my buildings v2 |
| POST   | `/api/v1/buildings/{buildingId}/upgrade/` | `Building.handleUpgradeBuilding` | Upgrade building |
| GET   | `/api/v3/companies/{companyId}/buildings/` | `Building.handleListCompanyBuildings` | List company buildings |
| GET   | `/api/v3/companies/me/buildings/` | `Building.handleListMyBuildings` | List my buildings v3 |

## Chat

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| POST   | `/api/v2/chat/room/` | `Chat.handleCreateRoom` | Create chat room |
| GET   | `/api/v2/chat/rooms/` | `Chat.handleListRooms` | List chat rooms |
| GET   | `/api/v2/chat/room/{roomId}/messages/` | `Chat.handleGetRoomMessages` | Get room messages |
| POST   | `/api/v2/chat/room/send/` | `Chat.handleSendMessage` | Send chat message |
| POST   | `/api/v2/chat/room/read/` | `Chat.handleMarkRead` | Mark chat read |

## Company

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| GET   | `/api/v2/players/me/companies/` | `Company.handleListMyCompanies` | List my companies |
| PATCH   | `/api/v2/companies/me/story-progress/` | `Company.handleUpdateStoryProgress` | Update story progress |
| POST   | `/api/v2/companies/me/tutorial/` | `Company.handleCompleteTutorial` | Complete tutorial |
| GET   | `/api/v3/companies/{companyId}/` | `Company.handleGetCompanyProfile` | Get company profile |

## Contract

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| GET   | `/api/v2/orders/daily/` | `Contract.handleListDailyOrders` | List daily orders |
| POST   | `/api/v2/orders/daily/complete/` | `Contract.handleCompleteDailyOrder` | Complete daily order |
| POST   | `/api/v2/orders/daily/claim/` | `Contract.handleClaimDailyOrder` | Claim daily order |
| GET   | `/api/v3/government-orders/` | `Contract.handleListGovContracts` | List government contracts |
| POST   | `/api/v3/government-orders/bid/` | `Contract.handleBidContract` | Bid on contract |

## Executive

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| POST   | `/api/v2/executives/search/` | `Executive.handleSearchExecutives` | Search executives |
| POST   | `/api/v2/executives/recruit/` | `Executive.handleRecruitExecutive` | Recruit executive |
| POST   | `/api/v2/executives/train/{executiveId}/` | `Executive.handleTrainExecutive` | Train executive |
| GET   | `/api/v3/executives/{id}/` | `Executive.handleGetExecutiveDetail` | Get executive detail |

## Finance

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| GET   | `/api/v2/companies/me/cashflow/recent/` | `Finance.handleRecentCashflow` | Recent cashflow |
| GET   | `/api/v2/companies/me/income-statement/` | `Finance.handleIncomeStatement` | Income statement |
| GET   | `/api/v2/companies/me/balance-sheet/` | `Finance.handleBalanceSheet` | Balance sheet |
| GET   | `/api/v2/companies/me/cashflow-statement/` | `Finance.handleCashflowStatement` | Cashflow statement |
| GET   | `/api/v3/companies/me/past-finances/` | `Finance.handlePastFinances` | Past finances |

## Leaderboard

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| GET   | `/api/v2/leaderboard/` | `Leaderboard.handleLeaderboard` | Get leaderboard |

## Market

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| POST   | `/api/v2/market-order/` | `Market.handleCreateOrder` | Create market order |
| DELETE   | `/api/v2/market-order/cancel/{orderId}/` | `Market.handleCancelOrder` | Cancel order |
| POST   | `/api/v2/market-order/take/` | `Market.handleTakeOrder` | Take/fulfill order |
| GET   | `/api/v3/resources/` | `Market.handleResources` | List resources |
| GET   | `/api/v3/market-ticker/{resourceId}/` | `Market.handleMarketTicker` | Market ticker |
| GET   | `/api/v3/market-depth/{resourceId}/{quality}/` | `Market.handleMarketDepth` | Market depth |
| GET   | `/api/v3/market/{resourceId}/{quality}/` | `Market.handleMarketOrders` | Market orders |
| GET   | `/api/v3/market-ticker/` | `Market.handleListTickers` | List all tickers |
| GET   | `/api/v3/companies/me/orders/` | `Market.handleListMyOrders` | List my orders |

## Player

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| GET   | `/api/v2/players/me/level/` | `Player.handleLevel` | Get player level |
| GET   | `/api/v2/players/simboosts/` | `Player.handleSimboostTypes` | List simboost types |
| GET   | `/api/v2/players/simboosts-use/` | `Player.handleSimboostsUse` | Get simboost usage |
| POST   | `/api/v2/players/simboosts-use/` | `Player.handleSimboostsUse` | Use simboost |
| POST   | `/api/v2/players/{playerId}/preferences/` | `Player.handleSavePreferences` | Save player preferences |

## Production

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| GET   | `/api/v2/production/jobs/` | `Production.handleListProductionJobs` | List production jobs |
| POST   | `/api/v2/production/start/` | `Production.handleStartProduction` | Start production |
| POST   | `/api/v2/production/claim/{jobId}/` | `Production.handleClaimProduction` | Claim production |
| GET   | `/api/v2/production/claimable/` | `Production.handleListClaimableJobs` | List claimable jobs |
| GET   | `/api/v2/production/queue/` | `Production.handleProductionQueue` | Production queue |
| POST   | `/api/v2/production/cancel/` | `Production.handleCancelProduction` | Cancel production |
| POST   | `/api/v2/production/claim-all/` | `Production.handleClaimAll` | Claim all production |
| GET   | `/api/v2/buildings/{buildingId}/production-options/` | `Production.handleProductionOptions` | Production options |
| POST   | `/api/v1/buildings/{buildingId}/busy/` | `Production.handleStartProductionV1` | Start production v1 |

## Report

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| POST   | `/api/v2/report/` | `Report.handleSubmitReport` | Submit report |

## Research

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| GET   | `/api/v2/research/` | `Research.handleListResearch` | List research |
| GET   | `/api/v2/research/progress/` | `Research.handleListResearch` | Research progress |
| POST   | `/api/v2/research/levelup/` | `Research.handleLevelUp` | Level up research |

## Social

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| GET   | `/api/messages/` | `Social.handleMessages` | Get messages |
| POST   | `/api/v2/message/` | `Social.handleV2Message` | Send v2 message |
| GET   | `/api/v2/message/{messageId}/read/` | `Social.handleMarkRead` | Mark message read |
| GET   | `/api/v2/chatroom/` | `Social.handleChatroom` | Get chatroom |
| GET   | `/api/v2/contacts/` | `Social.handleContacts` | Get contacts |

## Warehouse

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| GET   | `/api/v2/companies/me/warehouse/` | `Warehouse.handleGetMyWarehouse` | Get warehouse |
| POST   | `/api/v2/companies/me/warehouse/upgrade/` | `Warehouse.handleUpgradeWarehouse` | Upgrade warehouse |

## Admin (dev tool, no auth)

| 方法 | 路径 | Handler | 说明 |
|------|------|---------|------|
| POST   | `/api/admin/snapshot/save` | `Admin.handleSaveSnapshot` | Save snapshot |
| POST   | `/api/admin/snapshot/load` | `Admin.handleLoadSnapshot` | Load snapshot |

---

## 文件清单

| 文件 | 大小 | 主要 Handler |
|------|------|-------------|
| `auth_handler.go` | 2.4KB | AuthHandler: Register, Login |
| `auth_middleware.go` | 1.5KB | JWT auth middleware |
| `building_handler.go` | 8.4KB | BuildingHandler: buy, place, move, demolish, stock |
| `market_handler.go` | 6.1KB | MarketHandler: orders, tickers, depth, resources |
| `production_handler.go` | 5.6KB | ProductionHandler: start, claim, queue, cancel |
| `finance_handler.go` | 2.5KB | FinanceHandler: cashflow, income, balance sheet |
| `bond_handler.go` | 4.5KB | BondHandler: create, call, buy, settle |
| `company_handler.go` | 1.8KB | CompanyHandler: list, profile, story, tutorial |
| `company_profile_handler.go` | 1.8KB | CompanyProfileHandler (deprecated?) |
| `player_handler.go` | 3.4KB | PlayerHandler: level, simboosts, prefs |
| `social_handler.go` | 7.8KB | SocialHandler: messages, contacts, chatroom |
| `chat_handler.go` | 8.1KB | ChatHandler: rooms, send, read |
| `contract_handler.go` | 4.5KB | ContractHandler: daily orders, gov contracts |
| `research_handler.go` | 2.1KB | ResearchHandler: list, levelup |
| `executive_handler.go` | 7.9KB | ExecutiveHandler: search, recruit, train |
| `leaderboard_handler.go` | 2.7KB | LeaderboardHandler |
| `warehouse_handler.go` | 1.3KB | WarehouseHandler: get, upgrade |
| `admin_handler.go` | 1.2KB | AdminHandler: save/load snapshot |
| `report_handler.go` | 3.5KB | ReportHandler: submit report |
| `router.go` | 10.5KB | **Route registration (最佳入口点)** |
| `response.go` | 2.1KB | Standard response envelope + error codes |
| `errmap.go` | 2.1KB | Error code mapping |
| `ratelimit.go` | 2.2KB | Rate limiting middleware |
| `middleware.go` | 1.1KB | Logger, CORS middleware |
| `openapi-draft.yaml` | 53.2KB | Full OpenAPI 3.0 spec (2242 lines) |

---

## CLI 工具推荐

如果你在终端/VS Code终端里工作:

### 1. ripgrep (`rg`) — 搜索代码内容（比 grep 快 10x）
```bash
# 找一个 handler 的所有引用
rg "handleCreateOrder" heritage/httpapi/

# 搜错误码
rg "ErrorConflict" heritage/httpapi/

# 搜某个 DTO 字段
rg "BuildingName" heritage/ --type go

# 搜 openapi 里的 schema
rg "BuildingListResponse" heritage/openapi-draft.yaml
```

### 2. `bat` — 带语法高亮的 cat
```bash
bat heritage/httpapi/router.go
```

### 3. `yq` — 查询 OpenAPI YAML（比翻文件快）
```bash
# 列出所有路径
yq eval ".paths | keys" heritage/openapi-draft.yaml

# 查看某个路径的详情
yq eval '.paths.["/api/v3/resources/"]' heritage/openapi-draft.yaml

# 查询所有 schema 名字
yq eval ".components.schemas | keys" heritage/openapi-draft.yaml

# 查某个 schema 字段
yq eval '.components.schemas.BuildingListResponse' heritage/openapi-draft.yaml
```

### 4. `fzf` — 交互式模糊搜索 + 预览
```bash
# 交互式搜端点
rg "^    (get|post|put|patch|delete|options):" -B1 heritage/openapi-draft.yaml | fzf

# 交互式搜 handler 方法
rg "^func \(h \*" heritage/httpapi/ -n | fzf
```

### 5. `go doc` — 本地 Go 文档服务器
```bash
# 启动
go doc -http :6060 &
# 浏览器打开 http://localhost:6060/pkg/github.com/newhaven/backend-next/
```

### 6. `delta` — Git diff 高亮（配合 `bat` 用）

---

## 最佳实践总结

| 场景 | 做什么 | 打开什么 |
|------|--------|---------|
| 找一个端点 | `Ctrl+F` 搜路径 | `API-QUICKREF.md`（就是这个文件） |
| 看某个 Handler 实现了什么 | `read` 摘要模式 | 对应的 `xxx_handler.go` |
| 查请求/响应结构 | 搜 schema 名 | `openapi-draft.yaml` |
| 找 handler 被哪些路由引用 | `search`/`rg` | 搜 handler 方法名 |
| 理解代码流程 | `read` 展开具体行 | 从 `router.go` 开始跟 |
