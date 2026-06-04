package service

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
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
// ── Map / slot config (extensible: add new entries to add new maps) ──

type mapSlotDef struct {
	mapID          string
	slotID         string
	unlockOrder    int
	mapUnlockLevel int
}

// knownMapIDs returns all valid map IDs.
func knownMapIDs() []string {
	seen := map[string]bool{}
	ids := make([]string, 0, 4)
	for _, s := range mapSlotDefs {
		if seen[s.mapID] {
			continue
		}
		seen[s.mapID] = true
		ids = append(ids, s.mapID)
	}
	return ids
}

func isValidMapID(mapID string) bool {
	for _, id := range knownMapIDs() {
		if id == mapID {
			return true
		}
	}
	return false
}

// mapUnlockLevel returns the level required to place buildings on this map.
func mapUnlockLevel(mapID string) int {
	for _, s := range mapSlotDefs {
		if s.mapID == mapID {
			return s.mapUnlockLevel
		}
	}
	return 1
}

var mapSlotDefs = []mapSlotDef{
	{mapID: "harbor", slotID: "harbor-plot-01", unlockOrder: 1, mapUnlockLevel: 1},
	{mapID: "harbor", slotID: "harbor-plot-02", unlockOrder: 2, mapUnlockLevel: 1},
	{mapID: "harbor", slotID: "harbor-plot-03", unlockOrder: 3, mapUnlockLevel: 1},
	{mapID: "harbor", slotID: "harbor-plot-04", unlockOrder: 4, mapUnlockLevel: 1},
	{mapID: "harbor", slotID: "harbor-plot-05", unlockOrder: 5, mapUnlockLevel: 1},
	{mapID: "harbor", slotID: "harbor-plot-06", unlockOrder: 6, mapUnlockLevel: 1},
	{mapID: "harbor", slotID: "harbor-plot-07", unlockOrder: 7, mapUnlockLevel: 1},
	{mapID: "harbor", slotID: "harbor-plot-08", unlockOrder: 8, mapUnlockLevel: 1},
	{mapID: "inland", slotID: "inland-plot-01", unlockOrder: 1, mapUnlockLevel: 5},
	{mapID: "inland", slotID: "inland-plot-02", unlockOrder: 2, mapUnlockLevel: 5},
	{mapID: "inland", slotID: "inland-plot-03", unlockOrder: 3, mapUnlockLevel: 5},
	{mapID: "inland", slotID: "inland-plot-04", unlockOrder: 4, mapUnlockLevel: 5},
	{mapID: "inland", slotID: "inland-plot-05", unlockOrder: 5, mapUnlockLevel: 5},
	{mapID: "inland", slotID: "inland-plot-06", unlockOrder: 6, mapUnlockLevel: 5},
	{mapID: "inland", slotID: "inland-plot-07", unlockOrder: 7, mapUnlockLevel: 5},
	{mapID: "inland", slotID: "inland-plot-08", unlockOrder: 8, mapUnlockLevel: 5},
	{mapID: "desert", slotID: "desert-plot-01", unlockOrder: 1, mapUnlockLevel: 10},
	{mapID: "desert", slotID: "desert-plot-02", unlockOrder: 2, mapUnlockLevel: 10},
	{mapID: "desert", slotID: "desert-plot-03", unlockOrder: 3, mapUnlockLevel: 10},
}
func (s *Service) PlaceBuilding(companyID int, buildingID string, mapID, slotID string, x, y int) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}
	if companyBuildingCount(company) > BuildingSlotsForLevel(company.Level) {
		return nil, fmt.Errorf("building limit reached: %d/%d slots used", companyBuildingCount(company), BuildingSlotsForLevel(company.Level))
	}
	mapID, slotID, x, y, err := validateMapPlacement(company.PlacedBuildings, "", company.Level, mapID, slotID, x, y)
	if err != nil {
		return nil, err
	}
	for i, b := range company.UnplacedBuildings {
		if b["id"] != buildingID {
			continue
		}
		b["mapId"] = mapID
		b["slotId"] = slotID
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

func (s *Service) MoveBuilding(companyID int, buildingID string, mapID, slotID string, x, y int) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}
	mapID, slotID, x, y, err := validateMapPlacement(company.PlacedBuildings, buildingID, company.Level, mapID, slotID, x, y)
	if err != nil {
		return nil, err
	}
	for _, b := range company.PlacedBuildings {
		if b["id"] == buildingID {
			b["mapId"] = mapID
			b["slotId"] = slotID
			b["x"] = x
			b["y"] = y
			s.saveCompanyLocked(company)
			return map[string]any{"building": b, "status": "moved"}, nil
		}
	}
	return nil, fmt.Errorf("building not found")
}

