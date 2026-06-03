# Economic Formula Diff Analysis

> Generated 2026-06-03 by comparing `Food Economy Actuarial Design v1.3.1` / `Model v1.3.2` spec against live code.

## 1. Production Formula

### Spec (v1.3.1)
```
OutputPerHour_Lv1         = BuildingBaseOutputPerHour_Lv1 * (1 + FinalProductionSpeedBonus)
OutputPerHour_Level       = OutputPerHour_Lv1 * BuildingLevel
```
No WorkerCount, RecipeWorkload, BaseOutputPerWorkerHour.

### Code (`formula/production.go`)
```go
func BaseProductionRate(producedPerHourRaw, salaryModifier float64, salaryLevel int) float64 {
    sMid := SalaryMid[salaryLevel]
    return producedPerHourRaw * math.Pow(AverageSalary/sMid, salaryModifier)
}
func ProducedPerHour(size int, baseRate, salaryPercent float64, robotCount int,
    isAccumulator bool, speedModifierPct float64, qualityPct float64, isMining bool) float64 { ... }
```

### Gaps
| # | Variable | Status | Action |
|---|----------|--------|--------|
| 1 | `SalaryMid{0:655,1:700,2:745}` | Exists, should be deleted | **DELETE** |
| 2 | `AverageSalary = 345` | Exists, should be deleted | **DELETE** |
| 3 | `RobotBonus = 4` | Exists, should be deleted | **DELETE** |
| 4 | `salaryModifier` param | Exists, should be deleted | **DELETE** |
| 5 | `salaryLevel` param | Exists, should be deleted | **DELETE** |
| 6 | `robotCount` param | Exists, should be deleted | **DELETE** |
| 7 | `isAccumulator` flag | Exists, should be deleted | **DELETE** |
| 8 | `isMining` flag | Exists, should be deleted | **DELETE** |
| 9 | `salaryPercent` param | Exists, should be deleted | **DELETE** |
| 10 | `BuildingBaseOutputPerHour_Lv1` | Missing — no concept | **ADD** |
| 11 | `FinalProductionSpeedBonus` formula | Missing — speed mod exists in `ProducedPerHour` but mixed with salary | **ADD** |
| 12 | Linear level multiplier | Missing — `ProducedPerHour` divides by `(1-salaryPercent/100)` which is wrong | **ADD** |
| 13 | `ProductionTimeSeconds` still uses 345 magic constant | Exists | **REPLACE** |

---

## 2. Cost Indices

### Spec
```
LaborCostPerHour_Level   = BaseLaborCostPerHour_Lv1 * LaborCostIndex * BuildingLevel
InputCostPerHour_Level   = OutputPerHour_Level * InputQtyPerUnit * InputUnitPrice * MaterialCostIndex
EnergyCostPerHour_Level  = BaseEnergyCostPerHour_Lv1 * EnergyCostIndex * BuildingLevel
```

### Code
- No `LaborCostIndex`, `MaterialCostIndex`, `EnergyCostIndex` anywhere in `config/game.json` or `config.go`
- Input costs happen via `deductInputs` (inventory subtraction) — no cost index adjustment
- No labor/energy cost calculation at all
- `production_mod` and `bot_order_base` exist but are not cost indices

### Gaps
| # | Variable | Status |
|---|----------|--------|
| 14 | `LaborCostIndex` | **MISSING** — add to GameConfig |
| 15 | `MaterialCostIndex` | **MISSING** — add to GameConfig |
| 16 | `EnergyCostIndex` | **MISSING** — add to GameConfig |
| 17 | `BaseLaborCostPerHour_Lv1` | **MISSING** — no formula |
| 18 | `BaseEnergyCostPerHour_Lv1` | **MISSING** — no formula |
| 19 | Cost-index-adjusted pricing in trade | **MISSING** — no application point |

---

## 3. Saturation / Price Control

### Spec
```
MarketSaturationGroup     = TotalSupplyInGroup / TotalDemandInGroup
SaturationPriceMultiplier = CLAMP(0.70, 1.10, 1 - MAX(0, MarketSaturationGroup - 1) * SaturationK)
EffectivePrice            = BasePrice * SaturationPriceMultiplier * EventPriceMultiplier
```
Commodity groups: Grain(1), Dairy(2), Processed(3), Bakery(4), GeneralMarket(5), CafeDessert(6), StreetFood(7), RestaurantMeal(8), Finance(9)

