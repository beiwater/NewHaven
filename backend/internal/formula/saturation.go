package formula

// Commodity group constants.
const (
	GroupGrain          = 1
	GroupDairy          = 2
	GroupProcessed      = 3
	GroupBakery         = 4
	GroupGeneralMarket  = 5
	GroupCafeDessert    = 6
	GroupStreetFood     = 7
	GroupRestaurantMeal = 8
	GroupFinance        = 9
)

// ResourceCommodityGroup maps resource IDs to commodity groups.
// v1.3.1's minimal chain is Wheat -> Flour -> Bread -> Meals.
var ResourceCommodityGroup = map[int]int{
	1: GroupGrain,
	2: GroupProcessed,
	3: GroupBakery,
	4: GroupRestaurantMeal,
}

// GroupOf returns the commodity group for a resource ID, defaulting to GroupGeneralMarket.
func GroupOf(resourceID int) int {
	if g, ok := ResourceCommodityGroup[resourceID]; ok {
		return g
	}
	return GroupGeneralMarket
}

// SaturationPriceMultiplier computes the price multiplier from market saturation.
//
//	SaturationPriceMultiplier = CLAMP(0.70, 1.10, 1 + (1 - marketSaturation) * SaturationK)
//
// Symmetric: oversupply (>1) reduces prices, undersupply (<1) raises prices.
// Floor 0.70, Ceiling 1.10 prevent extreme swings.
func SaturationPriceMultiplier(marketSaturation float64, saturationK float64) float64 {
	if saturationK <= 0 {
		saturationK = 0.15
	}
	mul := 1.0 + (1.0-marketSaturation)*saturationK
	return clamp(mul, 0.70, 1.10)
}

// EffectivePrice applies saturation and event multipliers to a base price.
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
