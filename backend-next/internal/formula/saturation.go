package formula

const (
	GroupGrain          = 1
	GroupDairy          = 2
	GroupProcessed      = 3
	GroupBakery         = 4
	GroupRestaurantMeal = 8
	GroupGeneralMarket  = 5
)

var ResourceCommodityGroup = map[int]int{
	1: GroupGrain,
	2: GroupProcessed,
	3: GroupBakery,
	4: GroupRestaurantMeal,
}

func GroupOf(resourceID int) int {
	if g, ok := ResourceCommodityGroup[resourceID]; ok {
		return g
	}
	return GroupGeneralMarket
}

func SaturationPriceMultiplier(marketSaturation float64, saturationK float64) float64 {
	if saturationK <= 0 {
		saturationK = 0.15
	}
	mul := 1.0 + (1.0-marketSaturation)*saturationK
	return clamp(mul, 0.70, 1.10)
}

func EffectivePrice(basePrice, saturationMultiplier, eventMultiplier float64) float64 {
	if eventMultiplier <= 0 {
		eventMultiplier = 1.0
	}
	return basePrice * saturationMultiplier * eventMultiplier
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}
