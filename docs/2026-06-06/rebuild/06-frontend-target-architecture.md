# 前端目标架构

> Phase 1 重建 —— 前端架构文档
> 基于当前代码审计结果（2026-06-05），描述目标状态、目录结构、数据流边界和集成计划。

## 1. 总体原则

- **无重复数据源**：服务端数据只存在于 TanStack Query 缓存，Zustand 只持 UI 状态。
- **类型安全优先**：所有 API 响应类型通过 `openapi-typescript` 从 OpenAPI 规范生成，手写类型逐步淘汰。
- **hooks 聚合**：每个业务域通过一个自定义 hook 文件暴露所有 Query + Mutation，组件不直接调用 `api.get()`。
- **WebSocket 为实时通道**：市场和生产数据优先走 WS 推送，HTTP 查询作为降级和初始加载。
- **Canvas ↔ React 解耦**：PixiJS GameCanvas 懒加载，通过 Zustand + 回调与 React UI 通信。

## 2. 目录结构（目标）

```
src/api/
  generated/                # openapi-typescript 自动生成 ← 新增
    index.ts                # 所有后端类型定义
    schemas.ts              # 请求/响应 schema
  client.ts                 # 自定义 fetch 封装（已有，保留）
  queryKeys.ts              # 集中管理的 query key 工厂 ← 新增
  websocket.ts              # WebSocket 连接管理器（已有桩，需实现）
  hooks/                    # TanStack Query hooks 按域分文件 ← 新增目录
    useCompany.ts
    useBuildings.ts
    useProduction.ts
    useMarket.ts
    useResearch.ts
    useFinancial.ts
    useExecutives.ts
    useChat.ts
    useContracts.ts
    useLeaderboard.ts
    usePowerup.ts
    useInventory.ts
    useWebSocket.ts          # WS 订阅 hook（useMarketWebSocket / useProductionWebSocket）
  # 以下文件在 Phase 1 生命周期内保留，待 hooks/ 完善后逐步废弃：
  company.api.ts
  buildings.api.ts
  production.api.ts
  market.api.ts
  research.api.ts
  financial.api.ts
  executives.api.ts
  chat.api.ts
  contracts.api.ts
  leaderboard.api.ts
  powerup.api.ts
  inventory.api.ts
```

### 2.1 `generated/` 目录

`openapi-typescript` 从后端 OpenAPI 规范生成类型定义。以 "types from backend" 为唯一来源。

- 生成命令：`npx openapi-typescript https://<backend>/openapi.json -o src/api/generated/index.ts`
- 生成时机：CI 构建时，或本地 `npm run codegen`
- 所有 `hooks/*.ts` 中的 queryFn 返回类型直接从 `generated/` 引入，不再手写响应接口
- 现有 `api/*.api.ts` 中的手写类型（如 `ChatMessage`、`ResearchProject`、`IncomeStatement`）在 hooks 迁移到位后删除

### 2.2 `queryKeys.ts`

集中管理所有 query key，避免字符串散落在各文件：

```typescript
// src/api/queryKeys.ts
export const queryKeys = {
  company: {
    all: ['company'] as const,
    level: ['company', 'level'] as const,
  },
  buildings: {
    all: ['buildings'] as const,
    byCompany: (id: number) => ['buildings', 'company', id] as const,
  },
  production: {
    jobs: ['production', 'jobs'] as const,
    queue: ['production', 'queue'] as const,
    claimable: ['production', 'claimable'] as const,
    options: (buildingId: string) => ['production', 'options', buildingId] as const,
  },
  market: {
    ticker: (resourceId: number) => ['market', 'ticker', resourceId] as const,
    depth: (resourceId: number, quality: number) => ['market', 'depth', resourceId, quality] as const,
    orders: (resourceId: number, quality: number) => ['market', 'orders', resourceId, quality] as const,
  },
  research: {
    projects: ['research', 'projects'] as const,
    progress: ['research', 'projects', 'progress'] as const,
  },
  financial: {
    income: ['financial', 'income'] as const,
    balance: ['financial', 'balance'] as const,
    cashflow: ['financial', 'cashflow'] as const,
    past: ['financial', 'past'] as const,
    overview: ['financial', 'overview'] as const,
  },
  executives: {
    all: ['executives'] as const,
    my: ['executives', 'my'] as const,
    detail: (id: string) => ['executives', 'detail', id] as const,
  },
  chat: {
    messages: ['chat', 'messages'] as const,
    contacts: ['chat', 'contacts'] as const,
    chatroom: ['chat', 'chatroom'] as const,
  },
  contracts: {
    dailyOrders: ['contracts', 'dailyOrders'] as const,
    govContracts: ['contracts', 'govContracts'] as const,
  },
  leaderboard: {
    list: (sort: string, page: number) => ['leaderboard', sort, page] as const,
  },
  powerup: {
    types: ['powerup', 'types'] as const,
    active: ['powerup', 'active'] as const,
  },
  warehouse: {
    all: ['warehouse'] as const,
  },
}
```

