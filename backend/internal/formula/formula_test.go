package formula_test

import (
	"math"
	"testing"
	"time"

	"github.com/beiwater/NewHaven/backend/internal/formula"
)

// --- Exchange Fee ---

func TestExchangeFee(t *testing.T) {
	tests := []struct {
		fill    int
		price   float64
		feeRate float64
		want    float64
	}{
		{100, 23.0, 0.04, 92.0},  // ceil(100*23*0.04) = ceil(92) = 92
		{1, 5000.0, 0.04, 200.0}, // ceil(1*5000*0.04) = ceil(200) = 200
		{0, 100.0, 0.04, 0.0},    // ceil(0*100*0.04) = ceil(0) = 0
		{3, 1.99, 0.10, 1.0},     // ceil(3*1.99*0.10) = ceil(0.597) = 1
	}
	for _, tc := range tests {
		got := formula.ExchangeFee(tc.fill, tc.price, tc.feeRate)
		if got != tc.want {
			t.Errorf("ExchangeFee(%d, %.2f, %.2f) = %g; want %g", tc.fill, tc.price, tc.feeRate, got, tc.want)
		}
	}
}

// --- Tick Step ---

func TestTickStep(t *testing.T) {
	cases := []struct {
		price float64
		want  float64
	}{
		{25000, 500},
		{20000, 500}, // boundary
		{19999, 100},
		{10000, 100}, // boundary
		{9999, 25},
		{5000, 25}, // boundary
		{4999, 10},
		{1000, 10}, // boundary
		{999, 5},
		{500, 5}, // boundary
		{499, 2},
		{200, 2}, // boundary
		{199, 1},
		{100, 1}, // boundary
		{99, 0.5},
		{50, 0.5}, // boundary
		{49, 0.25},
		{20, 0.25}, // boundary
		{19, 0.1},
		{5, 0.1}, // boundary
		{4.99, 0.05},
		{2, 0.05}, // boundary
		{1.99, 0.01},
		{1, 0.01}, // boundary
		{0.99, 0.005},
		{0.5, 0.005}, // boundary
		{0.49, 0.001},
		{0.001, 0.001},
	}
	for _, c := range cases {
		got := formula.TickStep(c.price)
		if got != c.want {
			t.Errorf("TickStep(%.4f) = %g; want %g", c.price, got, c.want)
		}
	}
}

// --- Is Valid Tick ---

func TestIsValidTick(t *testing.T) {
	cases := []struct {
		price float64
		want  bool
	}{
		{25.0, true},
		{25.01, false},
		{100.0, true},
		{100.50, false},
		{2000.0, true},
		{1999.0, false},
		{0.5, true},
		{0.513, false},
		{10000.0, true},
		{10250.0, false},
	}
	for _, c := range cases {
		got := formula.IsValidTick(c.price)
		if got != c.want {
			t.Errorf("IsValidTick(%.3f) = %v; want %v", c.price, got, c.want)
		}
	}
}

// --- Bonds ---

func TestDailyBondInterest(t *testing.T) {
	tests := []struct {
		amount  int
		faceVal float64
		ratePct float64
		want    float64
	}{
		{1, 5000, 1.2, math.Floor(1 * 5000 * 1.2 / 100)},       // 60
		{1, 5000, 0.5, math.Floor(1 * 5000 * 0.5 / 100)},       // 25
		{0, 5000, 1.2, math.Floor(0 * 5000 * 1.2 / 100)},       // 0
		{10, 10000, 2.0, math.Floor(10 * 10000 * 2.0 / 100)},   // 2000
		{100, 5000, 0.01, math.Floor(100 * 5000 * 0.01 / 100)}, // 50
	}
	for _, tc := range tests {
		got := formula.DailyBondInterest(tc.amount, tc.faceVal, tc.ratePct)
		if got != tc.want {
			t.Errorf("DailyBondInterest(%d, %.0f, %.1f) = %g; want %g", tc.amount, tc.faceVal, tc.ratePct, got, tc.want)
		}
	}
}

// --- Max Issuable Bonds ---

