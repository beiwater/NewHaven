# 当前基线报告 — NewHaven 项目

**日期**: 2026-06-06  
**范围**: 后端 Go 服务 + 前端 React/PixiJS 客户端  
**说明**: 本文件记录 Phase 1 开始时的项目状态，作为后续变更的对照基线。

---

## 1. 后端目录结构

```
backend/
├── cmd/simapi/main.go            # 入口 (~98 行)
├── internal/
│   ├── handler/     [22 files]    HTTP 路由层
│   │   ├── handler.go              # New(), Register(), withAuth(), companyID(), writeJSON(), writeErr()
│   │   ├── auth.go                 # /api/register, /api/login
│   │   ├── company.go              # /api/v2/companies/*, /api/v3/companies/*
│   │   ├── market.go               # /api/v3/market*, /api/v2/market-order*
│   │   ├── production.go           # /api/v1/buildings/*, /api/v2/production/*
│   │   ├── production_queue.go     # /api/v2/production/queue/*
│   │   ├── building_shop.go        # /api/v2/buildings/buy/, place/move/demolish
│   │   ├── bond.go                 # /api/bonds/*
│   │   ├── financial.go            # /api/v2/companies/me/income-statement, balance-sheet, etc.
│   │   ├── research.go             # research 路由
│   │   ├── dev.go                  # /api/dev/*, /api/v4/*, contract/research stubs
│   │   ├── government.go           # /api/v3/government-orders/*
│   │   ├── auction.go              # /api/v2/auctions/*
│   │   ├── executive.go            # /api/v2/executives/*
│   │   ├── order.go                # /api/v2/orders/*
│   │   ├── message.go              # /api/messages/*, /api/v2/message/*
│   │   ├── leaderboard.go          # /api/v2/leaderboard/*
│   │   ├── player.go               # /api/v2/players/*
│   │   ├── recipe.go               # /api/v2/recipes/
│   │   ├── aerospace.go            # /api/v2/aerospace/*
│   │   ├── health.go               # /healthz, /readyz
│   │   └── helpers.go              # helper 函数
│   ├── service/     [41 files]    业务逻辑层（最肥大）
│   │   ├── service.go               # Service struct, New(), ResourcesWithMarket()
│   │   ├── service_player.go        # RegisterPlayer, LoginPlayer
│   │   ├── company.go               # Company CRUD
│   │   ├── building.go              # upgrade, demolish
│   │   ├── building_shop.go         # BuildingMarket, BuyBuilding, place/move
│   │   ├── production.go            # Start/claim/cancel production + slot logic
│   │   ├── production_claim.go      # claim/claimAll
│   │   ├── recipe.go                # recipe lookups
│   │   ├── market_trade.go          # CreateOrder, CancelOrder, TakeOrder, matchLimitOrders
│   │   ├── market_match.go          # executeMatch (market order execution)
│   │   ├── market_info.go           # Ticker, chain price
│   │   ├── market_depth.go          # Orderbook depth
│   │   ├── market_competition.go    # Bot orders, market lock, national team intervention
│   │   ├── bond.go                  # IssueOrAdjustBond, BuyBond, CallBond, SettleBondInterest
│   │   ├── government.go            # PlaceGovernmentBid, Award, Deliver
│   │   ├── auction.go               # PlaceBid
│   │   ├── order.go                 # Daily orders
│   │   ├── research.go              # StartResearch, ResearchProgress, CompleteResearch
│   │   ├── research_level.go        # Research level unlocks
│   │   ├── service_stubs.go         # Executive, Aerospace stubs
│   │   ├── leaderboard.go           # Leaderboard computation
│   │   ├── state_snapshot.go        # Snapshot(), clone*
│   │   ├── service_save.go          # Save helpers
│   │   ├── simboost.go              # SimBoost activation
│   │   ├── offline.go               # Offline income
│   │   ├── cleanup.go               # Order cleanup
│   │   ├── ids.go                   # ID generation
│   │   ├── level_unlocks.go         # Level unlock rules
│   │   ├── test_helpers.go
│   │   ├── xp.go
│   │   └── *_test.go  [12 files]
│   ├── model/        [1 file]     model/types.go — 所有数据类型
│   ├── formula/      [8 files]    纯函数经济公式
│   │   ├── production.go            # OutputPerHour, ProductionTimeSeconds, BaseProductionRate
│   │   ├── market.go                # TickStep, IsValidTick, ExchangeFee
│   │   ├── retail.go                # UnitsSoldPerHour
│   │   ├── bonds.go                 # DailyBondInterest, MaxIssuableBonds, BondFaceValue
│   │   ├── costs.go                 # LaborCostPerHour, EnergyCostPerHour, etc.
│   │   ├── saturation.go            # SaturationPriceMultiplier, EffectivePrice
│   │   ├── admin.go                 # AdminOverheadWithCOO, CTOProductionMultiplier
│   │   └── formula_test.go          # 19KB, 25+ tests
│   ├── config/       [2 files]    config.go + config_test.go
│   ├── middleware/   [2 files]    middleware.go (Recovery, Logger, CORS, RequestID) + auth.go (JWT)
│   ├── storage/      [3 files]    storage.go (interface + NoopStorage) + postgres.go (pgx) + storage_test.go
│   ├── scheduler/    [1 file]     scheduler.go — 60s tick, GameService interface
│   ├── data/         [1 file]     loader.go — 静态 JSON 加载
│   ├── anticheat/    [n files]    反作弊
│   └── aml/          [n files]    反洗钱
├── configs/game.json               # 经济参数
├── go.mod                          # 依赖: pgx/v5 + bcrypt + 标准库
└── go.sum
```

