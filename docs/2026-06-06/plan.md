对，现在该进入**标准重建流程**了。不是“AI 上来就改代码”，而是先做：

```txt
明确分工 → 定时间表 → 冻结现状 → 数据库设计 → API 契约 → 迁移方案 → 才开始写代码
```

你现在第一阶段不要叫“重构阶段”，应该叫：

# 第一阶段：重建准备与施工图阶段

目标不是改代码，而是把以后怎么改定死。因为你现在最大的问题不是某一行代码烂，而是**没有施工标准**：API 版本乱、service 太大、DTO 不明确、数据库边界不清、前端后端互猜字段。

技术路线我建议第一阶段直接按这个定：

```txt
Router: chi
API Contract: OpenAPI
Go codegen: oapi-codegen
DB: PostgreSQL + pgx
SQL layer: sqlc
Migration: goose
Frontend API: openapi-typescript + TanStack Query
Frontend local state: Zustand
Realtime later: coder/websocket
```

理由也很简单：`chi` 是轻量、可组合、适合大型 REST API 的 Go router，并且贴近 `net/http`，适合你从当前标准库路由迁移。([GitHub][1]) `oapi-codegen` 可以从 OpenAPI 生成 Go server、client、HTTP models，能减少手写 DTO 和样板代码。([GitHub][2]) `sqlc` 是写 SQL 再生成类型安全 Go 接口，适合你的市场订单、库存、金融流水这种不能靠 ORM 瞎猜的系统。([GitHub][3]) `goose` 用来管理数据库 schema migration，支持 SQL migration 和 Go migration。([GitHub][4])

---

# 第一阶段总目标

第一阶段只做 5 件事：

```txt
1. 确认重建边界：哪些保留，哪些重写，哪些迁移。
2. 明确分工：AI / 人 / 后端 / 前端 / 数据库 / 测试各做什么。
3. 冻结现状：API、公式、资源、数据库、前端调用全部记录下来。
4. 设计目标：数据库 schema 草案 + OpenAPI 草案 + 目录结构草案。
5. 生成第一批可执行任务：PR1、PR2、PR3 怎么切。
```

这阶段**不要大规模改代码**。最多允许加文档、测试、清单、脚本。否则 AI 很容易越改越乱。

---

# 第一阶段时间表

按 7 天算，比较稳。

| 天数    | 主题        | 产物                                                     | 是否改代码 |
| ----- | --------- | ------------------------------------------------------ | ----- |
| Day 1 | 项目冻结      | API 路由清单、前端调用清单、go test 结果、go vet 结果                   | 基本不改  |
| Day 2 | 分工和规则     | `rebuild-constitution.md`、AI 工作规则、禁止事项                 | 不改    |
| Day 3 | 数据库设计     | `database-design-v1.md`、ERD、表清单、字段草案                   | 不建表   |
| Day 4 | API 契约设计  | `openapi-draft.yaml`、DTO 命名规范、版本策略                     | 不生成代码 |
| Day 5 | 后端目标架构    | `backend-target-architecture.md`、service 切分图、chi 路由迁移图 | 不改业务  |
| Day 6 | 前端目标架构    | `frontend-api-plan.md`、TanStack Query hook 分组、页面补齐计划   | 不改页面  |
| Day 7 | 第一批 PR 计划 | PR1/PR2/PR3 任务单、验收标准、风险清单                              | 准备开工  |

如果你想更快，可以压成 3 天，但我不建议。因为你这项目有经济系统、市场、订单、债券、生产、库存，乱动会出隐藏 bug。

---

# 分工设计

你现在就算是一个人开发，也要模拟团队分工。否则 AI 会乱串职责。

## 角色 1：项目总架构 / Architect

负责：

```txt
- 定目标架构
- 定目录结构
- 定哪些旧 API 保留
- 定哪些模块先迁移
- 拒绝 AI 过度重写
```

第一阶段产物：

```txt
docs/rebuild/00-constitution.md
docs/rebuild/01-target-architecture.md
docs/rebuild/02-migration-map.md
```

## 角色 2：后端工程 / Backend

负责：

```txt
- chi 路由迁移方案
- handler 分组
- service 切分
- DTO 分离
- middleware 标准化
```

第一阶段产物：

```txt
docs/rebuild/backend-route-plan.md
docs/rebuild/backend-service-split-plan.md
```

## 角色 3：数据库工程 / Database

负责：

