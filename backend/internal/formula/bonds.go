package formula

import "math"

var BondFaceValue = 5000.0

func SetBondFaceValue(v float64) { BondFaceValue = v }

// DailyBondInterest computes daily interest on a bond.
//
// The service layer passes interestRatePct = b.Interest * 100 (e.g. 0.012 → 1.2).
// Formula: amount * BondFaceValue * (interestRatePct / 100)  =  amount * BondFaceValue * b.Interest
//
// The old 50x coefficient (amount * 50 * interestRatePct) was equivalent:
// 50 = BondFaceValue/100, so amount*(BFV/100)* (b.Interest*100) = amount*BFV*b.Interest.
// The "50x" was not a bug — just obfuscated arithmetic.
//
// At b.Interest=0.012 (1.2%), one bond (amount=1) yields 5000*0.012 = 60/day.
// This is 1.2%/day game convention for short-term bonds. Adjust rates in game.json
// (bond_min_interest, bond_max_interest) if balance changes are needed.
func DailyBondInterest(amount int, interestRatePct float64) float64 {
	return math.Floor(float64(amount) * BondFaceValue * interestRatePct / 100.0)
}

// PeriodBondInterest computes interest over a period. Semantically equivalent
// to DailyBondInterest. The old version had a /100 at the end that double-counted
// with the service-layer *100; this is now harmonized.
func PeriodBondInterest(amount int, interestRatePct float64) float64 {
	return DailyBondInterest(amount, interestRatePct)
}

func MaxIssuableBonds(totalBuildingValue float64, alreadySold int) int {
	v := int(math.Floor(totalBuildingValue/BondFaceValue)) - alreadySold
	if v < 0 {
		return 0
	}
	return v
}
