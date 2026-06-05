package market_test

import (
	"context"
	"testing"
	"time"

	appmarket "github.com/newhaven/backend-next/internal/app/market"
	"github.com/newhaven/backend-next/internal/catalog"
	domainmarket "github.com/newhaven/backend-next/internal/domain/market"
	"github.com/newhaven/backend-next/internal/platform"
	"github.com/newhaven/backend-next/internal/storage/memory"
)

func newTestSvc(resources map[int]*catalog.ResourceEntry) (*appmarket.Service, *memory.Store) {
	store := memory.New()
	if resources == nil {
		resources = make(map[int]*catalog.ResourceEntry)
	}
	clock := platform.NewFakeClock(time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC))
	svc := appmarket.NewService(store, resources, clock)
	return svc, store
}

func TestListResources_FiltersCorrectly(t *testing.T) {
	ctx := context.Background()
	resources := map[int]*catalog.ResourceEntry{
		1: {ID: 1, DbLetter: 1, Name: "Grain", IsExchangeTradable: true, IsResearch: false, ProducedPerHourRaw: 500, UnitsSoldAnHour: 150, HasEconomyModel: true, BasePrice: 23, ProducedFrom: map[int]int{}},
		2: {ID: 2, DbLetter: 2, Name: "Dairy Milk", IsExchangeTradable: true, IsResearch: false, ProducedPerHourRaw: 420, UnitsSoldAnHour: 130, HasEconomyModel: true, BasePrice: 28},
		3: {ID: 3, DbLetter: 3, Name: "ResearchOnly", IsExchangeTradable: false, IsResearch: true},
		4: {ID: 4, DbLetter: 0, Name: "NoDbLetter", IsExchangeTradable: true},
	}
	svc, _ := newTestSvc(resources)

	resp, err := svc.ListResources(ctx)
	if err != nil {
		t.Fatalf("ListResources failed: %v", err)
	}
	if resp.Resources == nil {
		t.Fatal("expected non-nil resources array")
	}
	if len(*resp.Resources) != 2 {
		t.Fatalf("expected 2 resources (filtered), got %d", len(*resp.Resources))
	}

	r := (*resp.Resources)[0]
	if r.ResourceId == nil || *r.ResourceId != 1 {
		t.Errorf("expected resource_id 1, got %v", r.ResourceId)
	}
	if r.Name == nil || *r.Name != "Grain" {
		t.Errorf("expected name 'Grain', got %v", r.Name)
	}
}

func TestListResources_Empty(t *testing.T) {
	ctx := context.Background()
	svc, _ := newTestSvc(nil)

	resp, err := svc.ListResources(ctx)
	if err != nil {
		t.Fatalf("ListResources failed: %v", err)
	}
	if resp.Resources == nil {
		t.Fatal("expected non-nil resources array (should be empty)")
	}
	if len(*resp.Resources) != 0 {
		t.Errorf("expected 0 resources, got %d", len(*resp.Resources))
	}
}

func TestGetMarketDepth_SortsAndTakesTop5(t *testing.T) {
	ctx := context.Background()
	svc, store := newTestSvc(nil)

	for _, b := range []struct {
		id    string
		price float64
		qty   int
	}{
		{"buy1", 100.0, 10}, {"buy2", 99.0, 20}, {"buy3", 101.0, 5},
		{"buy4", 98.0, 30}, {"buy5", 97.0, 15}, {"buy6", 96.0, 25},
	} {
		_ = store.CreateOrder(ctx, &domainmarket.MarketOrder{
			ID: b.id, ResourceID: 10, IsBuy: true, Price: b.price,
			Quantity: b.qty, Quality: 0, Status: domainmarket.StatusOpen,
		})
	}
	for _, s := range []struct {
		id    string
		price float64
		qty   int
	}{
		{"sell1", 105.0, 10}, {"sell2", 106.0, 20}, {"sell3", 104.0, 5},
		{"sell4", 107.0, 30}, {"sell5", 108.0, 15}, {"sell6", 109.0, 25},
	} {
		_ = store.CreateOrder(ctx, &domainmarket.MarketOrder{
			ID: s.id, ResourceID: 10, IsBuy: false, Price: s.price,
			Quantity: s.qty, Quality: 0, Status: domainmarket.StatusOpen,
		})
	}
	_ = store.CreateOrder(ctx, &domainmarket.MarketOrder{
		ID: "q1-sell", ResourceID: 10, IsBuy: false, Price: 110.0,
		Quantity: 5, Quality: 1, Status: domainmarket.StatusOpen,
	})

	resp, err := svc.GetMarketDepth(ctx, 10, 0)
	if err != nil {
		t.Fatalf("GetMarketDepth failed: %v", err)
	}
	if resp.Buys == nil || len(*resp.Buys) != 5 {
		t.Fatalf("expected 5 buys, got %d", len(*resp.Buys))
	}
	if resp.Sells == nil || len(*resp.Sells) != 5 {
		t.Fatalf("expected 5 sells, got %d", len(*resp.Sells))
	}

	expectedBuyPrices := []float64{101, 100, 99, 98, 97}
	for i, level := range *resp.Buys {
		if level.Price == nil || float64(*level.Price) != expectedBuyPrices[i] {
			t.Errorf("buy[%d] price: expected %.2f, got %.2f", i, expectedBuyPrices[i], *level.Price)
		}
		if level.Qty == nil || *level.Qty != *level.Quantity {
			t.Errorf("buy[%d]: qty and quantity mismatch", i)
		}
	}

	expectedSellPrices := []float64{104, 105, 106, 107, 108}
	for i, level := range *resp.Sells {
		if level.Price == nil || float64(*level.Price) != expectedSellPrices[i] {
			t.Errorf("sell[%d] price: expected %.2f, got %.2f", i, expectedSellPrices[i], *level.Price)
		}
	}
}