### 2.3 hooks 分域细节

每个 `hooks/useXxx.ts` 文件遵循统一模式：

```typescript
// src/api/hooks/useMarket.ts（示例）
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '@/api/client'
import { queryKeys } from '@/api/queryKeys'
import type { components } from '@/api/generated'  // openapi-typescript 生成

export function useMarketTicker(resourceId: number) {
  return useQuery({
    queryKey: queryKeys.market.ticker(resourceId),
    queryFn: () => api.get<components['schemas']['MarketTicker']>(`/api/v3/market-ticker/${resourceId}/`),
    refetchInterval: 10_000,       // 市场数据每 10s 刷新
    staleTime: 5_000,
  })
}

export function useMarketDepth(resourceId: number, quality = 0) {
  return useQuery({
    queryKey: queryKeys.market.depth(resourceId, quality),
    queryFn: () => api.get<components['schemas']['MarketDepth']>(`/api/v3/market-depth/${resourceId}/${quality}/`),
    refetchInterval: 10_000,
    staleTime: 5_000,
  })
}

export function useCreateOrder() {
  const qc = useQueryClient()
  return useMutation({
    mutationFn: (order: components['schemas']['CreateOrderPayload']) =>
      api.post('/api/v3/market/order/', order),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.market.orders(0, 0) }) // 更精确的 key 需动态
    },
  })
}
```

### 2.4 各领域 staleTime / refetchInterval 策略

| 域 | staleTime | refetchInterval | 说明 |
|---|---|---|---|
| Company | 5 min | 否 | 公司信息变化慢 |
| Buildings | 30 s | 否 | 建筑状态可通过 WS 推送 |
| Production Jobs | 5 s | 5 s（有数据时） | 生产进度需要接近实时 |
| Production Queue | 10 s | 10 s | 队列状态 |
| Market Ticker | 5 s | 10 s | 价格变化快 |
| Market Depth | 5 s | 10 s | 深度数据 |
| Research | 1 min | 15 s（进行中） | 研究进度可轮询 |
| Financial | 2 min | 否 | 财务报表变化慢 |
| Executives | 1 min | 否 | 高管数据 |
| Chat Messages | 0 | 15 s（有数据时） | 聊天需及时 |
| Chat Contacts | 1 min | 否 | 联系人变化慢 |
| Contracts | 1 min | 否 | 订单列表 |
| Leaderboard | 30 s | 60 s | 排行榜 |
| Powerup | 1 min | 30 s（生效中） | 加速状态 |
| Warehouse | 30 s | 否 | 库存变化不频繁 |

## 3. Zustand 边界

### 3.1 只有 UI 状态

Zustand 的三个 store 都只保存与后端无关的 UI 状态。

**`ui.store.ts`（已有，保留）**

```typescript
interface UIState {
  activeView: ActiveView            // 当前视图（map/build/market/...）
  selectedBuildingId: string | null // React UI 当前选中的建筑
  placementBuildingId: string | null
  movingBuildingId: string | null
  sidebarOpen: boolean
  marketPanelOpen: boolean
  chatOpen: boolean
  powerupOpen: boolean
  currentMapId: MapId
}
```

**`game.store.ts`（已有，保留）**

```typescript
interface GameState {
  zoom: number                      // PixiJS 画布缩放
  panX: number
  panY: number
  selectedMapBuildingId: string | null  // PixiJS 画布上选中的建筑（可能与 ui.selectedBuildingId 同步）
  tick: number                      // 生产动画刷新 tick
}
```

> `selectedMapBuildingId` 和 `ui.store.selectedBuildingId` 是两个不同来源的选择状态。前者由 Canvas 内部鼠标事件驱动，后者由 React 侧（侧边栏、面板）驱动。当一方变化时，通过 `GameCanvas` 中的 `useEffect` 同步到另一方。

**`building.store.ts`（已有，需重构合并到 `ui.store.ts`）**

当前 `building.store.ts` 只包含 `queuePanelOpen`，这是一个 UI 开关状态，不应独立成一个 store。Phase 1 将其字段并入 `ui.store.ts`，删除 `building.store.ts`。