## 2. 前端目录结构

```
client/atlas-foods-client/src/
├── app/
│   ├── App.tsx                   # 布局路由（无 react-router，基于 tab）
│   ├── ErrorBoundary.tsx
│   └── providers.tsx
├── api/           [14 files]     API 调用层
│   ├── client.ts                   # 自定义 fetch wrapper + JWT
│   ├── company.api.ts
│   ├── buildings.api.ts
│   ├── production.api.ts
│   ├── market.api.ts
│   ├── financial.api.ts
│   ├── research.api.ts
│   ├── executives.api.ts
│   ├── chat.api.ts
│   ├── contracts.api.ts
│   ├── leaderboard.api.ts
│   ├── powerup.api.ts
│   ├── inventory.api.ts
│   └── websocket.ts               # 空壳
├── store/         [2 files]
│   ├── ui.store.ts                 # Zustand — 仅 UI 状态
│   └── game.store.ts               # Zustand — zoom, pan, tick
├── features/      [18 目录]       功能模块
│   ├── auth/
│   ├── buildings/
│   ├── market/
│   ├── production/
│   ├── research/
│   ├── executives/
│   ├── financial/
│   ├── chat/
│   ├── contracts/
│   ├── leaderboard/
│   ├── inventory/
│   ├── sidebar/
│   ├── topbar/
│   ├── powerups/
│   ├── mobile/
│   ├── guidance/
│   ├── story/
│   ├── chain/
│   ├── inspect/
│   └── settings/
├── game/                          # PixiJS 渲染
│   ├── GameCanvas.tsx
│   ├── map.config.ts
│   ├── resources.ts
│   └── pixi/
└── audio/
```

## 3. 当前 Go 依赖

```
# backend/go.mod
module go-sim-api

go 1.25

require (
    github.com/jackc/pgx/v5      # PostgreSQL driver
    golang.org/x/crypto           # bcrypt
)
```

**注意**: go.mod 仅有 pgx 和 crypto 两个外部依赖。其余全部自造（router、middleware、auth、scheduler、storage abstraction）。

## 4. 当前前端依赖

```
# client/atlas-foods-client/package.json (from project tree)
    @tanstack/react-query           # 服务器状态管理
    @tanstack/react-query-devtools
    zustand                        # UI 状态管理
    pixi.js                        # 游戏画布渲染
    react (19.x)
    react-dom
    react-i18next + i18next        # 国际化
    tailwindcss + @tailwindcss/vite
    recharts                       # 图表（Leaderboard/Finance 用）
    clsx + tailwind-merge
    vite
    typescript (6.x 严格模式)
    eslint
```

## 5. 所有 API 路由清单

### 5.1 系统

| 方法 | 路径 | 文件 | 认证 |
|------|------|------|------|
| GET | `/healthz` | handler/health.go | 否 |
| GET | `/readyz` | handler/health.go | 否 |
| GET | `/api/dev/formulas/production` | handler/dev.go | 是 |
| GET | `/api/dev/formulas/retail` | handler/dev.go | 是 |
| GET | `/api/dev/formulas/retail/season-weather` | handler/dev.go | 是 |
| GET | `/api/dev/ledger` | handler/dev.go | 是 |
| POST | `/api/dev/time` | handler/dev.go | 是 |

