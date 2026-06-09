package formula

// FeatureUnlockLevels defines which features unlock at which player level.
var FeatureUnlockLevels = map[string]int{
	"map": 1, "build": 1, "warehouse": 1, "leaderboard": 1,
	"market": 2, "contracts": 3, "research": 4,
	"executives": 5, "finance": 6,
}

// BuildingUnlockLevels defines which building IDs unlock at which player level.
var BuildingUnlockLevels = map[int]int{
	1: 1, 2: 2, 3: 2, 4: 3, 5: 4,
	6: 5, 7: 6, 8: 7, 9: 8, 10: 9, 11: 10, 12: 11,
}

// XpToNextLevel returns XP needed to reach the next level.
func XpToNextLevel(level int) int {
	if level < 1 {
		return 100
	}
	return level * 100
}

// BuildingSlotsForLevel returns how many building slots a company has at a given level.
func BuildingSlotsForLevel(level int) int {
	return 2 + level/2
}

// FeatureUnlocksAtLevel returns features and their unlock levels for the given player level.
func FeatureUnlocksAtLevel(level int) map[string]any {
	features := make(map[string]bool)
	featureLevels := make(map[string]int)
	for f, lvl := range FeatureUnlockLevels {
		featureLevels[f] = lvl
		features[f] = level >= lvl
	}
	return map[string]any{
		"features":      features,
		"featureLevels": featureLevels,
	}
}