```txt
- PostgreSQL schema 草案
- 表关系设计
- 索引设计
- transaction 边界
- ledger / market / inventory 的一致性规则
```

第一阶段产物：

```txt
docs/rebuild/database-design-v1.md
docs/rebuild/database-transaction-rules.md
```

## 角色 4：API 契约 / API Contract

负责：

```txt
- OpenAPI 草案
- DTO 命名规范
- request/response 标准格式
- error response 标准格式
- API version 策略
```

第一阶段产物：

```txt
docs/rebuild/openapi-plan.md
docs/rebuild/api-version-policy.md
openapi/openapi-draft.yaml
```

## 角色 5：前端工程 / Frontend

负责：

```txt
- 前端 API 调用清单
- TanStack Query hook 分组
- Zustand 只管 UI 状态
- 哪些 mock 要删除
- Research / Financial / Executive / Chat 页面接入顺序
```

第一阶段产物：

```txt
docs/rebuild/frontend-api-plan.md
docs/rebuild/frontend-state-plan.md
```

## 角色 6：测试 / QA

负责：

```txt
- go test baseline
- go vet baseline
- API contract test 计划
- formula golden test
- market / production / inventory / ledger 高风险测试计划
```

第一阶段产物：

```txt
docs/rebuild/test-baseline.md
docs/rebuild/risk-test-plan.md
```

---

# 第一阶段数据库设计范围

注意：第一阶段**只设计，不建库**。

你要先把数据库分成这些核心域：

```txt
Auth / Player
Company / Profile
Buildings / Map
Resources / Inventory
Production / Queue
Market / Orders / Trades
Finance / Ledger
Bonds
Research
Executives
Government Orders
Daily Orders
Chat / Messages
Notifications
Scheduler / Events
Audit / Anti-cheat
```

## 第一版表清单

建议第一阶段先列这些表：

```txt
players
auth_sessions
companies
company_members

resource_catalog
building_catalog
recipe_catalog

company_buildings
placed_buildings
warehouse_items

production_jobs
production_queue_slots

market_orders
market_trades
market_tickers

ledger_entries
cashflow_snapshots
financial_snapshots

bonds
bond_holdings
bond_interest_payments

research_nodes
company_research_progress

executive_catalog
company_executives
executive_training_logs

daily_orders
government_orders
government_order_bids

chat_messages
notifications

scheduler_events
audit_events
anti_cheat_flags
```

第一阶段每张表只需要设计：

```txt
- 表名
- 归属领域
- 主要字段
- 主键
- 外键
- 重要索引
- 是否需要 transaction
- 是否是静态表
- 是否需要历史记录
```

暂时不要写最终 SQL migration。

---

# 第一阶段 API 设计范围

API 第一阶段要做的是**契约草案**，不是全量实现。

先分这几个 OpenAPI tag：

```txt
Auth
Players
Companies
Buildings
Warehouse
Production
Market
Resources
Finance
Bonds
Research
Executives
Orders
Government
Chat
Leaderboard
Notifications
Dev
Health
```

## API 统一响应格式

建议定成：

```json
{
  "data": {},
  "error": null,
  "meta": {}
}
```

错误统一：

```json
{
  "data": null,
  "error": {
    "code": "INSUFFICIENT_FUNDS",
    "message": "Not enough cash.",
    "details": {}
  },
  "meta": {
    "request_id": "..."
  }
}
```

这样前端不会每个接口单独猜格式。

## API 版本策略

建议这样定：

```txt
/api/v2/* = 保留兼容，当前前端继续能跑
/api/v3/* = 保留已有市场/政府等接口
/api/v4/* = 暂停新增，不再扩大
/api/new/* = 不建议
/api/internal/* = 只给后台任务/开发工具
```

后续新接口优先进入：

```txt
/api/v3 或 /api/v2 明确归属
```

不要再随手开 `/api/v4`。

---

# 第一阶段文件结构

建议你让 AI 先创建这些文档：

```txt
docs/rebuild/
  00-constitution.md
  01-current-baseline.md
  02-team-roles.md
  03-database-design-v1.md
  04-api-contract-plan.md
  05-backend-target-architecture.md
  06-frontend-target-architecture.md
  07-test-baseline.md
  08-risk-register.md
  09-first-pr-plan.md

openapi/
  openapi-draft.yaml

scripts/
  audit-routes.sh
  audit-frontend-api.sh
```

第一阶段最多允许加这些，不要直接改 `service/market_trade.go` 这种高风险业务文件。

