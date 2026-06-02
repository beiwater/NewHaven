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

// ResourceCommodityGroup maps resource IDs → commodity group.
// Built from the spec v1.3.2 commodity group table.
var ResourceCommodityGroup = map[int]int{
	// Grain
	1: GroupGrain, 2: GroupGrain, 3: GroupGrain, 4: GroupGrain, 5: GroupGrain, 6: GroupGrain,
	// Dairy
	66: GroupDairy, 72: GroupDairy, 120: GroupDairy,
	// GeneralMarket
	8: GroupGeneralMarket, 9: GroupGeneralMarket, 10: GroupGeneralMarket,
	11: GroupGeneralMarket, 12: GroupGeneralMarket,
	// Processed
	7: GroupProcessed, 121: GroupProcessed, 122: GroupProcessed, 127: GroupProcessed,
	133: GroupProcessed, 134: GroupProcessed, 135: GroupProcessed, 137: GroupProcessed,
	139: GroupProcessed, 141: GroupProcessed,
	// Bakery
	115: GroupBakery, 116: GroupBakery, 117: GroupBakery,
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
