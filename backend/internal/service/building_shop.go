package service

import (
	"fmt"
	"time"
)

func (s *Service) BuildingMarket() []map[string]any {
	return []map[string]any{
		{
			"id": "b-shop-1", "name": "Farm", "kind": 1, "cost": 10000.0,
			"description":     "Produces Wheat at 100 units per hour at level 1.",
			"produces":        []int{1},
			"starterProduces": []int{1},
			"starterRole":     "Start here: Farm -> Wheat",
		},
		{
			"id": "b-shop-2", "name": "Mill", "kind": 2, "cost": 20000.0,
			"description":     "Mills Wheat into Flour at 80 units per hour at level 1.",
			"produces":        []int{2},
			"starterProduces": []int{2},
			"starterRole":     "Then process Wheat -> Flour",
		},
		{
			"id": "b-shop-3", "name": "Bakery", "kind": 3, "cost": 30000.0,
			"description":     "Bakes Flour into Bread at 40 units per hour at level 1.",
			"produces":        []int{3},
			"starterProduces": []int{3},
			"starterRole":     "Then bake Flour -> Bread",
		},
		{
			"id": "b-shop-4", "name": "Restaurant", "kind": 4, "cost": 40000.0,
			"description":     "Serves Bread as Meals at 30 units per hour at level 1.",
			"produces":        []int{4},
			"starterProduces": []int{4},
			"starterRole":     "Finally serve Bread -> Meals",
		},
	}
}

func (s *Service) BuyBuilding(companyID int, buildingID string) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}
	market := s.BuildingMarket()
	var item map[string]any
	for _, b := range market {
		if b["id"] == buildingID {
			item = b
			break
		}
	}
	if item == nil {
		return nil, fmt.Errorf("building not found")
	}
	cost := floatFromAny(item["cost"])
	if company.Money < cost {
		return nil, fmt.Errorf("not enough money: need %.0f, have %.0f", cost, company.Money)
	}
	// Check warehouse capacity
	cap := (company.WarehouseLevel + 1) * 1000
	used := len(company.PlacedBuildings) + len(company.UnplacedBuildings)
	if used >= cap {
		return nil, fmt.Errorf("warehouse full: %d/%d slots used", used, cap)
	}
	company.Money -= cost
	b := map[string]any{
		"id":       fmt.Sprintf("b-%d", s.now().UnixNano()),
		"kind":     intFromAny(item["kind"]),
		"level":    1,
		"name":     item["name"],
		"baseCost": intFromAny(item["cost"]),
		"busy":     false,
		"placed":   false,
		"produces": item["produces"],
	}
	if starterProduces, ok := item["starterProduces"]; ok {
		b["starterProduces"] = starterProduces
	}
	if starterRole, ok := item["starterRole"]; ok {
		b["starterRole"] = starterRole
	}
	company.UnplacedBuildings = append(company.UnplacedBuildings, b)
	s.addLedger("buy_building", -cost, "out", map[string]any{"buildingId": b["id"]})
	s.saveCompanyLocked(company)
	return map[string]any{"building": b, "cost": cost}, nil
}

func (s *Service) PlaceBuilding(companyID int, buildingID string, x, y int) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}
	for i, b := range company.UnplacedBuildings {
		if b["id"] != buildingID {
			continue
		}
		b["x"] = x
		b["y"] = y
		b["placedAt"] = s.now().UTC().Format(time.RFC3339)
		company.PlacedBuildings = append(company.PlacedBuildings, b)
		company.UnplacedBuildings = append(company.UnplacedBuildings[:i], company.UnplacedBuildings[i+1:]...)
		s.saveCompanyLocked(company)
		return map[string]any{"building": b, "status": "placed"}, nil
	}
	return nil, fmt.Errorf("unplaced building not found")
}

func (s *Service) MoveBuilding(companyID int, buildingID string, x, y int) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}
	for _, b := range company.PlacedBuildings {
		if b["id"] == buildingID {
			b["x"] = x
			b["y"] = y
			s.saveCompanyLocked(company)
			return map[string]any{"building": b, "status": "moved"}, nil
		}
	}
	return nil, fmt.Errorf("building not found")
}

func (s *Service) DemolishBuilding(companyID int, buildingID string) (map[string]any, error) {
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
		refund := float64(intFromAny(b["baseCost"])) * 0.5
		company.Money += refund
		company.PlacedBuildings = append(company.PlacedBuildings[:i], company.PlacedBuildings[i+1:]...)
		s.addLedger("demolish_building", refund, "in", map[string]any{"buildingId": buildingID})
		s.saveCompanyLocked(company)
		return map[string]any{"refund": refund, "status": "demolished"}, nil
	}
	return nil, fmt.Errorf("building not found")
}

func (s *Service) WarehouseUpgrade(companyID int) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}
	lvl := company.WarehouseLevel
	cost := float64(lvl+1) * s.Cfg.Game.WarehouseUpgradeCost
	if company.Money < cost {
		return nil, fmt.Errorf("not enough money: need %.0f, have %.0f", cost, company.Money)
	}
	company.Money -= cost
	company.WarehouseLevel++
	newCap := (company.WarehouseLevel + 1) * 1000
	s.addLedger("warehouse_upgrade", -cost, "out", map[string]any{"newLevel": company.WarehouseLevel})
	s.saveCompanyLocked(company)
	return map[string]any{"level": company.WarehouseLevel, "capacity": newCap, "cost": cost}, nil
}
