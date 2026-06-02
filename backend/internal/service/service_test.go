package service

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"testing"

	"go-sim-api/internal/config"
	"go-sim-api/internal/data"
	"go-sim-api/internal/model"
)

type fakeStorage struct {
	state *model.GameState
}

func (f *fakeStorage) Close() {}
func (f *fakeStorage) LoadState(_ context.Context) (*model.GameState, error) {
	return f.state, nil
}
func (f *fakeStorage) SaveState(_ context.Context, state *model.GameState) error {
	f.state = state
	return nil
}
func (f *fakeStorage) SaveCompany(_ context.Context, c *model.Company) error { return nil }
func (f *fakeStorage) SaveOrders(_ context.Context, orders []model.MarketOrder) error {
	return nil
}
func (f *fakeStorage) SaveTrades(_ context.Context, trades []model.Trade) error { return nil }

func newCoreTestService() *Service {
	cfg := config.DefaultTestConfig()
	d := &data.StaticData{
		Resources: []map[string]any{
			{"id": 1, "name": "Wheat", "dbLetter": 1, "producedPerHourRaw": 100.0},
			{"id": 2, "name": "Flour", "dbLetter": 2, "producedPerHourRaw": 80.0, "producedFrom": map[string]any{"1": 2}},
			{"id": 3, "name": "Bread", "dbLetter": 3, "producedPerHourRaw": 40.0, "producedFrom": map[string]any{"2": 2}},
			{"id": 4, "name": "Meals", "dbLetter": 4, "producedPerHourRaw": 30.0, "producedFrom": map[string]any{"3": 2}},
		},
		EconomyModel: map[string]any{"models": map[string]any{}},
	}
	return New(d, cfg, nil)
}

func TestNewService(t *testing.T) {
	s := newCoreTestService()
	if s.State.Companies[0].Money != 200000 {
		t.Errorf("expected 200000, got %.0f", s.State.Companies[0].Money)
	}
	if s.State.Companies[0].Level != 42 {
		t.Errorf("expected level 42, got %d", s.State.Companies[0].Level)
	}
}

func TestAddXP(t *testing.T) {
	s := newCoreTestService()
	s.mu.Lock()
	s.State.XpToNextLevel = 100
	s.State.XP = 0
	s.State.Companies[0].Level = 1
	s.addXP(&s.State.Companies[0], 150)
	s.mu.Unlock()
	if s.State.Companies[0].Level != 2 {
		t.Errorf("expected level 2, got %d", s.State.Companies[0].Level)
	}
	if s.State.XP != 50 {
		t.Errorf("expected 50 XP remaining, got %d", s.State.XP)
	}
}

func TestInventoryQuality(t *testing.T) {
	s := newCoreTestService()
	s.mu.Lock()
	s.State.Companies[0].Inventory[3] = 100
	s.State.Companies[0].QualityInventory = map[string]int{}

	if v := s.inventoryGet(&s.State.Companies[0], 3, 0); v != 100 {
		t.Errorf("expected inventory 100, got %d", v)
	}
	if v := s.inventoryGet(&s.State.Companies[0], 3, 1); v != 0 {
		t.Errorf("expected quality inventory 0, got %d", v)
	}
	s.inventoryAdd(&s.State.Companies[0], 3, 1, 50)
	if v := s.inventoryGet(&s.State.Companies[0], 3, 1); v != 50 {
		t.Errorf("expected quality inventory 50, got %d", v)
	}
	if !s.inventorySub(&s.State.Companies[0], 3, 1, 30) {
		t.Fatal("inventorySub returned false")
	}
	if v := s.inventoryGet(&s.State.Companies[0], 3, 1); v != 20 {
		t.Errorf("expected quality inventory 20, got %d", v)
	}
	s.mu.Unlock()
}

func TestOrderBookDepth(t *testing.T) {
	s := newCoreTestService()
	s.State.Orders = []model.MarketOrder{
		{ID: "b1", ResourceID: 3, Kind: 1, Price: 10.0, Quality: 0, Quantity: 10, Remaining: 10, CompanyID: 1},
		{ID: "b2", ResourceID: 3, Kind: 1, Price: 9.5, Quality: 0, Quantity: 5, Remaining: 5, CompanyID: 1},
		{ID: "s1", ResourceID: 3, Kind: 0, Price: 10.5, Quality: 0, Quantity: 8, Remaining: 8, CompanyID: 2},
		{ID: "s2", ResourceID: 3, Kind: 0, Price: 11.0, Quality: 0, Quantity: 3, Remaining: 3, CompanyID: 2},
	}
	depth := s.OrderBookDepth(3, 0)
	buys := depth["buys"].([]depthLevel)
	sells := depth["sells"].([]depthLevel)
	if len(buys) != 2 {
		t.Errorf("expected 2 buy levels, got %d", len(buys))
	}
	if len(sells) != 2 {
		t.Errorf("expected 2 sell levels, got %d", len(sells))
	}
}

func TestFinancialStatements(t *testing.T) {
	s := newCoreTestService()
	s.mu.Lock()
	s.State.Companies[0].Money = 100000
	s.State.Companies[0].Inventory[3] = 50
	s.mu.Unlock()

	fs := s.FinancialStatements(s.State.Companies[0].ID)
	if fs == nil {
		t.Fatal("expected financial statements")
	}
}

