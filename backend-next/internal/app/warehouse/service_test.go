package warehouse_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/newhaven/backend-next/internal/app/warehouse"
	"github.com/newhaven/backend-next/internal/domain/auth"
	domain "github.com/newhaven/backend-next/internal/domain/company"
	whdomain "github.com/newhaven/backend-next/internal/domain/warehouse"
	"github.com/newhaven/backend-next/internal/platform"
	"github.com/newhaven/backend-next/internal/storage"
	"github.com/newhaven/backend-next/internal/storage/memory"
)

func TestGetMyWarehouse_ReturnsWarehouse(t *testing.T) {
	ctx := context.Background()
	store := memory.New()

	err := store.CreatePlayer(ctx, &auth.Player{ID: 1, Username: "alice"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID: 1,
		Name:     "Alice Co",
		Money:    1000.50,
		Level:    5,
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}

	// Get the company to find its assigned ID
	company, err := store.GetCompanyByPlayerID(ctx, 1)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}

	logger := platform.NewLogger(slog.Default())
	svc := warehouse.NewService(store, store, nil, logger)

	resp, err := svc.GetMyWarehouse(ctx, company.ID)
	if err != nil {
		t.Fatalf("GetMyWarehouse failed: %v", err)
	}

	if resp.CompanyId == nil || *resp.CompanyId != company.ID {
		t.Errorf("expected company_id %d, got %v", company.ID, resp.CompanyId)
	}
	if resp.Capacity == nil || *resp.Capacity != 1000 {
		t.Errorf("expected capacity 1000, got %v", resp.Capacity)
	}
	if resp.UsedCapacity == nil || *resp.UsedCapacity != 0 {
		t.Errorf("expected used_capacity 0, got %v", resp.UsedCapacity)
	}
	if resp.Items == nil {
		t.Fatal("expected non-nil items array")
	}
	if len(*resp.Items) != 0 {
		t.Errorf("expected empty items, got %d items", len(*resp.Items))
	}
}

func TestGetMyWarehouse_EmptyItemsReturnsEmptyArray(t *testing.T) {
	ctx := context.Background()
	store := memory.New()

	err := store.CreatePlayer(ctx, &auth.Player{ID: 2, Username: "bob"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID: 2,
		Name:     "Bob Co",
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}

	company, err := store.GetCompanyByPlayerID(ctx, 2)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}

	logger := platform.NewLogger(slog.Default())
	svc := warehouse.NewService(store, store, nil, logger)

	resp, err := svc.GetMyWarehouse(ctx, company.ID)
	if err != nil {
		t.Fatalf("GetMyWarehouse failed: %v", err)
	}

	if resp.Items == nil {
		t.Fatal("expected non-nil items array, got nil")
	}
	if len(*resp.Items) != 0 {
		t.Errorf("expected empty items, got %d items", len(*resp.Items))
	}
}

func TestGetMyWarehouse_CompanyNotFound(t *testing.T) {
	ctx := context.Background()
	store := memory.New()

	logger := platform.NewLogger(slog.Default())
	svc := warehouse.NewService(store, store, nil, logger)

	_, err := svc.GetMyWarehouse(ctx, 99999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, warehouse.ErrNotFound) {
		t.Logf("error: %v", err)
	}
}

// whCompanyFailStore wraps CompanyStorage to inject failures in UpdateCompany.
type whCompanyFailStore struct {
	storage.CompanyStorage
	failOn int
	calls  int
}

func (s *whCompanyFailStore) UpdateCompany(ctx context.Context, c *domain.Company) error {
	s.calls++
	if s.calls == s.failOn {
		return errors.New("injected UpdateCompany failure")
	}
	return s.CompanyStorage.UpdateCompany(ctx, c)
}

// whWarehouseFailStore wraps WarehouseStorage to inject failures in UpdateWarehouse.
type whWarehouseFailStore struct {
	storage.WarehouseStorage
	failOn int
	calls  int
}

