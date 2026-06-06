# Phase 17: Formula and Static Data Governance (v2)

**Date**: 2026-06-06
**Status**: Complete
**Validator**: `backend-next/scripts/phase17-governance.ps1` (12 checks)

## 1. Summary

Phase 17 established a governed formula boundary (`backend-next/internal/formula/`) and static-data provenance (`backend-next/internal/catalog/static-data-manifest.json`) plus legacy source drift detection (`backend-next/internal/formula/legacy-source-manifest.json`). The formula package is the target source of truth for economic calculations; future changes must update governed formulas, parity fixtures, and manifests intentionally.

### Key Deliverables

| Component | Files | Tests |
|-----------|-------|-------|
| Formula package | 5 `.go` files (market, bond, production, costs, saturation) | 18 tests passing |
| Formula wiring | 3 services consume governed functions | 3 parity tests passing |
| Catalog provenance | manifest.go, static-data-manifest.json | 4 tests (load, exists, integrity, required paths) |
| Legacy source manifest | legacy-source-manifest.json, legacy_source_test.go | 1 exact-set and integrity test |
| Validator | phase17-governance.ps1 | 12 checks |

### Formula Inventory

| Family | Functions | Status | Notes |
|--------|-----------|--------|-------|
| Market | ExchangeFee, TickStep, IsValidTick | **Governed** | Exact legacy parity for ExchangeFee |
| Bond | DailyBondInterest, MaxIssuableBonds | **Governed** | faceValue is explicit parameter; guards faceValue<=0 |
| Production | OutputPerHour, DurationSeconds | **Governed** | DurationSeconds preserves BN semantics; differences from legacy documented in tests |
| Costs | LaborCostPerHour, EnergyCostPerHour, InputCost, MaintenanceCostPerHour, ManagementCostPerHour, TaxCost, UpgradeCost, TotalBuildingCost | **Governed but unused** | Available for future core migrations |
| Saturation | SaturationPriceMultiplier, EffectivePrice, GroupOf | **Governed but unused** | GroupOf matches legacy mapping exactly |
| Admin (deferred) | AdminOverheadWithCOO, CTOProductionMultiplier | **Not copied** | Executives plugin -- deferred |
| Retail (deferred) | UnitsSoldPerHour | **Not copied** | Retail plugin -- deferred |
| Dead | PeriodBondInterest | **Not copied** | Alias only; no callers |

### Legacy Saturation Mapping

Exact copy from `backend/internal/formula/saturation.go`:

| Resource ID | Name | Group |
|-------------|------|-------|
| 1 | Wheat | GroupGrain (1) |
| 2 | Old legacy mapping | GroupProcessed (3) |
| 3 | Flour | GroupBakery (4) |
| 4 | Bread | GroupRestaurantMeal (8) |
| Default | Any other | GroupGeneralMarket (5) |

Note: Legacy mapping uses resource ID 2 -> GroupProcessed (3), not GroupDairy. The earlier GroupDairy constant has been removed from `backend-next/internal/formula/saturation.go` to match legacy exactly.

## 2. Governed Formula Boundary

### formula/market.go

```go
func ExchangeFee(amount int, price float64, feeRate float64) float64  // ceil(amount * price * feeRate)
func TickStep(price float64) float64                                   // price quantization
func IsValidTick(price float64) bool                                   // price tick validation
```

ExchangeFee matches legacy formula exactly. The who-pays decision (buyer vs seller) is service orchestration and differs by workflow (buy orders match against sells, sells match against buys).

### formula/bond.go

```go
func DailyBondInterest(amount int, faceValue float64, interestRatePct float64) float64  // floor(amount * faceValue * rate / 100)
func MaxIssuableBonds(totalBuildingValue float64, faceValue float64, alreadySold int) int  // guards faceValue <= 0
```

### formula/production.go

```go
func OutputPerHour(baseOutputPerHour float64, speedBonusPct float64, level int) float64
func DurationSeconds(quantity int, producedPerHourRaw int, level int, productionMod float64) float64
```

DurationSeconds preserves current BN semantics exactly. Differences from legacy documented in test file.

### formula/costs.go

8 cost functions: LaborCostPerHour, EnergyCostPerHour, InputCost, MaintenanceCostPerHour, ManagementCostPerHour, TaxCost, UpgradeCost, TotalBuildingCost. All match legacy formulas. Available for future core migrations.

### formula/saturation.go

3 functions: SaturationPriceMultiplier, EffectivePrice, GroupOf. GroupOf mapping matches legacy exactly.

## 3. Approved Behavioral Differences

| Formula | Legacy Behavior | Backend-Next Behavior | Reason |
|---------|----------------|----------------------|--------|
| DurationSeconds | Uses secondsPerUnit from data | Uses producedPerHourRaw from catalog | Different data source; BN ProductionMod is config divisor |
| ExchangeFee placement | Varies by legacy workflow | Existing backend-next workflows retain their current placement | Phase 17 governs only the pure fee calculation; fee ownership remains an explicit service-level parity question |
| Min duration | secondsPerUnit-based lower bound | Hard 30s minimum | BN explicit guard |

These are approved governed differences -- not bugs. Any future change must update the governed formula AND the fixtures simultaneously.

## 4. Static Data Provenance

### Manifest: `backend-next/internal/catalog/static-data-manifest.json`

Tracks version 1.0.0 SHA-256 digests for 5 static data files:

Digests are calculated after normalizing text line endings to LF so the gate
produces the same result on Windows and Linux.