### 5.2 认证

| 方法 | 路径 | 文件 | 认证 |
|------|------|------|------|
| POST | `/api/register` | handler/auth.go | 是 (withAuth 但内部不验证) |
| POST | `/api/login` | handler/auth.go | 是 (同上) |

### 5.3 公司

| 方法 | 路径 | 文件 | 认证 |
|------|------|------|------|
| GET | `/api/v2/companies/me/` | handler/company.go | 是 |
| GET | `/api/v2/companies/me/buildings/` | handler/company.go | 是 |
| GET | `/api/v2/companies/me/inventory/` | handler/company.go | 是 |
| GET | `/api/v3/companies/me/` | handler/company.go | 是 |
| GET | `/api/v3/companies/me/buildings/` | handler/company.go | 是 |
| GET | `/api/v3/companies/me/inventory/` | handler/company.go | 是 |
| PATCH | `/api/v2/companies/me/preferences/` | handler/company.go | 是 |

### 5.4 建筑

| 方法 | 路径 | 文件 | 认证 |
|------|------|------|------|
| GET | `/api/v2/buildings/buy/` | handler/building_shop.go | 是 |
| POST | `/api/v2/buildings/buy/` | handler/building_shop.go | 是 |
| POST | `/api/v2/buildings/place/` | handler/building_shop.go | 是 |
| POST | `/api/v2/buildings/move/` | handler/building_shop.go | 是 |
| POST | `/api/v2/buildings/demolish/` | handler/building_shop.go | 是 |
| POST | `/api/v2/buildings/warehouse-upgrade/` | handler/building_shop.go | 是 |
| POST | `/api/v1/buildings/:id/upgrade/` | handler/production.go | 是 |
| POST | `/api/v1/buildings/:id/busy/` | handler/production.go | 是 |

### 5.5 生产

| 方法 | 路径 | 文件 | 认证 |
|------|------|------|------|
| GET | `/api/v2/production/jobs/` | handler/production.go | 是 |
| GET | `/api/v2/production/claimable/` | handler/production.go | 是 |
| POST | `/api/v2/production/claim/` | handler/production.go | 是 |
| POST | `/api/v2/production/claim-all/` | handler/production.go | 是 |
| GET | `/api/v2/production/queue/` | handler/production_queue.go | 是 |
| POST | `/api/v2/production/queue/cancel/` | handler/production_queue.go | 是 |
| POST | `/api/v2/buildings/:id/production-options/` | handler/production.go | 是 |

### 5.6 市场

| 方法 | 路径 | 文件 | 认证 |
|------|------|------|------|
| GET | `/api/v3/market-ticker/` | handler/market.go | 是 |
| GET | `/api/v3/market-depth/` | handler/market.go | 是 |
| GET | `/api/v3/market-trades/` | handler/market.go | 是 |
| GET | `/api/v3/resources/` | handler/market.go | 是 |
| GET | `/api/v3/resources-info/` | handler/market.go | 是 |
| POST | `/api/v2/market-order/` | handler/market.go | 是 |
| POST | `/api/v2/market-order/cancel/` | handler/market.go | 是 |
| POST | `/api/v2/market-order/take/` | handler/market.go | 是 |
| GET | `/api/v2/market-order/my/` | handler/market.go | 是 |
| GET | `/api/v3/chain-price/` | handler/market.go | 是 |

### 5.7 金融

| 方法 | 路径 | 文件 | 认证 |
|------|------|------|------|
| GET | `/api/v2/companies/me/income-statement/` | handler/financial.go | 是 |
| GET | `/api/v2/companies/me/balance-sheet/` | handler/financial.go | 是 |
| GET | `/api/v2/companies/me/cashflow-statement/` | handler/financial.go | 是 |
| GET | `/api/v2/companies/me/cashflow/recent/` | handler/financial.go | 是 |
| GET | `/api/v2/companies/me/past-finances-overview/` | handler/financial.go | 是 |
| GET | `/api/v3/companies/me/past-finances/` | handler/financial.go | 是 |
| POST | `/api/bonds/issue/` | handler/bond.go | 是 |
| POST | `/api/bonds/buy/` | handler/bond.go | 是 |
| POST | `/api/bonds/call/` | handler/bond.go | 是 |
| POST | `/api/bonds/settle-interest/` | handler/bond.go | 是 |
| GET | `/api/bonds/market/` | handler/bond.go | 是 |
| GET | `/api/v2/bonds/rating-group/` | handler/bond.go | 是 |

