package formula

const (
	MinProductQuality = 0
	MaxProductQuality = 12

	// Each quality level raises retail demand speed by two percent. Quality
	// changes how quickly a committed batch sells; it does not bypass price
	// sensitivity or increase the player's chosen sale price.
	RetailQualityBonusPerLevel = 0.02

	// Producing above Q0 consumes twice the normal lower-quality inputs. Raw
	// products use two units of their own previous quality per output unit.
	QualityInputMultiplier = 2
)

// ValidProductQuality reports whether quality is one of the 13 playable tiers.
func ValidProductQuality(quality int) bool {
	return quality >= MinProductQuality && quality <= MaxProductQuality
}

// RetailQualitySpeedMultiplier returns 1.00 at Q0 and 1.24 at Q12.
func RetailQualitySpeedMultiplier(quality int) float64 {
	if quality < MinProductQuality {
		quality = MinProductQuality
	}
	if quality > MaxProductQuality {
		quality = MaxProductQuality
	}
	return 1 + float64(quality)*RetailQualityBonusPerLevel
}
