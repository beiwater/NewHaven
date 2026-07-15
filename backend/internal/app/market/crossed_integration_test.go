package market_test

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/beiwater/NewHaven/backend/internal/app"
	"github.com/beiwater/NewHaven/backend/internal/catalog"
	"github.com/beiwater/NewHaven/backend/internal/config"
	"github.com/beiwater/NewHaven/backend/internal/domain/auth"
	"github.com/beiwater/NewHaven/backend/internal/domain/company"
	domainmarket "github.com/beiwater/NewHaven/backend/internal/domain/market"
	openapi "github.com/beiwater/NewHaven/backend/internal/generated/openapi"
	"github.com/beiwater/NewHaven/backend/internal/storage/memory"
)

// TestCrossedMarket_MatchAllOrders verifies MatchAllOrders clears crossed orders.
func TestCrossedMarket_MatchAllOrders(t *testing.T) {
	store := memory.New()

	_ = store.CreatePlayer(context.Background(), &auth.Player{ID: 1, Username: "alice", PasswordHash: "hash"})
	_ = store.CreateCompany(context.Background(), &company.Company{
		PlayerID: 1, Name: "Alice Corp", Money: 100000, Inventory: map[int]int{},
	})
	alice, _ := store.GetCompanyByPlayerID(context.Background(), 1)

	_ = store.CreatePlayer(context.Background(), &auth.Player{ID: 2, Username: "bob", PasswordHash: "hash"})
	_ = store.CreateCompany(context.Background(), &company.Company{
		PlayerID: 2, Name: "Bob Corp", Money: 0, Inventory: map[int]int{1: 10000},
	})
	bob, _ := store.GetCompanyByPlayerID(context.Background(), 2)

	// Inject the exact crossed orders from the user's report.
	_ = store.CreateOrder(context.Background(), &domainmarket.MarketOrder{
		ID: "sell-10", ResourceID: 1, IsBuy: false, Price: 10.00, Quantity: 400, Quality: 0, Status: domainmarket.StatusOpen, CompanyID: bob.ID,
	})
	_ = store.CreateOrder(context.Background(), &domainmarket.MarketOrder{
		ID: "sell-1390", ResourceID: 1, IsBuy: false, Price: 13.90, Quantity: 100, Quality: 0, Status: domainmarket.StatusOpen, CompanyID: bob.ID,
	})
	_ = store.CreateOrder(context.Background(), &domainmarket.MarketOrder{
		ID: "sell-1400", ResourceID: 1, IsBuy: false, Price: 14.00, Quantity: 10, Quality: 0, Status: domainmarket.StatusOpen, CompanyID: bob.ID,
	})
	_ = store.CreateOrder(context.Background(), &domainmarket.MarketOrder{
		ID: "sell-1410", ResourceID: 1, IsBuy: false, Price: 14.10, Quantity: 10, Quality: 0, Status: domainmarket.StatusOpen, CompanyID: bob.ID,
	})
	_ = store.CreateOrder(context.Background(), &domainmarket.MarketOrder{
		ID: "buy-1410", ResourceID: 1, IsBuy: true, Price: 14.10, Quantity: 577, Quality: 0, Status: domainmarket.StatusOpen, CompanyID: alice.ID,
	})
	_ = store.CreateOrder(context.Background(), &domainmarket.MarketOrder{
		ID: "sell-1550", ResourceID: 1, IsBuy: false, Price: 15.50, Quantity: 10, Quality: 0, Status: domainmarket.StatusOpen, CompanyID: bob.ID,
	})

	// MUST provide resources so MatchAllOrders iterates over them.
	resources := map[int]*catalog.ResourceEntry{
		1: {ID: 1, DbLetter: 1, Name: "Test", IsExchangeTradable: true, BasePrice: 14},
	}
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	a := app.New(cfg, store, resources, nil, nil)
	svc := a.MarketService

	// Before: verify crossed orders exist in depth.
	depth, err := svc.GetMarketDepth(context.Background(), 1, 0)
	if err != nil {
		t.Fatalf("GetMarketDepth before: %v", err)
	}
	t.Logf("BEFORE — buys: %s, sells: %s", depthStr(depth.Buys), depthStr(depth.Sells))

	if depth.Sells == nil || len(*depth.Sells) == 0 {
		t.Fatal("expected sells before matching")
	}

	// Run MatchAllOrders (simulating scheduler tick).
	if err := svc.MatchAllOrders(context.Background()); err != nil {
		t.Fatalf("MatchAllOrders: %v", err)
	}

	// After: crossed orders should be gone.
	depth, err = svc.GetMarketDepth(context.Background(), 1, 0)
	if err != nil {
		t.Fatalf("GetMarketDepth after: %v", err)
	}
	t.Logf("AFTER — buys: %s, sells: %s", depthStr(depth.Buys), depthStr(depth.Sells))

	if depth.Buys == nil || len(*depth.Buys) == 0 {
		t.Fatal("expected buy levels after matching")
	}
	bestBuy := (*depth.Buys)[0]
	if bestBuy.Price == nil || *bestBuy.Price != 14.10 {
		t.Errorf("expected best buy price 14.10, got %v", bestBuy.Price)
	}
	if bestBuy.Quantity == nil || *bestBuy.Quantity != 57 {
		t.Errorf("expected best buy quantity 57, got %v", bestBuy.Quantity)
	}

	if depth.Sells != nil {
		for _, s := range *depth.Sells {
			if s.Price != nil && *s.Price <= 14.10 {
				t.Errorf("REGRESSION: crossed sell at $%.2f should have been matched away", *s.Price)
			}
		}
	}
}

