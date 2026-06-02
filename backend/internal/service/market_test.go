package service

import (
	"testing"

	"go-sim-api/internal/config"
	"go-sim-api/internal/data"
	"go-sim-api/internal/model"
)

func newTestService() *Service {
	cfg := config.Load()
	cfg.ACEnabled = false
	cfg.AMLEnabled = false
	cfg.ScriptDetectEnabled = false
	return New(&data.StaticData{}, cfg, nil)
}

func TestCreateQualitySellOrderUsesQualityInventory(t *testing.T) {
	s := newTestService()
	s.State.Companies[0].Inventory[8] = 0
	s.State.Companies[0].QualityInventory = map[string]int{"8_2": 10}

	if _, err := s.CreateOrder(s.State.Companies[0].ID, 8, 0, 2, 4, 10); err != nil {
		t.Fatalf("CreateOrder() unexpected error: %v", err)
	}
	if got := s.State.Companies[0].QualityInventory["8_2"]; got != 6 {
		t.Fatalf("quality inventory = %d, want 6", got)
	}
	if got := s.State.Companies[0].Inventory[8]; got != 0 {
		t.Fatalf("base inventory = %d, want 0", got)
	}
}

func TestCancelQualitySellOrderRefundsQualityInventory(t *testing.T) {
	s := newTestService()
	s.State.Companies[0].Inventory[8] = 0
	s.State.Companies[0].QualityInventory = map[string]int{"8_2": 10}

	resp, err := s.CreateOrder(s.State.Companies[0].ID, 8, 0, 2, 4, 10)
	if err != nil {
		t.Fatalf("CreateOrder() unexpected error: %v", err)
	}
	order := resp["order"].(model.MarketOrder)
	if _, err := s.CancelOrder(s.State.Companies[0].ID, order.ID); err != nil {
		t.Fatalf("CancelOrder() unexpected error: %v", err)
	}
	if got := s.State.Companies[0].QualityInventory["8_2"]; got != 10 {
		t.Fatalf("quality inventory after cancel = %d, want 10", got)
	}
}

func TestTakeQualitySellOrderAddsQualityInventory(t *testing.T) {
	s := newTestService()
	s.State.Companies[0].Inventory[8] = 0
	s.State.Orders = []model.MarketOrder{{
		ID: "sell-quality", ResourceID: 8, Kind: 0, Quality: 2,
		Quantity: 5, Remaining: 5, Price: 10, CompanyID: s.Cfg.Game.Bot1ID,
	}}

	if _, err := s.TakeOrder(s.State.Companies[0].ID, 8, 3, 2, 10); err != nil {
		t.Fatalf("TakeOrder() unexpected error: %v", err)
	}
	if got := s.State.Companies[0].QualityInventory["8_2"]; got != 3 {
		t.Fatalf("quality inventory = %d, want 3", got)
	}
	if got := s.State.Companies[0].Inventory[8]; got != 0 {
		t.Fatalf("base inventory = %d, want 0", got)
	}
}

func TestCreateOrderRejectsInvalidKind(t *testing.T) {
	s := newTestService()
	if _, err := s.CreateOrder(s.State.Companies[0].ID, 8, 7, 0, 1, 10); err == nil {
		t.Fatal("CreateOrder() expected invalid kind error")
	}
}

func TestCreateBuyOrderDeductsCash(t *testing.T) {
	s := newCoreTestService()
	initial := s.State.Companies[0].Money

	resp, err := s.CreateOrder(s.State.Companies[0].ID, 8, 1, 0, 5, 10.0)
	if err != nil {
		t.Fatalf("CreateOrder buy: %v", err)
	}
	order := resp["order"].(model.MarketOrder)
	if order.Kind != 1 {
		t.Errorf("expected buy order (kind=1), got %d", order.Kind)
	}
	if order.Price != 10.0 {
		t.Errorf("price = %f, want 10.0", order.Price)
	}
	// Cash should be deducted: 5 * 10 = 50
	expectedMoney := initial - 50.0
	if s.State.Companies[0].Money != expectedMoney {
		t.Errorf("money after buy = %f, want %f", s.State.Companies[0].Money, expectedMoney)
	}
}

func TestCreateSellOrderDeductsInventory(t *testing.T) {
	s := newCoreTestService()
	initial := s.State.Companies[0].Inventory[8]

	_, err := s.CreateOrder(s.State.Companies[0].ID, 8, 0, 0, 10, 12.0)
	if err != nil {
		t.Fatalf("CreateOrder sell: %v", err)
	}
	if got := s.State.Companies[0].Inventory[8]; got != initial-10 {
		t.Errorf("inventory after sell order = %d, want %d", got, initial-10)
	}
}

