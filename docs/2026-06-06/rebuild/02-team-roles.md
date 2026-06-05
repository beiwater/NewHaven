# 团队角色与分工说明

> Phase 1 — 重建准备与施工图阶段
>
> 即使只有一个人开发，也必须模拟团队分工，防止 AI 或单人思维在职责边界上越界。每个角色有明确的责任范围、交付物、以及必须由人确认的关键决策。

---

## 目录

1. [角色总览](#角色总览)
2. [架构师 / Architect](#1-架构师--architect)
3. [后端工程 / Backend](#2-后端工程--backend)
4. [数据库工程 / Database](#3-数据库工程--database)
5. [API 契约 / API Contract](#4-api-契约--api-contract)
6. [前端工程 / Frontend](#5-前端工程--frontend)
7. [测试与质量 / QA](#6-测试与质量--qa)
8. [游戏经济 / Game Economy](#7-游戏经济--game-economy)
9. [运维与基础设施 / DevOps](#8-运维与基础设施--devops)
10. [协作流程](#协作流程)

---

## 角色总览

| # | 角色 | Phase 1 核心产出 | 核心文件区域 |
|---|------|-------------------|-------------|
| 1 | 架构师 | 目标架构、迁移路线图、目录结构 | `docs/2026-06-06/rebuild/*` |
| 2 | 后端工程 | 路由迁移方案、Service 切分方案 | `backend/internal/handler/`, `service/` |
| 3 | 数据库工程 | Schema 草案、事务边界、索引策略 | `backend/internal/model/`, `storage/` |
| 4 | API 契约 | OpenAPI 草稿、DTO 规范、版本策略 | `docs/api-contract.md`, `openapi/` |
| 5 | 前端工程 | 前端 API 调用清单、Hook 分组、状态边界 | `client/.../src/api/`, `store/`, `features/` |
| 6 | QA | 测试基线、风险测试计划、回归保障 | `backend/internal/*/*_test.go` |
| 7 | 游戏经济 | 公式基线、经济模型校对、参数配置化 | `backend/internal/formula/`, `decompiled/data/` |
| 8 | DevOps | 开发环境标准化、CI 基础、配置分离 | `backend/go.mod`, `configs/`, `scripts/` |

---

## 1. 架构师 / Architect

### 职责

- 定义整体重建目标：哪些模块保留、哪些重写、哪些通过桥接迁移（参考 `backend-refactor-constitution.md` 的 migration-modes）
- 划定目录结构和包边界：domain 走向（auth → company → production → market → finance → research → social → system）
- 决定技术栈选型：chi router、oapi-codegen、sqlc、goose、pgx 的技术选型理由和集成方式
- 制定迁移策略：old / next / shadow 三种模式的定义和切换时机（见 `bridge-routing-plan.md`）
- 拒绝 AI 或开发者的过度重写冲动，保持"可玩"红线

### 非职责

- 不写具体 handler 或 service 代码
- 不做数据库字段级设计
- 不参与经济公式的数值调整

### Phase 1 交付物

- `docs/2026-06-06/rebuild/00-constitution.md` — 重建宪章、技术选型、阶段划分
- `docs/2026-06-06/rebuild/01-target-architecture.md` — 目标架构图、包依赖关系、目录树
- `docs/2026-06-06/rebuild/02-migration-map.md` — 模块迁移路线图，标注每个 domain 的当前状态和目标模式
- `docs/2026-06-06/rebuild/02-team-roles.md` （本文档）

### 后续阶段交付物

- PR 级架构审核：每个 PR 合入前的架构一致性检查
- ADR（架构决策记录）：在 `docs/2026-06-06/backend-refactor/adr/` 下记录关键决策（如"为什么选择 chi 而非 gin"）

### AI 如何辅助

- 对比当前 `backend/internal/handler/` 下所有 handler 的路由注册模式，自动生成 API 路由清单（已有 `api-route-inventory.md`）
- 分析 `service.Service` 的结构（`service.go:20-36`），建议 domain 拆分方案
- 从 handler 调用关系推断 service 依赖图

### 人类必须确认

- **模块迁移优先级**：哪个 domain 先拆、哪个 domain 保持不动，必须由人根据游戏实际运营需求决定
- **技术选型**：AI 推荐了 chi/oapi-codegen/sqlc/goose，但选型是否适配项目实际情况（Go 版本 1.25、已有 pgx 依赖），人需要确认兼容性，见 `backend/go.mod`
- **不可触碰的文件列表**：哪些 handler 或 API 因为前端正在用，必须保持兼容（如 `handler/dev.go` 中路由 `v4/` 和 `contracts-incoming` 看起来像产品端，不能随意改名）

---

## 2. 后端工程 / Backend

### 职责

- 实现 chi router 迁移方案，将当前 `http.ServeMux` 注册模式逐步改为 chi 分组路由
- 拆分 `service.Service` 巨型结构体（`backend/internal/service/service.go:20-36`），按 domain 切分为独立 service
- Handler 文件重组：当前 22 个 handler 文件（见 `handler/` 目录清单），需要按 API tag 分组、消除 `dev.go` 中混合的 v2/v3/v4 产品路由
- 统一错误响应格式：当前 handler 散落 `writeErr(w, 400, ...)` 调用，需要标准化为 JSON error envelope
- 标准化 middleware 链：认证、日志、恢复、CORS、请求 ID 的注册顺序和复用方式

### 非职责

- 不改动经济公式逻辑（`internal/formula/`）
- 不改动数据库 schema 或存储层接口（由 Database 角色定义）
- 不负责 API 契约定义（由 API Contract 角色输出 OpenAPI，后端按契约实现）

### Phase 1 交付物

- `docs/2026-06-06/rebuild/backend-route-plan.md` — chi 路由注册方案，含路由分组设计、middleware 顺序、旧路由兼容策略
- `docs/2026-06-06/rebuild/backend-service-split-plan.md` — Service 拆分方案，标注 `service/production.go`、`service/market_trade.go`、`service/bond.go`、`service/company.go` 等文件的目标 domain 归属
- `backend/internal/handler/` 下每个文件的职责标记（在文件头部加注释标注 domain 归属和 API version）

### 后续阶段交付物

- 按 PR 逐个 domain 迁移 handler 注册到 chi
- 按 PR 逐个拆分 service 文件到 domain package
- 为每个 domain 添加 `go doc` 兼容的 package 文档
- 删除已废弃的旧路由和 handler 方法

### AI 如何辅助

- 分析 `handler/production.go:9-16` 到 `handler/market.go:11-23` 的路由注册模式，自动生成 chi 分组代码
- 从 `service/service.go` 的公开方法签名出发，推理哪些方法属于同一个 domain，输出拆分建议
- 检测当前 handler 中内联定义的 request struct（如 `handler/market.go:53-60` 的 `CreateOrderRequest`），建议提取为共享 DTO

### 人类必须确认

- **API 兼容性**：在迁移路由时，旧的 `/api/v2/...` 路径不能 404，必须在 PR 合入前用 `go test ./...` 确认 handler_test 通过，见 `handler/handler_test.go`
- **Service 拆分粒度**：AI 可能过度拆分（每个方法一个文件），必须控制 domain 的边界合理性：比如 `market_trade.go` 和 `market_match.go` 是拆成两个 package 还是一个 package 的两个文件

---

## 3. 数据库工程 / Database

### 职责

- 设计 PostgreSQL schema 草案：表名、字段、主键、外键、索引、缺省值
- 定义 transaction 边界：哪些操作需要原子性（市场订单撮合、财务流水记账、债券利息结算）
- 制定一致性规则：ledger / market / inventory 三者的数据一致性保障策略
- 评估当前 memory-first 存储模式（`internal/storage/postgres.go`）与目标 pgx + sqlc 的差异

### 非职责

- 不创建实际数据库或运行 migration
- 不修改 `internal/storage/storage.go` 的接口定义（Phase 1 只设计，不实现）
- 不涉及经济公式的字段含义调整（属于 Game Economy 角色）

### Phase 1 交付物

- `docs/2026-06-06/rebuild/database-design-v1.md` — schema 草案，含所有表的结构描述（参考 `plan.md` 中列出的 40+ 张表）
- `docs/2026-06-06/rebuild/database-transaction-rules.md` — 事务边界文档，标注哪些操作需要 `BEGIN/COMMIT`、哪些可以最终一致性
- 每张表的字段来源标注：哪些来自 `model/types.go` 中的现有 struct（如 `MarketOrder`、`Trade`、`Company`、`Bond`、`LedgerEntry`），哪些是新加的

### 后续阶段交付物

- 正式的 SQL migration 文件（goose 格式），放置在 `backend/migrations/`
- sqlc 配置文件 `sqlc.yaml` 和生成的 Go 查询代码
- 存储层适配器：将当前 `storage.Storage` 接口的实现从 JSON dump 迁移到 pgx

### AI 如何辅助

- 从 `model/types.go` 的 struct 定义（如 `Company` 第 95-110 行、`MarketOrder` 第 71-82 行、`Bond` 第 9-21 行）反向生成建表 SQL 草稿
- 分析 `service/market_trade.go` 和 `service/market_match.go` 中的读写逻辑，标注需要事务保护的字段
- 对比当前 model 中 `map[string]any` 字段（如 `LedgerEntry.Meta` 第 49 行）和 JSON 字段，建议具体 schema

### 人类必须确认

- **表归属**：每个表属于哪个 domain，人必须核准。比如 `company_buildings` 归 Company 还是 Buildings 域
- **事务边界**：AI 可能建议过度事务化（每步都 BEGIN/COMMIT）或过松（撮合不事务），人需要根据 game 的实际一致性要求确认
- **软删除 vs 硬删除**：市场订单取消是软删除（status=cancelled）还是硬删除，决定了 schema 设计

---

## 4. API 契约 / API Contract

### 职责

- 编写 OpenAPI 3.0 草案（先 YAML，不生成代码）：定义所有端点的 request/response schema
- 制定 DTO 命名规范和服务分层规则：Handler DTO / Service DTO / Storage DTO 的转换责任
- 定义标准错误响应格式：error code、message、details、trace ID 的结构
- 设计 API 版本策略：当前混合了 `/api/`、`/api/v2/`、`/api/v3/`、`/api/v4/`，需要决定弃用路线

### 非职责

- 不生成 Go server/client 代码（Phase 1 后由 oapi-codegen 自动生成）
- 不改动 handler 实现（由 Backend 角色完成）
- 不定义前端 API 调用方式（由 Frontend 角色决定用哪个 endpoint）

### Phase 1 交付物

- `docs/2026-06-06/rebuild/openapi-plan.md` — OpenAPI 文档编写计划，含 tag 分组、DTO 命名约定、复用组件设计
- `docs/2026-06-06/rebuild/api-version-policy.md` — API 版本策略文档，规定每个版本的生命周期（如 v2 保持兼容到 2026Q3，v4 标记为 deprecated）
- `openapi/openapi-draft.yaml` — 初步 OpenAPI 定义，至少覆盖核心 domain（Auth、Player、Market、Production、Finance）

### 后续阶段交付物

- 完整的 OpenAPI 规范，覆盖所有 16+ domain
- oapi-codegen 生成的 Go server 接口和类型（作为 `backend-next` 的契约基础）
- openapi-typescript 生成的前端 TypeScript 类型
- API 契约测试套件：验证 handler 的 response 符合 OpenAPI schema

### AI 如何辅助

- 从当前 handler 的 `writeJSON(w, 200, ...)` 调用（如 `handler/player.go:29`、`handler/market.go:32`）反推 response schema
- 从 handler 中的内联 struct 定义（如 `handler/market.go:53-60` 的 `CreateOrderRequest`）提取为 OpenAPI components
- 对比 `docs/api-contract.md` 和实际 handler 实现的差异，标注偏离点

### 人类必须确认

- **DTO 边界**：是否真的需要三层 DTO（Handler/Service/Storage），还是 Handler 和 Service 可以共享。这决定了 oapi-codegen 生成的代码直接使用还是需要 adapter
- **版本弃用日程**：/api/v1/buildings/ 是否还有前端在用？必须确认后再决定是否标记 deprecated
- **错误码分类**：400（客户端错误） vs 500（服务端错误）的边界，以及业务错误（如余额不足）是否应该返回 200 + error code 还是 4xx

---

## 5. 前端工程 / Frontend

### 职责

- 建立前端 API 调用清单：当前 `client/atlas-foods-client/src/api/` 下 14 个 API 模块，一一映射到后端路由
- 设计 TanStack Query hook 分组方案：按 domain 划分 query key 结构、缓存策略、stale time
- 明确 Zustand 只管理 UI 状态（如 `game.store.ts` 中的 zoom/pan/selectedMapBuilding），业务真相全部来自后端
- 规划新页面的接入顺序：Research、Financial、Bonds、Executives、Auction 的优先级和依赖关系

### 非职责

- 不修改 PixiJS 地图渲染逻辑（`game/GameCanvas.tsx`）的业务行为
- 不新建业务 API endpoint（由 API Contract + Backend 定义）
- 不调整经济数值在前端的展示格式（由 Game Economy 决定）

### Phase 1 交付物

- `docs/2026-06-06/rebuild/frontend-api-plan.md` — 前端 API 调用清单，标注每个前端 API 调用对应的后端路由、当前状态（可用/缺失/需要 adapter）
- `docs/2026-06-06/rebuild/frontend-state-plan.md` — 前端状态边界文档，定义哪些数据走 TanStack Query（服务器状态）、哪些走 Zustand（UI 状态）、哪些走 React context（主题/临时）

### 后续阶段交付物

- TanStack Query hook 重构：当前 `src/api/` 下的原始 fetch 调用改为统一的 query hooks
- 缺失页面的补充：FinancialPage、ExecutivePage、BondMarketPage 等功能完整实现
- 删除过时的 mock 数据和硬编码 fallback

### AI 如何辅助

- 遍历 `src/api/*.api.ts` 中所有 fetch 调用，统计每个后端路由的使用频率和参数模式
- 从 `src/features/` 下各组件的 import 依赖，分析当前前端对后端的真实依赖关系
- 生成 TanStack Query hook 的样板代码（每个 domain 一个 hook 文件，如 `useMarket.ts`、`useProduction.ts`）

### 人类必须确认

- **API adapter 策略**：后端 API 返回的字段名（如 `resourceId` vs `resource_id`）是否需要前端 adapter 转换，还是统一 camelCase
- **页面优先级**：Research 页面和 Financial 页面哪个先做，取决于当前游戏运营需求。`docs/2026-06-02/dev-plan.md` 已有排序，人需要确认
- **TanStack Query 缓存策略**：每个资源的 stale time 和 cache time 不能一刀切：market ticker 需要短缓存（<30s），company info 可以长缓存（5min）

---

## 6. 测试与质量 / QA

### 职责

- 建立 Phase 1 的 go test 基线，确保 `go test ./...` 在当前主干上全部通过
- 识别高风险区域并制定专项测试计划：市场撮合（`service/market_match.go`）、生产加锁（`service/production.go`）、金融流水一致性（`service/bond.go`）
- 设计 API 契约测试方案：验证 handler response 与 OpenAPI schema 的一致性
- 维护回归测试保障：每次 PR 必须携带或更新测试

### 非职责

- 不编写性能测试或压力测试（Phase 1 不涉及）
- 不测试前端 UI 组件（属于后续阶段 E2E 测试）
- 不修复业务 bug（只报告和归档）

### Phase 1 交付物

- `docs/2026-06-06/rebuild/test-baseline.md` — 当前测试基线报告：每个 package 的测试覆盖率、通过率、已知失败测试清单
- `docs/2026-06-06/rebuild/risk-test-plan.md` — 高风险测试计划，重点覆盖 market / production / inventory / finance / bond 领域

### 后续阶段交付物

- API 契约测试套件：自动验证 handler response 符合 OpenAPI schema
- Golden test 文件：经济公式的输入输出对，防止公式重构导致数值偏移
- 集成测试环境：模拟完整玩家操作流程的端到端测试

### AI 如何辅助

- 运行 `go test ./... -v` 和 `go vet ./...`，收集失败测试和 vet warning，输出可执行的修复清单
- 分析 `service/*_test.go`（已有 12 个测试文件）的覆盖缺口，建议新增测试用例
- 对比 `service/production_test.go` 中的测试数据和实际 `service/production.go` 逻辑，检测测试是否覆盖了所有分支

### 人类必须确认

- **测试优先级**：AI 可能建议全覆盖，但 Phase 1 只有 7 天，人必须圈定核心 domain（Market、Production、Finance）优先保障
- **Golden test 数据**：经济公式的 golden 数据来源——从 `decompiled/data/economy_model.json` 提取还是从当前运行日志提取
- **已知失败测试的处理**：如果基线中已有失败测试，是否在 Phase 1 修复，还是标记为已知问题延迟处理

---

## 7. 游戏经济 / Game Economy

### 职责

- 冻结当前经济公式基线：`internal/formula/` 下所有公式的输入输出快照
- 校对反编译经济数据与当前运行公式的一致性：`decompiled/data/economy_model.json` vs `formula/*.go`
- 定义经济参数配置化方案：将硬编码常量（如手续费率 4%、基础产量、薪资系数）迁移到 `configs/game.json`
- 确保公式的纯函数和确定性：`formula/market.go`、`formula/production.go`、`formula/retail.go`、`formula/bonds.go` 不应依赖外部状态

### 非职责

- 不调整游戏数值平衡（属于 Game Design）
- 不修改 service 层的业务逻辑（只关注纯公式）
- 不参与数据库 schema 或 API 契约设计

### Phase 1 交付物

- `docs/2026-06-06/rebuild/formula-baseline.md` — 公式基线文档，记录每个公式函数的签名、输入范围、当前实现文件行号
- `docs/2026-06-06/rebuild/economy-config-plan.md` — 经济参数配置化方案，标注哪些常量应该从 `configs/game.json` 读取

### 后续阶段交付物

- 公式 golden test 文件：对 `formula.OutputPerHour()`、`formula.ProductionDurationSeconds()`、`formula.UnitsSoldPerHour()`、`formula.BondInterest()` 等关键函数的输入-输出快照
- 经济参数配置化重构：将 `formula/production.go:37-42` 中的硬编码常量抽取为 config 字段
- 经济状态影响模型：从 `economy_model.json` 的三种经济状态推导当前动态影响

### AI 如何辅助

- 对比 `decompiled/data/economy_model.json` 中的 `modeledProductionCostPerUnit` 和 `modeledUnitsSoldAnHour` 与当前 formula 实现是否一致
- 从 `formula/*.go` 中提取所有 magic number（如 `formula/market.go` 中的 `0.04` 手续费率、`formula/production.go` 中的 `500.0` 基础产量），生成参数清单
- 分析 `handler/dev.go:36-48` 中暴露的 formula 测试路由，验证前端展示的公式值与公式代码的对应关系

### 人类必须确认

- **公式正确性**：AI 无法验证经济模型的设计意图，必须由人（或原作者）确认公式逻辑是否正确。特别是 `formula/retail.go` 中的零售模型，反推得到的 `kle/J7r` 结构版本是否准确
- **参数配置化边界**：哪些参数属于游戏可调配置（放 `game.json`），哪些属于代码级常量（放 `formula/`），人需要根据运营频率决定
- **经济状态切换逻辑**：`economy_model.json` 定义了三种经济状态，但当前代码是否真正实现了状态切换？人需要确认当前是否只是静态使用状态 0

---

## 8. 运维与基础设施 / DevOps

### 职责

- 标准化开发环境：Go 版本 1.25、Node 版本、PostgreSQL 本地配置的统一约定
- 建立 CI 基础：至少保证 `go test ./...` 和 `go vet ./...` 作为合入门禁
- 配置分离：区分开发/测试/生产配置，当前 `backend/configs/game.json` 是单环境配置，需要拆分
- 建立文档生成和代码检查流水线：go fmt、go vet、staticcheck 等工具的集成

### 非职责

- 不涉及生产部署或容器化（Phase 1 不部署）
- 不搭建数据库集群或 CDN
- 不设计监控和告警系统

### Phase 1 交付物

- `docs/2026-06-06/rebuild/dev-environment-setup.md` — 开发环境标准化文档，含 Go 工具链版本、PostgreSQL 可选配置、前端构建步骤
- `docs/2026-06-06/rebuild/ci-baseline.md` — CI 基线配置说明，定义合入门禁规则（test pass、vet pass、format check）
- 根目录下的 `.github/workflows/ci.yml`（或等价 CI 配置）草稿

### 后续阶段交付物

- Docker Compose 开发环境（Go server + PostgreSQL + Redis 可选）
- Makefile 或 Taskfile：统一构建、测试、lint、migrate 命令
- Pre-commit hook 配置（go fmt、go vet、eslint 等）
- 生产环境部署剧本

### AI 如何辅助

- 从 `backend/go.mod` 分析当前 Go toolchain 和依赖版本，输出 `go vet` 已知问题
- 扫描 `backend/configs/game.json` 中与环境相关的配置（如 DB 连接串、端口），建议环境变量化
- 生成 CI 配置文件草稿（如 GitHub Actions 的 `ci.yml`）

### 人类必须确认

- **CI 门禁的严格程度**：是否要求 100% 测试通过才能合入，还是允许已知失败测试标记跳过
- **PostgreSQL 的必选/可选**：当前 `storage/postgres.go` 是可选持久化（`backend/internal/storage/storage.go` 中 `Storage` 接口是 interface），Phase 1 是否要求数据库必须运行
- **工具链版本**：Go 1.25 是当前版本，但 CI 镜像是否支持？Node 版本 v22+ 能否兼容当前项目

---

## 协作流程

### Phase 1 的 7 天分工节奏

| 天数 | 主导角色 | 其他角色配合 | 审核角色 |
|------|----------|-------------|---------|
| Day 1 项目冻结 | QA + 后端冻结现状 | 所有人提供清单 | 架构师审核完整性 |
| Day 2 分工和规则 | 架构师写 constitution | 所有人提约束 | 所有人 review |
| Day 3 数据库设计 | Database 输出 schema 草案 | API/Backend 输入字段需求 | 架构师审核域划分 |
| Day 4 API 契约设计 | API Contract 输出 OpenAPI | Frontend 确认调用方式 | 后端确认可行性 |
| Day 5 后端目标架构 | Backend 输出路由/服务拆分 | API/Database 对齐边界 | 架构师审核拆分粒度 |
| Day 6 前端目标架构 | Frontend 输出调用清单和状态计划 | API Contract 确认端点 | Backend 确认数据可用 |
| Day 7 第一批 PR 计划 | 架构师汇总 PR1/2/3 任务单 | 所有角色验收 | 所有人签字确认 |

### 跨角色依赖

```
Database → API Contract: 表结构决定 API response 字段
API Contract → Backend: OpenAPI 决定 handler 签名
API Contract → Frontend: DTO 决定前端 adapter 和类型
Backend → QA: 代码变更决定测试范围
Game Economy → Backend: 公式参数决定 service 行为
Architect → 所有人: 架构决策约束所有角色
```

### AI 协作红线

1. **AI 不得跨角色越权修改文件**。每个 agent 只允许修改其被分配角色对应的文件范围。
2. **AI 不得在 Phase 1 修改业务代码**。所有变更仅限于文档、测试、配置。
3. **AI 的经济公式分析结果必须经人复核**，特别是涉及游戏数值的部分。
4. **AI 生成的迁移计划必须有回退方案**（每个 PR 必须可独立回滚）。
5. **AI 不得删除或重命名任何被前端调用的 API 路由**，除非在 `frontend-api-plan.md` 中确认无引用。
