# Phase 1 后端目标架构

日期：2026-06-06  
状态：规划中  
基于：`docs/backend-refactor-constitution.md` v1

---

## 1. 目标目录树

```
backend/internal/
├── domain/                        # 纯 Go 域逻辑，无 IO 依赖
│   ├── player/                    # Player, Company 实体与规则
│   │   └── player.go
│   ├── building/                  # 建筑放置、升级、插槽约束
│   │   └── building.go
│   ├── production/                # ProductionJob, 槽位管理, 配方查找
│   │   └── production.go
│   ├── market/                    # MarketOrder, Trade, 撮合引擎
│   │   ├── order.go               # 订单类型、有效性
│   │   ├── matching.go            # 价格-时间优先撮合
│   │   └── market.go              # 市场状态查询
│   ├── finance/                   # LedgerEntry, Bond, 财务报表
│   │   ├── ledger.go
│   │   └── bond.go
│   ├── research/                  # ResearchProject, 研究树
│   │   └── research.go
│   └── social/                    # Message, Notification（轻量）
│       └── social.go
│
├── app/                           # 应用用例层：编排域逻辑 + 存储
│   ├── playerapp/
│   │   └── playerapp.go
│   ├── buildingapp/
│   │   └── buildingapp.go
│   ├── productionapp/
│   │   └── productionapp.go
│   ├── marketapp/
│   │   └── marketapp.go
│   ├── financeapp/
│   │   └── financeapp.go
│   ├── researchapp/
│   │   └── researchapp.go
│   └── socialapp/
│       └── socialapp.go
│
├── adapter/                       # IO 适配器：HTTP、数据库、实时
│   ├── http/
│   │   ├── middleware/            # 中间件链
│   │   │   ├── recovery.go
│   │   │   ├── requestid.go
│   │   │   ├── logger.go
│   │   │   ├── cors.go
│   │   │   ├── auth.go
│   │   │   └── ratelimit.go
│   │   ├── dto/                   # 请求/响应 DTO
│   │   │   ├── request.go
│   │   │   └── response.go
│   │   ├── v2/                    # 旧 API 版本适配器（兼容层）
│   │   │   └── v2.go
│   │   └── router.go              # chi 路由注册
│   ├── postgres/                  # PostgreSQL 仓储实现
│   │   ├── company_repo.go
│   │   ├── market_repo.go
│   │   └── migrations/            # SQL 迁移（后续使用 goose）
│   ├── memory/                    # 内存仓储实现（用于测试/dev）
│   │   ├── company_repo.go
│   │   └── market_repo.go
│   └── realtime/                  # WebSocket hub（预留）
│       └── hub.go
│
├── platform/                      # 运行时抽象层
│   ├── config/
│   │   ├── config.go              # 现有 config（保持不变）
│   │   └── config_test.go
│   ├── clock.go                   # time.Now 抽象
│   ├── idgen.go                   # ID 生成器接口 + 实现
│   └── logger.go                  # slog 包装器
│
├── formula/                       # 纯经济函数（**保持原位，不迁移**）
│   ├── admin.go
│   ├── bonds.go
│   ├── costs.go
│   ├── market.go
│   ├── production.go
│   ├── retail.go
│   └── saturation.go
│
├── model/                         # 领域类型（**保持原位**）
│   └── types.go
│
├── data/                          # 静态数据加载（**保持原位**）
│   └── ...
│
├── anticheat/                     # 反作弊检测（**保持原位，独立**）
│   └── ...
│
├── service/                       # **过渡期：旧 Service 逐渐瘦身**
│   ├── service.go                 # 核心 + 锁
│   ├── company.go
│   ├── production.go
│   ├── market.go
│   └── ...                        # 随着迁移逐步删除文件
│
├── handler/                       # **过渡期：旧 handler 逐渐退役**
│   ├── handler.go                 # 注册逻辑
│   ├── company.go
│   ├── production.go
│   └── ...                        # 随着迁移逐步删除文件
│
├── scheduler/                     # 后台循环（**保持原位**）
│   └── ...
│
└── storage/                       # **过渡期：旧 Storage 接口逐渐退役**
    ├── storage.go
    └── postgres.go
```

---

## 2. 各层职责与边界

### 2.1 `domain/` — 纯业务规则

**拥有：**

- 实体定义：`Player`、`Company`、`Building`、`ProductionJob`、`MarketOrder`、`Trade`、`LedgerEntry`、`Bond`、`ResearchProject`、`Message`、`Notification`
- 不变量的校验逻辑（例：建筑放置不能重叠、生产槽位数不能超出公司等级限制、订单不能自成交）
- 纯函数式业务规则（例：生产完成时间计算、订单有效状态判断、研究点数累计）
- 与 `model/types.go` 共享实体结构，或直接引用 `model` 中的类型（现阶段不动 `model`）

