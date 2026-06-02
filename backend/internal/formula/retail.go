package formula

import "math"

const (
	RetailQualityWeight = 0.3
	RetailZor           = 370.0
)

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// Simplified executable version of decompiled retail model.
func UnitsSoldPerHour(
	buildingKindModifier float64,
	buildingLevelsNeededPerUnitPerHour float64,
	modeledProductionCostPerUnit float64,
	modeledStoreWages float64,
	modeledUnitsSoldAnHour float64,
	price float64,
	quality float64,
	saturation float64,
	salesModifierPct float64,
	size int,
	acceleration float64,
	weatherSellingSpeedMultiplier float64,
) float64 {
	d := clamp(2-saturation, 0, 2)
	p := math.Max(0.9, d/2+0.5)
	f := quality / 12.0
	g := RetailZor * (buildingLevelsNeededPerUnitPerHour*modeledUnitsSoldAnHour + 1) * buildingKindModifier * (d / 2 * (1 + f*RetailQualityWeight))
	m := modeledUnitsSoldAnHour * p
	if m <= 0 {
		return 0
	}
	v := modeledProductionCostPerUnit + (g+modeledStoreWages)/m
	_ = v
	den := price - modeledProductionCostPerUnit
	if den <= 0.0001 {
		return 0
	}
	a := (modeledStoreWages + g) / (den * den)
	b := g - (m-price)*(m-price)*a
	if b+modeledStoreWages <= 0 {
		return 0
	}
	res := (acceleration*den*3600 - modeledStoreWages) / (b + modeledStoreWages)
	if size <= 0 {
		size = 1
	}
	if acceleration <= 0 {
		acceleration = 1
	}
	res = res / float64(size) / acceleration
	res = res - res*salesModifierPct/100.0
	if weatherSellingSpeedMultiplier > 0 {
		res = res / weatherSellingSpeedMultiplier
	}
	if math.IsNaN(res) || math.IsInf(res, 0) {
		return 0
	}
	if res < 0 {
		return 0
	}
	return res
}
