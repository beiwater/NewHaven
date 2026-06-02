package service

import "go-sim-api/internal/model"

var featureUnlockLevels = map[string]int{
	"map":         1,
	"build":       1,
	"warehouse":   1,
	"leaderboard": 1,
	"market":      2,
	"contracts":   3,
	"research":    4,
	"executives":  5,
	"finance":     6,
}

var buildingUnlockLevels = map[int]int{
	1: 1, // Farm
	2: 2, // Mill
	3: 4, // Bakery
	4: 6, // Restaurant
}

func BuildingSlotsForLevel(level int) int {
	if level < 1 {
		level = 1
	}
	return 2 + level/2
}

func BuildingUnlockLevel(kind int) int {
	if level, ok := buildingUnlockLevels[kind]; ok {
		return level
	}
	return 1
}

func FeatureUnlockPayload(level int) map[string]any {
	features := map[string]bool{}
	featureLevels := map[string]int{}
	for feature, unlockLevel := range featureUnlockLevels {
		features[feature] = level >= unlockLevel
		featureLevels[feature] = unlockLevel
	}
	return map[string]any{
		"features":      features,
		"featureLevels": featureLevels,
	}
}

func companyBuildingCount(company *model.Company) int {
	if company == nil {
		return 0
	}
	return len(company.PlacedBuildings) + len(company.UnplacedBuildings)
}
