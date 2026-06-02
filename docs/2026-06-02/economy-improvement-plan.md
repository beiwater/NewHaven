# 经济系统改进计划 (v1.3 → v2.0)

**基于:**  
- Food Economy Actuarial Design v1.3.1 - 极简建筑产量版.docx  
- Food Economy Actuarial Model v1.3.2 - 高管研究曲线精算表  

**当前项目版本:** 混乱的混合体（Sim Companies 反编译 + 自定义修改）  
**目标:** 全部替换为新精算模型，实现可调平衡的食品经济

---

## 一、现状 vs 新设计

| 维度 | 当前项目 | v1.3 新设计 |
|------|---------|------------|
| 工人管理 | 有 SalaryMid/工资等级 | ❌ 删除，改用 LaborCostIndex |
| 配方工作量 | RecipeWorkload / producedFrom | ❌ 删除，使用 BuildingBaseOutputPerHour |
| 生产时间 | 通过 formula/production.go 复杂公式 | ❌ 删除，直接由每小时产量反推 |
| 零售销量 | 13 参数 retail.go 零售模型 | ❌ 删除，默认产多少卖多少 |
| 价格控制 | 自由市场匹配 + 机器人 | MarketSaturation 按商品组控制 |
| 建筑等级 | 等级 × 5000 递增 | 明确的升级成本表 |
| 管理费 | admin_overhead_base 固定系数 | ManagementK × Level 递增，边际收敛 |
| 高管 | 搜索/招募/培训/挖角全有 | 精简为 10 级培训曲线 + 固定时薪 |
| 研究 | 4 个硬编码项目 | 10 节点树状研究曲线 |
| 销售建筑 | 无专门销售建筑 | 4 种（MarketStall/Cafe/FoodTruck/Restaurant） |
| 建筑数量 | 无限制 | BuildingSlotLimit 锁死 |
| 离线收入 | 8 小时上限 | ❌ 删除（或简化为固定比例） |
| 债券 | 发行/购买/利息完整 | 保留但简化 |
| 品质 | 多级品质系统 | ❌ 删除（默认为 0） |
| 资金水槽 | 无 | 升级成本 + 零售税 + 研究费用 |

---

## 二、新公式体系（完整替换）

### 2.1 建筑产量

```
OutputPerHour_Lv1 = BuildingBaseOutputPerHour_Lv1 × (1 + FinalProductionSpeedBonus)
OutputPerHour_Level = OutputPerHour_Lv1 × BuildingLevel
```

**含义：** 1 级建筑定基础产量，等级线性加倍。不再需要配方/工作量/工人。

### 2.2 成本系统

```
LaborCostPerHour_Level    = BaseLaborCostPerHour_Lv1 × LaborCostIndex × BuildingLevel
InputCostPerHour_Level     = OutputPerHour_Level × InputQtyPerUnit × InputUnitPrice × MaterialCostIndex
EnergyCostPerHour_Level    = BaseEnergyCostPerHour_Lv1 × EnergyCostIndex × BuildingLevel
MaintenanceCostPerHour_Level = BaseMaintenanceCostPerHour_Lv1 × BuildingLevel
BaseManagementCost         = BuildingLevel × ManagementK
TaxCostPerHour             = RevenuePerHour_Level × RetailTaxRate
```

**含义：** 三类成本指数（Labor/Material/Energy）作为全局调控阀门。

### 2.3 市场饱和度

```
MarketSaturationGroup = TotalSupplyInGroup / TotalDemandInGroup
SaturationPriceMultiplier = CLAMP(0.70, 1.10, 1 - MAX(0, MarketSaturationGroup - 1) × SaturationK)
EffectivePrice = BasePrice × SaturationPriceMultiplier × EventPriceMultiplier
```

**含义：** 按商品组（Grain/Dairy/Processed/Bakery/GeneralMarket）统计供需，价格在 0.7x-1.1x 之间浮动。不再需要订单簿撮合。

### 2.4 销售建筑

```
// MarketStall (傻瓜): 无加成
Revenue = Output × EffectivePrice

// Cafe (傻瓜): 菜单加成
EffectivePrice_Cafe = BasePrice × SPM × EventPM × (1 + SimpleMenuBonus)

// FoodTruck (机制): 地点加成
FoodTruckSpecialBonus = LocationBonus + EventLocationBonus - LocalSaturation

// Restaurant (机制): 综合加成
RestaurantSpecialBonus = MenuQualityBonus + GlobalReputationBonus + ExecutiveSalesBonus - IngredientSaturationImpact
```

