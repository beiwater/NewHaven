# Phase 2 — 核心业务域迁移

目标：在 Phase 1 的前端基础设施通过验收后，把 Market、Warehouse、Production、Buildings 四个核心业务域迁移到 `clinet-next/`，并统一接入 API hooks、query keys、React Hook Form + Zod、Recharts、MSW 和 shadcn/ui。

目标代码路径：`clinet-next/`（目录名沿用现状拼写，非 typo）。

源项目路径：`client/atlas-foods-client/`。

---

## Phase 2 入口条件

必须先满足 Phase 1：

- `clinet-next/` 已经是可构建的 Vite + React + TS 项目。
- `npm run build` 通过。
- `npm run lint` 通过。
- `npm test` 通过。
- `api/client.ts`、`api/errors.ts`、`api/queryKeys.ts` 已存在。
- `components/ui/` 基础组件已存在且可导入。
- `mocks/` 与 Vitest 测试框架已存在。

如果 Phase 1 未通过，不进入 Phase 2。

---

## 范围

### 做

1. 迁移四个核心业务域：
   - Market
   - Warehouse / Inventory
   - Production
   - Buildings
2. 把现有散落 query key 迁移到 `api/queryKeys.ts`。
3. 把业务 API hooks 放入统一目录，建议为 `src/api/hooks/`。
4. Market 下单表单改成 React Hook Form + Zod。
5. Market 价格图从手写 SVG 迁移到 Recharts。
6. 为四个业务域补 MSW handler。
7. 为 hooks、表单 schema、关键组件补最小测试。
8. 保留现有业务行为和后端 API 路径，不改后端。

### 不做

- 不迁移 Research、Financial、Executive、Chat、Power-up。
- 不重做完整 App shell。
- 不重做 GameCanvas / PixiJS 地图。
- 不迁移音频系统。
- 不引入 Redux、MUI、Ant Design、Three.js、完整开源游戏模板。
- 不把服务器数据放进 Zustand。
- 不一次性接入 OpenAPI 全量生成类型。

---

## 当前源项目对照

| 业务域 | 源 API 文件 | 源 Feature 文件 | Phase 2 目标 |
| --- | --- | --- | --- |
| Market | `src/api/market.api.ts` | `features/market/MarketPage.tsx`, `MarketTicker.tsx`, `PriceCurve.tsx`, `ParticipantList.tsx`, `constants.ts` | 统一 hooks、RHF+Zod 下单、Recharts 价格图 |
| Warehouse | `src/api/inventory.api.ts` | `features/inventory/InventoryBar.tsx` | 统一 hooks、库存条/仓库状态组件 |
| Production | `src/api/production.api.ts` | `features/production/ProductionQueue.tsx` | 统一 hooks、生产队列、领取/取消 mutation |
| Buildings | `src/api/buildings.api.ts` | `features/buildings/BuildView.tsx`, `BuildingPanel.tsx`, `BuildingCard.tsx`, `MapPicker.tsx`, `building.store.ts` | 统一 hooks、建造/放置/升级/移动/拆除 mutation |

---

## 目标 API hooks 结构

建议在 Phase 2 建立以下文件：

```txt
clinet-next/src/api/hooks/
  market.hooks.ts
  warehouse.hooks.ts
  production.hooks.ts
  buildings.hooks.ts
```

保留 `src/api/client.ts` 只做请求边界，不放业务 hooks。

### queryKeys 约定

`src/api/queryKeys.ts` 至少覆盖：

```ts
export const queryKeys = {
  company: () => ['company'] as const,
  buildings: {
    all: () => ['buildings'] as const,
    byCompany: (companyId: number | null) => ['buildings', 'company', companyId] as const,
    productionOptions: (buildingId: string | undefined) => ['buildings', 'production-options', buildingId] as const,
  },
  warehouse: () => ['warehouse'] as const,
  production: {
    jobs: () => ['production', 'jobs'] as const,
    queue: () => ['production', 'queue'] as const,
    claimable: () => ['production', 'claimable'] as const,
  },
  market: {
    resources: () => ['market', 'resources'] as const,
    ticker: (resourceId: number) => ['market', 'ticker', resourceId] as const,
    depth: (resourceId: number, quality: number) => ['market', 'depth', resourceId, quality] as const,
    orders: (resourceId: number, quality: number) => ['market', 'orders', resourceId, quality] as const,
  },
}
```

