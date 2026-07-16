package market_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	appmarket "github.com/beiwater/NewHaven/backend/internal/app/market"
	"github.com/beiwater/NewHaven/backend/internal/catalog"
	"github.com/beiwater/NewHaven/backend/internal/config"
	"github.com/beiwater/NewHaven/backend/internal/domain/company"
	domainmarket "github.com/beiwater/NewHaven/backend/internal/domain/market"
	"github.com/beiwater/NewHaven/backend/internal/formula"
	"github.com/beiwater/NewHaven/backend/internal/platform"
	"github.com/beiwater/NewHaven/backend/internal/storage"
	"github.com/beiwater/NewHaven/backend/internal/storage/memory"
)

type overlapDetectingCompanyStore struct {
	storage.CompanyStorage
	active     atomic.Int32
	overlapped atomic.Bool
}

func (s *overlapDetectingCompanyStore) GetCompany(ctx context.Context, companyID int) (*company.Company, error) {
	if s.active.Add(1) > 1 {
		s.overlapped.Store(true)
	}
	defer s.active.Add(-1)
	// Make simultaneous profile requests overlap reliably unless the service
	// serializes the whole settlement transaction.
	time.Sleep(25 * time.Millisecond)
	return s.CompanyStorage.GetCompany(ctx, companyID)
}

func newRetailTestSvc(economy map[int]*catalog.EconomyModelEntry) (*appmarket.Service, *memory.Store, *platform.FakeClock) {
	store := memory.New()
	clock := platform.NewFakeClock(time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC))
	svc := newRetailTestSvcWithCompanies(economy, store, store, clock)
	return svc, store, clock
}

