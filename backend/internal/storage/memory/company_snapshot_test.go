package memory_test

import (
	"context"
	"errors"
	"testing"

	"github.com/beiwater/NewHaven/backend/internal/domain/company"
	"github.com/beiwater/NewHaven/backend/internal/storage"
	"github.com/beiwater/NewHaven/backend/internal/storage/memory"
)

func TestGetCompanyReturnsDetachedSnapshot(t *testing.T) {
	// Given
	ctx := context.Background()
	store := memory.New()
	seed := &company.Company{
		PlayerID:    7,
		Money:       1000,
		Preferences: map[string]any{"story": map[string]any{"status": "not_started"}},
		Buildings: []company.Building{{
			ID: "shop-1", Level: 1,
			Shelves: []company.ShelfItem{{ResourceID: 1, Quantity: 5}},
		}},
		Inventory:   map[int]int{1: 10},
		Executives:  []company.Executive{{ID: "exec-1", Level: 1}},
		RetailCarry: map[string]float64{"shop-1": 0.25},
	}
	if err := store.CreateCompany(ctx, seed); err != nil {
		t.Fatal(err)
	}

	// When
	snapshot, err := store.GetCompany(ctx, seed.ID)
	if err != nil {
		t.Fatal(err)
	}
	snapshot.Money = 0
	snapshot.Preferences["story"].(map[string]any)["status"] = "completed"
	snapshot.Buildings[0].Level = 2
	snapshot.Buildings[0].Shelves[0].Quantity = 0
	snapshot.Inventory[1] = 0
	snapshot.Executives[0].Level = 2
	snapshot.RetailCarry["shop-1"] = 0.75

	// Then
	persisted, err := store.GetCompany(ctx, seed.ID)
	if err != nil {
		t.Fatal(err)
	}
	if persisted.Money != 1000 ||
		persisted.Preferences["story"].(map[string]any)["status"] != "not_started" ||
		persisted.Buildings[0].Level != 1 || persisted.Buildings[0].Shelves[0].Quantity != 5 ||
		persisted.Inventory[1] != 10 || persisted.Executives[0].Level != 1 ||
		persisted.RetailCarry["shop-1"] != 0.25 {
		t.Fatalf("company read leaked mutable store state: %+v", persisted)
	}
}

func TestUpdateCompanyRefreshesPlayerLookupAfterDetachedRead(t *testing.T) {
	// Given
	ctx := context.Background()
	store := memory.New()
	seed := &company.Company{PlayerID: 8, Money: 1000}
	if err := store.CreateCompany(ctx, seed); err != nil {
		t.Fatal(err)
	}
	snapshot, err := store.GetCompany(ctx, seed.ID)
	if err != nil {
		t.Fatal(err)
	}

	// When
	snapshot.Money = 750
	if err := store.UpdateCompany(ctx, snapshot); err != nil {
		t.Fatal(err)
	}
	snapshot.Money = 0

	// Then
	byPlayer, err := store.GetCompanyByPlayerID(ctx, seed.PlayerID)
	if err != nil {
		t.Fatal(err)
	}
	if byPlayer.Money != 750 {
		t.Fatalf("player lookup returned stale company money: got %.0f want 750", byPlayer.Money)
	}
}

func TestUpdateCompanyRejectsPlayerOwnershipChange(t *testing.T) {
	// Given
	ctx := context.Background()
	store := memory.New()
	owner := &company.Company{PlayerID: 9, Name: "Owner"}
	other := &company.Company{PlayerID: 10, Name: "Other"}
	if err := store.CreateCompany(ctx, owner); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateCompany(ctx, other); err != nil {
		t.Fatal(err)
	}
	snapshot, err := store.GetCompany(ctx, owner.ID)
	if err != nil {
		t.Fatal(err)
	}

	// When
	snapshot.PlayerID = other.PlayerID
	err = store.UpdateCompany(ctx, snapshot)

	// Then
	if !errors.Is(err, storage.ErrStateConflict) {
		t.Fatalf("ownership change should conflict: %v", err)
	}
	otherLookup, err := store.GetCompanyByPlayerID(ctx, other.PlayerID)
	if err != nil {
		t.Fatal(err)
	}
	if otherLookup.ID != other.ID {
		t.Fatalf("ownership change replaced another player's company: got %d want %d", otherLookup.ID, other.ID)
	}
}
