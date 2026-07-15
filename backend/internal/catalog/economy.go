package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// EconomyModelEntry mirrors one state_1 entry in decompiled/data/economy_model.json.
type EconomyModelEntry struct {
	ModeledProductionCostPerUnit       float64 `json:"modeledProductionCostPerUnit"`
	ModeledWagesPerUnitPerHour         float64 `json:"modeledWagesPerUnitPerHour"`
	ModeledStoreWages                  float64 `json:"modeledStoreWages"`
	ModeledSalesPerUnitPerHour         float64 `json:"modeledSalesPerUnitPerHour"`
	ModeledUnitsSoldAnHour             float64 `json:"modeledUnitsSoldAnHour"`
	BuildingLevelsNeededPerUnitPerHour float64 `json:"buildingLevelsNeededPerUnitPerHour"`
	BuildingKindModifier               float64 `json:"buildingKindModifier"`
}

type economyModelWrapper struct {
	Models map[string]map[string]*EconomyModelEntry `json:"models"`
}

// LoadEconomyModel reads economy_model.json from the project's decompiled/data directory.
// The file format is {"models": {"1": {"state_1": { ... }}, ...}}
// Returns a map keyed by resource ID (int).
func LoadEconomyModel(projectRoot string) (map[int]*EconomyModelEntry, error) {
	path := filepath.Join(projectRoot, "decompiled", "data", "economy_model.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read economy_model.json: %w", err)
	}
	var wrapper economyModelWrapper
	if err := json.Unmarshal(raw, &wrapper); err != nil {
		return nil, fmt.Errorf("unmarshal economy_model.json: %w", err)
	}
	byID := make(map[int]*EconomyModelEntry, len(wrapper.Models))
	for key, states := range wrapper.Models {
		id, err := strconv.Atoi(key)
		if err != nil {
			return nil, fmt.Errorf("parse economy model key %q: %w", key, err)
		}
		entry, ok := states["state_1"]
		if !ok {
			return nil, fmt.Errorf("economy model resource %d missing state_1", id)
		}
		byID[id] = entry
	}
	return byID, nil
}
