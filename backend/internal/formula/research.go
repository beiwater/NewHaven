package formula

import "math"

// ResearchBaseCost returns the money cost for level 1 of a resource at the given tier.
// Tier 1 = baseCost, Tier 2 = 2x, Tier 3 = 4x, Tier 4 = 8x.
func ResearchBaseCost(tier int, baseCost float64) float64 {
	if tier < 1 {
		tier = 1
	}
	return baseCost * math.Pow(2, float64(tier-1))
}

// ResearchLevelCost returns the money cost to research a specific level (1-indexed).
// Each level costs ResearchCostGrowth (1.2) times the previous level.
// Returns 0 for level < 1 or > MaxResearchLevel.
func ResearchLevelCost(baseCost float64, level int) float64 {
	if level < 1 {
		return 0
	}
	return baseCost * math.Pow(1.2, float64(level-1))
}

// QualityResearchCost returns the cash required to unlock one target quality.
// Product tier makes downstream goods more expensive to study, while the
// configured growth rate keeps each step meaningful without making Q12
// unreachable. Costs are rounded up to a player-readable $10 boundary.
func QualityResearchCost(tier, targetQuality int, baseCost, growth float64) float64 {
	if targetQuality < 1 || targetQuality > MaxProductQuality {
		return 0
	}
	if baseCost <= 0 {
		baseCost = 1000
	}
	if growth <= 1 {
		growth = 1.2
	}
	raw := ResearchBaseCost(tier, baseCost) * math.Pow(growth, float64(targetQuality-1))
	return math.Ceil(raw/10) * 10
}

// ResearchSpeedBonus returns the production speed multiplier from N research levels.
// Each level gives 0.2% (0.002) production speed bonus.
func ResearchSpeedBonus(level int) float64 {
	return 1.0 + float64(level)*0.002
}
