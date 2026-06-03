package formula

// --- Old salary-based formulas removed per v1.3.1 spec ---
// SalaryMid, AverageSalary, RobotBonus, BaseProductionRate, ProducedPerHour, ProductionTimeSeconds

// OutputPerHour computes building output using the minimal model:
//
//	OutputPerHour_Lv1  = BuildingBaseOutputPerHour_Lv1 * (1 + FinalProductionSpeedBonus)
//	OutputPerHour_N    = OutputPerHour_Lv1 * BuildingLevel
//
// speedBonusPct is a percentage (0-100+), e.g. 10 means +10%.
// level is the building level (1-based).
func OutputPerHour(baseOutputPerHour float64, speedBonusPct float64, level int) float64 {
	if level < 1 {
		level = 1
	}
	lv1 := baseOutputPerHour * (1.0 + speedBonusPct/100.0)
	return lv1 * float64(level)
}

// ProductionDurationSeconds estimates time to produce `amount` units.
// Uses secondsPerUnit (derived from base output rate), divided by level and boost.
func ProductionDurationSeconds(amount int, secondsPerUnit float64, level int, boostMultiplier float64) float64 {
	if secondsPerUnit <= 0 {
		secondsPerUnit = 3600.0 // default 1 hour
	}
	if level < 1 {
		level = 1
	}
	dur := float64(amount) * secondsPerUnit / float64(level)
	if boostMultiplier > 1.0 {
		dur /= boostMultiplier
	}
	return dur
}
