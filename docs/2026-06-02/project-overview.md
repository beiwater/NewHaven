# Project Overview

**Date:** 2026-06-02  
**Generated from:** codebase static audit (no runtime required)

---

## Project Structure

```
NewHaven/
├── backend/           Go API 服务 (59 Go 文件)
├── client/            React + PixiJS 前端 (28 TS/TSX 文件)
├── assets/            美术资源 (90 PNG 文件)
├── docs/              文档
├── scripts/           工具脚本
├── plans/             架构决策
└── dev/               开发产物
```

---

## Backend — Go API (`go-sim-api`)

### 技术栈
Go 1.25, 标准库 `net/http.ServeMux`, 内存存储 (可选 PostgreSQL via pgx/v5)

### 整体规模
| 统计 | 数量 |
|------|------|
| Go 源文件 | 59 |
| Handler 文件 | 20 |
| Service 文件 | 17 |
| 注册的 API 端点 | ~80+ |

### API 端点全表

#### 认证 (handler/auth.go)
| 方法 | 路径 | 状态 |
|------|------|------|
| POST | `/api/register` | 前端使用 |
| POST | `/api/login` | 前端使用 |

#### 公司 (handler/company.go + 部分 executive/auction 挂在里面注册)
| 方法 | 路径 | 状态 |
|------|------|------|
| GET | `/api/csrf/` | 前端未使用 |
| GET | `/api/v2/players/me/companies/` | 前端未使用 |
| GET | `/api/v2/players/{id}/` | 前端未使用 |
| GET/PUT | `/api/v2/companies/me/preferences/` | 前端未使用 |
| GET | `/api/v2/companies/me/buildings/` | 前端使用 |
| GET | `/api/v2/companies/me/administration-overhead/` | 前端未使用 |
| GET | `/api/v3/companies/{id}/` | 前端使用 |
| GET | `/api/v2/companies/{id}/` | 前端未使用 |
| GET | `/api/v2/companies/me/collectibles/` | 前端未使用 |
| GET | `/api/v2/companies/me/notifications/` | 前端未使用 |
| GET | `/api/v2/companies/me/market-orders/` | 前端未使用 |
| GET | `/api/v2/companies/me/certificates/` | 前端未使用 |
| GET | `/api/v2/companies/me/display-case/` | 前端未使用 |
| GET | `/api/v2/companies/me/former-executives/` | 前端未使用 |
| GET | `/api/v2/companies/me/royalties/` | 前端未使用 |
| GET | `/api/v2/companies/me/egg-collection/` | 前端未使用 |
| GET | `/api/v2/companies/me/tags/` | 前端未使用 |
| GET | `/api/v2/companies/me/auctions/` | 前端未使用 |
| GET/POST | `/api/v2/companies/me/warehouse/` | 前端使用 (GET) |

#### 玩家/等级 (handler/player.go)
| 方法 | 路径 | 状态 |
|------|------|------|
| GET | `/api/v2/companies/me/achievements/` | 前端未使用 |
| GET | `/api/v2/no-cache/companies/me/achievements/` | 前端未使用 |
| DELETE | `/api/v2/no-cache/companies/achievements/` | 前端未使用 |
| POST | `/api/v2/players/simboosts-use/` | 前端未使用 |
| GET | `/api/v2/players/simboosts/` | 前端未使用 |
| GET | `/api/v2/players/unlocked-hqs/` | 前端未使用 |
| GET | `/api/v2/players/devices/` | 前端未使用 |
| GET | `/api/v2/players/me/level/` | 前端使用 |
| POST | `/api/v2/players/me/xp/` | 前端未使用 |
| GET | `/api/v2/players/me/level-rewards/` | 前端未使用 |
| GET | `/api/v2/players/me/offline-income/` | 前端未使用 |