允许根据现有实现微调，但禁止继续散写裸字符串 key。

---

## 必须保留的后端 API 路径

### Market

```txt
GET    /api/v3/resources/
GET    /api/v3/market-ticker/{resourceId}/
GET    /api/v3/market-depth/{resourceId}/{quality}/
GET    /api/v3/market/{resourceId}/{quality}/
POST   /api/v2/market-order/
DELETE /api/v2/market-order/cancel/{orderId}/
POST   /api/v2/market-order/take/
```

### Warehouse

```txt
GET /api/v2/companies/me/warehouse/
```

### Production

```txt
GET  /api/v2/production/jobs/
GET  /api/v2/production/queue/
GET  /api/v2/production/claimable/
GET  /api/v2/buildings/{buildingId}/production-options/
POST /api/v1/buildings/{buildingId}/busy/
POST /api/v2/production/claim/{jobId}/
POST /api/v2/production/claim-all/
POST /api/v2/production/cancel/
```

### Buildings

```txt
GET  /api/v2/companies/me/buildings/
GET  /api/v2/companies/{companyId}/buildings/
POST /api/v2/buildings/buy/
POST /api/v2/buildings/place/
POST /api/v2/buildings/move/
POST /api/v1/buildings/{buildingId}/upgrade/
POST /api/v2/buildings/demolish/
```

---

## 共享迁移边界

四个业务域会依赖一小组共享类型和资源 helper。只迁移必要文件，不搬整个旧前端。

### 允许迁移

```txt
src/game/types.ts
src/game/resources.ts
src/game/icons.ts
src/game/map.config.ts
src/features/ui/Icon.tsx
src/constants.ts
```

### 暂不迁移

```txt
src/game/GameCanvas.tsx
src/game/pixi/
src/audio/
src/features/story/
src/features/mobile/
src/i18n/ 全量语言系统
```

说明：现有 Market 和 Building 组件有 `audio.playSfx()` 调用。Phase 2 不迁移音频系统，API hooks 不允许 import `audio/AudioManager`。若需要保留点击反馈，先在 UI 层留 `TODO(audio)`，或者通过可选回调处理，不能让 API hook 依赖音频模块。

---

## 子 Agent 分工

### 第 0 步：迁移审计（串行）

**Agent: Phase2-Audit**

- 读取 Phase 1 产物：
  - `clinet-next/src/api/client.ts`
  - `clinet-next/src/api/queryKeys.ts`
  - `clinet-next/src/components/ui/`
  - `clinet-next/src/mocks/`
  - `clinet-next/package.json`
- 读取源项目：
  - `client/atlas-foods-client/src/api/market.api.ts`
  - `client/atlas-foods-client/src/api/inventory.api.ts`
  - `client/atlas-foods-client/src/api/production.api.ts`
  - `client/atlas-foods-client/src/api/buildings.api.ts`
  - 四个对应 feature 目录
- 输出：
  - 需要复制的共享文件清单。
  - 需要改写的 import 清单。
  - 需要 mock 的 API 响应 shape。
  - 可能阻塞 build 的依赖清单。

验收：

- 明确哪些文件从旧 client 复制，哪些文件重新写。
- 明确是否需要临时 i18n adapter。若不迁移 i18n，页面文字先用固定英文/中文，不接 `useTranslation()`。

---

### 第 1 步：共享类型与资源迁移（串行）

**Agent: Shared-Slice**

操作文件：

```txt
clinet-next/src/game/types.ts
clinet-next/src/game/resources.ts
clinet-next/src/game/icons.ts
clinet-next/src/game/map.config.ts
clinet-next/src/features/ui/Icon.tsx
clinet-next/src/constants.ts
```

任务：

- 从源项目复制必要类型和 helper。
- 删除会拖入 PixiJS 地图、音频、story、mobile 的 import。
- 确保资源图片路径仍指向 `public/assets/...`。
- 如果图片资产未在 `clinet-next/public/assets/` 中，复制四个业务域需要的最小资产：
  - resource icons
  - building icons
  - building preview images

验收：

- `npm run build` 不因共享类型缺失失败。
- 共享文件不 import `GameCanvas`、`AudioManager`、`StoryPlayer`。

---

### 第 2 步：统一 API hooks（串行）