func newRetailTestSvcWithCompanies(economy map[int]*catalog.EconomyModelEntry, store *memory.Store, companies storage.CompanyStorage, clock *platform.FakeClock) *appmarket.Service {
	if economy == nil {
		economy = make(map[int]*catalog.EconomyModelEntry)
	}
	idgen := platform.NewIDGen()
	cfg := &config.GameConfig{ExchangeFeePct: 0.04}
	resources := make(map[int]*catalog.ResourceEntry)
	resources[1] = &catalog.ResourceEntry{ID: 1, BasePrice: 10.0, RetailDemandPerHour: 120, PrimaryBuildingKind: 6}
	buildings := map[int]*catalog.BuildingEntry{
		6: {ID: 6, Kind: 6, Name: "Market Stall", Type: "retail", Produces: []int{1}, RetailSlots: 2, SlotPerLevel: 1},
	}
	return appmarket.NewService(store, companies, store, resources, buildings, economy, cfg, clock, idgen)
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
	svc, store, _ := newRetailTestSvc(economy)
	setupRetailTicker(t, store, 1, 24.0)

	companyID := newTestCompany(t, store, -1001, "retail_test", 1000.0)
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

func TestProcessRetailSales_NoEconomyModel_UsesDemandParameter(t *testing.T) {
	t.Parallel()
	economy := map[int]*catalog.EconomyModelEntry{
		2: {BuildingKindModifier: 0.8, BuildingLevelsNeededPerUnitPerHour: 0.01, ModeledProductionCostPerUnit: 8.0, ModeledStoreWages: 200.0, ModeledUnitsSoldAnHour: 15.0},
	}
	svc, store, _ := newRetailTestSvc(economy)
	setupRetailTicker(t, store, 1, 24.0)

	companyID := newTestCompany(t, store, -1002, "no_eco_test", 500.0)
	c, _ := store.GetCompany(nil, companyID)
	c.Inventory[1] = 50
	store.UpdateCompany(nil, c)

	if err := svc.ProcessRetailSales(context.Background()); err != nil {
		t.Fatalf("ProcessRetailSales: %v", err)
	}
	updated, _ := store.GetCompany(nil, companyID)
	if updated.Inventory[1] >= 50 {
		t.Errorf("inventory should be deducted from the data-driven demand model, got %d", updated.Inventory[1])
	}
	if updated.Money <= 500.0 {
		t.Errorf("money should be credited, got %g", updated.Money)
	}
}

func TestProcessRetailSales_EmptyEconomy_StillSellsConfiguredResource(t *testing.T) {
	t.Parallel()
	svc, store, _ := newRetailTestSvc(nil)
	setupRetailTicker(t, store, 1, 24.0)
	companyID := newTestCompany(t, store, -1003, "empty_eco_test", 500.0)
	c, _ := store.GetCompany(nil, companyID)
	c.Inventory[1] = 50
	store.UpdateCompany(nil, c)

	if err := svc.ProcessRetailSales(context.Background()); err != nil {
		t.Fatalf("ProcessRetailSales: %v", err)
	}
	updated, _ := store.GetCompany(nil, companyID)
	if updated.Inventory[1] >= 50 || updated.Money <= 500.0 {
		t.Errorf("configured retail demand should not depend on a legacy economy model: %+v", updated)
	}
}

func TestCatchUpPlayerRetail_FirstTime_SetsBaseline(t *testing.T) {
	t.Parallel()
	economy := map[int]*catalog.EconomyModelEntry{
		1: {BuildingKindModifier: 0.8, BuildingLevelsNeededPerUnitPerHour: 0.01, ModeledProductionCostPerUnit: 8.0, ModeledStoreWages: 200.0, ModeledUnitsSoldAnHour: 15.0},
	}
	svc, store, _ := newRetailTestSvc(economy)
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

func TestCatchUpPlayerRetail_LocksLegacySaleBatchAtCurrentPrice(t *testing.T) {
	t.Parallel()
	svc, store, clock := newRetailTestSvc(map[int]*catalog.EconomyModelEntry{
		1: {BuildingKindModifier: 0.8, BuildingLevelsNeededPerUnitPerHour: 0.01, ModeledProductionCostPerUnit: 8.0, ModeledStoreWages: 200.0, ModeledUnitsSoldAnHour: 15.0},
	})
	setupRetailTicker(t, store, 1, 24.0)

	companyID := newTestCompany(t, store, 1014, "legacy_sale_batch", 500.0)
	c, _ := store.GetCompany(nil, companyID)
	c.Buildings = []company.Building{{
		ID: "bld-retail-legacy", Kind: 6, Name: "Market Stall", Level: 1,
		Shelves: []company.ShelfItem{{ResourceID: 1, Quantity: 10, MaxQty: 100}},
	}}
	c.LastRetailAt = clock.Now().UTC().Format(time.RFC3339)
	if err := store.UpdateCompany(nil, c); err != nil {
		t.Fatal(err)
	}

	if err := svc.CatchUpPlayerRetail(context.Background(), companyID); err != nil {
		t.Fatalf("CatchUpPlayerRetail: %v", err)
	}
	updated, _ := store.GetCompany(nil, companyID)
	shelf := updated.Buildings[0].Shelves[0]
	if !shelf.PriceLock || shelf.Price <= 24 {
		t.Fatalf("legacy batch was not locked at current price: %+v", shelf)
	}
}

func TestCatchUpPlayerRetail_SettlesElapsedSales(t *testing.T) {
	t.Parallel()
	economy := map[int]*catalog.EconomyModelEntry{
		1: {BuildingKindModifier: 0.8, BuildingLevelsNeededPerUnitPerHour: 0.01, ModeledProductionCostPerUnit: 8.0, ModeledStoreWages: 200.0, ModeledUnitsSoldAnHour: 15.0},
	}
	svc, store, _ := newRetailTestSvc(economy)
	setupRetailTicker(t, store, 1, 24.0)

	companyID := newTestCompany(t, store, 1005, "catchup_elapsed", 500.0)
	c, _ := store.GetCompany(nil, companyID)
	c.Buildings = []company.Building{
		{
			ID: "bld-retail-1", Kind: 6, Name: "Market Stall", Level: 1,
			Shelves: []company.ShelfItem{
				{ResourceID: 1, Quantity: 100, MaxQty: 100, Price: 100.0, PriceLock: true},
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
	if updated.Money == 500.0 {
		t.Errorf("expected retail settlement to change cash, got %g", updated.Money)
	}
}

func TestCatchUpPlayerRetail_AssignedCMOIncreasesDemandSpeed(t *testing.T) {
	t.Parallel()
	svc, store, clock := newRetailTestSvc(nil)
	setupRetailTicker(t, store, 1, 24.0)
	withoutCMO := newTestCompany(t, store, 1081, "without-cmo", 5000)
	withCMO := newTestCompany(t, store, 1082, "with-cmo", 5000)
	for _, companyID := range []int{withoutCMO, withCMO} {
		c, _ := store.GetCompany(nil, companyID)
		c.Buildings = []company.Building{{
			ID: "retail", Kind: 6, Name: "Market Stall", Level: 1,
			Shelves: []company.ShelfItem{{ResourceID: 1, Quantity: 1000, MaxQty: 1000, Price: 100, PriceLock: true}},
		}}
		c.LastRetailAt = clock.Now().Add(-time.Hour).Format(time.RFC3339)
		if companyID == withCMO {
			c.Executives = []company.Executive{{
				ID:       "cmo",
				Position: company.ExecutivePositionCMO,
				Skills:   company.ExecutiveSkills{Communication: 50},
			}}
		}
		if err := store.UpdateCompany(nil, c); err != nil {
			t.Fatal(err)
		}
	}
	if err := svc.CatchUpPlayerRetail(context.Background(), withoutCMO); err != nil {
		t.Fatal(err)
	}
	if err := svc.CatchUpPlayerRetail(context.Background(), withCMO); err != nil {
		t.Fatal(err)
	}
	without, _ := store.GetCompany(nil, withoutCMO)
	with, _ := store.GetCompany(nil, withCMO)
	withoutSold := 1000 - without.Buildings[0].Shelves[0].Quantity
	withSold := 1000 - with.Buildings[0].Shelves[0].Quantity
	if withSold <= withoutSold {
		t.Fatalf("assigned CMO sold %d, want more than no-CMO %d", withSold, withoutSold)
	}
}

func TestCatchUpPlayerRetail_OverpricedBatchStallsAndStillPaysPayroll(t *testing.T) {
	t.Parallel()
	economy := map[int]*catalog.EconomyModelEntry{
		1: {BuildingKindModifier: 0.8, BuildingLevelsNeededPerUnitPerHour: 0.01, ModeledProductionCostPerUnit: 8.0, ModeledStoreWages: 200.0, ModeledUnitsSoldAnHour: 15.0},
	}
	svc, store, clock := newRetailTestSvc(economy)
	setupRetailTicker(t, store, 1, 24.0)

	companyID := newTestCompany(t, store, 1015, "overpriced_payroll", 5000.0)
	c, _ := store.GetCompany(nil, companyID)
	c.Buildings = []company.Building{{
		ID: "bld-retail-1", Kind: 6, Name: "Market Stall", Level: 1,
		Shelves: []company.ShelfItem{{ResourceID: 1, Quantity: 100, MaxQty: 100, Price: 1_000_000, PriceLock: true}},
	}}
	c.LastRetailAt = clock.Now().Add(-time.Hour).Format(time.RFC3339)
	if err := store.UpdateCompany(nil, c); err != nil {
		t.Fatal(err)
	}

	if err := svc.CatchUpPlayerRetail(context.Background(), companyID); err != nil {
		t.Fatalf("CatchUpPlayerRetail: %v", err)
	}
	updated, _ := store.GetCompany(nil, companyID)
	if got := updated.Buildings[0].Shelves[0].Quantity; got != 100 {
		t.Fatalf("million-price batch sold %d units, want no practical sale", 100-got)
	}
	wantMoney := 5000.0 - formula.BuildingHourlyWage(6, 1)
	if updated.Money != wantMoney {
		t.Fatalf("money after stalled sale = %g, want %g after one active hour of payroll", updated.Money, wantMoney)
	}
	entries, err := store.GetLedgerEntries(context.Background(), companyID, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 || entries[0].Kind != "retail_wages" || entries[0].Direction != "out" {
		t.Fatalf("expected one retail_wages debit, got %+v", entries)
	}
}

func TestCatchUpPlayerRetail_StopsPayrollWhenBatchSellsOut(t *testing.T) {
	t.Parallel()
	economy := map[int]*catalog.EconomyModelEntry{
		1: {BuildingKindModifier: 0.8, BuildingLevelsNeededPerUnitPerHour: 0.01, ModeledProductionCostPerUnit: 8.0, ModeledStoreWages: 200.0, ModeledUnitsSoldAnHour: 15.0},
	}
	svc, store, clock := newRetailTestSvc(economy)
	setupRetailTicker(t, store, 1, 24.0)

	companyID := newTestCompany(t, store, 1016, "sellout_payroll", 5000.0)
	c, _ := store.GetCompany(nil, companyID)
	c.Buildings = []company.Building{{
		ID: "bld-retail-1", Kind: 6, Name: "Market Stall", Level: 1,
		Shelves: []company.ShelfItem{{ResourceID: 1, Quantity: 1, MaxQty: 100, Price: 100, PriceLock: true}},
	}}
	c.LastRetailAt = clock.Now().Add(-time.Hour).Format(time.RFC3339)
	if err := store.UpdateCompany(nil, c); err != nil {
		t.Fatal(err)
	}

	if err := svc.CatchUpPlayerRetail(context.Background(), companyID); err != nil {
		t.Fatalf("CatchUpPlayerRetail: %v", err)
	}
	entries, err := store.GetLedgerEntries(context.Background(), companyID, 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Kind == "retail_wages" {
			if entry.Amount <= 0 || entry.Amount >= formula.BuildingHourlyWage(6, 1) {
				t.Fatalf("sell-out payroll = %g, want a partial active-hour wage", entry.Amount)
			}
			return
		}
	}
	t.Fatalf("retail payroll entry not found: %+v", entries)
}

func TestProcessRetailSales_SkipsPlayerWithoutSettlementBaseline(t *testing.T) {
	t.Parallel()
	economy := map[int]*catalog.EconomyModelEntry{
		1: {BuildingKindModifier: 0.8, BuildingLevelsNeededPerUnitPerHour: 0.01, ModeledProductionCostPerUnit: 8.0, ModeledStoreWages: 200.0, ModeledUnitsSoldAnHour: 15.0},
	}
	svc, store, _ := newRetailTestSvc(economy)
	setupRetailTicker(t, store, 1, 24.0)

	companyID := newTestCompany(t, store, 1006, "player_skip", 500.0)
	c, _ := store.GetCompany(nil, companyID)
	c.Inventory[1] = 100
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

func TestCatchUpPlayerRetail_CarriesFractionalDemandAcrossProfilePolls(t *testing.T) {
	t.Parallel()
	economy := map[int]*catalog.EconomyModelEntry{
		1: {BuildingKindModifier: 0.8, BuildingLevelsNeededPerUnitPerHour: 0.01, ModeledProductionCostPerUnit: 8.0, ModeledStoreWages: 200.0, ModeledUnitsSoldAnHour: 15.0},
	}
	svc, store, clock := newRetailTestSvc(economy)
	setupRetailTicker(t, store, 1, 24.0)

	companyID := newTestCompany(t, store, 1007, "fractional_demand", 500.0)
	c, _ := store.GetCompany(nil, companyID)
	c.Buildings = []company.Building{{
		ID: "bld-retail-1", Kind: 6, Name: "Market Stall", Level: 1,
		Shelves: []company.ShelfItem{{ResourceID: 1, Quantity: 100, MaxQty: 100, Price: 100.0, PriceLock: true}},
	}}
	c.LastRetailAt = clock.Now().Add(-15 * time.Second).Format(time.RFC3339)
	store.UpdateCompany(nil, c)

	if err := svc.CatchUpPlayerRetail(context.Background(), companyID); err != nil {
		t.Fatalf("first short catch-up: %v", err)
	}
	updated, _ := store.GetCompany(nil, companyID)
	if got := updated.Buildings[0].Shelves[0].Quantity; got != 100 {
		t.Fatalf("first 15-second poll should not yet sell a whole item, got %d", got)
	}
	if len(updated.RetailCarry) != 1 {
		t.Fatalf("expected fractional demand to be retained, got %#v", updated.RetailCarry)
	}

	clock.Advance(5 * time.Minute)
	if err := svc.CatchUpPlayerRetail(context.Background(), companyID); err != nil {
		t.Fatalf("second short catch-up: %v", err)
	}
	updated, _ = store.GetCompany(nil, companyID)
	if got := updated.Buildings[0].Shelves[0].Quantity; got >= 100 {
		t.Errorf("repeated short polls should eventually sell stock, got %d", got)
	}
	if updated.Money == 500.0 {
		t.Errorf("expected retained demand to settle cash, got %g", updated.Money)
	}
}

func TestCatchUpPlayerRetail_DeductsTheShelfThatMadeTheSale(t *testing.T) {
	t.Parallel()
	economy := map[int]*catalog.EconomyModelEntry{
		1: {BuildingKindModifier: 0.8, BuildingLevelsNeededPerUnitPerHour: 0.01, ModeledProductionCostPerUnit: 8.0, ModeledStoreWages: 200.0, ModeledUnitsSoldAnHour: 15.0},
	}
	svc, store, clock := newRetailTestSvc(economy)
	setupRetailTicker(t, store, 1, 24.0)

	companyID := newTestCompany(t, store, 1008, "shelf_identity", 500.0)
	c, _ := store.GetCompany(nil, companyID)
	c.Buildings = []company.Building{
		{ID: "bld-no-sale", Kind: 6, Name: "No Sale", Level: 1, Shelves: []company.ShelfItem{{ResourceID: 1, Quantity: 100, MaxQty: 100, Price: 1_000_000, PriceLock: true}}},
		{ID: "bld-sale", Kind: 6, Name: "Sale", Level: 1, Shelves: []company.ShelfItem{{ResourceID: 1, Quantity: 100, MaxQty: 100, Price: 100.0, PriceLock: true}}},
	}
	c.LastRetailAt = clock.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	store.UpdateCompany(nil, c)

	if err := svc.CatchUpPlayerRetail(context.Background(), companyID); err != nil {
		t.Fatalf("CatchUpPlayerRetail: %v", err)
	}
	updated, _ := store.GetCompany(nil, companyID)
	if got := updated.Buildings[0].Shelves[0].Quantity; got != 100 {
		t.Errorf("non-selling shelf was deducted, got %d", got)
	}
	if len(updated.Buildings[1].Shelves) > 0 && updated.Buildings[1].Shelves[0].Quantity >= 100 {
		t.Errorf("selling shelf was not deducted, got %d", updated.Buildings[1].Shelves[0].Quantity)
	}
}

func TestCatchUpPlayerRetail_SerializesConcurrentSettlements(t *testing.T) {
	economy := map[int]*catalog.EconomyModelEntry{
		1: {BuildingKindModifier: 0.8, BuildingLevelsNeededPerUnitPerHour: 0.01, ModeledProductionCostPerUnit: 8.0, ModeledStoreWages: 200.0, ModeledUnitsSoldAnHour: 15.0},
	}
	store := memory.New()
	clock := platform.NewFakeClock(time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC))
	detector := &overlapDetectingCompanyStore{CompanyStorage: store}
	svc := newRetailTestSvcWithCompanies(economy, store, detector, clock)
	setupRetailTicker(t, store, 1, 24.0)

	companyID := newTestCompany(t, store, 1009, "concurrent_settlement", 500.0)
	c, _ := store.GetCompany(nil, companyID)
	c.Buildings = []company.Building{{
		ID: "bld-retail-1", Kind: 6, Name: "Market Stall", Level: 1,
		Shelves: []company.ShelfItem{{ResourceID: 1, Quantity: 100, MaxQty: 100, Price: 100.0, PriceLock: true}},
	}}
	c.LastRetailAt = clock.Now().Add(-1 * time.Hour).Format(time.RFC3339)
	store.UpdateCompany(nil, c)

	const requests = 8
	start := make(chan struct{})
	errs := make(chan error, requests)
	var wg sync.WaitGroup
	for range requests {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			errs <- svc.CatchUpPlayerRetail(context.Background(), companyID)
		}()
	}
	close(start)
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent catch-up failed: %v", err)
		}
	}
	if detector.overlapped.Load() {
		t.Fatal("retail settlements overlapped for the same service")
	}
	updated, _ := store.GetCompany(nil, companyID)
	if updated.Money <= 500.0 {
		t.Fatal("expected the elapsed retail interval to settle")
	}
}

func TestCatchUpPlayerRetail_AllowsDifferentCompaniesToSettleConcurrently(t *testing.T) {
	economy := map[int]*catalog.EconomyModelEntry{
		1: {BuildingKindModifier: 0.8, BuildingLevelsNeededPerUnitPerHour: 0.01, ModeledProductionCostPerUnit: 8.0, ModeledStoreWages: 200.0, ModeledUnitsSoldAnHour: 15.0},
	}
	store := memory.New()
	clock := platform.NewFakeClock(time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC))
	detector := &overlapDetectingCompanyStore{CompanyStorage: store}
	svc := newRetailTestSvcWithCompanies(economy, store, detector, clock)

	companyA := newTestCompany(t, store, 1010, "parallel_a", 500.0)
	companyB := newTestCompany(t, store, 1011, "parallel_b", 500.0)

	start := make(chan struct{})
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for _, companyID := range []int{companyA, companyB} {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			<-start
			errs <- svc.CatchUpPlayerRetail(context.Background(), id)
		}(companyID)
	}
	close(start)
	wg.Wait()
	close(errs)

	for err := range errs {
		if err != nil {
			t.Fatalf("parallel catch-up failed: %v", err)
		}
	}
	if !detector.overlapped.Load() {
		t.Fatal("different companies were serialized behind one global retail lock")
	}
}