**不拥有：**

- IO：不调用 HTTP、数据库、文件系统、`time.Now`（通过 `platform/clock.go` 注入）
- 事务边界：不管理数据库事务或锁
- 编排逻辑：不调用存储库或其他域
- 副作用：不发送通知、不写日志、不修改全局状态

**迁移方式：**

从 `service/` 中提取纯业务逻辑，一个域一个域地进行。每个 `domain/` 包的初始内容是将 `service/` 中对应的**纯数据操作**和**规则判断**函数剥离出来，不涉及锁、不涉及 `GameState` 的直接遍历。

### 2.2 `app/` — 应用用例层

**拥有：**

- 用例编排：接收输入 → 加载聚合 → 调用域规则 → 持久化 → 返回结果
- 事务边界：每个公开方法为一个事务单元
- 跨域协调（例如：生产完成后更新库存并创建市场订单）
- 调用 `platform/clock.go` 获取当前时间
- 调用 `platform/idgen.go` 生成 ID

**不拥有：**

- HTTP 细节：不解析请求、不写响应、不处理 cookie/header
- 原始域逻辑：不包含 `if job.IsComplete()` 之外的业务规则（规则在 `domain/` 中）
- 锁：不持有 `sync.Mutex`（事务和并发由存储层或 app 层协调）

**App 层方法签名示例：**

```go
type ProductionApp struct {
    repo    production.Repository   // 接口，由 adapter/postgres 或 adapter/memory 实现
    clock   platform.Clock
    idGen   platform.IDGen
    domain  *domain.ProductionRules
}

func (a *ProductionApp) StartJob(ctx context.Context, companyID int, recipeID string, slot int) (*model.ProductionJob, error)
func (a *ProductionApp) ClaimJob(ctx context.Context, companyID int, jobID string) (*ClaimResult, error)
```

### 2.3 `adapter/http/` — HTTP 适配器

**拥有：**

- chi 路由器：负责路由注册、中间件链挂载、版本分组
- 中间件链：Recovery → RequestID → Logger → CORS → Auth → RateLimit
- DTO 转换：请求反序列化、响应序列化、错误格式统一
- 新版 handler（每个域一个文件）：接收 `*app.ProductionApp` 等，调用 app 层、返回 JSON
- `v2/` 兼容层：将旧 `handler.Handler` 适配为 chi `http.Handler`，挂载到旧路径

**不拥有：**

- 用例逻辑：不调用域规则或存储库
- 业务决策：不包含 `if order.Quantity > max` 之外的校验（校验在 domain 或 dto 层）

**中间件链（按顺序）：**

```
Recovery         →  panic 恢复，返回 500
RequestID        →  注入 X-Request-ID
Logger           →  结构化请求日志（method, path, status, duration）
CORS             →  跨域头
Auth             →  JWT 解析 → 注入 companyID/playerID 到 context
RateLimit        →  基于 IP 或公司 ID 的限流（预留）
```

**路由分组设计：**

```
/healthz, /readyz           # 无认证，存活/就绪探针
/api/v2/...                 # 旧 API 兼容路径（映射到旧 handler）
/api/v3/auth/login          # 新版登陆
/api/v3/company/profile     # 新版公司信息
/api/v3/production/jobs     # 新版生产
/api/v3/market/orders       # 新版市场
/api/v3/market/trades
/api/v3/finance/ledger      # 新版财务
/api/v3/finance/bonds
/api/v3/research/projects   # 新版研究
/api/v3/building/list       # 新版建筑
/api/v3/social/messages     # 新版社交
```

### 2.4 `adapter/postgres/` — PostgreSQL 仓储

**拥有：**

- 仓储接口的实现：`CompanyRepository`、`MarketRepository`、`ProductionRepository`、`FinanceRepository`、`ResearchRepository`
- SQL 查询和事务管理
- model ↔ 数据库行之间的映射

**不拥有：**

- 业务逻辑
- 连接池管理（由 `pgx` 池管理，仓储接收连接或池）

### 2.5 `adapter/memory/` — 内存仓储

**拥有：**

- 与 `adapter/postgres/` 相同接口的内存实现
- 用于测试和开发模式（替换 `NoopStorage`）

**不拥有：**

- 持久化保证（数据随进程消失）

### 2.6 `adapter/realtime/` — WebSocket Hub（预留）

**拥有：**

- WebSocket 连接管理（连接注册、心跳、断开清理）
- 消息广播（市场 tick、通知推送）

**不拥有：**

- 业务逻辑
- 持久化

