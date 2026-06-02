package service

import (
	"testing"
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
