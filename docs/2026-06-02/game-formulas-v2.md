# Game Formulas (v2)

**Date:** 2026-06-02  
**Source:** `backend/internal/formula/` (pure functions) + embedded formulas in `backend/internal/service/`  
**Status:** Active (all formulas verified against Go 1.25 source)

---

## Formula Index

| # | Formula | Location | Status |
|---|---------|----------|--------|
| 1 | Tick Size & Validation | `formula/market.go` | ✅ Stable |
| 2 | Exchange Fee | `formula/market.go` | ✅ Stable |
| 3 | Base Production Rate | `formula/production.go` | ✅ Stable |
| 4 | Produced Per Hour | `formula/production.go` | ✅ Stable |
| 5 | Production Time | `formula/production.go` | ✅ Stable |
| 6 | Retail Units Sold Per Hour | `formula/retail.go` | ✅ Stable |
| 7 | Admin Overhead (COO) | `formula/admin.go` | ✅ Stable |
| 8 | CTO Production Multiplier | `formula/admin.go` | ✅ Stable |
| 9 | Bond Daily Interest | `formula/bonds.go` | ✅ Stable |
| 10 | Bond Period Interest | `formula/bonds.go` | ✅ Stable |
| 11 | Max Issuable Bonds | `formula/bonds.go` | ✅ Stable |
| 12 | Building Upgrade Cost | `service/building.go` | ✅ Stable |
| 13 | Warehouse Upgrade Cost | `service/building_shop.go` | ✅ Stable |
| 14 | Slot Upgrade Cost | `service/production.go` | ✅ Stable |
| 15 | Production Cancel Refund | `service/production.go` | ✅ Stable |
| 16 | Production Duration | `service/production.go` | ✅ Stable |
| 17 | Output Quality Resolution | `service/production.go` | ✅ Stable |
| 18 | XP & Leveling | `service/service.go` | ✅ Stable |
| 19 | Offline Income | `service/offline.go` | ✅ Stable |
| 20 | Bot Market Cycle Pricing | `service/market_competition.go` | ✅ Stable |
| 21 | Bot Order Sizing | `service/market_competition.go` | ✅ Stable |
| 22 | Bot Replacement | `service/market_competition.go` | ✅ Stable |
| 23 | Market Lock Detection | `service/market_competition.go` | ✅ Stable |
| 24 | National Team Intervention | `service/market_competition.go` | ✅ Stable |
| 25 | Chain Price Model | `service/market_info.go` | ✅ Stable |
| 26 | Daily Order Reward | `service/order.go` | ✅ Stable |

---

## 1. Market Formulas

**File:** `backend/internal/formula/market.go`

### 1.1 Tick Size

```
TickStep(price):
  price >= 20000  -> 500
  price >= 10000  -> 100
  price >=  5000  ->  25
  price >=  1000  ->  10
  price >=   500  ->   5
  price >=   200  ->   2
  price >=   100  ->   1
  price >=    50  ->   0.5
  price >=    20  ->   0.25
  price >=     5  ->   0.1
  price >=     2  ->   0.05
  price >=     1  ->   0.01
  price >=   0.5  ->   0.005
  default          ->   0.001
```

Used for price validation and order entry UX. See `configs/game.json` for market parameters.

### 1.2 Price Validation

```
IsValidTick(price):
  step = TickStep(price)
  return |price / step - round(price / step)| < 1e-9
```

Orders with invalid tick prices are rejected at submission.

### 1.3 Exchange Fee

```
ExchangeFee(amount, price, feeRate) = ceil(amount * price * feeRate)
```

| Parameter | Source | Default |
|-----------|--------|---------|
| `feeRate` | `game.json` -> `exchange_fee_pct` | 0.04 (4%) |

Fee is deducted from the buyer's total settlement. Fee revenue is not credited to any player account.

---

## 2. Production Formulas

**File:** `backend/internal/formula/production.go`

### 2.1 Constants

| Constant | Value | Description |
|----------|-------|-------------|
| `SalaryMid[0]` | 655 | Low salary level benchmark |
| `SalaryMid[1]` | 700 | Medium salary level benchmark |
| `SalaryMid[2]` | 745 | High salary level benchmark |
| `AverageSalary` | 345.0 | Global average salary constant |
| `RobotBonus` | 4.0 | Per-robot salary bonus for accumulator buildings |

### 2.2 Base Production Rate