### 2.7 `platform/` — 运行时抽象

**拥有：**

- `clock.go`：`type Clock interface { Now() time.Time }`，默认实现 `SystemClock`，测试用 `FrozenClock`
- `idgen.go`：`type IDGen interface { NewID() string }`，默认基于时间戳/随机
- `logger.go`：`slog.Logger` 的轻量包装，便于全局注入
- `config/`：现有配置加载器，保持不变

**注入方式：**

所有依赖通过构造函数注入，不使用全局变量：

```go
app := productionapp.New(
    repo,
    platform.SystemClock,
    platform.NewTimestampIDGen(),
    domain.NewProductionRules(data),
)
```

---

## 3. 迁移策略

### 3.1 `domain/` — 从 `service/` 逐域提取

**原则：**

- 不重写逻辑，只移动。提取后保留原 `service/` 中的调用点，逐步替换为新接口。
- 每个域的提取作为一个独立的 PR，不影响其他域。
- 提取前先为被移动的行为编写测试（`service/` 已有测试时，提取后补 `domain/` 测试）。

**提取顺序：**

| 顺序 | 域 | 原因 | 提取内容 |
|------|-----|------|---------|
| 1 | `building` | 最独立，无跨域依赖 | 放置规则、升级路径、槽位数限制 |
| 2 | `production` | 仅依赖配方数据 | 生产耗时计算、完成状态判断、槽位管理 |
| 3 | `market` | 最复杂，需要隔离测试 | 订单验证、撮合算法、深度计算 |
| 4 | `finance` | 依赖市场/公司 | 账本分录、债券到期计算、报表生成 |
| 5 | `research` | 依赖公司等级 | 研究点数计算、前置条件检查、解锁条件 |
| 6 | `player` | 基础实体 | 公司创建、等级升级、初始化 |
| 7 | `social` | 轻量 | 消息验证、通知生成 |

**具体例子：production 域的提取**

1. 创建 `backend/internal/domain/production/production.go`，从 `service/production.go` 复制纯逻辑函数（可能仅 `func CalculateCompletionTime(start time.Time, recipe *model.Recipe) time.Time` 级别）
2. `domain/production` 不得引用 `service.Service`、`storage.Storage`、`model.GameState`。
3. `service/production.go` 中的锁和方法头保持不变，内部调用改为委托 `domain/production`。
4. 提取完成后，`service/production.go` 行数减半，不再包含纯业务逻辑的副本。

### 3.2 `app/` — 新代码

**何时创建：**

- 当某个域完成 `domain/` 提取、且对应的 `adapter/memory/` 仓储可用时，创建对应的 `app/` 包。
- `app/` 使用 Go 接口来依赖仓储，不直接引用 `service.Service` 或 `storage.Storage`。

**与旧 Service 的共存期：**

```
                        ┌─────────────────────────────────────┐
  chi router            │  adapter/http/ (new handlers)        │
     │                  │  → 调用 app.ProductionApp            │
     ├──────────────────│  →  app 调用 domain + adapter/memory │
     │                  └─────────────────────────────────────┘
     │
     ├── 旧 handler (继续运行)
     │   → 调用旧 service.Service
     │   → service 内部逐渐委托给 app 和 domain
     │
     └── 写路径二选一（新 handler 优先）
```

过渡期内：

- 新 handler 调用 `app.ProductionApp`
- 旧 handler 继续调用 `service.Service`
- 数据源统一：都通过 `adapter/postgres/` 或 `adapter/memory/` 读写，不各自维护一份状态

### 3.3 `adapter/http/` — 新 chi 路由器

**引入 chi 的理由：**

- 当前 `http.ServeMux` + `strings.TrimPrefix` 模式的路径解析难以维护版本分组和中间件链。
- chi 是轻量（无框架化）、与 `net/http` 兼容、允许按组挂载中间件。
- chi 注册函数签名与 `http.Handler` 兼容，旧 handler 可以整体挂载到 chi 上。

**共存方案：**

```go
// adapter/http/router.go
func NewRouter(
    oldHandler *handler.Handler,    // 旧 handler，继续服务 /api/v2/ 路径
    productionApp *app.ProductionApp,
    marketApp *app.MarketApp,
    // ... 其他 app
) http.Handler {
    r := chi.NewRouter()

    // 全局中间件
    r.Use(middleware.Recovery)
    r.Use(middleware.RequestID)
    r.Use(middleware.Logger)
    r.Use(middleware.CORS)

    // 健康检查（无认证）
    r.Get("/healthz", healthHandler)
    r.Get("/readyz", readyHandler)

    // 旧 API 兼容：将整个旧 handler 挂载到 /api/v2/
    r.Group(func(r chi.Router) {
        r.Use(middleware.Auth)    // 旧 handler 的 auth 被中间件取代
        r.HandleFunc("/api/v2/*", oldHandler.ServeHTTP)   // 通配符转发
    })

    // 新 API v3
    r.Group(func(r chi.Router) {
        r.Use(middleware.Auth)
        r.Use(middleware.RateLimit)
        // v3 路由...
        r.Post("/api/v3/production/jobs", newProductionHandler.StartJob)
    })

    return r
}
```

