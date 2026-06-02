# Game Formulas

**Date:** 2026-06-02  
**Source:** code audit of `backend/internal/formula/` + embedded formulas in `backend/internal/service/`

---

## 1. Market Formulas (formula/market.go)

### Tick Size

```
TickStep(price):
  price ≥ 20000  → 500
  price ≥ 10000  → 100
  price ≥  5000  →  25
  price ≥  1000  →  10
  price ≥   500  →   5
  price ≥   200  →   2
  price ≥   100  →   1
  price ≥    50  →   0.5
  price ≥    20  →   0.25
  price ≥     5  →   0.1
  price ≥     2  →   0.05
  price ≥     1  →   0.01
  price ≥   0.5  →   0.005
  default         →   0.001
```

### Price Validation

```
IsValidTick(price):
  step = TickStep(price)
  return abs(price / step - round(price / step)) < 1e-9
```

### Exchange Fee

```
ExchangeFee(amount, price, feeRate) = ceil(amount × price × feeRate)
```

`feeRate` 来自 `game.json` → `exchange_fee_pct` (默认 0.04 = 4%)

---

## 2. Production Formulas (formula/production.go)

### Constants

```
SalaryMid = { 0: 655, 1: 700, 2: 745 }
AverageSalary = 345.0
RobotBonus = 4.0
```

### Base Production Rate

```
BaseProductionRate(producedPerHourRaw, salaryModifier, salaryLevel) =
  producedPerHourRaw × (AverageSalary / SalaryMid[salaryLevel]) ^ salaryModifier
```

### Produced Per Hour

```
ProducedPerHour(size, baseRate, salaryPercent, robotCount, isAccumulator,
                speedModifierPct, qualityPct, isMining):

  adjusted = baseRate
  if isMining:
    adjusted ×= qualityPct / 100
  adjusted ×= (speedModifierPct / 100) + 1

  effectiveSalary = salaryPercent
  if isAccumulator:
    effectiveSalary += RobotBonus × robotCount

  den = 1 - effectiveSalary / 100
  if den ≤ 0.01: den = 0.01

  return size × adjusted / den
```

### Production Time

```
ProductionTimeSeconds(size, buildingSalaryModifier, producedPerHour,
                      eventMultiplier):

  if producedPerHour ≤ 0: return +inf
  if eventMultiplier ≤ 0: eventMultiplier = 1

  return (345 × buildingSalaryModifier × size / producedPerHour) × eventMultiplier
```

---

## 3. Retail Formula (formula/retail.go)

### Units Sold Per Hour (零售销量公式)

完整的解构版零售模型，来自反编译数据：

```
Constants:
  RetailQualityWeight = 0.3
  RetailZor = 370.0

UnitsSoldPerHour(
  buildingKindModifier,
  buildingLevelsNeededPerUnitPerHour,
  modeledProductionCostPerUnit,
  modeledStoreWages,
  modeledUnitsSoldAnHour,
  price,
  quality,
  saturation,
  salesModifierPct,
  size,
  acceleration,
  weatherSellingSpeedMultiplier,
):
  d = clamp(2 - saturation, 0, 2)          // 需求系数 [0, 2]
  p = max(0.9, d/2 + 0.5)                  // 需求修正
  f = quality / 12                          // 品质系数
  g = RetailZor × (buildingLevelsNeeded... × modeledUnitsSoldAnHour + 1)
      × buildingKindModifier × (d/2 × (1 + f × RetailQualityWeight))
  m = modeledUnitsSoldAnHour × p
  if m ≤ 0: return 0

  den = price - modeledProductionCostPerUnit
  if den ≤ 0.0001: return 0               // 价格低于成本 → 不卖

  a = (modeledStoreWages + g) / (den × den)
  b = g - (m - price)² × a
  if b + modeledStoreWages ≤ 0: return 0

  res = (acceleration × den × 3600 - modeledStoreWages) / (b + modeledStoreWages)
  if size ≤ 0: size = 1
  if acceleration ≤ 0: acceleration = 1
  res = res / size / acceleration
  res = res - res × salesModifierPct / 100
  if weatherSellingSpeedMultiplier > 0:
    res = res / weatherSellingSpeedMultiplier
  if NaN or Inf or res < 0: return 0

  return res
```

---

## 4. Admin Overhead (formula/admin.go)

### 管理费 (COO 加成)

```
AdminOverheadWithCOO(adminOverhead, cooSkill) =
  adminOverhead - (adminOverhead - 1) × cooSkill / 100
```

