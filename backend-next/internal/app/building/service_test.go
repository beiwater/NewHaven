package building_test

import (
	"context"
	"testing"

	"github.com/newhaven/backend-next/internal/app/building"
	"github.com/newhaven/backend-next/internal/domain/auth"
	domain "github.com/newhaven/backend-next/internal/domain/company"
	"github.com/newhaven/backend-next/internal/storage/memory"
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
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}

	company, err := store.GetCompanyByPlayerID(ctx, 1)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}

	svc := building.NewService(store)

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

	svc := building.NewService(store)

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

	svc := building.NewService(store)

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