### 5.8 研究

| 方法 | 路径 | 文件 | 认证 |
|------|------|------|------|
| GET | `/api/v3/research/` | handler/dev.go | 是 |
| POST | `/api/v3/research/start/` | handler/dev.go | 是 |
| GET | `/api/v3/research/progress/` | handler/dev.go | 是 |
| POST | `/api/v3/research/complete/` | handler/dev.go | 是 |

### 5.9 高管

| 方法 | 路径 | 文件 | 认证 |
|------|------|------|------|
| GET | `/api/v2/executives/market/` | handler/executive.go | 是 |
| GET | `/api/v2/executives/my/` | handler/executive.go | 是 |
| POST | `/api/v2/executives/recruit/` | handler/executive.go | 是 |
| POST | `/api/v2/executives/train/` | handler/executive.go | 是 |
| GET | `/api/v2/executives/detail/` | handler/executive.go | 是 |

### 5.10 政府合同

| 方法 | 路径 | 文件 | 认证 |
|------|------|------|------|
| GET | `/api/v3/government-orders/` | handler/government.go | 是 |
| POST | `/api/v3/government-orders/bid/` | handler/government.go | 是 |
| POST | `/api/v3/government-orders/deliver/` | handler/government.go | 是 |

### 5.11 其他 v4 路由（在 dev.go 中注册）

| 方法 | 路径 | 文件 | 认证 |
|------|------|------|------|
| GET | `/api/v4/contracts-incoming/` | handler/dev.go | 是 |
| GET | `/api/v4/contracts-outgoing/` | handler/dev.go | 是 |
| GET | `/api/v4/contracts-history-incoming/` | handler/dev.go | 是 |
| GET | `/api/v4/contracts-history-outgoing/` | handler/dev.go | 是 |
| GET | `/api/v4/warehouse-contracts-summary/` | handler/dev.go | 是 |

### 5.12 消息

| 方法 | 路径 | 文件 | 认证 |
|------|------|------|------|
| GET | `/api/messages/` | handler/message.go | 是 |
| POST | `/api/messages/send/` | handler/message.go | 是 |
| GET | `/api/v2/message/notifications/` | handler/message.go | 是 |

## 6. 前端调用 API 清单

| 前端文件 | 调用的 API 路径 |
|---------|----------------|
| api/company.api.ts | `/api/v2/companies/me/`, `/api/v2/companies/me/buildings/`, `/api/v2/companies/me/inventory/` |
| api/auth/authGate.ts | `/api/login`, `/api/register` |
| api/buildings.api.ts | `/api/v2/buildings/buy/`, `/api/v2/buildings/place/`, `/api/v2/buildings/move/`, `/api/v2/buildings/demolish/`, `/api/v2/buildings/warehouse-upgrade/` |
| api/production.api.ts | `/api/v2/production/jobs/`, `/api/v2/production/claim/`, `/api/v2/production/claim-all/`, `/api/v2/production/queue/`, `/api/v1/buildings/:id/upgrade/` |
| api/market.api.ts | `/api/v3/market-ticker/`, `/api/v3/market-depth/`, `/api/v2/market-order/`, `/api/v2/market-order/cancel/`, `/api/v2/market-order/take/`, `/api/v3/chain-price/` |
| api/financial.api.ts | `/api/v2/companies/me/income-statement/`, `/api/v2/companies/me/balance-sheet/`, `/api/v2/companies/me/cashflow-statement/`, `/api/v2/bonds/rating-group/`, `/api/bonds/market/`, `/api/v3/companies/me/past-finances/` |
| api/research.api.ts | `/api/v3/research/`, `/api/v3/research/start/`, `/api/v3/research/progress/`, `/api/v3/research/complete/` |
| api/executives.api.ts | `/api/v2/executives/market/`, `/api/v2/executives/my/`, `/api/v2/executives/recruit/`, `/api/v2/executives/train/`, `/api/v2/executives/detail/` |
| api/chat.api.ts | `/api/messages/`, `/api/messages/send/` |
| api/contracts.api.ts | `/api/v3/government-orders/`, `/api/v3/government-orders/bid/` |
| api/leaderboard.api.ts | `/api/v2/leaderboard/` |
| api/inventory.api.ts | Varies |

