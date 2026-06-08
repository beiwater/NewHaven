# 销售建筑实现分析 — backend-next

日期: 2026-06-08

## 一、建筑定义 (静态数据)

**文件**: `decompiled/data/buildings.json`

两种销售类建筑：

| ID | Kind | 名称 | 成本 | 产出 | 描述 |
|---|---|---|---|---|---|
| 6 | 6 | **Market Stall** | 62,000 | Coffee (ID 11) | 为本地市场客流生产咖啡 |
| 12 | 12 | **Shop** | 150,000 | Steak, Meal, Bread, Coffee (ID 8,9,10,11) | 高价值加工食品零售点 |

> ⚠️ `buildings.json` 中所有建筑的 `type` 都是 `"production"`，包括 Market Stall 和 Shop。但 domain 层 `CatalogEntry.Type` 注释写着 `"production" or "retail"`，说明设计上有分类意图但未落地。

## 二、经济模型

**文件**: `decompiled/data/economy_model.json`

12 个资源全部有 `state_1` 零售参数：

- `modeledProductionCostPerUnit` — 每单位生产成本
- `modeledStoreWages` — 商店工资
- `modeledUnitsSoldAnHour` — 每小时预估销售单位
- `buildingLevelsNeededPerUnitPerHour` — 每单位每小时所需建筑等级
- `buildingKindModifier` — 建筑种类修正值

这些参数在 `formula.UnitsSoldPerHour()` 中使用。

## 三、零售销售系统

### 核心组件

| 组件 | 文件 | 功能 |
|---|---|---|
| 零售公式 | `internal/formula/retail.go` | `UnitsSoldPerHour()` — 综合价格、质量、饱和度、工时、天气等计算每小时销量 |
| NPC 零售 | `internal/app/market/retail.go` | `ProcessRetailSales()` — 每 tick 遍历 NPC 公司卖货 |
| 玩家零售补算 | 同上 | `CatchUpPlayerRetail()` — 按需结算玩家未处理的零售收入 |
| 定时器 | `internal/scheduler/scheduler.go` | 每 60s 调用 `ProcessRetailSales` |
| HTTP 入口 | `internal/httpapi/company_handler.go` | `handleListMyCompanies` 和 `handleCompleteTutorial` 中调用补算 |
| 高管加成 | `internal/httpapi/executive_handler.go` | `salesBonusAtLevel()` 已定义但未接零售 |

### 数据流

```
Scheduler (60s tick)
  └─ ProcessRetailSales()
       └─ 每个 NPC 公司 (LastRetailAt == "")
            └─ 遍历公司库存
                 └─ 查找资源的 economy model
                 └─ 取市场价 (ticker.lastPrice / basePrice)
                 └─ UnitsSoldPerHour(...)
                 └─ 扣库存 + 加钱 + 记账

玩家请求 (ListMyCompanies / CompleteTutorial)
  └─ CatchUpPlayerRetail()
       └─ 同上，按 elapsedSeconds 缩放
```

## 四、现有硬编码 / 待办

| 问题 | 当前值 | 说明 |
|---|---|---|
| **建筑与零售脱钩** | — | 零售遍历所有库存物品，不检查公司是否有 Shop/Market Stall。生产建筑（Farm/Barn/Mill）产品也会自动零售 |
| **quality** | `4.0` | 硬编码中间值，无逐物品质量追踪 |
| **saturation** | `1.0` | 硬编码平衡市场 |
| **salesModPct** | `0.0` | 未接入高管加成（`salesBonusAtLevel` 已实现但未引用） |
| **storeSize** | `1` | 硬编码，未读取建筑等级/类型 |
| **acceleration** | `1.0` | 硬编码默认值 |
| **weather** | `1.06` | 硬编码天气乘数 |
| **building type** | `"production"` | Market Stall 和 Shop 的 type 字段未正确标记为 "retail" |

## 五、建筑系统完整度

| 功能 | 状态 | 文件 |
|---|---|---|
| 建筑目录加载 | ✅ | `internal/catalog/catalog.go` |
| 建筑购买 (BuyBuilding) | ✅ | `internal/app/building/shop.go` |
| 建筑升级 (UpgradeBuilding) | ✅ | 同上 |
| 建筑放置/移动/拆除/储藏 | ✅ | `internal/app/building/placement.go` |
| 建筑市场列表 | ✅ | `internal/app/building/service.go` |
| 建筑解锁等级 | ✅ | 同上 |
| 零售 NPC 定时结算 | ✅ | `internal/app/market/retail.go` |
| 零售玩家补算 | ✅ | 同上 |
| 高管加成函数 | ✅ 已定义未集成 | `internal/httpapi/executive_handler.go` |
| 建筑-零售关联 | ❌ | 零售不检查建筑类型 |
| 质量/饱和度追踪 | ❌ | 硬编码 |
| 天气/加速度集成 | ❌ | 硬编码 |
| storeSize 与建筑等级联动 | ❌ | 硬编码 |

## 六、关键结论

**零售销售系统已基本完工**（公式、NPC 定时结算、玩家补算），但**建筑与零售之间没有关联**。当前状态下：

1. 玩家不需要拥有 Shop 或 Market Stall，只要库存有物品就会自动零售
2. Shop/Market Stall 只是像其他生产建筑一样产出物品，没有零售专属优势
3. 高管对销售的加成已定义但未接入零售流程

完整功能需要：限制零售只能由零售建筑/品类触发、接入建筑等级/类型作为 storeSize、接入饱和度追踪、连接高管加成系统。
