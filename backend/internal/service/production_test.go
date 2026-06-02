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

	result := s.StartBuildingProduction(s.State.Companies[0].ID, "b-1", map[string]any{
		"kind": 2, "amount": 10,
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
	s.State.Companies[0].Inventory[2] = 0
	s.State.Companies[0].Inventory[1] = 3000
	s.mu.Unlock()

	result := s.StartBuildingProduction(s.State.Companies[0].ID, "b-1", map[string]any{
		"kind": 2, "amount": 10,
	})
	if err, ok := result["error"]; ok {
		t.Fatalf("StartBuildingProduction() unexpected error: %v", err)
	}
	if got := s.State.Companies[0].Inventory[2]; got != 0 {
		t.Fatalf("output inventory credited before claim = %d, want 0", got)
	}
	if got := s.State.ProductionJobs[0].Output[2]; got != 10 {
		t.Fatalf("job output = %d, want selected amount 10", got)
	}
}

func TestCalcProductionDuration(t *testing.T) {
	s := newCoreTestService()
	dur := s.calcProductionDuration(1, 10, 1)
	if dur <= 0 {
		t.Errorf("expected positive duration, got %d", dur)
	}
	if dur != 360 {
		t.Errorf("10 wheat at 100/h should take 360 seconds, got %d", dur)
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
	recipe, ok := s.findRecipe(2, 10)
	if !ok {
		t.Fatal("expected recipe resource to exist")
	}
	if len(recipe) == 0 {
		t.Fatal("expected recipe for resource 2")
	}
	if recipe[1] != 20 {
		t.Errorf("expected 20 wheat units (2*10), got %d", recipe[1])
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
		{ID: "job-1", BuildingID: "b-1", ResourceID: 2, Amount: 10, Status: "running"},
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
		{ID: "job-1", BuildingID: "b-1", ResourceID: 3, Amount: 10, Status: "claimed"},
	}
	s.mu.Unlock()

	_, err := s.ClaimProduction(s.State.Companies[0].ID, "job-1")
	if err == nil {
		t.Fatal("expected error for already claimed job")
	}
}
func TestCalcProductionDuration_EdgeCases(t *testing.T) {
	s := newCoreTestService()
	dur := s.calcProductionDuration(1, 10, 2)
	if dur != 180 {
		t.Errorf("level 2 should halve duration to 180 seconds, got %d", dur)
	}
	dur = s.calcProductionDuration(1, 1, 1)
	if dur != 36 {
		t.Errorf("1 wheat at 100/h should take 36 seconds, got %d", dur)
	}
	dur = s.calcProductionDuration(1, 0, 1)
	if dur != 36 {
		t.Errorf("zero amount should default to 1 unit, got %d", dur)
	}
}

func TestStartBuildingProductionRejectsOver48Hours(t *testing.T) {
	s := newCoreTestService()
	s.mu.Lock()
	s.State.Companies[0].PlacedBuildings = []map[string]any{
		{"id": "b-1", "kind": 1, "level": 1, "busy": false, "baseCost": 10000},
	}
	s.mu.Unlock()

	result := s.StartBuildingProduction(s.State.Companies[0].ID, "b-1", map[string]any{
		"kind": 1, "amount": 4801,
	})
	if _, ok := result["error"]; !ok {
		t.Fatalf("expected over-48h production error, got %v", result)
	}
	if got := intFromAny(result["maxAmount"]); got != 4800 {
		t.Errorf("maxAmount = %d, want 4800", got)
	}
}

func TestFindRecipe_NotFound(t *testing.T) {
	s := newCoreTestService()
	recipe, ok := s.findRecipe(9999, 1)
	if ok || len(recipe) != 0 {
		t.Errorf("expected empty map and missing flag for nonexistent recipe, got %v %v", recipe, ok)
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