**旧 handler 如何共存：**

- 旧 `handler.Handler` 的 `Register(*http.ServeMux)` 方法继续存在，但只注册在 chi 的 `/api/v2/` 子路径下。
- 旧 handler 内部路径提取方式（`strings.TrimPrefix(r.URL.Path, "/api/v2/...")`）需做兼容：当挂载在 `/api/v2/` 下时，路径不变；当旧 handler 单独运行时，路径仍为原样。
- 新 handler 一律注册在 `/api/v3/` 下，使用标准 `chi` 参数提取。

**旧 handler 的三种兼容模式：**

| 模式 | 说明 | 过渡阶段 |
|------|------|---------|
| `only` | 仅旧 handler 运行（当前状态） | Phase 1 开始 |
| `shadow` | 新旧并行，旧 handler 处理请求，新 handler 同步记录结果但不返回 | 各域 app 编写完成 |
| `next` | 新 handler 处理请求，旧 handler 关闭 | 各域迁移完成后 |

### 3.4 `adapter/postgres/` — 新仓储

**迁移路径：**

1. 定义 Go 接口（每个域一个仓储接口）：`CompanyRepository`、`ProductionRepository`、`MarketRepository`、`FinanceRepository`、`ResearchRepository`
2. 实现内存版本（`adapter/memory/company_repo.go`）
3. 实现 PostgreSQL 版本（`adapter/postgres/company_repo.go`）
4. 新 app 层只依赖接口，不依赖实现
5. 当前的 `storage.Storage` 接口作为一个粗粒度的过渡继续存在，直到所有域都迁移完毕

```go
// 过渡期：新旧接口共存
//
// storage.Storage（旧）— 粗粒度，保存整个 GameState
// adapter/postgres/CompanyRepository（新）— 细粒度，只操作 company 表
//
// 迁移完成前：旧 handler 通过 storage.Storage 读写
// 迁移完成后：新 app 通过 adapter/postgres/*Repository 读写
```

**旧 `storage.Storage` 的生命周期：**

- Phase 1 初始：`storage.Storage` 继续存在，旧 handler 和 service 依赖它
- 各域迁移时：新增的 `adapter/postgres/*Repository` 独立运行
- 迁移完成后：`storage.Storage` 接口废弃，代码删除

### 3.5 `platform/clock.go` — 注入旧代码

**策略：**

1. 先创建 `platform/clock.go`，同时不修改任何现有代码
2. 在 `service.Service` 中新增可选的 `Clock` 字段，默认为 `SystemClock`
3. 逐个域将 `s.now()` 替换为 `s.clock.Now()`
4. 测试中使用 `FrozenClock` 替代 `time.Now` 的 mock

```go
// platform/clock.go
type Clock interface {
    Now() time.Time
}

type systemClock struct{}
func (systemClock) Now() time.Time { return time.Now() }

var SystemClock Clock = systemClock{}

type FrozenClock struct {
    T time.Time
}
func (f FrozenClock) Now() time.Time { return f.T }
```

```go
// service/service.go — 过渡期修改
type Service struct {
    mu    sync.Mutex
    state model.GameState
    clock platform.Clock   // 新增字段
    // ...
}

func New(...) *Service {
    svc := &Service{
        clock: platform.SystemClock,
        // ...
    }
    return svc
}
```

---

## 4. 保持原位不动的文件

| 目录 | 原因 | 备注 |
|------|------|------|
| `formula/` (8 个文件) | 已经是纯函数，无 IO，无依赖 | 其余代码直接调用 `formula.XXX()` |
| `config/` (2 个文件) | 工作正常，设计合理 | 未来可能移到 `platform/config/`，但阶段 1 不动 |
| `data/` | 静态数据加载，与 gameplay 解耦 | 未来可能加入类型化查询，现阶段保持 |
| `anticheat/` | 独立子系统，无重构必要 | 保持原样 |
| `scheduler/` | 后台循环，逻辑固定 | 阶段 1 不修改；阶段 5 关注生命周期 |
| `model/types.go` | 所有域共享的类型定义 | 阶段 1 不拆分，未来可能按域拆分 |
| `cmd/simapi/main.go` | 主入口 | 仅需修改依赖注入部分（替换 `http.ServeMux` 为 chi） |
| `tests/` | 集成测试 | 追加新测试，不删除旧测试 |

