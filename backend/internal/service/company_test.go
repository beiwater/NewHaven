package service

import (
	"testing"

	"go-sim-api/internal/model"
)

func TestCompaniesByPlayer(t *testing.T) {
	s := newCoreTestService()
	result := s.CompaniesByPlayer(0)
	if len(result) == 0 {
		t.Fatal("expected companies")
	}
}

func TestCOOandCTOSkill(t *testing.T) {
	s := newCoreTestService()
	s.mu.Lock()
	s.State.Executives = []map[string]any{
		{"id": "ex-coo-1", "role": "COO", "skill": 8},
		{"id": "ex-cto-1", "role": "CTO", "skill": 7},
	}
	s.mu.Unlock()

	coo, cto := s.COOandCTOSkill()
	if coo != 8 {
		t.Errorf("expected COO skill 8, got %.0f", coo)
	}
	if cto != 7 {
		t.Errorf("expected CTO skill 7, got %.0f", cto)
	}
}

func TestUpgradeBuilding(t *testing.T) {
	s := newCoreTestService()
	s.mu.Lock()
	s.State.Companies[0].PlacedBuildings = []map[string]any{
		{"id": "b-1", "kind": 2, "level": 1, "baseCost": 10000},
	}
	s.State.Companies[0].Money = 100000
	s.mu.Unlock()

	result, err := s.UpgradeBuilding(s.State.Companies[0].ID, "b-1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	newLv := result["newLevel"].(int)
	if newLv != 2 {
		t.Errorf("expected level 2, got %d", newLv)
	}
	if s.State.Companies[0].Money >= 100000 {
		t.Errorf("money should have been deducted")
	}
}

func TestUpgradeBuilding_NotEnoughMoney(t *testing.T) {
	s := newCoreTestService()
	s.mu.Lock()
	s.State.Companies[0].PlacedBuildings = []map[string]any{
		{"id": "b-1", "kind": 5, "level": 10, "baseCost": 10000},
	}
	s.State.Companies[0].Money = 1000
	s.mu.Unlock()

	_, err := s.UpgradeBuilding(s.State.Companies[0].ID, "b-1")
	if err == nil {
		t.Fatal("expected error for insufficient funds")
	}
}

func TestMarketOrderDepth(t *testing.T) {
	s := newCoreTestService()
	s.mu.Lock()
	s.State.Orders = []model.MarketOrder{
		{ID: "b1", ResourceID: 3, Kind: 1, Price: 10.0, Quality: 0, Quantity: 5, Remaining: 5, CompanyID: 1},
		{ID: "b2", ResourceID: 3, Kind: 1, Price: 9.0, Quality: 0, Quantity: 3, Remaining: 3, CompanyID: 1},
		{ID: "s1", ResourceID: 3, Kind: 0, Price: 11.0, Quality: 0, Quantity: 4, Remaining: 4, CompanyID: 2},
	}
	s.mu.Unlock()

	d := s.OrderBookDepth(3, 0)
	buys := d["buys"].([]depthLevel)
	sells := d["sells"].([]depthLevel)
	if len(buys) != 2 || len(sells) != 1 {
		t.Errorf("expected 2 buys, 1 sell; got %d buys, %d sells", len(buys), len(sells))
	}
}
