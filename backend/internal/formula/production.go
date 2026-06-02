package formula

import "math"

var SalaryMid = map[int]float64{0: 655, 1: 700, 2: 745}

const (
	AverageSalary = 345.0
	RobotBonus    = 4.0
)

func BaseProductionRate(producedPerHourRaw, salaryModifier float64, salaryLevel int) float64 {
	sMid := SalaryMid[salaryLevel]
	if sMid == 0 {
		sMid = SalaryMid[1]
	}
	return producedPerHourRaw * math.Pow(AverageSalary/sMid, salaryModifier)
}

func ProducedPerHour(size int, baseRate, salaryPercent float64, robotCount int, isAccumulator bool, speedModifierPct float64, qualityPct float64, isMining bool) float64 {
	adjusted := baseRate
	if isMining {
		adjusted *= qualityPct / 100.0
	}
	adjusted *= (speedModifierPct / 100.0) + 1.0
	effectiveSalary := salaryPercent
	if isAccumulator {
		effectiveSalary += RobotBonus * float64(robotCount)
	}
	den := 1.0 - effectiveSalary/100.0
	if den <= 0.01 {
		den = 0.01
	}
	return float64(size) * adjusted / den
}

func ProductionTimeSeconds(size int, buildingSalaryModifier, producedPerHour float64, eventMultiplier float64) float64 {
	if producedPerHour <= 0 {
		return math.Inf(1)
	}
	if eventMultiplier <= 0 {
		eventMultiplier = 1
	}
	return (345.0 * buildingSalaryModifier * float64(size) / producedPerHour) * eventMultiplier
}
