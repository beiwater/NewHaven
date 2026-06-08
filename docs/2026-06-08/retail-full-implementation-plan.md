# 零售全链路实施方案

日期: 2026-06-08

## 概述

修复前后端实现完整的"市场买入 → 生产 → 零售上架 → NPC 卖出 → 结算"链条。

---

## 后端变更

### Phase 1: 数据模型

**1. `internal/domain/building/types.go`** — CatalogEntry 新增零售字段
```go
type CatalogEntry struct {
    // ... 现有
    Type        string `json:"type"` // "production" | "retail"
    RetailSlots int    `json:"retail_slots"`    // 基础货架数
    SlotPerLevel int   `json:"slot_per_level"`  // 每级新增货架数
}
```

**2. `internal/domain/company/types.go`** — Building 新增 Shelves
```go
type ShelfItem struct {
    ResourceID int     `json:"resource_id"`
    Quantity   int     `json:"quantity"`
    MaxQty     int     `json:"max_qty"`
    Price      float64 `json:"price"`
    PriceLock  bool    `json:"price_lock"`
    Revenue    float64 `json:"revenue,omitempty"` // 累计未提取收入
}

type Building struct {
    // ... 现有
    Shelves    []ShelfItem `json:"shelves,omitempty"`
}
```

**3. `internal/catalog/catalog.go`** — BuildingEntry 新增字段，load 时映射

**4. `decompiled/data/buildings.json`** — 5 个零售建筑补充字段
```json
{
  "6": { "retailSlots": 2, "slotPerLevel": 1 },
  "7": { "retailSlots": 3, "slotPerLevel": 1 },
  "8": { "retailSlots": 3, "slotPerLevel": 1 },
  "9": { "retailSlots": 4, "slotPerLevel": 1 },
  "12": { "retailSlots": 5, "slotPerLevel": 1 }
}
```

### Phase 2: Building 服务

**5. `internal/app/building/shop.go`** — BuyBuilding 初始化 `Shelves: []ShelfItem{}`

**6. `internal/app/building/service.go`** — 
- `BuildingMarket`: 在 `BuildingMarketItem` 中加 `IsRetail` 字段标记零售建筑
- `buildingToDTO`: 返回 `shelves` 信息

**7. `internal/app/building/retail.go`** (新建)
```go
func (s *Service) StockShelf(ctx, companyID, buildingID, resourceID, quantity int, price *float64) error
  // 校验建筑是零售类型
  // 查货架是否已满
  // 查仓库库存够不够
  // 减仓库库存，加货架数量
  // 如果传了 price 则锁定，否则跟市场价

func (s *Service) UnstockShelf(ctx, companyID, buildingID, resourceID, quantity int) error
  // 减货架，加回仓库

func (s *Service) SetShelfPrice(ctx, companyID, buildingID, resourceID int, price float64, lock bool) error
```
各方法加 `s.mu.Lock()` / `s.mu.Unlock()`。

### Phase 3: 零售结算改造

**8. `internal/app/market/retail.go`** — 核心改造

`ProcessRetailSales()` 改为：
```
for each NPC company:
    for each building where isRetail:
        for each shelfItem:
            → UnitsSoldPerHour(building.kind→modifier, shelfItem.price, ...)
            → 扣 shelf.Quantity
            → company.Money += earned
            → 记账
            → 如果 shelf 卖空, 从 shelves 移除
```

`CatchUpPlayerRetail()` 同理，按 elapsed 缩放。

**同时修复硬编码问题：**
- `storeSize` → 取 `building.Level`（之前硬编码 1）
- `salesModPct` → 从 executive 的 `SalesBonus` 汇总（之前硬编码 0.0）
- `saturation` → 保持 1.0（后续做维度追踪）

### Phase 4: HTTP API

**9. `internal/httpapi/building_handler.go`** — 新增 3 个 handler
```go
func (h *BuildingHandler) handleStockShelf(w, r)     // POST
func (h *BuildingHandler) handleUnstockShelf(w, r)   // POST
func (h *BuildingHandler) handleSetShelfPrice(w, r)  // POST
```

**10. `internal/httpapi/router.go`** — 注册新路由
```go
r.Post("/buildings/{buildingId}/stock/", h.Building.handleStockShelf)
r.Post("/buildings/{buildingId}/unstock/", h.Building.handleUnstockShelf)
r.Post("/buildings/{buildingId}/shelf-price/", h.Building.handleSetShelfPrice)
```

**11. `openapi/openapi-draft.yaml`** — 新增 Schema
- `ShelfItem` (resourceId, quantity, maxQty, price, priceLock, revenue)
- `StockShelfRequest`, `UnstockShelfRequest`, `SetShelfPriceRequest`
- `BuildingDTO` 新增 `is_retail`, `shelves` 字段

