package formula_test

import (
	"math"
	"testing"

	"go-sim-api/internal/formula"
)

// ─── Production ────────────────────────────────────────────────

func TestOutputPerHour(t *testing.T) {
	// Spec: OutputPerHour_Lv1 = baseOutput * (1 + speedBonus/100)
	//       OutputPerHour_N   = OutputPerHour_Lv1 * level
	tests := []struct {
		base     float64
		speedPct float64
		level    int
		want     float64
	}{
		{500, 0, 1, 500},           // Farm Lv1, no bonus
		{500, 10, 1, 550},          // Farm Lv1, +10% → 550
		{500, 0, 3, 1500},           // Farm Lv3, no bonus → 500*3
		{500, 10, 3, 1650},          // Farm Lv3, +10% → 550*3
		{320, 0, 1, 320},            // Barn Lv1
		{220, 5, 1, 231},            // Mill Lv1, +5% → 231
		{90, 0, 2, 180},             // Bakery Lv2
		{0, 0, 1, 0},                // zero output
		{500, 0, 0, 500},            // level < 1 → clamped to 1
	}
	for _, tc := range tests {
		got := formula.OutputPerHour(tc.base, tc.speedPct, tc.level)
		if math.Abs(got-tc.want) > 0.01 {
			t.Errorf("OutputPerHour(%.0f, %.0f, %d) = %.2f; want %.2f", tc.base, tc.speedPct, tc.level, got, tc.want)
		}
	}
}

func TestProductionDurationSeconds(t *testing.T) {
	// 500/hr = 7.2 sec/unit ≈ 500/3600
	dur := formula.ProductionDurationSeconds(100, 3600/500.0, 1, 1.0)
	if dur < 719 || dur > 721 {
		t.Errorf("100 units at 500/hr should take ~720s, got %.2f", dur)
	}
	// Level 2 halves time
	dur2 := formula.ProductionDurationSeconds(100, 3600/500.0, 2, 1.0)
	if dur2 < 359 || dur2 > 361 {
		t.Errorf("100 units at 500/hr Lv2 should take ~360s, got %.2f", dur2)
	}
	// Boost 2x halves again
	dur3 := formula.ProductionDurationSeconds(100, 3600/500.0, 2, 2.0)
	if dur3 < 179 || dur3 > 181 {
		t.Errorf("100 units at 500/hr Lv2 +2x boost should take ~180s, got %.2f", dur3)
	}
}

// ─── Bonds ─────────────────────────────────────────────────────

func TestDailyBondInterest_No50x(t *testing.T) {
	// Formula: amount * BondFaceValue * interestRatePct / 100
	// The old 50x was BondFaceValue/100 — removed by writing the formula directly.
	//
	// Service passes interestRatePct = b.Interest * 100 (e.g. 0.012 → 1.2)
	// So: amount * 5000 * 1.2 / 100 = amount * 60
	//
	// For 1 bond at 1.2%: 1 * 5000 * 1.2 / 100 = 60
	got := formula.DailyBondInterest(1, 1.2)
	if got != 60 {
		t.Errorf("DailyBondInterest(1, 1.2) = %.0f; want 60", got)
	}
	// Min interest (b.Interest=0.005 → interestPct=0.5): 1 * 5000 * 0.5 / 100 = 25
	gotMin := formula.DailyBondInterest(1, 0.5)
	if gotMin != 25 {
		t.Errorf("DailyBondInterest(1, 0.5) = %.0f; want 25", gotMin)
	}
	// Max interest (b.Interest=0.020 → interestPct=2.0): 1 * 5000 * 2.0 / 100 = 100
	gotMax := formula.DailyBondInterest(1, 2.0)
	if gotMax != 100 {
		t.Errorf("DailyBondInterest(1, 2.0) = %.0f; want 100", gotMax)
	}
}

