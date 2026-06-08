# 零售建筑设计方案

日期: 2026-06-08

## 一、零售建筑列表

以下 5 个建筑应被归类为 **retail**（零售型），它们作为终端销售点，直接从库存中取出物品卖给 NPC 顾客：

| Kind | 建筑 | 当前产出 | 可售品类 | 描述 |
|---|---|---|---|---|
| 6 | **Market Stall** | Coffee | beverage | 小摊，快速零售 Coffee |
| 7 | **Cafe** | Pizza, Coffee | meal, beverage | 堂食/外带 |
| 8 | **Food Truck** | Pizza, Vegetables | meal, raw | 移动售卖 |
| 9 | **Restaurant** | Steak, Pizza, Cake | protein, meal, dessert | 正餐 |
| 12 | **Shop** | Steak, Pizza, Cake, Coffee | protein, meal, dessert, beverage | 综合零售终端 |

其余 7 个建筑保持 **production** 类型。

## 二、核心设计思路

**当前问题**：零售系统遍历整个公司库存自动卖，Farm 的 Grain 也能"零售"。

**解决方向**：零售不再卖全部库存。玩家必须把商品**上架**到零售建筑的**货架**（Shelf）上，零售系统只卖货架上的商品。每个零售建筑有独立的货架、库存和售价。

## 三、数据模型变更

### 3.1 BuildingEntry（静态目录）

```go
// catalog/catalog.go — BuildingEntry 新增
type BuildingEntry struct {
    // ... 现有字段
    RetailSlots    int  `json:"retailSlots"`    // 基础货架数 (默认 0 = 不可零售)
    SlotPerLevel   int  `json:"slotPerLevel"`  // 每级增加货架数
}
```

`buildings.json` 也要对应更新，production 建筑 `retailSlots: 0`，retail 建筑设合理值。

### 3.2 Building（公司持有的建筑，domain/company/types.go）

```go
// ShelfItem 货架上的商品
type ShelfItem struct {
    ResourceID int     `json:"resource_id"`
    Quantity   int     `json:"quantity"`    // 当前库存
    MaxQty     int     `json:"max_qty"`     // 最大容量（受建筑等级影响）
    Price      float64 `json:"price"`       // 单价
    PriceLock  bool    `json:"price_lock"`  // true=手动定价, false=跟随市场价
}

// Building 新增 Shelves 字段
type Building struct {
    // ... 现有字段
    Shelves    []ShelfItem `json:"shelves,omitempty"`
}
```

### 3.3 CatalogEntry（domain/building/types.go）

```go
type CatalogEntry struct {
    // ... 现有字段
    RetailSlots  int `json:"retail_slots"`
    SlotPerLevel int `json:"slot_per_level"`
}
```

## 四、定价机制

### 4.1 价格来源（优先级从高到低）

1. **手动定价**（`PriceLock = true`）→ 玩家自行输入的单价
2. **市场价**（`PriceLock = false` 或默认） → `ticker.lastPrice`，无 ticker 则 `basePrice`
3. **最低价** → `modeledProductionCostPerUnit`（不得低于成本价）

### 4.2 定价策略（后续可扩展）

| 模式 | 说明 | 当前是否做 |
|---|---|---|
| Auto（跟随市场） | 价格自动取市场最新成交价，tick 刷新 | ✅ 做 |
| Manual（固定价） | 玩家输入固定价，锁定 | ✅ 做 |
| Markup（加成%） | 玩家设一个加成百分比，自动计算 | ❌ 后续 |

### 4.3 价格对销量的影响（已有）

`formula.UnitsSoldPerHour` 已经包含价格参数 `price`，价格越高销量越低。直接复用。

## 五、零售流程（新逻辑）

### 5.1 定时零售（Scheduler tick）

```
ProcessRetailSales() 修改为：
  遍历所有 NPC 公司：
    遍历公司的 Buildings：
      只处理 type=retail 的建筑（kind 6/7/8/9/12）
        遍历建筑的 Shelves：
          取 Shelf.Price（如果 PriceLock=false 则刷新为市场价）
          用 shelf 的 resourceId/quantity/price + 建筑 kind 参数
          → UnitsSoldPerHour(...)
          → 扣货架库存
          → 公司加钱
          → 记账
```

### 5.2 玩家零售补算（CatchUpPlayerRetail）

```
同上，但按 elapsed 时间缩放。
只处理零售建筑的货架商品。
```

### 5.3 建筑等级 → 销售参数映射

| 参数 | 来源 |
|---|---|
| `storeSize` | 建筑 Level（等级越高卖得越快） |
| `buildingKindModifier` | 静态目录 BuildingEntry 中的值（目前全 1.0，可按建筑类型差异化） |
| `saturation` | 保持 1.0（全市场统一，后续可做维度追踪） |
| `salesModPct` | 连接高管 `salesBonusAtLevel` |

