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
