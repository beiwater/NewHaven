package service

import (
	"testing"

	"go-sim-api/internal/config"
	"go-sim-api/internal/data"
)

// TestExecutiveStateIsolation verifies that two Service instances
// have independent executive hired state (not shared via package globals).
func TestExecutiveStateIsolation(t *testing.T) {
	s1 := New(&data.StaticData{Resources: []map[string]any{}, EconomyModel: map[string]any{"models": map[string]any{}}}, config.DefaultTestConfig(), nil)
	s2 := New(&data.StaticData{Resources: []map[string]any{}, EconomyModel: map[string]any{"models": map[string]any{}}}, config.DefaultTestConfig(), nil)

	// Both start with empty hired lists
	execs1 := s1.MyExecutives()
	execs2 := s2.MyExecutives()
	if len(execs1) != 0 {
		t.Fatalf("s1 started with %d executives, want 0", len(execs1))
	}
	if len(execs2) != 0 {
		t.Fatalf("s2 started with %d executives, want 0", len(execs2))
	}

	// Get catalog for s1 (this populates s1's market)
	catalog1 := s1.ExecutiveCatalog()
	if len(catalog1) == 0 {
		t.Fatal("s1 catalog should have executives")
	}
	recruitID := catalog1[0]["id"].(string)

	// Recruit into s1
	result := s1.RecruitExecutive(recruitID)
	if ok, _ := result["ok"].(bool); !ok {
		t.Fatalf("s1 recruit failed: %v", result["error"])
	}

	// Verify s1 has 1 hired, s2 has 0 hired
	execs1 = s1.MyExecutives()
	execs2 = s2.MyExecutives()
	if len(execs1) != 1 {
		t.Fatalf("s1 has %d hired executives after recruit, want 1", len(execs1))
	}
	if len(execs2) != 0 {
		t.Fatalf("s2 has %d hired executives after s1 recruit, want 0 (leak)", len(execs2))
	}

	// Recruit the same ID into s2 — should succeed (different instance)
	catalog2 := s2.ExecutiveCatalog()
	if len(catalog2) == 0 {
		t.Fatal("s2 catalog should have executives")
	}
	// Find matching exec in s2's catalog
	var s2recruitID string
	for _, e := range catalog2 {
		if e["id"] == recruitID {
			s2recruitID = recruitID
			break
		}
	}
	if s2recruitID == "" && len(catalog2) > 0 {
		s2recruitID = catalog2[0]["id"].(string)
	}

	result2 := s2.RecruitExecutive(s2recruitID)
	if ok, _ := result2["ok"].(bool); !ok {
		t.Fatalf("s2 recruit failed: %v", result2["error"])
	}

	execs1 = s1.MyExecutives()
	execs2 = s2.MyExecutives()
	if len(execs1) != 1 {
		t.Fatalf("s1 hired count changed after s2 recruit = %d, want 1", len(execs1))
	}
	if len(execs2) != 1 {
		t.Fatalf("s2 hired count = %d after recruit, want 1", len(execs2))
	}
}
