package catalog_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/beiwater/NewHaven/backend/internal/catalog"
)

func TestStaticDataManifestLoads(t *testing.T) {
	root := findProjectRoot()
	m, err := catalog.LoadManifest(root)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	if m.Version == "" {
		t.Error("manifest version is empty")
	}
	if len(m.Files) == 0 {
		t.Error("manifest has no file entries")
	}
}

func TestStaticDataManifestEntriesExist(t *testing.T) {
	root := findProjectRoot()
	m, err := catalog.LoadManifest(root)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	for _, f := range m.Files {
		path := filepath.Join(root, f.Path)
		info, err := os.Stat(path)
		if err != nil {
			t.Errorf("entry not found: %s: %v", f.Path, err)
			continue
		}
		if info.IsDir() {
			t.Errorf("entry is a directory: %s", f.Path)
		}
	}
}

func TestStaticDataManifestIntegrity(t *testing.T) {
	root := findProjectRoot()
	err := catalog.ValidateManifest(root)
	if err != nil {
		t.Errorf("integrity check failed: %v", err)
	}
}

func TestManifestRequiredPaths(t *testing.T) {
	root := findProjectRoot()
	m, err := catalog.LoadManifest(root)
	if err != nil {
		t.Fatalf("LoadManifest: %v", err)
	}
	required := map[string]bool{
		"decompiled/data/resources.json":        false,
		"decompiled/data/buildings.json":        false,
		"decompiled/data/economy_model.json":    false,
		"decompiled/data/resource_lookups.json": false,
		"backend/configs/game.json":             false,
	}
	if len(m.Files) != len(required) {
		t.Fatalf("static data manifest has %d entries; want exactly %d", len(m.Files), len(required))
	}
	for _, f := range m.Files {
		if _, ok := required[f.Path]; !ok {
			t.Errorf("unexpected static data manifest path: %s", f.Path)
			continue
		}
		required[f.Path] = true
	}
	for path, found := range required {
		if !found {
			t.Errorf("required path missing: %s", path)
		}
	}
}

// findProjectRoot walks up from wd looking for decompiled/data directory.
func findProjectRoot() string {
	wd, _ := os.Getwd()
	dir := wd
	for {
		if _, err := os.Stat(filepath.Join(dir, "decompiled", "data")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return wd
		}
		dir = parent
	}
}
