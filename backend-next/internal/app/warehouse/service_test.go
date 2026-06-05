package warehouse_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/newhaven/backend-next/internal/app/warehouse"
	"github.com/newhaven/backend-next/internal/domain/auth"
	domain "github.com/newhaven/backend-next/internal/domain/company"
	"github.com/newhaven/backend-next/internal/platform"
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
	svc := warehouse.NewService(store, store, logger)

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
	svc := warehouse.NewService(store, store, logger)

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
	svc := warehouse.NewService(store, store, logger)

	_, err := svc.GetMyWarehouse(ctx, 99999)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, warehouse.ErrNotFound) {
		t.Logf("error: %v", err)
	}
}
