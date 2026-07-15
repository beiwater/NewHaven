package building_test

import (
	"context"
	"errors"
	"testing"

	"github.com/beiwater/NewHaven/backend/internal/app/building"
	"github.com/beiwater/NewHaven/backend/internal/catalog"
	"github.com/beiwater/NewHaven/backend/internal/domain/auth"
	domain "github.com/beiwater/NewHaven/backend/internal/domain/company"
	openapi "github.com/beiwater/NewHaven/backend/internal/generated/openapi"
	"github.com/beiwater/NewHaven/backend/internal/platform"
	"github.com/beiwater/NewHaven/backend/internal/storage"
	"github.com/beiwater/NewHaven/backend/internal/storage/memory"
)

func TestListMyBuildings_ReturnsBuildings(t *testing.T) {
	ctx := context.Background()
	store := memory.New()

	err := store.CreatePlayer(ctx, &auth.Player{ID: 1, Username: "alice"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID:  1,
		Name:      "Alice Corp",
		Money:     100000,
		Level:     1,
		XP:        0,
		Inventory: make(map[int]int),
		Buildings: []domain.Building{
			{ID: "bld-1-1", BuildingID: 1, Kind: 1, Name: "Bakery", Level: 1, MapID: "map_1", SlotID: "slot_a1", X: 5, Y: 10},
			{ID: "bld-1-2", BuildingID: 2, Kind: 2, Name: "Workshop", Level: 1, MapID: "map_1", SlotID: "slot_b1", X: 15, Y: 20},
		},
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}

	company, err := store.GetCompanyByPlayerID(ctx, 1)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}

	svc := building.NewService(store, nil, nil, nil, nil)

	resp, err := svc.ListMyBuildings(ctx, company.ID)
	if err != nil {
		t.Fatalf("ListMyBuildings failed: %v", err)
	}

	if resp.Buildings == nil {
		t.Fatal("expected non-nil buildings array")
	}

	if len(*resp.Buildings) != 2 {
		t.Fatalf("expected 2 buildings, got %d", len(*resp.Buildings))
	}

	// Check first building
	b := (*resp.Buildings)[0]
	if b.Name == nil || *b.Name != "Bakery" {
		t.Errorf("expected first building name 'Bakery', got %v", b.Name)
	}
	if b.Level == nil || *b.Level != 1 {
		t.Errorf("expected level 1, got %v", b.Level)
	}
	if b.Id == nil || *b.Id == "" {
		t.Error("expected non-empty building id")
	}
	if b.BuildingId == nil || *b.BuildingId <= 0 {
		t.Errorf("expected positive building_id, got %v", b.BuildingId)
	}

	// Check second building
	b2 := (*resp.Buildings)[1]
	if b2.Name == nil || *b2.Name != "Workshop" {
		t.Errorf("expected second building name 'Workshop', got %v", b2.Name)
	}
}

func TestListMyBuildings_CompanyNotFound(t *testing.T) {
	ctx := context.Background()
	store := memory.New()

	svc := building.NewService(store, nil, nil, nil, nil)

	_, err := svc.ListMyBuildings(ctx, 99999)
	if err == nil {
		t.Fatal("expected error for non-existent company, got nil")
	}
}

func TestListMyBuildings_EmptyBuildings(t *testing.T) {
	ctx := context.Background()
	store := memory.New()

	err := store.CreatePlayer(ctx, &auth.Player{ID: 2, Username: "bob"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}

	c := &domain.Company{
		PlayerID:  2,
		Name:      "Bob Corp",
		Money:     100000,
		Level:     1,
		XP:        0,
		Inventory: make(map[int]int),
	}
	err = store.CreateCompany(ctx, c)
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}

	// Clear buildings created by CreateCompany to test empty case
	company, err := store.GetCompanyByPlayerID(ctx, 2)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}
	company.Buildings = nil
	if err := store.UpdateCompany(ctx, company); err != nil {
		t.Fatalf("UpdateCompany: %v", err)
	}

	svc := building.NewService(store, nil, nil, nil, nil)

	resp, err := svc.ListMyBuildings(ctx, company.ID)
	if err != nil {
		t.Fatalf("ListMyBuildings failed: %v", err)
	}

	if resp.Buildings == nil {
		t.Fatal("expected non-nil buildings array (should be empty)")
	}

	if len(*resp.Buildings) != 0 {
		t.Errorf("expected 0 buildings, got %d", len(*resp.Buildings))
	}
}

