package formula

const buildingCostFallbackPerUnit = 3450.0

// BuildingMarketCost values a construction recipe using live market prices.
// If any required quote is unavailable, the deterministic fallback is used.
func BuildingMarketCost(tickerPrices map[int]float64, costUnits float64, qualityPoints map[int]float64) float64 {
	if costUnits <= 0 {
		return 0
	}
	total := 0.0
	for resourceID, points := range qualityPoints {
		if points <= 0 {
			continue
		}
		price := tickerPrices[resourceID]
		if price <= 0 {
			return costUnits * buildingCostFallbackPerUnit
		}
		total += price * costUnits * points
	}
	if total <= 0 {
		return costUnits * buildingCostFallbackPerUnit
	}
	return total
}

// UpgradeResourceCost returns the material consumption for one upgrade.
func UpgradeResourceCost(qualityPoints map[int]float64, costUnits float64, currentSize int) map[int]float64 {
	if currentSize < 1 {
		currentSize = 1
	}
	result := make(map[int]float64, len(qualityPoints))
	for resourceID, points := range qualityPoints {
		if points > 0 && costUnits > 0 {
			result[resourceID] = points * costUnits * float64(currentSize)
		}
	}
	return result
}

// WagePerTick computes 345 * salaryModifier * building size.
func WagePerTick(salaryModifier float64, size int) float64 {
	if size < 1 {
		size = 1
	}
	if salaryModifier < 0 {
		salaryModifier = 0
	}
	return 345 * salaryModifier * float64(size)
}

// LaborCostPerHour computes per-hour labor cost, scaled linearly by the
// building level and a labor-cost index.
//
//	baseLaborCost * laborCostIndex * level
func LaborCostPerHour(baseLaborCost float64, laborCostIndex float64, level int) float64 {
	if level < 1 {
		level = 1
	}
	return baseLaborCost * laborCostIndex * float64(level)
}

// EnergyCostPerHour computes per-hour energy cost, scaled linearly by the
// building level and an energy-cost index.
//
//	baseEnergyCost * energyCostIndex * level
func EnergyCostPerHour(baseEnergyCost float64, energyCostIndex float64, level int) float64 {
	if level < 1 {
		level = 1
	}
	return baseEnergyCost * energyCostIndex * float64(level)
}

// InputCost computes the total cost of material inputs for a given output
// volume, scaled by a material-cost index.
//
//	output * inputQtyPerUnit * inputUnitPrice * materialCostIndex
func InputCost(output float64, inputQtyPerUnit float64, inputUnitPrice float64, materialCostIndex float64) float64 {
	return output * inputQtyPerUnit * inputUnitPrice * materialCostIndex
}

// MaintenanceCostPerHour computes per-hour maintenance cost, scaled linearly
// by the building level.
//
//	baseMaintenance * level
func MaintenanceCostPerHour(baseMaintenance float64, level int) float64 {
	if level < 1 {
		level = 1
	}
	return baseMaintenance * float64(level)
}

// ManagementCostPerHour computes per-hour management cost. Below the
// sweet-spot level the cost scales linearly; beyond it costs scale
// quadratically (level*level / sweetSpotLevel) to penalise over-staffing.
//
//	baseManagement * level (level <= sweetSpotLevel)
//	baseManagement * level * level / sweetSpotLevel (level > sweetSpotLevel)
func ManagementCostPerHour(baseManagement float64, level int, sweetSpotLevel int) float64 {
	if level < 1 {
		level = 1
	}
	if sweetSpotLevel <= 0 {
		sweetSpotLevel = 7
	}
	if level <= sweetSpotLevel {
		return baseManagement * float64(level)
	}
	return baseManagement * float64(level*level) / float64(sweetSpotLevel)
}

// TaxCost computes tax on revenue at the given rate. Returns 0 when taxRate
// is zero or negative.
func TaxCost(revenue float64, taxRate float64) float64 {
	if taxRate <= 0 {
		return 0
	}
	return revenue * taxRate
}

// UpgradeCost computes the cost to upgrade a building at its current size.
//
//	baseBuildCost * currentSize
func UpgradeCost(baseBuildCost float64, currentSize int) float64 {
	if currentSize < 1 {
		currentSize = 1
	}
	return baseBuildCost * float64(currentSize)
}

// TotalBuildingCost computes the cumulative cost of building and upgrading
// to a given level.
//
//	baseBuildCost * level * (level + 1) / 2
func TotalBuildingCost(baseBuildCost float64, level int) float64 {
	if level < 1 {
		level = 1
	}
	return baseBuildCost * float64(level) * float64(level+1) / 2.0
}
