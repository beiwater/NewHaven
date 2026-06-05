# Phase 1 首批 PR 计划

> 本文档定义 Phase 1 的前 5 个 PR 的具体范围、改动的文件、验收标准、回滚方式、风险以及需要人工确认的问题。
>
> 语境：New Haven Go 后端重构（单体 → 可测试、可维护架构）。
>
> 当前栈：`net/http ServeMux` + 手动路径解析 + `map[string]any` DTO + pgx + 无 OpenAPI。
>
> 目标栈（Phase 1 不全部落地，仅作为远景）：chi router、OpenAPI + oapi-codegen、sqlc、goose migration、typed DTO。

---

## PR1：文档 + 基线

### 目标

为整个重构周期建立文档基础设施和代码基线快照，不修改任何业务代码。

### 新增文件

| 路径 | 内容 |
|------|------|
| `docs/rebuild/01-goal-and-baseline.md` | 重构目标、非目标、当前架构快照、关键指标（端点数量、service 方法数、test 覆盖率基线） |
| `docs/rebuild/02-api-inventory.md` | 按域名分类的 API 端点清单（路径、方法、handler 函数、前端消费端、是否经测试） |
| `docs/rebuild/03-service-method-inventory.md` | `service.Service` 方法清单，按域名分组，标注锁所有权和 `time.Now()` / ID 生成点 |
| `docs/rebuild/04-risk-register.md` | 高风险区域清单（market order matching、production claim race、auth token 解析等） |
| `docs/rebuild/README.md` | rebuild 文档索引，指向各子 doc 和 PR 计划 |
| `scripts/audit-endpoints.sh` | 从 `internal/handler/*.go` 提取 `Register*` 调用和 `mux.HandleFunc` 注册的脚本，输出端点清单 |
| `scripts/audit-deps.sh` | 扫描 `go.mod` 和 `internal/` 下的 import 依赖图的脚本，输出第三方依赖树 |

### 不改动的文件

- `backend/internal/service/*` — 无任何业务逻辑改动
- `backend/internal/handler/*` — 无任何 handler 改动
- `backend/internal/model/*` — 无数据模型改动
- `backend/cmd/simapi/main.go` — 不修改主入口
- `backend/go.mod` / `go.sum` — 不添加依赖

### 验收标准

1. `scripts/audit-endpoints.sh` 运行后输出完整的端点列表，与 `handler.Register()` 中的注册一致
2. `scripts/audit-deps.sh` 运行后输出依赖树不报错
3. `go test ./...` 在 `backend/` 下通过（基线验证，失败则记录在 `01-goal-and-baseline.md` 中）
4. 所有新增文档路径能被 `docs/rebuild/README.md` 索引到

### 回滚方法

`git revert` 本 PR 即可，无运行时影响。

### 风险

- 极低。仅含文档和 shell 脚本，不改变任何运行时行为。
- 若 `audit-endpoints.sh` 依赖特定的包结构或字符串模式，未来重组 handler 目录时需同步更新。

### 人工确认问题

1. 当前 `go test ./...` 是否已经全部通过？若有失败用例，是否同意在 `01-goal-and-baseline.md` 中记录为已知问题？
2. `audit-endpoints.sh` 提取端点清单的模式（`mux.HandleFunc` / `Register*`）是否覆盖所有注册方式？
3. 是否需要在 `api-inventory.md` 中额外记录每个端点对应的前端组件或路由？

---

## PR2：OpenAPI 草稿

### 目标

创建 OpenAPI 3.1 规范草稿和 DTO 命名约定文档，为后续代码生成做准备，**不运行任何代码生成器**。

### 新增 / 修改文件

| 路径 | 操作 | 内容 |
|------|------|------|
| `docs/openapi/openapi-draft.yaml` | 新增 | OpenAPI 3.1 规范草稿，覆盖 `auth`、`company`、`production`、`market`、`finance`、`research` 域的核心端点 |
| `docs/rebuild/dto-naming-conventions.md` | 新增 | 请求/响应 DTO 命名规范（例如 `CompanyProfileResponse`、`CreateOrderRequest`）、字段类型映射规则、`snake_case`/`camelCase` 选择 |
| `docs/rebuild/openapi-review-checklist.md` | 新增 | 每次审查 OpenAPI 规范时的检查清单（端点覆盖、响应码、字段类型、一致错误格式） |