`adminOverhead` 来自 `game.json` → `admin_overhead_base` (默认 1.35)

### CTO 生产倍率

```
CTOProductionMultiplier(ctoSkill) = (100 + ctoSkill × 2) / 100
```

即每点 CTO 技能 +2% 产量。

---

## 5. Bond Formulas (formula/bonds.go)

### Constants

```
BondFaceValue = 5000.0  (可通过 game.json 覆盖)
```

### Daily Interest

```
DailyBondInterest(amount, interestRatePct) =
  floor(amount × 50 × interestRatePct)
```

### Period Interest (per-bond calculation)

```
PeriodBondInterest(amount, interestRatePct) =
  floor(amount × BondFaceValue × interestRatePct) / 100
```

### Max Issuable Bonds

```
MaxIssuableBonds(totalBuildingValue, alreadySold) =
  max(0, floor(totalBuildingValue / BondFaceValue) - alreadySold)
```

即最多发行面值不超过总建筑价值的债券。

---

## 6. Embedded Service Formulas

### 6.1 建筑升级费用 (service/building.go)

```
UpgradeCost(currLevel, baseCost):
  nextLevel = currLevel + 1
  if baseCost ≤ 0: baseCost = kind × 5000
  cost = nextLevel × baseCost

产出倍率: outputMultiplier = nextLevel
```

### 6.2 建筑购买 (service/building_shop.go)

每种建筑固定价格（目前写死 4 种）：
- Farm Plot: 50,000
- Food Factory: 120,000
- Warehouse: 30,000
- Research Lab: 80,000

### 6.3 仓库升级 (service/building_shop.go)

```
WarehouseUpgradeCost(lvl) = (lvl + 1) × WarehouseUpgradeCost
新容量 = (WarehouseLevel + 1) × 1000
```

`WarehouseUpgradeCost` 来自 `game.json` → `warehouse_upgrade_cost` (默认 25,000)  
`WarehouseBaseCap` 来自 `game.json` → 1000

### 6.4 生产槽位升级 (service/production.go)

```
SlotUpgradeCost(current) = (current + 1) × SlotUpgradeCost
```

`SlotUpgradeCost` 来自 `game.json` → `slot_upgrade_cost` (默认 50,000)

### 6.5 取消生产退还 (service/production.go)

```
CancelRefund(inputs):
  for each (resourceID, qty) in inputs:
    退还 qty / 2 (向下取整)
```

退还 50% 原料。

### 6.6 生产持续时间 (service/production.go)

```
calcProductionDuration(resourceID, amount, durSec):
  if durSec ≤ 0:
    durSec = max(30, amount × 6)
  // 如果有经济模型数据:
  if buildingLevelsNeededPerUnitPerHour > 0:
    durSec = max(durSec, round(amount / bl / 20))
  return durSec

最终时间（有 Boost 时）:
  if BoostMultiplier > 1:
    durSec = durSec / BoostMultiplier
```

### 6.7 产出质量 (service/production.go)

```
resolveQuality(company, reqQuality, input):
  if reqQuality == 0 or input is empty: return reqQuality
  inputQuality = reqQuality - 1
  for each input resource:
    查找 inventory 中 >= inputQuality 的最低品质库存
    如果找不到 → return 0 (无原料)
  return minQualityFound + 1
```

### 6.8 品质销售系数 (未独立使用)

`game.json` → `quality_sales_factor` (默认 0.0833)  
在零售公式中通过 `quality / 12 × RetailQualityWeight` 体现品质影响。

### 6.9 等级与 XP (service/service.go)

```
addXP(company, amount):
  XP += amount
  while XP ≥ XpToNextLevel:
    XP -= XpToNextLevel
    company.Level++
    XpToNextLevel = company.Level × 100      // 每级需要 Level × 100 XP
```

### 6.10 离线收入 (service/offline.go)

```
CalculateOfflineIncome(companyID):
  offlineHours = now - lastActive
  if offlineHours < 0.1: return 0            // <6分钟不算
  if offlineHours > 8: offlineHours = 8      // 最多8小时

  // 生产任务
  for each running production job:
    elapsed = now - job.StartedAt
    completeCycles = floor(elapsed / job.Duration)
    if completeCycles > 0:
      produced = baseQty × completeCycles
      cap at maxCapacity (10000/资源)
      添加到库存
      重置 job 时间线

  // 债券收入
  for each bond where owner == player AND issuer != player:
    daily = floor(amount × 50 × interest × 100)
    bondIncome += daily × (offlineHours / 24)
```