#### 市场 (handler/market.go)
| 方法 | 路径 | 状态 |
|------|------|------|
| GET | `/api/v3/market-ticker/{id}/` | 前端使用 |
| GET | `/api/v3/market/{id}/{quality}/` | 前端使用 |
| POST | `/api/v2/market-order/` | 前端使用 |
| POST | `/api/v2/market-order/cancel/{id}/` | 前端使用 |
| GET | `/api/market/buy/orders/` | 前端未使用 |
| POST | `/api/v2/market-order/take/` | 前端使用 |
| GET | `/api/v2/weather/` | 前端未使用 |
| GET | `/api/v2/production-modifiers/` | 前端未使用 |
| GET | `/api/v3/market-depth/{id}/{quality}/` | 前端使用 |
| GET | `/api/v3/resources/` | 前端使用 |
| GET | `/api/v3/resources-info/{id}/` | 前端未使用 |

#### 生产 (handler/production.go)
| 方法 | 路径 | 状态 |
|------|------|------|
| GET/POST | `/api/v1/buildings/` | 前端使用 (GET detail) |
| GET/POST | `/api/v2/buildings/` | 前端使用 (production-options) |
| GET | `/api/v2/production/jobs/` | 前端使用 |
| POST | `/api/v2/production/claim/` | 前端使用 |
| GET | `/api/v2/production/claimable/` | 前端使用 |
| POST | `/api/v2/production/claim-all/` | 前端使用 |

#### 生产队列 (handler/production_queue.go)
| 方法 | 路径 | 状态 |
|------|------|------|
| GET | `/api/v2/production/queue/` | 前端使用 |
| POST | `/api/v2/production/slots/add/` | 前端未使用 |
| POST | `/api/v2/production/cancel/` | 前端使用 |

#### 金融 (handler/financial.go)
| 方法 | 路径 | 状态 |
|------|------|------|
| GET | `/api/v2/companies/me/income-statement/` | 前端未使用 |
| GET | `/api/v2/companies/me/balance-sheet/` | 前端未使用 |
| GET | `/api/v2/companies/me/cashflow-statement/` | 前端未使用 |
| GET | `/api/v2/companies/me/cashflow/recent/` | 前端未使用 |
| GET | `/api/v2/companies/me/past-finances-overview/` | 前端未使用 |
| GET | `/api/v3/companies/me/past-finances/` | 前端未使用 |

#### 债券 (handler/bond.go)
| 方法 | 路径 | 状态 |
|------|------|------|
| POST | `/api/bonds/settle-interest/` | 前端未使用 |
| GET/POST | `/api/bonds/` | 前端未使用 |
| GET | `/api/v2/companies/me/bonds/owned/` | 前端未使用 |
| GET | `/api/v2/companies/me/bonds/sold/` | 前端未使用 |

#### 政府合同 (handler/government.go)
| 方法 | 路径 | 状态 |
|------|------|------|
| GET | `/api/v3/government-orders/` | 前端使用 |
| POST | `/api/v3/government-orders/bid/` | 前端使用 |
| POST | `/api/v3/government-orders/award/` | Scheduler 使用，前端未使用 |
| POST | `/api/v3/government-orders/deliver/` | 前端未使用 |
| POST | `/api/v3/government-orders/resolve-defaults/` | Scheduler 使用，前端未使用 |

#### 消息/聊天 (handler/message.go)
| 方法 | 路径 | 状态 |
|------|------|------|
| GET/POST | `/api/messages/` | 前端未使用 |
| GET | `/api/messages_by_company/` | 前端未使用 |
| POST | `/api/v2/message/` | 前端未使用 |
| GET | `/api/v2/message/{id}/read/` | 前端未使用 |
| GET | `/api/v2/chatroom/` | 前端未使用 |
| GET | `/api/v2/contacts/` | 前端未使用 |
| GET | `/api/v2/newspaper/articles-by-author/` | 前端未使用 |
| GET/POST | `/api/v2/newspaper/articles/` | 前端未使用 |
| GET | `/api/v2/newspaper/publishing-costs/` | 前端未使用 |

#### 配方 (handler/recipe.go)
| 方法 | 路径 | 状态 |
|------|------|------|
| GET | `/api/v2/recipes/` | 前端未使用 |
| GET | `/api/v2/recipes/{id}/` | 前端未使用 |

