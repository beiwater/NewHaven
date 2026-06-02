package service

import (
	"context"
	"log"
	"math"
	"time"

	"go-sim-api/internal/model"
)

func (s *Service) CleanupOrders() {
	s.mu.Lock()
	defer s.mu.Unlock()
	active := make([]model.MarketOrder, 0, len(s.State.Orders))
	botIDs := map[int]bool{s.Cfg.Game.Bot1ID: true, s.Cfg.Game.Bot2ID: true}
	for _, o := range s.State.Orders {
		// Remove filled/cancelled orders
		if o.Status == "filled" || o.Status == "cancelled" {
			continue
		}
		// Keep active non-bot orders with remaining
		if !botIDs[o.CompanyID] {
			active = append(active, o)
			continue
		}
		// Keep bot orders that have remaining (they provide liquidity)
		if o.Remaining > 0 {
			active = append(active, o)
		}
	}
	s.State.Orders = active
	// Trim ledger
	if len(s.State.Ledger) > s.Cfg.Game.MaxLedgerEntries {
		s.State.Ledger = s.State.Ledger[:s.Cfg.Game.MaxLedgerEntries]
	}
	// Decay market pressure toward 0
	for k, v := range s.State.MarketPressure {
		s.State.MarketPressure[k] = v * 0.9
		if math.Abs(s.State.MarketPressure[k]) < 0.001 {
			delete(s.State.MarketPressure, k)
		}
	}
}

func (s *Service) SaveAll() {
	if s.Store == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.Store.SaveState(ctx, &s.State); err != nil {
		log.Printf("[persist] save error: %v", err)
	}
}
