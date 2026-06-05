# NewHaven 框架重构与开源库选型检测报告

**日期**: 2026-06-06  
**版本**: 1.0  
**范围**: 后端 Go 服务 (backend/) + 前端 React/PixiJS 客户端 (client/)

---

## 1. 执行摘要

**是否应该换框架**: 部分应该。

- **HTTP 路由层应该升级**：当前 `net/http ServeMux` + 手工路径切割 (`strings.TrimPrefix`, `strings.Split`) 是最大的技术债来源。建议迁移到 **go-chi/chi**。
- **业务逻辑不需要重写框架**：Service 层和 Formula 层的核心代码质量尚可，不需要推到重来。
- **基础设施需要换开源库**：认证、日志、配置、调度、WebSocket 应迁移到成熟库。
- **替换路径是渐进的**：先替换路由层，再引入 OpenAPI 契约，再切分 Service 层，再补充前端 API 层。

核心诊断：项目跑得起来、能玩，但 70% 的工程成本花在了自己写的薄基础设施上。这些基础设施每个单独看问题不大，但加在一起让每一次增加功能都变得很重。

---

## 2. 当前项目结构

### 2.1 后端目录结构

```
backend/
├── cmd/simapi/                # Entry point (~100 lines)
│   └── main.go                # net/http ServeMux wiring
├── internal/
│   ├── handler/               # 22 files - HTTP handlers
│   │   ├── handler.go         # Handler struct, Register, withAuth, helpers
│   │   ├── auth.go            # /api/register, /api/login
│   │   ├── company.go         # /api/v2/companies/*, /api/v3/companies/*
│   │   ├── market.go          # /api/v3/market*, /api/v2/market-order*
│   │   ├── production.go      # /api/v1/buildings/*, /api/v2/production/*
│   │   ├── production_queue.go
│   │   ├── building_shop.go   # /api/v2/buildings/buy/, place, move, demolish
│   │   ├── bond.go            # /api/bonds/*
│   │   ├── financial.go       # /api/v2/companies/me/income-statement etc.
│   │   ├── research.go        # (in dev.go currently)
│   │   ├── dev.go             # /api/dev/*, /api/v4/*, contract/research exec stubs
│   │   ├── government.go      # /api/v3/government-orders/*
│   │   ├── auction.go         # /api/v2/auctions/*
│   │   ├── executive.go       # /api/v2/executives/*
│   │   ├── order.go           # /api/v2/orders/*
│   │   ├── message.go         # /api/messages/*, /api/v2/message/*
│   │   ├── leaderboard.go     # /api/v2/leaderboard/
│   │   ├── player.go          # /api/v2/players/*
│   │   ├── recipe.go          # /api/v2/recipes/
│   │   ├── aerospace.go       # /api/v2/aerospace/*
│   │   ├── health.go          # /healthz, /readyz
│   │   └── helpers.go
│   ├── service/               # 41 files - business logic (fattest layer)
│   │   ├── service.go         # Service struct, New(), helpers
│   │   ├── service_player.go  # Register/Login player
│   │   ├── company.go         # Company CRUD
│   │   ├── building.go        # Building upgrade/demolish
│   │   ├── building_shop.go   # Building market/place/move
│   │   ├── production.go      # Start/claim/cancel production
│   │   ├── production_claim.go
│   │   ├── recipe.go          # Recipe lookups
│   │   ├── market_trade.go    # Create/cancel/take + matchLimitOrders
│   │   ├── market_match.go    # TakeOrder (market order)
│   │   ├── market_info.go     # Ticker, chain price
│   │   ├── market_depth.go    # Orderbook depth
│   │   ├── market_competition.go # Bot orders, market lock, national team
│   │   ├── bond.go            # Bond issue/buy/call/settle
│   │   ├── government.go      # Gov contract bid/award/deliver
│   │   ├── auction.go         # Auction bid
│   │   ├── order.go           # Daily orders
│   │   ├── research.go        # Research start/progress/complete
│   │   ├── research_level.go
│   │   ├── service_stubs.go   # Executive, Aerospace stubs (mock data)
│   │   ├── leaderboard.go     # Leaderboard computation
│   │   ├── state_snapshot.go  # Snapshot(), clone helpers
│   │   ├── service_save.go    # Save helpers
│   │   ├── simboost.go        # SimBoost activation
│   │   ├── offline.go         # Offline income
│   │   ├── cleanup.go         # Order cleanup
│   │   ├── ids.go             # ID generation
│   │   ├── level_unlocks.go   # Level unlock rules
│   │   ├── test files         # 12 test files (auth_test, bond_test, etc.)
│   ├── formula/               # 8 files - pure economic functions ✅
│   │   ├── production.go      # OutputPerHour, ProductionDurationSeconds
│   │   ├── market.go          # TickStep, IsValidTick, ExchangeFee
│   │   ├── retail.go          # UnitsSoldPerHour (complex retail model)
│   │   ├── bonds.go           # DailyBondInterest, MaxIssuableBonds
│   │   ├── costs.go           # Labor/Energy/Maintenance/Management/Tax cost
│   │   ├── saturation.go      # Saturation price multiplier
│   │   ├── admin.go           # AdminOverheadWithCOO, CTOProductionMultiplier
│   │   └── formula_test.go    # 19KB comprehensive tests ✅
│   ├── model/
│   │   └── types.go           # All data types + GameState
│   ├── config/
│   │   ├── config.go          # Env config + game.json loader
│   │   └── config_test.go
│   ├── middleware/
│   │   ├── middleware.go       # Recovery, Logger, CORS, RequestID
│   │   └── auth.go            # JWT sign/parse (HS256), CSRF token
│   ├── storage/
│   │   ├── storage.go          # Storage interface + NoopStorage
│   │   ├── postgres.go         # pgx implementation
│   │   └── storage_test.go
│   ├── scheduler/
│   │   └── scheduler.go        # 60s tick, GameService interface
│   ├── data/
│   │   └── loader.go           # Static JSON loader (resources, buildings, etc.)
│   ├── anticheat/
│   │   └── (detection logic)
│   └── aml/
│       └── (anti-money-laundering)
├── configs/
│   └── game.json               # Economy tuning
└── go.mod                      # Only dep: pgx/v5 (others built from stdlib)
```