#### 高管 (handler/executive.go)
| 方法 | 路径 | 状态 |
|------|------|------|
| POST | `/api/v2/executives/search/` | 前端未使用 |
| POST | `/api/v2/executives/recruit/` | 前端未使用 |
| POST | `/api/v2/executives/train/{id}/` | 前端未使用 |
| POST | `/api/v3/executives/poach/` | 前端未使用 |
| GET/POST | `/api/v3/executives/offers/` | 前端未使用 |
| GET | `/api/v3/executives/{id}/` | 前端未使用 |

#### 建筑商店 (handler/building_shop.go)
| 方法 | 路径 | 状态 |
|------|------|------|
| GET | `/api/v2/buildings/market/` | 前端未使用（BuildView 用`/api/v2/buildings/` 但实际 API 注册路径不同） |
| POST | `/api/v2/buildings/buy/` | 前端使用 |
| POST | `/api/v2/buildings/place/` | 前端使用 |
| POST | `/api/v2/buildings/move/` | 前端未使用 |
| POST | `/api/v2/buildings/demolish/` | 前端未使用 |
| POST | `/api/v2/companies/me/warehouse/upgrade/` | 前端未使用 |

#### 日常订单 (handler/order.go)
| 方法 | 路径 | 状态 |
|------|------|------|
| GET | `/api/v2/orders/daily/` | 前端使用 |
| POST | `/api/v2/orders/daily/complete/` | 前端使用 |
| POST | `/api/v2/orders/daily/claim/` | 前端使用 |

#### 航空/火箭 (handler/aerospace.go)
| 方法 | 路径 | 状态 |
|------|------|------|
| GET | `/api/v2/aerospace/projects/` | 前端未使用 |
| POST | `/api/v2/aerospace/projects/create/` | 前端未使用 |
| GET | `/api/v2/aerospace/launches/` | 前端未使用 |
| POST | `/api/v2/aerospace/launch/` | 前端未使用 |
| GET | `/api/v2/aerospace/components/` | 前端未使用 |

#### 拍卖 (handler/auction.go)
| 方法 | 路径 | 状态 |
|------|------|------|
| GET | `/api/v2/auctions/` | 前端未使用 |
| GET | `/api/v2/auctions/{id}/` | 前端未使用 |
| POST | `/api/v2/auctions/{id}/bid/` | 前端未使用 |

#### 研发 (handler/dev.go)
| 方法 | 路径 | 状态 |
|------|------|------|
| GET | `/api/dev/ledger/` | 前端未使用 |
| GET | `/api/dev/formulas/production/` | 前端未使用 |
| GET | `/api/dev/formulas/retail/` | 前端未使用 |
| GET | `/api/dev/formulas/retail-season-weather/` | 前端未使用 |
| GET | `/api/v4/` | 前端未使用 |
| GET | `/api/v3/contracts-incoming/` | 前端未使用 |
| GET | `/api/v3/contracts-outgoing/me/` | 前端未使用 |
| GET | `/api/v2/contracts-history-incoming/` | 前端未使用 |
| GET | `/api/v2/contracts-history-outgoing/` | 前端未使用 |
| GET | `/api/v2/warehouse-contracts-summary/` | 前端未使用 |
| GET | `/api/v2/research/` | 前端未使用 |
| POST | `/api/v2/research/start/` | 前端未使用 |
| GET | `/api/v2/research/progress/` | 前端未使用 |
| POST | `/api/v2/research/complete/` | 前端未使用 |
| POST | `/api/dev/time/` | 前端未使用 |

#### 健康检查 (handler/health.go)
| 方法 | 路径 | 状态 |
|------|------|------|
| GET | `/healthz` | 基础设施 |
| GET | `/readyz` | 基础设施 |