func TestGetMarketDepth_AggregatesOpenAndPartialOrders(t *testing.T) {
	ctx := context.Background()
	svc, store := newTestSvc(nil)

	orders := []*domainmarket.MarketOrder{
		{ID: "buy-open-1", ResourceID: 12, IsBuy: true, Price: 31.5, Quantity: 10, Quality: 0, Status: domainmarket.StatusOpen},
		{ID: "buy-open-2", ResourceID: 12, IsBuy: true, Price: 31.5, Quantity: 8, Quality: 0, Status: domainmarket.StatusOpen},
		{ID: "buy-partial", ResourceID: 12, IsBuy: true, Price: 30.5, Quantity: 10, FilledQuantity: 4, Quality: 0, Status: domainmarket.StatusPartial},
		{ID: "buy-filled", ResourceID: 12, IsBuy: true, Price: 35.5, Quantity: 10, FilledQuantity: 10, Quality: 0, Status: domainmarket.StatusFilled},
		{ID: "sell-open", ResourceID: 12, IsBuy: false, Price: 33.5, Quantity: 6, Quality: 0, Status: domainmarket.StatusOpen},
	}
	for _, o := range orders {
		if err := store.CreateOrder(ctx, o); err != nil {
			t.Fatalf("CreateOrder: %v", err)
		}
	}

	resp, err := svc.GetMarketDepth(ctx, 12, 0)
	if err != nil {
		t.Fatalf("GetMarketDepth failed: %v", err)
	}
	if resp.Buys == nil || len(*resp.Buys) != 2 {
		t.Fatalf("expected 2 buy price levels, got %v", resp.Buys)
	}
	first := (*resp.Buys)[0]
	if first.Price == nil || float64(*first.Price) != 31.5 {
		t.Fatalf("expected best buy price 31.5, got %v", first.Price)
	}
	if first.Quantity == nil || *first.Quantity != 18 {
		t.Fatalf("expected aggregated quantity 18, got %v", first.Quantity)
	}
	second := (*resp.Buys)[1]
	if second.Quantity == nil || *second.Quantity != 6 {
		t.Fatalf("expected partial remaining quantity 6, got %v", second.Quantity)
	}
	if resp.Sells == nil || len(*resp.Sells) != 1 {
		t.Fatalf("expected 1 sell price level, got %v", resp.Sells)
	}
}

func TestGetMarketDepth_Empty(t *testing.T) {
	ctx := context.Background()
	svc, _ := newTestSvc(nil)

	resp, err := svc.GetMarketDepth(ctx, 99, 0)
	if err != nil {
		t.Fatalf("GetMarketDepth failed: %v", err)
	}
	if resp.Buys == nil || len(*resp.Buys) != 0 {
		t.Errorf("expected empty buys")
	}
	if resp.Sells == nil || len(*resp.Sells) != 0 {
		t.Errorf("expected empty sells")
	}
}

