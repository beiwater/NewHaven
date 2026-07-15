package formula

import "math"

// SalaryAdjustedBase applies the reference salary midpoint adjustment (B6t).
func SalaryAdjustedBase(producedPerHourRaw, salaryMid, salaryModifier float64) float64 {
	if producedPerHourRaw <= 0 || salaryMid <= 0 {
		return 0
	}
	return producedPerHourRaw * math.Pow(345/salaryMid, salaryModifier)
}

// ProducedPerHour applies size, speed, salary, and optional mining quality.
func ProducedPerHour(size int, salaryPct, qualityPct, baseProduction, speedModifierPct float64, mining bool) float64 {
	if size < 1 {
		size = 1
	}
	if baseProduction <= 0 {
		return 0
	}
	if salaryPct >= 100 {
		salaryPct = 99
	}
	adjustedBase := baseProduction * (1 + speedModifierPct/100)
	result := float64(size) * adjustedBase / (1 - salaryPct/100)
	if mining {
		result *= math.Max(0, qualityPct) / 100
	}
	return result
}

// AccumulatorProducedPerHour adds four salary points per robot.
func AccumulatorProducedPerHour(size, robots int, salaryPct, baseProduction, speedModifierPct float64) float64 {
	if robots < 0 {
		robots = 0
	}
	return ProducedPerHour(size, salaryPct+4*float64(robots), 100, baseProduction, speedModifierPct, false)
}

// FinalProductionSpeed applies production, recreation, and accumulator bonuses.
func FinalProductionSpeed(perHour, totalBonusPct float64) float64 {
	if perHour <= 0 {
		return 0
	}
	if totalBonusPct >= 100 {
		totalBonusPct = 99
	}
	return perHour / (1 - totalBonusPct/100)
}

// ProductionTimePerUnit computes wage-cost time per produced unit.
func ProductionTimePerUnit(salaryModifier float64, size int, producedPerHour float64) float64 {
	if producedPerHour <= 0 {
		return 0
	}
	return WagePerTick(salaryModifier, size) / producedPerHour
}

// QueueProductionTime applies spare capacity after administrative reduction.
func QueueProductionTime(spareCapacity int, timePerUnit float64) float64 {
	return math.Max(0, float64(spareCapacity-1)) * math.Max(0, timePerUnit)
}

func OutputPerHour(baseOutputPerHour float64, speedBonusPct float64, level int) float64 {
	if level < 1 {
		level = 1
	}
	lv1 := baseOutputPerHour * (1.0 + speedBonusPct/100.0)
	return lv1 * float64(level)
}

func DurationSeconds(quantity int, producedPerHourRaw int, level int, productionMod float64) float64 {
	if quantity <= 0 {
		return 0
	}
	if level < 1 {
		level = 1
	}
	if producedPerHourRaw <= 0 {
		return 0
	}
	if productionMod <= 0 {
		productionMod = 1.0
	}
	rate := float64(producedPerHourRaw) * float64(level) * productionMod
	duration := math.Ceil(float64(quantity) / rate * 3600)
	if duration < 30 {
		duration = 30
	}
	return duration
}
