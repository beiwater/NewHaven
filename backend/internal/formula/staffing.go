package formula

import "math"

// WorkerHourlyWage is the company-wide hourly cost of one worker. Buildings
// never choose their headcount: their type and level determine it. Keeping the
// hourly rate shared makes staffing legible while still making larger buildings
// materially more expensive to run.
const (
	// WorkerHourlyWage is shared across the economy so headcount, not an
	// arbitrary per-building salary slider, determines operating cost.
	WorkerHourlyWage = 345.0
	// TargetBuildingProfitPerHour is the level-one balance target used by the
	// market anchor and retail recommendation. Higher levels scale this target
	// linearly with their production line and fixed workforce.
	TargetBuildingProfitPerHour = 300.0
)

// BuildingWorkerCount returns the fixed headcount required by a building. A
// level adds one full base team, so a level-two Cafe employs twice the workers
// of a level-one Cafe. This intentionally avoids a separate hire/fire system.
func BuildingWorkerCount(buildingKind, level int) int {

	baseByKind := map[int]int{
		1:  2, // Farm
		2:  3, // Barn
		3:  4, // Mill
		4:  5, // Kitchen
		5:  6, // Bakery
		6:  3, // Market Stall
		7:  5, // Cafe
		8:  4, // Food Truck
		9:  8, // Restaurant
		10: 7,
		11: 8,
		12: 9,
	}
	if level < 1 {
		level = 1
	}
	base, ok := baseByKind[buildingKind]
	if !ok {
		base = 3
	}
	return base * level
}

// BuildingHourlyWage is the payroll rate while a building is operating. Idle
// and upgrading buildings have no active payroll.
func BuildingHourlyWage(buildingKind, level int) float64 {
	return float64(BuildingWorkerCount(buildingKind, level)) * WorkerHourlyWage
}

// RetailPriceSpeedMultiplier describes customer sensitivity around the current
// recommended price. Discounting can provide at most a modest 25% speed boost;
// pricing above the reference slows demand quadratically. That means revenue
// per hour falls at extreme prices instead of allowing arbitrary price exploits.
func RetailPriceSpeedMultiplier(price, recommendedPrice float64) float64 {
	if price <= 0 || recommendedPrice <= 0 || math.IsNaN(price) || math.IsInf(price, 0) {
		return 0
	}
	ratio := price / recommendedPrice
	if ratio <= 1 {
		return math.Min(1.25, 1+(1-ratio)*0.25)
	}
	return 1 / (ratio * ratio)
}