func TestMaxIssuableBonds(t *testing.T) {
	tests := []struct {
		val  float64
		fv   float64
		sold int
		want int
	}{
		{500000, 5000, 0, 100}, // floor(500000/5000) = 100
		{500000, 5000, 50, 50}, // floor(500000/5000) - 50 = 50
		{5000, 5000, 2, 0},     // floor(5000/5000) - 2 = -1 -> 0
		{0, 5000, 0, 0},        // zero value
		{500000, 0, 0, 0},      // faceValue=0 returns 0
		{500000, -1, 0, 0},     // negative faceValue returns 0
	}
	for _, tc := range tests {
		got := formula.MaxIssuableBonds(tc.val, tc.fv, tc.sold)
		if got != tc.want {
			t.Errorf("MaxIssuableBonds(%.0f, %.0f, %d) = %d; want %d", tc.val, tc.fv, tc.sold, got, tc.want)
		}
	}
}

// --- Output Per Hour ---

func TestOutputPerHour(t *testing.T) {
	tests := []struct {
		base  float64
		bonus float64
		level int
		want  float64
	}{
		{500, 0, 1, 500},  // no bonus
		{500, 10, 1, 550}, // 10% bonus
		{500, 0, 3, 1500}, // level 3
		{500, 0, 0, 500},  // level 0 clamped to 1
	}
	for _, tc := range tests {
		got := formula.OutputPerHour(tc.base, tc.bonus, tc.level)
		if got != tc.want {
			t.Errorf("OutputPerHour(%.0f, %.0f, %d) = %g; want %g", tc.base, tc.bonus, tc.level, got, tc.want)
		}
	}
}

// --- Duration Seconds ---

func TestDurationSeconds(t *testing.T) {
	tests := []struct {
		qty  int
		rate int
		lvl  int
		mod  float64
		want float64
	}{
		{10, 500, 1, 1.0, math.Ceil(10.0 / 500.0 / 1.0 / 1.0 * 3600)}, // 72
		{1, 500, 1, 1.0, 30.0}, // min 30s
		{10, 500, 2, 1.0, math.Ceil(10.0 / 1000.0 / 1.0 * 3600)},        // 36 (level 2)
		{10, 500, 1, 1.02, math.Ceil(10.0 / 500.0 / 1.0 / 1.02 * 3600)}, // 71 (ProductionMod)
		{0, 500, 1, 1.0, 0}, // qty=0
		{10, 0, 1, 1.0, 0},  // rate=0 returns 0
	}
	for _, tc := range tests {
		got := formula.DurationSeconds(tc.qty, tc.rate, tc.lvl, tc.mod)
		if got != tc.want {
			t.Errorf("DurationSeconds(%d, %d, %d, %.2f) = %g; want %g", tc.qty, tc.rate, tc.lvl, tc.mod, got, tc.want)
		}
	}
}

// --- Cost Formulas ---

func TestLaborCostPerHour(t *testing.T) {
	tests := []struct {
		base  float64
		idx   float64
		level int
		want  float64
	}{
		{500, 1.0, 1, 500 * 1.0 * 1},
		{500, 1.0, 5, 500 * 1.0 * 5},
		{500, 1.2, 3, 500 * 1.2 * 3},
		{500, 1.0, 0, 500 * 1.0 * 1}, // level 0 clamped to 1
	}
	for _, tc := range tests {
		got := formula.LaborCostPerHour(tc.base, tc.idx, tc.level)
		if got != tc.want {
			t.Errorf("LaborCostPerHour(%.0f, %.1f, %d) = %.2f; want %.2f", tc.base, tc.idx, tc.level, got, tc.want)
		}
	}
}

func TestEnergyCostPerHour(t *testing.T) {
	tests := []struct {
		base  float64
		idx   float64
		level int
		want  float64
	}{
		{300, 1.0, 1, 300 * 1.0 * 1},
		{300, 1.0, 4, 300 * 1.0 * 4},
		{300, 1.5, 2, 300 * 1.5 * 2},
		{300, 1.0, 0, 300 * 1.0 * 1}, // level 0 clamped to 1
	}
	for _, tc := range tests {
		got := formula.EnergyCostPerHour(tc.base, tc.idx, tc.level)
		if got != tc.want {
			t.Errorf("EnergyCostPerHour(%.0f, %.1f, %d) = %.2f; want %.2f", tc.base, tc.idx, tc.level, got, tc.want)
		}
	}
}

