package formula

import "math"

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