### 3.2 禁止存入 Zustand 的数据

以下数据必须保持在 TanStack Query 缓存中，禁止放入 Zustand：

- 建筑列表、建筑详情
- 生产任务、队列
- 市场订单、深度、行情
- 研究项目列表和进度
- 财务报表
- 高管列表和详情
- 聊天消息
- 仓库库存
- 排行榜数据
- 公司信息、玩家等级

### 3.3 Zustand 升级为 TypeScript 严格模式

```typescript
import { create } from 'zustand'
import { useShallow } from 'zustand/react/shallow'

// 组件中避免全量订阅
const activeView = useUIStore((s) => s.activeView)
// 或
const { activeView, chatOpen } = useUIStore(useShallow((s) => ({ activeView: s.activeView, chatOpen: s.chatOpen })))
```

## 4. Mock 数据清理

### 4.1 当前 Mock 现状

经审计，所有主要页面已使用真实 API 钩子，不存在内联 mock 数据。具体情况：

| 页面 | 状态 | 说明 |
|---|---|---|
| ChatPanel | ✅ 已接入 | 使用 `useMessages` / `useSendMessage` / `useMarkRead` / `useContacts`（来自 `chat.api.ts`） |
| ResearchPage | ✅ 已接入 | 使用 `useResearch` / `useResearchProgress` / `useStartResearch` / `useCompleteResearch` |
| FinancialPage | ✅ 已接入 | 使用 `useFinancialOverview` / `useIncomeStatement` / `useBalanceSheet` / `useCashflowStatement` |
| ExecutivePage | ✅ 已接入 | 使用 `useMyExecutives` / `useTrainExecutive`，子组件 `ExecutiveMarket` 使用 `useExecutiveSearch` |

### 4.2 尚存的桩和需要替换的部分

| 文件 | 问题 | 处理方式 |
|---|---|---|
| `api/websocket.ts` | 三个导出函数均为空实现 | 实现 WebSocket 连接管理器（见第 7 节） |
| `api/chat.api.ts` | `useMessages` 使用 GET `/api/messages/`（v1 路径），`useSendMessage` 使用 POST `/api/v2/message/`，`useMarkRead` 使用 GET（非 POST） | 统一路径风格为 `/api/v2/...`，POST 替代 GET 修改操作 |
| `api/financial.api.ts` | 类型为手写 | 接入 `generated/` 类型后删除手写定义 |
| `api/research.api.ts` | 类型为手写 | 同上 |
| `api/executives.api.ts` | 类型为手写，`transformExecutive` 格式转换逻辑 | 后端稳定后移除转换函数 |
| `features/buildings/building.store.ts` | 独立小 store | 合并到 `ui.store.ts` |
| `features/research/ResearchPage.tsx` | `inferCategory` 硬编码推断研究类别 | 后端返回 `category` 字段后删除该函数 |

### 4.3 清理 checklist

- [ ] `api/websocket.ts` 实现 WS 连接管理器
- [ ] `api/chat.api.ts` API 路径规范化
- [ ] `api/*.api.ts` 手写类型替换为 `generated/` 导入
- [ ] 删除 `building.store.ts`，字段迁入 `ui.store.ts`
- [ ] 删除 `api/executives.api.ts` 中的 `transformExecutive` 等过时转换

## 5. 新页面集成顺序

Research、Financial、Executives、Chat 四个页面的集成按以下优先级：

### Phase 1.1 — Research（已完成基线）

- 状态：已接入真实 API，页面完整
- 目标：替换手写类型为 `generated/`，确认 `refetchInterval` 逻辑正确

### Phase 1.2 — Financial（已完成基线）

- 状态：已接入真实 API，页面完整
- 目标：添加错误状态 UI（目前无 `isError` 分支），接入 WS 推送（如未来有财务变动事件）

### Phase 1.3 — Executives（已完成基线）

- 状态：已接入真实 API，页面完整
- 目标：替换 `transformExecutive` 格式转换，接入 `generated/` 类型

### Phase 1.4 — Chat（已完成基线，但需优化）

- 状态：已接入真实 API，但后端 API 路径不一致
- 目标：
  - 规范化 API 路径
  - 接入 WebSocket 实现实时消息推送（代替 15s 轮询）
  - 添加 `useMarkRead` 的成功回调更新本地缓存（乐观更新）

### 集成顺序总表

```
Week 1: Research 类型替换 + 错误边界  |  Financial 错误态补充
Week 2: Executives 转换函数移除      |  Chat API 路径规范化
Week 3: WebSocket 集成（市场 + 生产 + 聊天）
Week 4: 全页面类型迁移到 generated/  |  删除废弃文件
```

