// Package catalog loads static game data from decompiled JSON files.
// It is a small, self-contained loader; not a broad engine abstraction.
package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ResourceEntry mirrors one entry in decompiled/data/resources.json.
type ResourceEntry struct {
	ID                 int         `json:"id"`
	DbLetter           int         `json:"dbLetter"`
	Name               string      `json:"name"`
	Category           string      `json:"category"`
	Tier               int         `json:"tier"`
	ProducedFrom       map[int]int `json:"producedFrom"` // resourceID -> amount per unit
	ProducedPerHourRaw int         `json:"producedPerHourRaw"`
	ProducedAnHour     int         `json:"producedAnHour"`
	UnitsSoldAnHour    int         `json:"unitsSoldAnHour"`
	IsExchangeTradable bool        `json:"isExchangeTradable"`
	HasEconomyModel    bool        `json:"hasEconomyModel"`
	BasePrice          float64     `json:"basePrice"`
}

// BuildingEntry mirrors one entry in decompiled/data/buildings.json.
type BuildingEntry struct {
	ID                int    `json:"id"`
	Kind              int    `json:"kind"`
	Name              string `json:"name"`
	Type              string `json:"type"`
	BaseCost          int    `json:"baseCost"`
	BaseOutput        int    `json:"baseOutput"`
	BaseOutputPerHour int    `json:"baseOutputPerHourLv1"`
	Produces          []int  `json:"produces"` // resource IDs this building can produce
	Description       string `json:"description"`
}

// LoadResources reads resources.json from the project's decompiled/data directory.
func LoadResources(projectRoot string) (map[int]*ResourceEntry, error) {
	path := filepath.Join(projectRoot, "decompiled", "data", "resources.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read resources.json: %w", err)
	}
	var list []*ResourceEntry
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, fmt.Errorf("unmarshal resources.json: %w", err)
	}
	byID := make(map[int]*ResourceEntry, len(list))
	for _, r := range list {
		key := r.ID
		if r.DbLetter > 0 {
			key = r.DbLetter
		}
		byID[key] = r
	}
	return byID, nil
}

// LoadBuildings reads buildings.json from the project's decompiled/data directory.
// The JSON is keyed by building ID string; it returns a map keyed by numeric ID.
func LoadBuildings(projectRoot string) (map[int]*BuildingEntry, error) {
	path := filepath.Join(projectRoot, "decompiled", "data", "buildings.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read buildings.json: %w", err)
	}
	var rawMap map[string]*BuildingEntry
	if err := json.Unmarshal(raw, &rawMap); err != nil {
		return nil, fmt.Errorf("unmarshal buildings.json: %w", err)
	}
	byID := make(map[int]*BuildingEntry, len(rawMap))
	for key, b := range rawMap {
		var id int
		if _, err := fmt.Sscanf(key, "%d", &id); err != nil {
			return nil, fmt.Errorf("parse building key %q: %w", key, err)
		}
		byID[id] = b
	}
	return byID, nil
}
