package market_test

import (
	"context"
	"testing"
	"time"

	appmarket "github.com/beiwater/NewHaven/backend/internal/app/market"
	"github.com/beiwater/NewHaven/backend/internal/catalog"
	"github.com/beiwater/NewHaven/backend/internal/config"
	"github.com/beiwater/NewHaven/backend/internal/domain/company"
	domainmarket "github.com/beiwater/NewHaven/backend/internal/domain/market"
	"github.com/beiwater/NewHaven/backend/internal/platform"
	"github.com/beiwater/NewHaven/backend/internal/storage/memory"
)

func newRetailTestSvc(economy map[int]*catalog.EconomyModelEntry) (*appmarket.Service, *memory.Store) {
	store := memory.New()
	if economy == nil {
		economy = make(map[int]*catalog.EconomyModelEntry)
	}
	clock := platform.NewFakeClock(time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC))
	idgen := platform.NewIDGen()
	cfg := &config.GameConfig{ExchangeFeePct: 0.04}
	resources := make(map[int]*catalog.ResourceEntry)
	resources[1] = &catalog.ResourceEntry{ID: 1, BasePrice: 10.0}
	buildings := map[int]*catalog.BuildingEntry{
		6: {ID: 6, Kind: 6, Name: "Market Stall", Type: "retail", Produces: []int{1}, RetailSlots: 2, SlotPerLevel: 1},
	}
	svc := appmarket.NewService(store, store, store, resources, buildings, economy, cfg, clock, idgen)
	return svc, store
}

func setupRetailTicker(t *testing.T, store *memory.Store, resourceID int, price float64) {
	t.Helper()
	now := time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC)
	ticker := &domainmarket.Ticker{ResourceID: resourceID, LastPrice: price, UpdatedAt: now.Format(time.RFC3339)}
	if err := store.UpdateTicker(nil, ticker); err != nil {
		t.Fatalf("UpdateTicker: %v", err)
	}
}

func TestProcessRetailSales_SellsGoodsAndCreditsMoney(t *testing.T) {
	t.Parallel()
	economy := map[int]*catalog.EconomyModelEntry{
		1: {BuildingKindModifier: 0.8, BuildingLevelsNeededPerUnitPerHour: 0.01, ModeledProductionCostPerUnit: 8.0, ModeledStoreWages: 200.0, ModeledUnitsSoldAnHour: 15.0},
	}
	svc, store := newRetailTestSvc(economy)
	setupRetailTicker(t, store, 1, 24.0)

	companyID := newTestCompany(t, store, 1001, "retail_test", 1000.0)
	c, _ := store.GetCompany(nil, companyID)
	c.Inventory[1] = 100
	store.UpdateCompany(nil, c)

	if err := svc.ProcessRetailSales(context.Background()); err != nil {
		t.Fatalf("ProcessRetailSales: %v", err)
	}
	updated, _ := store.GetCompany(nil, companyID)
	if updated.Inventory[1] >= 100 {
		t.Errorf("expected NPC inventory to be deducted, got %d", updated.Inventory[1])
	}
	if updated.Money <= 1000.0 {
		t.Errorf("expected NPC money to increase, got %g", updated.Money)
	}
}

func TestProcessRetailSales_NoEconomyModel_SkipsResource(t *testing.T) {
	t.Parallel()
	economy := map[int]*catalog.EconomyModelEntry{
		2: {BuildingKindModifier: 0.8, BuildingLevelsNeededPerUnitPerHour: 0.01, ModeledProductionCostPerUnit: 8.0, ModeledStoreWages: 200.0, ModeledUnitsSoldAnHour: 15.0},
	}
	svc, store := newRetailTestSvc(economy)
	setupRetailTicker(t, store, 1, 24.0)

	companyID := newTestCompany(t, store, 1002, "no_eco_test", 500.0)
	c, _ := store.GetCompany(nil, companyID)
	c.Inventory[1] = 50
	store.UpdateCompany(nil, c)

	svc.ProcessRetailSales(context.Background())
	updated, _ := store.GetCompany(nil, companyID)
	if updated.Inventory[1] != 50 {
		t.Errorf("inventory should be unchanged (no economy model), got %d", updated.Inventory[1])
	}
	if updated.Money != 500.0 {
		t.Errorf("money should be unchanged, got %g", updated.Money)
	}
}

func TestProcessRetailSales_EmptyEconomy_NoOp(t *testing.T) {
	t.Parallel()
	svc, store := newRetailTestSvc(nil)
	companyID := newTestCompany(t, store, 1003, "empty_eco_test", 500.0)
	c, _ := store.GetCompany(nil, companyID)
	c.Inventory[1] = 50
	store.UpdateCompany(nil, c)

	svc.ProcessRetailSales(context.Background())
	updated, _ := store.GetCompany(nil, companyID)
	if updated.Money != 500.0 {
		t.Errorf("money should be unchanged with empty economy, got %g", updated.Money)
	}
}