func validateMapPlacement(buildings []map[string]any, movingID string, level int, mapID, slotID string, x, y int) (string, string, int, int, error) {
	slot, err := normalizeMapSlot(mapID, slotID, x, y)
	if err != nil {
		return "", "", 0, 0, err
	}
	if err := validateSlotUnlocked(slot, level); err != nil {
		return "", "", 0, 0, err
	}
	legacyX, legacyY := legacyCoordsForOrder(slot.unlockOrder)
	for _, b := range buildings {
		if movingID != "" && fmt.Sprint(b["id"]) == movingID {
			continue
		}
		buildingMapID, buildingSlotID := buildingMapSlot(b)
		if buildingMapID == slot.mapID && buildingSlotID == slot.slotID {
			return "", "", 0, 0, fmt.Errorf("map position already occupied")
		}
	}
	return slot.mapID, slot.slotID, legacyX, legacyY, nil
}

func normalizeMapSlot(mapID, slotID string, x, y int) (mapSlotDef, error) {
	if !isValidMapID(mapID) {
		mapID = knownMapIDs()[0] // fallback to first known map (harbor)
	}
	if slotID == "" {
		slotID = legacySlotID(mapID, x, y)
	}
	for _, slot := range mapSlotDefs {
		if slot.mapID == mapID && slot.slotID == slotID {
			return slot, nil
		}
	}
	return mapSlotDef{}, fmt.Errorf("map position out of bounds")
}

func validateSlotUnlocked(slot mapSlotDef, level int) error {
	requiredLevel := slot.mapUnlockLevel
	if slot.unlockOrder > 3 {
		requiredLevel += (slot.unlockOrder - 3) * 2
	}
	if level < requiredLevel {
		if requiredLevel == slot.mapUnlockLevel && slot.mapUnlockLevel > 1 {
			return fmt.Errorf("%s map unlocks at level %d", slot.mapID, slot.mapUnlockLevel)
		}
		return fmt.Errorf("map plot unlocks at level %d", requiredLevel)
	}
	return nil
}

func buildingMapSlot(building map[string]any) (string, string) {
	mapID := fmt.Sprint(building["mapId"])
	if !isValidMapID(mapID) {
		mapID = knownMapIDs()[0]
	}
	slotID := fmt.Sprint(building["slotId"])
	if slotID == "" || slotID == "<nil>" {
		slotID = legacySlotID(mapID, intFromAny(building["x"]), intFromAny(building["y"]))
	}
	return mapID, slotID
}

func legacySlotID(mapID string, x, y int) string {
	if x < 1 {
		x = 1
	}
	if y < 1 {
		y = 1
	}
	order := (y-1)*3 + x
	if order < 1 {
		order = 1
	}
	if order > 8 {
		order = 8
	}
	return fmt.Sprintf("%s-plot-%02d", mapID, order)
}

func legacyCoordsForOrder(order int) (int, int) {
	return ((order - 1) % 3) + 1, ((order - 1) / 3) + 1
}

func slotOrderFromID(slotID string) int {
	parts := strings.Split(slotID, "-plot-")
	if len(parts) != 2 {
		return 1
	}
	order, err := strconv.Atoi(parts[1])
	if err == nil && order > 0 {
		return order
	}
	return 1
}

func firstOpenSlotID(mapID string, occupied map[string]bool) string {
	for _, slot := range mapSlotDefs {
		if slot.mapID != mapID {
			continue
		}
		key := slot.mapID + ":" + slot.slotID
		if !occupied[key] {
			return slot.slotID
		}
	}
	return legacySlotID(mapID, 1, 1)
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