### 后台定时任务 (Scheduler)
每 60 秒执行一次:
1. 债券利息结算
2. 政府合同授标
3. 政府违约处理
4. 机器人市场循环
5. 市场锁定检查 + 国家队介入
6. 过期订单清理
7. 生产任务刷新
8. 日常订单刷新
9. 持久化保存

### 数据模型 (model/types.go)
| 模型 | 描述 |
|------|------|
| `ResourceAmount` | 资源数量 (ResourceID + Quality + Amount) |
| `Bond` | 债券 (发行人, 金额, 利率, 评级, 状态) |
| `GovContract` | 政府合同 (需求资源, 数量, 竞标, 交付) |
| `LedgerEntry` | 财务流水 (方向, 金额, 分类, 时间) |
| `Notification` | 通知 |
| `Message` | 聊天消息 |
| `MarketOrder` | 市场订单 (方向, 资源, 价格, 数量, 品质) |
| `Trade` | 成交记录 |
| `Company` | 公司 (资金, 库存, 建筑, 等级) |
| `Player` | 玩家 (用户名, token, 公司列表) |
| `ProductionJob` | 生产任务 (建筑, 资源, 开始/结束时间, 状态) |
| `Auction` | 建筑拍卖 |
| `Order` | 日常订单 |
| `AuctionBid` | 拍卖出价 |
| `ResearchProject` | 研发项目 |
| `GameState` | 全局游戏状态 (以上所有集合) |

### 公式系统 (formula/)
| 文件 | 功能 |
|------|------|
| `production.go` | 生产效率/时间计算 |
| `market.go` | 市场相关 |
| `bonds.go` | 债券面值/利息/发行上限 |
| `retail.go` | 零售公式 |
| `admin.go` | 管理费公式 |

### 静态游戏数据 (from `decompiled/data/*.json`)
- `resources.json` — 所有资源定义 (ID, 名称, 配方, 类型)
- `buildings.json` — 建筑定义 (类型, 成本, 产出)
- `economy_model.json` — 经济模型参数
- `resource_lookups.json` — 资源查找映射

### 后端其他子系统
| 模块 | 功能 |
|------|------|
| `anticheat/` | 反作弊速率检测 |
| `aml/` | 反洗钱 |
| `storage/` | PostgreSQL 持久化 (可选) |
| `middleware/` | Recovery, Logger, CORS, RequestID, Auth |
| `config/` | 27 个环境变量配置 |

---

## Frontend — React + PixiJS (`atlas-foods-client`)

### 技术栈
Vite + React 19 + TypeScript + PixiJS + Zustand + TanStack Query + Tailwind CSS

### 整体规模
| 统计 | 数量 |
|------|------|
| TS/TSX 源文件 | 28 |
| 页面/视图 | 6 个主视图 |
| API 调用模块 | 7 个 |
| 自定义类型 | 11 个接口 |

### 页面/功能视图

| 视图 | 组件 | 描述 | API 覆盖 |
|------|------|------|----------|
| `map` | `GameCanvas` + PixiJS | 游戏地图 (建筑渲染, 交互) | `buildings` |
| `build` | `BuildView` | 购买/放置建筑 | `buildings/buy`, `buildings/place`, `buildings/market` |
| `warehouse` | `InventoryBar` | 仓库库存 (在 sidebar 中显示) | `warehouse` |
| `market` | `MarketPage` | 市场行情/深度/挂单/交易 | `market-ticker`, `market-depth`, `market`, `market-order`, `resources` |
| `contracts` | `ContractList` | 日常订单 + 政府合同 | `orders/daily`, `government-orders` |
| `research` | — | 侧边栏有入口但无实际页面内容 | **无 API 调用** |

### 面板块 (非视图, 一直显示)

| 组件 | 位置 | 描述 |
|------|------|------|
| `TopBar` | 顶栏 | 公司名, 资金, 等级/XP, 能量, 工人 |
| `LeftSidebar` | 左侧导航 | 6 个功能区图标导航 |
| `BuildingPanel` | 右侧面板 | 建筑列表 + 选择后显示 BuildingCard |
| `ChatPanel` | 浮动窗口 | 聊天 (目前仅 mock 数据) |
| `AuthGate` | 登录页 | 登录/注册表单 |