// rollbackStore wraps storage.CompanyStorage to inject failures in UpdateCompany.
type rollbackStore struct {
	storage.CompanyStorage
	failOn int
	calls  int
}

func (s *rollbackStore) UpdateCompany(ctx context.Context, c *domain.Company) error {
	s.calls++
	if s.calls == s.failOn {
		return errors.New("injected UpdateCompany failure")
	}
	return s.CompanyStorage.UpdateCompany(ctx, c)
}

func TestRollbackBuyBuildingOnFailure(t *testing.T) {
	ctx := context.Background()
	store := memory.New()

	err := store.CreatePlayer(ctx, &auth.Player{ID: 20, Username: "buyrollback"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID:  20,
		Name:      "BuyRollback Corp",
		Money:     100000,
		Level:     5,
		XP:        0,
		Inventory: make(map[int]int),
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}

	company, err := store.GetCompanyByPlayerID(ctx, 20)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}
	origMoney := company.Money
	origLen := len(company.Buildings)

	// Wrap store to fail on 1st UpdateCompany
	wrap := &rollbackStore{CompanyStorage: store, failOn: 1}

	buildings := map[int]*catalog.BuildingEntry{
		1: {ID: 1, Kind: 1, Name: "Bakery", BaseCost: 5000},
	}
	svc := building.NewService(wrap, buildings, nil, nil, platform.NewIDGen())

	_, err = svc.BuyBuilding(ctx, company.ID, "b-shop-1")
	if err == nil {
		t.Fatal("expected error from injected UpdateCompany failure")
	}

	// Re-fetch and verify rollback
	updated, err := store.GetCompany(ctx, company.ID)
	if err != nil {
		t.Fatalf("GetCompany: %v", err)
	}
	if updated.Money != origMoney {
		t.Errorf("expected money %f, got %f", origMoney, updated.Money)
	}
	if len(updated.Buildings) != origLen {
		t.Errorf("expected %d buildings, got %d", origLen, len(updated.Buildings))
	}
}

func TestRollbackUpgradeBuildingOnFailure(t *testing.T) {
	ctx := context.Background()
	store := memory.New()

	err := store.CreatePlayer(ctx, &auth.Player{ID: 21, Username: "upgrollback"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID:  21,
		Name:      "UpgRollback Corp",
		Money:     100000,
		Level:     5,
		XP:        0,
		Inventory: make(map[int]int),
		Buildings: []domain.Building{
			{ID: "bld-21-1", BuildingID: 1, Kind: 1, Name: "Bakery", Level: 1, MapID: "map_1", SlotID: "slot_a1", X: 5, Y: 10},
			{ID: "bld-21-2", BuildingID: 2, Kind: 2, Name: "Workshop", Level: 1, MapID: "map_1", SlotID: "slot_b1", X: 15, Y: 20},
		},
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}

	company, err := store.GetCompanyByPlayerID(ctx, 21)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}
	origMoney := company.Money
	origLevel := company.Buildings[0].Level
	buildingID := company.Buildings[0].ID

	wrap := &rollbackStore{CompanyStorage: store, failOn: 1}
	svc := building.NewService(wrap, nil, nil, nil, nil)

	_, err = svc.UpgradeBuilding(ctx, company.ID, buildingID)
	if err == nil {
		t.Fatal("expected error from injected UpdateCompany failure")
	}

	// Re-fetch and verify rollback
	updated, err := store.GetCompany(ctx, company.ID)
	if err != nil {
		t.Fatalf("GetCompany: %v", err)
	}
	if updated.Money != origMoney {
		t.Errorf("expected money %f, got %f", origMoney, updated.Money)
	}
	if updated.Buildings[0].Level != origLevel {
		t.Errorf("expected building level %d, got %d", origLevel, updated.Buildings[0].Level)
	}
}