### Code
- `MarketPressure[resourceID]` per-resource float (0-1), not per-group
- Pressure applied via `cost * (1 + pressure*0.2)` in `ComputeChainPrice` — ±20% on cost, not a price cap
- `CheckMarketLock` detects extreme sell/buy ratio and locks market + deploys national team — circuit breaker, not saturation curve
- `ResourcesWithMarket()` returns `{8,9,10,11,12}` — only 5 resources get market-lock checks

### Gaps
| # | Variable | Status |
|---|----------|--------|
| 20 | `MarketSaturationGroup` | **MISSING** — no group system |
| 21 | `SaturationK` config param | **MISSING** |
| 22 | `SaturationPriceMultiplier` formula | **MISSING** |
| 23 | `EffectivePrice = BasePrice * SaturationMultiplier` | **MISSING** — price never adjusted for saturation |
| 24 | `EventPriceMultiplier` | **MISSING** — no event system |
| 25 | Commodity group → resource mapping | **MISSING** |

---

## 4. Executive + Reputation

### Spec
```
ExecutiveBonus + GlobalReputationBonus  (additive, not multiplicative)
```

### Code
- `AdminOverheadWithCOO(adminOverhead, cooSkill)` — additive reduction ✅
- `CTOProductionMultiplier(ctoSkill)` — `(100 + ctoSkill*2)/100` = `1 + 0.02*ctoSkill` — additive to base ✅
- No `GlobalReputationBonus` concept exists ✅ (trivial to add, but not present)

### Gaps
| # | Variable | Status |
|---|----------|--------|
| 26 | Production/sales bonus applied from executives | **NOT IMPLEMENTED** — execs are stubs |
| 27 | `CEO`/`CFO` executive roles | Missing from stubs |
| 28 | `ExecutiveCost` in per-hour cost calculation | **MISSING** |
| 29 | `GlobalReputationBonus` | Simple to add when reputation system exists |

---

## 5. Auto Trading Bots

### Spec
```
BotBidPrice = FairPrice * (1 - BotSpread)
BotAskPrice = FairPrice * (1 + BotSpread)
```
- Spread-limited
- Inventory limited (can't sell what they don't have)
- Budget limited (can't buy what they can't afford)
- Participation limited (not every market, not infinite orders)

### Code
| Check | Status |
|-------|--------|
| `bot_spread: 0.05` → bid = FairPrice * 0.95, ask = FairPrice * 1.05 | ✅ Correct |
| Bot inventory: starts at config, resets if < BotOrderQty*25 to BotOrderQty*80 | ⚠️ Replenishes aggressively — almost unlimited inventory |
| Bot budget: resets to BotMoney (5M) if < BotMoney/2 | ⚠️ Effectively unlimited budget |
| `MaxBotOrders: 600` total across all resources & bots | ✅ Hard cap prevents infinite orders |
| `BotReplacementRate: 0.3` — player orders replace bot orders | ✅ Anti-arbitrage |
| Bot order qty varies by inventory pressure | ✅ Prevents over-accumulation |
| `ResourcesWithMarket()` only returns `{8,9,10,11,12}` → scheduler only creates bot orders for these 5 | ⚠️ Meanwhile `bot_resources` has 24 resource IDs — discrepancies |
| `ComputeChainPrice` generates fair price from 3-tier profit model, not `BasePrice * SaturationMultiplier` | ⚠️ Because saturation doesn't exist yet |

---

## 6. Management Cost Convergence

### Spec
- After sweet-spot level (7), management costs accelerate.
- `TotalBuildingCost(Level) = BaseBuildCost * Level * (Level + 1) / 2`

### Code
- `admin_overhead_base: 1.35` — single scalar, no level scaling
- No `BaseManagementCost` formula
- Building upgrade costs: `UpgradeCost = BaseBuildCost * level` (linear) — spec wants the cumulative sum formula

### Gaps
| # | Variable | Status |
|---|----------|--------|
| 30 | Level-based management cost formula | **MISSING** |
| 31 | `BaseManagementCost` per hour | **MISSING** |
| 32 | Sweet-spot level concept (7) | **MISSING** |
| 33 | Cumulative upgrade cost formula | `UpgradeCost = baseCost * level` is simple; spec says cumulative is `baseCost * level * (level+1)/2` |

---

## 7. Bond 50x Coefficient

### Code (`formula/bonds.go`)
```go
func DailyBondInterest(amount int, interestRatePct float64) float64 {
    return math.Floor(float64(amount) * 50.0 * interestRatePct)
}
```

### Impact
At `bond_face_value: 5000`, `bond_min_interest: 0.5` (0.5%):
- `5000 * 50 * 0.005 = 1250/day` = 25%/day = **9125%/yr**