---

## 7. Bot Market Cycle (service/market_competition.go)

### Bot 价格循环

```
cycleVol = 1 + BotCycleAmplitude × sin(hour / 24 × 2π)

basePrice = ComputeChainPrice → processorPrice
buyBase   = ComputeChainPrice → producerPrice

// 买单价格
buyPrice  = round(buyBase × cycleVol × (1 - BotSpread))

// 卖单价格
sellPrice = round(basePrice × cycleVol × (1 + BotSpread))
```

`BotCycleAmplitude` 默认 0.06 (6% 日振幅)  
`BotSpread` 默认 0.05 (5% 买卖价差)

### Bot 数量动态

```
target = BotOrderQty × 50
pressure = (target - currentInventory) / target

buyQty = max(qty/3, qty/2 + rand(qty/2) - max(0, pressure × qty/3))
sellQty = max(qty/3, qty/2 + rand(qty/2) - max(0, -pressure × qty/3))
```

库存偏低时增加买量、减少卖量；反之亦然。

### 玩家取代 Bot

```
replaceBotOrders(resourceID, quality, qty, kind):
  toRemove = ceil(qty × BotReplacementRate)   // 默认 0.3 = 30%
  移除同方向同资源同品质的 bot 订单
```

即玩家每挂单，系统移除同等方向 30% 的机器人订单，给玩家让路。

---

## 8. Market Lock & National Team (service/market_competition.go)

### 市场锁定条件（三个条件任一触发）

```
sellRatio  = 当前卖单总量 / 昨日成交量
buyRatio   = 当前买单总量 / 昨日成交量
threshold  = MarketLockThreshold (默认 0.05)

// 触发条件:
// 1. 卖盘枯竭 — 无货可买
  sellRatio < threshold → 锁定市场（价格上限）
// 2. 买盘枯竭 — 无人收购
  buyRatio < threshold  → 锁定市场（价格下限）
// 3. 价格跌破昨日低点 90%
  lastPrice < dailyLow × 0.9 → 锁定市场
```

### 国家队干预

```
deployNationalTeam(resourceID):
  avgVol = (昨日成交量 + 今日成交量) / 2
  volume = ceil(avgVol × NationalTeamVolumePct)  // 默认 0.3

  // 供给端 — 价格上限
  ntPrice   = highPrice × NationalTeamPricePct   // 默认 1.5（昨日高价 × 1.5）
  sellQty   = volume / 2
  // 需求端 — 价格下限
  floorPrice = highPrice × 0.80
  buyQty     = volume / 2

  国家队以限价挂单，平抑市场
```

---

## 9. Chain Pricing (service/market_info.go)

### 链上价格计算

```
ComputeChainPrice(resourceID):
  cost  = from economy_model (或 fallback: BotOrderBase + resourceID % 7)
  wages = modeledStoreWages / modeledUnitsSoldAnHour

  // 供需压力修正（±20%）
  cost = cost × (1 + MarketPressure[resourceID] × 0.2)

  // 终端价 = 成本 + 30% 利润 + 单位工资
  terminal = cost × 1.30 + wagesPerUnit

  // 三层利润分配（生产/加工/零售）
  gross = terminal - cost - wagesPerUnit
  variance = 0.85 + rand() × 0.30
  baseShare = gross / 3 × variance

  producerPrice  = cost + baseShare
  processorPrice = cost + baseShare × 2 + wages × 0.5
  retailerPrice  = cost + baseShare × 3 + wages
```

价格由经济模型数据驱动，10-15% 随机波动防止静态。

---

## 10. Daily Orders (service/order.go)

### 订单奖励计算

```
computeOrderReward(resourceID, rng):
  basePrice = ComputeChainPrice → processorPrice

  // 根据价格分档确定 tier
  tier:
    basePrice > 50  → 0.5
    basePrice > 20  → 0.75
    basePrice > 10  → 1.0
    basePrice > 3   → 2.0
    default         → 4.0

  qty   = 50 × tier × (0.5 + rand())
  rewardMult = 1.2 + rand() × 0.6

  cash  = round(basePrice × qty × rewardMult)
  xp    = round(DailyOrderXPBase × tier × (0.8 + rand() × 0.4))
         min 5 XP

  // 25% 概率获得品质奖励
  if rand() < 0.25:
    quality = 1 + rand()(min(MaxQuality, 5))
```

