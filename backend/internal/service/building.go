package service

import "fmt"

func (s *Service) UpgradeBuilding(companyID int, buildingID string) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}
	for i, b := range company.PlacedBuildings {
		if b["id"] != buildingID {
			continue
		}
		currLevel := intFromAny(b["level"])
		if currLevel <= 0 {
			currLevel = 1
		}
		nextLevel := currLevel + 1
		baseCost := intFromAny(b["baseCost"])
		if baseCost <= 0 {
			baseCost = intFromAny(b["kind"]) * 5000
		}
		// Cost to go from LvN to Lv(N+1) = (N+1) × baseCost
		cost := float64(nextLevel) * float64(baseCost)
		if company.Money < cost {
			return nil, fmt.Errorf("need %.0f to upgrade to level %d, have %.0f", cost, nextLevel, company.Money)
		}
		company.Money -= cost
		company.PlacedBuildings[i]["level"] = float64(nextLevel)
		s.addLedger("building_upgrade", cost, "out", map[string]any{"buildingId": buildingID, "oldLevel": currLevel, "newLevel": nextLevel})
		s.saveCompanyLocked(company)
		return map[string]any{
			"buildingId":       buildingID,
			"oldLevel":         currLevel,
			"newLevel":         nextLevel,
			"cost":             cost,
			"outputMultiplier": nextLevel,
		}, nil
	}
	return nil, fmt.Errorf("building not found")
}
