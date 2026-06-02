package storage

import (
	"context"
	"encoding/json"
	"testing"

	"go-sim-api/internal/model"
)

// ---------- NoopStorage ----------

func TestNoopStorageRoundTrip(t *testing.T) {
	n := &NoopStorage{}
	ctx := context.Background()

	// LoadState returns nil,nil
	state, err := n.LoadState(ctx)
	if err != nil {
		t.Fatalf("LoadState error: %v", err)
	}
	if state != nil {
		t.Fatalf("expected nil state, got %+v", state)
	}

	// All save methods are no-ops
	if err := n.SaveState(ctx, &model.GameState{}); err != nil {
		t.Fatalf("SaveState error: %v", err)
	}
	if err := n.SaveCompany(ctx, &model.Company{ID: 1}); err != nil {
		t.Fatalf("SaveCompany error: %v", err)
	}
	if err := n.SaveOrders(ctx, []model.MarketOrder{}); err != nil {
		t.Fatalf("SaveOrders error: %v", err)
	}
	if err := n.SaveTrades(ctx, []model.Trade{}); err != nil {
		t.Fatalf("SaveTrades error: %v", err)
	}

	// Close is a no-op
	n.Close()
}

// ---------- GameState serialization consistency ----------

// TestGameStateRoundTrip verifies that all GameState fields that the
// storage layer depends on serialize/deserialize correctly through JSON.
// This validates the JSON tags used by the JSONB-based storage strategy.
func TestGameStateRoundTrip(t *testing.T) {
	state := &model.GameState{
		Players: []model.Player{
			{ID: 1, Username: "alice", Token: "tok1", CompanyID: 100, RegisteredAt: "2026-01-01T00:00:00Z"},
		},
		NextPlayerID: 2,
		Companies: []model.Company{
			{
				ID: 100, Name: "TestCo", Money: 50000, Level: 3,
				Inventory:        map[int]int{1: 100, 2: 200},
				QualityInventory: map[string]int{},
				PlacedBuildings: []map[string]any{
					{"id": "b-1", "kind": 1, "level": 2, "busy": false, "baseCost": 10000.0},
				},
				UnplacedBuildings: []map[string]any{
					{"id": "b-2", "kind": 2, "level": 1, "busy": false, "baseCost": 50000.0},
				},
				WarehouseLevel: 0,
			},
		},
		Orders: []model.MarketOrder{
			{ID: "o-1", ResourceID: 8, Kind: 0, Price: 10.5, Quantity: 100, Remaining: 100, CompanyID: 100, Status: "open"},
		},
		Trades: []model.Trade{
			{ID: "t-1", ResourceID: 8, Quantity: 50, Price: 10.0, BuyOrderID: "o-1", SellOrderID: "o-2", CreatedAt: "2026-01-01T00:00:00Z"},
		},
		Messages: []model.Message{
			{ID: "m-1", Chatroom: "N", Body: "hello", At: "2026-01-01T00:00:00Z"},
		},
		Ledger: []model.LedgerEntry{
			{ID: "l-1", At: "2026-01-01T00:00:00Z", Kind: "test", Amount: 1000, Direction: "in"},
		},
		Notifications: []model.Notification{
			{ID: "n-1", Title: "test", Body: "test", Read: false},
		},
		GovernmentContracts: []model.GovContract{
			{ID: "gov-1", ResourceID: 8, Quantity: 100, MaxPrice: 12.0, Status: "open"},
		},
		Executives: []map[string]any{
			{"id": "ex-1", "role": "COO", "skill": 8},
		},
		Achievements: []map[string]any{
			{"id": "a-1", "code": "first_trade"},
		},
		ContractsIn:  []map[string]any{{"id": "ci-1"}},
		ContractsOut: []map[string]any{{"id": "co-1"}},
		Bonds: []model.Bond{
			{ID: "bond-1", Amount: 5000, Interest: 0.012},
		},
		ProductionJobs: []model.ProductionJob{
			{ID: "job-1", BuildingID: "b-1", ResourceID: 8, Amount: 10, Status: "running"},
		},
		Auctions: []model.Auction{
			{ID: "auction-1", Item: "Test", StartingBid: 1000, Status: "open"},
		},
		ResearchProjects: []model.ResearchProject{
			{ID: "rp-1", Name: "Test Research", Status: "available"},
		},
		MarketPressure:      map[int]float64{8: 0.5},
		UnlockedRecipes:     map[int]bool{100: true},
		ResearchedQuality:   map[int]int{3: 1},
		DailyTradeVolume:    map[int]float64{8: 5000},
		DailyTradeQty:       map[int]int{8: 100},
		DailyHighPrice:      map[int]float64{8: 12.0},
		DailyLowPrice:       map[int]float64{8: 8.0},
		YesterdayVolume:     map[int]float64{8: 4000},
		YesterdayHighPrice:  map[int]float64{8: 11.0},
		YesterdayClose:      map[int]float64{8: 10.0},
		LastTradePrice:      map[int]float64{8: 10.5},
		MarketLocked:        map[int]bool{8: false},
		NationalTeamActive:  map[int]bool{8: false},
		PlayerPreferences:   map[string]any{"sound": true},
		MarketTicks:         map[int][]map[string]any{8: {{"price": 10.0}}},
		CSRFToken:           "csrf-abc",
		LastActiveAt:        "2026-01-01T00:00:00Z",
		XP:                  100,
		XpToNextLevel:       500,
		BoostType:           "speed",
		BoostEndsAt:         "2026-01-02T00:00:00Z",
		BoostMultiplier:     2.0,
		LastBotCycleAt:      "2026-01-01T01:00:00Z",
	}

	data, err := json.Marshal(state)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored model.GameState
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	// Verify that all array fields survive round-trip
	cases := []struct {
		name string
		got  int
		want int
	}{
		{"companies", len(restored.Companies), len(state.Companies)},
		{"orders", len(restored.Orders), len(state.Orders)},
		{"trades", len(restored.Trades), len(state.Trades)},
		{"messages", len(restored.Messages), len(state.Messages)},
		{"ledger", len(restored.Ledger), len(state.Ledger)},
		{"bonds", len(restored.Bonds), len(state.Bonds)},
		{"production_jobs", len(restored.ProductionJobs), len(state.ProductionJobs)},
		{"auctions", len(restored.Auctions), len(state.Auctions)},
		{"research_projects", len(restored.ResearchProjects), len(state.ResearchProjects)},
		{"notifications", len(restored.Notifications), len(state.Notifications)},
		{"government_contracts", len(restored.GovernmentContracts), len(state.GovernmentContracts)},
		{"players", len(restored.Players), len(state.Players)},
	}
	for _, c := range cases {
		if c.got != c.want {
			t.Errorf("%s: got %d, want %d", c.name, c.got, c.want)
		}
	}

	// Verify scalar fields
	if restored.XP != state.XP {
		t.Errorf("xp: got %d, want %d", restored.XP, state.XP)
	}
	if restored.BoostMultiplier != state.BoostMultiplier {
		t.Errorf("boost_multiplier: got %f, want %f", restored.BoostMultiplier, state.BoostMultiplier)
	}
	if restored.CSRFToken != state.CSRFToken {
		t.Errorf("csrf_token: got %q, want %q", restored.CSRFToken, state.CSRFToken)
	}
	if restored.BoostType != state.BoostType {
		t.Errorf("boost_type: got %q, want %q", restored.BoostType, state.BoostType)
	}
	if restored.LastBotCycleAt != state.LastBotCycleAt {
		t.Errorf("last_bot_cycle_at: got %q, want %q", restored.LastBotCycleAt, state.LastBotCycleAt)
	}

	// Verify map fields survived
	if len(restored.MarketPressure) != len(state.MarketPressure) {
		t.Errorf("market_pressure: got %d, want %d", len(restored.MarketPressure), len(state.MarketPressure))
	}
	if len(restored.UnlockedRecipes) != len(state.UnlockedRecipes) {
		t.Errorf("unlocked_recipes: got %d, want %d", len(restored.UnlockedRecipes), len(state.UnlockedRecipes))
	}
	if len(restored.ResearchedQuality) != len(state.ResearchedQuality) {
		t.Errorf("researched_quality: got %d, want %d", len(restored.ResearchedQuality), len(state.ResearchedQuality))
	}
	if restored.LastTradePrice[8] != state.LastTradePrice[8] {
		t.Errorf("last_trade_price[8]: got %f, want %f", restored.LastTradePrice[8], state.LastTradePrice[8])
	}
}