## 6. PixiJS Canvas ↔ React UI 边界

### 6.1 当前架构

```
                    ┌──────────────────────┐
                    │   App.tsx             │
                    │  (路由 / 视图切换)     │
                    └───────┬──────────────┘
                            │ lazy()
                    ┌───────▼──────────────┐
                    │  GameCanvas.tsx        │
                    │  (PixiJS Application) │
                    │                       │
                    │  - 读取 Zustand:      │
                    │    selectedBuildingId  │
                    │    placementBuildingId │
                    │    movingBuildingId    │
                    │    currentMapId        │
                    │  - 调用 API hooks:    │
                    │    useBuildings()      │
                    │    useProductionJobs() │
                    │  - 回调 setActiveView  │
                    │    selectBuilding      │
                    └──────────────────────┘
                    ▲
                    │ (DOM ref div)
                    │
               React 组件树（TopBar / Sidebar / Panels...）
```

### 6.2 通信方式

**React → PixiJS（数据流入 Canvas）**

通过 Zustand store 传递：
- `ui.store.selectedBuildingId` → Canvas 高亮对应建筑精灵
- `ui.store.placementBuildingId` → Canvas 显示放置预览
- `ui.store.currentMapId` → Canvas 切换地图
- `game.store.zoom / panX / panY` → Canvas 视口变换（由 Canvas 内的滚轮/拖拽事件写入）

**PixiJS → React（事件流出 Canvas）**

通过直接调用 Zustand actions：
- Canvas 内点击建筑 → `useUIStore.getState().selectBuilding(id)` → React 侧 Sidebar/BuildingPanel 更新
- Canvas 内切换地图 → `useUIStore.getState().setCurrentMapId(newMapId)`
- Canvas 放置完成 → `useUIStore.getState().clearBuildingPlacement()`

**PixiJS 内部状态**

保持在 `useRef` 中，不放入 Zustand：
- `appRef` — PixiJS Application 实例
- `buildingLayerRef` — 建筑图层 Container
- 纹理缓存、缩放因子等

### 6.3 懒加载策略

```typescript
// App.tsx（已有，保留）
const GameCanvas = lazy(() => import('@/game/GameCanvas'))

// 只在 activeView === 'map' || activeView === 'build' 时加载
// Suspense fallback 显示加载骨架
```

### 6.4 性能隔离

- GameCanvas 使用 `React.memo` 包裹，仅在 `currentMapId`、`selectedBuildingId`、`placementBuildingId`、`movingBuildingId`、`tick` 变化时重渲染
- PixiJS 内部动画循环通过 `tick` 驱动（`game.store.tick`），不由 React render 驱动
- 纹理预加载在 `GameCanvas` 挂载时执行一次

## 7. WebSocket 集成计划

### 7.1 `api/websocket.ts` — 连接管理器

目标实现结构：

```typescript
// src/api/websocket.ts

type WsMessageHandler = (data: unknown) => void

interface WsOptions {
  url: string
  reconnectInterval?: number    // 默认 3000
  maxReconnectAttempts?: number // 默认 10
}

class WebSocketManager {
  private ws: WebSocket | null = null
  private handlers = new Map<string, Set<WsMessageHandler>>()
  private reconnectAttempts = 0
  private options: WsOptions
  private destroyed = false

  constructor(options: WsOptions) { ... }

  connect(): void { ... }
  disconnect(): void { ... }
  subscribe(event: string, handler: WsMessageHandler): () => void { ... }
  send(type: string, payload?: Record<string, unknown>): void { ... }

  private onMessage(event: MessageEvent): void { ... }
  private reconnect(): void { ... }
}

// 单例
let manager: WebSocketManager | null = null

export function getWS(): WebSocketManager {
  if (!manager) {
    manager = new WebSocketManager({
      url: `${location.protocol === 'https:' ? 'wss' : 'ws'}://${location.host}/ws/`,
      reconnectInterval: 3000,
      maxReconnectAttempts: 10,
    })
  }
  return manager
}

export function useMarketWebSocket(handler: (data: MarketWsMessage) => void) {
  useEffect(() => {
    const ws = getWS()
    const unsub = ws.subscribe('market_update', handler)
    ws.connect()
    return unsub
  }, [handler])
}

export function useProductionWebSocket(handler: (data: ProductionWsMessage) => void) {
  useEffect(() => {
    const ws = getWS()
    const unsub = ws.subscribe('production_update', handler)
    ws.connect()
    return unsub
  }, [handler])
}