---

# 第一阶段验收标准

第一阶段完成的标准是：

```txt
1. 能说清楚项目现在有哪些 API。
2. 能说清楚前端用了哪些 API。
3. 能说清楚哪些 service 文件必须拆。
4. 能说清楚数据库第一版有哪些表。
5. 能说清楚 OpenAPI 第一版有哪些 tag 和 DTO。
6. 能说清楚第一批 PR 改什么，不改什么。
7. 所有高风险模块都有测试计划。
8. AI 不允许绕过文档直接改业务。
```

最重要的一句话：

> 第一阶段结束时，你应该拿到“施工图”，不是拿到一堆半成品代码。

---

# 给 AI 的第一阶段 Prompt

下面这个直接丢给 Codex / Claude Code / Cursor。它的任务是**只做第一阶段，不许改业务代码**。

你是一名资深 Go 后端架构师、React/PixiJS 前端架构师、PostgreSQL 数据库设计师、OpenAPI API 契约工程师和游戏服务器重构顾问。

项目名称：NewHaven
项目类型：多人网页经济模拟游戏
技术栈：Go 后端、React + PixiJS 前端、食品产业链、市场订单、生产队列、金融流水、债券、研究、高管、政府订单、聊天、排行榜。
当前目标：进入“第一阶段：重建准备与施工图阶段”。

非常重要：
本阶段禁止直接重构业务代码。
本阶段禁止大规模修改 handler/service/model/formula。
本阶段只允许读取、审计、整理、生成文档、生成草案、生成检查脚本。
如果必须修改代码，只能新增无侵入性脚本或文档，不能改变运行行为。

你的任务是完成第一阶段重建准备包。

请按以下步骤执行。

第一步：读取项目和文档

请优先阅读：

* README.md
* docs/
* docs/rebuild/ 如果存在
* backend/
* backend/internal/handler/
* backend/internal/service/
* backend/internal/formula/
* backend/internal/model/
* backend/internal/storage/
* backend/internal/scheduler/
* backend/configs/game.json
* decompiled/data/resources.json
* decompiled/data/buildings.json
* decompiled/data/economy_model.json
* decompiled/data/resource_lookups.json
* client/
* client/src/api/
* client/src/features/
* client/src/store/
* go.mod
* package.json

如果文件不存在，请在报告中说明，不要猜。

第二步：生成当前基线报告

请生成：

docs/rebuild/01-current-baseline.md

内容必须包括：

1. 后端目录结构
2. 前端目录结构
3. 当前 Go 依赖
4. 当前前端依赖
5. 所有 API 路由清单
6. 前端调用 API 清单
7. handler 文件职责表
8. service 文件职责表
9. formula 文件职责表
10. storage / scheduler / middleware 当前职责
11. 当前自造基础设施清单

自造基础设施包括：

* 手写 router
* 手写 path parsing
* 手写 API response
* 手写 error format
* 手写 auth
* 手写 storage abstraction
* 手写 scheduler
* 手写 frontend API client
* 手写 websocket stub
* 手写 DTO / map[string]any response

第三步：生成重建宪法

请生成：

docs/rebuild/00-constitution.md

内容必须包括：

1. 重建目标
2. 不允许做的事情
3. 允许做的事情
4. API 兼容原则
5. 公式保护原则
6. 资源 baseline 保护原则
7. 数据库 migration 原则
8. OpenAPI 契约原则
9. 前端状态管理原则
10. AI 修改代码规则
11. PR 拆分规则
12. 验收标准

必须写入这些硬规则：

* 不允许直接删除旧 API。
* 不允许让 runtime 读取 docs/reference。
* 不允许把公式写进 handler。
* 不允许把 domain model 直接暴露给前端新 API。
* 不允许一次性重写 market / production / finance。
* 不允许在没有测试 baseline 的情况下修改订单撮合、库存扣减、生产队列、债券、ledger。
* OpenAPI 是前后端契约来源。
* formula 层应该保持纯函数。
* 数据库 schema 必须通过 migration 管理。
* 每个 PR 必须能独立运行和回滚。

第四步：生成分工文档

请生成：

docs/rebuild/02-team-roles.md

即使开发者只有一个人，也要按角色分工：

* Architect
* Backend
* Database
* API Contract
* Frontend
* QA
* Game Economy
* DevOps

每个角色写清楚：

* 负责什么
* 不负责什么
* 第一阶段产物
* 后续阶段产物
* AI 可以怎么帮
* 人类必须确认什么