### 不改动的文件

- `backend/internal/service/*` — 无变更
- `backend/internal/handler/*` — 无变更
- `backend/internal/model/*` — 无变更
- `backend/go.mod` / `go.sum` — **尤其不能引入 oapi-codegen 或任何 codegen 工具**
- `backend/cmd/simapi/main.go` — 不修改

### 验收标准

1. `openapi-draft.yaml` 通过 `redocly lint` 或 `swagger-cli validate` 验证（若本地无这些工具，则用 `npx @redocly/cli lint`）
2. 规范中每个端点至少标注 `operationId`、`summary`、`requestBody`/`parameters`、`responses`（至少一个成功响应和一个错误响应）
3. DTO 命名约定文档明确说明 `*Request` / `*Response` 后缀、数组字段复数形式、时间字段格式（`RFC3339`）
4. openapi-draft.yaml 中引用的所有类型均有 `$ref` 指向 `#/components/schemas/`

### 回滚方法

`git revert`。无运行时影响，因为没有任何代码引用该规范。

### 风险

- 低。规范可能包含与当前实际 API 行为不一致的地方，但这是 draft，允许迭代修正。
- 若规范与前端需求不匹配，可能在后续 PR 中产生返工。风险可接受，因为这是第一次写 openapi，旨在暴露不一致。

### 人工确认问题

1. 当前 API 中哪些端点属于内部调试/开发专用（例如 `/api/dev/...`），是否纳入 OpenAPI 规范？
2. 错误响应格式目前是 `{"error": string}`，是否接受这个格式作为标准，还是需要扩展为 `{"code": int, "message": string, "details": ...}`？
3. 谁负责审查 openapi-draft.yaml 的前端兼容性？是否需要一个前端开发者 review？

---

## PR3：chi Bridge 原型

### 目标

在现有 `net/http ServeMux` 旁**增设** chi 路由器，通过 chi 的 `RouteContext` 实现路径参数解析，旧路由保持 100% 可用，**不改变任何业务逻辑**。

### 新增 / 修改文件

| 路径 | 操作 | 内容 |
|------|------|------|
| `backend/go.mod` | 修改 | 增加 `github.com/go-chi/chi/v5` 依赖 |
| `backend/go.sum` | 自动更新 | 新增 chi 包的校验和 |
| `backend/internal/router/router.go` | 新增 | chi 路由器的初始化函数，注册空路由组占位，返回 `http.Handler` |
| `backend/internal/router/chi_bridge.go` | 新增 | **可选**：将现有 `*Handler` 方法适配为 chi 的 `http.HandlerFunc` 形式；包含桥接测试 |
| `backend/cmd/simapi/main.go` | 修改 | 创建 chi 路由器实例，将其作为顶层 `http.Handler`，通过 `chi.Router` 的 `Handle("/*", oldMux)` 将未知路由回退到旧的 `http.ServeMux` |
| `backend/internal/middleware/middleware.go` | 修改 | 保持现有中间件不变；增加 chi 兼容的 `chi.Router` 中间件版本的简单封装（可选，如果不需要则不做） |

### 不改动的文件

- `backend/internal/service/*` — 全部不改
- `backend/internal/handler/*` — 全部不改；handler 的注册和实现完全不动
- `backend/internal/model/*` — 不改
- `backend/internal/data/*`、`formula/*`、`storage/*`、`scheduler/*` — 不改

### 验收标准

1. `go build ./cmd/simapi` 成功
2. `go vet ./...` 无新警告
3. 服务启动后，`GET /healthz`、`POST /api/auth/login`、`GET /api/v3/company/profile` 等**所有已有端点**仍然返回与 PR1 基线一致的响应（用 curl 或 test 验证至少 5 个关键端点）
4. 新增 `TestChiBridge` 测试用例验证 chi 路由分发正确，旧 ServeMux 路径参数解析行为未受影响
5. 在 chi 路由表中注册一条显式的新路由（例如 `GET /rebuild/health`）后，该路由可访问且不干扰旧路由

### 回滚方法

`git revert` 本 PR。若 `main.go` 修改导致冲突，直接删除 chi 相关代码并回退 `go.mod`。

### 风险