```
BaseProductionRate(producedPerHourRaw, salaryModifier, salaryLevel) =
  producedPerHourRaw * (AverageSalary / SalaryMid[salaryLevel]) ^ salaryModifier
```

- `salaryModifier` comes from the executive/CEO salary skill (typically 0-100)
- `salaryLevel` 0 = low, 1 = medium, 2 = high
- Higher `salaryLevel` reduces the base rate (more expensive labor)
- Higher `salaryModifier` amplifies the sensitivity to salary level

### 2.3 Produced Per Hour

```
ProducedPerHour(size, baseRate, salaryPercent, robotCount, isAccumulator,
                speedModifierPct, qualityPct, isMining):
  adjusted = baseRate
  if isMining:
    adjusted *= qualityPct / 100
  adjusted *= (speedModifierPct / 100) + 1

  effectiveSalary = salaryPercent
  if isAccumulator:
    effectiveSalary += RobotBonus * robotCount

  den = 1 - effectiveSalary / 100
  if den <= 0.01: den = 0.01

  return size * adjusted / den
```

| Parameter | Source | Notes |
|-----------|--------|-------|
| `size` | Building size level | Scales output linearly |
| `baseRate` | From `BaseProductionRate` | Per-unit raw output rate |
| `salaryPercent` | Company setting | 0-100 |
| `robotCount` | Building robot count | Only applied to accumulators |
| `speedModifierPct` | CTO skill + level bonuses | See CTO Production Multiplier |
| `qualityPct` | Resource quality | Only applied to mining buildings |
| `isMining` | Building type flag | Mining includes quality scaling |

**Key behavior:** The denominator `den` caps effective salary at 99%, preventing division-by-zero. The $1 - salaryPercent/100$ structure means higher salary **increases** production (better-paid workers produce more).

### 2.4 Production Time

```
ProductionTimeSeconds(size, buildingSalaryModifier, producedPerHour,
                      eventMultiplier):
  if producedPerHour <= 0: return +inf
  if eventMultiplier <= 0: eventMultiplier = 1

  return (345 * buildingSalaryModifier * size / producedPerHour) * eventMultiplier
```

- `buildingSalaryModifier` is the building's inherent salary modifier (from `buildings.json`)
- `345` is the `AverageSalary` constant
- `eventMultiplier` is > 1 for event-based production boosts (Default: 1.0 via `production_mod` from `game.json`)
- Returns `+inf` when `producedPerHour <= 0` (production cannot complete)

**Relationship:** Production time is inversely proportional to `ProducedPerHour`. The `345 * buildingSalaryModifier * size` numerator converts output units to time.

---

## 3. Retail Formula

**File:** `backend/internal/formula/retail.go`

### 3.1 Units Sold Per Hour

```
Constants:
  RetailQualityWeight = 0.3
  RetailZor           = 370.0

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
  d = clamp(2 - saturation, 0, 2)          // demand coefficient [0, 2]
  p = max(0.9, d/2 + 0.5)                 // demand modifier
  f = quality / 12                        // quality factor
  g = RetailZor
      * (buildingLevelsNeededPerUnitPerHour * modeledUnitsSoldAnHour + 1)
      * buildingKindModifier
      * (d/2 * (1 + f * RetailQualityWeight))
  m = modeledUnitsSoldAnHour * p
  if m <= 0: return 0

  den = price - modeledProductionCostPerUnit
  if den <= 0.0001: return 0             // price below cost -> no sales

  a = (modeledStoreWages + g) / (den * den)
  b = g - (m - price)^2 * a
  if b + modeledStoreWages <= 0: return 0

  res = (acceleration * den * 3600 - modeledStoreWages) / (b + modeledStoreWages)
  if size <= 0: size = 1
  if acceleration <= 0: acceleration = 1
  res = res / size / acceleration
  res = res - res * salesModifierPct / 100
  if weatherSellingSpeedMultiplier > 0:
    res = res / weatherSellingSpeedMultiplier
  if NaN or Inf or res < 0: return 0

  return res
```

| Parameter | Source | Notes |
|-----------|--------|-------|
| `buildingKindModifier` | Economy model per building kind | Retail building modifier |
| `buildingLevelsNeededPerUnitPerHour` | Economy model per resource | Production density |
| `modeledProductionCostPerUnit` | Economy model per resource | Baseline cost |
| `modeledStoreWages` | Economy model per resource | Retail wage baseline |
| `modeledUnitsSoldAnHour` | Economy model per resource | Market potential |
| `price` | Player-set selling price | Must exceed cost |
| `quality` | Resource quality (0-100) | Scales with `quality / 12` |
| `saturation` | Market saturation [0, 2] | 0=underserved, 2=glutted |
| `salesModifierPct` | Configurable modifier | From `game.json` or events |
| `size` | Building size | Scales revenue linearly |
| `acceleration` | Boost multiplier | >= 1 when boosted |
| `weatherSellingSpeedMultiplier` | Weather effect | > 0 means slower sales |

