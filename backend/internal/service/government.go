package service

import (
	"fmt"
	"time"

	"go-sim-api/internal/anticheat"
	"go-sim-api/internal/model"
)

func (s *Service) PlaceGovernmentBid(companyID int, contractID string, unitPrice float64) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}
	pid := company.ID
	if ok, msg := s.AC.CheckRateLimit(pid); !ok {
		return nil, fmt.Errorf("cheat detected: %s", msg)
	}
	s.AC.RecordAction(pid, anticheat.ActGovBid, fmt.Sprintf("contract=%s price=%.2f", contractID, unitPrice))
	s.SD.RecordAction(pid)

	if unitPrice <= 0 {
		return nil, fmt.Errorf("invalid price")
	}
	for i := range s.State.GovernmentContracts {
		c := &s.State.GovernmentContracts[i]
		if c.ID != contractID {
			continue
		}
		if c.Status != "open" {
			return nil, fmt.Errorf("contract closed")
		}
		deposit := float64(c.Quantity) * unitPrice * c.DepositRate
		if company.Money < deposit {
			return nil, fmt.Errorf("not enough money for deposit")
		}
		company.Money -= deposit
		s.addLedger("gov_bid_deposit", deposit, "out", map[string]any{"contractId": contractID})
		bid := map[string]any{
			"companyId": company.ID,
			"unitPrice": unitPrice,
			"deposit":   deposit,
			"placedAt":  s.now().UTC().Format(time.RFC3339),
		}
		bids := c.Bids
		if bids == nil {
			bids = []map[string]any{}
		}
		// Check existing bids
		for _, existing := range bids {
			if intFromAny(existing["companyId"]) == company.ID {
				return nil, fmt.Errorf("already placed a bid on this contract")
			}
		}

		bids = append(bids, bid)
		c.Bids = bids
		s.saveStateLocked()
		return bid, nil
	}
	return nil, fmt.Errorf("contract not found")
}

func (s *Service) AwardGovernmentContracts() []model.GovContract {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := []model.GovContract{}
	for i := range s.State.GovernmentContracts {
		c := &s.State.GovernmentContracts[i]
		if c.Status != "open" {
			continue
		}
		bids := c.Bids
		if len(bids) == 0 {
			continue
		}
		best := bids[0]
		for _, b := range bids[1:] {
			if floatFromAny(b["unitPrice"]) < floatFromAny(best["unitPrice"]) {
				best = b
			}
		}
		c.WinnerCompanyID = intFromAny(best["companyId"])
		c.WinningPrice = floatFromAny(best["unitPrice"])
		c.Status = "awarded"
		c.AwardedAt = s.now().UTC().Format(time.RFC3339)
		c.DueAt = s.now().UTC().Add(24 * time.Hour).Format(time.RFC3339)
		for _, b := range bids {
			bidCompanyID := intFromAny(b["companyId"])
			if bidCompanyID == intFromAny(best["companyId"]) {
				continue
			}
			bidCompany := s.getCompanyLocked(bidCompanyID)
			if bidCompany != nil {
				dep := floatFromAny(b["deposit"])
				bidCompany.Money += dep * s.Cfg.Game.GovBidRefundRate
				s.addLedger("gov_bid_refund_partial", dep*s.Cfg.Game.GovBidRefundRate, "in", map[string]any{"contractId": c.ID})
			}
		}
		out = append(out, *c)
	}
	return out
}

func (s *Service) DeliverGovernmentContract(companyID int, contractID string) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}

	for i := range s.State.GovernmentContracts {
		c := &s.State.GovernmentContracts[i]
		if c.ID != contractID {
			continue
		}
		if c.Status != "awarded" {
			return nil, fmt.Errorf("contract not awarded")
		}
		if company.Inventory[c.ResourceID] < c.Quantity {
			return nil, fmt.Errorf("insufficient inventory")
		}
		reward := float64(c.Quantity) * c.WinningPrice
		company.Inventory[c.ResourceID] -= c.Quantity
		company.Money += reward
		s.addLedger("gov_contract_reward", reward, "in", map[string]any{"contractId": contractID})
		c.Status = "delivered"
		s.addXP(company, 50)
		c.DeliveredAt = s.now().UTC().Format(time.RFC3339)
		// return deposit when fulfilled
		for _, b := range c.Bids {
			if intFromAny(b["companyId"]) == company.ID {
				company.Money += floatFromAny(b["deposit"])
				s.addLedger("gov_bid_refund_full", floatFromAny(b["deposit"]), "in", map[string]any{"contractId": contractID})
				break
			}
		}
		s.saveCompanyLocked(company)
		return map[string]any{"id": contractID, "reward": reward, "money": company.Money}, nil
	}
	return nil, fmt.Errorf("contract not found")
}

func (s *Service) ResolveGovernmentDefaults() []model.GovContract {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.now().UTC()
	defaulted := []model.GovContract{}
	for i := range s.State.GovernmentContracts {
		c := &s.State.GovernmentContracts[i]
		if c.Status != "awarded" {
			continue
		}
		dueAt, err := time.Parse(time.RFC3339, c.DueAt)
		if err != nil || now.Before(dueAt) {
			continue
		}
		c.Status = "defaulted"
		c.DefaultedAt = now.Format(time.RFC3339)
		c.Penalty = "deposit_forfeited"
		s.addLedger("gov_contract_default", 0, "out", map[string]any{"contractId": c.ID})
		defaulted = append(defaulted, *c)
	}
	return defaulted
}