### 2.2 前端目录结构

```
client/atlas-foods-client/src/
├── app/
│   ├── App.tsx                 # Layout routing (no react-router, tab-based)
│   ├── ErrorBoundary.tsx
│   └── providers.tsx
├── api/
│   ├── client.ts               # Custom fetch wrapper + JWT storage
│   ├── company.api.ts           # TanStack Query hooks
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
│   └── websocket.ts            # Empty stub
├── store/
│   ├── ui.store.ts              # Zustand - activeView, selectedBuilding, etc.
│   └── game.store.ts            # Zustand - zoom, pan, tick
├── features/                    # 18 feature directories
│   ├── auth/                    # AuthGate, login/register
│   ├── buildings/               # BuildView, BuildingCard, BuildingPanel
│   ├── market/                  # MarketPage, MarketTicker, PriceCurve
│   ├── production/              # ProductionQueue
│   ├── research/                # ResearchPage (new, complete)
│   ├── executives/              # ExecutivePage, ExecutiveMarket, etc.
│   ├── financial/               # FinancialPage (new, complete)
│   ├── chat/                    # ChatPanel (uses mock → needs real API)
│   ├── contracts/               # ContractList
│   ├── leaderboard/             # LeaderboardPage
│   ├── inventory/               # InventoryBar
│   ├── sidebar/                 # LeftSidebar
│   ├── topbar/                  # TopBar
│   ├── powerups/                # PowerPanel
│   ├── mobile/                  # MobileLayout
│   ├── guidance/                # FarmNotes
│   ├── story/                   # StoryGate
│   ├── chain/                   # SupplyChainPage
│   ├── inspect/                 # InspectPage
│   └── settings/                # SettingsPage
├── game/
│   ├── GameCanvas.tsx            # PixiJS canvas (lazy-loaded)
│   ├── map.config.ts
│   ├── types.ts
│   ├── resources.ts
│   └── pixi/                    # Pixi rendering
└── audio/                       # Audio system
```

### 2.3 数据与文档

```
docs/
├── requirements.md               # 完整需求规格 (1423 lines)
├── 2026-06-02/
│   ├── dev-plan.md               # 前后端补齐计划
│   ├── game-formulas-v2.md       # 公式文档 (780 lines)
│   ├── food-chain-simplify.md    # 食品链简化方案
│   ├── project-overview.md       # (可能不存在, 内容在 requirements.md)
│   └── ...
└── 2026-06-06/
    └── backend-refactor/
        ├── api-route-inventory.md  # API 路由清单
        ├── backend-inventory.md    # 后端代码清单
        ├── resource-baseline.md    # 资源基线
        ├── bridge-routing-plan.md  # 桥接路由方案
        ├── backend-next-plan.md    # 下一代后端计划
        └── prototypes/
            └── 00-target-architecture.md

decompiled/data/
├── resources.json           # 资源定义 (19 种)
├── buildings.json           # 建筑定义 (12 种)
├── economy_model.json       # 经济模型参数
└── resource_lookups.json    # 资源查询映射
```

---

## 3. 主要屎山来源