**Design note:** The formula produces units/hour of a selling building. It accounts for saturation (too many sellers = less per-seller volume), quality-based premiums, and price competition (lower price relative to cost yields higher volume).

---

## 4. Admin & Executive Formulas

**File:** `backend/internal/formula/admin.go`

### 4.1 Admin Overhead with COO

```
AdminOverheadWithCOO(adminOverhead, cooSkill):
  return adminOverhead - (adminOverhead - 1) * cooSkill / 100
```

| Parameter | Source | Default |
|-----------|--------|---------|
| `adminOverhead` | `game.json` -> `admin_overhead_base` | 1.35 |
| `cooSkill` | COO executive skill (0-100) | 0 |

Reduces overhead proportionally to COO skill. At `cooSkill = 100`, overhead becomes 1.0 (no overhead).

### 4.2 CTO Production Multiplier

```
CTOProductionMultiplier(ctoSkill) = (100 + ctoSkill * 2) / 100
```

Each CTO skill point yields +2% production speed.

---

## 5. Bond Formulas

**File:** `backend/internal/formula/bonds.go`

### 5.1 Constants

| Constant | Overridable | Default |
|----------|-------------|---------|
| `BondFaceValue` | `game.json` -> `bond_face_value` | 5000.0 |

### 5.2 Daily Interest

```
DailyBondInterest(amount, interestRatePct):
  return floor(amount * 50 * interestRatePct)
```

- `amount`: number of bonds held
- `interestRatePct`: annual interest rate (e.g. 1.5 means 1.5%)
- The constant `50` is `BondFaceValue / 100`
- Result is daily interest in dollars (floor-rounded)

### 5.3 Period Interest

```
PeriodBondInterest(amount, interestRatePct):
  return floor(amount * BondFaceValue * interestRatePct) / 100
```

Used for settlement periods longer than one day.

### 5.4 Max Issuable Bonds

```
MaxIssuableBonds(totalBuildingValue, alreadySold):
  return max(0, floor(totalBuildingValue / BondFaceValue) - alreadySold)
```

Total bond issuance cannot exceed the company's building asset value at face value.

---

## 6. Embedded Service Formulas

### 6.1 Building Upgrade Cost

**File:** `service/building.go`

```
UpgradeCost(currLevel, baseCost):
  nextLevel = currLevel + 1
  if baseCost <= 0: baseCost = kind * 5000
  cost = nextLevel * baseCost
  outputMultiplier = nextLevel
```

### 6.2 Building Purchase Prices

**File:** `service/building_shop.go`

| Building | Price |
|----------|-------|
| Farm Plot | 50,000 |
| Food Factory | 120,000 |
| Warehouse | 30,000 |
| Research Lab | 80,000 |

### 6.3 Warehouse Upgrade

**File:** `service/building_shop.go`

```
WarehouseUpgradeCost(lvl) = (lvl + 1) * WarehouseUpgradeCostBase

NewCapacity = (WarehouseLevel + 1) * WarehouseBaseCap
```

| Parameter | Source | Default |
|-----------|--------|---------|
| `WarehouseUpgradeCostBase` | `game.json` -> `warehouse_upgrade_cost` | 25,000 |
| `WarehouseBaseCap` | `game.json` -> `warehouse_base_cap` | 1,000 |

### 6.4 Production Slot Upgrade

**File:** `service/production.go`

```
SlotUpgradeCost(current) = (current + 1) * SlotUpgradeCostBase
```

| Parameter | Source | Default |
|-----------|--------|---------|
| `SlotUpgradeCostBase` | `game.json` -> `slot_upgrade_cost` | 50,000 |
| `BaseProductionSlots` | `game.json` -> `base_production_slots` | 3 |

### 6.5 Cancel Production Refund

**File:** `service/production.go`

```
CancelRefund(inputs):
  for each (resourceID, qty) in inputs:
    refund qty / 2 (floor)
```

Cancelling returns 50% of input resources.

### 6.6 Production Duration

**File:** `service/production.go`

