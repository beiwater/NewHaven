package building_test

import (
	"context"
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

	if _, err := service.StockShelf(ctx, companyID, "retail-1", 11, 10, 78); err != nil {
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

	if _, err := service.StockShelf(ctx, companyID, "retail-1", 11, 5, 80); !apperr.HasKind(err, apperr.KindConflict) {
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
	if _, err := service.StockShelf(ctx, companyID, "retail-1", 11, 5, 80); err != nil {
		t.Fatalf("new batch after sell-out: %v", err)
	}
}

func TestRetailSaleRequiresPositiveQuantityAndPrice(t *testing.T) {
	ctx := context.Background()
	service, _, companyID := newRetailBuildingService(t)
	if _, err := service.StockShelf(ctx, companyID, "retail-1", 11, 0, 78); !apperr.HasKind(err, apperr.KindBadRequest) {
		t.Fatalf("zero quantity error=%v, want bad request", err)
	}
	if _, err := service.StockShelf(ctx, companyID, "retail-1", 11, 10, 0); !apperr.HasKind(err, apperr.KindBadRequest) {
		t.Fatalf("zero price error=%v, want bad request", err)
	}
}