---

## 5. Service 拆分顺序与计划

### 5.1 整体路线图

```
Phase 1a: 准备基础设施
  ├── 引入 chi 路由，挂载旧 handler 到 /api/v2/
  ├── 创建 adapter/http/middleware/（从 internal/middleware/ 迁移）
  ├── 创建 platform/clock.go + platform/idgen.go
  └── 定义各域仓储接口（repository interfaces）

Phase 1b: building 域（最独立）
  ├── domain/building/ — 放置规则、升级路径
  ├── adapter/postgres/building_repo.go
  ├── adapter/memory/building_repo.go
  ├── app/buildingapp/
  └── adapter/http/v3 路由

Phase 1c: production 域（次独立）
  ├── domain/production/ — 槽位、完成时间
  ├── app/productionapp/
  └── 旧 service/production.go 委托给 productionapp

Phase 1d: market 域（最复杂，需隔离测试）
  ├── domain/market/ — 订单、撮合、深度
  ├── app/marketapp/
  └── marketapp 接管订单创建/取消/撮合

Phase 1e: finance 域
  ├── domain/finance/ — 账本、债券
  ├── app/financeapp/
  └── 从 service 提取债券和账本逻辑

Phase 1f: research 域
  ├── domain/research/ — 研究项目、前置
  └── app/researchapp/

Phase 1g: company + player 域
  ├── domain/player/ — Company, Player
  └── app/playerapp/

Phase 1h: social 域（轻量）
  ├── domain/social/
  └── app/socialapp/

Phase 1i: 旧 service 瘦身收尾
  ├── service/ 中所有逻辑已委托到 domain/app
  ├── service.go 仅保留锁和状态快照
  └── 准备 Phase 2（时间/ID 可测化）
```

### 5.2 各域迁移详细范围

#### 5.2.1 building — 首批迁移

**现有文件：** `service/building.go` (1.3KB), `handler/building_shop.go` (3.4KB)

**迁移目标：**

- `domain/building/building.go`：建筑放置约束（地块是否可用、等级限制）、建筑类型定义
- `app/buildingapp/`：购买建筑、放置建筑、升级建筑
- `adapter/http/`：`POST /api/v3/building/purchase`, `POST /api/v3/building/place`

**依赖：** 无（仅 `model.Company` 和静态建筑数据）

#### 5.2.2 production — 第二批

**现有文件：** `service/production.go` (11.9KB), `service/production_claim.go` (2.7KB), `handler/production.go` (4.3KB), `handler/production_queue.go` (1.2KB)

**迁移目标：**

- `domain/production/production.go`：ProductionJob 槽位管理、配方查找、完成时间计算
- `app/productionapp/`：开始生产、领取产出、队列管理
- 旧 `service/production.go` 逐步委托给 `app/productionapp`

**依赖：** building（确定生产槽位）

#### 5.2.3 market — 第三批

**现有文件：** `service/market.go`, `service/market_match.go`, `service/market_depth.go`, `service/market_trade.go`, `service/market_info.go`, `service/market_competition.go`, `service/order.go`, `handler/market.go`, `handler/order.go`

**迁移目标：**

- `domain/market/order.go`：订单类型、有效性、过期
- `domain/market/matching.go`：价格-时间优先撮合（纯函数，可单元测试）
- `domain/market/market.go`：订单簿深度计算、价差计算
- `app/marketapp/`：下单、撤单、撮合（事务性）
- 旧 handler 的订单列表/市场概况仍可运行，但下单逻辑由 `marketapp` 接管

**注意：** 撮合引擎的并发和事务一致性是迁移重点。

#### 5.2.4 finance — 第四批

**现有文件：** `service/bond.go` (8.4KB), `handler/financial.go` (2.2KB), `handler/bond.go` (3.6KB)

**迁移目标：**

- `domain/finance/ledger.go`：账本条目结构、余额计算
- `domain/finance/bond.go`：债券发行、利息、到期
- `app/financeapp/`：发行债券、还款、生成财务报表

**依赖：** market（债券交易）、company（信用评级）

#### 5.2.5 research — 第五批

**现有文件：** `service/research.go` (6.5KB), `service/research_level.go` (2.0KB), `handler/executive.go` (2.8KB)

**迁移目标：**

- `domain/research/research.go`：研究项目前置条件、点数需求、解锁效果
- `app/researchapp/`：开始研究、加速完成、应用解锁

**依赖：** company（等级限制）、finance（研究经费）

#### 5.2.6 company & player — 第六批