### 前端 API 调用总表 (实际发起的请求)

| 端点 | 组件 | 刷新间隔 |
|------|------|----------|
| GET `/api/v3/companies/{id}/` | TopBar | 30s |
| GET `/api/v2/players/me/level/` | TopBar | 30s |
| GET `/api/v3/market-ticker/{id}/` | MarketPage | 30s |
| GET `/api/v3/market-depth/{id}/{quality}/` | MarketPage | 15s |
| GET `/api/v3/market/{id}/{quality}/` | MarketPage | 15s |
| GET `/api/v3/resources/` | MarketPage | 10min (stale) |
| POST `/api/v2/market-order/` | MarketPage | — (mutation) |
| DELETE `/api/v2/market-order/cancel/{id}/` | MarketPage | — (mutation) |
| POST `/api/v2/market-order/take/` | MarketPage | — (mutation) |
| GET `/api/v2/companies/me/warehouse/` | InventoryBar | — |
| GET `/api/v2/companies/me/buildings/` | GameCanvas, BuildingPanel, BuildView | — |
| GET `/api/v2/production/jobs/` | BuildingCard | — |
| GET `/api/v2/production/queue/` | BuildingPanel | — |
| GET `/api/v2/production/claimable/` | BuildingPanel | — |
| POST `/api/v2/production/start/` | BuildingCard | — (mutation) |
| POST `/api/v2/production/claim/` | BuildingCard | — (mutation) |
| POST `/api/v2/production/claim-all/` | BuildingPanel | — (mutation) |
| POST `/api/v2/production/cancel/` | BuildingCard | — (mutation) |
| GET `/api/v2/buildings/{id}/production-options/` | BuildingCard | — |
| GET `/api/v1/buildings/{id}/` | BuildingCard | — |
| POST `/api/v2/buildings/buy/` | BuildView | — (mutation) |
| POST `/api/v2/buildings/place/` | BuildView | — (mutation) |
| GET `/api/v2/buildings/market/` | BuildView | — |
| GET `/api/v2/orders/daily/` | ContractList | — |
| POST `/api/v2/orders/daily/complete/` | ContractList | — (mutation) |
| POST `/api/v2/orders/daily/claim/` | ContractList | — (mutation) |
| GET `/api/v3/government-orders/` | ContractList | — |
| POST `/api/v3/government-orders/bid/` | ContractList | — (mutation) |
| POST `/api/login` | AuthGate | — (mutation) |
| POST `/api/register` | AuthGate | — (mutation) |

### 前端已实现但后端 API 缺失的
- **WebSocket/实时推送**: 前端有 `websocket.ts` 的 hooks (`useMarketWebSocket`, `useProductionWebSocket`) 但后端未实现 WS 端点，实际为空操作。
- **ChatPanel**: 聊天 UI 完整 (包括消息输入、发送按钮)，但只用本地 mock 数据，未连接后端 `/api/v2/message/` 或 WebSocket。

### 前端状态管理
| Store | 用途 |
|-------|------|
| `ui.store.ts` | activeView (6 个视图), selectedBuildingId, sidebar/market/chat 开关 |
| `game.store.ts` | PixiJS 地图 zoom/pan, 选中建筑, 生产计时 tick |
| `building.store.ts` | 建筑相关状态 |

### PixiJS 游戏地图
| 文件 | 功能 |
|------|------|
| `GameCanvas.tsx` | 主组件, 加载建筑数据, 渲染地图 |
| `createApp.ts` | 初始化 PixiJS Application |
| `mapScene.ts` | 创建地图背景 |
| `buildingLayer.ts` | 建筑精灵 + 选中高亮 + 等级/名称/状态标签 |
| `interactionLayer.ts` | 交互层 |
| `progressLayer.ts` | 进度层 |

---

## Gap Analysis: 后端有前端没有 vs 前端有后端没有

### 后端有完整实现但前端未接入的子系统