func TestInputCost(t *testing.T) {
	tests := []struct {
		output     float64
		qtyPerUnit float64
		unitPrice  float64
		matIdx     float64
		want       float64
	}{
		{10, 2, 5, 1.0, 10 * 2 * 5 * 1.0}, // normal
		{0, 2, 5, 1.0, 0},                 // zero output
		{10, 0, 5, 1.0, 0},                // zero qty per unit
		{10, 2, 5, 1.5, 10 * 2 * 5 * 1.5}, // with material cost index
	}
	for _, tc := range tests {
		got := formula.InputCost(tc.output, tc.qtyPerUnit, tc.unitPrice, tc.matIdx)
		if got != tc.want {
			t.Errorf("InputCost(%.0f, %.0f, %.0f, %.1f) = %.2f; want %.2f", tc.output, tc.qtyPerUnit, tc.unitPrice, tc.matIdx, got, tc.want)
		}
	}
}

func TestMaintenanceCostPerHour(t *testing.T) {
	tests := []struct {
		base  float64
		level int
		want  float64
	}{
		{200, 1, 200 * 1},
		{200, 5, 200 * 5},
		{200, 0, 200 * 1}, // level 0 clamped to 1
		{200, 10, 200 * 10},
	}
	for _, tc := range tests {
		got := formula.MaintenanceCostPerHour(tc.base, tc.level)
		if got != tc.want {
			t.Errorf("MaintenanceCostPerHour(%.0f, %d) = %.2f; want %.2f", tc.base, tc.level, got, tc.want)
		}
	}
}

func TestManagementCostPerHour(t *testing.T) {
	tests := []struct {
		base  float64
		level int
		sweet int
		want  float64
	}{
		{500, 1, 7, 500 * 1},              // linear (level <= sweet)
		{500, 7, 7, 500 * 7},              // linear at sweet spot
		{500, 10, 7, 500.0 * 100.0 / 7.0}, // quadratic (level > sweet)
		{500, 0, 7, 500 * 1},              // level 0 clamped to 1
	}
	for _, tc := range tests {
		got := formula.ManagementCostPerHour(tc.base, tc.level, tc.sweet)
		if got != tc.want {
			t.Errorf("ManagementCostPerHour(%.0f, %d, %d) = %.2f; want %.2f", tc.base, tc.level, tc.sweet, got, tc.want)
		}
	}
}

func TestTaxCost(t *testing.T) {
	tests := []struct {
		revenue float64
		rate    float64
		want    float64
	}{
		{1000, 0.1, 100}, // 10% tax
		{0, 0.1, 0},      // zero revenue
		{1000, 0, 0},     // zero rate returns 0
		{1000, -0.1, 0},  // negative rate returns 0
	}
	for _, tc := range tests {
		got := formula.TaxCost(tc.revenue, tc.rate)
		if got != tc.want {
			t.Errorf("TaxCost(%.0f, %.2f) = %.2f; want %.2f", tc.revenue, tc.rate, got, tc.want)
		}
	}
}

func TestUpgradeCost(t *testing.T) {
	tests := []struct {
		base  float64
		level int
		want  float64
	}{
		{10000, 1, 10000 * 1},
		{10000, 5, 10000 * 5},
		{10000, 0, 10000 * 1}, // level 0 clamped to 1
		{5000, 10, 5000 * 10},
	}
	for _, tc := range tests {
		got := formula.UpgradeCost(tc.base, tc.level)
		if got != tc.want {
			t.Errorf("UpgradeCost(%.0f, %d) = %.2f; want %.2f", tc.base, tc.level, got, tc.want)
		}
	}
}