func TestCreateOrderInsufficientInventory(t *testing.T) {
	s := newCoreTestService()
	s.State.Companies[0].Inventory[8] = 5
	_, err := s.CreateOrder(s.State.Companies[0].ID, 8, 0, 0, 10, 12.0)
	if err == nil {
		t.Fatal("expected error for insufficient inventory")
	}
}

func TestCreateOrderInsufficientCash(t *testing.T) {
	s := newCoreTestService()
	s.State.Companies[0].Money = 10
	_, err := s.CreateOrder(s.State.Companies[0].ID, 8, 1, 0, 100, 100.0)
	if err == nil {
		t.Fatal("expected error for insufficient cash")
	}
}

func TestCancelBuyOrderRefundsCash(t *testing.T) {
	s := newCoreTestService()
	initial := s.State.Companies[0].Money

	resp, err := s.CreateOrder(s.State.Companies[0].ID, 8, 1, 0, 5, 10.0)
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	order := resp["order"].(model.MarketOrder)

	cancelResp, err := s.CancelOrder(s.State.Companies[0].ID, order.ID)
	if err != nil {
		t.Fatalf("CancelOrder: %v", err)
	}
	if cancelResp["status"] != "cancelled" {
		t.Errorf("expected cancelled, got %v", cancelResp["status"])
	}
	// Cash should be refunded
	if s.State.Companies[0].Money != initial {
		t.Errorf("money after cancel = %f, want %f", s.State.Companies[0].Money, initial)
	}
}

func TestCancelSellOrderRefundsInventory(t *testing.T) {
	s := newCoreTestService()
	initial := s.State.Companies[0].Inventory[8]

	resp, err := s.CreateOrder(s.State.Companies[0].ID, 8, 0, 0, 10, 12.0)
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	order := resp["order"].(model.MarketOrder)

	_, err = s.CancelOrder(s.State.Companies[0].ID, order.ID)
	if err != nil {
		t.Fatalf("CancelOrder: %v", err)
	}
	if got := s.State.Companies[0].Inventory[8]; got != initial {
		t.Errorf("inventory after cancel = %d, want %d", got, initial)
	}
}

func TestCancelNonexistentOrder(t *testing.T) {
	s := newCoreTestService()
	_, err := s.CancelOrder(s.State.Companies[0].ID, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent order")
	}
}

func TestMatchLimitOrders(t *testing.T) {
	s := newCoreTestService()
	// Bot has a sell order, create a matching buy order
	s.State.Orders = []model.MarketOrder{
		{ID: "sell-1", ResourceID: 8, Kind: 0, Price: 10.0, Quantity: 20, Remaining: 20, CompanyID: s.Cfg.Game.Bot1ID, Status: "open"},
	}
	initialInv := s.State.Companies[0].Inventory[8]

	_, err := s.CreateOrder(s.State.Companies[0].ID, 8, 1, 0, 10, 10.0)
	if err != nil {
		t.Fatalf("CreateOrder buy: %v", err)
	}
	// After match, company should have inventory + filled quantity
	if got := s.State.Companies[0].Inventory[8]; got != initialInv+10 {
		t.Errorf("inventory after match = %d, want %d", got, initialInv+10)
	}
}

func TestMatchLimitOrdersPricePriority(t *testing.T) {
	s := newCoreTestService()
	// Multiple sells at different prices
	s.State.Orders = []model.MarketOrder{
		{ID: "sell-high", ResourceID: 8, Kind: 0, Price: 12.0, Quantity: 10, Remaining: 10, CompanyID: s.Cfg.Game.Bot1ID, Status: "open"},
		{ID: "sell-low", ResourceID: 8, Kind: 0, Price: 9.0, Quantity: 10, Remaining: 10, CompanyID: s.Cfg.Game.Bot2ID, Status: "open"},
	}

	_, err := s.CreateOrder(s.State.Companies[0].ID, 8, 1, 0, 10, 15.0)
	if err != nil {
		t.Fatalf("CreateOrder: %v", err)
	}
	// Should match with lowest sell price (9.0) first
	for _, o := range s.State.Orders {
		if o.ID == "sell-low" && o.Remaining != 0 {
			t.Errorf("lowest sell should be fully filled, remaining=%d", o.Remaining)
		}
	}
}

