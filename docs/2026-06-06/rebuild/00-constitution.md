# 重建宪法 — NewHaven 项目第一阶段

**版本**: 1.0  
**日期**: 2026-06-06  
**适用范围**: 后端 (backend/)、前端 (client/)、数据 (decompiled/data/)、配置 (configs/)

---

## 1. 重建目标

1.1. 将当前可运行但基础设施自造的项目升级为工程标准清晰、可扩展的经济游戏后端。  
1.2. 不破坏现有游戏行为，不改动经济数值，不引入玩家可见的回归。  
1.3. 分阶段迁移，每个阶段都能独立部署和回滚。  
1.4. 最终目标技术栈：

```
Router:           go-chi/chi v5
API Contract:     OpenAPI 3.1 + oapi-codegen + openapi-typescript
DB Driver:        pgx/v5 (保持现有)
SQL Layer:        sqlc (后续引入)
Migration:        goose 或 golang-migrate
WebSocket:        coder/websocket
Validation:       go-playground/validator
Logging:          log/slog (Go 标准库)
Server State:     TanStack Query (保持现有)
UI State:         Zustand (保持现有)
Game Render:      PixiJS (保持现有)
```

---

## 2. 不允许做的事情（硬红线）

2.1. **不允许直接删除旧 API。** 旧路由必须保持可用直到确认前端完全迁移。

2.2. **不允许让 runtime 读取 docs/reference。** `docs/` 目录是文档和测试对照快照，不是数据源。

2.3. **不允许把公式写进 handler。** 所有经济公式必须放在 `internal/formula/` 包内，保持纯函数。

2.4. **不允许把 domain model 直接暴露给新 API 的前端。** `model/types.go` 中的类型是领域模型，不是 API DTO。新 API 必须通过 DTO 层转换。

2.5. **不允许一次性重写 market / production / finance。** 每个域必须按"读端点 → 写端点 → scheduler"的顺序逐阶段迁移。

2.6. **不允许在没有测试基线的情况下修改订单撮合、库存扣减、生产队列、债券、ledger。** 必须先有 contract test 或 golden test 证明行为不变。

2.7. **不允许在 Phase 1 修改业务代码。** Phase 1 只加文档、脚本、draft 文件，不改 `service/`、`handler/`、`formula/`、`model/`、`client/src/features/` 的业务逻辑。

2.8. **不允许跳过"冻结现状"直接开工。** 必须先记录当前 API 路由、前端调用点、公式测试、go.mod 依赖锁。

2.9. **不允许 AI 在没有人类确认的情况下合并 PR。** 每个 PR 必须有人类审查且确认验收标准满足。

2.10. **不允许引入 AGPL / GPL 许可证的依赖。** 只允许 MIT、Apache-2.0、BSD、ISC 等宽松许可证。

---

## 3. 允许做的事情

3.1. 新增文档到 `docs/rebuild/`、`docs/api/`、`docs/db/` 等目录。  
3.2. 新增审计脚本到 `scripts/` 目录（route 枚举、依赖检查等）。  
3.3. 新增 OpenAPI spec 到 `openapi/` 目录，不生成代码。  
3.4. 新增数据库 schema 草案到 `db/` 目录，不应用 migration。  
3.5. 新增测试文件（在现有测试包内），不改现有测试逻辑。  
3.6. 新增 chi router 作为兼容桥接层，不改旧 handler 函数体。  
3.7. 新增 DTO 类型定义，不改 domain model。  
3.8. 新增 clock 抽象 (`platform/clock.go`)，不改现有 `time.Now` 调用点（先写 interface，不改消费方）。  
3.9. 升级 go.mod 依赖版本（补丁版本），不改代码。

---

## 4. API 兼容原则

4.1. `/api/v2/*` = 保留兼容。当前前端继续使用，不改响应格式。

4.2. `/api/v3/*` = 保留市场、政府等已有接口。

4.3. `/api/v4/*` = 暂停新增。不再向 `/api/v4` 添加新路由。

4.4. 新接口优先进入 `/api/v3` 或 `/api/v2` 明确归属域，不随手开新版本号。

4.5. 弃用的 API 标记为 `legacy: true` 但继续可用至少 2 个发布周期。

4.6. 错误响应统一格式：
```json
{
  "code": "INSUFFICIENT_FUNDS",
  "message": "Not enough cash.",
  "details": {}
}
```

4.7. 成功的 DTO 响应不再用 `map[string]any`。使用 typed struct。

---

## 5. 公式保护原则

5.1. `internal/formula/` 包永远是纯函数。不接受 `*Service`、`*GameState`、DB 连接等参数。