| # | 问题 | 严重度 | 证据 | 影响 | 换库解决？ | 架构切分？ | 优先级 |
|---|------|--------|------|------|-----------|-----------|--------|
| A | **HTTP 路由层手工解析路径** | **Critical** | `handler/handler.go:62-81` `Register()` 用 `HandleFunc` + 每个 handler 自己 `strings.TrimPrefix` + `strings.Split` 解析路径参数。示例：`production.go:18-26` `handleV2Buildings` 手工切 path parts 判断 action | 新增一个路由需要手写路径片段提取 + 字符串常量比对；路径参数极易冲突；无法声明式定义中间件作用域 | ✅ chi 替换 | ✅ | **P0** |
| B | **handler 返回 `map[string]any` 而不是 DTO** | **High** | 几乎所有 service 方法返回 `map[string]any`。示例 `service/market_trade.go:70` `return map[string]any{"order": order}, nil`。前端无法获得类型保证，后端改字段名前端无声断裂 | 前后端契约全靠默契；JSON 字段名变更无编译检查；重构恐惧症 | ✅ OpenAPI + oapi-codegen | ✅ | **P0** |
| C | **service 层过大, 41 个文件混合多个职责** | **High** | `service/market_trade.go` 包含：订单创建、取消、吃单、限价撮合、实际成交执行、市场数据更新。`service/production.go` 包含：生产启动、配方查找、品质解析、输入扣除、时间计算、刷新、队列、槽位、取消、退款 | 一个 service 改了可能影响所有领域；测试需要 setup 整个 GameState | ❌ | ✅ DDD 切分 | **P0** |
| D | **API 版本混乱** | **High** | 混合了 `/api`, `/api/v1`, `/api/v2`, `/api/v3`, `/api/v4`。同一领域有多个版本：公司 GET 有 `/api/v2/companies/` 和 `/api/v3/companies/`。handler/dev.go 包含生产环境路由 | 新人无法判断该用哪个版本；弃用无法清理因前端可能还在用 | ❌ | ✅ 先清单后合并 | **P1** |
| E | **DTO 和 Model 混用** | **High** | `model/types.go` 定义了 Company、MarketOrder 等 domain 类型。handler 层直接序列化这些 model 返回给前端（`company.go:39` `writeJSON(w, 200, h.svc.CompaniesByPlayer(...))`）。前端拿到的是 domain 模型字段 | 后端改 domain 字段名 = 改 API 响应；无法做版本兼容；domain 加了内部字段就会泄漏 | ❌ | ✅ DTO 层隔离 | **P1** |
| F | **存储层耦合在 Service 内** | **Medium** | `service.go` 的 `saveOrdersLocked()`、`saveCompanyLocked()` 等方法直接在业务逻辑中穿插持久化；`SaveAll()` 在 scheduler 中每 60s 调一次 | 业务逻辑修改时还要考虑持久化副作用；内存模式和 Postgres 模式的 save 路径不统一 | ❌ | ✅ 仓储模式 | **P1** |
| G | **认证体系自造且脆弱** | **Medium** | `middleware/auth.go`：自己实现 HS256 JWT sign/parse（约 60 行）、自己 base64 编码 header；`service_player.go` 用 `bcrypt` 存密码但 token 验证在 middleware 层 | 无 refresh token、无 JWT 标准库验证（如过期时间是否被正确校验）、CSRF token 生成但未使用 | ✅ 换标准库 | ✅ | **P1** |
| H | **公式散落在 service 和 formula** | **Medium** | `game-formulas-v2.md` 列出的 26 个公式中：10 个在 `formula/`、16 个在 `service/`（`building.go`、`production.go`、`market_competition.go` 等） | 修改经济平衡时需要同时改 service 和 formula；公式没有统一注册表 | ❌ | ✅ 全部移入 formula | **P1** |
| I | **前端 API client 自造** | **Low-Medium** | `api/client.ts`：自己写 fetch wrapper、自己处理 401 清除 auth、自己解析 error body。约 91 行，功能完整但缺少拦截器、请求取消、类型推断 | 已使用 TanStack Query 做 hooks，但 client 层没有标准的 retry、timeout、abort 控制 | ✅ 可用标准 fetch / 或 axios | ❌ | **P2** |
| J | **测试覆盖严重不足** | **High** | formula: 19KB 测试 ✅；service: 12 个测试文件 ✅（auth、bond、market、production 等）；handler: 1 个测试文件（handler_test.go 3.7KB）❌ | Handler 层几乎无测试意味着重构路由层时没有安全网；API 变更只能靠手动测试 | ❌ | ✅ 先加 contract test | **P1** |
| K | **WebSocket 是空壳** | **Low-Medium** | 前端 `api/websocket.ts` 全是空函数；后端无 WS handler | 生产完成不能实时推送到前端；市场成交不能实时更新 ticker；聊天靠轮询 | ✅ 需要 ws 库 | ✅ | **P2** |
| L | **Scheduler 直接依赖 Service 实现** | **Low-Medium** | `scheduler/scheduler.go:10-22` 定义了 `GameService` interface，但 interface 方法是 Service 的具体方法名（`SettleAllBonds`、`RunBotMarketCycle` 等），不是事件驱动的 | Scheduler 增加 tick 动作时需要改 interface；无法独立测试 scheduler | ❌ | ✅ 事件总线 | **P2** |
| M | **静态数据用 `map[string]any`** | **Medium** | `data/loader.go:10-15` `StaticData` 的 Resources、Buildings 等都是 `map[string]any`，无 typed struct | 所有消费方都要手动类型断言；改 JSON 字段无声断裂 | ❌ | ✅ 生成类型 | **P1** |
| N | **handler/dev.go 混合生产和开发路由** | **Medium** | `dev.go` 注册了 `/api/v4/` 生产路由、`/api/v3/contracts-incoming/`、research 路由等 | 无法明确区分哪些是 dev-only 哪些是生产 API | ❌ | ✅ 按域名拆分 | **P1** |

---

## 4. 是否应该换框架

**判断：部分应该，但不应全换。**

### 应该立即换的
1. **HTTP Router**：`net/http ServeMux` 手工路径解析 → **chi**。这是当前最大的工程效率瓶颈。标准库 ServeMux 在 Go 1.22+ 支持了路径参数，但全局中间件设计、子路由挂载、分组中间件等能力仍然缺失。chi 是标准库兼容方案，可以渐进替换。
2. **OpenAPI 契约**：没有契约 → **oapi-codegen + openapi-typescript**。前后端对接靠读代码和肉眼对比，必须解决。
3. **WebSocket**：空壳 → **coder/websocket**。gorilla/websocket 已归档，coder/websocket 是活跃的后续维护分支。
4. **Validation**：无请求校验 → **go-playground/validator** 或 **ozzo-validation**。
5. **Logging**：`log.Printf` → **slog** (标准库, Go 1.21+ 自带，无需引入外部依赖)。

### 不应该换的
1. **Service 层架构**：不需要全盘重写。当前问题在**边界不清晰**而不是**代码质量差**。切分 use case、引入 interface 隔离即可。
2. **Formula 层**：质量最好，只改导入路径不改逻辑。
3. **前端整体架构**：TanStack Query + Zustand + PixiJS 已经是最佳组合。只需要补充 WebSocket client 管理。
4. **数据库驱动**：`pgx/v5` 已经是 Go 生态最好的 PostgreSQL driver，无需更换。
5. **不要引入重框架**：不需要 Gin 的重中间件体系、不需要 Fiber 的 fasthttp 兼容性损失、不需要全栈框架。

### 为什么 chi 而不是 Gin/Fiber/Echo

