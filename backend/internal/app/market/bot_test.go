package market_test

import (
	"context"
	"testing"

	"github.com/beiwater/NewHaven/backend/internal/catalog"
	"github.com/beiwater/NewHaven/backend/internal/domain/company"
)

func TestEnsureBotCompanies_RepairsQualityZeroWarehouseDrift(t *testing.T) {
	t.Parallel()
	svc, store := newTestSvc(map[int]*catalog.ResourceEntry{
		1: {ID: 1, DbLetter: 1, Name: "Grain", IsExchangeTradable: true},
	})
	ctx := context.Background()

	bot := &company.Company{
		PlayerID:  -900001,
		Name:      "Atlas Trading Bot",
		Money:     5_000_000,
		Inventory: map[int]int{1: 7},
	}
	if err := store.CreateCompany(ctx, bot); err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}

	// Reproduce an old restart path that updated only the legacy inventory map.
	bot.Inventory[1] = 999999
	if err := store.UpdateCompany(ctx, bot); err != nil {
		t.Fatalf("UpdateCompany: %v", err)
	}

	if err := svc.EnsureBotCompanies(ctx); err != nil {
		t.Fatalf("EnsureBotCompanies: %v", err)
	}

	updated, err := store.GetCompany(ctx, bot.ID)
	if err != nil {
		t.Fatalf("GetCompany: %v", err)
	}
	if updated.Inventory[1] != 999999 {
		t.Fatalf("legacy Q0 inventory = %d, want 999999", updated.Inventory[1])
	}

	stock, err := store.GetWarehouse(ctx, bot.ID)
	if err != nil {
		t.Fatalf("GetWarehouse: %v", err)
	}
	if len(stock.Items) != 1 || stock.Items[0].ResourceID != 1 || stock.Items[0].Quality != 0 || stock.Items[0].Amount != 999999 {
		t.Fatalf("warehouse was not repaired: %+v", stock.Items)
	}
}
