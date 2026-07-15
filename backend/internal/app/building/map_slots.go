package building

import (
	"fmt"

	domain "github.com/beiwater/NewHaven/backend/internal/domain/company"
)

// -- Map / slot config (extensible: add new entries to add new maps) --

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

// normalizeMapSlot normalizes mapID and slotID. Invalid or empty mapID falls
// back to the first known map (harbor). An empty slotID is derived from x,y
// using the legacy 3-column grid formula.
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

// validateSlotUnlocked checks whether the given slot is available at the
// given company level. First 3 slots unlock at the map's base unlock level;
// each subsequent slot requires 2 additional levels.
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

// buildingMapSlot extracts canonical mapID and slotID from a domain.Building,
// normalizing empty or invalid values the same way as the legacy code.
func buildingMapSlot(b domain.Building) (string, string) {
	mapID := b.MapID
	if !isValidMapID(mapID) {
		mapID = knownMapIDs()[0]
	}
	slotID := b.SlotID
	if slotID == "" {
		slotID = legacySlotID(mapID, b.X, b.Y)
	}
	return mapID, slotID
}

// legacySlotID derives a canonical slot ID from x,y coordinates using the
// legacy 3-column grid formula: order = (y-1)*3 + x.
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

// legacyCoordsForOrder returns canonical (x,y) coordinates for a slot's
// unlock order, using the legacy 3-column grid.
func legacyCoordsForOrder(order int) (int, int) {
	return ((order - 1) % 3) + 1, ((order - 1) / 3) + 1
}

// validateMapPlacement validates and normalizes a building placement/move
// request. It returns the canonical mapID, slotID, x, y coordinates after
// normalization, or an error if the position is invalid or occupied.
func validateMapPlacement(buildings []domain.Building, movingID string, level int, mapID, slotID string, x, y int) (string, string, int, int, error) {
	slot, err := normalizeMapSlot(mapID, slotID, x, y)
	if err != nil {
		return "", "", 0, 0, err
	}
	if err := validateSlotUnlocked(slot, level); err != nil {
		return "", "", 0, 0, err
	}
	legacyX, legacyY := legacyCoordsForOrder(slot.unlockOrder)
	for _, b := range buildings {
		if movingID != "" && b.ID == movingID {
			continue
		}
		// Skip buildings that aren't actually placed on any map.
		if b.MapID == "" && b.SlotID == "" {
			continue
		}
		buildingMapID, buildingSlotID := buildingMapSlot(b)
		if buildingMapID == slot.mapID && buildingSlotID == slot.slotID {
			return "", "", 0, 0, fmt.Errorf("map position already occupied")
		}
	}
	return slot.mapID, slot.slotID, legacyX, legacyY, nil
}