func (s *whWarehouseFailStore) UpdateWarehouse(ctx context.Context, w *whdomain.Warehouse) error {
	s.calls++
	if s.calls == s.failOn {
		return errors.New("injected UpdateWarehouse failure")
	}
	return s.WarehouseStorage.UpdateWarehouse(ctx, w)
}

func TestRollbackWarehouseUpgradeOnFailure(t *testing.T) {
	ctx := context.Background()
	store := memory.New()

	err := store.CreatePlayer(ctx, &auth.Player{ID: 30, Username: "whupgfail"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID:  30,
		Name:      "WhUpgFail Corp",
		Money:     100000,
		Level:     5,
		XP:        0,
		Inventory: make(map[int]int),
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}

	company, err := store.GetCompanyByPlayerID(ctx, 30)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}
	origMoney := company.Money
	origWarehouseLevel := company.WarehouseLevel

	// Get original warehouse capacity
	origWh, err := store.GetWarehouse(ctx, company.ID)
	if err != nil {
		t.Fatalf("GetWarehouse: %v", err)
	}
	origCapacity := origWh.Capacity

	logger := platform.NewLogger(slog.Default())
	wrapCompanies := &whCompanyFailStore{CompanyStorage: store, failOn: 1}
	svc := warehouse.NewService(store, wrapCompanies, nil, logger)

	_, err = svc.UpgradeWarehouse(ctx, company.ID)
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
	if updated.WarehouseLevel != origWarehouseLevel {
		t.Errorf("expected warehouse level %d, got %d", origWarehouseLevel, updated.WarehouseLevel)
	}

	// Verify warehouse capacity was not changed in storage
	updatedWh, err := store.GetWarehouse(ctx, company.ID)
	if err != nil {
		t.Fatalf("GetWarehouse: %v", err)
	}
	if updatedWh.Capacity != origCapacity {
		t.Errorf("expected warehouse capacity %d, got %d", origCapacity, updatedWh.Capacity)
	}
}

func TestRollbackWarehouseUpgradeOnWarehouseFailure(t *testing.T) {
	ctx := context.Background()
	store := memory.New()

	err := store.CreatePlayer(ctx, &auth.Player{ID: 31, Username: "whwhfail"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID:  31,
		Name:      "WhWhFail Corp",
		Money:     100000,
		Level:     5,
		XP:        0,
		Inventory: make(map[int]int),
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}

	company, err := store.GetCompanyByPlayerID(ctx, 31)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}
	origMoney := company.Money
	origWarehouseLevel := company.WarehouseLevel

	// Get original warehouse capacity
	origWh, err := store.GetWarehouse(ctx, company.ID)
	if err != nil {
		t.Fatalf("GetWarehouse: %v", err)
	}
	origCapacity := origWh.Capacity

	logger := platform.NewLogger(slog.Default())
	wrapWarehouses := &whWarehouseFailStore{WarehouseStorage: store, failOn: 1}
	svc := warehouse.NewService(wrapWarehouses, store, nil, logger)

	_, err = svc.UpgradeWarehouse(ctx, company.ID)
	if err == nil {
		t.Fatal("expected error from injected UpdateWarehouse failure")
	}

	// Re-fetch and verify money & level rollback
	updated, err := store.GetCompany(ctx, company.ID)
	if err != nil {
		t.Fatalf("GetCompany: %v", err)
	}
	if updated.Money != origMoney {
		t.Errorf("expected money %f, got %f", origMoney, updated.Money)
	}
	if updated.WarehouseLevel != origWarehouseLevel {
		t.Errorf("expected warehouse level %d, got %d", origWarehouseLevel, updated.WarehouseLevel)
	}

	// Verify warehouse capacity was restored to originals
	updatedWh, err := store.GetWarehouse(ctx, company.ID)
	if err != nil {
		t.Fatalf("GetWarehouse: %v", err)
	}
	if updatedWh.Capacity != origCapacity {
		t.Errorf("expected warehouse capacity %d, got %d", origCapacity, updatedWh.Capacity)
	}
}