- 中低。若 chi 的 `Handle("/*", oldMux)` 在路径匹配上与 ServeMux 默认行为有微妙差异（例如 404 处理、trailing slash 行为），可能导致某些端点行为变化。
- chi 作为新依赖加入 `go.mod`，若后续弃用 chi 则需移除依赖。
- **关键风险**：chi 的 `RouteContext` 路径参数解析会改变 `r.URL.Path` 的读取方式，若 handler 中同时使用 `r.URL.Path` 做字符串匹配和 chi 的 `URLParam`，可能导致读取到不同的路径片段。需在桥接层确保 handler 层仍然读取原始的 `r.URL.Path`。

### 人工确认问题

1. chi 版本选哪个？推荐 `v5` 的最新 patch（`v5.1.0` 或更高），但需确认与 Go 1.25 的兼容性。
2. 桥接策略：① `chi.Router` 作为顶层，`Handle("/*", oldServeMux)` 兜底；② 还是逐条路由迁移？PR3 采用方案①，方案②留到后续 PR。
3. 是否需要在 chi 层统一添加 `RequestID` / `Logger` / `CORS` / `Recovery` 中间件，还是仍然委托给旧 ServeMux 的中间件链？

---

## PR4：测试基线

### 目标

为最高风险的 handler（market、production）添加合约级 HTTP 测试，确保重构期间 API 行为可被自动验证。

### 新增 / 修改文件

| 路径 | 操作 | 内容 |
|------|------|------|
| `backend/internal/handler/handler_test.go` | 修改 | 重构 test helper：抽 `newTestServer()` 返回 `*httptest.Server`，添加公共断言辅助函数 |
| `backend/internal/handler/market_test.go` | 新增 | `TestMarketDepth`、`TestCreateOrder`、`TestCancelOrder`、`TestMarketTicker` — 使用 `httptest` 验证 JSON 响应结构、HTTP 状态码、关键字段存在 |
| `backend/internal/handler/production_test.go` | 新增 | `TestStartProduction`、`TestClaimProduction` — 覆盖正常流程和边界条件（资源不足、队列满等） |
| `backend/internal/handler/company_test.go` | 新增 | `TestCompanyProfile`、`TestLevelUp` — 覆盖公司信息获取和升级流程 |
| `backend/internal/handler/auth_test.go` | 新增 | `TestLogin`、`TestRegister` — 覆盖认证流程的正向和异常路径 |
| `backend/internal/handler/health_test.go` | 新增 | `TestHealthz` — 简单存活检测 |
| `docs/rebuild/test-coverage-gap.md` | 新增 | 记录当前测试未覆盖的 handler 路径和决策（例如：哪些路由暂不写测试、原因是什么） |
| `scripts/run-handler-tests.sh` | 新增 | 一键运行所有 handler 测试的脚本 |

### 不改动的文件

- `backend/internal/service/*` — 不改，但测试中会用 `service.NewTestService()`
- `backend/internal/handler/*.go` — **不修改 handler 的业务逻辑**，仅添加 `_test.go` 文件
- `backend/internal/model/*` — 不改
- `backend/internal/formula/*`、`storage/*` — 不改

### 验收标准

1. `cd backend && go test ./internal/handler/ -v -count=1` 全部通过
2. 每个新增测试文件至少包含 3 个测试用例（1 正向 + 1 异常 + 1 边界）
3. 测试使用 `httptest.Server` 而非直接调用 handler 函数，确保中间件链路完整
4. 测试不依赖外部数据库（使用 `service.NewTestService()` 的内存模式）
5. `test-coverage-gap.md` 明确列出本 PR **未覆盖**的端点及原因

### 回滚方法

`git revert`。测试文件不影响生产编译产物。

### 风险

- 低。纯新增文件 + 对 `handler_test.go` 的辅助函数重构。
- 若 `NewTestService()` 返回的测试服务状态与生产路径不一致，可能导致误报。需确保测试 service 使用了和 main.go 相同的 `cfg` 默认值。

### 人工确认问题

1. `NewTestService()` 的行为是否稳定？是否所有 handler 都能用它构造完整的测试环境？
2. 测试覆盖率目标：本 PR 的目标是覆盖**高风险**端点，还是覆盖**所有**端点？建议只覆盖 market、production、company、auth 四个域，其他域留到后续 PR。
3. 是否需要对未导出的 handler（如 `handleMarketTicker`）做测试？建议通过路由路径而非函数名进行测试。

---

