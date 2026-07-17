package formula

import "math"

// CTOProductionMultiplier follows the documented executive formula: every
// effective CTO science point adds two percent to production speed. The input
// is capped to keep malformed persistence data from creating a negative or
// unbounded job duration.
func CTOProductionMultiplier(skill float64) float64 {
	skill = math.Max(0, math.Min(100, skill))
	return (100 + skill*2) / 100
}

// CMOSalesBonusPct converts communication skill into the retail speed bonus.
// A full 100-point skill would be too dominant in the early economy, so this
// intentionally reaches a capped +50% demand speed at 100 effective skill.
func CMOSalesBonusPct(skill float64) float64 {
	skill = math.Max(0, math.Min(100, skill))
	return skill / 2
}
