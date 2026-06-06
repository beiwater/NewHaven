package catalog

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// ManifestEntry describes a single file tracked in the asset manifest.
type ManifestEntry struct {
	Path    string `json:"path"`
	Version string `json:"version"`
	SHA256  string `json:"sha256"`
}

// Manifest is the top-level version manifest for the decompiled game data set.
type Manifest struct {
	Version string          `json:"version"`
	Files   []ManifestEntry `json:"files"`
}

// LoadManifest reads and parses the static-data-manifest.json inside internal/catalog.
func LoadManifest(projectRoot string) (*Manifest, error) {
	path := filepath.Join(projectRoot, "backend-next", "internal", "catalog", "static-data-manifest.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read version manifest: %w", err)
	}
	var m Manifest
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, fmt.Errorf("unmarshal version manifest: %w", err)
	}
	return &m, nil
}

// ValidateManifest loads the manifest and verifies the SHA-256 digest of every
// listed file. It returns an error listing all mismatches (or a single wrapping
// error when the manifest itself cannot be loaded).
func ValidateManifest(projectRoot string) error {
	m, err := LoadManifest(projectRoot)
	if err != nil {
		return err
	}

	var errs []string
	for _, f := range m.Files {
		path := filepath.Join(projectRoot, f.Path)
		raw, err := os.ReadFile(path)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", f.Path, err))
			continue
		}
		sum := sha256.Sum256(canonicalText(raw))
		got := hex.EncodeToString(sum[:])
		if got != f.SHA256 {
			errs = append(errs, fmt.Sprintf("%s: digest mismatch (expected %s, got %s)", f.Path, f.SHA256, got))
		}
	}

	if len(errs) > 0 {
		msg := "manifest integrity check failed:\n"
		for _, e := range errs {
			msg += "  - " + e + "\n"
		}
		return fmt.Errorf("%s", msg)
	}
	return nil
}

func canonicalText(raw []byte) []byte {
	raw = bytes.ReplaceAll(raw, []byte("\r\n"), []byte("\n"))
	return bytes.ReplaceAll(raw, []byte("\r"), []byte("\n"))
}