**Agent: API-Hooks**

操作文件：

```txt
clinet-next/src/api/hooks/market.hooks.ts
clinet-next/src/api/hooks/warehouse.hooks.ts
clinet-next/src/api/hooks/production.hooks.ts
clinet-next/src/api/hooks/buildings.hooks.ts
clinet-next/src/api/queryKeys.ts
```

任务：

- 从旧 `*.api.ts` 迁移 hooks 到新目录。
- 所有 query key 使用 `queryKeys`。
- mutation 成功后统一 invalidate 相关 query。
- API hooks 不 import React UI、Zustand store、音频、PixiJS。
- 保持后端 API 路径不变。

验收：

- 四个 hooks 文件可被独立 import。
- TypeScript 编译通过。
- `queryKeys.ts` 覆盖四个业务域。
- hooks 测试可通过 MSW 返回数据。

---

### 第 3 步：MSW handler 补齐（可与第 2 步后并行）

**Agent: MSW-Domain-Handlers**

操作文件：

```txt
clinet-next/src/mocks/handlers/market.ts
clinet-next/src/mocks/handlers/warehouse.ts
clinet-next/src/mocks/handlers/production.ts
clinet-next/src/mocks/handlers/buildings.ts
clinet-next/src/mocks/handlers/index.ts
```

任务：

- 补齐四个业务域的 happy path handler。
- 每个 handler 返回字段贴近现有前端类型，不乱造不使用的深层字段。
- 至少覆盖：
  - Market ticker/depth/orders/resources/create/cancel/take
  - Warehouse inventory/capacity/used
  - Production jobs/queue/claimable/options/start/claim/cancel
  - Buildings list/buy/place/move/upgrade/demolish

验收：

- `npm test` 可启动 MSW server。
- hooks 测试不需要真实后端。

---

### 第 4 步：业务域组件迁移（分批）

第 4 步不要五个组件同时乱改。建议按依赖从轻到重：

```txt
4A Warehouse
4B Market
4C Production
4D Buildings
```

#### Agent: Warehouse-Migration

操作文件：

```txt
clinet-next/src/features/warehouse/
clinet-next/src/features/inventory/
```

任务：

- 迁移 `InventoryBar.tsx`。
- 如果需要页面容器，新建 `WarehousePage.tsx`。
- 使用 `useWarehouse()` from `api/hooks/warehouse.hooks.ts`。
- 显示 loading / error / empty / normal 状态。

验收：

- 可渲染库存、容量、已用量。
- 不把库存复制进 Zustand。

#### Agent: Market-Migration

操作文件：

```txt
clinet-next/src/features/market/
```

任务：

- 迁移 `MarketPage.tsx`、`MarketTicker.tsx`、`ParticipantList.tsx`、`constants.ts`。
- 将 `PriceCurve.tsx` 改成 Recharts 实现，建议命名 `PriceHistoryChart.tsx`。
- 将下单区域拆成 `CreateOrderForm.tsx`。
- `CreateOrderForm` 使用 React Hook Form + Zod。
- 使用 `api/hooks/market.hooks.ts`。
- 删除 API hook 内音频副作用；UI 层可先留 TODO。

表单 schema 建议：

```ts
const createOrderSchema = z.object({
  resourceId: z.number().int().positive(),
  kind: z.enum(['buy', 'sell']),
  quality: z.number().int().min(0),
  quantity: z.coerce.number().int().positive(),
  price: z.coerce.number().positive(),
})
```

验收：

- 能查看资源列表、价格曲线、买卖盘、我的订单。
- 能提交买单/卖单 mutation。
- 本地库存/现金不足校验保留。
- Recharts 图表在空数据和正常数据下都不报错。

#### Agent: Production-Migration

操作文件：

```txt
clinet-next/src/features/production/
```

任务：

- 迁移 `ProductionQueue.tsx`。
- 使用 `api/hooks/production.hooks.ts`。
- 保留领取、全部领取、取消生产操作。
- 保留倒计时和 ready 状态展示。

验收：

- jobs、queue、claimable 三组数据各自 loading/error/empty 状态完整。
- claim/cancel mutation 后 invalidate production + warehouse + company 相关 query。

#### Agent: Buildings-Migration

操作文件：

```txt
clinet-next/src/features/buildings/
clinet-next/src/store/ui.store.ts
clinet-next/src/store/game.store.ts
```