func TestUpgradeDurationSchedule(t *testing.T) {
	base := 2 * time.Minute // farm base construction time
	tests := []struct {
		level int
		want  time.Duration
	}{
		{1, base},
		{2, 2 * base},
		{3, 3 * base},
		{4, 6 * base},
		{5, 10 * base},
		{6, 15 * base},
		{7, 21 * base},
	}
	for _, tc := range tests {
		if got := formula.UpgradeDuration(1, tc.level); got != tc.want {
			t.Errorf("UpgradeDuration(farm, level %d) = %s; want %s", tc.level, got, tc.want)
		}
	}
	if got := formula.UpgradeDuration(999, 1); got != 4*time.Minute {
		t.Errorf("unknown building base = %s; want 4m", got)
	}
}

func TestBuildingMarketAndUpgradeResourceCosts(t *testing.T) {
	qp := map[int]float64{101: 4, 102: 55, 108: 16, 111: 1}
	prices := map[int]float64{101: 10, 102: 2, 108: 5, 111: 20}
	if got := formula.BuildingMarketCost(prices, 2, qp); got != 500 {
		t.Fatalf("BuildingMarketCost = %v; want 500", got)
	}
	delete(prices, 108)
	if got := formula.BuildingMarketCost(prices, 2, qp); got != 6900 {
		t.Fatalf("fallback BuildingMarketCost = %v; want 6900", got)
	}
	upgrade := formula.UpgradeResourceCost(qp, 2, 3)
	if upgrade[101] != 24 || upgrade[102] != 330 || upgrade[108] != 96 || upgrade[111] != 6 {
		t.Fatalf("UpgradeResourceCost = %v", upgrade)
	}
	if got := formula.WagePerTick(1.2, 3); got != 1242 {
		t.Fatalf("WagePerTick = %v; want 1242", got)
	}
}

func TestReferenceProductionChain(t *testing.T) {
	base := formula.SalaryAdjustedBase(100, 345, 1)
	if base != 100 {
		t.Fatalf("SalaryAdjustedBase = %v; want 100", base)
	}
	produced := formula.ProducedPerHour(2, 20, 50, base, 10, true)
	if math.Abs(produced-137.5) > 1e-9 {
		t.Fatalf("ProducedPerHour = %v; want 137.5", produced)
	}
	withRobots := formula.AccumulatorProducedPerHour(2, 3, 20, 100, 0)
	if math.Abs(withRobots-294.11764705882354) > 1e-9 {
		t.Fatalf("AccumulatorProducedPerHour = %v", withRobots)
	}
	final := formula.FinalProductionSpeed(100, 20)
	if final != 125 {
		t.Fatalf("FinalProductionSpeed = %v; want 125", final)
	}
	timePerUnit := formula.ProductionTimePerUnit(1, 2, 345)
	if timePerUnit != 2 || formula.QueueProductionTime(4, timePerUnit) != 6 {
		t.Fatalf("queue timing mismatch: unit=%v queue=%v", timePerUnit, formula.QueueProductionTime(4, timePerUnit))
	}
}

func TestTotalBuildingCost(t *testing.T) {
	tests := []struct {
		base  float64
		level int
		want  float64
	}{
		{10000, 1, 10000 * 1 * 2 / 2}, // level 1: 10000
		{10000, 3, 10000 * 3 * 4 / 2}, // level 3: 60000
		{10000, 0, 10000 * 1 * 2 / 2}, // level 0 clamped to 1
		{5000, 5, 5000 * 5 * 6 / 2},   // level 5: 75000
	}
	for _, tc := range tests {
		got := formula.TotalBuildingCost(tc.base, tc.level)
		if got != tc.want {
			t.Errorf("TotalBuildingCost(%.0f, %d) = %.2f; want %.2f", tc.base, tc.level, got, tc.want)
		}
	}
}

// --- Saturation ---