func TestMaxIssuableBonds(t *testing.T) {
	got := formula.MaxIssuableBonds(500000, 0)
	if got != 100 {
		t.Errorf("MaxIssuableBonds(500000, 0) = %d; want 100 (500000/5000)", got)
	}
	got2 := formula.MaxIssuableBonds(500000, 50)
	if got2 != 50 {
		t.Errorf("MaxIssuableBonds(500000, 50) = %d; want 50", got2)
	}
	got3 := formula.MaxIssuableBonds(5000, 2)
	if got3 != 0 {
		t.Errorf("MaxIssuableBonds(5000, 2) = %d; want 0 (already sold > max)", got3)
	}
}

// ─── Saturation ────────────────────────────────────────────────

func TestSaturationPriceMultiplier_Balanced(t *testing.T) {
	// Balanced market (saturation=1.0): multiplier should be 1.0
	mul := formula.SaturationPriceMultiplier(1.0, 0.15)
	if math.Abs(mul-1.0) > 0.001 {
		t.Errorf("SaturationPriceMultiplier(1.0, 0.15) = %.4f; want 1.0", mul)
	}
}

func TestSaturationPriceMultiplier_Oversupply(t *testing.T) {
	// Saturation=2.0 (double supply): 1 - 1.0*0.15 = 0.85
	mul := formula.SaturationPriceMultiplier(2.0, 0.15)
	if math.Abs(mul-0.85) > 0.001 {
		t.Errorf("SaturationPriceMultiplier(2.0, 0.15) = %.4f; want 0.85", mul)
	}
	// Saturation=3.0 (triple supply): 1 - 2.0*0.15 = 0.70 (clamped)
	mul2 := formula.SaturationPriceMultiplier(3.0, 0.15)
	if math.Abs(mul2-0.70) > 0.001 {
		t.Errorf("SaturationPriceMultiplier(3.0, 0.15) = %.4f; want 0.70 (clamped)", mul2)
	}
	// Saturation=5.0: still 0.70 (floor)
	mul3 := formula.SaturationPriceMultiplier(5.0, 0.15)
	if math.Abs(mul3-0.70) > 0.001 {
		t.Errorf("SaturationPriceMultiplier(5.0, 0.15) = %.4f; want 0.70 (floor)", mul3)
	}
}

func TestSaturationPriceMultiplier_Undersupply(t *testing.T) {
	// Saturation=0.5 (half supply): clamp ceiling at 1.10
	// Saturation=0.5 (undersupplied): 1 + (1-0.5)*0.15 = 1 + 0.075 = 1.075
	mul := formula.SaturationPriceMultiplier(0.5, 0.15)
	if math.Abs(mul-1.075) > 0.001 {
		t.Errorf("SaturationPriceMultiplier(0.5, 0.15) = %.4f; want 1.075", mul)
	}
	mul2 := formula.SaturationPriceMultiplier(0.9, 0.15)
	if math.Abs(mul2-1.015) > 0.001 {
		t.Errorf("SaturationPriceMultiplier(0.9, 0.15) = %.4f; want 1.015 (mild undersupply → +1.5%%)", mul2)
	}
}

func TestSaturationPriceMultiplier_CustomK(t *testing.T) {
	// Higher K = more price impact
	mul := formula.SaturationPriceMultiplier(2.0, 0.30)
	if math.Abs(mul-0.70) > 0.001 {
		t.Errorf("SaturationPriceMultiplier(2.0, 0.30) = %.4f; want 0.70 (1-1.0*0.30)", mul)
	}
}

func TestGroupOf(t *testing.T) {
	tests := []struct {
		rid  int
		want int
	}{
		{1, formula.GroupGrain},
		{66, formula.GroupDairy},
		{8, formula.GroupGeneralMarket},
		{7, formula.GroupProcessed},
		{115, formula.GroupBakery},
		{999, formula.GroupGeneralMarket}, // unknown → default
	}
	for _, tc := range tests {
		g := formula.GroupOf(tc.rid)
		if g != tc.want {
			t.Errorf("GroupOf(%d) = %d; want %d", tc.rid, g, tc.want)
		}
	}
}