| 维度 | chi | Gin | Fiber | Echo |
|------|-----|-----|-------|------|
| 与现有代码兼容性 | ✅ 直接包装 `http.Handler` | ❌ 需要 `gin.Context` 重写 | ❌ fasthttp 不兼容 `http.Handler` | ❌ `echo.Context` 重写 |
| 迁移成本 | **低** | **高** | **最高** | **高** |
| 可渐进替换 | ✅ 逐个 route group 迁移 | ❌ 一次性切 | ❌ 一次性切 | ❌ 一次性切 |
| 性能需求匹配 | ✅ 足够 | ✅ 但不需要 | 过剩 | ✅ |
| 社区活跃度 | ✅ 活跃 | ✅ 最活跃 | ✅ 活跃 | ⚠️ 中等 |

关键决策点：**chi 直接接受标准 `http.Handler`**，意味着现有 handler 函数签名 `func(w http.ResponseWriter, r *http.Request)` 一个字都不用改。只需要改 `Register()` 方法从 `mux.HandleFunc` → `chi.NewRouter().HandleFunc` 并增加中间件链。

---

## 5. 框架候选对比

| 方案 | 许可证 | 生态 | net/http 兼容 | 中间件生态 | OpenAPI 集成 | WS 集成 | 测试难度 | 迁移成本 | 推荐度 |
|------|--------|------|--------------|--------|-------------|---------|---------|---------|--------|
| 保留 `net/http ServeMux` | MIT (Go stdlib) | 标准 | - | 只支持全局 | 手动 | 手动 | 低 | 零 | **不推荐** (已触天花板) |
| **go-chi/chi v5** | MIT | 活跃 | ✅ 完全兼容 | ✅ 丰富 | ✅ chi-swagger | ✅ 无侵入 | 低 | 低 | **✅ 强烈推荐** |
| Gin | MIT | 最大 | ❌ gin.Context | ✅ gin middlewares | ✅ 多个 | ✅ gin-gorilla/ws | 中 | 高 | ❌ 不推荐 |
| Fiber | MIT | 大 | ❌ fasthttp | ✅ fiber middlewares | ⚠️ 有限 | ✅ | 中 | 最高 | ❌ 不推荐 |
| Echo | MIT | 中 | ❌ echo.Context | ✅ echo middlewares | ✅ | ✅ | 中 | 高 | ⚠️ 可选但不如 chi |

**为什么 chi 胜出**：

1. `chi.Router` 实现 `http.Handler` 接口 → 现有 `func(w, r)` handler 直接可用
2. 支持子路由挂载：`r.Route("/api/v2", func(r chi.Router) { ... })` → 可以让旧版 API 和新版 API 共存
3. 中间件作用域精确到路由组
4. URL 路径参数：`r.Get("/companies/{id}/", handler)` 替代手工 TrimPrefix
5. 零外部依赖，chi 自己纯标准库实现
6. Go 1.25 的 `net/http` 自带路由参数是 `ServeMux` 的功能增强，但仍然是扁平的全局注册模式，没有 chi 的分组中间件能力

---

## 6. 推荐技术栈

### 后端最终推荐

| 领域 | 方案 | 替换对象 | 等级 |
|------|------|---------|------|
| HTTP Router | **go-chi/chi v5** | 手工 path parsing | **Must** |
| API Contract | **OpenAPI 3.1 + oapi-codegen** | 无契约 | **Must** |
| DB Driver | **pgx/v5** (已有) | - | **Keep** |
| SQL 生成 | **sqlc** (后续) | 手工 SQL | **Should** |
| Migration | **golang-migrate/migrate** 或 **pressly/goose** | 手工建表 | **Should** |
| WebSocket | **github.com/coder/websocket** (原 nhooyr) | 空壳 | **Must** |
| Validation | **go-playground/validator v10** | 无校验 | **Should** |
| Logging | **log/slog** (stdlib Go 1.21+) | log.Printf | **Must** |
| Config | **保持当前方案** (env + game.json) | - | **Keep** |
| Metrics | **prometheus/client_golang** | 无 metrics | **Optional** |
| Background jobs | **保持当前 scheduler + interface** | - | **Keep** (后续可 EventBus) |
| ID generation | **sony/sonyflake** 或 保持 nanosecond | 当前 ID 生成 | **Optional** |
| Clock abstraction | **自写 `now()` interface** | 全局 time.Now | **Should** |
| Testing | **testify** (已有间接使用) | - | **Keep** |
| Event Bus | **自写轻量 EventBus**（后面需要时引入） | 无 | **Optional** |

### 前端最终推荐

| 领域 | 方案 | 替换对象 | 等级 |
|------|------|---------|------|
| API 类型生成 | **openapi-typescript**  (从 OpenAPI spec 生成) | 手写 .api.ts | **Must** |
| Server State | **TanStack Query** (已有) | - | **Keep** |
| UI State | **Zustand** (已有) | - | **Keep** |
| Game Rendering | **PixiJS** (已有) | - | **Keep** |
| UI Components | **Radix UI** + **shadcn/ui** | 自建组件 | **Should** |
| Charts | **Lightweight Charts** (TradingView) 或 **Recharts** | - | **Should** |
| WS Client | **@tanstack/react-query** 的 WS 集成 或 **自建 hook** | 空壳 | **Must** |
| Form/Validation | **React Hook Form** + **Zod** | 手写表单 | **Optional** |
| Table | **@tanstack/react-table** | 手写表格 | **Optional** |

---

## 7. 开源库清单

### 后端库

