package service

import (
	"fmt"
	"math"
	"math/rand"
	"time"

	"go-sim-api/internal/model"
)

// DailyOrders returns the current daily orders. If the orders are stale (wrong date),
// it regenerates them automatically.
func (s *Service) DailyOrders() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureDailyOrdersLocked()
	return map[string]any{
		"orders": s.State.DailyOrders,
		"date":   s.State.DailyOrdersDate,
	}
}

func (s *Service) ClaimDailyOrderReward(companyID int, orderID string) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}
	s.ensureDailyOrdersLocked()
	for i := range s.State.DailyOrders {
		o := &s.State.DailyOrders[i]
		if o.ID != orderID {
			continue
		}
		if o.Status != "completed" {
			return nil, fmt.Errorf("order is not completed")
		}
		if o.RewardCash == 0 && o.RewardXP == 0 {
			return nil, fmt.Errorf("order already claimed")
		}
		cash := o.RewardCash
		xp := o.RewardXP
		company.Money += cash
		s.addXP(company, xp)
		s.addLedger("daily_order_reward", cash, "in", map[string]any{"orderId": o.ID})
		o.RewardCash = 0
		o.RewardXP = 0
		s.saveStateLocked()
		return map[string]any{
			"id": o.ID, "cash": cash, "xp": xp, "money": company.Money,
		}, nil
	}
	return nil, fmt.Errorf("order not found")
}

func (s *Service) CompleteDailyOrder(companyID int, orderID string) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}
	s.ensureDailyOrdersLocked()
	for i := range s.State.DailyOrders {
		o := &s.State.DailyOrders[i]
		if o.ID != orderID {
			continue
		}
		if o.Status == "completed" {
			return nil, fmt.Errorf("order already completed")
		}
		if o.Status == "expired" {
			return nil, fmt.Errorf("order has expired")
		}
		if !s.inventorySub(company, o.ResourceID, o.Quality, o.Quantity) {
			return nil, fmt.Errorf("not enough resources: need %d of resource %d (quality %d), have %d",
				o.Quantity, o.ResourceID, o.Quality, s.inventoryGet(company, o.ResourceID, o.Quality))
		}
		o.Status = "completed"
		o.CompletedAt = s.now().UTC().Format(time.RFC3339)
		s.saveCompanyLocked(company)
		s.saveStateLocked()
		return map[string]any{
			"id": o.ID, "status": o.Status,
		}, nil
	}
	return nil, fmt.Errorf("order not found")
}

// RefreshDailyOrders is called by the scheduler to regenerate daily orders.
func (s *Service) RefreshDailyOrders() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.ensureDailyOrdersLocked()
}

// ensureDailyOrdersLocked checks whether today's orders exist and generates them if not.
// Caller MUST hold s.mu.
func (s *Service) ensureDailyOrdersLocked() {
	today := s.now().UTC().Format("2006-01-02")
	if s.State.DailyOrdersDate == today && len(s.State.DailyOrders) > 0 {
		return
	}
	s.generateDailyOrdersLocked(today)
}

// generateDailyOrdersLocked creates a fresh set of daily orders.
// Caller MUST hold s.mu.
func (s *Service) generateDailyOrdersLocked(date string) {
	for i := range s.State.DailyOrders {
		if s.State.DailyOrders[i].Status == "active" {
			s.State.DailyOrders[i].Status = "expired"
		}
	}
	count := s.Cfg.Game.DailyOrderCount
	if count <= 0 {
		count = 5
	}
	tradeable := s.pickTradeableResources()
	if len(tradeable) == 0 {
		for i := 1; i <= 28; i++ {
			tradeable = append(tradeable, i)
		}
	}
	rng := rand.New(rand.NewSource(s.now().UnixNano()))
	orders := make([]model.Order, 0, count)
	used := map[int]bool{}
	for len(orders) < count {
		if len(used) >= len(tradeable) {
			used = map[int]bool{}
		}
		rid := tradeable[rng.Intn(len(tradeable))]
		if used[rid] {
			continue
		}
		used[rid] = true
		qty, rewardCash, rewardXP, quality := s.computeOrderReward(rid, rng)
		orders = append(orders, model.Order{
			ID:         fmt.Sprintf("dorder-%d-%d", s.now().UnixNano(), len(orders)),
			ResourceID: rid, Quality: quality, Quantity: qty,
			RewardCash: rewardCash, RewardXP: rewardXP, Status: "active",
			CreatedAt: s.now().UTC().Format(time.RFC3339),
		})
	}
	s.State.DailyOrders = orders
	s.State.DailyOrdersDate = date
	s.saveStateLocked()
}

// pickTradeableResources returns resource IDs that are tradeable on the exchange.
func (s *Service) pickTradeableResources() []int {
	var tradeable []int
	for _, r := range s.Data.Resources {
		rid := intFromAny(r["id"])
		if rid <= 0 || rid > 155 {
			continue
		}
		if intFromAny(r["isResearch"]) == 1 {
			continue
		}
		if tradable, ok := r["isExchangeTradable"].(bool); ok && tradable {
			tradeable = append(tradeable, rid)
		}
	}
	return tradeable
}

// computeOrderReward determines quantity, cash reward, XP reward, and quality
// for a daily order based on the resource's chain price.
func (s *Service) computeOrderReward(rid int, rng *rand.Rand) (qty int, cash float64, xp int, quality int) {
	prices := s.ComputeChainPrice(rid)
	basePrice := prices["processorPrice"]
	if basePrice <= 0 {
		basePrice = s.Cfg.Game.BotOrderBase + float64(rid%7)
	}
	var tier float64
	switch {
	case basePrice > 50:
		tier = 0.5
	case basePrice > 20:
		tier = 0.75
	case basePrice > 10:
		tier = 1.0
	case basePrice > 3:
		tier = 2.0
	default:
		tier = 4.0
	}
	qty = int(50 * tier * (0.5 + rng.Float64()))
	if qty < 1 {
		qty = 1
	}
	rewardMult := 1.2 + rng.Float64()*0.6
	cash = math.Round(basePrice*float64(qty)*rewardMult*100) / 100
	xp = int(math.Round(float64(s.Cfg.Game.DailyOrderXPBase) * tier * (0.8 + rng.Float64()*0.4)))
	if xp < 5 {
		xp = 5
	}
	if rng.Float64() < 0.25 && s.Cfg.Game.MaxQuality > 0 {
		quality = 1 + rng.Intn(min(s.Cfg.Game.MaxQuality, 5))
	}
	return
}
