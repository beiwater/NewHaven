package service

import (
	"testing"

	"go-sim-api/internal/model"
)

func TestDailyOrdersGenerated(t *testing.T) {
	s := NewTestService()
	s.mu.Lock()
	s.State.DailyOrders = nil
	s.State.DailyOrdersDate = ""
	s.mu.Unlock()

	resp := s.DailyOrders()
	date := resp["date"].(string)
	orders := resp["orders"].([]model.Order)
	if len(orders) == 0 {
		t.Fatal("expected at least 1 daily order")
	}
	if date == "" {
		t.Fatal("expected a date string")
	}
}

func TestDailyOrdersPersistAcrossCalls(t *testing.T) {
	s := NewTestService()
	first := s.DailyOrders()
	firstDate := first["date"]

	second := s.DailyOrders()
	secondDate := second["date"]

	if firstDate != secondDate {
		t.Errorf("expected same date across calls, got %v vs %v", firstDate, secondDate)
	}
}

func TestCompleteDailyOrderInsufficientResources(t *testing.T) {
	s := NewTestService()
	s.mu.Lock()
	s.State.Companies[0].Inventory = map[int]int{}
	s.State.Companies[0].QualityInventory = map[string]int{}
	s.mu.Unlock()

	resp := s.DailyOrders()
	orders := resp["orders"].([]model.Order)
	if len(orders) == 0 {
		t.Fatal("no daily orders")
	}
	orderID := orders[0].ID

	_, err := s.CompleteDailyOrder(s.State.Companies[0].ID, orderID)
	if err == nil {
		t.Fatal("expected error when completing order without resources")
	}
}

func TestCompleteAndClaimDailyOrder(t *testing.T) {
	s := NewTestService()

	resp := s.DailyOrders()
	orders := resp["orders"].([]model.Order)
	if len(orders) == 0 {
		t.Fatal("no daily orders")
	}
	order := orders[0]

	s.mu.Lock()
	s.inventoryAdd(&s.State.Companies[0], order.ResourceID, order.Quality, order.Quantity+100)
	s.mu.Unlock()

	completed, err := s.CompleteDailyOrder(s.State.Companies[0].ID, order.ID)
	if err != nil {
		t.Fatalf("unexpected error completing order: %v", err)
	}
	if completed["status"] != "completed" {
		t.Errorf("expected status completed, got %v", completed["status"])
	}

	claim, err := s.ClaimDailyOrderReward(s.State.Companies[0].ID, order.ID)
	if err != nil {
		t.Fatalf("unexpected error claiming reward: %v", err)
	}

	cashVal := claim["cash"]
	cash, ok := cashVal.(float64)
	if !ok {
		var cashInt int
		cashInt, ok = cashVal.(int)
		if ok {
			cash = float64(cashInt)
		}
	}
	if !ok || cash <= 0 {
		t.Errorf("expected positive cash reward, got %v (type %T)", cashVal, cashVal)
	}
	// XP may be int or float64 depending on JSON round-trip
	xpVal := claim["xp"]
	xpNum, ok := xpVal.(float64)
	if !ok {
		var xpInt int
		xpInt, ok = xpVal.(int)
		if ok {
			xpNum = float64(xpInt)
		}
	}
	if !ok || xpNum <= 0 {
		t.Errorf("expected positive XP reward, got %v (type %T)", xpVal, xpVal)
	}
}

func TestClaimAlreadyClaimedOrderFails(t *testing.T) {
	s := NewTestService()

	resp := s.DailyOrders()
	orders := resp["orders"].([]model.Order)
	if len(orders) == 0 {
		t.Fatal("no daily orders")
	}
	order := orders[0]

	s.mu.Lock()
	s.inventoryAdd(&s.State.Companies[0], order.ResourceID, order.Quality, order.Quantity+100)
	s.mu.Unlock()

	if _, err := s.CompleteDailyOrder(s.State.Companies[0].ID, order.ID); err != nil {
		t.Fatalf("complete: %v", err)
	}
	if _, err := s.ClaimDailyOrderReward(s.State.Companies[0].ID, order.ID); err != nil {
		t.Fatalf("claim: %v", err)
	}
	if _, err := s.ClaimDailyOrderReward(s.State.Companies[0].ID, order.ID); err == nil {
		t.Fatal("expected error on second claim")
	}
}

func TestCompleteExpiredOrderFails(t *testing.T) {
	s := NewTestService()

	// Get fresh orders then expire them
	resp := s.DailyOrders()
	orders := resp["orders"].([]model.Order)
	if len(orders) == 0 {
		t.Fatal("no daily orders")
	}
	orderID := orders[0].ID

	s.mu.Lock()
	for i := range s.State.DailyOrders {
		s.State.DailyOrders[i].Status = "expired"
	}
	s.mu.Unlock()

	_, err := s.CompleteDailyOrder(s.State.Companies[0].ID, orderID)
	if err == nil {
		t.Fatal("expected error completing expired order")
	}
}

func TestClaimUncompletedOrderFails(t *testing.T) {
	s := NewTestService()

	resp := s.DailyOrders()
	orders := resp["orders"].([]model.Order)
	if len(orders) == 0 {
		t.Fatal("no daily orders")
	}
	orderID := orders[0].ID

	_, err := s.ClaimDailyOrderReward(s.State.Companies[0].ID, orderID)
	if err == nil {
		t.Fatal("expected error claiming uncompleted order")
	}
}

func TestDailyOrderNotFoundFails(t *testing.T) {
	s := NewTestService()

	_, err := s.CompleteDailyOrder(s.State.Companies[0].ID, "nonexistent-order")
	if err == nil {
		t.Fatal("expected error for nonexistent order")
	}
	_, err = s.ClaimDailyOrderReward(s.State.Companies[0].ID, "nonexistent-order")
	if err == nil {
		t.Fatal("expected error for nonexistent order")
	}
}