At `bond_max_interest: 2.0` (2%):
- `5000 * 50 * 0.02 = 5000/day` = 100%/day

This is clearly broken. The 50x multiplier was likely a placeholder or unit conversion error.

**Recommended fix:** `amount * interestRatePct / 100.0` → flat rate applied to face value:
`5000 * 0.012 / 1.0 = 60/day` = 1.2%/day = 438%/yr — still high but the interest rates in the config (0.5-2.0) are already daily percentages, so `faceValue * annualRate / 365` would be more appropriate. But since the system uses daily settlement, the simplest fix: `amount * interestRatePct / 100.0` to make `5000 * 0.012 = 60/day` = 1.2%/day reasonable for a bond-like instrument.

Actually wait, looking more carefully:
- `bond_min_interest: 0.5` — is this 0.5% (0.005) or 0.5% as a decimal (0.5)?
- The formula treats it as a fraction: `5000 * 50 * 0.5 = 125000` which is absurd
- So `0.5` probably means 0.5% as a decimal fraction, so interestRatePct = 0.005
- `5000 * 50 * 0.005 = 1250/day` — still absurd

Verdict: The 50x must go. Replace with `amount * interestRatePct / 100.0` or simple `amount * interestRatePct`.

---

## 8. Retail Formula

### Spec
"Do not use global UnitsSold to throttle sales. Use saturation-based pricing instead."

### Code
`UnitsSoldPerHour` in retail.go — 12 parameters, decompiled model from original Sim Business. The spec says to delete `UnitsSold`, `PriceAcceptance`, `RecipeWorkload`.

### Gap
| # | Variable | Status |
|---|----------|--------|
| 34 | `UnitsSoldPerHour` complex formula | Still exists, should be replaced by saturation-driven effective price model |
| 35 | `UnitsSold` as a throttling mechanism | Spec says delete — use saturation pricing instead |
| 36 | `PriceAcceptance` | Not in code, but the retail formula effectively computes it via price-denominator check |

---

## Summary: What to Change

### Must change (breaks core economics)
| Area | Change | Risk of not changing |
|------|--------|---------------------|
| `formula/production.go` | Replace with `OutputPerHour = baseOutput * (1+speedBonus) * level` | Player output curve completely wrong |
| `formula/bonds.go` | Remove 50x coefficient | Bond economy broken (60%/day) |
| `config/game.json` | Add `SaturationK`, cost indices | Missing tuning valves |

### Should change (important but non-breaking)
| Area | Change |
|------|--------|
| New `formula/saturation.go` | `SaturationPriceMultiplier = CLAMP(0.70, 1.10, 1 - MAX(0, ratio-1)*K)` |
| New `formula/costs.go` | `CostPerHour = Base * Index * Level` formulas |
| `service/production.go` | Wire new production formulas |
| `service/market_competition.go` | Track group-level saturation, apply price multiplier |

### Nice to have (formula hygiene)
| Area | Change |
|------|--------|
| `formula/retail.go` | Simplify or mark deprecated |
| `service/market_info.go` | Replace ComputeChainPrice with saturation-adjusted pricing |
| `config/game.json` | Add management cost curve parameters |

---

## Regression Test Plan

```
Test Case: Lv1 Farm producing Grain
  baseOutput     = 500   (per spec v1.3.2 Building Balance table)
  speedBonus     = 0     (no upgrades)
  Output/hr      = 500 * (1 + 0) = 500
  BasePrice      = 23    (per spec)
  Saturation     = 1.0   (balanced market)
  SaturationK    = 0.15
  SaturationMul  = CLAMP(0.70, 1.10, 1 - MAX(0, 1.0 - 1) * 0.15) = 1.0
  EffectivePrice = 23 * 1.0 = 23
  Revenue/hr     = 500 * 23 = 11500
  LaborCost/hr   = LaborBase * LaborIndex * level = 500 * 1.0 * 1 = 500
  MaterialCost/hr = 500 * 0 * 23 (no inputs for base farm) = 0 (depends on recipe)
  EnergyCost/hr  = EnergyBase * EnergyIndex * level = 300 * 1.0 * 1 = 300
  ManagementCost = ManagementBase * level² = 500 * 1 = 500
  TotalCost/hr   = 500 + 0 + 300 + 500 = 1300
  NetProfit/hr   = 11500 - 1300 = 10200/hr

  Target: ~6000/hr
  Gap: 10200 is above 6000 target → adjust BaseLaborCost or ManagementBase to calibrate
  (This is expected — the formulas need per-building tweaking to hit precise targets)
```

Let me proceed to implement changes. Start with the formula files.
