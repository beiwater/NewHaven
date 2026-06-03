package data

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type StaticData struct {
	Resources       []map[string]any
	Buildings       map[string]any
	EconomyModel    map[string]any
	ResourceLookups map[string]any
}

func readJSON(path string, target any) error {
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, target)
}

func Load(root string) (*StaticData, error) {
	base := filepath.Join(root, "decompiled", "data")
	result := &StaticData{}
	if err := readJSON(filepath.Join(base, "resources.json"), &result.Resources); err != nil {
		return nil, fmt.Errorf("load resources.json: %w", err)
	}
	if err := readJSON(filepath.Join(base, "buildings.json"), &result.Buildings); err != nil {
		return nil, fmt.Errorf("load buildings.json: %w", err)
	}
	if err := readJSON(filepath.Join(base, "economy_model.json"), &result.EconomyModel); err != nil {
		return nil, fmt.Errorf("load economy_model.json: %w", err)
	}
	if err := readJSON(filepath.Join(base, "resource_lookups.json"), &result.ResourceLookups); err != nil {
		return nil, fmt.Errorf("load resource_lookups.json: %w", err)
	}
	return result, nil
}
