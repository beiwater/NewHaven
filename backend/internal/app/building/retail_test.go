package building_test

import (
	"context"
	"sync"
	"testing"

	"github.com/beiwater/NewHaven/backend/internal/app/building"
	"github.com/beiwater/NewHaven/backend/internal/apperr"
	"github.com/beiwater/NewHaven/backend/internal/catalog"
	"github.com/beiwater/NewHaven/backend/internal/domain/auth"
	domain "github.com/beiwater/NewHaven/backend/internal/domain/company"
	"github.com/beiwater/NewHaven/backend/internal/storage/memory"
)

func newRetailBuildingService(t *testing.T) (*building.Service, *memory.Store, int) {
	t.Helper()
	ctx := context.Background()
	store := memory.New()
	if err := store.CreatePlayer(ctx, &auth.Player{ID: 71, Username: "retail-batch"}); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateCompany(ctx, &domain.Company{
		PlayerID: 71,
		Name:     "Retail Batch Corp",
		Money:    1000,
		Inventory: map[int]int{
			11: 40,
		},
		Buildings: []domain.Building{{
			ID: "retail-1", BuildingID: 6, Kind: 6, Name: "Market Stall", Level: 1,
		}},
	}); err != nil {
		t.Fatal(err)
	}
	company, err := store.GetCompanyByPlayerID(ctx, 71)
	if err != nil {
		t.Fatal(err)
	}
	service := building.NewService(store, map[int]*catalog.BuildingEntry{
		6: {ID: 6, Kind: 6, Name: "Market Stall", Type: "retail", Produces: []int{11}, RetailSlots: 2, SlotPerLevel: 1},
	}, nil, nil, nil)
	return service, store, company.ID
}

func TestRetailSaleBatchLocksGoodsAndPriceUntilSoldOut(t *testing.T) {
	ctx := context.Background()
	service, store, companyID := newRetailBuildingService(t)

	if _, err := service.StockShelf(ctx, companyID, "retail-1", 11, 0, 10, 78); err != nil {
		t.Fatalf("StockShelf: %v", err)
	}
	company, _ := store.GetCompany(ctx, companyID)
	if company.Inventory[11] != 30 || len(company.Buildings[0].Shelves) != 1 {
		t.Fatalf("sale batch was not reserved correctly: %+v", company)
	}
	shelf := company.Buildings[0].Shelves[0]
	if shelf.Quantity != 10 || shelf.Price != 78 || !shelf.PriceLock {
		t.Fatalf("unexpected committed shelf: %+v", shelf)
	}

	if _, err := service.StockShelf(ctx, companyID, "retail-1", 11, 0, 5, 80); !apperr.HasKind(err, apperr.KindConflict) {
		t.Fatalf("restock active batch error=%v, want conflict", err)
	}
	if _, err := service.UnstockShelf(ctx, companyID, "retail-1", 11, 10); !apperr.HasKind(err, apperr.KindConflict) {
		t.Fatalf("cancel active batch error=%v, want conflict", err)
	}
	if _, err := service.SetShelfPrice(ctx, companyID, "retail-1", 11, 80, true); !apperr.HasKind(err, apperr.KindConflict) {
		t.Fatalf("reprice active batch error=%v, want conflict", err)
	}
	company, _ = store.GetCompany(ctx, companyID)
	if company.Inventory[11] != 30 || company.Buildings[0].Shelves[0].Quantity != 10 || company.Buildings[0].Shelves[0].Price != 78 {
		t.Fatalf("rejected mutation changed the active batch: %+v", company.Buildings[0].Shelves[0])
	}

	// Retail settlement removes an exhausted shelf. That is the boundary at
	// which the player may choose a new quantity and price.
	company.Buildings[0].Shelves = nil
	if err := store.UpdateCompany(ctx, company); err != nil {
		t.Fatal(err)
	}
	if _, err := service.StockShelf(ctx, companyID, "retail-1", 11, 0, 5, 80); err != nil {
		t.Fatalf("new batch after sell-out: %v", err)
	}
}

func TestRetailSaleRequiresPositiveQuantityAndPrice(t *testing.T) {
	ctx := context.Background()
	service, _, companyID := newRetailBuildingService(t)
	if _, err := service.StockShelf(ctx, companyID, "retail-1", 11, 0, 0, 78); !apperr.HasKind(err, apperr.KindBadRequest) {
		t.Fatalf("zero quantity error=%v, want bad request", err)
	}
	if _, err := service.StockShelf(ctx, companyID, "retail-1", 11, 0, 10, 0); !apperr.HasKind(err, apperr.KindBadRequest) {
		t.Fatalf("zero price error=%v, want bad request", err)
	}
}