| 子系统 | 后端端点数 | 前端状态 | 影响 |
|--------|-----------|---------|------|
| **财务 (Financial)** | 6 | ❌ 完全未调用 | 三大报表、现金流、历史财务 |
| **债券 (Bond)** | 6 | ❌ 完全未调用 | 债券发行、购买、赎回、利息 |
| **高管 (Executive)** | 6 | ❌ 完全未调用 | 搜索/招募/培训/挖角/报价 |
| **拍卖 (Auction)** | 3 | ❌ 完全未调用 | 建筑竞标 |
| **航空/火箭 (Aerospace)** | 5 | ❌ 完全未调用 | 火箭项目/发射/组件 |
| **研发 (Research)** | 4 | ⚠️ 仅侧边栏入口 | 研究项目/进度/完成 |
| **消息/聊天 (Message)** | 8 | ⚠️ 仅 mock UI | 聊天、报纸、联系人 |
| **SimBoost** | 3 | ❌ 完全未调用 | 加速道具 |
| **成就 (Achievement)** | 3 | ❌ 完全未调用 | 成就系统 |
| **等级奖励** | 1 | ❌ 完全未调用 | 等级奖励领取 |
| **离线收入** | 1 | ❌ 完全未调用 | 离线收益计算 |
| **天气/生产修正** | 2 | ❌ 完全未调用 | 天气预报、生产修正系数 |

### 后端实现但前端只用了部分的功能

| 功能 | 后端实现 | 前端使用 |
|------|---------|---------|
| 建筑详情 (v1) | `GET /api/v1/buildings/{id}/` | 只获取详情, 未调用 upgrade/move/demolish |
| 建筑市场 | `GET /api/v2/buildings/market/` | BuildView 通过 `useQuery` 调用 |
| 政府合同授标/交付/违约 | 完整 CRUD | 仅 bid + 列表 |
| 仓库 | 完整仓库 CRUD | 仅 GET 查看, 未用 upgrade |
| 玩家级别 | 级别+XP+奖励 | 仅级别查看, 未用 addXP/rewards/offline-income |

### 前端有而后端没有对应支持的功能

| 功能 | 描述 | 后端状态 |
|------|------|---------|
| **PixiJS 游戏地图** | 地图渲染、平移缩放、建筑交互 | N/A (纯前端渲染) |
| **WebSocket 实时推送** | 前端有 hooks 但无后端 WS 端点 | ❌ 后端未实现 |
| **聊天 UI** | 完整的聊天窗口 UI | 后端有 REST API (`/api/v2/message/`) 但前端未对接 |
| **公司声望硬编码** | 侧边栏底部 "Company Reputation" 显示固定值 5 | 后端无此模型 |
| **能量/工人** | TopBar 显示固定值 120/120 和 8/10 | 后端无此概念 |

---

## Items & Assets Mapping

### 资源/物品 (Resources)

前端 `resources.ts` 定义了 23 种资源和 3 个市场分组:

| ID | 名称 | 分组 | 有物品图? |
|----|------|------|-----------|
| 1 | Power | Staples | ❌ |
| 2 | Water | Staples | ❌ |
| 3 | Apples | Farm Goods | ❌ |
| 4 | Oranges | Farm Goods | ❌ |
| 5 | Grapes | Farm Goods | ❌ |
| 6 | Grain | Staples | ✅ `item_wheat_v1.png` |
| 7 | Steak | Kitchen Chain | ❌ |
| 8 | Sausages | Kitchen Chain | ❌ |
| 9 | Eggs | Farm Goods | ❌ |
| 66 | Seeds | Staples | ❌ |
| 72 | Sugarcane | Staples | ❌ |
| 115 | Cow | Farm Goods | ❌ |
| 116 | Pig | Farm Goods | ❌ |
| 117 | Milk | Farm Goods | ❌ |
| 120 | Vegetables | Staples | ❌ |
| 121 | Bread | Kitchen Chain | ✅ `item_bread_v1.png` |
| 122 | Cheese | Kitchen Chain | ❌ |
| 127 | Pizza/Meal | Kitchen Chain | ✅ `item_meal_v1.png` |
| 133 | Flour | Kitchen Chain | ✅ `item_flour_v1.png` |
| 134 | Butter | Kitchen Chain | ❌ |
| 135 | Sugar | Kitchen Chain | ❌ |
| 137 | Dough | Kitchen Chain | ❌ |
| 139 | Fodder | Kitchen Chain | ❌ |
| 141 | Vegetable Oil | Kitchen Chain | ❌ |