**现有文件：** `service/company.go` (3.9KB), `service/player.go` (3.7KB), `handler/company.go` (8.9KB), `handler/player.go` (4.0KB)

**迁移目标：**

- `domain/player/player.go`：Player 实体、Company 创建规则
- `app/playerapp/`：注册、登录、公司信息

**依赖：** 所有其他域（公司聚合了建筑、生产、市场、财务信息）

---

## 6. 旧 `service.Service` 外观模式

**过渡期内，`service.Service` 的角色：**

```
┌──────────────────────────────────────────────────┐
│  service.Service （外观）                          │
│                                                   │
│  type Service struct {                            │
│      mu     sync.Mutex                            │
│      state  model.GameState                       │
│      clock  platform.Clock                        │
│      cfg    *config.Config                        │
│      data   *data.StaticData                      │
│      prodApp *app.ProductionApp    // 可选        │
│      marketApp *app.MarketApp       // 可选        │
│      // ...                                       │
│  }                                                │
│                                                   │
│  // 旧方法：对外接口不变                            │
│  func (s *Service) StartProduction(...)            │
│  func (s *Service) ClaimProduction(...)            │
│                                                   │
│  // 内部实现逐渐改为委托                            │
│  func (s *Service) StartProduction(...) {          │
│      s.mu.Lock()                                  │
│      defer s.mu.Unlock()                          │
│      // 通过 s.prodApp.StartJob(...) 实现          │
│  }                                                │
└──────────────────────────────────────────────────┘
```

**关键点：**

- 旧 handler 只调用 `service.Service` 的方法，签名不变
- `service.Service` 内部逐渐将实现委托给 `app.*App`，保持锁语义不变
- 新 handler 直接调用 `app.*App`，不经过 `service.Service`
- `service.Service` 在迁移过程中逐渐瘦身，最终可能只保留状态快照和 Save/Load

**何时可以删除旧的 Service 方法：**

当满足以下所有条件时：

1. 对应的新 handler 已上线并验证可用
2. 旧 handler 的所有调用方（前端的旧 API 请求）已被重定向到新 handler
3. 旧 handler 代码已移除或标记为已弃用
4. 运行一个月以上未出现兼容性问题

---

## 7. chi 路由注册 + 新旧 handler 共存（详细方案）

### 7.1 路由注册入口

```go
// adapter/http/router.go
package http

import (
    "net/http"
    "go-sim-api/internal/handler"       // 旧 handler
    "go-sim-api/internal/middleware"    // 旧中间件（过渡期引用）
    apphttp "go-sim-api/adapter/http"  // 新中间件 + handler
    "github.com/go-chi/chi/v5"
)

type RouterConfig struct {
    JWTSecret string

    // 旧依赖
    OldHandler *handler.Handler

    // 新 app 依赖
    ProductionApp *app.ProductionApp
    MarketApp     *app.MarketApp
    // ...
}

func NewRouter(cfg RouterConfig) http.Handler {
    r := chi.NewRouter()

    // === 全局中间件 ===
    r.Use(apphttp.middleware.Recovery)
    r.Use(apphttp.middleware.RequestID)
    r.Use(apphttp.middleware.Logger)
    r.Use(apphttp.middleware.CORS)

    // === 健康检查（无认证） ===
    r.Get("/healthz", apphttp.middleware.HealthHandler)
    r.Get("/readyz", apphttp.middleware.ReadyHandler)

    // === 旧 API v2（兼容层） ===
    r.Group(func(r chi.Router) {
        r.Use(apphttp.middleware.Auth(cfg.JWTSecret))
        // 旧 handler 的路径模式：/api/company/info → 变为 /api/v2/company/info
        r.HandleFunc("/api/v2/*", func(w http.ResponseWriter, r *http.Request) {
            // 路径兼容：将 /api/v2/xxx 转发为旧 handler 期望的 /api/xxx
            r.URL.Path = "/api" + strings.TrimPrefix(r.URL.Path, "/api/v2")
            cfg.OldHandler.ServeHTTP(w, r)
        })
    })

    // === 新 API v3（按域分组） ===
    r.Group(func(r chi.Router) {
        r.Use(apphttp.middleware.Auth(cfg.JWTSecret))
        r.Use(apphttp.middleware.RateLimit)

        // Auth（登录、注册在 v3 下不变，但 handler 用新的）
        r.Post("/api/v3/auth/login", v3handlers.Login(cfg.PlayerApp))
        r.Post("/api/v3/auth/register", v3handlers.Register(cfg.PlayerApp))

        // Production
        r.Get("/api/v3/production/jobs", v3handlers.ListJobs(cfg.ProductionApp))
        r.Post("/api/v3/production/jobs", v3handlers.StartJob(cfg.ProductionApp))
        r.Post("/api/v3/production/jobs/{jobID}/claim", v3handlers.ClaimJob(cfg.ProductionApp))

        // Market
        r.Get("/api/v3/market/orders", v3handlers.ListOrders(cfg.MarketApp))
        r.Post("/api/v3/market/orders", v3handlers.PlaceOrder(cfg.MarketApp))
        r.Delete("/api/v3/market/orders/{orderID}", v3handlers.CancelOrder(cfg.MarketApp))
        r.Get("/api/v3/market/depth", v3handlers.MarketDepth(cfg.MarketApp))

        // Finance
        r.Get("/api/v3/finance/ledger", v3handlers.Ledger(cfg.FinanceApp))
        r.Post("/api/v3/finance/bonds", v3handlers.IssueBond(cfg.FinanceApp))

        // Research
        r.Get("/api/v3/research/projects", v3handlers.ListResearch(cfg.ResearchApp))
        r.Post("/api/v3/research/projects/{projectID}/start", v3handlers.StartResearch(cfg.ResearchApp))

        // Building
        r.Get("/api/v3/building/shop", v3handlers.BuildingShop(cfg.BuildingApp))
        r.Post("/api/v3/building/purchase", v3handlers.PurchaseBuilding(cfg.BuildingApp))
        r.Post("/api/v3/building/place", v3handlers.PlaceBuilding(cfg.BuildingApp))

        // Company
        r.Get("/api/v3/company/profile", v3handlers.CompanyProfile(cfg.PlayerApp))
        r.Get("/api/v3/company/executives", v3handlers.Executives(cfg.PlayerApp))
    })

    return r
}
```

