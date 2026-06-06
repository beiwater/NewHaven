package formula

import "math"

// DailyBondInterest computes daily interest on a bond.
// Formula: floor(amount * faceValue * interestRatePct / 100).
// interestRatePct is a percentage (e.g., 1.2 means 1.2%).
// Matches legacy formula.DailyBondInterest with BondFaceValue passed explicitly.
func DailyBondInterest(amount int, faceValue float64, interestRatePct float64) float64 {
	return math.Floor(float64(amount) * faceValue * interestRatePct / 100.0)
}

// MaxIssuableBonds computes how many bonds a company can issue given its building value.
// Formula: floor(totalBuildingValue / faceValue) - alreadySold, clamped to 0.
func MaxIssuableBonds(totalBuildingValue float64, faceValue float64, alreadySold int) int {
	if faceValue <= 0 {
		return 0
	}
	v := int(math.Floor(totalBuildingValue/faceValue)) - alreadySold
	if v < 0 {
		return 0
	}
	return v
}