**后端** 的 `bot_resources` 配置 (`game.json` line 25) 包含完全相同的一组 ID，且通过 `resources.json` 加载完整定义（含配方配方输入输出）。

**只有 4/23 的资源有前端物品图:** Wheat, Flour, Bread, Meal.

还有一个 `resource_items_chain_v1_sheet.png` 精灵图集在 `assets/game/items/`。

### 建筑 (Buildings)

| Kind | 名称 | 前端纹理 | 资产 |
|------|------|---------|------|
| 1 | Grain Plot | `grain_plot_lv1_idle_trimmed.png` | ✅ 4 建筑均有去背 PNG |
| 2 | Mill House | `mill_house_lv1_idle_trimmed.png` | ✅ |
| 3 | Bakery Shop | `bakery_shop_lv1_idle_trimmed.png` | ✅ |
| 4 | Meal Kiosk | `meal_kiosk_lv1_idle_trimmed.png` | ✅ |

后端通过 `buildings.json` 加载完整建筑定义，`building_shop.go` 中 `BuildingMarket()` 返回 4 种建筑。

### 前端 Icon 清单 (16 个)

全部在 `assets/game/icons/` 有对应 PNG, 后端无相关概念。

| Icon | 用途 |
|------|------|
| `icon_level_badge_v1` | 等级徽章 |
| `icon_builder_v1` | 建造 |
| `icon_warehouse_v1` | 仓库 |
| `icon_market_v1` | 市场 |
| `icon_contract_v1` | 合同 |
| `icon_timer_v1` | 研发 (当前用作 research 图标) |
| `icon_cash_v1` | 现金 |
| `icon_coin_v1` | 金币 |
| `icon_energy_v1` | 能量 |
| `icon_factory_v1` | 工厂 |
| `icon_wheat_resource_v1` | 小麦 |
| `icon_xp_v1` | 经验 |
| `icon_restaurant_v1` | 餐厅 |
| `icon_upgrade_v1` | 升级 |
| `icon_collect_v1` | 收集 |
| `icon_refresh_v1` | 刷新 |

### 地图背景

`map_background_v1.png` — 单张地图背景, 前后端共用 (前端 PixiJS 加载渲染)。

---

## 总结

### 前端已对接成熟的功能
- **认证**: 登录/注册 ✅
- **市场**: 行情、深度、挂单、成交、资源列表 ✅
- **生产**: 查看任务、开始/收取/取消生产、生产队列 ✅
- **建筑**: 购买、放置、查看详情 ✅
- **日常+政府合同**: 查看、竞标、完成、领取奖励 ✅
- **地图渲染**: PixiJS 地图 + 建筑精灵 ✅

### 前端完全缺失的功能 (后端已就绪)
- **财务**: 三大报表 (损益表/资产负债表/现金流量表)
- **债券**: 发行、购买、赎回、利息结算
- **高管**: 搜索、招募、培训、挖角
- **拍卖**: 建筑竞标
- **航空**: 火箭项目
- **研发**: 研究项目 (侧边栏有入口但无页面)
- **SimBoost**: 加速道具使用
- **成就/等级奖励**: 成就系统
- **聊天对接**: 后端有完整 REST API 但前端用 mock

### 资产缺口
- 23 种资源中仅 4 种有前端物品图标
- 后端配方系统完整 (resources.json 中 `producedFrom` 定义了所有配方链) 但前端未直接使用 `/api/v2/recipes` 端点
