package formula

import "math"

// ExchangeFee computes the fee charged for a market order execution.
// Matches legacy formula: ceil(amount * price * feeRate).
// feeRate is a decimal fraction (e.g., 0.04 = 4%).
func ExchangeFee(amount int, price float64, feeRate float64) float64 {
	return math.Ceil(float64(amount) * price * feeRate)
}

// TickStep returns the minimum price increment at the given price level.
// Matches legacy formula.TickStep exactly.
func TickStep(price float64) float64 {
	switch {
	case price >= 20000:
		return 500
	case price >= 10000:
		return 100
	case price >= 5000:
		return 25
	case price >= 1000:
		return 10
	case price >= 500:
		return 5
	case price >= 200:
		return 2
	case price >= 100:
		return 1
	case price >= 50:
		return 0.5
	case price >= 20:
		return 0.25
	case price >= 5:
		return 0.1
	case price >= 2:
		return 0.05
	case price >= 1:
		return 0.01
	case price >= 0.5:
		return 0.005
	default:
		return 0.001
	}
}

// IsValidTick checks whether price is on a valid tick boundary.
// Matches legacy formula.IsValidTick exactly.
func IsValidTick(price float64) bool {
	step := TickStep(price)
	scaled := price / step
	return math.Abs(scaled-math.Round(scaled)) < 1e-9
}