## PR5：数据库 Schema 草稿

### 目标

产出仅包含文档的数据库 schema 设计草案和迁移策略，**不切换任何运行时存储方式**。

### 新增文件

| 路径 | 内容 |
|------|------|
| `docs/db/schema-draft.sql` | 完整的 PostgreSQL schema 草稿：表定义（`companies`、`buildings`、`productions`、`orders`、`trades`、`bonds`、`research_projects` 等），索引，外键约束 |
| `docs/db/migration-plan.md` | 从当前内存模式迁移到 PostgreSQL 持久化的分阶段计划：数据迁移顺序、回退策略、零停机切换条件 |
| `docs/db/entity-relationship.md` | ER 图文字描述和字段映射：当前 `model.GameState` / `model.Company` 等 Go 结构体如何映射到 SQL 表 |
| `scripts/audit-sqlc-compatibility.sh` | 检查 `schema-draft.sql` 是否与 sqlc 兼容的脚本（语法验证，不生成代码） |

### 不改动的文件

- **所有** `backend/internal/` 下的文件 — 不触碰
- `backend/go.mod` — 不加 goose/sqlc 依赖
- `backend/cmd/simapi/main.go` — 不修改
- `backend/internal/storage/storage.go`、`postgres.go` — 不修改

### 验收标准

1. `schema-draft.sql` 经过至少一次 PostgreSQL 语法验证（`psql -f schema-draft.sql` 或 `pg_validate`），报告无语法错误
2. 迁移计划文档说明每步的变更内容、回退 SQL、预期的数据兼容性
3. 实体关系文档中每个表都标注了对应的 Go 结构体路径（如 `model.Company` → `companies` 表）
4. `audit-sqlc-compatibility.sh` 验证 schema-draft.sql 无 sqlc 不支持的语法（如不支持的 PostgreSQL 扩展、不支持的类型等）

### 回滚方法

`git revert`。仅为文档，无运行时依赖。

### 风险

- 极低。纯文档 + shell 脚本验证。
- schema-draft.sql 可能与未来实际的代码生成需求有偏差，但这是 draft，允许修改。
- 若 `schema-draft.sql` 中的表结构与现有的 `internal/storage/postgres.go` 中的 SQL 查询不一致，需在后续 PR 中协调。

### 人工确认问题

1. `schema-draft.sql` 是否应该考虑现有的 `storage` 包中已实现的 PostgreSQL 查询（`internal/storage/postgres.go`），确保草稿表结构与其兼容？
2. 迁移路径选择：是先「db schema draft → sqlc 生成 → 存储层适配」还是先「存储层整理 → db schema 定型」？建议前者，但需确认。
3. 是否需要在 schema 中预留 `version` / `created_at` / `updated_at` 列作为公共审计字段？

---

## 先做哪个 PR？为什么？

**从 PR1（文档 + 基线）开始。**

理由：

1. **零风险入门。** PR1 只改文档和 shell 脚本，不碰任何 Go 代码、依赖、运行时路径。即使出错，回滚代价为零。
2. **建立知识基线。** 在开始任何实质性改动前，PR1 的输出（API 清单、service 方法清单、风险登记表）为后续所有 PR 提供了参考锚点。没有这个基线，后续 PR 的验收标准无法客观衡量（"这段代码之前长什么样？"）。
3. **暴露隐藏问题。** 写 audit 脚本和 API 清单的过程中，会自然发现：哪些端点没有被测试、哪些路径有冲突、哪些 handler 之间存在隐含耦合。这些问题在 PR1 阶段暴露出来，比在 PR3（chi 桥接）或 PR4（测试基线）阶段才发现要便宜得多。
4. **为并行奠基。** PR1 完成后，PR2（OpenAPI 草稿）、PR3（chi 原型）、PR4（测试基线）可以并行展开，因为它们的改动域不重叠，且都以 PR1 的文档为共同参考。

**建议的执行流水线：**

```
PR1 (Docs + Baseline)
    ├─ PR2 (OpenAPI Draft)
    ├─ PR3 (chi Bridge)
    └─ PR4 (Test Baseline)
            └─ PR5 (Database Schema Draft)
```

PR2、PR3、PR4 在 PR1 合并后可并行启动。PR5 依赖 PR4 的测试基础设施来验证 schema 兼容性，建议排在 PR4 之后。