| 名称 | 许可证 | GitHub | 活跃度 | 替换 | 迁移成本 | 推荐 |
|------|--------|--------|--------|------|---------|------|
| **chi v5** | MIT | go-chi/chi | ⭐ 19k+ 活跃 | 手工 path parsing | **Low** | **Must** |
| **oapi-codegen** | MIT | deepmap/oapi-codegen | ⭐ 6k+ 活跃 | 无 API 契约 | **Medium** | **Must** |
| **coder/websocket** | MIT | coder/websocket | ⭐ 5k+ (gorilla 后继) | 空壳 WS | **Low** | **Must** |
| **golang-migrate** | MIT | golang-migrate/migrate | ⭐ 16k+ 活跃 | 手工建表 | **Low** | **Should** |
| **sqlc** | MIT | sqlc-dev/sqlc | ⭐ 14k+ 活跃 | 手工 SQL | **Medium** | **Should** |
| **go-playground/validator** | MIT | go-playground/validator | ⭐ 17k+ 活跃 | 无校验 | **Low** | **Should** |
| **prometheus/client_golang** | Apache-2.0 | prometheus/client_golang | ⭐ 5k+ 活跃 | 无 metrics | **Low** | **Should** |
| **testify** | MIT | stretchr/testify | ⭐ 24k+ 活跃 | 已有间接用 | **Low** | **Keep** |
| **sony/sonyflake** | MIT | sony/sonyflake | ⭐ 3k+ | ID 生成 | **Low** | **Optional** |
| **open-telemetry/opentelemetry-go** | Apache-2.0 | open-telemetry/opentelemetry-go | ⭐ 5k+ 活跃 | 无 tracing | **High** | **Optional** |
| **koanf** | MIT | knadh/koanf | ⭐ 3k+ | config 增强 | **Low** | **Optional** |

### 前端库

| 名称 | 许可证 | GitHub | 活跃度 | 替换 | 迁移成本 | 推荐 |
|------|--------|--------|--------|------|---------|------|
| **openapi-typescript** | MIT | openapi-typescript/openapi-typescript | ⭐ 4k+ 活跃 | 手写 API 类型 | **Low** | **Must** |
| **@tanstack/react-query** (已有) | MIT | TanStack/query | ⭐ 43k+ | ✅ 已有 | **Keep** | **Keep** |
| **zustand** (已有) | MIT | pmndrs/zustand | ⭐ 50k+ | ✅ 已有 | **Keep** | **Keep** |
| **pixi.js** (已有) | MIT | pixijs/pixijs | ⭐ 45k+ | ✅ 已有 | **Keep** | **Keep** |
| **shadcn/ui** | MIT | shadcn-ui/ui | ⭐ 82k+ 活跃 | 自建 UI 组件 | **Medium** | **Should** |
| **tailwindcss** (已有) | MIT | tailwindlabs/tailwindcss | ⭐ 86k+ | ✅ 已有 | **Keep** | **Keep** |
| **recharts** | MIT | recharts/recharts | ⭐ 25k+ | - | **Low** | **Should** |
| **@tanstack/react-table** | MIT | TanStack/table | ⭐ 26k+ | 手写表格 | **Low** | **Optional** |
| **react-hook-form** | MIT | react-hook-form/react-hook-form | ⭐ 42k+ | 手写表单 | **Low** | **Optional** |
| **zod** | MIT | colinhacks/zod | ⭐ 35k+ | - | **Low** | **Optional** |

---

## 8. 目标架构图

目标是逐步演进到以下结构，当前代码不改文件名，仅逐步添加新包和 facade。

### 后端目标（渐进式，非一次性）

```
backend/
├── cmd/simapi/                    # 保持不变，只替换路由注册
├── internal/
│   ├── app/                        # [新建] Use case / application service
│   │   ├── auth/
│   │   ├── company/
│   │   ├── production/
│   │   ├── market/
│   │   ├── finance/
│   │   ├── research/
│   │   ├── social/
│   │   └── system/
│   │
│   ├── domain/                     # [新建] 纯 domain 逻辑（从 service 抽出）
│   │   ├── company/                # Company entity + rules
│   │   ├── production/             # ProductionJob entity + rules
│   │   ├── market/                 # MarketOrder, Trade, matching engine
│   │   ├── finance/                # Bond, Ledger
│   │   ├── research/               # ResearchProject
│   │   ├── building/               # Building, placement rules
│   │   └── formula/                # 从当前 formula/ 移入（保持纯函数）
│   │
│   ├── handler/                    # [保留兼容] 当前 handler 作为 facade
│   │   ├── auth.go                 # 仅路由转发到 app/auth
│   │   ├── company.go              # 仅路由转发到 app/company
│   │   ├── … (22 个文件保持不变)     # 每个 handler eventually 变薄
│   │   └── helpers.go
│   │
│   ├── httpapi/                    # [新建] 新 handler（OpenAPI 生成的）
│   │   ├── middleware/             # chi middleware 链
│   │   ├── dto/                    # OpenAPI 生成的请求/响应类型
│   │   ├── v1/                     # 旧版兼容 adapter
│   │   ├── v2/                     # 旧版兼容 adapter
│   │   └── v3/                     # 目标 API 版本
│   │
│   ├── service/                    # [保留] 当前 service 慢慢变薄
│   │   └── (41 个文件逐步瘦身)
│   │
│   ├── storage/
│   │   ├── storage.go              # interface
│   │   ├── memory/                 # [新建] 内存存储（当前 NoopStorage 升级）
│   │   ├── postgres/               # [新建] pgx 实现（当前 postgres.go 移入）
│   │   └── migration/              # [新建] golang-migrate scripts
│   │
│   ├── platform/                   # [新建] 运行时抽象
│   │   ├── clock.go                # time.Now abstraction
│   │   ├── ids.go                  # ID generation
│   │   └── logger.go               # slog wrapper
│   │
│   ├── bridge/                     # [已有] 桥接路由方案
│   │   ├── client.go               # 转发到 backend-next
│   │   └── compare.go              # shadow 比较
│   │
│   ├── config/                     # [保留]
│   ├── middleware/                 # [保留后合并到 httpapi/middleware]
│   ├── data/                       # [保留后加 typed struct]
│   ├── anticheat/                  # [保留]
│   └── aml/                        # [保留]
│
└── scheduler/                      # [保留，但 event-driven 化]
```

