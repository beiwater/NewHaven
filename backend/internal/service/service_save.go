package service

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"go-sim-api/internal/model"
)

// Now returns the effective current time (exported).
func (s *Service) Now() time.Time {
	return s.now()
}

func (s *Service) addLedger(kind string, amount float64, direction string, meta map[string]any) {
	if amount < 0 {
		amount = math.Abs(amount)
	}
	entry := model.LedgerEntry{
		ID: fmt.Sprintf("led-%d", s.now().UnixNano()), At: s.now().UTC().Format(time.RFC3339),
		Kind: kind, Amount: amount, Direction: direction, Meta: meta,
	}
	s.State.Ledger = append([]model.LedgerEntry{entry}, s.State.Ledger...)
	if len(s.State.Ledger) > s.Cfg.Game.MaxLedgerEntries {
		s.State.Ledger = s.State.Ledger[:s.Cfg.Game.MaxLedgerEntries]
	}
}

// --- Lock-free variants (caller already holds s.mu) ---

func (s *Service) saveStateLocked() {
	if s.Store == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := s.Store.SaveState(ctx, &s.State); err != nil {
		log.Printf("ERROR save state: %v", err)
		s.setSaveError(err)
		return
	}
	s.setSaveError(nil)
}

func (s *Service) saveCompanyLocked(company *model.Company) {
	if s.Store == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := s.Store.SaveCompany(ctx, company); err != nil {
		log.Printf("ERROR save company: %v", err)
		s.setSaveError(err)
		return
	}
	if err := s.Store.SaveState(ctx, &s.State); err != nil {
		log.Printf("ERROR save state after company save: %v", err)
		s.setSaveError(err)
		return
	}
	s.setSaveError(nil)
}

func (s *Service) saveOrdersLocked() {
	if s.Store == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := s.Store.SaveOrders(ctx, s.State.Orders); err != nil {
		log.Printf("ERROR save orders: %v", err)
		s.setSaveError(err)
		return
	}
	s.setSaveError(nil)
}

func (s *Service) saveTradesLocked() {
	if s.Store == nil {
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := s.Store.SaveTrades(ctx, s.State.Trades); err != nil {
		log.Printf("ERROR save trades: %v", err)
		s.setSaveError(err)
		return
	}
	s.setSaveError(nil)
}

func (s *Service) checkIdempotent(requestID string) (map[string]any, bool) {
	if requestID == "" {
		return nil, false
	}
	if s.State.ProcessedRequests == nil {
		s.State.ProcessedRequests = map[string]map[string]any{}
	}
	if cached, ok := s.State.ProcessedRequests[requestID]; ok {
		return cached, true
	}
	return nil, false
}

func (s *Service) markIdempotent(requestID string, result map[string]any) {
	if requestID == "" {
		return
	}
	if s.State.ProcessedRequests == nil {
		s.State.ProcessedRequests = map[string]map[string]any{}
	}
	s.State.ProcessedRequests[requestID] = result
	// Trim old entries (keep last 1000)
	if len(s.State.ProcessedRequests) > 1000 {
		for k := range s.State.ProcessedRequests {
			delete(s.State.ProcessedRequests, k)
			break
		}
	}
}

func (s *Service) CheckIdempotent(requestID string) (map[string]any, bool) {
	return s.checkIdempotent(requestID)
}

func (s *Service) MarkIdempotent(requestID string, result map[string]any) {
	s.markIdempotent(requestID, result)
}
