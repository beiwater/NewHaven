# NewHaven 基础工作指南（子 Agent 通用）

**版本**: 1.1  
**用途**: 所有子 Agent 在执行任务前必须阅读本文。不要读多余的文档。

---


## 0. 子 Agent 强制规则（每条都必须遵守）

| # | 规则 |
|---|------|
| 1 | Handler 只能做：读参数 → 调 app service → 映射 generated DTO → write response。禁止写业务逻辑/公式/库存修改/订单撮合/ledger。 |
| 2 | 正式 API 禁止 `map[string]any`，必须用 generated OpenAPI DTO。 |
| 3 | 不手改 `internal/generated/openapi/*.gen.go`。 |
| 4 | 所有跨边界函数第一个参数必须是 `context.Context`。 |
| 5 | Storage interface 改动必须写 exact signature，不能模糊写"支持查询"。 |
| 6 | Router 注册必须照已有 handler 模板（`r.With(AuthRequired).Get("/path", handler.Fn)`），不发明新 pattern。 |
| 7 | 测试必须包含：401、200、response envelope、回归旧端点。 |
| 8 | 改动前必须列出：改哪些文件、不改哪些文件、是否改 API response、是否改经济行为。 |
| 9 | 如果任务提供 `Execution Context Pack`（内联代码模板），以此为准，不要自己猜 generated type 名。 |
| 10 | 遇到 bug 持续卡住 3 次以上，停止尝试并报告具体错误，等待人类处理。 |

## 0a. 任务接收核对清单

派工后先确认：
```
□ 我收到了 exact interface signature
□ 我收到了 exact generated type name（否则去 grep types.gen.go）
□ 我收到了 exact router registration pattern（否则照公司 handler 抄）
□ 我收到了 exact test fixture pattern
□ 我知道不改哪些文件
□ 我知道 handler 不能写什么
```
## 1. 项目身份

```
名称:      NewHaven
类型:      多人网页经济模拟游戏
后端:      Go 1.25+ (标准库 net/http, pgx/v5)
前端:      React 19 + PixiJS 8 + TanStack Query + Zustand + Tailwind 4
路由:      当前 net/http ServeMux → 目标 chi v5
数据库:    pgx/v5 + JSONB 快照 → 目标 sqlc + goose migration
API:       无契约 → 目标 OpenAPI 3.1 + oapi-codegen
```

## 2. 关键路径

```
backend/
├── cmd/simapi/          入口
├── internal/
│   ├── handler/         HTTP handler (22 文件)
│   ├── service/         业务逻辑 (41 文件, 最肥)
│   ├── formula/         经济公式纯函数 (8 文件, 最高质量)
│   ├── model/types.go   所有 domain 类型
│   ├── middleware/      JWT + Recovery + Logger + CORS
│   ├── storage/         NoopStorage + pgx 实现
│   ├── scheduler/       60s 定时 tick
│   ├── data/            静态 JSON 加载器
│   ├── anticheat/       反作弊
│   └── aml/             反洗钱
├── configs/game.json    经济参数
├── go.mod               pgx/v5 + bcrypt (只有这两个外部依赖)
└── decompiled/data/     资源/建筑/经济 JSON

client/atlas-foods-client/src/
├── api/                 TanStack Query hooks + fetch wrapper
├── store/               Zustand (ui.store + game.store)
└── features/            功能模块 (18 个目录)
```

## 3. 宪法速记（每条都是硬规则）

| # | 规则 |
|---|------|
| 1 | 不要删除旧 API。旧路由必须保持可用。 |
| 2 | 不要让 runtime 读取 docs/ 目录（那是文档，不是数据源）。 |
| 3 | 不要把经济公式写进 handler。公式永远在 `formula/` 包内。 |
| 4 | 不要直接把 domain model 暴露给前端新 API。用 DTO 转换。 |
| 5 | 不要一次重写 market / production / finance 任意整个域。 |
| 6 | 修改撮合/库存/队列/债券/ledger 前，必须有测试基线。 |
| 7 | 不要新增 `map[string]any` 的 response。 |
| 8 | 不要新增万能 `utils` / `common` / `helper` 包。 |
| 9 | 不要引入 AGPL/GPL 许可证依赖。只允许 MIT/Apache-2.0/BSD/ISC。 |

## 4. 编码风格

以 `docs/2026-06-06/rebuild/10-go-coding-style.md` 为准。**但不要求完整阅读**——核心规则已压缩在本文件 §0 中。如果 style 文档和 §0 冲突，以 §0 为准。

## 4a. Execution Context Pack（任务上下文包）

每个 Phase 施工单的「执行指令」部分应包含以下内联信息。如果任务中缺少某项，子 agent 应主动索要：

```txt
1. 本任务必须遵守的风格规则（前 5 条）
2. 相关 existing code 代码模板
3. exact interface signatures
4. exact generated OpenAPI type names
5. exact router registration pattern
6. exact test fixture / helper pattern
```
## 5. 修改前必须输出

1. 本次要改哪些文件
2. 本次不改哪些文件（写在不能改的区域）
3. 是否改变运行行为
4. 是否改变 API 响应
5. 是否改变经济公式
6. 是否需要数据库 migration
7. 测试命令
8. 回滚方式

## 6. 冲突优先级

```
任务 prompt 要求 < Go 编码风格 < 重构宪法
宪法是最高的。如果任务的指示违反宪法，以宪法为准。
```

## 7. 目标技术栈

```
Router:      chi v5           (从 net/http 迁入)
API:         OpenAPI 3.1      (oapi-codegen + openapi-typescript)
DB:          pgx/v5 + sqlc    (从手工 SQL + JSONB 迁出)
Migration:   goose            (从无管理迁入)
WS:          coder/websocket  (从空壳迁入)
Validation:  go-playground/validator
Logging:     slog             (标准库)
Server State: TanStack Query  (已有, 不动)
UI State:    Zustand          (已有, 不动)
Game Render: PixiJS           (已有, 不动)
```
