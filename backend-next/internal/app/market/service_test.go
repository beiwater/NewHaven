package market_test

import (
	"context"
	"errors"
	"github.com/newhaven/backend-next/internal/config"
	"strings"
	"testing"
	"time"

	appmarket "github.com/newhaven/backend-next/internal/app/market"
	"github.com/newhaven/backend-next/internal/catalog"
	"github.com/newhaven/backend-next/internal/domain/auth"
	"github.com/newhaven/backend-next/internal/domain/company"
	domainmarket "github.com/newhaven/backend-next/internal/domain/market"
	openapi "github.com/newhaven/backend-next/internal/generated/openapi"
	"github.com/newhaven/backend-next/internal/platform"
	"github.com/newhaven/backend-next/internal/storage/memory"
)

func newTestSvc(resources map[int]*catalog.ResourceEntry) (*appmarket.Service, *memory.Store) {
	store := memory.New()
	if resources == nil {
		resources = make(map[int]*catalog.ResourceEntry)
	}
	clock := platform.NewFakeClock(time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC))
	idgen := platform.NewIDGen()
	cfg := &config.GameConfig{ExchangeFeePct: 0.04}
	svc := appmarket.NewService(store, store, store, resources, cfg, clock, idgen)
	return svc, store
}

type createOrderFailStore struct {
	*memory.Store
}

func (s *createOrderFailStore) CreateOrder(context.Context, *domainmarket.MarketOrder) error {
	return errors.New("create order failed")
}

