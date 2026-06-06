package formula_test

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type sourceManifest struct {
	Version string `json:"version"`
	Files   []struct {
		Path   string `json:"path"`
		SHA256 string `json:"sha256"`
	} `json:"files"`
}

func TestLegacySourceManifest(t *testing.T) {
	root := findProjectRoot(t)
	raw, err := os.ReadFile(filepath.Join(root, "backend-next", "internal", "formula", "legacy-source-manifest.json"))
	if err != nil {
		t.Fatalf("read legacy source manifest: %v", err)
	}

	var manifest sourceManifest
	if err := json.Unmarshal(raw, &manifest); err != nil {
		t.Fatalf("unmarshal legacy source manifest: %v", err)
	}
	if manifest.Version == "" {
		t.Fatal("legacy source manifest version is empty")
	}

	required := map[string]bool{
		"backend/internal/formula/market.go":     false,
		"backend/internal/formula/bonds.go":      false,
		"backend/internal/formula/production.go": false,
		"backend/internal/formula/costs.go":      false,
		"backend/internal/formula/saturation.go": false,
	}
	if len(manifest.Files) != len(required) {
		t.Fatalf("legacy source manifest has %d entries; want exactly %d", len(manifest.Files), len(required))
	}

	for _, entry := range manifest.Files {
		if _, ok := required[entry.Path]; !ok {
			t.Errorf("unexpected legacy source manifest path: %s", entry.Path)
			continue
		}
		required[entry.Path] = true

		data, err := os.ReadFile(filepath.Join(root, entry.Path))
		if err != nil {
			t.Errorf("read %s: %v", entry.Path, err)
			continue
		}
		data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
		data = bytes.ReplaceAll(data, []byte("\r"), []byte("\n"))
		sum := sha256.Sum256(data)
		if got := hex.EncodeToString(sum[:]); got != entry.SHA256 {
			t.Errorf("%s digest mismatch: got %s, want %s", entry.Path, got, entry.SHA256)
		}
	}
	for path, found := range required {
		if !found {
			t.Errorf("required legacy source manifest path missing: %s", path)
		}
	}
}

func findProjectRoot(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "decompiled", "data")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("project root not found")
		}
		dir = parent
	}
}