// TestCompanyJSONTags verifies that Company fields serialize/deserialize correctly.
func TestCompanyJSONTags(t *testing.T) {
	c := model.Company{
		ID: 100, Name: "TestCo", Money: 50000.0, Level: 5,
		ProductionSlots:  3,
		Inventory:        map[int]int{1: 100, 2: 200},
		PlacedBuildings:  []map[string]any{{"id": "b-1", "kind": 1, "level": 2}},
		UnplacedBuildings: []map[string]any{{"id": "b-2", "kind": 2, "level": 1}},
		WarehouseLevel:   1,
	}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored model.Company
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if restored.ID != c.ID {
		t.Errorf("id: got %d, want %d", restored.ID, c.ID)
	}
	if restored.Money != c.Money {
		t.Errorf("money: got %f, want %f", restored.Money, c.Money)
	}
	if restored.Inventory[1] != c.Inventory[1] {
		t.Errorf("inventory[1]: got %d, want %d", restored.Inventory[1], c.Inventory[1])
	}
	if len(restored.PlacedBuildings) != len(c.PlacedBuildings) {
		t.Errorf("placed_buildings len: got %d, want %d", len(restored.PlacedBuildings), len(c.PlacedBuildings))
	}
}

// TestMarketOrderJSONTags verifies MarketOrder serialization.
func TestMarketOrderJSONTags(t *testing.T) {
	o := model.MarketOrder{
		ID: "o-1", ResourceID: 8, Kind: 0, Price: 10.5,
		Quantity: 100, Remaining: 75, CompanyID: 100, Status: "open",
	}
	data, err := json.Marshal(o)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored model.MarketOrder
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if restored.ID != o.ID {
		t.Errorf("id: got %q, want %q", restored.ID, o.ID)
	}
	if restored.Price != o.Price {
		t.Errorf("price: got %f, want %f", restored.Price, o.Price)
	}
	if restored.Remaining != o.Remaining {
		t.Errorf("remaining: got %d, want %d", restored.Remaining, o.Remaining)
	}
}

// TestLedgerEntryJSONTags verifies LedgerEntry serialization.
func TestLedgerEntryJSONTags(t *testing.T) {
	l := model.LedgerEntry{
		ID: "l-1", At: "2026-01-01T00:00:00Z",
		Kind: "test", Amount: 1000.0, Direction: "in",
		Meta: map[string]any{"buildingId": "b-1"},
	}
	data, err := json.Marshal(l)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var restored model.LedgerEntry
	if err := json.Unmarshal(data, &restored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if restored.ID != l.ID {
		t.Errorf("id: got %q, want %q", restored.ID, l.ID)
	}
	if restored.Amount != l.Amount {
		t.Errorf("amount: got %f, want %f", restored.Amount, l.Amount)
	}
	if restored.Kind != l.Kind {
		t.Errorf("kind: got %q, want %q", restored.Kind, l.Kind)
	}
}