func newTestCompany(t *testing.T, store *memory.Store, playerID int, username string, money float64) int {
	t.Helper()
	err := store.CreatePlayer(nil, &auth.Player{ID: playerID, Username: username, PasswordHash: "hash"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	err = store.CreateCompany(nil, &company.Company{
		PlayerID:  playerID,
		Name:      username + " Corp",
		Money:     money,
		Level:     1,
		XP:        0,
		Inventory: make(map[int]int),
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}
	c, err := store.GetCompanyByPlayerID(nil, playerID)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}
	return c.ID
}

// --- ListResources tests ---

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
	if resp.Resources == nil || len(*resp.Resources) != 2 {
		t.Fatalf("expected 2 resources, got %d", len(*resp.Resources))
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
	if resp.Resources == nil || len(*resp.Resources) != 0 {
		t.Fatal("expected empty resources")
	}
}

// --- MarketDepth tests ---

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
		_ = store.CreateOrder(ctx, o)
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

// --- Ticker tests ---

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

// --- ListMarketOrders tests ---

func TestListMarketOrders_FiltersByQuality(t *testing.T) {
	ctx := context.Background()
	svc, store := newTestSvc(nil)
	orders := []*domainmarket.MarketOrder{
		{ID: "o1", ResourceID: 20, IsBuy: true, Price: 50.0, Quantity: 10, Quality: 0, Status: domainmarket.StatusOpen, CompanyID: 1},
		{ID: "o2", ResourceID: 20, IsBuy: false, Price: 55.0, Quantity: 5, Quality: 0, Status: domainmarket.StatusOpen, CompanyID: 2},
		{ID: "o3", ResourceID: 20, IsBuy: true, Price: 52.0, Quantity: 8, Quality: 1, Status: domainmarket.StatusOpen, CompanyID: 1},
	}
	for _, o := range orders {
		_ = store.CreateOrder(ctx, o)
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

// --- CreateOrder tests ---

func TestCreateOrder_Buy_ReservesCashCreatesOpenOrderAndLedger(t *testing.T) {
	ctx := context.Background()
	resources := map[int]*catalog.ResourceEntry{
		5: {ID: 5, DbLetter: 5, Name: "Butter", IsExchangeTradable: true, BasePrice: 30},
	}
	svc, store := newTestSvc(resources)
	cid := newTestCompany(t, store, 100, "buyer", 10000)

	resp, err := svc.CreateOrder(ctx, cid, &openapi.CreateOrderRequestFrontend{
		ResourceId: 5, Kind: 1, Quality: 0, Quantity: 10, Price: 25.0,
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}
	if resp.Order == nil {
		t.Fatal("expected non-nil order in response")
	}
	if *resp.Order.Kind != 1 {
		t.Errorf("expected kind 1 (buy), got %d", *resp.Order.Kind)
	}
	if *resp.Order.Status != "open" {
		t.Errorf("expected status open, got %s", *resp.Order.Status)
	}
	if *resp.Order.Remaining != 10 {
		t.Errorf("expected remaining 10, got %d", *resp.Order.Remaining)
	}

	// Cash deducted.
	company, _ := store.GetCompany(nil, cid)
	if company.Money != 10000-250 {
		t.Errorf("expected money 9750, got %.2f", company.Money)
	}

	// Ledger entry present.
	entries, _ := store.GetLedgerEntries(nil, cid, 10)
	found := false
	for _, e := range entries {
		if e.Kind == "market_buy_reserve" && e.Amount == 250.0 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected market_buy_reserve ledger entry for 250")
	}
}

func TestCreateOrder_Sell_ReservesInventoryCreatesOpenOrder(t *testing.T) {
	ctx := context.Background()
	resources := map[int]*catalog.ResourceEntry{
		5: {ID: 5, DbLetter: 5, Name: "Butter", IsExchangeTradable: true, BasePrice: 30},
	}
	svc, store := newTestSvc(resources)
	cid := newTestCompany(t, store, 101, "seller", 5000)
	// Seed inventory.
	err := store.UpdateInventory(nil, cid, 5, 20)
	if err != nil {
		t.Fatalf("UpdateInventory: %v", err)
	}

	resp, err := svc.CreateOrder(ctx, cid, &openapi.CreateOrderRequestFrontend{
		ResourceId: 5, Kind: 0, Quality: 0, Quantity: 8, Price: 20.0,
	})
	if err != nil {
		t.Fatalf("CreateOrder failed: %v", err)
	}
	if resp.Order == nil {
		t.Fatal("expected non-nil order")
	}
	if *resp.Order.Kind != 0 {
		t.Errorf("expected kind 0 (sell), got %d", *resp.Order.Kind)
	}
	if *resp.Order.Status != "open" {
		t.Errorf("expected status open, got %s", *resp.Order.Status)
	}

	// Inventory deducted.
	c, _ := store.GetCompany(nil, cid)
	if c.Inventory[5] != 12 {
		t.Errorf("expected 12 of resource 5 remaining, got %d", c.Inventory[5])
	}

	// No ledger entry for sell.
	entries, _ := store.GetLedgerEntries(nil, cid, 10)
	for _, e := range entries {
		if e.Kind == "market_buy_reserve" {
			t.Error("sell orders should not produce reserve ledger entries")
		}
	}
}

func TestCreateOrder_InvalidPayloads(t *testing.T) {
	ctx := context.Background()
	resources := map[int]*catalog.ResourceEntry{
		5: {ID: 5, DbLetter: 5, Name: "Butter", IsExchangeTradable: true, BasePrice: 30},
	}
	svc, store := newTestSvc(resources)
	cid := newTestCompany(t, store, 102, "invalid", 5000)

	tests := []struct {
		name string
		req  openapi.CreateOrderRequestFrontend
	}{
		{"missing resource", openapi.CreateOrderRequestFrontend{ResourceId: 999, Kind: 1, Quality: 0, Quantity: 1, Price: 10}},
		{"bad kind", openapi.CreateOrderRequestFrontend{ResourceId: 5, Kind: 2, Quality: 0, Quantity: 1, Price: 10}},
		{"zero quantity", openapi.CreateOrderRequestFrontend{ResourceId: 5, Kind: 1, Quality: 0, Quantity: 0, Price: 10}},
		{"zero price", openapi.CreateOrderRequestFrontend{ResourceId: 5, Kind: 1, Quality: 0, Quantity: 1, Price: 0}},
		{"non-zero quality", openapi.CreateOrderRequestFrontend{ResourceId: 5, Kind: 1, Quality: 1, Quantity: 1, Price: 10}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.CreateOrder(ctx, cid, &tt.req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestCreateOrder_InsufficientFunds_NoOrder(t *testing.T) {
	ctx := context.Background()
	resources := map[int]*catalog.ResourceEntry{
		5: {ID: 5, DbLetter: 5, Name: "Butter", IsExchangeTradable: true, BasePrice: 30},
	}
	svc, store := newTestSvc(resources)
	cid := newTestCompany(t, store, 103, "poor", 100)

	_, err := svc.CreateOrder(ctx, cid, &openapi.CreateOrderRequestFrontend{
		ResourceId: 5, Kind: 1, Quality: 0, Quantity: 20, Price: 25.0, // total=500 > 100
	})
	if err == nil {
		t.Fatal("expected insufficient funds error")
	}
	if !strings.Contains(err.Error(), "insufficient funds") {
		t.Errorf("expected 'insufficient funds', got: %v", err)
	}
	// Money unchanged.
	c, _ := store.GetCompany(nil, cid)
	if c.Money != 100 {
		t.Errorf("expected money unchanged 100, got %.2f", c.Money)
	}
	// No orders created.
	orders, _ := store.GetOrdersByCompany(nil, cid)
	if len(orders) != 0 {
		t.Error("expected no orders created")
	}
}

func TestCreateOrder_Buy_RollsBackCashWhenOrderCreateFails(t *testing.T) {
	ctx := context.Background()
	resources := map[int]*catalog.ResourceEntry{
		5: {ID: 5, DbLetter: 5, Name: "Butter", IsExchangeTradable: true, BasePrice: 30},
	}
	store := memory.New()
	failingStore := &createOrderFailStore{Store: store}
	clock := platform.NewFakeClock(time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC))
	cfg := &config.GameConfig{ExchangeFeePct: 0.04}
	svc := appmarket.NewService(failingStore, failingStore, failingStore, resources, cfg, clock, platform.NewIDGen())
	cid := newTestCompany(t, store, 110, "rollbackbuyer", 1000)

	_, err := svc.CreateOrder(ctx, cid, &openapi.CreateOrderRequestFrontend{
		ResourceId: 5, Kind: 1, Quality: 0, Quantity: 4, Price: 25.0,
	})
	if err == nil {
		t.Fatal("expected create order failure")
	}

	c, _ := store.GetCompany(nil, cid)
	if c.Money != 1000 {
		t.Errorf("expected money rolled back to 1000, got %.2f", c.Money)
	}
}

func TestCreateOrder_InsufficientInventory_NoOrder(t *testing.T) {
	ctx := context.Background()
	resources := map[int]*catalog.ResourceEntry{
		5: {ID: 5, DbLetter: 5, Name: "Butter", IsExchangeTradable: true, BasePrice: 30},
	}
	svc, store := newTestSvc(resources)
	cid := newTestCompany(t, store, 104, "poorinv", 5000)

	_, err := svc.CreateOrder(ctx, cid, &openapi.CreateOrderRequestFrontend{
		ResourceId: 5, Kind: 0, Quality: 0, Quantity: 5, Price: 10.0,
	})
	if err == nil {
		t.Fatal("expected insufficient inventory error")
	}
	if !strings.Contains(err.Error(), "insufficient inventory") {
		t.Errorf("expected 'insufficient inventory', got: %v", err)
	}
}

// --- CancelOrder tests ---

func TestCancelOrder_Buy_RefundsCashAndLedger(t *testing.T) {
	ctx := context.Background()
	resources := map[int]*catalog.ResourceEntry{
		5: {ID: 5, DbLetter: 5, Name: "Butter", IsExchangeTradable: true, BasePrice: 30},
	}
	svc, store := newTestSvc(resources)
	cid := newTestCompany(t, store, 105, "canbuy", 5000)

	cr, err := svc.CreateOrder(ctx, cid, &openapi.CreateOrderRequestFrontend{
		ResourceId: 5, Kind: 1, Quality: 0, Quantity: 10, Price: 25.0,
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	beforeMoney, _ := store.GetCompany(nil, cid)
	beforeAmt := beforeMoney.Money

	resp, err := svc.CancelOrder(ctx, cid, *cr.Order.Id)
	if err != nil {
		t.Fatalf("CancelOrder failed: %v", err)
	}
	if resp.Id == nil || *resp.Id != *cr.Order.Id {
		t.Errorf("expected id %s, got %v", *cr.Order.Id, resp.Id)
	}
	if resp.Status == nil || *resp.Status != "cancelled" {
		t.Errorf("expected status cancelled, got %v", resp.Status)
	}

	// Money refunded full remaining (10 * 25 = 250).
	afterMoney, _ := store.GetCompany(nil, cid)
	if afterMoney.Money != beforeAmt+250 {
		t.Errorf("expected money %v, got %.2f", beforeAmt+250, afterMoney.Money)
	}

	// Order marked cancelled.
	order, _ := store.GetOrder(nil, *cr.Order.Id)
	if order.Status != domainmarket.StatusCancelled {
		t.Errorf("expected cancelled, got %s", order.Status)
	}
	if order.FilledQuantity != 0 {
		t.Errorf("expected FilledQuantity to preserve actual fills, got %d", order.FilledQuantity)
	}

	// Ledger refund entry.
	entries, _ := store.GetLedgerEntries(nil, cid, 10)
	foundRefund := false
	for _, e := range entries {
		if e.Kind == "market_buy_refund" && e.Amount == 250.0 {
			foundRefund = true
			break
		}
	}
	if !foundRefund {
		t.Error("expected market_buy_refund ledger entry")
	}
}

func TestCancelOrder_Sell_ReturnsInventory(t *testing.T) {
	ctx := context.Background()
	resources := map[int]*catalog.ResourceEntry{
		5: {ID: 5, DbLetter: 5, Name: "Butter", IsExchangeTradable: true, BasePrice: 30},
	}
	svc, store := newTestSvc(resources)
	cid := newTestCompany(t, store, 106, "cansell", 100)
	_ = store.UpdateInventory(nil, cid, 5, 15)

	cr, err := svc.CreateOrder(ctx, cid, &openapi.CreateOrderRequestFrontend{
		ResourceId: 5, Kind: 0, Quality: 0, Quantity: 10, Price: 20.0,
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	_, err = svc.CancelOrder(ctx, cid, *cr.Order.Id)
	if err != nil {
		t.Fatalf("CancelOrder: %v", err)
	}

	// Inventory restored.
	c, _ := store.GetCompany(nil, cid)
	if c.Inventory[5] != 15 {
		t.Errorf("expected inventory 15 restored, got %d", c.Inventory[5])
	}
}

func TestCancelOrder_WrongCompany_Error(t *testing.T) {
	ctx := context.Background()
	resources := map[int]*catalog.ResourceEntry{
		5: {ID: 5, DbLetter: 5, Name: "Butter", IsExchangeTradable: true, BasePrice: 30},
	}
	svc, store := newTestSvc(resources)
	cidA := newTestCompany(t, store, 107, "owner", 1000)
	cidB := newTestCompany(t, store, 108, "intruder", 1000)

	cr, err := svc.CreateOrder(ctx, cidA, &openapi.CreateOrderRequestFrontend{
		ResourceId: 5, Kind: 1, Quality: 0, Quantity: 5, Price: 10.0,
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	_, err = svc.CancelOrder(ctx, cidB, *cr.Order.Id)
	if err == nil {
		t.Fatal("expected error for wrong company")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found', got: %v", err)
	}

	// Order still open.
	order, _ := store.GetOrder(nil, *cr.Order.Id)
	if order.Status != domainmarket.StatusOpen {
		t.Errorf("expected order still open, got %s", order.Status)
	}
}

func TestCancelOrder_AlreadyCancelled_Error(t *testing.T) {
	ctx := context.Background()
	resources := map[int]*catalog.ResourceEntry{
		5: {ID: 5, DbLetter: 5, Name: "Butter", IsExchangeTradable: true, BasePrice: 30},
	}
	svc, store := newTestSvc(resources)
	cid := newTestCompany(t, store, 109, "dblcan", 1000)

	cr, err := svc.CreateOrder(ctx, cid, &openapi.CreateOrderRequestFrontend{
		ResourceId: 5, Kind: 1, Quality: 0, Quantity: 5, Price: 10.0,
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	_, err = svc.CancelOrder(ctx, cid, *cr.Order.Id)
	if err != nil {
		t.Fatalf("First cancel: %v", err)
	}

	_, err = svc.CancelOrder(ctx, cid, *cr.Order.Id)
	if err == nil {
		t.Fatal("expected error for second cancel")
	}
	if !strings.Contains(err.Error(), "already settled") {
		t.Errorf("expected 'already settled', got: %v", err)
	}
}

// --- TakeOrder tests ---

func TestTakeOrder_BuysFromBestSellOrders(t *testing.T) {
	ctx := context.Background()
	resources := map[int]*catalog.ResourceEntry{
		5: {ID: 5, DbLetter: 5, Name: "Butter", IsExchangeTradable: true, BasePrice: 30},
	}
	svc, store := newTestSvc(resources)
	cid := newTestCompany(t, store, 200, "taker", 100000)
	sellerCid := newTestCompany(t, store, 201, "seller", 1000)
	_ = store.UpdateInventory(nil, sellerCid, 5, 100)

	// Create sell orders at different prices.
	order10 := &domainmarket.MarketOrder{ID: "sell-10", CompanyID: sellerCid, ResourceID: 5, IsBuy: false, Price: 10.0, Quantity: 20, Quality: 0, Status: domainmarket.StatusOpen, CreatedAt: "2026-06-06T11:00:00Z"}
	order12 := &domainmarket.MarketOrder{ID: "sell-12", CompanyID: sellerCid, ResourceID: 5, IsBuy: false, Price: 12.0, Quantity: 10, Quality: 0, Status: domainmarket.StatusOpen, CreatedAt: "2026-06-06T11:00:01Z"}
	order11 := &domainmarket.MarketOrder{ID: "sell-11", CompanyID: sellerCid, ResourceID: 5, IsBuy: false, Price: 11.0, Quantity: 15, Quality: 0, Status: domainmarket.StatusOpen, CreatedAt: "2026-06-06T11:00:02Z"}
	_ = store.CreateOrder(ctx, order10)
	_ = store.CreateOrder(ctx, order12)
	_ = store.CreateOrder(ctx, order11)
	// Buy order that should NOT match.
	_ = store.CreateOrder(ctx, &domainmarket.MarketOrder{ID: "buy-20", CompanyID: sellerCid, ResourceID: 5, IsBuy: true, Price: 20.0, Quantity: 5, Quality: 0, Status: domainmarket.StatusOpen})

	resp, err := svc.TakeOrder(ctx, cid, &openapi.TakeOrderRequest{
		Resource: 5, Quantity: 30, Quality: 0, MaxPrice: 100.0,
	})
	if err != nil {
		t.Fatalf("TakeOrder failed: %v", err)
	}
	if resp.AmountBought == nil || *resp.AmountBought != 30 {
		t.Fatalf("expected amountBought 30, got %v", resp.AmountBought)
	}
	// sell-10(20@10) + sell-11(10@11) = 30; sell-12(10@12) not needed.
	if resp.Trades == nil || len(*resp.Trades) != 2 {
		t.Fatalf("expected 2 trades, got %d", len(*resp.Trades))
	}
	firstTrade := (*resp.Trades)[0]
	if float64(*firstTrade.Price) != 10.0 {
		t.Errorf("expected first trade price 10.0, got %v", firstTrade.Price)
	}
	// Verify order statuses.
	o10, _ := store.GetOrder(ctx, "sell-10")
	if o10.Status != domainmarket.StatusFilled {
		t.Errorf("expected sell-10 filled, got %s", o10.Status)
	}
	o12, _ := store.GetOrder(ctx, "sell-12")
	if o12.Status != domainmarket.StatusOpen {
		t.Errorf("expected sell-12 open (unfilled), got %s", o12.Status)
	}
	o11, _ := store.GetOrder(ctx, "sell-11")
	if o11.Status != domainmarket.StatusPartial {
		t.Errorf("expected sell-11 partial, got %s", o11.Status)
	}
	if o11.FilledQuantity != 10 {
		t.Errorf("expected sell-11 filled 10, got %d", o11.FilledQuantity)
	}
	// Taker inventory.
	c, _ := store.GetCompany(nil, cid)
	if c.Inventory[5] != 30 {
		t.Errorf("expected taker inventory 30, got %d", c.Inventory[5])
	}
}

func TestTakeOrder_PartialFillWhenAvailableSellsRunOut(t *testing.T) {
	ctx := context.Background()
	resources := map[int]*catalog.ResourceEntry{
		5: {ID: 5, DbLetter: 5, Name: "Butter", IsExchangeTradable: true, BasePrice: 30},
	}
	svc, store := newTestSvc(resources)
	cid := newTestCompany(t, store, 208, "partialtaker", 100000)
	sellerCid := newTestCompany(t, store, 209, "partialseller", 1000)

	_ = store.CreateOrder(ctx, &domainmarket.MarketOrder{
		ID: "only-sell", CompanyID: sellerCid, ResourceID: 5, IsBuy: false,
		Price: 7.0, Quantity: 6, Quality: 0, Status: domainmarket.StatusOpen,
		CreatedAt: "2026-06-06T11:00:00Z",
	})

	resp, err := svc.TakeOrder(ctx, cid, &openapi.TakeOrderRequest{
		Resource: 5, Quantity: 10, Quality: 0, MaxPrice: 100.0,
	})
	if err != nil {
		t.Fatalf("TakeOrder failed: %v", err)
	}
	if resp.AmountBought == nil || *resp.AmountBought != 6 {
		t.Fatalf("expected amountBought 6, got %v", resp.AmountBought)
	}
	if resp.Trades == nil || len(*resp.Trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(*resp.Trades))
	}

	order, _ := store.GetOrder(ctx, "only-sell")
	if order.Status != domainmarket.StatusFilled {
		t.Errorf("expected sell order filled, got %s", order.Status)
	}

	taker, _ := store.GetCompany(nil, cid)
	if taker.Inventory[5] != 6 {
		t.Errorf("expected taker inventory 6, got %d", taker.Inventory[5])
	}
}

func TestTakeOrder_NoQualifyingSells_ReturnsZero(t *testing.T) {
	ctx := context.Background()
	resources := map[int]*catalog.ResourceEntry{
		5: {ID: 5, DbLetter: 5, Name: "Butter", IsExchangeTradable: true, BasePrice: 30},
	}
	svc, store := newTestSvc(resources)
	cid := newTestCompany(t, store, 202, "noorder", 100000)

	resp, err := svc.TakeOrder(ctx, cid, &openapi.TakeOrderRequest{
		Resource: 5, Quantity: 10, Quality: 0, MaxPrice: 100.0,
	})
	if err != nil {
		t.Fatalf("TakeOrder failed: %v", err)
	}
	if resp.AmountBought == nil || *resp.AmountBought != 0 {
		t.Errorf("expected amountBought 0, got %v", resp.AmountBought)
	}
	if resp.Trades == nil || len(*resp.Trades) != 0 {
		t.Errorf("expected 0 trades, got %d", len(*resp.Trades))
	}
	if resp.MoneyDelta == nil || *resp.MoneyDelta != 0 {
		t.Errorf("expected moneyDelta 0, got %v", resp.MoneyDelta)
	}
}

func TestTakeOrder_StopsWhenCannotAffordFill(t *testing.T) {
	ctx := context.Background()
	resources := map[int]*catalog.ResourceEntry{
		5: {ID: 5, DbLetter: 5, Name: "Butter", IsExchangeTradable: true, BasePrice: 30},
	}
	svc, store := newTestSvc(resources)
	cid := newTestCompany(t, store, 203, "poortaker", 50)
	sellerCid := newTestCompany(t, store, 204, "richseller", 1000)
	_ = store.UpdateInventory(nil, sellerCid, 5, 100)

	_ = store.CreateOrder(ctx, &domainmarket.MarketOrder{ID: "s1", CompanyID: sellerCid, ResourceID: 5, IsBuy: false, Price: 2.0, Quantity: 10, Quality: 0, Status: domainmarket.StatusOpen, CreatedAt: "2026-06-06T11:00:00Z"})
	_ = store.CreateOrder(ctx, &domainmarket.MarketOrder{ID: "s2", CompanyID: sellerCid, ResourceID: 5, IsBuy: false, Price: 3.0, Quantity: 10, Quality: 0, Status: domainmarket.StatusOpen, CreatedAt: "2026-06-06T11:00:01Z"})

	resp, err := svc.TakeOrder(ctx, cid, &openapi.TakeOrderRequest{
		Resource: 5, Quantity: 20, Quality: 0, MaxPrice: 100.0,
	})
	if err != nil {
		t.Fatalf("TakeOrder failed: %v", err)
	}
	// First order at $2: fill=10, fee=ceil(10*2*0.04)=ceil(0.8)=1, cost=20+1=21
	// Remaining money: 50-21=29
	// Second order at $3: fill=10, fee=ceil(10*3*0.04)=ceil(1.2)=2, cost=30+2=32
	// 29 < 32, so should stop before second fill.
	if resp.AmountBought == nil || *resp.AmountBought != 10 {
		t.Fatalf("expected amountBought 10, got %v", resp.AmountBought)
	}
	if resp.Trades == nil || len(*resp.Trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(*resp.Trades))
	}
	// Second order untouched.
	o2, _ := store.GetOrder(ctx, "s2")
	if o2.FilledQuantity != 0 {
		t.Errorf("expected s2 unfilled, got filled=%d", o2.FilledQuantity)
	}
}

func TestTakeOrder_RecordsTradesAndUpdatesTicker(t *testing.T) {
	ctx := context.Background()
	resources := map[int]*catalog.ResourceEntry{
		5: {ID: 5, DbLetter: 5, Name: "Butter", IsExchangeTradable: true, BasePrice: 30},
	}
	svc, store := newTestSvc(resources)
	cid := newTestCompany(t, store, 205, "tradetest", 100000)
	sellerCid := newTestCompany(t, store, 206, "selltrades", 1000)
	_ = store.UpdateInventory(nil, sellerCid, 5, 100)

	_ = store.CreateOrder(ctx, &domainmarket.MarketOrder{ID: "ts1", CompanyID: sellerCid, ResourceID: 5, IsBuy: false, Price: 8.0, Quantity: 10, Quality: 0, Status: domainmarket.StatusOpen, CreatedAt: "2026-06-06T11:00:00Z"})

	resp, err := svc.TakeOrder(ctx, cid, &openapi.TakeOrderRequest{
		Resource: 5, Quantity: 10, Quality: 0, MaxPrice: 100.0,
	})
	if err != nil {
		t.Fatalf("TakeOrder failed: %v", err)
	}
	if resp.Trades == nil || len(*resp.Trades) != 1 {
		t.Fatalf("expected 1 trade")
	}
	trade := (*resp.Trades)[0]
	if float64(*trade.Price) != 8.0 {
		t.Errorf("expected trade price 8.0, got %v", trade.Price)
	}
	if *trade.Quantity != 10 {
		t.Errorf("expected trade qty 10, got %v", trade.Quantity)
	}

	// Ticker updated.
	ticker, err := store.GetTicker(ctx, 5)
	if err != nil {
		t.Fatalf("GetTicker: %v", err)
	}
	if ticker.LastPrice != 8.0 {
		t.Errorf("expected LastPrice 8.0, got %.2f", ticker.LastPrice)
	}
	if ticker.Volume24h <= 0 {
		t.Errorf("expected positive Volume24h, got %.2f", ticker.Volume24h)
	}

	// Ledger entry.
	entries, _ := store.GetLedgerEntries(nil, cid, 10)
	found := false
	for _, e := range entries {
		if e.Kind == "market_take_buy" && e.Amount > 0 {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected market_take_buy ledger entry")
	}

	// Seller money credited.
	seller, _ := store.GetCompany(nil, sellerCid)
	if seller.Money <= 1000 {
		t.Errorf("expected seller money > 1000, got %.2f", seller.Money)
	}
}

func TestTakeOrder_ValidatesBadPayloads(t *testing.T) {
	ctx := context.Background()
	resources := map[int]*catalog.ResourceEntry{
		5: {ID: 5, DbLetter: 5, Name: "Butter", IsExchangeTradable: true, BasePrice: 30},
	}
	svc, store := newTestSvc(resources)
	cid := newTestCompany(t, store, 207, "badpayload", 100000)

	tests := []struct {
		name string
		req  openapi.TakeOrderRequest
	}{
		{"missing resource", openapi.TakeOrderRequest{Resource: 999, Quantity: 1, Quality: 0, MaxPrice: 10}},
		{"zero quantity", openapi.TakeOrderRequest{Resource: 5, Quantity: 0, Quality: 0, MaxPrice: 10}},
		{"zero maxPrice", openapi.TakeOrderRequest{Resource: 5, Quantity: 1, Quality: 0, MaxPrice: 0}},
		{"non-zero quality", openapi.TakeOrderRequest{Resource: 5, Quantity: 1, Quality: 1, MaxPrice: 10}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := svc.TakeOrder(ctx, cid, &tt.req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

// --- Auto-match on CreateOrder tests ---

func TestCreateOrder_BuyAutoMatchesExistingSell(t *testing.T) {
	ctx := context.Background()
	resources := map[int]*catalog.ResourceEntry{
		5: {ID: 5, DbLetter: 5, Name: "Butter", IsExchangeTradable: true, BasePrice: 30},
	}
	svc, store := newTestSvc(resources)
	sellerCid := newTestCompany(t, store, 300, "seller", 1000)
	buyerCid := newTestCompany(t, store, 301, "buyer", 10000)
	_ = store.UpdateInventory(nil, sellerCid, 5, 100)

	// Pre-create a sell order at $8.
	_, err := svc.CreateOrder(ctx, sellerCid, &openapi.CreateOrderRequestFrontend{
		ResourceId: 5, Kind: 0, Quality: 0, Quantity: 20, Price: 8.0,
	})
	if err != nil {
		t.Fatalf("create sell: %v", err)
	}

	// Create a buy order at $10. It should match the sell at $8.
	resp, err := svc.CreateOrder(ctx, buyerCid, &openapi.CreateOrderRequestFrontend{
		ResourceId: 5, Kind: 1, Quality: 0, Quantity: 10, Price: 10.0,
	})
	if err != nil {
		t.Fatalf("create buy: %v", err)
	}

	// Buy order should be filled (10 units matched against sell's 20, remaining 10 on sell).
	if resp.Order == nil {
		t.Fatal("expected order in response")
	}
	if *resp.Order.Status != "filled" {
		t.Errorf("expected buy order status 'filled', got %s", *resp.Order.Status)
	}
	if *resp.Order.Remaining != 0 {
		t.Errorf("expected buy order remaining 0, got %d", *resp.Order.Remaining)
	}

	// Buyer got inventory: 10 units.
	buyer, _ := store.GetCompany(nil, buyerCid)
	if buyer.Inventory[5] != 10 {
		t.Errorf("expected buyer inventory 10, got %d", buyer.Inventory[5])
	}
	// Buyer paid: reserved 100, refund 20 (10 * (10-8)), net 80 spent.
	// Money was 10000, reserved 100, refunded 20, so 9920.
	if buyer.Money != 9920 {
		t.Errorf("expected buyer money 9920 (10000-100+20), got %.2f", buyer.Money)
	}

	// Seller got money: 10*8 - fee where fee=ceil(80*0.04)=ceil(3.2)=4, so 80-4=76.
	seller, _ := store.GetCompany(nil, sellerCid)
	if seller.Money != 1000+76 {
		t.Errorf("expected seller money 1076, got %.2f", seller.Money)
	}

	// Sell order partially filled (20->10 remaining).
	sellOrders, _ := store.GetOrdersByCompany(nil, sellerCid)
	if len(sellOrders) != 1 {
		t.Fatalf("expected 1 sell order, got %d", len(sellOrders))
	}
	sellOrder := sellOrders[0]
	if sellOrder.Status != domainmarket.StatusPartial {
		t.Errorf("expected sell order partial, got %s", sellOrder.Status)
	}
	if sellOrder.FilledQuantity != 10 {
		t.Errorf("expected sell order filled 10, got %d", sellOrder.FilledQuantity)
	}

	// Trade recorded.
	trades, _ := store.GetTrades(nil, 5, 10)
	if len(trades) == 0 {
		t.Error("expected at least one trade")
	}
	ticker, err := store.GetTicker(ctx, 5)
	if err != nil {
		t.Fatalf("expected ticker update: %v", err)
	}
	if ticker.LastPrice != 8.0 {
		t.Errorf("expected ticker last price 8.0, got %.2f", ticker.LastPrice)
	}

	// Ledger entries: seller has market_trade and market_fee.
	entries, _ := store.GetLedgerEntries(nil, sellerCid, 10)
	foundTrade := false
	foundFee := false
	for _, e := range entries {
		if e.Kind == "market_trade" && e.Amount == 80.0 {
			foundTrade = true
		}
		if e.Kind == "market_fee" && e.Amount == 4.0 {
			foundFee = true
		}
	}
	if !foundTrade {
		t.Error("expected market_trade ledger entry for seller")
	}
	if !foundFee {
		t.Error("expected market_fee ledger entry for seller")
	}

	// Buyer has market_buy_refund.
	buyerEntries, _ := store.GetLedgerEntries(nil, buyerCid, 10)
	foundRefund := false
	for _, e := range buyerEntries {
		if e.Kind == "market_buy_refund" && e.Amount == 20.0 {
			foundRefund = true
		}
	}
	if !foundRefund {
		t.Error("expected market_buy_refund ledger entry for buyer")
	}
}

func TestCreateOrder_SellAutoMatchesExistingBuy(t *testing.T) {
	ctx := context.Background()
	resources := map[int]*catalog.ResourceEntry{
		5: {ID: 5, DbLetter: 5, Name: "Butter", IsExchangeTradable: true, BasePrice: 30},
	}
	svc, store := newTestSvc(resources)
	buyerCid := newTestCompany(t, store, 302, "buyer", 10000)
	sellerCid := newTestCompany(t, store, 303, "seller", 1000)
	_ = store.UpdateInventory(nil, sellerCid, 5, 100)

	// Pre-create a buy order at $18 for 15 units.
	_, err := svc.CreateOrder(ctx, buyerCid, &openapi.CreateOrderRequestFrontend{
		ResourceId: 5, Kind: 1, Quality: 0, Quantity: 15, Price: 18.0,
	})
	if err != nil {
		t.Fatalf("create buy: %v", err)
	}

	// Reset buyer money tracking; we know the buy reserved 18*15=270 from 10000.
	// Now create a sell at $15 for 10 units.
	resp, err := svc.CreateOrder(ctx, sellerCid, &openapi.CreateOrderRequestFrontend{
		ResourceId: 5, Kind: 0, Quality: 0, Quantity: 10, Price: 15.0,
	})
	if err != nil {
		t.Fatalf("create sell: %v", err)
	}

	if resp.Order == nil {
		t.Fatal("expected order in response")
	}
	if *resp.Order.Status != "filled" {
		t.Errorf("expected sell order status 'filled', got %s", *resp.Order.Status)
	}

	// Buyer gets 10 units, refunded (18-15)*10=30.
	buyer, _ := store.GetCompany(nil, buyerCid)
	if buyer.Inventory[5] != 10 {
		t.Errorf("expected buyer inventory 10, got %d", buyer.Inventory[5])
	}
	// Buyer reserved 270, refunded 30, net spent 240.
	// But wait: buyer initially had 10000. Reserved 270 (9730). Refunded 30 (9760). So buyer.Money should be ~9760.
	if buyer.Money != 10000-270+30 {
		t.Errorf("expected buyer money 9760, got %.2f", buyer.Money)
	}

	// Seller gets paid: 10*15 - fee = 150 - ceil(150*0.04)=150-6=144.
	seller, _ := store.GetCompany(nil, sellerCid)
	if seller.Money != 1000+144 {
		t.Errorf("expected seller money 1144, got %.2f", seller.Money)
	}

	// Buy order partially filled (15->5 remaining).
	buyOrders, _ := store.GetOrdersByCompany(nil, buyerCid)
	if len(buyOrders) != 1 {
		t.Fatalf("expected 1 buy order, got %d", len(buyOrders))
	}
	buyOrder := buyOrders[0]
	if buyOrder.Status != domainmarket.StatusPartial {
		t.Errorf("expected buy order partial, got %s", buyOrder.Status)
	}
	if buyOrder.FilledQuantity != 10 {
		t.Errorf("expected buy order filled 10, got %d", buyOrder.FilledQuantity)
	}
}

func TestCreateOrder_PartialMatch(t *testing.T) {
	ctx := context.Background()
	resources := map[int]*catalog.ResourceEntry{
		5: {ID: 5, DbLetter: 5, Name: "Butter", IsExchangeTradable: true, BasePrice: 30},
	}
	svc, store := newTestSvc(resources)
	sellerCid := newTestCompany(t, store, 304, "seller", 1000)
	buyerCid := newTestCompany(t, store, 305, "buyer", 10000)
	_ = store.UpdateInventory(nil, sellerCid, 5, 100)

	// Pre-create a sell order for 30 units at $8.
	_, err := svc.CreateOrder(ctx, sellerCid, &openapi.CreateOrderRequestFrontend{
		ResourceId: 5, Kind: 0, Quality: 0, Quantity: 30, Price: 8.0,
	})
	if err != nil {
		t.Fatalf("create sell: %v", err)
	}

	// Create a buy order for 50 units at $10.
	resp, err := svc.CreateOrder(ctx, buyerCid, &openapi.CreateOrderRequestFrontend{
		ResourceId: 5, Kind: 1, Quality: 0, Quantity: 50, Price: 10.0,
	})
	if err != nil {
		t.Fatalf("create buy: %v", err)
	}

	if resp.Order == nil {
		t.Fatal("expected order in response")
	}
	// Buy order should be partial: 30 filled, 20 remaining.
	if *resp.Order.Status != "partial" {
		t.Errorf("expected buy order status 'partial', got %s", *resp.Order.Status)
	}
	if *resp.Order.Remaining != 20 {
		t.Errorf("expected buy order remaining 20, got %d", *resp.Order.Remaining)
	}

	// Buyer got 30 units.
	buyer, _ := store.GetCompany(nil, buyerCid)
	if buyer.Inventory[5] != 30 {
		t.Errorf("expected buyer inventory 30, got %d", buyer.Inventory[5])
	}

	// Sell order filled.
	sellOrders, _ := store.GetOrdersByCompany(nil, sellerCid)
	if len(sellOrders) != 1 {
		t.Fatalf("expected 1 sell order, got %d", len(sellOrders))
	}
	if sellOrders[0].Status != domainmarket.StatusFilled {
		t.Errorf("expected sell order filled, got %s", sellOrders[0].Status)
	}
}

func TestCreateOrder_NoMatchWhenPriceDoesNotCross(t *testing.T) {
	ctx := context.Background()
	resources := map[int]*catalog.ResourceEntry{
		5: {ID: 5, DbLetter: 5, Name: "Butter", IsExchangeTradable: true, BasePrice: 30},
	}
	svc, store := newTestSvc(resources)
	sellerCid := newTestCompany(t, store, 306, "expensiveseller", 1000)
	buyerCid := newTestCompany(t, store, 308, "cheapbuyer", 10000)
	_ = store.UpdateInventory(nil, sellerCid, 5, 100)

	_, err := svc.CreateOrder(ctx, sellerCid, &openapi.CreateOrderRequestFrontend{
		ResourceId: 5, Kind: 0, Quality: 0, Quantity: 10, Price: 8.0,
	})
	if err != nil {
		t.Fatalf("create sell: %v", err)
	}

	resp, err := svc.CreateOrder(ctx, buyerCid, &openapi.CreateOrderRequestFrontend{
		ResourceId: 5, Kind: 1, Quality: 0, Quantity: 10, Price: 5.0,
	})
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}

	// Order stays open, no inventory given.
	if resp.Order == nil || *resp.Order.Status != "open" {
		t.Errorf("expected status open, got %v", resp.Order.Status)
	}
	company, _ := store.GetCompany(nil, buyerCid)
	if company.Inventory == nil || company.Inventory[5] != 0 {
		t.Errorf("expected no inventory added, got %d", company.Inventory[5])
	}
	sellOrders, _ := store.GetOrdersByCompany(nil, sellerCid)
	if len(sellOrders) != 1 {
		t.Fatalf("expected one seller order, got %d", len(sellOrders))
	}
	if sellOrders[0].Status != domainmarket.StatusOpen || sellOrders[0].FilledQuantity != 0 {
		t.Errorf("expected sell order untouched, got status=%s filled=%d", sellOrders[0].Status, sellOrders[0].FilledQuantity)
	}
}

func TestCreateOrder_NoMatchSameCompany(t *testing.T) {
	ctx := context.Background()
	resources := map[int]*catalog.ResourceEntry{
		5: {ID: 5, DbLetter: 5, Name: "Butter", IsExchangeTradable: true, BasePrice: 30},
	}
	svc, store := newTestSvc(resources)
	cid := newTestCompany(t, store, 307, "sameco", 100000)
	_ = store.UpdateInventory(nil, cid, 5, 100)

	// Create a sell order first.
	_, err := svc.CreateOrder(ctx, cid, &openapi.CreateOrderRequestFrontend{
		ResourceId: 5, Kind: 0, Quality: 0, Quantity: 20, Price: 8.0,
	})
	if err != nil {
		t.Fatalf("create sell: %v", err)
	}

	// Now create a buy from the same company at a higher price.
	resp, err := svc.CreateOrder(ctx, cid, &openapi.CreateOrderRequestFrontend{
		ResourceId: 5, Kind: 1, Quality: 0, Quantity: 10, Price: 10.0,
	})
	if err != nil {
		t.Fatalf("create buy: %v", err)
	}

	// Should NOT match (same company). Both orders should remain open.
	if resp.Order == nil || *resp.Order.Status != "open" {
		t.Errorf("expected status open (no same-company match), got %v", resp.Order.Status)
	}
	orders, _ := store.GetOrdersByCompany(nil, cid)
	if len(orders) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(orders))
	}
	for _, o := range orders {
		if o.Status != domainmarket.StatusOpen {
			t.Errorf("expected order %s to remain open, got %s", o.ID, o.Status)
		}
	}
}
