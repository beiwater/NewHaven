# Game Project Index

> 项目文件索引 — 最后更新: 2026-06-02

```
game/
├── backend/          Go 后端 API (原 go-sim-api)
├── client/           React 前端 (atlas-foods-client)
├── assets/           UI/美术资源
│   ├── source/       源文件 (生成用 PSD/PNG/提示词)
│   └── game/         运行时资源副本
├── docs/             项目文档合集
├── scripts/          工具脚本
├── plans/            架构/技术决策文档
└── dev/              开发产物 (日志/反编译数据/覆盖率/截图)
```

---

## backend/ — Go 后端

> 路径: `backend/`
> 进入: `go run ./cmd/simapi/` (端口 8088)

| 目录 | 说明 |
|------|------|
| `cmd/simapi/` | 入口 main.go |
| `internal/config/` | 环境变量配置 (27 项) |
| `internal/middleware/` | HTTP 中间件 (Recovery, Logger, CORS, Auth) |
| `internal/handler/` | HTTP 处理器 (By domain: company, market, bond, etc.) |
| `internal/service/` | 业务逻辑层 |
| `internal/model/` | 数据模型 (`types.go`) |
| `internal/storage/` | PostgreSQL 持久化 (可选, 默认内存) |
| `internal/data/` | JSON 静态数据加载 |
| `internal/formula/` | 经济公式纯函数 |
| `internal/scheduler/` | 定时任务 (市场撮合 / 机器人) |
| `internal/anticheat/` | 反作弊检测 |
| `internal/aml/` | 反洗钱逻辑 |
| `configs/` | 运行时配置 (game.json) |
| `tests/` | 集成测试 |

**模块:** `go-sim-api` (Go 1.25)

---

## client/ — React 前端

> 路径: `client/atlas-foods-client/`
> 进入: `pnpm dev` (端口 5173)

基于 **Vite + React 19 + TypeScript + PixiJS + Zustand + TanStack Query + Tailwind CSS**。

| 目录 | 说明 |
|------|------|
| `src/app/` | 应用壳 (App.tsx, providers, ErrorBoundary) |
| `src/api/` | API 客户端及领域 API |
| `src/game/` | 游戏地图 (PixiJS 渲染层) |
| `src/features/` | 业务功能模块 |
| `src/features/market/` | 市场页面/报价牌 |
| `src/features/buildings/` | 建筑面板/卡片/商店 |
| `src/features/production/` | 生产队列 |
| `src/features/inventory/` | 库存条 |
| `src/features/contracts/` | 合同列表 |
| `src/features/chat/` | 聊天面板 |
| `src/features/auth/` | 登录鉴权 |
| `src/features/topbar/` | 顶栏 |
| `src/features/sidebar/` | 侧边栏 |
| `src/features/ui/` | 通用 UI 组件 |
| `src/store/` | Zustand 状态管理 |
| `src/styles/` | 全局样式 |
| `public/assets/` | 运行时资源 (图标/物品/建筑/背景) |

---

## assets/ — 美术资源

### assets/source/ — 源文件

AI 生成/手工制作的原始素材与中间产物:

| 目录 | 说明 |
|------|------|
| `ui-icons/` | UI 图标源文件 (含分层 PSD/slice 切片/透明 PNG) |
| `items/` | 物品资源 (小麦/面粉/面包/套餐) |
| `generated/` | AI 生成的原图 (chromakey) |
| `transparent/` | 建筑去背 PNG |
| `backgrounds/` | 地图背景 |
| `avatars/` | NPC 头像 |
| `spritesheets/` | 精灵图集 |
| `prompts/` | AI 绘图提示词 |

### assets/game/ — 运行时资源

从 `client/public/assets/` 复制的运行时资源副本，方便美术人员直接预览:
- `icons/` — 16 枚功能图标
- `items/` — 5 枚物品图
- `buildings/` — 4 栋建筑图
- `backgrounds/` — 1 张地图背景

---

## docs/ — 文档

| 文件 | 说明 |
|------|------|
| `api-contract.md` | API 接口契约 |
| `game-wiki.md` | 游戏维基 |
| `requirements.md` | 需求文档 |
| `newbie-tutorial.md` | 新手教程 |
| `game-materials.md` | 项目材料总览 (Binder) |
| `restaurant-math-model.md` | 餐厅数学模型 |
| `gui.md` | GUI 设计说明 |
| `ECONOMY_SYSTEM.md` | 经济系统设计 |
| `project-conventions.md` | 项目规范 |
| `game-art-pipeline/` | 美术流水线文档 (00-vision ~ 07-generation-log) |

---

## scripts/ — 工具脚本

| 文件 | 说明 |
|------|------|
| `count_loc.go` | Go 代码行数统计 |

---

## plans/ — 架构决策

| 文件 | 说明 |
|------|------|
| `frontend-tech-decision.md` | 前端技术选型分析 (React+PixiJS vs Unity/Phaser) |

---

## dev/ — 开发产物

| 目录 | 说明 |
|------|------|
| `logs/` | 运行日志 + 日志分析报告 |
| `screenshots/` | 前端截图 |
| `decompiled/` | 反编译参考数据 (只读) |
| `cover-reports/` | Go 测试覆盖率报告 |
| `.vite/` | Vite 依赖缓存 |
| `simapi.exe` | Go 后端编译产物 |

---

## 目录迁移记录

| 原路径 | 新路径 |
|--------|--------|
| `go-sim-api/` | `backend/` |
| `go-sim-api/docs/` | `docs/` (合并) |
| `go-sim-api/scripts/count_loc.go` | `scripts/count_loc.go` |
| `农场小游戏 前端/atlas-foods-client/` | `client/atlas-foods-client/` |
| `农场小游戏 前端/index.md` | `plans/frontend-tech-decision.md` |
| `农场小游戏 前端/art-src/` | `assets/source/` (合并) |
| `农场小游戏 前端/log/` | `dev/logs/` |
| `农场小游戏 前端/screenshot*.png` | `dev/screenshots/` |
| `art-src/` | `assets/source/` (合并) |
| `decompiled/` | `dev/decompiled/` |
| `docs/restaurant-math-model.md` 等 | `docs/` (保留原位置) |