### 7.2 旧 handler 的内置路径前缀兼容

旧 handler `Register` 方法内部使用原始的路径模式（如 `/api/company/info`）。在 chi 挂载下，请求到达时 URL 已被 `r.URL.Path = "/api" + ...` 修改，所以旧 handler 内部无需改动。

当旧 handler 不再需要时（所有流量切换到 v3），删除 `v2/` 组和 `handler/` 目录。

### 7.3 旧 handler 驻留期间的路由覆盖检查

| 路径 | v2（旧 handler） | v3（新 handler） |
|------|------------------|-----------------|
| `/api/(v2/)company/info` | 由旧 handler 处理 | v3 下由新 handler 处理 |
| `/api/(v2/)production/list` | 由旧 handler 处理 | v3 下由新 handler 处理 |
| `/api/(v2/)market/buy` | 由旧 handler 处理 | v3 下由新 handler 处理 |
| ... | ... | ... |

新旧版本互不冲突：旧路径在 `v2` 下，新路径在 `v3` 下。

### 7.4 cmd/main.go 的修改

```go
func main() {
    // ... 现有初始化代码 ...

    // 旧依赖
    h := handler.New(svc, cfg.JWTSigningKey)

    // 新依赖（按域逐步添加）
    // var prodApp *app.ProductionApp    // 初始为 nil，逐步启用
    // var marketApp *app.MarketApp

    router := apphttp.NewRouter(apphttp.RouterConfig{
        JWTSecret: cfg.JWTSigningKey,
        OldHandler: h,
        // ProductionApp: prodApp,    // 逐步传入
        // MarketApp: marketApp,
    })

    srv := &http.Server{
        Addr:    cfg.Addr,
        Handler: router,    // 替换旧的 mux
        // ...
    }
    // ...
}
```

---

## 8. 关键接口定义（过渡期）

### 8.1 仓储接口（以 Production 为例）

```go
// adapter/postgres/repository.go（或者每个域一个文件）

type ProductionRepository interface {
    FindJob(ctx context.Context, companyID int, jobID string) (*model.ProductionJob, error)
    ListJobs(ctx context.Context, companyID int) ([]model.ProductionJob, error)
    SaveJob(ctx context.Context, job *model.ProductionJob) error
    DeleteJob(ctx context.Context, companyID int, jobID string) error
}

type MarketRepository interface {
    FindOrder(ctx context.Context, orderID string) (*model.MarketOrder, error)
    ListOrders(ctx context.Context, companyID int) ([]model.MarketOrder, error)
    SaveOrder(ctx context.Context, order *model.MarketOrder) error
    CancelOrder(ctx context.Context, orderID string) error
    ListActiveOrders(ctx context.Context) ([]model.MarketOrder, error)
    SaveTrade(ctx context.Context, trade *model.Trade) error
}

type CompanyRepository interface {
    FindByID(ctx context.Context, companyID int) (*model.Company, error)
    Save(ctx context.Context, company *model.Company) error
}

// ... 其他域类似
```

