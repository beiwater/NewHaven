package formula

// ph(adminOverhead, cooSkill) = adminOverhead - (adminOverhead - 1) * cooSkill / 100
func AdminOverheadWithCOO(adminOverhead float64, cooSkill float64) float64 {
	return adminOverhead - (adminOverhead-1.0)*cooSkill/100.0
}

func CTOProductionMultiplier(ctoSkill float64) float64 {
	// hb = 2 (% per skill level)
	return (100.0 + ctoSkill*2.0) / 100.0
}