func TestSaturationPriceMultiplier(t *testing.T) {
	tests := []struct {
		sat  float64
		k    float64
		want float64
	}{
		{1.0, 0.15, 1.0},    // balanced
		{1.5, 0.15, 0.925},  // oversupply: 1+(1-1.5)*0.15 = 0.925
		{0.5, 0.15, 1.075},  // undersupply: 1+(1-0.5)*0.15 = 1.075
		{10.0, 0.15, 0.70},  // extreme oversupply: clamped at 0.70
		{-10.0, 0.15, 1.10}, // extreme undersupply: clamped at 1.10
	}
	for _, tc := range tests {
		got := formula.SaturationPriceMultiplier(tc.sat, tc.k)
		if got != tc.want {
			t.Errorf("SaturationPriceMultiplier(%.2f, %.2f) = %g; want %g", tc.sat, tc.k, got, tc.want)
		}
	}
}

func TestEffectivePrice(t *testing.T) {
	tests := []struct {
		base float64
		sat  float64
		evt  float64
		want float64
	}{
		{100, 1.0, 1.0, 100 * 1.0 * 1.0}, // normal
		{100, 0.8, 0, 100 * 0.8 * 1.0},   // zero event multiplier clamped to 1.0
		{100, 0, 1.1, 100 * 0 * 1.1},     // zero saturation multiplier
	}
	for _, tc := range tests {
		got := formula.EffectivePrice(tc.base, tc.sat, tc.evt)
		if got != tc.want {
			t.Errorf("EffectivePrice(%.0f, %.2f, %.2f) = %g; want %g", tc.base, tc.sat, tc.evt, got, tc.want)
		}
	}
}

// --- Commodity Groups ---

func TestGroupOf(t *testing.T) {
	tests := []struct {
		rid  int
		want int
	}{
		{1, formula.GroupGrain},          // Grain (known)
		{2, formula.GroupProcessed},      // Processed (known)
		{3, formula.GroupBakery},         // Bakery (known)
		{4, formula.GroupRestaurantMeal}, // RestaurantMeal (known)
		{99, formula.GroupGeneralMarket}, // unknown falls back to GeneralMarket
	}
	for _, tc := range tests {
		got := formula.GroupOf(tc.rid)
		if got != tc.want {
			t.Errorf("GroupOf(%d) = %d; want %d", tc.rid, got, tc.want)
		}
	}
}