第五步：生成数据库设计草案

请生成：

docs/rebuild/03-database-design-v1.md

本阶段只设计，不写 migration。

请至少包含这些领域：

* Auth / Player
* Company
* Buildings / Map
* Resources / Inventory
* Production
* Market / Orders / Trades
* Finance / Ledger
* Bonds
* Research
* Executives
* Government Orders
* Daily Orders
* Chat / Messages
* Notifications
* Scheduler / Events
* Audit / Anti-cheat

请至少提出这些表：

* players
* auth_sessions
* companies
* company_members
* resource_catalog
* building_catalog
* recipe_catalog
* company_buildings
* placed_buildings
* warehouse_items
* production_jobs
* production_queue_slots
* market_orders
* market_trades
* market_tickers
* ledger_entries
* cashflow_snapshots
* financial_snapshots
* bonds
* bond_holdings
* bond_interest_payments
* research_nodes
* company_research_progress
* executive_catalog
* company_executives
* executive_training_logs
* daily_orders
* government_orders
* government_order_bids
* chat_messages
* notifications
* scheduler_events
* audit_events
* anti_cheat_flags

每张表请写：

* purpose
* owner domain
* primary key
* important fields
* foreign keys
* indexes
* transaction notes
* history/audit requirement
* migration risk

请特别分析这些高风险 transaction：

* 创建市场订单
* 撮合市场订单
* 取消市场订单
* 开始生产
* 领取生产
* 扣库存
* 加库存
* 写 ledger
* 发行债券
* 购买债券
* 结算债券利息
* 政府订单竞标
* 政府订单交付

第六步：生成 API 契约计划

请生成：

docs/rebuild/04-api-contract-plan.md

同时生成：

openapi/openapi-draft.yaml

openapi-draft.yaml 只需要覆盖第一批核心接口，不需要完整实现全部接口。

第一批 OpenAPI tag：

* Auth
* Companies
* Buildings
* Warehouse
* Production
* Market
* Resources
* Finance
* Research
* Executives
* Chat
* Health

请设计统一响应格式：

SuccessResponse:
{
data: ...
error: null
meta: ...
}

ErrorResponse:
{
data: null
error: {
code: string
message: string
details: object
}
meta: {
request_id: string
}
}

请设计 DTO 命名规范：

* xxxRequest
* xxxResponse
* xxxDTO
* xxxListResponse
* xxxErrorCode

请明确：

* 哪些旧 API 保留
* 哪些 API 标记 legacy
* 哪些 API 后续迁移到 OpenAPI
* 哪些 API 不再新增
* API version 策略

第七步：生成后端目标架构

请生成：

docs/rebuild/05-backend-target-architecture.md

目标结构参考：

backend/internal/
domain/
market/
production/
finance/
research/
building/
player/
formula/
app/
marketapp/
productionapp/
financeapp/
researchapp/
buildingapp/
adapter/
http/
middleware/
v1/
v2/
v3/
postgres/
memory/
realtime/
platform/
config/
clock/
idgen/
log/
scheduler/

请说明：

* 现有文件迁移建议
* 哪些文件先不动
* 哪些 service 应该先拆
* 哪些 handler 应该先迁移到 chi
* 哪些 DTO 应该先定义
* 哪些 domain model 不应该直接暴露
* 旧 service.Service 如何作为 facade 保留
* 第一批 chi route group 怎么设计
* middleware chain 怎么设计

第八步：生成前端目标架构

请生成：

docs/rebuild/06-frontend-target-architecture.md

请说明：

* openapi-typescript 生成目录
* api client 目录
* TanStack Query hook 分组
* Zustand 只保存哪些 UI/local state
* 哪些 mock 数据要移除
* Research / Financial / Executives / Chat 接入顺序
* PixiJS 地图层和 React UI 层边界
* WebSocket 后续接入位置

建议结构：

client/src/api/
generated/
client.ts
queryKeys.ts
hooks/

client/src/features/
market/
production/
research/
financial/
executives/
chat/
realtime/

第九步：生成测试基线计划

请生成：

docs/rebuild/07-test-baseline.md

请包含：

* go test ./... 当前结果
* go vet ./... 当前结果
* 前端 build/test 当前结果
* formula golden test 计划
* handler contract test 计划
* market order matching test 计划
* inventory mutation test 计划
* production queue test 计划
* finance ledger test 计划
* bond settlement test 计划
* government order test 计划
* scheduler tick test 计划