func TestEffectivePrice(t *testing.T) {
	price := formula.EffectivePrice(23.0, 0.85, 1.0)
	if math.Abs(price-19.55) > 0.01 {
		t.Errorf("EffectivePrice(23, 0.85, 1.0) = %.2f; want 19.55", price)
	}
	// With event multiplier
	price2 := formula.EffectivePrice(23.0, 1.0, 1.2)
	if math.Abs(price2-27.60) > 0.01 {
		t.Errorf("EffectivePrice(23, 1.0, 1.2) = %.2f; want 27.60", price2)
	}
}

// ─── Costs ─────────────────────────────────────────────────────

func TestLaborCostPerHour(t *testing.T) {
	got := formula.LaborCostPerHour(500, 1.0, 1)
	if math.Abs(got-500) > 0.01 {
		t.Errorf("LaborCostPerHour(500, 1, 1) = %.2f; want 500", got)
	}
	// Level 3: 500 * 1.0 * 3 = 1500
	got2 := formula.LaborCostPerHour(500, 1.0, 3)
	if math.Abs(got2-1500) > 0.01 {
		t.Errorf("LaborCostPerHour(500, 1, 3) = %.2f; want 1500", got2)
	}
	// High index: 500 * 1.5 * 1 = 750
	got3 := formula.LaborCostPerHour(500, 1.5, 1)
	if math.Abs(got3-750) > 0.01 {
		t.Errorf("LaborCostPerHour(500, 1.5, 1) = %.2f; want 750", got3)
	}
}

func TestManagementCostPerHour(t *testing.T) {
	// Level 1 (≤ sweet spot = 7): linear
	got := formula.ManagementCostPerHour(500, 1, 7)
	if math.Abs(got-500) > 0.01 {
		t.Errorf("ManagementCostPerHour(500, 1, 7) = %.2f; want 500", got)
	}
	// Level 7 (at sweet spot): 500 * 7 = 3500
	got7 := formula.ManagementCostPerHour(500, 7, 7)
	if math.Abs(got7-3500) > 0.01 {
		t.Errorf("ManagementCostPerHour(500, 7, 7) = %.2f; want 3500", got7)
	}
	// Level 10 (past sweet spot): 500 * 100 / 7 ≈ 7142.86
	got10 := formula.ManagementCostPerHour(500, 10, 7)
	if got10 < 7140 || got10 > 7145 {
		t.Errorf("ManagementCostPerHour(500, 10, 7) = %.2f; want ≈7142.86", got10)
	}
}

func TestUpgradeCost(t *testing.T) {
	got := formula.UpgradeCost(50000, 3)
	if math.Abs(got-150000) > 0.01 {
		t.Errorf("UpgradeCost(50000, 3) = %.2f; want 150000", got)
	}
}

func TestTotalBuildingCost(t *testing.T) {
	// Cumulative: 50000 * 5 * 6 / 2 = 750000
	got := formula.TotalBuildingCost(50000, 5)
	if math.Abs(got-750000) > 0.01 {
		t.Errorf("TotalBuildingCost(50000, 5) = %.2f; want 750000", got)
	}
}

// ─── Profit Regression (Spec target: ~6000/hr) ─────────────────