```
calcProductionDuration(resourceID, amount, durSec):
  if durSec <= 0:
    durSec = max(30, amount * 6)
  // If economy model provides per-unit density data:
  if buildingLevelsNeededPerUnitPerHour > 0:
    durSec = max(durSec, round(amount / bl / 20))
  return durSec

// Final time when boosted:
if BoostMultiplier > 1:
  durSec = durSec / BoostMultiplier
```

### 6.7 Output Quality Resolution

**File:** `service/production.go`

```
resolveQuality(company, reqQuality, input):
  if reqQuality == 0 or input is empty: return reqQuality
  inputQuality = reqQuality - 1
  for each input resource:
    find inventory entry >= inputQuality with lowest quality
    if not found -> return 0 (missing input)
  return minQualityFound + 1
```

Quality of output = min input quality + 1. If any required input of sufficient quality is missing, production fails (returns 0).

### 6.8 XP & Leveling

**File:** `service/service.go`

```
addXP(company, amount):
  XP += amount
  while XP >= XpToNextLevel:
    XP -= XpToNextLevel
    company.Level++
    XpToNextLevel = company.Level * 100
```

| Level | XP Required | Cumulative XP | Building Slots |
|-------|-------------|---------------|----------------|
| 1 | 100 | 0 | 1 |
| 2 | 200 | 100 | 1 |
| 5 | 500 | 1,000 | 2 |
| 10 | 1,000 | 4,500 | 3 |
| 42 | 4,200 | 86,100 | 9 |
| 60 | 6,000 | 183,000 | 13 |

Building slots = `1 + floor(level / 5)`.

### 6.9 Offline Income

**File:** `service/offline.go`

```
CalculateOfflineIncome(companyID):
  offlineHours = now - lastActive
  if offlineHours < 0.1: return 0           // <6 min threshold
  if offlineHours > 8: offlineHours = 8     // cap at 8 hours

  // Production jobs
  for each running production job:
    elapsed = now - job.StartedAt
    completeCycles = floor(elapsed / job.Duration)
    if completeCycles > 0:
      produced = baseQty * completeCycles
      cap at maxCapacity (10,000/resource)
      add to inventory
      reset job timeline

  // Bond income
  for each bond where owner == player AND issuer != player:
    daily = floor(amount * 50 * interest * 100)
    bondIncome += daily * (offlineHours / 24)
```

---

## 7. Bot Market Cycle

**File:** `service/market_competition.go`

Scheduler runs hourly. Bot prices follow a sinusoidal cycle to simulate natural market rhythms.

### 7.1 Bot Price Cycle

```
cycleVol = 1 + BotCycleAmplitude * sin(hour / 24 * 2 * pi)

// Base prices from chain model
basePrice = ComputeChainPrice -> processorPrice
buyBase   = ComputeChainPrice -> producerPrice

// Buy order (bid) price
buyPrice  = round(buyBase * cycleVol * (1 - BotSpread))

// Sell order (ask) price
sellPrice = round(basePrice * cycleVol * (1 + BotSpread))
```

| Parameter | Source | Default |
|-----------|--------|---------|
| `BotCycleAmplitude` | `game.json` -> `bot_cycle_amplitude` | 0.06 (6%) |
| `BotSpread` | `game.json` -> `bot_spread` | 0.05 (5%) |

### 7.2 Bot Order Sizing

```
target = BotOrderQty * 50
pressure = (target - currentInventory) / target

buyQty  = max(qty/3, qty/2 + rand(qty/2) - max(0, pressure * qty/3))
sellQty = max(qty/3, qty/2 + rand(qty/2) - max(0, -pressure * qty/3))
```

| Parameter | Source | Default |
|-----------|--------|---------|
| `BotOrderQty` | `game.json` -> `bot_order_qty` | 200 |

**Behavior:** Low inventory -> more buy orders, fewer sell orders. High inventory -> more sell orders, fewer buy orders. Random variance prevents static pricing.

### 7.3 Player Bot Replacement

```
replaceBotOrders(resourceID, quality, qty, kind):
  toRemove = ceil(qty * BotReplacementRate)
  remove matching bot orders (same direction, resource, quality)
```

| Parameter | Source | Default |
|-----------|--------|---------|
| `BotReplacementRate` | `game.json` -> `bot_replacement_rate` | 0.3 (30%) |

When a player places an order, 30% of matching bot orders in the same direction are removed, creating market room for the player.

---

## 8. Market Lock & National Team

**File:** `service/market_competition.go`