### 前端目标

```
src/
├── api/
│   ├── client.ts                   # 保持（或换 axios）
│   ├── generated/                  # [新建] openapi-typescript 生成
│   │   ├── types.ts
│   │   ├── auth.ts
│   │   ├── company.ts
│   │   ├── production.ts
│   │   └── ...
│   └── websocket.ts                # [重写] 真实 WS 连接
├── store/                          # 保持 zustand
├── features/                       # 保持，已对接真实 API
└── game/                           # 保持
```

---

## 9. 分阶段迁移计划

### 阶段 0：行为冻结（1-2 天）

**目标**：建立重构前的行为基线，确保重构过程中不会无声改变游戏行为。

| 步骤 | 验收标准 |
|------|---------|
| 0.1 `go test ./...` 全部通过 | ✅ `go test ./...` exit 0 |
| 0.2 `go vet ./...` 全部通过 | ✅ `go vet ./...` exit 0 |
| 0.3 保存 API 路由清单 | ✅ `docs/backend-refactor/api-route-inventory.md` 已存在 |
| 0.4 保存资源基线 | ✅ `docs/backend-refactor/resource-baseline.md` 已存在 |
| 0.5 保存 go.mod 和 go.sum | ✅ `reference/dependency-baseline/` 已存在 |
| 0.6 保存前端 API 调用清单 | ❌ 需要补充（提取所有 .api.ts 中的 path） |
| 0.7 保存当前 formula 测试覆盖率 | ❌ 需要补充（`go test -coverprofile`） |

**关键规则**：
- 阶段 0 期间**不修改任何业务逻辑代码**
- 新增工具和脚本**不得影响编译产物**
- 基线文件只追加不修改

### 阶段 1：HTTP 层迁移（3-5 天）

**目标**：`net/http ServeMux` → `go-chi/chi`，保持所有旧路由可用。

| 步骤 | 文件 | 改动 |
|------|------|------|
| 1.1 引入 chi | `go.mod` | `go get github.com/go-chi/chi/v5` |
| 1.2 修改 main.go | `cmd/simapi/main.go` | `mux := chi.NewRouter()` 替代 `http.ServeMux{}` |
| 1.3 移植中间件 | `handler/handler.go` | 将 `middleware.RequestID(Logger(...))` 改为 `mux.Use(middleware.RequestID)` 链 |
| 1.4 修改 Register 签名 | `handler/handler.go` | `func (h *Handler) Register(r chi.Router)` 替代 `*http.ServeMux` |
| 1.5 逐个移植 handler 注册 | 所有 `RegisterXxx` 方法 | `r.Get("/api/v2/companies/me/buildings/", h.withAuth(h.handle...))` 替代 `HandleFunc` |
| 1.6 引入 URL 参数 | path parsing 处 | `r.Get("/api/v3/companies/{id}/", h.withAuth(h.handleV3Company))` 替代 `strings.TrimPrefix` |
| 1.7 添加统一 error response | `handler/helpers.go` | 确保 `writeErr` 格式统一，所有错误包含 `error` 字段 |

**验收标准**：
- `go test ./...` 全部通过
- 启动 server 后所有旧 API 路由仍然可用
- 路由不再使用 `strings.TrimPrefix` 解析路径参数
- `go mod tidy` 后无多余依赖

**不需改动的**：
- handler 函数体（`handleXxx` 内部逻辑 ⚠️ 一个字符不改）
- service 层
- formula 层
- model 层

### 阶段 2：OpenAPI 契约（3-5 天）

**目标**：前后端共享 API 契约，不再互猜字段。

| 步骤 | 说明 |
|------|------|
| 2.1 创建 `docs/openapi.yaml` | 从当前 api-route-inventory.md 提取所有路由，已存在的稳定路由优先（market, company, production 读端点） |
| 2.2 安装 oapi-codegen | `go get github.com/oapi-codegen/oapi-codegen/v2` |
| 2.3 生成 Go 服务端类型 | `oapi-codegen -generate types,chi-server docs/openapi.yaml` |
| 2.4 创建 DTO 映射层 | `internal/httpapi/dto/` - 从 domain model 到 API response 的转换函数 |
| 2.5 安装 openapi-typescript | `npm install openapi-typescript` |
| 2.6 生成前端类型 | `npx openapi-typescript docs/openapi.yaml -o src/api/generated/types.ts` |
| 2.7 前端 api hooks 改用生成类型 | 逐个修改 `.api.ts` 的泛型类型从 `any` 改为生成的类型 |

**验收标准**：
- 后端编译通过，生成的 server type 可用
- 前端有 `.api.ts` 不再返回 `any`
- OpenAPI spec 与实际行为一致（通过对比测试验证）

**不需改动的**：
- 业务逻辑
- 数据库结构
- 服务发现

### 阶段 3：Service 切分（5-10 天）

**目标**：将 41 文件的巨型 service 按领域切分。

**方法**：不移动文件（短期），新建 app/ 目录的 use case，旧 Service 做 facade。

| 步骤 | 说明 |
|------|------|
| 3.1 定义每个 app service 的 interface | `app/production/service.go` 定义 `ProductionService` interface |
| 3.2 实现 ProductionService | 从 `service/production.go` 复制相关方法，分离时间依赖、ID 生成、持久化 |
| 3.3 旧 Service 保留 facade | `service/production.go` 改为调用 `app/production.Service` |
| 3.4 handler 逐渐切换 | handler 可以注入 app service 直接调用，绕过旧 Service |
| 3.5 重复 3.1-3.4 逐域 | market → finance → research → company → social |