## 六、新增业务流程

### 6.1 上架（Stock）

玩家从仓库取物品放到零售建筑的货架上。

```
POST /api/v2/buildings/{buildingId}/stock
Request: { resourceId, quantity, price? }
Response: { shelf: ShelfItem }
```

约束：
- 建筑必须是零售类型
- 货架未满（当前数量 < maxSlots）
- 仓库有足够库存
- price 可选；不传则自动跟随市场价

### 6.2 下架（Unstock）

从货架取回仓库。

```
POST /api/v2/buildings/{buildingId}/unstock
Request: { resourceId, quantity }
Response: { shelf: ShelfItem }
```

### 6.3 调价（SetPrice）

调整货架商品售价。

```
POST /api/v2/buildings/{buildingId}/shelf-price
Request: { resourceId, price, lock }
Response: { shelf: ShelfItem }
```

### 6.4 建筑详情（已有 GetBuilding 增强）

返回货架信息：

```json
{
  "id": "...",
  "kind": 12,
  "name": "Shop",
  "level": 3,
  "shelves": [
    { "resourceId": 10, "quantity": 50, "maxQty": 200, "price": 180, "priceLock": false }
  ]
}
```

## 七、需要变更的文件清单

### 后端（backend-next）

| 文件 | 变更 |
|---|---|
| `internal/domain/building/types.go` | CatalogEntry 加 RetailSlots, SlotPerLevel |
| `internal/domain/company/types.go` | Building 加 Shelves []ShelfItem；定义 ShelfItem |
| `internal/catalog/catalog.go` | BuildingEntry 加 RetailSlots, SlotPerPerLevel；加载 buildings.json 新字段 |
| `internal/app/building/service.go` | buildingToDTO 加 shelves 字段；BuildingMarket 标记零售类型 |
| `internal/app/building/shop.go` | BuyBuilding 初始空 Shelves；附加零售类型校验 |
| `internal/app/building/retail.go` | **新建** — StockShelf, UnstockShelf, SetShelfPrice |
| `internal/app/market/retail.go` | ProcessRetailSales 改为只遍历零售建筑货架 |
| `internal/httpapi/building_handler.go` | 新增 stock/unstock/set-price 端点注册 |
| `internal/httpapi/route.go` | 注册新路由 |
| `internal/scheduler/scheduler.go` | 无需改动（ProcessRetailSales 内部逻辑已变） |
| `internal/formula/retail.go` | 无需改动 |
| `decompiled/data/buildings.json` | 补充 retailSlots / slotPerLevel 字段 |
| `openapi/openapi-draft.yaml` | 新增 StockRequest/UnstockRequest/ShelfPriceRequest 等 Schema |

### 前端（后续）

| 文件 | 变更 |
|---|---|
| `src/features/buildings/` | 零售建筑卡片增加货架管理 UI |
| `src/api/buildings.api.ts` | 新增 stock/unstock/setPrice 接口 |
| 翻译文件 | 新增零售相关文案 |

## 八、实施步骤

```
Phase 1 — 数据模型
  □ 修改 BuildingEntry (catalog) + CatalogEntry (domain) + Building (company domain)
  □ 修改 buildings.json 补充 retail 字段
  □ 更新 catalog 加载逻辑

Phase 2 — 零售服务
  □ 新建 internal/app/building/retail.go (Stock/Unstock/SetPrice)
  □ 修改 ProcessRetailSales → 从零售建筑货架卖货
  □ 修改 CatchUpPlayerRetail → 同上

Phase 3 — HTTP API
  □ 新增 stock/unstock/shelf-price 端点
  □ 更新 openapi-draft.yaml
  □ buildingToDTO 暴露 shelves

Phase 4 — 验证
  □ go test
  □ 手动验证：dev 建 Shop → stock 商品 → 验证零售卖货
```

## 九、关键决策记录

| 决策 | 选择 | 理由 |
|---|---|---|
| 库存模型 | 货架独立数量（从公司库存扣减到货架） | 区分"仓库库存"和"上架待售"，玩家明确知道在卖什么 |
| 定价默认 | 跟随市场价（`PriceLock=false`） | 零操作门槛，需要固定价时再手动锁定 |
| 零售触发 | 只卖零售建筑货架上的商品 | 解决 raw material 自动零售问题 |
| 建筑升级 | 增加货架数量和容量 | 线性递进，鼓励升级 |
| 高管加成 | 接入 `salesBonusAtLevel` → `salesModPct` | 已实现，只差连线 |