### 8.1 Market Lock Conditions

```
sellRatio  = total open sell quantity / (yesterday volume / 100)
buyRatio   = total open buy quantity / (yesterday volume / 100)
threshold  = MarketLockThreshold

Conditions (any triggers lock):
  1. sellRatio < threshold   -> market locked (sell-side dried up)
  2. buyRatio  < threshold   -> market locked (buy-side dried up)
  3. lastPrice < dailyLow * 0.9  -> market locked (price crash)
```

| Parameter | Source | Default |
|-----------|--------|---------|
| `MarketLockThreshold` | `game.json` -> `market_lock_threshold` | 0.05 (5%) |

### 8.2 National Team Intervention

When a market lock is detected, the national team deploys emergency orders:

```
deployNationalTeam(resourceID):
  avgVol = (yesterdayVolume + todayVolume) / 2
  if avgVol <= 0: avgVol = 10000
  volume = ceil(avgVol * NationalTeamVolumePct)

  // Supply cap (sell order)
  ntPrice   = highPrice * NationalTeamPricePct
  sellQty   = volume / 2

  // Price floor (buy order)
  floorPrice = highPrice * 0.80
  buyQty     = volume / 2
```

| Parameter | Source | Default |
|-----------|--------|---------|
| `NationalTeamVolumePct` | `game.json` -> `national_team_volume_pct` | 0.3 (30%) |
| `NationalTeamPricePct` | `game.json` -> `national_team_price_pct` | 1.5 (150%) |

---

## 9. Chain Pricing Model

**File:** `service/market_info.go`

### 9.1 Compute Chain Price

```
ComputeChainPrice(resourceID):
  // Base cost from economy model or fallback
  cost  = economy_model -> modeledProductionCostPerUnit
         fallback: BotOrderBase + resourceID % 7
  wages = economy_model -> modeledStoreWages
  sales = economy_model -> modeledUnitsSoldAnHour

  // Supply/demand pressure (+/- 20%)
  cost = cost * (1 + MarketPressure[resourceID] * 0.2)

  wpu = wages / sales

  // Terminal (retail) price = cost + 30% margin + per-unit wages
  terminal = cost * 1.30 + wpu

  // Gross margin split across 3 tiers
  gross = terminal - cost - wpu
  variance = 0.85 + rand() * 0.30  // +/- 15% random
  baseShare = gross / 3 * variance

  producerPrice  = cost + baseShare
  processorPrice = cost + baseShare * 2 + wpu * 0.5
  retailerPrice  = cost + baseShare * 3 + wpu
```

Returns: `{ terminalPrice, producerPrice, processorPrice, retailerPrice, productionCost, wagesPerUnit, tierProfit }`

| Parameter | Source | Default |
|-----------|--------|---------|
| `BotOrderBase` | `game.json` -> `bot_order_base` | 8.0 |

**Three-tier model:**
- **Producer:** Farmers / miners — lowest share
- **Processor:** Manufacturers — middle share + half wages
- **Retailer:** Stores — full share + full wages
- Prices are rounded to 2 decimal places; random variance prevents static markets

---

## 10. Daily Orders

**File:** `service/order.go`

### 10.1 Order Reward Calculation

```
computeOrderReward(resourceID, rng):
  basePrice = ComputeChainPrice -> processorPrice

  // Price tier determines reward multipliers
  tier:
    basePrice > 50  -> 0.5
    basePrice > 20  -> 0.75
    basePrice > 10  -> 1.0
    basePrice > 3   -> 2.0
    default         -> 4.0

  qty         = 50 * tier * (0.5 + rand())
  rewardMult  = 1.2 + rand() * 0.6
  cash        = round(basePrice * qty * rewardMult)
  xp          = round(DailyOrderXPBase * tier * (0.8 + rand() * 0.4))
                min 5 XP

  // 25% chance for quality reward
  if rand() < 0.25:
    quality = 1 + rand() * min(MaxQuality, 5)
```

### 10.2 Daily Refresh

Scheduler checks every 60 seconds for date change. Generates `DailyOrderCount` (default 5) new orders from available resources.

---

## 11. Research

**File:** `service/research.go`

### 11.1 Research Projects

Currently 4 hardcoded projects:

| Project | Output/Hour | Cost | Duration |
|---------|-------------|------|----------|
| Plant Research | 12 | 95 | Hardcoded |
| Energy Research | 11 | 150 | Hardcoded |
| Mining Research | 10 | 150 | Hardcoded |
| Chemical Research | 10 | 180 | Locked |

