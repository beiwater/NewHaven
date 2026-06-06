package market_test

import (
	"context"
	"testing"
	"time"

	appmarket "github.com/newhaven/backend-next/internal/app/market"
	"github.com/newhaven/backend-next/internal/catalog"
	"github.com/newhaven/backend-next/internal/config"
	domainmarket "github.com/newhaven/backend-next/internal/domain/market"
	"github.com/newhaven/backend-next/internal/platform"
	"github.com/newhaven/backend-next/internal/storage/memory"
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
	svc := appmarket.NewService(store, store, store, resources, economy, cfg, clock, idgen)
	return svc, store
}

// setupRetailTicker creates a ticker entry so ProcessRetailSales uses it as sale price.
func setupRetailTicker(t *testing.T, store *memory.Store, resourceID int, price float64) {
	t.Helper()
	now := time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC)
	ticker := &domainmarket.Ticker{
		ResourceID: resourceID,
		LastPrice:  price,
		UpdatedAt:  now.Format(time.RFC3339),
	}
	if err := store.UpdateTicker(nil, ticker); err != nil {
		t.Fatalf("UpdateTicker: %v", err)
	}
}

func TestProcessRetailSales_SellsGoodsAndCreditsMoney(t *testing.T) {
	t.Parallel()

	economy := map[int]*catalog.EconomyModelEntry{
		1: {
			BuildingKindModifier:               0.8,
			BuildingLevelsNeededPerUnitPerHour: 0.01,
			ModeledProductionCostPerUnit:       8.0,
			ModeledStoreWages:                  200.0,
			ModeledUnitsSoldAnHour:             15.0,
		},
	}
	svc, store := newRetailTestSvc(economy)
	setupRetailTicker(t, store, 1, 24.0)

	companyID := newTestCompany(t, store, 1001, "retail_test", 1000.0)
	c, err := store.GetCompany(nil, companyID)
	if err != nil {
		t.Fatalf("GetCompany: %v", err)
	}
	c.Inventory[1] = 100
	if err := store.UpdateCompany(nil, c); err != nil {
		t.Fatalf("UpdateCompany: %v", err)
	}

	if err := svc.ProcessRetailSales(context.Background()); err != nil {
		t.Fatalf("ProcessRetailSales: %v", err)
	}

	updated, err := store.GetCompany(nil, companyID)
	if err != nil {
		t.Fatalf("GetCompany after: %v", err)
	}
	if updated.Inventory[1] >= 100 {
		t.Errorf("expected inventory to be deducted, got %d", updated.Inventory[1])
	}
	if updated.Money <= 1000.0 {
		t.Errorf("expected money to increase above 1000, got %g", updated.Money)
	}
	t.Logf("Inventory: 100 → %d, Money: 1000 → %g", updated.Inventory[1], updated.Money)
}

func TestProcessRetailSales_NoEconomyModel_SkipsResource(t *testing.T) {
	t.Parallel()

	economy := map[int]*catalog.EconomyModelEntry{
		2: {
			BuildingKindModifier:               0.8,
			BuildingLevelsNeededPerUnitPerHour: 0.01,
			ModeledProductionCostPerUnit:       8.0,
			ModeledStoreWages:                  200.0,
			ModeledUnitsSoldAnHour:             15.0,
		},
	}
	svc, store := newRetailTestSvc(economy)
	setupRetailTicker(t, store, 1, 24.0)

	companyID := newTestCompany(t, store, 1002, "no_eco_test", 500.0)
	c, err := store.GetCompany(nil, companyID)
	if err != nil {
		t.Fatalf("GetCompany: %v", err)
	}
	c.Inventory[1] = 50
	if err := store.UpdateCompany(nil, c); err != nil {
		t.Fatalf("UpdateCompany: %v", err)
	}

	if err := svc.ProcessRetailSales(context.Background()); err != nil {
		t.Fatalf("ProcessRetailSales: %v", err)
	}

	updated, err := store.GetCompany(nil, companyID)
	if err != nil {
		t.Fatalf("GetCompany after: %v", err)
	}
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
	c, err := store.GetCompany(nil, companyID)
	if err != nil {
		t.Fatalf("GetCompany: %v", err)
	}
	c.Inventory[1] = 50
	if err := store.UpdateCompany(nil, c); err != nil {
		t.Fatalf("UpdateCompany: %v", err)
	}

	if err := svc.ProcessRetailSales(context.Background()); err != nil {
		t.Fatalf("ProcessRetailSales: %v", err)
	}

	updated, err := store.GetCompany(nil, companyID)
	if err != nil {
		t.Fatalf("GetCompany after: %v", err)
	}
	if updated.Money != 500.0 {
		t.Errorf("money should be unchanged with empty economy, got %g", updated.Money)
	}
}
