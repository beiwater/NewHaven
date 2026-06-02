# Source Digest - 2026-06-03

本文件整理自 2026-06-02 的三个源材料：

- `Food Economy Actuarial Model v1.3.2 - 高管研究曲线精算表/`
- `dev-plan.md`
- `Food Economy Actuarial Design v1.3.1 - 极简建筑产量版.docx`

用途：给后续模型执行任务前快速建立共同上下文，避免反复读取杂乱文件夹。

## 1. 经济模型核心口径

当前经济系统应采用“极简建筑产量版”：

- 建筑直接决定产量：`BuildingBaseOutputPerHour_Lv1` 是生产能力核心变量。
- 不做玩家工人管理：删除 `WorkerCount_Lv1`、`WorkerProductivity`、`BaseOutputPerWorkerHour` 等工人产量变量。
- 不用配方工作量黑箱：删除 `RecipeWorkload`，高级产品通过低基础产量、更多原料和更高价格体现。
- 默认产出可卖：不要用全局 `UnitsSold` 卡死销量，利润主要由价格和成本调节。
- 市场饱和度按商品组计算：`MarketSaturationGroup = TotalSupplyInGroup / TotalDemandInGroup`。
- 饱和度控制价格：供给过多时降低 `EffectivePrice`，而不是让库存完全卖不出去。
- 成本指数控制经济压力：`LaborCostIndex`、`MaterialCostIndex`、`EnergyCostIndex` 作为全局调参阀门。
- 高管和口碑使用相加：`ExecutiveBonus + GlobalReputationBonus`，避免相乘导致数值爆炸。
- 自动交易机器人提供流动性：Market Maker 负责基础买盘/卖盘，但不能无限买卖或制造套利。
- 建筑数量锁死：用建筑槽位、切换成本和建造时间防止玩家每天无成本追逐最低饱和度商品。

## 2. 关键公式

```text
OutputPerHour_Lv1 =
  BuildingBaseOutputPerHour_Lv1 * (1 + FinalProductionSpeedBonus)

OutputPerHour_Level =
  OutputPerHour_Lv1 * BuildingLevel

LaborCostPerHour_Level =
  BaseLaborCostPerHour_Lv1 * LaborCostIndex * BuildingLevel

InputCostPerHour_Level =
  OutputPerHour_Level * InputQtyPerUnit * InputUnitPrice * MaterialCostIndex

EnergyCostPerHour_Level =
  BaseEnergyCostPerHour_Lv1 * EnergyCostIndex * BuildingLevel

SaturationPriceMultiplier =
  CLAMP(0.70, 1.10, 1 - MAX(0, MarketSaturationGroup - 1) * SaturationK)

EffectivePrice =
  BasePrice * SaturationPriceMultiplier * EventPriceMultiplier

RevenuePerHour_Level =
  OutputPerHour_Level * EffectivePrice

TotalCostPerHour_Level =
  InputCostPerHour_Level
  + LaborCostPerHour_Level
  + EnergyCostPerHour_Level
  + MaintenanceCostPerHour_Level
  + BaseManagementCost
  + ExecutiveCost
  + TaxCostPerHour

NetProfitPerHour_Level =
  RevenuePerHour_Level - TotalCostPerHour_Level
```

## 3. 精算表 v1.3.2 摘要

### Dashboard

- 目标每小时利润：约 `6000`
- 建筑平均利润：约 `6113`
- 平均饱和度：`1.00`
- 食品研究市场指数：`1.00`
- 甜蜜点等级：`7`

### Market Control

核心总控变量：

- `LaborCostIndex`
- `MaterialCostIndex`
- `EnergyCostIndex`
- `GlobalDemandIndex`
- `EventPriceMultiplier`
- `SaturationK`
- `RetailTaxRate`
- `BotSpread`

### Building Balance Lv1

Lv1 建筑样例：

| 建筑 | 类型 | 商品组 | Lv1产量 | 基础价格 |
|---|---|---|---:|---:|
| Farm | Production | Grain | 500 | 23 |
| Barn | Production | Dairy | 320 | 42 |
| Mill | Processing | Processed | 220 | 68 |
| Kitchen | Processing | Processed | 140 | 139 |
| Bakery | Production | Bakery | 90 | 245 |
| Market Stall | Sales | GeneralMarket | 700 | 18 |

### Level Curve

- 产能按等级线性增长。
- 建筑升级成本递增：`UpgradeCost(Level) = BaseBuildCost * Level`。
- 总建筑成本：`TotalBuildingCost(Level) = BaseBuildCost * Level * (Level + 1) / 2`。
- 甜蜜点后管理成本加速上升，用来防止无限升级。

### Research Curve

研究节点使用食品研究数量和现金费用推进：

- `BaseFoodResearchQty`
- `FloatingFoodResearchQty`
- `FoodResearchPrice`
- `MarketPurchaseCost`
- `CashFee`
- `TotalResearchCost`

研发页面实现时，必须能展示研究项目、消耗、进度和完成收益。

### Executive Curve

高管升级曲线需要表现：

- 早期升级快。
- 中后期成本明显上升。
- 生产加成、销售加成、管理折扣边际收敛。
- 高管工资是后期运营成本的一部分，不是免费增益。

### Saturation & Market Maker

商品组包括：

- `Grain`
- `Dairy`
- `Processed`
- `Bakery`
- `GeneralMarket`
- `CafeDessert`
- `StreetFood`
- `RestaurantMeal`
- `Finance`

机器人价格：

```text
BotBidPrice = FairPrice * (1 - BotSpread)
BotAskPrice = FairPrice * (1 + BotSpread)
```

机器人只提供流动性，不能无限接盘、无限出货或制造无风险套利。

### Sales Mechanics

销售建筑分为：

- `Market Stall`：傻瓜操作，稳定低风险回血。
- `Cafe`：低操作，轻菜单组合，稳定现金流。
- `Food Truck`：地点和活动驱动，高波动高上限。
- `Restaurant`：后期机制建筑，高成本、高上限，受菜单、品牌和高管影响。
- `Trading Hub`：金融/流动性建筑，不应产生无限套利。

## 4. dev-plan 中明确要做的事项

### 前端优先

- Research 页面：侧边栏已有入口，但页面为空。
- Financial 页面：后端端点已存在，前端为 0。
- Executive 页面：高管市场、招募、培训、详情；挖角和报价先不做。
- Chat：把 mock 数据替换为后端消息 API。
- Power-up UI：原 SimBoost 不能叫 SimBoost，前端展示名改为 `Power-up`。
- 次要加固：建筑市场 API 路径、Research 图标语义、TopBar 硬编码能量/工人。

### 后端优先

- 多 Player 与真实认证。
- Postgres 持久化完整性审计和补齐。
- 排行榜 API：按净资产/等级排名，并配套前端页面。
- 搜索/筛选：市场订单、高管、债券等列表加分页和查询参数。
- API 版本逐步对齐到 v3。
- 单元测试：公式、市场交易、债券、政府合同、生产链、反作弊、反洗钱。

### 暂缓

- Bond Market 页面。
- Auction 页面。
- Aerospace 页面。
- WebSocket 实时推送。
- 资源图标补齐。
- 动态天气/生产修正。