### 8.2 App 层接口（新旧切换的桥梁）

```go
// app/productionapp/productionapp.go
type ProductionApp struct {
    repo   ProductionRepository
    clock  platform.Clock
    idGen  platform.IDGen
    rules  *domain.ProductionRules
    data   *data.StaticData
}

func (a *ProductionApp) StartJob(ctx context.Context, companyID int, recipeID string, slot int) (*model.ProductionJob, error) {
    // 1. 检查公司是否有效（通过 repo 或 CompanyRepository）
    // 2. 检查配方是否存在（data）
    // 3. 检查槽位是否可用（domain rules）
    // 4. 创建 ProductionJob
    // 5. repo.SaveJob
    // 6. 返回结果
}

// old service/production.go 的委托版本
func (s *Service) StartProduction(companyID int, recipeID string, slot int) (*model.ProductionJob, error) {
    s.mu.Lock()
    defer s.mu.Unlock()

    if s.prodApp != nil {
        // 新路径：委托给 app，但 app 自己也加锁？需要协调
        // 过渡方案：service 持有锁，app 不持有
        return s.prodApp.StartJob(context.Background(), companyID, recipeID, slot)
    }

    // 旧路径：保留原有实现
    // ...
}
```

**锁协调策略：**

过渡期内，旧 handler 仍然通过 `service.Service` 进入，`service.Service` 持有 `sync.Mutex`。新 app 层设计为不自持锁（由调用方保证线程安全）。当 `service.Service` 委托给 app 时：

1. `service.Service` 加锁
2. 调用 `app.*` 方法（app 内部不加锁）
3. `service.Service` 解锁

当新 handler 直接调用 app 时：
1. 新 handler 不持有锁
2. `app.*` 方法通过仓储的事务隔离保证一致性
3. 数据库级别（PostgreSQL 行锁/SERIALIZABLE 隔离）或内存级（乐观锁）保证

---

## 9. 文件删除计划

| 阶段 | 删除内容 | 条件 |
|------|----------|------|
| Phase 1c 之后 | `service/building.go` | building 域迁移完成，旧 handler 不再调用 |
| Phase 1d 之后 | `service/production.go`, `service/production_claim.go` | production 域迁移完成 |
| Phase 1e 之后 | `service/market*.go`, `service/order.go` | market 域迁移完成 |
| Phase 1f 之后 | `service/bond.go` | finance 域迁移完成 |
| Phase 1g 之后 | `service/research*.go` | research 域迁移完成 |
| Phase 1h 之后 | `service/company.go`, `service/player.go`, `service/*` | 公司/玩家域迁移完成 |
| Phase 1i 收尾 | `handler/*` 中所有已替换的文件 | 所有 v3 新 handler 上线并验证后 |
| Phase 2 准备 | `storage/storage.go`（整体废弃） | 所有域仓储迁移完成，无代码引用该接口 |

---

## 10. 风险与缓解

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| 新旧 handler 数据不一致 | 玩家看到旧数据和新数据不同步 | v3 和 v2 共享同一套仓储和数据库，不存在独立副本 |
| `service.Service` 锁与新 app 的事务冲突 | 死锁或数据竞争 | app 层不自持锁；PostgreSQL 模式下使用行级锁，内存模式下由 service 的 Mutex 保护 |
| 迁移周期过长导致代码腐烂 | 新旧接口并存太久，维护成本高 | 每个域的迁移控制在 1-2 周内；同域内新旧代码不跨周并存 |
| chi 引入未预料的路由行为 | 旧 handler 路径解析异常 | 充分测试旧路径在 chi 下的行为，特别是 `/*` 通配符和路径前缀 Trim |
| 前端不兼容 v3 路径 | 前端用户无感知 | 前端继续使用 v2 路径，直到 v3 经过 QA 且前端明确适配 |

---

## 11. Phase 1 完成标志

- [ ] chi 路由运行，旧 handler 挂载在 `/api/v2/` 下正常响应
- [ ] `platform/clock.go` 创建并注入到 `service.Service`
- [ ] `platform/idgen.go` 创建并可供 app 层使用
- [ ] `adapter/http/middleware/` 目录创建，中间件从 `internal/middleware/` 迁移
- [ ] `adapter/memory/` 目录创建，提供至少两个域的内存仓储
- [ ] 至少两个域完成 `domain/` 提取 + `app/` 创建
- [ ] 新 v3 handler 可处理请求并返回正确结果
- [ ] 所有现有测试通过（`go test ./...`）
- [ ] 旧 `service/` 目录中至少有 2-3 个文件被清空或降为委托层
- [ ] 文档 `docs/api-contract.md` 同步更新
