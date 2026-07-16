package formula

import "time"

// UpgradeDuration returns the downtime for upgrading a building from
// currentLevel to the next level. Each building family has its own base
// construction time, while later levels deliberately take longer so an
// established town cannot be upgraded instantly.
//
// The first five upgrades use 1x, 2x, 3x, 6x, and 10x of the base time.
// From level 6 onward the multiplier continues as triangular numbers
// (15x, 21x, ...). This keeps the rule predictable without imposing a hard
// level cap on future content.
func UpgradeDuration(buildingKind, currentLevel int) time.Duration {
	baseMinutes := upgradeBaseMinutes(buildingKind)
	if currentLevel < 1 {
		currentLevel = 1
	}

	multiplier := 1
	switch currentLevel {
	case 1:
		multiplier = 1
	case 2:
		multiplier = 2
	case 3:
		multiplier = 3
	case 4:
		multiplier = 6
	default:
		// 5 -> 6 is 10x; 6 -> 7 is 15x, and so on.
		multiplier = currentLevel * (currentLevel - 1) / 2
	}
	return time.Duration(baseMinutes*multiplier) * time.Minute
}

func upgradeBaseMinutes(buildingKind int) int {
	// Keep early construction short enough for a casual session while still
	// letting each building feel distinct. Unknown future buildings get a
	// conservative four-minute base rather than silently becoming instant.
	if minutes, ok := map[int]int{
		1:  2, // Farm
		2:  3, // Mill
		3:  3, // Bakery
		4:  4, // Kitchen
		5:  5,
		6:  3,
		7:  4,
		8:  4,
		9:  5,
		10: 6,
		11: 6,
		12: 5,
	}[buildingKind]; ok {
		return minutes
	}
	return 4
}