// TestCrossedMarket_CreateOrderMatches verifies placing orders through CreateOrder
// immediately triggers matching (user-facing flow).
func TestCrossedMarket_CreateOrderMatches(t *testing.T) {
	store := memory.New()
	_ = store.CreatePlayer(context.Background(), &auth.Player{ID: 1, Username: "alice", PasswordHash: "hash"})
	_ = store.CreateCompany(context.Background(), &company.Company{
		PlayerID: 1, Name: "Alice Corp", Money: 100000, Inventory: map[int]int{},
	})
	alice, _ := store.GetCompanyByPlayerID(context.Background(), 1)

	_ = store.CreatePlayer(context.Background(), &auth.Player{ID: 2, Username: "bob", PasswordHash: "hash"})
	_ = store.CreateCompany(context.Background(), &company.Company{
		PlayerID: 2, Name: "Bob Corp", Money: 0, Inventory: map[int]int{1: 10000},
	})
	bob, _ := store.GetCompanyByPlayerID(context.Background(), 2)

	resources := map[int]*catalog.ResourceEntry{
		1: {ID: 1, DbLetter: 1, Name: "Test", IsExchangeTradable: true, BasePrice: 14},
	}
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	a := app.New(cfg, store, resources, nil, nil)
	svc := a.MarketService

	// Seller first: place sell at $10.00 x 400 via CreateOrder.
	_, err := svc.CreateOrder(context.Background(), bob.ID, &openapi.CreateOrderRequestFrontend{
		Kind: 0, ResourceId: 1, Quality: 0, Quantity: 400, Price: 10.00,
	})
	if err != nil {
		t.Fatalf("create sell: %v", err)
	}

	// Buyer: place buy at $14.10 x 577.
	_, err = svc.CreateOrder(context.Background(), alice.ID, &openapi.CreateOrderRequestFrontend{
		Kind: 1, ResourceId: 1, Quality: 0, Quantity: 577, Price: 14.10,
	})
	if err != nil {
		t.Fatalf("create buy: %v", err)
	}

	depth, _ := svc.GetMarketDepth(context.Background(), 1, 0)
	t.Logf("After buy — buys: %s, sells: %s", depthStr(depth.Buys), depthStr(depth.Sells))

	if depth.Buys == nil {
		t.Fatal("expected buys after matching")
	}
	if len(*depth.Buys) != 1 {
		t.Fatalf("expected 1 buy level, got %d", len(*depth.Buys))
	}
	bestBuy := (*depth.Buys)[0]
	if bestBuy.Price == nil || *bestBuy.Price != 14.10 {
		t.Errorf("expected best buy price 14.10, got %v", bestBuy.Price)
	}
	if bestBuy.Quantity == nil || *bestBuy.Quantity != 177 {
		t.Errorf("expected best buy quantity 177 (after matching 400 of 577), got %v", bestBuy.Quantity)
	}

	// Now place the remaining sells: $13.90 x 100, $14.00 x 10, $14.10 x 10.
	_, _ = svc.CreateOrder(context.Background(), bob.ID, &openapi.CreateOrderRequestFrontend{
		Kind: 0, ResourceId: 1, Quality: 0, Quantity: 100, Price: 13.90,
	})
	_, _ = svc.CreateOrder(context.Background(), bob.ID, &openapi.CreateOrderRequestFrontend{
		Kind: 0, ResourceId: 1, Quality: 0, Quantity: 10, Price: 14.00,
	})
	_, _ = svc.CreateOrder(context.Background(), bob.ID, &openapi.CreateOrderRequestFrontend{
		Kind: 0, ResourceId: 1, Quality: 0, Quantity: 10, Price: 14.10,
	})
	_, _ = svc.CreateOrder(context.Background(), bob.ID, &openapi.CreateOrderRequestFrontend{
		Kind: 0, ResourceId: 1, Quality: 0, Quantity: 10, Price: 15.50,
	})

	depth, _ = svc.GetMarketDepth(context.Background(), 1, 0)
	t.Logf("Final — buys: %s, sells: %s", depthStr(depth.Buys), depthStr(depth.Sells))

	// Buy should have 57 remaining.
	if depth.Buys != nil && len(*depth.Buys) > 0 {
		bestBuy := (*depth.Buys)[0]
		if bestBuy.Price == nil || *bestBuy.Price != 14.10 {
			t.Errorf("expected best buy price 14.10, got %v", bestBuy.Price)
		}
		if bestBuy.Quantity == nil || *bestBuy.Quantity != 57 {
			t.Errorf("expected best buy quantity 57, got %v", bestBuy.Quantity)
		}
	}
	// No sells <= 14.10.
	if depth.Sells != nil {
		for _, s := range *depth.Sells {
			if s.Price != nil && *s.Price <= 14.10 {
				t.Errorf("REGRESSION: sell at $%.2f should have been matched away", *s.Price)
			}
		}
	}
}

func depthStr(levels *[]openapi.MarketDepthLevel) string {
	if levels == nil {
		return "∅"
	}
	parts := make([]string, 0, len(*levels))
	for _, l := range *levels {
		p := "?"
		if l.Price != nil {
			p = fmt.Sprintf("%.2f", *l.Price)
		}
		q := "?"
		if l.Quantity != nil {
			q = fmt.Sprintf("%d", *l.Quantity)
		}
		parts = append(parts, "$"+p+"x"+q)
	}
	return "[" + strings.Join(parts, ", ") + "]"
}