**切分顺序建议**：
1. **Production** — 相对独立，依赖少
2. **Market** — 最复杂，但边界清晰（订单簿撮合）
3. **Finance** — Bond + 财务报表
4. **Research** — 研发系统
5. **Company** — 公司 + 等级
6. **Social** — 消息 + 聊天

**验收标准**：
- 每个 app service 有自己的 interface 和 tests
- 旧 facade 方法签名不变
- `go test ./...` 仍然通过

### 阶段 4：存储层重构（3-5 天）

**目标**：pgx 保留，引入 sqlc + migration 工具。

| 步骤 | 说明 |
|------|------|
| 4.1 引入 golang-migrate | 创建迁移目录 `storage/migration/` |
| 4.2 创建初始 migration | 当前 Postgres 建表 SQL 转为迁移文件 |
| 4.3 引入 sqlc 定义查询 | 对常用查询使用 sqlc 生成 |
| 4.4 Storage interface 拆分 | 按领域拆分：`CompanyStorage`、`MarketStorage`、`FinanceStorage` |
| 4.5 当前 postgres.go 重构 | 拆分为多个文件，每个 domain 一个 |
| 4.6 内存存储升级 | 当前 `NoopStorage` 改为真正的 in-memory 实现（可用于测试） |

**注意**：存储层重构不改变业务逻辑，只改数据访问路径。当前项目默认使用内存模式，所以这个阶段优先级可以降低到 P1。

### 阶段 5：前端 API 重构（3-5 天）

**目标**：TanStack Query 管 server-state，Zustand 管 UI/local state，WebSocket 接真实数据。

| 步骤 | 说明 |
|------|------|
| 5.1 清理所有 `map[string]any` 返回 | 用生成类型替代 |
| 5.2 ChatPanel 接真实 API | 取代 mock 数据 |
| 5.3 Research/Financial/Executives 确认全接 | 代码刚写完，需端到端验证 |
| 5.4 WebSocket 连接管理 | 实现 `useMarketWebSocket`、`useProductionWebSocket` |
| 5.5 Zustand 清理 | 确保 UI-only 状态在 ui.store，server data 不在 zustand 重复缓存 |

### 阶段 6：实时系统（5-7 天）

**目标**：WebSocket 推送必要事件。

| 步骤 | 说明 |
|------|------|
| 6.1 后端引入 coder/websocket | 添加 `/ws` endpoint |
| 6.2 连接管理 | 连接池 + company ID 绑定 + 心跳 |
| 6.3 事件广播 | market executeMatch 后推送 ticker 更新；production claim 推送通知 |
| 6.4 前端 WS hook | 重写 `websocket.ts`，实现自动重连 |
| 6.5 Chat 实时化 | 聊天通过 WS，不再轮询 |

### 阶段 7：经济系统保护（1-2 天）

**目标**：确保重构不改变经济数值。

| 步骤 | 说明 |
|------|------|
| 7.1 所有公式移入 formula 包 | 从 service 中提取剩余 16 个公式 |
| 7.2 formula 包加 regression test | 每个公式 + 边界 + 经济 baseline 对照 |
| 7.3 resource baseline 用于测试 | 测试读取 reference 目录的基线数据做比较 |
| 7.4 建筑/资源 JSON loader 加 validation | 验证必须字段存在，类型正确 |

---

## 10. 高风险区域

以下区域在重构过程中需要特别谨慎：

### 10.1 Market Order Matching (`service/market_match.go`, `service/market_trade.go`)

**风险**：撮合引擎是游戏经济的核心。执行顺序、填充逻辑、费用计算、状态更新都必须准确。

**保护措施**：
- 拆分匹配逻辑到独立的 `domain/market/match.go` 之前，先加完整的撮合测试（`market_test.go` 已有 6.8KB 测试 ✅）
- 先在 shadow 模式下运行新版匹配，锁定响应
- `executeMatch` (market_trade.go:149-212) 的价格/数量更新和资金转移必须原子

### 10.2 Production Queue (`service/production.go`)

**风险**：生产时间的计算、刷新逻辑、状态机（running→ready→claimed）涉及多个协程和计时。

**保护措施**：
- `refreshProductionJobs` (production.go:291-332) 的 time-based 刷新逻辑要加时钟抽象
- `calcProductionDuration` 使用 `formula.ProductionDurationSeconds` 确保一致性

### 10.3 Inventory Mutation (`service/service.go:477-505`)

**风险**：`inventoryGet`、`inventoryAdd`、`inventorySub` 是底层方法，在 market、production、claim、deliver 中都被调用。库存计算错误直接导致经济漏洞。

**保护措施**：
- 提取为 `inventory` 独立 struct，加单元测试
- 所有库存变更通过 ledger 记录审计轨迹（已部分实现）

### 10.4 Finance Ledger (`model/types.go:42-50`, `service/bond.go`)

**风险**：账本是可审计的财务记录，不能有舍入错误或丢失。

**保护措施**：
- 财务计算的舍入方向必须统一（`math.Floor` vs `math.Ceil`）
- `addLedger` 调用必须覆盖所有资金变更路径

### 10.5 Bonds (`service/bond.go`)

**风险**：债券涉及发行、购买、利息结算、违约处理，跨公司状态。

**保护措施**：
- `SettleAllBonds` 加 regression test
- `bond_test.go` 已有 7KB ✅，不要删减

### 10.6 Government Contracts (`service/government.go`)

**风险**：投标、授标、交付、履约、违约，有 bid 退款和保证金。

**保护措施**：
- `government_test.go` 已有 4.1KB ✅
- 授标算法（最低价中标？）要文档化

### 10.7 Scheduler (`scheduler/scheduler.go`)

**风险**：每 60s 的 Tick 驱动了债券结算、bot 市场、合同授标、订单清理、生产刷新。重构时不能丢失这些定时任务。