### 11.2 Start Research

```
StartResearch:
  for each (resourceID, qty) in ResourceCost:
    deduct from inventory
  if Money < CashCost: fail
  Money -= CashCost
  Status = "in_progress"
  StartedAt = now
  CompletesAt = now + DurationHours hours
```

### 11.3 Research Progress

```
ResearchProgress:
  for each in_progress project:
    if now >= CompletesAt:
      Progress = 100, Status = "completed"
    else:
      Progress = (elapsed / duration) * 100
```

---

## 12. Government Contracts

**File:** `service/government.go`

### 12.1 Bidding

Players submit bids (`price` per unit) on `GovContract`. The system records `bids[]`.

### 12.2 Award

```
AwardGovernmentContracts:
  for each contract in bidding phase:
    select lowest bidder
    Status = "awarded"
    (or mark "expired" if past deadline)
```

### 12.3 Delivery

```
DeliverGovernmentContract:
  deduct contract resources from company inventory
  settle at bid price * quantity + partial prepayment
  Status = "delivered"
```

### 12.4 Default

```
ResolveGovernmentDefaults:
  past-due undelivered -> marked "defaulted"
  GovBidRefundRate (default 0.8) -> 80% of pledged funds refunded
```

---

## 13. Game Config Parameters (`game.json`)

| Parameter | Default | Description |
|-----------|---------|-------------|
| `start_money` | 200,000 | New company starting cash |
| `start_level` | 42 | New company starting level |
| `exchange_fee_pct` | 0.04 | Market transaction fee (4%) |
| `admin_overhead_base` | 1.35 | Base admin overhead multiplier |
| `bond_face_value` | 5,000 | Face value per bond |
| `bond_min_interest` | 0.5 | Minimum annual bond interest (0.5%) |
| `bond_max_interest` | 2.0 | Maximum annual bond interest (2.0%) |
| `max_bot_orders` | 600 | Maximum bot orders in the book |
| `max_ledger_entries` | 5,000 | Maximum financial ledger entries |
| `weather_speed_mult` | 1.06 | Weather speed modifier |
| `production_mod` | 1.02 | Production modifier coefficient |
| `gov_bid_refund_rate` | 0.8 | Government contract bid refund ratio |
| `bot_cycle_amplitude` | 0.06 | Bot price daily amplitude (+/-6%) |
| `bot_spread` | 0.05 | Bot bid-ask spread (5%) |
| `bot_order_qty` | 200 | Bot base order quantity |
| `bot_resources` | (23 IDs) | Resources bots participate in |
| `bot_order_base` | 8.0 | Bot base price |
| `base_building_cost` | 50,000 | Base building purchase price |
| `warehouse_base_cap` | 1,000 | Base warehouse capacity |
| `warehouse_upgrade_cost` | 25,000 | Warehouse upgrade cost per level |
| `max_quality` | 100 | Maximum quality level |
| `quality_sales_factor` | 0.0833 | Quality sales impact factor |
| `quality_research_cost` | 5,000 | Quality research cost |
| `daily_order_count` | 5 | Number of daily orders |
| `daily_order_reward_base` | 1,000 | Base daily order reward |
| `daily_order_xp_base` | 50 | Base daily order XP |
| `base_production_slots` | 3 | Base production slot count |
| `slot_upgrade_cost` | 50,000 | Slot upgrade cost per level |
| `market_lock_threshold` | 0.05 | Market lock threshold (5%) |
| `market_lock_cap_pct` | 1.2 | Market lock price cap ratio |
| `national_team_volume_pct` | 0.3 | National team intervention volume ratio |
| `national_team_price_pct` | 1.5 | National team intervention price ratio |
| `bot_replacement_rate` | 0.3 | Player bot order replacement rate |

All parameters are hot-reloadable via `game.json` (restart required for changes to take effect).

---

## 14. Planned / In-Progress Changes

| Change | Reference | Status |
|--------|-----------|--------|
| Food chain simplification (10 buildings, 19 resources) | `docs/2026-06-02/food-chain-simplify.md` | Planned |
| Restaurant math model (menu, staff, style, rating) | `docs/restaurant-math-model.md` | Design phase |
| Weather dynamic variation | `docs/requirements.md` #weather | Backlog |
| Leaderboard API | `docs/2026-06-02/dev-plan.md` | Backlog |
| Search/filter for market, executives, bonds | `docs/2026-06-02/dev-plan.md` | Backlog |