### 每日刷新

Scheduler 每 60 秒检查日期变更，自动生成 `DailyOrderCount` (默认 5) 个新订单。

---

## 11. Research (service/research.go)

### 研究项目

当前写死了 4 个项目：

| 项目 | 产出/小时 | 费用 | 时长 |
|------|----------|------|------|
| Plant Research | 12 | 95 | 硬编码 |
| Energy Research | 11 | 150 | 硬编码 |
| Mining Research | 10 | 150 | 硬编码 |
| Chemical Research | 10 | 180 | 锁定 |

### 开始研究

```
StartResearch:
  // 检查资源扣库存
  for each (resourceID, qty) in ResourceCost:
    从 inventory 扣除
  // 检查资金
  if Money < CashCost: 失败
  Money -= CashCost
  // 设置时间线
  Status = "in_progress"
  StartedAt = now
  CompletesAt = now + DurationHours hours
```

### 研究进度

```
ResearchProgress:
  for each in_progress project:
    if now ≥ CompletesAt:
      Progress = 100, Status = "completed"
    else:
      Progress = (elapsed / duration) × 100
```

---

## 12. 政府合同 (service/government.go)

### 竞标价格

玩家对 `GovContract` 出价，系统记录 `bids[]`。

### 授标

```
AwardGovernmentContracts:
  for each contract in bidding phase:
    选择最低价竞标者
    Status = "awarded" (或 contract 过期则标记 "expired")
```

### 交付

```
DeliverGovernmentContract:
  从公司库存扣除 contract 要求的资源 + 数量
  按竞标价 × 数量 + 部分预付结算
  Status = "delivered"
```

### 违约处理

```
ResolveGovernmentDefaults:
  过期未交付 → 标记 "defaulted"
  GovBidRefundRate (默认 0.8) → 退还 80% 已质押资金
```

---

## 13. Game Config 参数表 (`game.json`)

| 参数 | 默认值 | 用途 |
|------|--------|------|
| `start_money` | 200,000 | 新公司初始资金 |
| `start_level` | 42 | 新公司初始等级 |
| `exchange_fee_pct` | 0.04 | 市场交易费率 (4%) |
| `admin_overhead_base` | 1.35 | 基础管理费系数 |
| `bond_face_value` | 5,000 | 每张债券面值 |
| `bond_min_interest` | 0.5 | 债券最低年利率 (0.5%) |
| `bond_max_interest` | 2.0 | 债券最高年利率 (2.0%) |
| `max_bot_orders` | 600 | 机器人最大订单数 |
| `max_ledger_entries` | 5,000 | 最大流水条目数 |
| `weather_speed_mult` | 1.06 | 天气速度修正 |
| `production_mod` | 1.02 | 生产修正系数 |
| `gov_bid_refund_rate` | 0.8 | 政府合同竞标退款比例 |
| `bot_cycle_amplitude` | 0.06 | 机器人价格日振幅 (±6%) |
| `bot_spread` | 0.05 | 机器人买卖价差 (5%) |
| `bot_order_qty` | 200 | 机器人单次挂单基数 |
| `bot_resources` | (23 个 ID) | 机器人参与的资源列表 |
| `bot_order_base` | 8.0 | 机器人基础价格 |
| `base_building_cost` | 50,000 | 基础建筑价格 |
| `warehouse_base_cap` | 1,000 | 仓库基础容量 |
| `warehouse_upgrade_cost` | 25,000 | 仓库升级单价基数 |
| `max_quality` | 100 | 最高品质等级 |
| `quality_sales_factor` | 0.0833 | 品质对销量影响系数 |
| `quality_research_cost` | 5,000 | 品质研发费用 |
| `daily_order_count` | 5 | 每日订单数量 |
| `daily_order_reward_base` | 1,000 | 每日订单基础奖励 |
| `daily_order_xp_base` | 50 | 每日订单基础 XP |
| `base_production_slots` | 3 | 基础生产槽位数 |
| `slot_upgrade_cost` | 50,000 | 槽位升级单价基数 |
| `market_lock_threshold` | 0.05 | 市场锁定阈值 (5%) |
| `market_lock_cap_pct` | 1.2 | 市场锁定价格上限比例 |
| `national_team_volume_pct` | 0.3 | 国家队干预成交量比例 |
| `national_team_price_pct` | 1.5 | 国家队干预价格比例 |
| `bot_replacement_rate` | 0.3 | 玩家替代机器人订单比例 |
