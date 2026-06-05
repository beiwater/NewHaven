# NewHaven 基础工作指南（子 Agent 通用）

**版本**: 1.1  
**用途**: 所有子 Agent 在执行任务前必须阅读本文。不要读多余的文档。

---

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

以 `docs/2026-06-06/rebuild/10-go-coding-style.md` 为准。

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
