package catalog_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/beiwater/NewHaven/backend/internal/catalog"
)

// TestLoadResources_KeysByDbLetter verifies that resources are keyed by dbLetter
// when dbLetter > 0, falling back to id when dbLetter is 0.
func TestLoadResources_KeysByDbLetter(t *testing.T) {
	// Create a temporary directory with a minimal resources.json that has
	// entries with dbLetter matching id, and one where they could differ.
	dir := t.TempDir()
	dataDir := filepath.Join(dir, "decompiled", "data")
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	// Write a resources.json with:
	// - Entry 1: id=1, dbLetter=1 (normal)
	// - Entry 2: id=2, dbLetter=2 (normal)
	// - Entry 3: id=99, dbLetter=0 (dbLetter 0 should fall back to id)
	content := `[
		{"id":1,"dbLetter":1,"name":"Grain","producedFrom":{},"producedPerHourRaw":500,"producedAnHour":500,"unitsSoldAnHour":150,"isExchangeTradable":true,"hasEconomyModel":true,"basePrice":23},
		{"id":2,"dbLetter":2,"name":"Dairy Milk","producedFrom":{},"producedPerHourRaw":420,"producedAnHour":420,"unitsSoldAnHour":130,"isExchangeTradable":true,"hasEconomyModel":true,"basePrice":28},
		{"id":99,"dbLetter":0,"name":"NoDbLetter","producedFrom":{},"producedPerHourRaw":100,"producedAnHour":100,"unitsSoldAnHour":50,"isExchangeTradable":false,"hasEconomyModel":false,"basePrice":10}
	]`
	if err := os.WriteFile(filepath.Join(dataDir, "resources.json"), []byte(content), 0644); err != nil {
		t.Fatalf("write resources.json: %v", err)
	}

	resources, err := catalog.LoadResources(dir)
	if err != nil {
		t.Fatalf("LoadResources: %v", err)
	}

	// Entry with id=1, dbLetter=1 -> keyed by 1
	r, ok := resources[1]
	if !ok {
		t.Fatal("expected resource keyed by 1 (dbLetter=1)")
	}
	if r.Name != "Grain" {
		t.Errorf("expected Grain, got %s", r.Name)
	}

	// Entry with id=2, dbLetter=2 -> keyed by 2
	r, ok = resources[2]
	if !ok {
		t.Fatal("expected resource keyed by 2 (dbLetter=2)")
	}
	if r.Name != "Dairy Milk" {
		t.Errorf("expected Dairy Milk, got %s", r.Name)
	}

	// Entry with id=99, dbLetter=0 -> keyed by id=99 (fallback)
	r, ok = resources[99]
	if !ok {
		t.Fatal("expected resource keyed by 99 (fallback id when dbLetter=0)")
	}
	if r.Name != "NoDbLetter" {
		t.Errorf("expected NoDbLetter, got %s", r.Name)
	}

	// Should NOT be keyed by the id that wasn't dbLetter
	// (For the first two, id == dbLetter so both keys coincidentally match)
	// For the dbLetter=0 entry, id=99 is the only key.
	// The dbLetter=0 entry should NOT also be at key 0.
	if _, ok := resources[0]; ok {
		t.Error("expected no resource keyed by 0 (dbLetter=0 should not produce key 0)")
	}
}

// TestLoadResources_RealKeys verifies that loading the actual resources.json
// produces a well-populated map (smoke test).
func TestLoadResources_RealKeys(t *testing.T) {
	// Walk up from working dir to find the project root (has decompiled/data/).
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Getwd: %v", err)
	}
	projectRoot := cwd
	for {
		if _, err := os.Stat(filepath.Join(projectRoot, "decompiled", "data")); err == nil {
			break
		}
		parent := filepath.Dir(projectRoot)
		if parent == projectRoot {
			t.Skip("project root not found (no decompiled/data/)")
		}
		projectRoot = parent
	}
	resources, err := catalog.LoadResources(projectRoot)
	if err != nil {
		t.Fatalf("LoadResources: %v", err)
	}
	if len(resources) == 0 {
		t.Fatal("expected some resources to be loaded")
	}
	// In real data, dbLetter matches id for all entries.
	// Verify that key 1 exists and has the expected name.
	r, ok := resources[1]
	if !ok {
		t.Fatal("expected resource with key 1 (Grain)")
	}
	if r.Name != "Grain" {
		t.Errorf("expected name Grain, got %s", r.Name)
	}
}

// TestLoadEconomyModel_RealKeys verifies that the actual economy_model.json
// loads successfully and contains the expected number of resource entries (12).
func TestLoadEconomyModel_RealKeys(t *testing.T) {
	projectRoot := findProjectRoot()
	economy, err := catalog.LoadEconomyModel(projectRoot)
	if err != nil {
		t.Fatalf("LoadEconomyModel: %v", err)
	}
	if len(economy) != 12 {
		t.Errorf("expected 12 economy model entries, got %d", len(economy))
	}
	// Spot-check a few well-known entries.
	for _, id := range []int{1, 3, 6, 12} {
		if _, ok := economy[id]; !ok {
			t.Errorf("missing economy model entry for resource %d", id)
		}
	}
}