---

## 前端变更

### Phase 5: 类型 & API

**12. `src/game/types.ts`** — 新增 ShelfItem 接口，Building 加字段
```ts
export interface ShelfItem {
  resourceId: number
  quantity: number
  maxQty: number
  price: number
  priceLock: boolean
  revenue: number
}
// Building 新增:
export interface Building {
  // ...
  isRetail?: boolean
  shelves?: ShelfItem[]
}
```

**13. `src/api/buildings.api.ts`** — 新增 3 个 hook
```ts
export function useStockShelf() { ... }
export function useUnstockShelf() { ... }
export function useSetShelfPrice() { ... }
```

**14. `src/api/compat.ts`** — `normalizeBuilding` 加 `isRetail`、`shelves` 字段

### Phase 6: BuildingCard 零售 UI

**15. `src/features/buildings/BuildingCard.tsx`** — 零售建筑改版

建筑打开时，根据 `isRetail` 区分两种 layout：

```
┌─────────────────────────────┐
│ [Header: Shop Lv.3]         │
├─────────────────────────────┤
│ 🏪 Sells: Cake, Pizza ...   │  ← 原 "Produces" 改为 "Sells"
├─────────────────────────────┤
│ [Production Tab / Sell Tab] │  ← Tab 切换
│                             │
│ Sell Tab:                   │
│  ┌─ Shelf 1 ────────────┐  │
│  │ Cake    Qty: 45/100  │  │
│  │ Price: $180  [锁]    │  │
│  │ [下架] [调价]        │  │
│  ├──────────────────────┤  │
│  │ Revenue: $3,600      │  │
│  │ [提取]               │  │
│  └──────────────────────┘  │
│  [+ 上架新品]              │
│                             │
│ Production Tab:             │
│  (原生产界面，不变)          │
├─────────────────────────────┤
│ [升级] [移动] [拆除]        │
└─────────────────────────────┘
```

关键点：
- **零售建筑默认显示 Sell Tab**（而非 Production）
- **"Sells" 标签**显示该建筑可卖的资源（即 catalog entry 的 produces）
- **Production Tab** 保留，供玩家选择配方生产物品（跟现在一样）
- **生产出来的物品自动进入仓库**，需手动上架到 Shelf 才能卖出

### Phase 7: BuildView 市场标识

**16. `src/features/buildings/BuildView.tsx`** 
- 零售建筑在市场列表中标记 🏪 或 "零售" badge
- `produces` 标签改为 `sells`（对零售建筑）

---

## 完整数据流

```
交易所买 Flour
      ↓
仓库有 Flour
      ↓
BuildingCard → Production Tab → 选 Cake 配方
      ↓
生产完成 → Claim → 仓库有 Cake
      ↓
切换到 Sell Tab → 上架(Stock) Cake 到货架, 定价 $180
      ↓
每 60s tick:
  NPC 零售 → 遍历 Shop 货架 → UnitsSoldPerHour($180, Lv3, ...)
  → 扣货架 2 个 Cake → 公司 +$360 → 记账
      ↓
玩家打开 BuildingCard → 看到 Revenue、货架余量
      ↓
可随时下架、调价、提取收入
```

---

## 变更汇总

| # | 文件 | 变更类型 |
|---|---|---|
| 1 | `backend-next/internal/domain/building/types.go` | 修改 |
| 2 | `backend-next/internal/domain/company/types.go` | 修改 |
| 3 | `backend-next/internal/catalog/catalog.go` | 修改 |
| 4 | `decompiled/data/buildings.json` | 修改 |
| 5 | `backend-next/internal/app/building/shop.go` | 修改 |
| 6 | `backend-next/internal/app/building/service.go` | 修改 |
| 7 | `backend-next/internal/app/building/retail.go` | **新建** |
| 8 | `backend-next/internal/app/market/retail.go` | 修改 |
| 9 | `backend-next/internal/httpapi/building_handler.go` | 修改 |
| 10 | `backend-next/internal/httpapi/router.go` | 修改 |
| 11 | `backend-next/openapi/openapi-draft.yaml` | 修改 |
| 12 | `client/.../src/game/types.ts` | 修改 |
| 13 | `client/.../src/api/buildings.api.ts` | 修改 |
| 14 | `client/.../src/api/compat.ts` | 修改 |
| 15 | `client/.../src/features/buildings/BuildingCard.tsx` | **大改** |
| 16 | `client/.../src/features/buildings/BuildView.tsx` | 修改 |