5.2. 公式的输入必须全部通过函数参数传递，不读全局变量。

5.3. 每个公式必须至少有一个对应的测试用例，覆盖正常值、边界值、零值、负值。

5.4. 公式变更必须同步更新 `docs/2026-06-02/game-formulas-v2.md`（或后续版本文档）。

5.5. 禁止在 handler 或 service 中手算经济公式。遇到 `price * 0.04` 这种 magic number 立即提取到 formula。

---

## 6. 资源 Baseline 保护原则

6.1. `decompiled/data/` 下的 JSON 文件是运行时数据源，有且仅有这 4 个文件：

- `resources.json`
- `buildings.json`
- `economy_model.json`
- `resource_lookups.json`

6.2. `docs/backend-refactor/reference/` 中的 baseline 快照只用于测试对照，不可被运行时读取。

6.3. 数据加载器 (`internal/data/loader.go`) 必须返回 typed struct，当前 `map[string]any` 应逐步替换。

6.4. 新增 JSON 字段必须加默认值兼容逻辑。

---

## 7. 数据库 Migration 原则

7.1. 所有 schema 变更必须通过 migration 文件（SQL 或 Go），不允许直接手改生产数据库。

7.2. migration 工具从 goose 或 golang-migrate 中选择。

7.3. migration 文件名格式：`YYYYMMDD_description.sql`

7.4. 每个 migration 必须可回滚（提供 `down` 语句）。

7.5. Phase 1 不创建 migration 文件，只设计 schema 草案。

---

## 8. OpenAPI 契约原则

8.1. `openapi/openapi-draft.yaml` 是前后端共享契约的唯一真实来源。

8.2. 所有新的 API 端点必须先出现在 OpenAPI spec 中，不允许"先写代码后补文档"。

8.3. DTO 命名规范：

- 请求: `XxxRequest`
- 响应: `XxxResponse`
- 列表: `XxxListResponse`
- 错误码: `XxxErrorCode`

8.4. 响应格式统一：
```json
{
  "data": {},
  "error": null,
  "meta": {}
}
```

---

## 9. 前端状态管理原则

9.1. **服务器状态（buildings、orders、inventory、production jobs、financial data）** → TanStack Query cache。不允许在 Zustand 中缓存服务器数据。

9.2. **UI 状态（activeView、selectedBuildingId、sidebarOpen、chatOpen）** → Zustand。不允许这些状态进入 TanStack Query。

9.3. **WebSocket 推送的事件** → 直接更新 TanStack Query cache（通过 `queryClient.setQueryData`），不经过 Zustand。

9.4. 前端 `api/client.ts` 可以保留 fetch wrapper，但所有请求/响应类型必须来自 `openapi-typescript` 生成的类型。

---

## 10. AI 修改代码规则

10.1. Phase 1（当前阶段）：AI **不允许修改**业务代码。只允许：
- 创建/修改 `docs/` 下的文档
- 创建 `scripts/` 下的审计脚本
- 创建 `openapi/` 下的 draft 文件
- 创建 `db/` 下的 schema 草案

10.2. Phase 2+（后续阶段）：AI 修改代码必须遵守：
- 每个 PR 只改一个域
- 先加测试再改代码
- 不改 not in scope 的文件
- 每个改动必须可以通过 `go test ./...` 和 `go vet ./...`

10.3. AI 在开始修改之前必须先读取最新的 `00-constitution.md`，确认当前阶段的允许范围。

---

## 11. PR 拆分规则

11.1. 每个 PR 必须独立可运行、可回滚。

11.2. PR 不应跨越多个域（例如市场和生产不能在一个 PR 中）。

11.3. PR 大小原则：不超过 20 个文件变更，不超过 500 行新增。

11.4. 每个 PR 必须有明确的验收标准。

11.5. 紧急修复（live bug）可以走 shortcut，但必须事后补文档和测试。

---

## 12. 验收标准

12.1. `go test ./...` 全部通过。  
12.2. `go vet ./...` 全部通过。  
12.3. 前端 `npm run build` 通过（TypeScript 检查 + 生产构建）。  
12.4. 所有旧 API 仍然可调用且响应格式不变。  
12.5. 所有 formula 测试通过且值无变化。  
12.6. 迁移后的 API 响应与迁移前的 contract test 一致。  
12.7. 没有新的 `map[string]any` 出现在 handler 响应中（新建代码）。  
12.8. 没有新的 `strings.TrimPrefix` 路径解析（新建代码）。

---

*本宪法在 Phase 1 结束时由项目所有者审核。后续阶段可修订，但修订必须记录版本号和变更说明。*