## 7. Handler 文件职责表

| 文件 | 路由前缀 | 职责 |
|------|---------|------|
| handler.go | — | New(), Register(), withAuth(), companyID(), writeJSON(), writeErr() |
| auth.go | `/api/register`, `/api/login` | 注册 + 登录 |
| company.go | `/api/v2/companies/me/*`, `/api/v3/companies/me/*` | 公司信息、建筑列表、库存、偏好 |
| market.go | `/api/v3/market*`, `/api/v2/market-order*`, `/api/v3/chain-price*` | 市场行情、订单、深度 |
| production.go | `/api/v1/buildings/*`, `/api/v2/production/*` | 生产操作、槽位 |
| production_queue.go | `/api/v2/production/queue/*` | 队列查看、取消 |
| building_shop.go | `/api/v2/buildings/*` | 购买、放置、移动、拆除、仓库升级 |
| bond.go | `/api/bonds/*`, `/api/v2/bonds/*` | 债券发行、购买、赎回、利息 |
| financial.go | `/api/v2/companies/me/*` | 财务报表、现金流 |
| dev.go | `/api/dev/*`, `/api/v4/*`, `/api/v3/research*`, `/api/v3/contract*` | 开发路由 + 部分生产路由 |
| government.go | `/api/v3/government-orders/*` | 政府订单竞标、交付 |
| auction.go | `/api/v2/auctions/*` | 竞拍 |
| executive.go | `/api/v2/executives/*` | 高管市场、招聘、培训 |
| order.go | `/api/v2/orders/*` | 日订单 |
| message.go | `/api/messages/*`, `/api/v2/message/*` | 消息、通知 |
| leaderboard.go | `/api/v2/leaderboard/*` | 排行榜 |
| player.go | `/api/v2/players/*` | 玩家信息 |
| recipe.go | `/api/v2/recipes/` | 配方查询 |
| aerospace.go | `/api/v2/aerospace/*` | 航空航天（stub） |
| health.go | `/healthz`, `/readyz` | 健康检查 |
| helpers.go | — | writeJSON, writeErr, parseID 等 |

## 8. Service 文件职责表

| 文件 | 行数(约) | 职责 |
|------|---------|------|
| service.go | 550+ | Service 结构体、New()、ResourcesWithMarket()、inventoryGet/Add/Sub、save*、helper 方法 |
| service_player.go | 140+ | RegisterPlayer、LoginPlayer、ValidateToken |
| company.go | 200+ | CompaniesByPlayer、UpdatePreferences |
| building.go | 200+ | UpgradeBuilding、DemolishBuilding |
| building_shop.go | 420+ | BuildingMarket、BuyBuilding、PlaceBuilding、MoveBuilding、WarehouseUpgrade |
| production.go | 350+ | StartBuildingProduction、productionOptions、refreshProductionJobs |
| production_claim.go | 100+ | ClaimProduction、ClaimAll |
| recipe.go | 100+ | 配方查询 |
| market_trade.go | 250+ | CreateOrder、CancelOrder、TakeOrder、matchLimitOrders |
| market_match.go | 80+ | executeMatch |
| market_info.go | 150+ | MarketTicker、ChainPrice |
| market_depth.go | 100+ | MarketDepth |
| market_competition.go | 350+ | Bot market making、market lock、national team |
| bond.go | 250+ | 债券发行、购买、赎回、利息结算 |
| government.go | 200+ | 政府订单竞标、授标、交付 |
| auction.go | 100+ | 竞拍 |
| order.go | 100+ | 日订单 |
| research.go | 200+ | 研究项目、进度、完成 |
| research_level.go | 80+ | 研究等级解锁 |
| service_stubs.go | 410+ | Executive、Aerospace 等 stub 方法 |
| leaderboard.go | 80+ | 排行榜计算 |
| state_snapshot.go | 140+ | Snapshot() 快照 + deep clone |
| service_save.go | 60+ | 持久化保存 |
| simboost.go | 60+ | SimBoost 激活 |
| offline.go | 60+ | 离线收入计算 |
| cleanup.go | 40+ | 订单清理 |
| ids.go | 40+ | ID 生成 |
| level_unlocks.go | 80+ | 等级解锁规则 |

## 9. Formula 文件职责表