func TestFormulaGolden(t *testing.T) {
	// Known inputs — chosen to produce one distinct, checkable value per formula.
	type inputs struct {
		amount      int
		price       float64
		feeRate     float64
		baseOutput  float64
		speedBonus  float64
		level       int
		baseLabor   float64
		laborIdx    float64
		baseEnergy  float64
		energyIdx   float64
		baseMaint   float64
		baseMgmt    float64
		sweetSpot   int
		revenue     float64
		taxRate     float64
		upgradeBase float64
		faceVal     float64
		ratePct     float64
		sat         float64
		satK        float64
		eventMul    float64
		qty         int
		prodRate    int
		prodMod     float64
		output      float64
		qtyPerUnit  float64
		unitPrice   float64
		matIdx      float64
		buildingVal float64
		alreadySold int
	}

	in := inputs{
		amount:      100,
		price:       23.0,
		feeRate:     0.04,
		baseOutput:  500,
		speedBonus:  0,
		level:       1,
		baseLabor:   500,
		laborIdx:    1.0,
		baseEnergy:  300,
		energyIdx:   1.0,
		baseMaint:   200,
		baseMgmt:    500,
		sweetSpot:   7,
		revenue:     1000,
		taxRate:     0.1,
		upgradeBase: 10000,
		faceVal:     5000,
		ratePct:     1.2,
		sat:         1.0,
		satK:        0.15,
		eventMul:    1.0,
		qty:         10,
		prodRate:    500,
		prodMod:     1.0,
		output:      10,
		qtyPerUnit:  2,
		unitPrice:   5,
		matIdx:      1.0,
		buildingVal: 500000,
		alreadySold: 0,
	}

	expected := map[string]float64{
		"ExchangeFee(100, 23, 0.04)":           92.0,
		"OutputPerHour(500, 0, 1)":             500.0,
		"DurationSeconds(10, 500, 1, 1.0)":     72.0,
		"DailyBondInterest(100, 5000, 1.2)":    6000.0,
		"MaxIssuableBonds(500000, 5000, 0)":    100.0,
		"LaborCostPerHour(500, 1.0, 1)":        500.0,
		"EnergyCostPerHour(300, 1.0, 1)":       300.0,
		"MaintenanceCostPerHour(200, 1)":       200.0,
		"ManagementCostPerHour(500, 1, 7)":     500.0,
		"TaxCost(1000, 0.1)":                   100.0,
		"UpgradeCost(10000, 1)":                10000.0,
		"TotalBuildingCost(10000, 1)":          10000.0,
		"InputCost(10, 2, 5, 1.0)":             100.0,
		"SaturationPriceMultiplier(1.0, 0.15)": 1.0,
		"EffectivePrice(100, 1.0, 1.0)":        100.0,
	}

	for key, want := range expected {
		t.Run(key, func(t *testing.T) {
			var got float64
			switch key {
			case "ExchangeFee(100, 23, 0.04)":
				got = formula.ExchangeFee(in.amount, in.price, in.feeRate)
			case "OutputPerHour(500, 0, 1)":
				got = formula.OutputPerHour(in.baseOutput, in.speedBonus, in.level)
			case "DurationSeconds(10, 500, 1, 1.0)":
				got = formula.DurationSeconds(in.qty, in.prodRate, in.level, in.prodMod)
			case "DailyBondInterest(100, 5000, 1.2)":
				got = formula.DailyBondInterest(in.amount, in.faceVal, in.ratePct)
			case "MaxIssuableBonds(500000, 5000, 0)":
				got = float64(formula.MaxIssuableBonds(in.buildingVal, in.faceVal, in.alreadySold))
			case "LaborCostPerHour(500, 1.0, 1)":
				got = formula.LaborCostPerHour(in.baseLabor, in.laborIdx, in.level)
			case "EnergyCostPerHour(300, 1.0, 1)":
				got = formula.EnergyCostPerHour(in.baseEnergy, in.energyIdx, in.level)
			case "MaintenanceCostPerHour(200, 1)":
				got = formula.MaintenanceCostPerHour(in.baseMaint, in.level)
			case "ManagementCostPerHour(500, 1, 7)":
				got = formula.ManagementCostPerHour(in.baseMgmt, in.level, in.sweetSpot)
			case "TaxCost(1000, 0.1)":
				got = formula.TaxCost(in.revenue, in.taxRate)
			case "UpgradeCost(10000, 1)":
				got = formula.UpgradeCost(in.upgradeBase, in.level)
			case "TotalBuildingCost(10000, 1)":
				got = formula.TotalBuildingCost(in.upgradeBase, in.level)
			case "InputCost(10, 2, 5, 1.0)":
				got = formula.InputCost(in.output, in.qtyPerUnit, in.unitPrice, in.matIdx)
			case "SaturationPriceMultiplier(1.0, 0.15)":
				got = formula.SaturationPriceMultiplier(in.sat, in.satK)
			case "EffectivePrice(100, 1.0, 1.0)":
				got = formula.EffectivePrice(100, in.sat, in.eventMul)
			default:
				t.Fatalf("unknown formula key: %s", key)
			}
			if got != want {
				t.Errorf("%s = %g; want %g", key, got, want)
			}
		})
	}
}

// --- Retail ---

func TestUnitsSoldPerHour_StableMarket(t *testing.T) {
	// Saturation=1.0: balanced market
	got := formula.UnitsSoldPerHour(
		0.8,   // buildingKindModifier
		0.01,  // buildingLevelsNeededPerUnitPerHour
		8.0,   // modeledProductionCostPerUnit
		200.0, // modeledStoreWages
		15.0,  // modeledUnitsSoldAnHour
		24.0,  // price
		6.0,   // quality
		1.0,   // saturation (balanced)
		0,     // salesModifierPct
		1,     // size
		1,     // acceleration
		1,     // weatherSellingSpeedMultiplier
	)
	if got <= 0 {
		t.Errorf("UnitsSoldPerHour stable market = %g, want > 0", got)
	}
	if got > 300 {
		t.Errorf("UnitsSoldPerHour stable market = %g, want reasonable (< 300)", got)
	}
	t.Logf("UnitsSoldPerHour (stable) = %.2f", got)
}