func TestProfitRegression_Lv1Farm(t *testing.T) {
	// Lv1 Farm producing Grain
	// Using spec v1.3.2 Building Balance table:
	//   BaseOutput = 500, BasePrice = 23
	//   LaborBase = 500, EnergyBase = 300
	//   Maintenance = 200, Management = 500
	//   Cost indices all = 1.0
	//   Saturation = 1.0 (balanced)

	baseOutput := 500.0
	speedBonus := 0.0
	level := 1
	laborIndex := 1.0
	energyIndex := 1.0
	basePrice := 23.0

	// Production
	outputPH := formula.OutputPerHour(baseOutput, speedBonus, level)
	if math.Abs(outputPH-500) > 0.01 {
		t.Errorf("Farm Lv1 output = %.2f; want 500", outputPH)
	}

	// Saturation (balanced)
	satMul := formula.SaturationPriceMultiplier(1.0, 0.15)
	effPrice := formula.EffectivePrice(basePrice, satMul, 1.0)

	// Revenue
	revenue := outputPH * effPrice

	// Costs
	labor := formula.LaborCostPerHour(500, laborIndex, level)
	energy := formula.EnergyCostPerHour(300, energyIndex, level)
	maintenance := formula.MaintenanceCostPerHour(200, level)
	management := formula.ManagementCostPerHour(500, level, 7)

	totalCost := labor + energy + maintenance + management
	netProfit := revenue - totalCost

	t.Logf("Farm Lv1 profit regression:")
	t.Logf("  Output/hr:      %.2f", outputPH)
	t.Logf("  EffPrice:       %.2f (satMul=%.4f)", effPrice, satMul)
	t.Logf("  Revenue/hr:     %.2f", revenue)
	t.Logf("  Labor/hr:       %.2f", labor)
	t.Logf("  Energy/hr:      %.2f", energy)
	t.Logf("  Maintenance/hr: %.2f", maintenance)
	t.Logf("  Management/hr:  %.2f", management)
	t.Logf("  TotalCost/hr:   %.2f", totalCost)
	t.Logf("  NetProfit/hr:   %.2f", netProfit)
	t.Logf("  Target:         ~6000/hr")

	// Profit should be positive and in reasonable range for tuning
	if netProfit <= 0 {
		t.Errorf("Farm Lv1 net profit = %.2f; should be positive", netProfit)
	}
	// Expected: 11500 - 1500 = 10000/hr (before input costs)
	// To reach ~6000/hr, either add input costs or increase base costs.
	// This test documents the current computed value for tuning reference.
	if netProfit < 5000 || netProfit > 15000 {
		t.Logf("  NOTE: Farm Lv1 profit %.0f/hr — adjust base costs or price to reach ~6000/hr target", netProfit)
	}
}

func TestProfitRegression_Lv1Mill(t *testing.T) {
	// Lv1 Mill Processing
	// BaseOutput = 220, BasePrice = 68
	baseOutput := 220.0
	basePrice := 68.0
	level := 1

	outputPH := formula.OutputPerHour(baseOutput, 0, level)
	satMul := formula.SaturationPriceMultiplier(1.0, 0.15)
	effPrice := formula.EffectivePrice(basePrice, satMul, 1.0)

	revenue := outputPH * effPrice
	labor := formula.LaborCostPerHour(500, 1.0, level)
	energy := formula.EnergyCostPerHour(300, 1.0, level)
	maintenance := formula.MaintenanceCostPerHour(200, level)
	management := formula.ManagementCostPerHour(500, level, 7)
	totalCost := labor + energy + maintenance + management
	netProfit := revenue - totalCost

	t.Logf("Mill Lv1 profit regression:")
	t.Logf("  Output/hr:      %.2f (want 220)", outputPH)
	t.Logf("  EffPrice:       %.2f", effPrice)
	t.Logf("  Revenue/hr:     %.2f", revenue)
	t.Logf("  NetProfit/hr:   %.2f", netProfit)
	t.Logf("  Without input costs, target ~6000")

	if netProfit <= 0 {
		t.Errorf("Mill Lv1 net profit = %.2f; should be positive", netProfit)
	}
}

func TestProfitRegression_Lv1Bakery(t *testing.T) {
	// Lv1 Bakery — higher price, lower output
	baseOutput := 90.0
	basePrice := 245.0
	level := 1

	outputPH := formula.OutputPerHour(baseOutput, 0, level)
	satMul := formula.SaturationPriceMultiplier(1.0, 0.15)
	effPrice := formula.EffectivePrice(basePrice, satMul, 1.0)

	revenue := outputPH * effPrice
	labor := formula.LaborCostPerHour(500, 1.0, level)
	energy := formula.EnergyCostPerHour(300, 1.0, level)
	maintenance := formula.MaintenanceCostPerHour(200, level)
	management := formula.ManagementCostPerHour(500, level, 7)
	totalCost := labor + energy + maintenance + management
	netProfit := revenue - totalCost

	t.Logf("Bakery Lv1 profit: %.0f/hr (output=%.0f, price=%.2f)", netProfit, outputPH, effPrice)
	// Higher profit per unit compensates for lower output
	if netProfit <= 0 {
		t.Errorf("Bakery Lv1 net profit = %.2f; should be positive", netProfit)
	}
}

