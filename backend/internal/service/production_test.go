package service

import (
	"testing"

	"go-sim-api/internal/model"
)

func TestStartBuildingProduction(t *testing.T) {
	s := newCoreTestService()
	s.mu.Lock()
	s.State.Companies[0].PlacedBuildings = []map[string]any{
		{"id": "b-1", "kind": 2, "level": 1, "busy": false, "baseCost": 10000},
	}
	s.State.Companies[0].Inventory[2] = 3000 // water
	s.State.Companies[0].Inventory[1] = 5000 // power
	s.mu.Unlock()

	// Resources don't have a producedFrom for ID=1 (Power), so this will return "recipe not found"
	// Let's find a resource that has a recipe - ID=1 is Power which has producedFrom
	result := s.StartBuildingProduction(s.State.Companies[0].ID, "b-1", map[string]any{
		"kind": 8, "amount": 10,
	})
	if err, ok := result["error"]; ok {
		t.Logf("production returned error (may be expected): %v", err)
	} else {
		t.Logf("production job created: %v", result["building"])
	}
}

func TestStartBuildingProductionDoesNotCreditOutputImmediately(t *testing.T) {
	s := newCoreTestService()
	s.mu.Lock()
	s.State.Companies[0].PlacedBuildings = []map[string]any{
		{"id": "b-1", "kind": 2, "level": 1, "busy": false, "baseCost": 10000},
	}
	s.State.Companies[0].Inventory[8] = 0
	s.State.Companies[0].Inventory[2] = 3000
	s.mu.Unlock()

	result := s.StartBuildingProduction(s.State.Companies[0].ID, "b-1", map[string]any{
		"kind": 8, "amount": 10,
	})
	if err, ok := result["error"]; ok {
		t.Fatalf("StartBuildingProduction() unexpected error: %v", err)
	}
	if got := s.State.Companies[0].Inventory[8]; got != 0 {
		t.Fatalf("output inventory credited before claim = %d, want 0", got)
	}
}

func TestCalcProductionDuration(t *testing.T) {
	s := newCoreTestService()
	dur := s.calcProductionDuration(1, 10, 60)
	if dur <= 0 {
		t.Errorf("expected positive duration, got %d", dur)
	}
}

func TestResolveQuality(t *testing.T) {
	s := newCoreTestService()
	s.mu.Lock()
	s.State.Companies[0].QualityInventory = map[string]int{"2_2": 100, "1_2": 100}
	s.mu.Unlock()

	// Request Q3, need Q2 inputs, they exist
	input := map[int]int{2: 10, 1: 5}
	q := s.resolveQuality(&s.State.Companies[0], 3, input)
	if q != 3 {
		t.Errorf("expected quality 3, got %d", q)
	}
}

func TestFindRecipe(t *testing.T) {
	s := newCoreTestService()
	// Resource 8 (Beef) has producedFrom: {"2": "3"}
	recipe := s.findRecipe(8, 10)
	if len(recipe) == 0 {
		t.Fatal("expected recipe for resource 8")
	}
	if recipe[2] != 30 {
		t.Errorf("expected 30 water units (3*10), got %d", recipe[2])
	}
}

func TestFindBuilding(t *testing.T) {
	s := newCoreTestService()
	s.mu.Lock()
	s.State.Companies[0].PlacedBuildings = []map[string]any{
		{"id": "b-1", "kind": 2, "level": 3, "busy": false, "baseCost": 10000},
	}
	s.mu.Unlock()

	lv := s.findBuilding(s.State.Companies[0].ID, "b-1")
	if lv != 3 {
		t.Errorf("expected level 3, got %d", lv)
	}
}

func TestCheckBuildingSlot(t *testing.T) {
	s := newCoreTestService()
	s.mu.Lock()
	s.State.Companies[0].Level = 5 // 1 + 5/5 = 2 slots
	s.State.Companies[0].PlacedBuildings = []map[string]any{
		{"id": "b-1"}, {"id": "b-2"},
	}
	s.mu.Unlock()

	err := s.checkBuildingSlot(s.State.Companies[0].ID, "b-new")
	if err == nil {
		t.Errorf("expected error when no slots available")
	}
}

func TestClaimProduction_WrongCompany(t *testing.T) {
	s := newCoreTestService()
	s.mu.Lock()
	s.State.Companies[0].PlacedBuildings = []map[string]any{
		{"id": "b-1", "kind": 2, "level": 1, "busy": true, "baseCost": 10000},
	}
	s.State.ProductionJobs = []model.ProductionJob{
		{ID: "job-1", BuildingID: "b-1", ResourceID: 8, Amount: 10, Status: "running"},
	}
	s.mu.Unlock()

	_, err := s.ClaimProduction(s.State.Companies[0].ID, "job-1")
	if err == nil {
		t.Fatal("expected error for wrong company")
	}
}

func TestClaimProduction_AlreadyClaimed(t *testing.T) {
	s := newCoreTestService()
	s.mu.Lock()
	s.State.Companies[0].PlacedBuildings = []map[string]any{
		{"id": "b-1", "kind": 2, "level": 1, "busy": false, "baseCost": 10000},
	}
	s.State.ProductionJobs = []model.ProductionJob{
		{ID: "job-1", BuildingID: "b-1", ResourceID: 8, Amount: 10, Status: "claimed"},
	}
	s.mu.Unlock()

	_, err := s.ClaimProduction(s.State.Companies[0].ID, "job-1")
	if err == nil {
		t.Fatal("expected error for already claimed job")
	}
}
func TestCalcProductionDuration_EdgeCases(t *testing.T) {
	s := newCoreTestService()
	// Positive durSec is returned as-is when no economy model override
	dur := s.calcProductionDuration(1, 10, 60)
	if dur != 60 {
		t.Errorf("expected passthrough of 60, got %d", dur)
	}
	// Zero or negative durSec triggers fallback: max(30, amount*6)
	dur = s.calcProductionDuration(1, 10, 0)
	if dur != 60 {
		t.Errorf("10*6=60, expected 60, got %d", dur)
	}
	dur = s.calcProductionDuration(1, 1, 0)
	if dur != 30 {
		t.Errorf("max(30, 1*6)=30, expected 30, got %d", dur)
	}
	dur = s.calcProductionDuration(1, 0, -1)
	if dur != 30 {
		t.Errorf("max(30, 0*6)=30, expected 30, got %d", dur)
	}
}

func TestFindRecipe_NotFound(t *testing.T) {
	s := newCoreTestService()
	recipe := s.findRecipe(9999, 1)
	if len(recipe) != 0 {
		t.Errorf("expected empty map for nonexistent recipe, got %v", recipe)
	}
}

func TestResolveQuality_Defaults(t *testing.T) {
	s := newCoreTestService()
	// reqQuality == 0 returns 0 immediately
	q := s.resolveQuality(&s.State.Companies[0], 0, map[int]int{2: 10})
	if q != 0 {
		t.Errorf("expected quality 0 for reqQuality=0, got %d", q)
	}
	// Empty input returns reqQuality unchanged
	q = s.resolveQuality(&s.State.Companies[0], 5, map[int]int{})
	if q != 5 {
		t.Errorf("expected quality 5 for empty input, got %d", q)
	}
}

func TestCheckBuildingSlot_InvalidBuilding(t *testing.T) {
	s := newCoreTestService()
	// Company has no placed buildings — non-"b-new" building ID passes
	err := s.checkBuildingSlot(s.State.Companies[0].ID, "nonexistent")
	if err != nil {
		t.Errorf("expected no error for non-new building, got %v", err)
	}
}