func TestCompanyProfile(t *testing.T) {
	s := newCoreTestService()
	prof := s.CompanyProfile(s.State.Companies[0].ID)
	if prof == nil {
		t.Fatal("expected profile")
	}
	auth, ok := prof["authCompany"].(map[string]any)
	if !ok {
		t.Fatal("expected authCompany")
	}
	if auth["money"] != 200000.0 {
		t.Errorf("expected 200000, got %v", auth["money"])
	}
}

func TestBuyBuildingRespectsUnlockLevel(t *testing.T) {
	s := newCoreTestService()
	companyID := s.State.Companies[0].ID
	s.State.Companies[0].Level = 1

	_, err := s.BuyBuilding(companyID, "b-shop-2")
	if err == nil {
		t.Fatal("expected level lock error")
	}
	if !strings.Contains(err.Error(), "level 2") {
		t.Fatalf("expected level 2 error, got %v", err)
	}
}

func TestBuyBuildingRespectsBuildingSlots(t *testing.T) {
	s := newCoreTestService()
	companyID := s.State.Companies[0].ID
	s.State.Companies[0].Level = 1
	s.State.Companies[0].PlacedBuildings = []map[string]any{
		{"id": "b-1", "kind": 1, "x": 1, "y": 1},
		{"id": "b-2", "kind": 1, "x": 2, "y": 1},
	}
	s.State.Companies[0].UnplacedBuildings = nil

	_, err := s.BuyBuilding(companyID, "b-shop-1")
	if err == nil {
		t.Fatal("expected building slot error")
	}
	if !strings.Contains(err.Error(), "building limit") {
		t.Fatalf("expected building limit error, got %v", err)
	}
}

func TestRegisterPlayerUsesPersistedNextPlayerID(t *testing.T) {
	s := newCoreTestService()
	s.State.NextPlayerID = 7
	result, err := s.RegisterPlayer("alice", "password")
	if err != nil {
		t.Fatalf("RegisterPlayer() unexpected error: %v", err)
	}
	player := result["player"].(model.Player)
	if player.ID != 7 {
		t.Fatalf("player ID = %d, want 7", player.ID)
	}
	if player.CompanyID != 1000007 {
		t.Fatalf("company ID = %d, want 1000007", player.CompanyID)
	}
	if s.State.NextPlayerID != 8 {
		t.Fatalf("next player ID = %d, want 8", s.State.NextPlayerID)
	}
	lastIdx := len(s.State.Companies) - 1
	if s.State.Companies[lastIdx].ID != player.CompanyID {
		t.Fatalf("new company ID = %d, want %d", s.State.Companies[lastIdx].ID, player.CompanyID)
	}
}

func TestRegisterPlayerRecoversNextPlayerIDWhenMissing(t *testing.T) {
	s := newCoreTestService()
	s.State.NextPlayerID = 0
	s.State.Players = []model.Player{{ID: 4, Username: "existing", CompanyID: 1000004}}

	result, err := s.RegisterPlayer("bob", "password")
	if err != nil {
		t.Fatalf("RegisterPlayer() unexpected error: %v", err)
	}
	player := result["player"].(model.Player)
	if player.ID != 5 {
		t.Fatalf("player ID = %d, want 5", player.ID)
	}
	if s.State.NextPlayerID != 6 {
		t.Fatalf("next player ID = %d, want 6", s.State.NextPlayerID)
	}
}

func TestSaveCompanyKeepsCompaniesInSync(t *testing.T) {
	s := newCoreTestService()
	s.State.Companies[0].Money = 12345

	s.mu.Lock()
	s.saveCompanyLocked(&s.State.Companies[0])
	s.mu.Unlock()

	for _, c := range s.State.Companies {
		if c.ID == s.State.Companies[0].ID {
			if c.Money != 12345 {
				t.Fatalf("companies snapshot money = %.0f, want 12345", c.Money)
			}
			return
		}
	}
	t.Fatalf("active company %d missing from companies", s.State.Companies[0].ID)
}

func TestNewInitializesRuntimeMapsForLoadedState(t *testing.T) {
	cfg := newCoreTestService().Cfg
	loaded := &model.GameState{
		Companies: []model.Company{{ID: 1, Name: "Loaded Co"}},
	}
	s := New(&data.StaticData{}, cfg, &fakeStorage{state: loaded})

	s.State.MarketPressure[3] += 0.1
	s.State.DailyTradeVolume[3] += 100
	s.State.MarketLocked[3] = true
	s.State.ProcessedRequests["req-1"] = map[string]any{"ok": true}

	if s.State.Companies[0].Inventory == nil {
		t.Fatal("company inventory should be initialized")
	}
	if len(s.State.Companies) != 1 {
		t.Fatalf("companies length = %d, want 1", len(s.State.Companies))
	}
}

func TestConcurrentStateAccessSmoke(t *testing.T) {
	s := newCoreTestService()
	s.State.Companies[0].Money = 1_000_000
	s.State.Companies[0].Inventory[3] = 10_000

	var wg sync.WaitGroup
	errs := make(chan error, 200)
	for i := 0; i < 25; i++ {
		wg.Add(4)
		go func(i int) {
			defer wg.Done()
			_ = s.Snapshot()
		}(i)
		go func(i int) {
			defer wg.Done()
			s.CleanupOrders()
		}(i)
		go func(i int) {
			defer wg.Done()
			if _, err := s.CreateOrder(s.State.Companies[0].ID, 3, 0, 0, 1, 10); err != nil {
				errs <- fmt.Errorf("create order %d: %w", i, err)
			}
		}(i)
		go func(i int) {
			defer wg.Done()
			// just a read to verify concurrent safety
			_ = s.State.Companies[0]
		}(i)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Error(err)
	}
}