export function sendWS(type: string, payload?: Record<string, unknown>) {
  getWS().send(type, payload)
}
```

### 7.2 WebSocket 消息类型

```typescript
// 后续在 generated/ 或 types.ts 中定义

interface MarketWsMessage {
  type: 'market_update'
  resourceId: number
  price: number
  volume: number
  timestamp: string
}

interface ProductionWsMessage {
  type: 'production_update'
  buildingId: string
  jobId: string
  status: 'progress' | 'completed' | 'claimed'
  progress?: number
}

interface ChatWsMessage {
  type: 'chat_message'
  message: ChatMessage
}
```

### 7.3 WebSocket 与 TanStack Query 的关系

| 场景 | 策略 |
|---|---|
| 初始加载 | HTTP GET（现有 useQuery） |
| 实时更新 | WS 推送到达后，通过 `queryClient.setQueryData()` 直接更新缓存 |
| 连接断开 | WS 自动重连；断连期间使用 HTTP 轮询作为降级（`refetchInterval` 作为 fallback） |
| Mutation 后 | 仍然通过 `queryClient.invalidateQueries()` 触发刷新（乐观更新 + WS 推送双重保证） |

### 7.4 实现步骤

1. 实现 `WebSocketManager` 类（连接、重连、订阅分发）
2. 导出 `getWS()` 单例
3. 实现 `useMarketWebSocket` — 订阅 `market_update` 事件，更新 market 查询缓存
4. 实现 `useProductionWebSocket` — 订阅 `production_update` 事件，更新 production 查询缓存
5. 在 `App.tsx` 中调用 `useMarketWebSocket` 和 `useProductionWebSocket`（保持与现有代码一致）
6. 聊天 WS 订阅后续在 ChatPanel 内部通过 `useEffect` + `getWS().subscribe('chat_message', ...)` 接入

### 7.5 当前 stub 清理

当前 `api/websocket.ts` 内容：

```typescript
export function useMarketWebSocket() { /* Not implemented */ }
export function useProductionWebSocket() { /* Not implemented */ }
export function sendWS(type: string, payload?: Record<string, unknown>) { /* Not implemented */ }
```

Phase 1 中的实现将替换这些空函数，并保持与 `App.tsx` 中 `import { useMarketWebSocket, useProductionWebSocket } from '@/api/websocket'` 的兼容。

## 8. 迁移策略

### 8.1 文件迁移顺序

```
Step 1: 创建 src/api/generated/（openapi-typescript 生成）
Step 2: 创建 src/api/queryKeys.ts
Step 3: 实现 api/websocket.ts（连接管理器）
Step 4: 创建 src/api/hooks/ 目录，逐步从 *.api.ts 复制并改写为使用 generated/ 类型
Step 5: 合并 building.store.ts 到 ui.store.ts
Step 6: 组件逐步从 '@/api/xxx.api' 迁移到 '@/api/hooks/useXxx'
Step 7: 删除废弃的 *.api.ts 文件（保留 client.ts）
```

### 8.2 兼容期间

- 旧 `*.api.ts` 文件和新 `hooks/useXxx.ts` 文件同时存在
- 旧文件不再新增 hook，只维护 bug 修复
- 组件迁移到新 hook 后，删除其导入的旧模块
- 所有新组件直接使用 `hooks/useXxx.ts`

### 8.3 删除清单（Phase 1 结束时）

- [ ] `src/api/company.api.ts` → 存在 `hooks/useCompany.ts`
- [ ] `src/api/buildings.api.ts` → 存在 `hooks/useBuildings.ts`
- [ ] `src/api/production.api.ts` → 存在 `hooks/useProduction.ts`
- [ ] `src/api/market.api.ts` → 存在 `hooks/useMarket.ts`
- [ ] `src/api/research.api.ts` → 存在 `hooks/useResearch.ts`
- [ ] `src/api/financial.api.ts` → 存在 `hooks/useFinancial.ts`
- [ ] `src/api/executives.api.ts` → 存在 `hooks/useExecutives.ts`
- [ ] `src/api/chat.api.ts` → 存在 `hooks/useChat.ts`
- [ ] `src/api/contracts.api.ts` → 存在 `hooks/useContracts.ts`
- [ ] `src/api/leaderboard.api.ts` → 存在 `hooks/useLeaderboard.ts`
- [ ] `src/api/powerup.api.ts` → 存在 `hooks/usePowerup.ts`
- [ ] `src/api/inventory.api.ts` → 存在 `hooks/useInventory.ts`
- [ ] `src/features/buildings/building.store.ts` → 合并到 `ui.store.ts`