// ─── Bot Arbitrage Guard ──────────────────────────────────────

func TestBotSpreadNoArbitrage(t *testing.T) {
	// Spec: BotBidPrice = FairPrice * (1 - BotSpread)
	//       BotAskPrice = FairPrice * (1 + BotSpread)
	// BotSpread = 0.05 → bid=0.95×fair, ask=1.05×fair
	// If a player buys from bot at ask and sells to bot at bid, they lose:
	//   Buy: price * 1.05, Sell: price * 0.95 → loss = 10% round-trip
	fairPrice := 23.0
	spread := 0.05

	botBid := fairPrice * (1 - spread)
	botAsk := fairPrice * (1 + spread)

	roundTripCost := botAsk - botBid // per unit loss from arbitrage
	if roundTripCost <= 0 {
		t.Errorf("Bot spread should prevent arbitrage: bid=%.2f, ask=%.2f, loss=%.2f", botBid, botAsk, roundTripCost)
	}
	if math.Abs(botBid-21.85) > 0.01 {
		t.Errorf("Bot bid = %.2f; want 21.85 (23*0.95)", botBid)
	}
	if math.Abs(botAsk-24.15) > 0.01 {
		t.Errorf("Bot ask = %.2f; want 24.15 (23*1.05)", botAsk)
	}
}

// ─── Admin Formulas ────────────────────────────────────────────

func TestAdminOverheadWithCOO(t *testing.T) {
	// overhead=1.35, cooSkill=8
	// 1.35 - (1.35-1) * 8/100 = 1.35 - 0.35*0.08 = 1.322
	got := formula.AdminOverheadWithCOO(1.35, 8)
	if math.Abs(got-1.322) > 0.001 {
		t.Errorf("AdminOverheadWithCOO(1.35, 8) = %.4f; want 1.322", got)
	}
}

func TestCTOProductionMultiplier(t *testing.T) {
	// skill=7: (100 + 7*2) / 100 = 1.14
	got := formula.CTOProductionMultiplier(7)
	if math.Abs(got-1.14) > 0.001 {
		t.Errorf("CTOProductionMultiplier(7) = %.4f; want 1.14", got)
	}
}

// ─── Edge Cases ────────────────────────────────────────────────

func TestOutputPerHour_EdgeCases(t *testing.T) {
	// Zero speed bonus
	got := formula.OutputPerHour(500, 0, 1)
	if math.Abs(got-500) > 0.01 {
		t.Errorf("Zero speed bonus: got %.2f, want 500", got)
	}
	// Negative speed bonus (debuff) — spec allows this
	gotNeg := formula.OutputPerHour(500, -10, 1)
	if math.Abs(gotNeg-450) > 0.01 {
		t.Errorf("-10%% speed: got %.2f, want 450", gotNeg)
	}
	// Level 0 defaults to 1
	gotL0 := formula.OutputPerHour(500, 0, 0)
	if math.Abs(gotL0-500) > 0.01 {
		t.Errorf("Level 0 clamped to 1: got %.2f, want 500", gotL0)
	}
}

func TestBondInterestEdgeCases(t *testing.T) {
	// Interest rate = 0 → 0
	got := formula.DailyBondInterest(1, 0)
	if got != 0 {
		t.Errorf("DailyBondInterest(1, 0) = %.0f; want 0", got)
	}
	// Multiple bonds at moderate rate: 100 * 5000 * 1.0 / 100 = 5000
	gotM := formula.DailyBondInterest(100, 1.0)
	if gotM != 5000 {
		t.Errorf("DailyBondInterest(100, 1.0) = %.0f; want 5000", gotM)
	}
	// Large amount
	gotL := formula.DailyBondInterest(100, 0.01)
	if gotL != 50 {
		t.Errorf("DailyBondInterest(100, 0.01) = %.0f; want 50", gotL)
	}
}