| File | Purpose |
|------|---------|
| decompiled/data/resources.json | Resource definitions |
| decompiled/data/buildings.json | Building definitions |
| decompiled/data/economy_model.json | Economy model parameters |
| decompiled/data/resource_lookups.json | Resource lookup tables |
| backend/configs/game.json | Game tuning configuration |

### Legacy Source Manifest: `backend-next/internal/formula/legacy-source-manifest.json`

Tracks SHA-256 digests for 5 legacy formula source files that the BN formula package copies from or governs:

| File | Purpose |
|------|---------|
| backend/internal/formula/market.go | ExchangeFee, TickStep, IsValidTick |
| backend/internal/formula/bonds.go | DailyBondInterest, MaxIssuableBonds |
| backend/internal/formula/production.go | OutputPerHour, ProductionDurationSeconds |
| backend/internal/formula/costs.go | All 8 cost functions |
| backend/internal/formula/saturation.go | Saturation, GroupOf |

**Workflow for legacy formula changes**: After modifying any of these legacy source files, compute the new SHA-256, update `legacy-source-manifest.json`, review BN formula parity, and update parity fixtures. The manifest integrity test (`TestLegacySourceManifest`) will fail until the manifest is updated.

Legacy source digests use the same canonical LF normalization as static data.

## 5. Service Wire Map

| Service | Replaced Function | Governed Call | Lines |
|---------|------------------|---------------|-------|
| production/service.go | Inline `math.Ceil(qty/rate*3600)` | `formula.DurationSeconds(...)` | ~239 |
| market/service.go (TakeOrder) | Inline `math.Ceil(fill*price*feeRate)` | `formula.ExchangeFee(...)` | ~567 |
| market/service.go (executeMatchFill) | Inline `math.Ceil(fill*price*feeRate)` | `formula.ExchangeFee(...)` | ~861 |
| finance/service.go (GetBond/ListBonds) | Private `dailyBondInterest()` wrapper | Direct `formula.DailyBondInterest(...)` at DTO site | ~311 |

### Intentionally Not Governed Yet

| Location | Reason |
|----------|--------|
| production/service.go: productionXPForClaim | Stateful -- tracks XPAwarded across claims; not a pure formula |
| All cost/saturation formulas | Not called from any service (no retail/profit computation in BN yet) |
| TickStep/IsValidTick | Available for future price validation wiring |

## 6. Non-Goals

- No economy rebalance
- No tick validation wired (functions available)
- No bond issuance cap wired (function available)
- No boost multiplier in BN production
- No API/response changes
- No scheduler, DB, frontend, or plugin implementation
- No legacy backend edits

## 7. Risks and Mitigations

| Risk | Mitigation |
|------|-----------|
| Formula drift -- service bypasses governed function | Validator checks 3 sites + absence of old wrapper + absence of old inline expressions |
| Static data change without manifest update | TestStaticDataManifestIntegrity fails |
| Legacy source change without parity review | TestLegacySourceManifest fails |
| Future dev copies legacy formula with unknown differences | Documentation in DurationSeconds test and legacy manifest workflow make process explicit |

## 8. Validator Checks

The validator (`backend-next/scripts/phase17-governance.ps1`) runs 12 checks:

1. Formula package builds
2. Formula tests pass (18)
3. Catalog provenance tests pass (8)
4. Service parity tests pass (3)
5. Deferred formula names absent (4 checked)
6. Static-data manifest required paths (5)
7. Static-data manifest hashes match (5)
8. Legacy source manifest required paths (5)
9. Legacy source manifest hashes match (5)
10. Governed formula calls present (production: 1, market: 2, finance: 1)
11. No old private wrapper (dailyBondInterest)
12. No old critical inline fee, bond-interest, or production-duration expressions

## 9. Verification

```powershell
# From backend-next directory:
go test ./internal/formula/           # 18 parity tests
go test ./internal/catalog/           # 8 provenance tests
go test ./internal/app/production/    # +1 duration parity test
go test ./internal/app/market/        # +1 fee parity test
go test ./internal/app/finance/       # +1 bond interest parity test

# From repo root:
powershell -NoProfile backend-next/scripts/phase17-governance.ps1
```

## 10. Completion Evidence

| Metric | Value |
|--------|-------|
| Formula functions | 18 (7 market/bond/production + 8 costs + 3 saturation) |
| Deferred formulas NOT copied | 4 (AdminOverheadWithCOO, CTOProductionMultiplier, UnitsSoldPerHour, PeriodBondInterest) |
| Formula parity tests | 18 |
| Service parity fixtures | 3 |
| Static data files tracked | 5 |
| Legacy source files tracked | 5 |
| Validator checks | 12 |
| gofmt compliance | All files pass |
| Non-ASCII in changed files | 0 |

## 11. Rollback

If governance introduces unexpected behavior:
1. Revert Phase 17 commit: `git revert <phase17-commit-hash>`
2. Verify: `go test ./backend-next/...`
3. Legacy remains operational throughout -- legacy `backend/internal/formula/` is untouched.
4. After revert, investigate the issue, update formula files, run tests, and re-commit.

## 12. Future Phases

- **Phase 18**: Wire cost formulas into production P&L computation
- **Phase 19**: Wire TickStep/IsValidTick into market order creation
- **Phase 20+**: Wire saturation/retail formulas when retail domain is implemented
- **Ongoing**: Any economy change must update governed formula AND update manifest SHA-256 AND update parity fixtures