### 2.5 速度加成（全相加）

```
FinalProductionSpeedBonus = ResearchBonus + LevelPoints + ExecutiveBonus + ReputationBonus + BuildingBonus
```

**不再相乘。** 所有加成相加，防止数值爆炸。

### 2.6 最终利润

```
NetProfitPerHour_Level = Revenue - InputCost - LaborCost - EnergyCost - Maintenance - MgmtCost - Tax - ExecutiveCost
```

---

## 三、新建筑与商品组

### 生产建筑

| 建筑 | Type | 商品组 | BasePrice | 产出 (Lv1/h) |
|------|------|--------|-----------|-------------|
| Farm | Production | Grain | 23 | 500 |
| Barn | Production | Dairy | 42 | 240 |
| Mill | Production | Processed | 103.5 | 200 |
| Kitchen | Production | Processed | 103.5 | 160 |
| Bakery | Production | Bakery | 245 | 80 |

### 销售建筑

| 建筑 | Type | 风险区间 | 基准加成 | 成本特征 |
|------|------|---------|---------|---------|
| Market Stall | Simple | -5% ~ +8% | 0% | 低运营成本 |
| Cafe | Simple | -8% ~ +12% | SimpleMenuBonus | 菜单成本 |
| Food Truck | Mechanic | -25% ~ +40% | 地点加成 | 燃油费 |
| Restaurant | Mechanic | -20% ~ +35% | 综合加成 | 高管 + 品牌 |

### 市场饱和度组

| 商品组 | 需求量 | Base FairPrice | BotBid | BotAsk |
|--------|-------|---------------|--------|--------|
| Grain | 500/h | 23 | 21.62 | 24.38 |
| Dairy | 320/h | 42 | 39.48 | 44.52 |
| Processed | 360/h | 103.5 | 97.29 | 109.71 |
| Bakery | 90/h | 245 | 230.3 | 259.7 |
| GeneralMarket | 700/h | 18 | 16.92 | 19.08 |

---

## 四、成长曲线

### 4.1 等级 (Level) 曲线

| Level | 产出倍率 | 升级费 | 累计费 | 净利 | 管理费 | 实际净利 |
|-------|---------|--------|--------|------|--------|---------|
| 1 | 1x | 10k | 10k | 6k | 0 | 6,000 |
| 2 | 2x | 20k | 30k | 12k | 0 | 12,000 |
| 3 | 3x | 30k | 60k | 18k | 0 | 18,000 |
| 4 | 4x | 40k | 100k | 24k | 0 | 24,000 |
| 5 | 5x | 50k | 150k | 30k | 5k | 25,000 |
| 6 | 6x | 60k | 210k | 36k | 12k | 24,000 |
| 7 | 7x | 70k | 280k | 42k | 21k | 21,000 |
| 8 | 8x | 80k | 360k | 48k | 32k | 16,000 |

**甜蜜点 (Sweet Spot): Level 4-6**，超过后管理费加速吞噬利润。

### 4.2 研究曲线（10 节点）

| ID | 名称 | Tier | 商品组 | 节点热度 | 研究量 | 价格 | 总成本 | 加成 |
|----|------|------|--------|---------|-------|------|--------|------|
| R01 | 基础种植 | 1 | Grain | 0% | 88 | 960 | 1,460 | 产+1% |
| R02 | 基础加工 | 1 | Processed | 0% | 12 | 1,440 | 2,340 | 产+1% |
| R03 | 奶蛋优化 | 1 | Dairy | 2% | 18 | 2,160 | 3,660 | 产+1.2% |
| R04 | 标准化厨房 | 1 | Processed | 2% | 24 | 2,880 | 4,830 | 产+1.2% |
| R05 | 发酵工艺 | 2 | Dairy | 5% | 32 | 4,800 | 7,970 | 产+1.5% |
| R06 | 自动烘焙线 | 2 | Bakery | 5% | 40 | 6,000 | 9,910 | 产+1.5% |
| R07 | 冷链物流 | 2 | Dairy | 8% | 50 | 7,500 | 12,350 | 销+1.5% |
| R08 | 精品菜单 | 3 | Bakery | 10% | 65 | 11,700 | 19,090 | 销+2% |
| R09 | 快餐连锁 | 3 | Processed | 12% | 80 | 14,400 | 23,410 | 产+2% |
| R10 | 全能厨房 | 3 | Mixed | 15% | 100 | 18,000 | 29,170 | 产+2.5% |