| 文件 | 函数 | 职责 |
|------|------|------|
| production.go | BaseProductionRate, ProducedPerHour, ProductionDurationSeconds | 生产速率 + 时间计算 |
| market.go | TickStep, IsValidTick, ExchangeFee | 价格刻度、有效性、交易费 |
| retail.go | UnitsSoldPerHour | 零售销售模型 |
| bonds.go | DailyBondInterest, PeriodBondInterest, SetBondFaceValue, MaxIssuableBonds | 债券利息计算 |
| costs.go | LaborCostPerHour, EnergyCostPerHour, MaintenanceCostPerHour, ManagementCostPerHour, TaxPerHour, UpgradeCost, TotalBuildingCost | 运营成本 |
| saturation.go | SaturationPriceMultiplier, GroupOf, EffectivePrice | 市场饱和度价格修正 |
| admin.go | AdminOverheadWithCOO, CTOProductionMultiplier | 高管加成计算 |

**注意**: service/ 中还散落着 16 个公式（升级成本、仓库成本、生产时长、品质解析、离线收入、bot 定价等），未在 formula 包中。

## 10. Storage / Scheduler / Middleware 当前职责

### Storage

- `storage/storage.go`: `Storage` interface（SaveCompany, SaveOrders, SaveTrades, SaveState, LoadState, Close）+ `NoopStorage`（内存模式无持久化）
- `storage/postgres.go`: pgx 实现，使用 JSONB 存储
- 当前内存模式是默认，PostgreSQL 可选

### Scheduler

- `scheduler/scheduler.go`: 定时器每 60 秒执行一次 tick
- Tick 执行: SettleAllBonds → SaveAll → ResetDailyMarket → RunBotMarketCycle → AwardGovernmentContracts → ResolveGovernmentDefaults → SimBoostClean
- 使用 `GameService` interface，直接绑定 Service 方法名

### Middleware

- `middleware/middleware.go`: Recovery（panic 恢复）、Logger（请求日志）、CORS（跨域）、RequestID（请求 ID）
- `middleware/auth.go`: 自造 HS256 JWT sign/parse + CSRF token 生成

## 11. 当前自造基础设施清单

| 自造组件 | 文件 | 说明 | 严重度 |
|---------|------|------|--------|
| HTTP 路由注册 | `handler/handler.go:Register()` + 所有 RegisterXxx | 纯手工注册，无路由分组 | ⛔ Critical |
| 手工路径参数解析 | 每个 handler 中 `strings.TrimPrefix` + `strings.Split` | 路径参数靠手工切 | ⛔ Critical |
| API 响应序列化 | `handler/handler.go:writeJSON()` 自行调用 `json.Marshal`+ `w.Write` | 约 15 行函数，但缺乏统一的 envelope | ⚠️ High |
| 错误响应格式 | `handler/handler.go:writeErr()` 调用 `http.Error` | 纯文本错误，无结构化的 code/message/details | ⚠️ High |
| JWT 认证 | `middleware/auth.go` 中自造 HS256 sign/parse | 60+ 行自造实现 | ⚠️ Medium |
| 存储抽象 | `storage/storage.go` interface + NoopStorage + pgx | 存在但接口粒度粗（全量 Save 为主） | ⚠️ Medium |
| Scheduler | `scheduler/scheduler.go` 自造定时器 | 简单可用，但缺乏事件驱动 | ⚠️ Medium |
| WebSocket | 无实现（前端 ws.ts 空壳 + 后端无 WS 端点） | 全部未实现 | ⚠️ Medium |
| Service 返回 map[string]any | 几乎所有 service 方法返回 `map[string]any` | 无类型安全 | ⛔ Critical |
| 前端 API client | `api/client.ts` 自造 fetch wrapper | 91 行，功能完整但少类型推断 | ⚠️ Low-Med |
| 数据加载器 | `data/loader.go` 使用 `map[string]any` | 无类型安全 | ⚠️ Medium |
| ID 生成 | `service/ids.go` 基于 nanosecond timestamp | 简单可用 | ✅ Low |
| 静态数据加载 | `data/loader.go` 直接 `os.ReadFile` + `json.Unmarshal` | 无 schema validation | ⚠️ Medium |
| Middleware 链 | `middleware/middleware.go` 自造 | Nesting 风格，可读性尚可 | ✅ Low |

---

*本基线在 Phase 1 结束时审核，后续变更必须对比此基线确认行为未改变。*