如果不能实际运行命令，请写明原因，并列出应运行命令。

第十步：生成风险登记表

请生成：

docs/rebuild/08-risk-register.md

必须覆盖：

* market order matching
* production queue
* inventory mutation
* finance ledger
* bonds
* government contracts
* scheduler
* bot market maker
* formula
* static game data loader
* API compatibility
* frontend API mismatch
* database migration
* WebSocket consistency
* anti-cheat / abuse

每个风险写：

* severity
* probability
* evidence
* impact
* mitigation
* test required
* owner

第十一步：生成第一批 PR 计划

请生成：

docs/rebuild/09-first-pr-plan.md

请规划至少 5 个 PR：

PR1：文档和基线
PR2：OpenAPI 草案和 DTO 规范
PR3：chi 路由桥接，不改业务
PR4：测试基线和 contract test
PR5：数据库 schema 草案和 migration 准备

每个 PR 写：

* 目标
* 改哪些文件
* 不改哪些文件
* 验收标准
* 回滚方式
* 风险
* 人类需要确认的问题

第十二步：最终输出

请输出一个总结，包含：

1. 第一阶段生成了哪些文件
2. 哪些地方你没有动
3. 哪些风险最高
4. 第一批真正应该开工的 PR 是什么
5. 是否建议现在直接开始 chi 迁移
6. 是否建议现在直接开始数据库 migration
7. 是否建议现在直接生成 OpenAPI 类型

最终原则：

* 不要空话。
* 不要直接改业务。
* 不要为了“看起来完成”而隐藏风险。
* 如果发现已有文档和代码冲突，以代码为准，但在报告中指出冲突。

---

# 第一阶段你自己要给 AI 的硬限制

你还应该在 prompt 前面加一句：

```txt
本次任务不是让你重构代码，而是让你生成第一阶段施工图。任何业务代码修改都算失败。
```

这句话很重要。AI 很容易看到“重构”就直接改代码。

---

# 第一批 PR 怎么切

第一阶段之后，真正动代码我建议这样切：

## PR1：Docs + Baseline

```txt
只加 docs/rebuild/*
只加 scripts/audit-*.sh
不改 backend/internal/service
不改 backend/internal/handler
不改 client/src/features
```

验收：

```txt
go test ./...
go vet ./...
前端 build 通过
API route 清单生成成功
```

## PR2：OpenAPI Draft

```txt
新增 openapi/openapi-draft.yaml
新增 DTO 命名规范
不生成代码
不替换现有 API
```

验收：

```txt
OpenAPI lint 通过
核心接口至少覆盖 Auth / Company / Production / Market
```

## PR3：chi Bridge Prototype

```txt
新增 adapter/http 或 handler/router_chi.go
旧路由继续存在
chi 只做桥接，不改业务
```

验收：

```txt
旧 API 全部还能访问
测试通过
前端不需要改
```

## PR4：Test Baseline

```txt
补 handler contract test
补 market / production 高风险测试
不改行为
```

验收：

```txt
测试能证明迁移前后的行为一致
```

## PR5：Database Schema Draft

```txt
新增 db/schema-draft.sql 或 docs 版 schema
新增 migration 计划
不真正切换 runtime storage
```

验收：

```txt
能解释每张表为什么存在
能解释交易/库存/ledger transaction 顺序
```

---

# 第一阶段最重要的产物

如果你嫌多，至少要让 AI 产出这 5 个：

```txt
1. docs/rebuild/00-constitution.md
2. docs/rebuild/03-database-design-v1.md
3. docs/rebuild/04-api-contract-plan.md
4. docs/rebuild/05-backend-target-architecture.md
5. docs/rebuild/09-first-pr-plan.md
```

有了这 5 个，后面 AI 才不会像无头苍蝇一样乱改。
你现在最该做的是：**让 AI 先当建筑师，不要先当工人。**

[1]: https://github.com/go-chi/chi?utm_source=chatgpt.com "go-chi/chi: lightweight, idiomatic and composable router ..."
[2]: https://github.com/oapi-codegen/oapi-codegen?utm_source=chatgpt.com "oapi-codegen"
[3]: https://github.com/sqlc-dev/sqlc?utm_source=chatgpt.com "sqlc-dev/sqlc: Generate type-safe code from SQL"
[4]: https://github.com/pressly/goose?utm_source=chatgpt.com "pressly/goose: A database migration tool. Supports SQL ..."