任务：

- 迁移 `BuildView.tsx`、`BuildingPanel.tsx`、`BuildingCard.tsx`、`MapPicker.tsx`、`building.store.ts`。
- 使用 `api/hooks/buildings.hooks.ts` 和 `api/hooks/production.hooks.ts`。
- 保留建造、放置、升级、移动、拆除。
- 不迁移完整 PixiJS 地图；MapPicker 可以先保留简化版或静态槽位版。
- 若需要地图状态，只使用 Zustand UI 状态，不存服务器建筑列表。

验收：

- 可展示建筑列表和建筑详情。
- 可触发 buy/place/move/upgrade/demolish mutation。
- 删除确认、移动状态、选中状态可用。
- 不 import `GameCanvas` 或 PixiJS internals。

---

### 第 5 步：App shell 接入（串行）

**Agent: App-Shell-Wire**

操作文件：

```txt
clinet-next/src/App.tsx
clinet-next/src/app/providers.tsx
clinet-next/src/store/ui.store.ts
```

任务：

- 建立最小可用导航：
  - Market
  - Warehouse
  - Production
  - Buildings
- 使用 Zustand 保存 `activeView`。
- 使用 Phase 1 的 QueryProvider 包裹 App。
- 不接完整旧版布局，不迁移 Story/Mobile/Audio。

验收：

- 四个页面能通过导航切换。
- 刷新后默认进入 Market 或 Buildings。
- build 通过。

---

### 第 6 步：测试与验证（串行）

**Agent: Verify-Phase2**

测试要求：

```txt
hooks:
  market.hooks.test.ts
  warehouse.hooks.test.ts
  production.hooks.test.ts
  buildings.hooks.test.ts

forms:
  CreateOrderForm.schema.test.ts

components:
  MarketPage smoke test
  Warehouse/Inventory smoke test
  ProductionQueue smoke test
  BuildingCard smoke test
```

命令：

```bash
cd clinet-next
npm run build
npm run lint
npm test
```

额外核对：

- `package.json` 不含 Redux、MUI、Ant Design、Three.js。
- `src/api/hooks/*` 不 import UI、Zustand、Audio、PixiJS。
- `src/store/*` 不保存 warehouse、orders、jobs、company money 等服务器数据。
- `src/features/market/*` 使用 Recharts，不保留手写 SVG 价格图作为主图表。

---

## 目标目录结构

```txt
clinet-next/src/
  api/
    hooks/
      market.hooks.ts
      warehouse.hooks.ts
      production.hooks.ts
      buildings.hooks.ts
    queryKeys.ts

  features/
    market/
      MarketPage.tsx
      MarketTicker.tsx
      ParticipantList.tsx
      PriceHistoryChart.tsx
      CreateOrderForm.tsx
      createOrder.schema.ts
      constants.ts
    warehouse/
      WarehousePage.tsx
    inventory/
      InventoryBar.tsx
    production/
      ProductionQueue.tsx
    buildings/
      BuildView.tsx
      BuildingPanel.tsx
      BuildingCard.tsx
      MapPicker.tsx
      building.store.ts
    ui/
      Icon.tsx

  game/
    types.ts
    resources.ts
    icons.ts
    map.config.ts

  mocks/
    handlers/
      market.ts
      warehouse.ts
      production.ts
      buildings.ts

  tests/
    unit/
      hooks/
      forms/
      components/
```

---

## 验收标准

- `npm run build` 通过。
- `npm run lint` 通过。
- `npm test` 通过。
- 四个业务域页面可在 `clinet-next` 中打开。
- Market 下单使用 RHF + Zod。
- Market 价格图使用 Recharts。
- 四个业务域 hooks 全部使用 `queryKeys`。
- 四个业务域有 MSW happy path handler。
- API hooks 不 import UI、Zustand、Audio、PixiJS。
- Zustand 不保存服务器数据。
- 没有迁移 Research、Financial、Executive、Chat、Power-up。

---

## Phase 3 入口条件

Phase 2 通过后，进入 Phase 3：

- Research 迁移到统一 hooks + 表单 mutation。
- Financial 迁移到 Recharts + 表格组件。
- Executive 迁移到 RHF + Zod + dialog 表单。
- Chat 迁移到 API hooks + MSW handler。
- Power-up UI 接入统一 hooks。

