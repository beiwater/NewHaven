package service

import (
	"fmt"
	"time"
)

func (s *Service) BuildingMarket() []map[string]any {
	return []map[string]any{
		{
			"id": "b-shop-1", "name": "Farm Plot", "kind": 1, "cost": 50000.0,
			"description":     "Grows staples and crops",
			"produces":        []int{3, 4, 5, 6, 66, 72, 120},
			"starterProduces": []int{66, 6},
			"starterRole":     "Start here: Water -> Seeds -> Grain",
		},
		{
			"id": "b-shop-2", "name": "Food Factory", "kind": 2, "cost": 120000.0,
			"description":     "Processes farm goods into food products",
			"produces":        []int{7, 8, 9, 121, 122, 127, 133, 134, 135, 137, 139, 141},
			"starterProduces": []int{133},
			"starterRole":     "Then process Grain -> Flour",
		},
		{
			"id": "b-shop-3", "name": "Warehouse", "kind": 3, "cost": 30000.0,
			"description": "Storage building",
			"produces":    []int{},
			"starterRole": "Stores Seeds, Grain, and Flour while the chain runs",
		},
		{
			"id": "b-shop-4", "name": "Research Lab", "kind": 4, "cost": 80000.0,
			"description": "Research new technologies",
			"produces":    []int{},
			"starterRole": "Future unlock point after the starter chain is stable",
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