func TestUnitsSoldPerHour_HighSaturation(t *testing.T) {
	// Saturation=2.0: oversupply, should sell less
	got := formula.UnitsSoldPerHour(0.8, 0.01, 8.0, 200.0, 15.0, 24.0, 6.0, 2.0, 0, 1, 1, 1)
	if got <= 0 {
		t.Errorf("UnitsSoldPerHour high saturation = %g, want > 0 (still sells but less)", got)
	}
	t.Logf("UnitsSoldPerHour (saturation=2.0) = %.2f", got)
}

func TestUnitsSoldPerHour_ZeroSaturation(t *testing.T) {
	// Saturation=0: undersupply, should sell more
	got := formula.UnitsSoldPerHour(0.8, 0.01, 8.0, 200.0, 15.0, 24.0, 6.0, 0, 0, 1, 1, 1)
	if got <= 0 {
		t.Errorf("UnitsSoldPerHour zero saturation = %g, want > 0", got)
	}
	t.Logf("UnitsSoldPerHour (saturation=0) = %.2f", got)
}

func TestUnitsSoldPerHour_LowPrice(t *testing.T) {
	// Price below production cost → no units sold
	got := formula.UnitsSoldPerHour(0.8, 0.01, 8.0, 200.0, 15.0, 5.0, 6.0, 1.0, 0, 1, 1, 1)
	if got != 0 {
		t.Errorf("UnitsSoldPerHour below-cost price = %g, want 0", got)
	}
}

// --- Research Formulas ---

func TestResearchFormulas(t *testing.T) {
	t.Run("ResearchBaseCost", func(t *testing.T) {
		tests := []struct {
			tier     int
			baseCost float64
			want     float64
		}{
			{tier: 1, baseCost: 1000, want: 1000},
			{tier: 2, baseCost: 1000, want: 2000},
			{tier: 3, baseCost: 1000, want: 4000},
			{tier: 4, baseCost: 1000, want: 8000},
		}
		for _, tc := range tests {
			got := formula.ResearchBaseCost(tc.tier, tc.baseCost)
			if got != tc.want {
				t.Errorf("ResearchBaseCost(tier=%d, baseCost=%g) = %g, want %g", tc.tier, tc.baseCost, got, tc.want)
			}
		}
	})

	t.Run("ResearchLevelCost", func(t *testing.T) {
		tests := []struct {
			baseCost float64
			level    int
			want     float64
			approx   bool
		}{
			{baseCost: 1000, level: 1, want: 1000},
			{baseCost: 1000, level: 2, want: 1200},
			{baseCost: 1000, level: 3, want: 1440},
			{baseCost: 1000, level: 10, want: 5159.78, approx: true},
		}
		for _, tc := range tests {
			got := formula.ResearchLevelCost(tc.baseCost, tc.level)
			if tc.approx {
				if math.Abs(got-tc.want) > 0.01 {
					t.Errorf("ResearchLevelCost(baseCost=%g, level=%d) = %g, want ~%g", tc.baseCost, tc.level, got, tc.want)
				}
			} else {
				if got != tc.want {
					t.Errorf("ResearchLevelCost(baseCost=%g, level=%d) = %g, want %g", tc.baseCost, tc.level, got, tc.want)
				}
			}
		}
	})

	t.Run("ResearchSpeedBonus", func(t *testing.T) {
		tests := []struct {
			level int
			want  float64
		}{
			{level: 0, want: 1.0},
			{level: 50, want: 1.1},
			{level: 100, want: 1.2},
		}
		for _, tc := range tests {
			got := formula.ResearchSpeedBonus(tc.level)
			if got != tc.want {
				t.Errorf("ResearchSpeedBonus(level=%d) = %g, want %g", tc.level, got, tc.want)
			}
		}
	})
}
