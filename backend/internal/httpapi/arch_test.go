// Package httpapi provides architecture enforcement tests.
package httpapi

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestNoAppImportsHttpapi verifies that app sub-packages do not import
// internal/httpapi. Only the top-level app.go wiring layer may import it.
func TestNoAppImportsHttpapi(t *testing.T) {
	root := findBackendNextRoot(t)
	appDir := filepath.Join(root, "internal", "app")
	err := filepath.Walk(appDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip the top-level wiring file — it necessarily imports httpapi.
		if strings.HasSuffix(path, "app.go") && !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		fset := token.NewFileSet()
		f, parseErr := parser.ParseFile(fset, path, nil, parser.ImportsOnly)
		if parseErr != nil {
			return nil
		}
		for _, imp := range f.Imports {
			importPath := strings.Trim(imp.Path.Value, "\"")
			if importPath == "github.com/beiwater/NewHaven/backend/internal/httpapi" {
				t.Errorf("%s imports httpapi (forbidden: app sub-packages must not depend on HTTP)", path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", appDir, err)
	}
}

// TestNoStringsContainsInHandlers verifies that production handler files
// do not contain strings.Contains(err.Error() calls. They should use writeAppErr instead.
func TestNoStringsContainsInHandlers(t *testing.T) {
	root := findBackendNextRoot(t)
	httpapiDir := filepath.Join(root, "internal", "httpapi")
	err := filepath.Walk(httpapiDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		if strings.Contains(string(data), `strings.Contains(err.Error()`) {
			t.Errorf("%s still uses strings.Contains(err.Error()) pattern; migrate to writeAppErr", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", httpapiDir, err)
	}
}

// findBackendNextRoot walks up from the test file location to find backend root.
// It looks for go.mod with "backend" module or the go.mod file path.
func findBackendNextRoot(t *testing.T) string {
	t.Helper()
	// Try common working directories.
	candidates := []string{
		".",                       // test runner cwd
		"..",                      // backend root
		filepath.Join("..", ".."), // one level up
	}
	for _, c := range candidates {
		gm := filepath.Join(c, "go.mod")
		if _, err := os.Stat(gm); err == nil {
			return c
		}
	}
	// Fallback: use the test file's own location
	wd, _ := os.Getwd()
	return wd
}