func TestCatchUpPlayerRetail_FirstTime_SetsBaseline(t *testing.T) {
	t.Parallel()
	economy := map[int]*catalog.EconomyModelEntry{
		1: {BuildingKindModifier: 0.8, BuildingLevelsNeededPerUnitPerHour: 0.01, ModeledProductionCostPerUnit: 8.0, ModeledStoreWages: 200.0, ModeledUnitsSoldAnHour: 15.0},
	}
	svc, store := newRetailTestSvc(economy)
	setupRetailTicker(t, store, 1, 24.0)

	companyID := newTestCompany(t, store, 1004, "catchup_first", 500.0)
	c, _ := store.GetCompany(nil, companyID)
	c.Inventory[1] = 100
	store.UpdateCompany(nil, c)

	if err := svc.CatchUpPlayerRetail(context.Background(), companyID); err != nil {
		t.Fatalf("CatchUpPlayerRetail (first): %v", err)
	}
	updated, _ := store.GetCompany(nil, companyID)
	if updated.LastRetailAt == "" {
		t.Fatal("expected LastRetailAt to be set after first catch-up")
	}
	if updated.Inventory[1] != 100 {
		t.Errorf("inventory should be unchanged on first catch-up, got %d", updated.Inventory[1])
	}
	if updated.Money != 500.0 {
		t.Errorf("money should be unchanged on first catch-up, got %g", updated.Money)
	}
}

func TestCatchUpPlayerRetail_SettlesElapsedSales(t *testing.T) {
	t.Parallel()
	economy := map[int]*catalog.EconomyModelEntry{
		1: {BuildingKindModifier: 0.8, BuildingLevelsNeededPerUnitPerHour: 0.01, ModeledProductionCostPerUnit: 8.0, ModeledStoreWages: 200.0, ModeledUnitsSoldAnHour: 15.0},
	}
	svc, store := newRetailTestSvc(economy)
	setupRetailTicker(t, store, 1, 24.0)

	companyID := newTestCompany(t, store, 1005, "catchup_elapsed", 500.0)
	c, _ := store.GetCompany(nil, companyID)
	c.Buildings = []company.Building{
		{
			ID: "bld-retail-1", Kind: 6, Name: "Market Stall", Level: 1,
			Shelves: []company.ShelfItem{
				{ResourceID: 1, Quantity: 100, MaxQty: 100, Price: 24.0, PriceLock: true},
			},
		},
	}
	now := time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC)
	c.LastRetailAt = now.Add(-1 * time.Hour).Format(time.RFC3339)
	store.UpdateCompany(nil, c)

	if err := svc.CatchUpPlayerRetail(context.Background(), companyID); err != nil {
		t.Fatalf("CatchUpPlayerRetail: %v", err)
	}
	updated, _ := store.GetCompany(nil, companyID)

	// Shelf should be gone (all sold) or have fewer items
	soldOut := len(updated.Buildings) == 0 || len(updated.Buildings[0].Shelves) == 0
	var shelfQty int
	if !soldOut {
		shelfQty = updated.Buildings[0].Shelves[0].Quantity
	}
	if !soldOut && shelfQty >= 100 {
		t.Errorf("expected shelf quantity to be deducted, got %d", shelfQty)
	}
	if updated.Money <= 500.0 {
		t.Errorf("expected money to increase, got %g", updated.Money)
	}
}

func TestCatchUpPlayerRetail_SkipsPlayerCompanyInScheduler(t *testing.T) {
	t.Parallel()
	economy := map[int]*catalog.EconomyModelEntry{
		1: {BuildingKindModifier: 0.8, BuildingLevelsNeededPerUnitPerHour: 0.01, ModeledProductionCostPerUnit: 8.0, ModeledStoreWages: 200.0, ModeledUnitsSoldAnHour: 15.0},
	}
	svc, store := newRetailTestSvc(economy)
	setupRetailTicker(t, store, 1, 24.0)

	companyID := newTestCompany(t, store, 1006, "player_skip", 500.0)
	c, _ := store.GetCompany(nil, companyID)
	c.Inventory[1] = 100
	c.LastRetailAt = time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC).Format(time.RFC3339)
	store.UpdateCompany(nil, c)

	svc.ProcessRetailSales(context.Background())
	updated, _ := store.GetCompany(nil, companyID)
	if updated.Inventory[1] != 100 {
		t.Errorf("player inventory should not be touched by scheduler, got %d", updated.Inventory[1])
	}
	if updated.Money != 500.0 {
		t.Errorf("player money should not be touched by scheduler, got %g", updated.Money)
	}
}