**保护措施**：
- Scheduler 的 `GameService` interface 拆分为多个小 interface
- 每个 tick action 加 error handling 和 metric

### 10.8 Bot Market Maker (`service/market_competition.go`)

**风险**：机器人流动性是市场不崩盘的关键。`replaceBotOrders`、`CheckMarketLock`、`RunBotMarketCycle`、`deployNationalTeam` 这些相互作用复杂。

**保护措施**：
- 加 bot 行为的经济影响测试
- 提取参数到 game.json（已部分做到）
- 不要把 bot 逻辑和玩家逻辑混在一起

### 10.9 Formula (`formula/`)

**当前状态**：✅ 质量最好。纯函数、1967 行测试、25+ 测试函数、文档齐全。

**保护措施**：
- 不要往 formula 里加任何 IO 或状态
- 新增公式必须加测试
- 公式变更必须更新 `game-formulas-v2.md`

### 10.10 Static Game Data Loader (`data/loader.go`)

**风险**：当前用 `map[string]any` 加载 JSON，消费方散落在 10+ 个文件中做类型断言。改 JSON 结构无声断裂。

**保护措施**：
- 生成 typed struct（可用 `encoding/json` 的 typed unmarshal 逐步替换）
- 加 schema validation
- 加加载时字段存在性检查

---

## 11. 不应该现在做的事

### 11.1 不要微服务化
当前项目规模完全不需要微服务。单体足够 -> 模块化 -> 只在必要时才考虑拆服务。

### 11.2 不要引入 Kafka / NATS / RabbitMQ
事件总线对当前规模过度设计。一个自写的 Go channel + interface 的轻量 EventBus 足够应付到 K 线功能上线。

### 11.3 不要全量 WebSocket 同步
**不要**一开始就通过 WS 推送整个 GameState。应该只推送增量事件（ticker update、production complete、notification）。全量同步仍然走 REST。

### 11.4 不要直接删除旧 API
在确认前端完全迁移到新版本前，旧 API 必须保留。用 chi 的子路由让新旧共存。

### 11.5 不要让 runtime 读取 docs/reference
resource-baseline.md 已明确声明：`reference/` 只作为测试对照快照，不要在生产代码中读取。

### 11.6 不要把公式散落进 handler
见到有人在 handler 里手算 tick 或费用，立即拦截。

### 11.7 不要把 domain model 直接暴露给前端
OpenAPI DTO 层是必须的。domain model 的字段名和结构是面向后端的，前端需要稳定契约。

### 11.8 不要一上来就 `backend-next`
`backend-next` 作为规划概念很好，但不要现在就创建 `backend-next/` 目录并行开发。当前项目的瓶颈是**工程效率**，不是**代码不可救药**。chi + OpenAPI + service 切分可以在原地完成。

### 11.9 不要替换 TanStack Query
前端已经有 TanStack Query v5，这是正确的选择。不要换成 SWR 或其他。

### 11.10 不要替换 PixiJS
PixiJS v8 已经是最佳选择。不要考虑 Three.js（太重）或 Canvas API（太低层）。

---

## 12. 最终建议

### 最短路线（Minimum Viable Migration Path）

**第一周应该做什么**：
1. 引入 chi，替换路由注册（不改 handler 函数体）
2. 引入 `coder/websocket`，实现基本 WS 端点（仅推送 ticker）
3. 引入 `log/slog` 替代 `log.Printf`
4. 创建 `docs/openapi.yaml` 第一版（只包含已稳定的读端点）

**第二周应该做什么**：
1. 引入 `oapi-codegen` + `openapi-typescript`，前后端类型共享
2. 创建 `app/production/` use case，将 production 从旧 service 拆分
3. 创建 `platform/clock.go` 抽象 time.Now

**第一个 PR 应该改什么**：
```
PR #1: net/http → go-chi/chi 路由迁移
- go.mod: +github.com/go-chi/chi/v5
- cmd/simapi/main.go: chi.NewRouter() + middleware chain
- handler/handler.go: Register 签名改为 chi.Router
- 逐个 handler 的 RegisterXxx: HandleFunc → r.Get/Post/...
- handler 内路径参数: strings.TrimPrefix → chi.URLParam(r, "id")
- 删除所有手工 path parsing
- 不修改任何业务逻辑
```

验收标准：
- `go test ./...` 全部通过
- `go vet ./...` 全部通过
- 启动后所有 API 可调用
- 前端不需要任何改动

**哪些库应该立即引入**：
| 库 | 引入时间 | 原因 |
|---|---------|------|
| go-chi/chi/v5 | **立即（第 1 天）** | 替换手工路由 |
| coder/websocket | **立即（第 1 天）** | 替换 WS 空壳 |
| go-playground/validator | **第 1 周** | 请求参数校验 |
| golang-migrate/migrate | **第 2 周** | 管理 DB schema |

**哪些库暂时只写 interface，不真正引入**：
| interface | 计划引入 | 说明 |
|-----------|---------|------|
| `platform/Clock` | 第 2 周 | 先抽象 `now()`，后面测试需要 |
| `eventbus/Bus` | 第 3 周（scheduler 解耦时） | 轻量 channel-based 实现即可 |
| `prometheus.Metrics` | 第 3 周 | 不影响功能，运筹需要 |
| `sqlc` | 第 4 周（存储层重构时） | 不影响功能，只改善 SQL 流程 |

### 一句话总结

> **不要重写，要重组织。** 项目最大的问题是工程效率（手工路由、无契约、胖 service），不是代码质量。chi + OpenAPI + service 切分可以原地解决 80% 的问题，不需要 `backend-next` 的 parallel universe。

---

报告结束。

*生成时间：2026-06-06*
*工具：手动审计 + 代码基线对比*