func TestUpgradeBuildingUsesCurrentSizeCost(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	if err := store.CreatePlayer(ctx, &auth.Player{ID: 31, Username: "upgradecost"}); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateCompany(ctx, &domain.Company{
		PlayerID: 31, Name: "Upgrade Cost Corp", Money: 100000, Level: 10,
		Inventory: make(map[int]int),
		Buildings: []domain.Building{{ID: "upgrade-me", BuildingID: 1, Kind: 1, Name: "Farm", Level: 3}},
	}); err != nil {
		t.Fatal(err)
	}
	company, _ := store.GetCompanyByPlayerID(ctx, 31)
	svc := building.NewService(store, map[int]*catalog.BuildingEntry{
		1: {ID: 1, Kind: 1, Name: "Farm", BaseCost: 10000},
	}, nil, nil, platform.NewIDGen())

	resp, err := svc.UpgradeBuilding(ctx, company.ID, "upgrade-me")
	if err != nil {
		t.Fatal(err)
	}
	if resp.Cost == nil || *resp.Cost != 30000 {
		t.Fatalf("level-3 upgrade should cost currentSize*baseCost = 30000, got %v", resp.Cost)
	}
}

func TestRollbackPlaceBuildingOnFailure(t *testing.T) {
	ctx := context.Background()
	store := memory.New()

	err := store.CreatePlayer(ctx, &auth.Player{ID: 22, Username: "placerollback"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID:  22,
		Name:      "PlaceRollback Corp",
		Money:     100000,
		Level:     5,
		XP:        0,
		Inventory: make(map[int]int),
		Buildings: []domain.Building{
			{ID: "bld-22-1", BuildingID: 1, Kind: 1, Name: "Bakery", Level: 1, MapID: "map_1", SlotID: "slot_a1", X: 5, Y: 10},
			{ID: "bld-22-2", BuildingID: 2, Kind: 2, Name: "Workshop", Level: 1, MapID: "map_1", SlotID: "slot_b1", X: 15, Y: 20},
		},
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}

	company, err := store.GetCompanyByPlayerID(ctx, 22)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}

	// Make both buildings unplaced so there's no slot conflict
	for i := range company.Buildings {
		company.Buildings[i].MapID = ""
		company.Buildings[i].SlotID = ""
	}
	if err := store.UpdateCompany(ctx, company); err != nil {
		t.Fatalf("UpdateCompany: %v", err)
	}

	buildingID := company.Buildings[0].ID
	origMapID := company.Buildings[0].MapID
	origSlotID := company.Buildings[0].SlotID
	origX := company.Buildings[0].X
	origY := company.Buildings[0].Y

	wrap := &rollbackStore{CompanyStorage: store, failOn: 1}
	svc := building.NewService(wrap, nil, nil, nil, nil)

	mapID := "harbor"
	slotID := "harbor-plot-01"
	xVal := 10
	yVal := 20
	_, err = svc.PlaceBuilding(ctx, company.ID, &openapi.PlaceBuildingRequest{
		BuildingId: buildingID,
		MapId:      &mapID,
		SlotId:     &slotID,
		X:          &xVal,
		Y:          &yVal,
	})
	if err == nil {
		t.Fatal("expected error from injected UpdateCompany failure")
	}

	// Re-fetch and verify rollback
	updated, err := store.GetCompany(ctx, company.ID)
	if err != nil {
		t.Fatalf("GetCompany: %v", err)
	}
	for _, b := range updated.Buildings {
		if b.ID == buildingID {
			if b.MapID != origMapID {
				t.Errorf("expected MapID %q, got %q", origMapID, b.MapID)
			}
			if b.SlotID != origSlotID {
				t.Errorf("expected SlotID %q, got %q", origSlotID, b.SlotID)
			}
			if b.X != origX {
				t.Errorf("expected X %d, got %d", origX, b.X)
			}
			if b.Y != origY {
				t.Errorf("expected Y %d, got %d", origY, b.Y)
			}
			break
		}
	}
}

