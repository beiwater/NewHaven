package service

import (
	"fmt"
	"sort"
	"time"
)

func (s *Service) BuildingMarket() []map[string]any {
	items := make([]map[string]any, 0, len(s.Data.Buildings))
	keys := make([]int, 0, len(s.Data.Buildings))
	for k := range s.Data.Buildings {
		keys = append(keys, intFromAny(k))
	}
	sort.Ints(keys)
	for _, key := range keys {
		raw, ok := s.Data.Buildings[fmt.Sprint(key)].(map[string]any)
		if !ok {
			continue
		}
		kind := intFromAny(raw["kind"])
		if kind <= 0 {
			kind = intFromAny(raw["id"])
		}
		produces := intSliceFromAny(raw["produces"])
		if len(produces) == 0 {
			if output := intFromAny(raw["outputResourceId"]); output > 0 {
				produces = []int{output}
			}
		}
		item := map[string]any{
			"id":          fmt.Sprintf("b-shop-%d", kind),
			"name":        raw["name"],
			"kind":        kind,
			"cost":        floatFromAny(raw["baseCost"]),
			"unlockLevel": BuildingUnlockLevel(kind),
			"description": raw["description"],
			"produces":    produces,
		}
		if item["cost"] == 0.0 {
			item["cost"] = float64(s.Cfg.Game.BaseBuildingCost + kind*10000)
		}
		if starterProduces := intSliceFromAny(raw["starterProduces"]); len(starterProduces) > 0 {
			item["starterProduces"] = starterProduces
		}
		if starterRole, ok := raw["starterRole"]; ok {
			item["starterRole"] = starterRole
		}
		items = append(items, item)
	}
	if len(items) == 0 {
		for _, resource := range s.Data.Resources {
			id := intFromAny(resource["dbLetter"])
			if id <= 0 {
				id = intFromAny(resource["id"])
			}
			if id <= 0 {
				continue
			}
			items = append(items, map[string]any{
				"id":              fmt.Sprintf("b-shop-%d", id),
				"name":            fmt.Sprintf("%s Producer", resource["name"]),
				"kind":            id,
				"cost":            float64(s.Cfg.Game.BaseBuildingCost + id*10000),
				"unlockLevel":     BuildingUnlockLevel(id),
				"description":     fmt.Sprintf("Produces %s.", resource["name"]),
				"produces":        []int{id},
				"starterProduces": []int{id},
			})
		}
	}
	return items
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
	kind := intFromAny(item["kind"])
	if company.Level < BuildingUnlockLevel(kind) {
		return nil, fmt.Errorf("building unlocks at level %d", BuildingUnlockLevel(kind))
	}
	maxBuildings := BuildingSlotsForLevel(company.Level)
	usedBuildings := companyBuildingCount(company)
	if usedBuildings >= maxBuildings {
		return nil, fmt.Errorf("building limit reached: %d/%d slots used", usedBuildings, maxBuildings)
	}
	cost := floatFromAny(item["cost"])
	if company.Money < cost {
		return nil, fmt.Errorf("not enough money: need %.0f, have %.0f", cost, company.Money)
	}
	company.Money -= cost
	b := map[string]any{
		"id":       fmt.Sprintf("b-%d", s.now().UnixNano()),
		"kind":     kind,
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
	s.addLedger("buy_building", cost, "out", map[string]any{"buildingId": b["id"]})
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
	if companyBuildingCount(company) > BuildingSlotsForLevel(company.Level) {
		return nil, fmt.Errorf("building limit reached: %d/%d slots used", companyBuildingCount(company), BuildingSlotsForLevel(company.Level))
	}
	if err := validateMapPlacement(company.PlacedBuildings, "", x, y); err != nil {
		return nil, err
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
	if err := validateMapPlacement(company.PlacedBuildings, buildingID, x, y); err != nil {
		return nil, err
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

func validateMapPlacement(buildings []map[string]any, movingID string, x, y int) error {
	if x < 1 || x > 12 || y < 1 || y > 10 {
		return fmt.Errorf("map position out of bounds")
	}
	for _, b := range buildings {
		if movingID != "" && fmt.Sprint(b["id"]) == movingID {
			continue
		}
		if intFromAny(b["x"]) == x && intFromAny(b["y"]) == y {
			return fmt.Errorf("map position already occupied")
		}
	}
	return nil
}

func intSliceFromAny(v any) []int {
	switch t := v.(type) {
	case []int:
		return t
	case []any:
		out := make([]int, 0, len(t))
		for _, item := range t {
			if n := intFromAny(item); n > 0 {
				out = append(out, n)
			}
		}
		return out
	default:
		return []int{}
	}
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
	newCap := s.WarehouseCapacity(company.WarehouseLevel)
	s.addLedger("warehouse_upgrade", cost, "out", map[string]any{"newLevel": company.WarehouseLevel})
	s.saveCompanyLocked(company)
	return map[string]any{"level": company.WarehouseLevel, "capacity": newCap, "cost": cost}, nil
}

func (s *Service) WarehouseCapacity(level int) int {
	if level < 0 {
		level = 0
	}
	base := s.Cfg.Game.WarehouseBaseCap
	if base <= 0 {
		base = 1000
	}
	return (level + 2) * base
}