func TestRetailSaleReservesOnlyTheSelectedQuality(t *testing.T) {
	ctx := context.Background()
	service, store, companyID := newRetailBuildingService(t)
	if err := store.UpdateInventoryQuality(ctx, companyID, 11, 4, 12); err != nil {
		t.Fatal(err)
	}
	resp, err := service.StockShelf(ctx, companyID, "retail-1", 11, 4, 10, 78)
	if err != nil {
		t.Fatalf("StockShelf Q4: %v", err)
	}
	if resp.Shelf == nil || resp.Shelf.Quality == nil || *resp.Shelf.Quality != 4 {
		t.Fatalf("shelf quality = %+v, want Q4", resp.Shelf)
	}
	warehouse, _ := store.GetWarehouse(ctx, companyID)
	qualityAmount := 0
	for _, item := range warehouse.Items {
		if item.ResourceID == 11 && item.Quality == 4 {
			qualityAmount = item.Amount
		}
	}
	if qualityAmount != 2 {
		t.Fatalf("remaining Q4 stock = %d, want 2", qualityAmount)
	}
	company, _ := store.GetCompany(ctx, companyID)
	if company.Inventory[11] != 40 {
		t.Fatalf("stocking Q4 changed legacy Q0 inventory: %d", company.Inventory[11])
	}
	if _, err := service.StockShelf(ctx, companyID, "retail-1", 11, 13, 1, 78); !apperr.HasKind(err, apperr.KindBadRequest) {
		t.Fatalf("Q13 stock error = %v, want bad request", err)
	}
}

func TestRetailSaleBatchIsAtomicAcrossServicesAndCompanyScoped(t *testing.T) {
	ctx := context.Background()
	firstService, store, firstCompanyID := newRetailBuildingService(t)
	buildings := map[int]*catalog.BuildingEntry{
		6: {ID: 6, Kind: 6, Name: "Market Stall", Type: "retail", Produces: []int{11}, RetailSlots: 2, SlotPerLevel: 1},
	}
	secondService := building.NewService(store, buildings, nil, nil, nil)

	errs := make([]error, 2)
	services := []*building.Service{firstService, secondService}
	var wg sync.WaitGroup
	for i := range services {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, errs[i] = services[i].StockShelf(ctx, firstCompanyID, "retail-1", 11, 0, 10, 78)
		}(i)
	}
	wg.Wait()

	successes, conflicts := 0, 0
	for _, err := range errs {
		switch {
		case err == nil:
			successes++
		case apperr.HasKind(err, apperr.KindConflict):
			conflicts++
		default:
			t.Fatalf("unexpected concurrent stock error: %v", err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("concurrent stock successes=%d conflicts=%d, want 1/1", successes, conflicts)
	}
	firstCompany, _ := store.GetCompany(ctx, firstCompanyID)
	if firstCompany.Inventory[11] != 30 || len(firstCompany.Buildings[0].Shelves) != 1 || firstCompany.Buildings[0].Shelves[0].Quantity != 10 {
		t.Fatalf("concurrent stock changed inventory or shelf twice: %+v", firstCompany)
	}

	if err := store.CreatePlayer(ctx, &auth.Player{ID: 72, Username: "other-retailer"}); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateCompany(ctx, &domain.Company{
		PlayerID: 72, Name: "Other Retail Corp", Money: 1000,
		Inventory: map[int]int{11: 40},
		Buildings: []domain.Building{{ID: "retail-1", BuildingID: 6, Kind: 6, Name: "Market Stall", Level: 1}},
	}); err != nil {
		t.Fatal(err)
	}
	otherCompany, _ := store.GetCompanyByPlayerID(ctx, 72)
	if _, err := firstService.StockShelf(ctx, otherCompany.ID, "retail-1", 11, 0, 10, 78); err != nil {
		t.Fatalf("another company must have an independent shelf namespace: %v", err)
	}
	otherCompany, _ = store.GetCompany(ctx, otherCompany.ID)
	if otherCompany.Inventory[11] != 30 || len(otherCompany.Buildings[0].Shelves) != 1 {
		t.Fatalf("other company stock was not isolated: %+v", otherCompany)
	}
}