func TestRollbackMoveBuildingOnFailure(t *testing.T) {
	ctx := context.Background()
	store := memory.New()

	err := store.CreatePlayer(ctx, &auth.Player{ID: 23, Username: "moverollback"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID:  23,
		Name:      "MoveRollback Corp",
		Money:     100000,
		Level:     5,
		XP:        0,
		Inventory: make(map[int]int),
		Buildings: []domain.Building{
			{ID: "bld-23-1", BuildingID: 1, Kind: 1, Name: "Bakery", Level: 1, MapID: "map_1", SlotID: "slot_a1", X: 5, Y: 10},
			{ID: "bld-23-2", BuildingID: 2, Kind: 2, Name: "Workshop", Level: 1, MapID: "map_1", SlotID: "slot_b1", X: 15, Y: 20},
		},
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}

	company, err := store.GetCompanyByPlayerID(ctx, 23)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}

	buildingID := company.Buildings[0].ID
	origMapID := company.Buildings[0].MapID
	origSlotID := company.Buildings[0].SlotID
	origX := company.Buildings[0].X
	origY := company.Buildings[0].Y

	wrap := &rollbackStore{CompanyStorage: store, failOn: 1}
	svc := building.NewService(wrap, nil, nil, nil, nil)

	mapID := "desert"
	slotID := "desert-plot-01"
	xVal := 30
	yVal := 40
	_, err = svc.MoveBuilding(ctx, company.ID, &openapi.MoveBuildingRequest{
		BuildingId: buildingID,
		MapId:      &mapID,
		SlotId:     &slotID,
		X:          &xVal,
		Y:          &yVal,
	})
	if err == nil {
		t.Fatal("expected error from injected UpdateCompany failure")
	}

	// Re-fetch and verify rollback
	updated, err := store.GetCompany(ctx, company.ID)
	if err != nil {
		t.Fatalf("GetCompany: %v", err)
	}
	for _, b := range updated.Buildings {
		if b.ID == buildingID {
			if b.MapID != origMapID {
				t.Errorf("expected MapID %q, got %q", origMapID, b.MapID)
			}
			if b.SlotID != origSlotID {
				t.Errorf("expected SlotID %q, got %q", origSlotID, b.SlotID)
			}
			if b.X != origX {
				t.Errorf("expected X %d, got %d", origX, b.X)
			}
			if b.Y != origY {
				t.Errorf("expected Y %d, got %d", origY, b.Y)
			}
			break
		}
	}
}

func TestRollbackDemolishBuildingOnFailure(t *testing.T) {
	ctx := context.Background()
	store := memory.New()

	err := store.CreatePlayer(ctx, &auth.Player{ID: 24, Username: "demorollback"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID:  24,
		Name:      "DemoRollback Corp",
		Money:     100000,
		Level:     5,
		XP:        0,
		Inventory: make(map[int]int),
		Buildings: []domain.Building{
			{ID: "bld-24-1", BuildingID: 1, Kind: 1, Name: "Bakery", Level: 1, MapID: "map_1", SlotID: "slot_a1", X: 5, Y: 10},
			{ID: "bld-24-2", BuildingID: 2, Kind: 2, Name: "Workshop", Level: 1, MapID: "map_1", SlotID: "slot_b1", X: 15, Y: 20},
		},
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}

	company, err := store.GetCompanyByPlayerID(ctx, 24)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}
	origMoney := company.Money
	origBuildings := make([]domain.Building, len(company.Buildings))
	copy(origBuildings, company.Buildings)

	buildingID := company.Buildings[0].ID

	wrap := &rollbackStore{CompanyStorage: store, failOn: 1}
	svc := building.NewService(wrap, nil, nil, nil, nil)

	_, err = svc.DemolishBuilding(ctx, company.ID, &openapi.DemolishBuildingRequest{
		BuildingId: buildingID,
	})
	if err == nil {
		t.Fatal("expected error from injected UpdateCompany failure")
	}

	// Re-fetch and verify rollback
	updated, err := store.GetCompany(ctx, company.ID)
	if err != nil {
		t.Fatalf("GetCompany: %v", err)
	}
	if updated.Money != origMoney {
		t.Errorf("expected money %f, got %f", origMoney, updated.Money)
	}
	if len(updated.Buildings) != len(origBuildings) {
		t.Fatalf("expected %d buildings, got %d", len(origBuildings), len(updated.Buildings))
	}
	// Verify each building matches original order
	for i := range origBuildings {
		if updated.Buildings[i].ID != origBuildings[i].ID {
			t.Errorf("building[%d] ID mismatch: expected %s, got %s", i, origBuildings[i].ID, updated.Buildings[i].ID)
		}
	}
}