**节点热度（NodeHeat）：** 完成前一节点后解锁下一节点概率/成本递进。  
**设计节奏：** Tier1 → 当天完成，Tier2 → 1-2 天，Tier3 → 3-5 天。

### 4.3 高管曲线（10 级）

| 等级 | 培养成本 | 时薪 | 产量加成 | 销量加成 | 管理折扣 | PowerScore |
|------|---------|------|---------|---------|---------|-----------|
| 0 | 0 | 0 | 0% | 0% | 0% | 0 |
| 1 ⭐ | 2,196 | 3 | 1% | 1% | 0.5% | 2.5 |
| 2 ⭐⭐ | 2,809 | 5 | 1.8% | 1.8% | 1% | 4.6 |
| 3 ⭐⭐⭐ | 3,545 | 8 | 2.5% | 2.5% | 1.5% | 6.5 |
| 4 | 4,370 | 12 | 3.2% | 3.2% | 2% | 8.4 |
| 5 | 5,289 | 16 | 3.8% | 3.8% | 2.5% | 10.1 |
| 6 | 6,305 | 20 | 4.4% | 4.4% | 3% | 11.8 |
| 7 | 7,424 | 25 | 5% | 5% | 3.5% | 13.5 |
| 8 | 8,649 | 30 | 5.5% | 5.5% | 4% | 15.0 |
| 9 | 9,985 | 35 | 6% | 6% | 4.5% | 16.5 |
| 10 | 11,437 | 40 | 6.5% | 6.5% | 5% | 18.0 |

**边际收敛：** 高管加成逐级递减（1% → 0.6% → 0.5%），成本加速上升。

---

## 五、全局调控变量

| 变量 | 默认值 | 用途 |
|------|--------|------|
| `LaborCostIndex` | 1.0 | 全市场人工成本 |
| `MaterialCostIndex` | 1.0 | 全市场原料成本 |
| `EnergyCostIndex` | 1.0 | 全市场能源成本 |
| `GlobalDemandIndex` | 1.0 | 全市场需求 |
| `EventPriceMultiplier` | 1.0 | 活动价格倍率 |
| `SaturationK` | 0.3 | 饱和度→价格敏感度 |
| `RetailTaxRate` | 3% | 收入税（金币水槽） |
| `BotSpread` | 0.06 | 机器人买卖价差 |
| `MaxBuildingsPerPlayer` | 6 | 建筑数量上限 |
| `ManagementEdgeK` | 2.0 | 管理费加速系数 |

---

## 六、代码修改清单

### 6.1 删除（完全移除的模块）

| 文件/模块 | 原因 |
|----------|------|
| `backend/internal/formula/production.go` | 替换为简单线性公式 |
| `backend/internal/formula/retail.go` | 替换为 MarketSaturation |
| `backend/internal/formula/admin.go` | 替换为 ManagementK |
| `backend/internal/formula/market.go` | 只保留 TickSize（给机器人用）|
| `backend/internal/service/market_match.go` | 市场撮合不再需要 |
| `backend/internal/service/market_depth.go` | 订单簿深度不再需要 |
| `backend/internal/service/market_trade.go` | 取消订单系统 |
| `backend/internal/service/market_competition.go` | 替换为新 Bot 系统 |
| `backend/internal/service/production.go` | 重写为 OutputPerHour 模式 |
| `backend/internal/service/building.go` | 重写升级公式 |
| `backend/internal/service/offline.go` | 简化或删除 |
| `backend/internal/service/research.go` | 替换为 10 节点树 |
| `backend/internal/service/simboost.go` | 删除或简化 |
| `backend/internal/service/auction.go` | 删除 |
| `backend/internal/service/aerospace.go` | 删除 |
| `backend/internal/service/executive.go` | 替换为新曲线 |
| `backend/internal/handler/market.go` | 大量端点删除或简化 |
| `backend/internal/handler/bond.go` | 保留但简化 |
| `backend/internal/anticheat/` | 可保留 |
| `backend/internal/aml/` | 可保留 |

### 6.2 新增

