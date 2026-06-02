package formula

import "math"

var BondFaceValue = 5000.0

func SetBondFaceValue(v float64) { BondFaceValue = v }

func DailyBondInterest(amount int, interestRatePct float64) float64 {
	return math.Floor(float64(amount) * 50.0 * interestRatePct)
}

func PeriodBondInterest(amount int, interestRatePct float64) float64 {
	return math.Floor(float64(amount)*BondFaceValue*interestRatePct) / 100.0
}

func MaxIssuableBonds(totalBuildingValue float64, alreadySold int) int {
	v := int(math.Floor(totalBuildingValue/BondFaceValue)) - alreadySold
	if v < 0 {
		return 0
	}
	return v
}
