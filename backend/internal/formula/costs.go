package formula

// Cost formulas using global cost indices.
//
// Spec v1.3.1:
//
//	LaborCostPerHour_Level    = BaseLaborCostPerHour_Lv1 * LaborCostIndex * BuildingLevel
//	InputCostPerHour_Level    = OutputPerHour_Level * InputQtyPerUnit * InputUnitPrice * MaterialCostIndex
//	EnergyCostPerHour_Level   = BaseEnergyCostPerHour_Lv1 * EnergyCostIndex * BuildingLevel
//	MaintenanceCostPerHour    = BaseMaintenancePerHour_Lv1 * BuildingLevel
//	ManagementCostPerHour     = BaseManagementCost_Lv1 * BuildingLevel²  (sweet-spot convergence)
//	TaxCostPerHour            = RevenuePerHour * EffectiveTaxRate

// LaborCostPerHour computes per-hour labor cost scaled by level and index.
func LaborCostPerHour(baseLaborCost float64, laborCostIndex float64, level int) float64 {
	if level < 1 {
		level = 1
	}
	return baseLaborCost * laborCostIndex * float64(level)
}

// EnergyCostPerHour computes per-hour energy cost scaled by level and index.
func EnergyCostPerHour(baseEnergyCost float64, energyCostIndex float64, level int) float64 {
	if level < 1 {
		level = 1
	}
	return baseEnergyCost * energyCostIndex * float64(level)
}

// InputCost computes cost of material inputs, scaled by material cost index.
// output = units produced per hour
// inputPerUnit = quantity of input per output unit
// inputPrice = current market price of the input
func InputCost(output float64, inputQtyPerUnit float64, inputUnitPrice float64, materialCostIndex float64) float64 {
	return output * inputQtyPerUnit * inputUnitPrice * materialCostIndex
}

// MaintenanceCostPerHour computes maintenance cost scaled by level.
func MaintenanceCostPerHour(baseMaintenance float64, level int) float64 {
	if level < 1 {
		level = 1
	}
	return baseMaintenance * float64(level)
}

// ManagementCostPerHour computes management/admin cost with quadratic scaling
// after the sweet-spot level (7). This creates convergence:
//
//	Level 1-7:   managementCost * level
//	Level 8+:    managementCost * level² / sweetSpot
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
	// After sweet spot: quadratic acceleration
	// At level 10: baseManagement * 100 / 7 ≈ 14.3x linear equivalent
	return baseManagement * float64(level*level) / float64(sweetSpotLevel)
}

// TaxCost computes tax on revenue at given rate.
func TaxCost(revenue float64, taxRate float64) float64 {
	if taxRate <= 0 {
		return 0
	}
	return revenue * taxRate
}

// UpgradeCost computes the cost to upgrade a building to a given level.
// Spec: UpgradeCost(level) = baseBuildCost * level
// Cumulative: TotalBuildingCost(level) = baseBuildCost * level * (level + 1) / 2
func UpgradeCost(baseBuildCost float64, targetLevel int) float64 {
	if targetLevel < 1 {
		targetLevel = 1
	}
	return baseBuildCost * float64(targetLevel)
}

func TotalBuildingCost(baseBuildCost float64, level int) float64 {
	if level < 1 {
		level = 1
	}
	return baseBuildCost * float64(level) * float64(level+1) / 2.0
}