func TestGetMarketTicker_FallbackSeries(t *testing.T) {
	ctx := context.Background()
	svc, _ := newTestSvc(map[int]*catalog.ResourceEntry{
		12: {ID: 12, DbLetter: 12, Name: "Bread", IsExchangeTradable: true, BasePrice: 42.5},
	})

	resp, err := svc.GetMarketTicker(ctx, 12)
	if err != nil {
		t.Fatalf("GetMarketTicker failed: %v", err)
	}
	if resp.Resource == nil || *resp.Resource != 12 {
		t.Fatalf("expected resource 12, got %v", resp.Resource)
	}
	if resp.Series == nil || len(*resp.Series) != 48 {
		t.Fatalf("expected 48 ticker points, got %v", resp.Series)
	}
	expectedFirstTime := time.Date(2026, 6, 4, 13, 0, 0, 0, time.UTC)
	if (*resp.Series)[0].Time == nil || !(*resp.Series)[0].Time.Equal(expectedFirstTime) {
		t.Fatalf("expected first ticker time %s, got %v", expectedFirstTime, (*resp.Series)[0].Time)
	}
	expectedLastTime := time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC)
	if (*resp.Series)[47].Time == nil || !(*resp.Series)[47].Time.Equal(expectedLastTime) {
		t.Fatalf("expected last ticker time %s, got %v", expectedLastTime, (*resp.Series)[47].Time)
	}
	for i, point := range *resp.Series {
		if point.Price == nil || *point.Price <= 0 {
			t.Fatalf("point %d has invalid price %v", i, point.Price)
		}
		if point.Time == nil {
			t.Fatalf("point %d has nil time", i)
		}
	}
}

func TestListMarketOrders_FiltersByQuality(t *testing.T) {
	ctx := context.Background()
	svc, store := newTestSvc(nil)

	orders := []*domainmarket.MarketOrder{
		{ID: "o1", ResourceID: 20, IsBuy: true, Price: 50.0, Quantity: 10, Quality: 0, Status: domainmarket.StatusOpen, CompanyID: 1},
		{ID: "o2", ResourceID: 20, IsBuy: false, Price: 55.0, Quantity: 5, Quality: 0, Status: domainmarket.StatusOpen, CompanyID: 2},
		{ID: "o3", ResourceID: 20, IsBuy: true, Price: 52.0, Quantity: 8, Quality: 1, Status: domainmarket.StatusOpen, CompanyID: 1},
	}
	for _, o := range orders {
		if err := store.CreateOrder(ctx, o); err != nil {
			t.Fatalf("CreateOrder: %v", err)
		}
	}

	resp, err := svc.ListMarketOrders(ctx, 20, 0)
	if err != nil {
		t.Fatalf("ListMarketOrders failed: %v", err)
	}
	if resp.Orders == nil {
		t.Fatal("expected non-nil orders array")
	}
	if len(*resp.Orders) != 2 {
		t.Fatalf("expected 2 orders for quality 0, got %d", len(*resp.Orders))
	}

	// Check each returned order (map iteration is non-deterministic).
	ids := make(map[string]bool)
	for _, dto := range *resp.Orders {
		ids[*dto.Id] = true
		switch *dto.Id {
		case "o1":
			if dto.Kind == nil || *dto.Kind != 1 {
				t.Errorf("order o1 expected kind 1 (buy), got %v", dto.Kind)
			}
			if dto.Price == nil || float64(*dto.Price) != 50.0 {
				t.Errorf("order o1 expected price 50.0, got %v", dto.Price)
			}
			if dto.Remaining == nil || *dto.Remaining != 10 {
				t.Errorf("order o1 expected remaining 10, got %v", dto.Remaining)
			}
			if dto.CompanyId == nil || *dto.CompanyId != 1 {
				t.Errorf("order o1 expected companyId 1, got %v", dto.CompanyId)
			}
		case "o2":
			if dto.Kind == nil || *dto.Kind != 0 {
				t.Errorf("order o2 expected kind 0 (sell), got %v", dto.Kind)
			}
			if dto.Remaining == nil || *dto.Remaining != 5 {
				t.Errorf("order o2 expected remaining 5, got %v", dto.Remaining)
			}
		default:
			t.Errorf("unexpected order id: %s", *dto.Id)
		}
	}
	if !ids["o1"] || !ids["o2"] {
		t.Errorf("expected orders o1 and o2, got ids: %v", ids)
	}
}

func TestListMarketOrders_Empty(t *testing.T) {
	ctx := context.Background()
	svc, _ := newTestSvc(nil)

	resp, err := svc.ListMarketOrders(ctx, 99, 0)
	if err != nil {
		t.Fatalf("ListMarketOrders failed: %v", err)
	}
	if resp.Orders == nil {
		t.Fatal("expected non-nil orders array (should be empty)")
	}
	if len(*resp.Orders) != 0 {
		t.Errorf("expected 0 orders, got %d", len(*resp.Orders))
	}
}