| 文件 | 内容 |
|------|------|
| `internal/service/economy_core.go` | 核心经济循环（每小时 tick 计算所有建筑产出/成本/利润）|
| `internal/service/market_saturation.go` | 按商品组统计供需、计算 EffectivePrice |
| `internal/service/sales_building.go` | 4 种销售建筑逻辑 |
| `internal/service/research_tree.go` | 10 节点研究树 |
| `internal/service/executive_curve.go` | 10 级高管曲线 |
| `internal/service/building_limit.go` | 建筑数量限制 |
| `internal/formula/building_output.go` | 新产量公式 |
| `internal/formula/cost.go` | 成本公式 |
| `internal/formula/saturation.go` | 饱和度公式 |
| `internal/handler/economy.go` | 新 API 端点 |
| `internal/scheduler/economy_tick.go` | 经济 tick（替代现有 scheduler）|

### 6.3 需要保留并适配的

| 模块 | 改动 |
|------|------|
| `handler/auth.go` | 保留 |
| `handler/player.go` | 简化（删除成就/SimBoost/离线等）|
| `handler/company.go` | 适配新数据模型 |
| `handler/recipe.go` | 删除（由 building output 替代）|
| `handler/health.go` | 保留 |
| `handler/bond.go` | 保留但简化 |
| `handler/government.go` | 保留（与销售建筑联动）|
| `handler/order.go` | 保留（每日订单）|
| `service/bond.go` | 适配新经济模型 |
| `service/government.go` | 适配新价格体系 |
| `service/order.go` | 适配新奖励公式 |
| `model/types.go` | 大量精简，删除废弃字段 |
| `config/game.json` | 全部参数替换为 v1.3 参数表 |
| `data/loader.go` | 加载新 JSON 数据结构 |

### 6.4 前端改动

| 文件 | 改动 |
|------|------|
| `src/game/resources.ts` | 更新为 5 个商品组 |
| `src/features/market/MarketPage.tsx` | 改为饱和度看板，非订单簿 |
| `src/features/buildings/` | 适配新建筑系统 |
| `src/features/buildings/BuildingCard.tsx` | 显示 NetProfitPerHour |
| 新建 `src/features/research/` | 研究树页面 |
| 新建 `src/features/executives/` | 高管培训页面 |
| 新建 `src/features/sales/` | 销售建筑管理 |
| 新建 `src/features/dashboard/` | 经济看板（饱和度/价格/利润）|
| `src/store/` | 简化状态 |

---

## 七、实施阶段

### Phase 1: 数据准备
1. 根据 BuildingBalance 表生成新 `buildings.json`
2. 根据 ProductInputs 表生成新 `resources.json`
3. 根据 MarketControl 表生成新 `game.json`
4. 根据 ResearchCurve 表生成研究树 JSON
5. 根据 ExecutiveCurve 表生成高管曲线 JSON

### Phase 2: 核心公式替换
1. 实现 `formula/building_output.go`（简单线性）
2. 实现 `formula/cost.go`（三类成本指数）
3. 实现 `formula/saturation.go`（按商品组饱和度）
4. 删除旧的 formula/*（production/retail/market/admin）

### Phase 3: 服务层重写
1. `service/economy_core.go` — 核心经济 tick
2. `service/market_saturation.go` — 市场饱和度计算
3. `service/sales_building.go` — 4 种销售建筑
4. 删除旧的 market_match/trade/depth/competition
5. 重写 production.go

### Phase 4: 成长系统
1. `service/research_tree.go` — 10 节点树
2. `service/executive_curve.go` — 10 级曲线
3. 适配 player level → 新等级曲线

### Phase 5: Handler 适配
1. 删除废弃端点
2. 新增经济相关端点
3. 更新现有端点到新数据模型

### Phase 6: 前端
1. 市场页改为饱和度看板
2. 研究树页面
3. 高管培训页面
4. 销售建筑页面
5. 经济仪表盘

### Phase 7: 平衡调优
1. 跑经济循环，验证 NetProfitPerHour 合理
2. 调整 SaturationK、成本指数
3. 确认金字塔利润率

---

## 八、设计原则总结

```
建筑产量    = 定死的     (BuildingBaseOutputPerHour_Lv1)
建筑等级    = 线性翻倍   (× BuildingLevel)
成本        = 指数调控   (Labor/Material/Energy CostIndex)
价格        = 饱和度压制  (0.7x-1.1x 浮动)
加成        = 全相加     (不乘)
销售建筑    = 4 种      (2 简单 + 2 机制)
建筑数量    = 锁死       (MaxBuildingsPerPlayer)
成长        = 边际收敛   (等级/研究/高管都有甜蜜点)
资金水槽    = 升级+税+研究
流动性      = 机器人     (自动买卖)
```

从当前项目迁移到 v1.3 核心是 **砍掉大部分现有代码**，用全新的精简公式替代。工作量主要在删除和替换，不是新增。
