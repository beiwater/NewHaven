package service

import (
	"fmt"
	"time"

	"go-sim-api/internal/anticheat"
	"go-sim-api/internal/model"
)

func (s *Service) AvailableAuctionList() []model.Auction {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.processAuctionDeadlines()
	out := make([]model.Auction, 0, len(s.State.Auctions))
	for _, a := range s.State.Auctions {
		if a.Status == "open" {
			out = append(out, a)
		}
	}
	return out
}

func (s *Service) MyAuctionList(companyID int) ([]model.Auction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}
	s.processAuctionDeadlines()
	out := make([]model.Auction, 0, len(s.State.Auctions))
	for _, a := range s.State.Auctions {
		if a.HighestBidder == company.ID || a.Status == "awarded" {
			out = append(out, a)
		}
	}
	return out, nil
}

func (s *Service) AuctionDetail(id string) (*model.Auction, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.processAuctionDeadlines()
	for i := range s.State.Auctions {
		if s.State.Auctions[i].ID == id {
			return &s.State.Auctions[i], nil
		}
	}
	return nil, fmt.Errorf("auction not found")
}

func (s *Service) PlaceAuctionBid(companyID int, auctionID string, amount float64) (map[string]any, error) {
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
	s.AC.RecordAction(pid, anticheat.ActAuctionBid, fmt.Sprintf("auction=%s amount=%.2f", auctionID, amount))
	s.SD.RecordAction(pid)
	s.processAuctionDeadlines()
	for i := range s.State.Auctions {
		a := &s.State.Auctions[i]
		if a.ID != auctionID {
			continue
		}
		if a.Status != "open" {
			return nil, fmt.Errorf("auction %s is not open", auctionID)
		}
		minBid := a.CurrentBid * 1.05 // must exceed current by at least 5%
		if amount < minBid {
			return nil, fmt.Errorf("bid too low: minimum is %.2f (current %.2f + 5%%)", minBid, a.CurrentBid)
		}
		if company.Money < amount {
			return nil, fmt.Errorf("insufficient funds: have %.2f, need %.2f", company.Money, amount)
		}
		// If previous highest bidder is someone else, refund their deposit
		if a.HighestBidder > 0 && a.HighestBidder != company.ID {
			for j := range s.State.Companies {
				if s.State.Companies[j].ID == a.HighestBidder {
					s.State.Companies[j].Money += a.CurrentBid
					break
				}
			}
		}
		// Deduct new bid amount
		company.Money -= amount
		a.CurrentBid = amount
		a.HighestBidder = company.ID
		a.Bids = append(a.Bids, model.AuctionBid{
			CompanyID: company.ID,
			Amount:    amount,
			At:        s.now().UTC().Format(time.RFC3339),
		})
		s.addLedger("auction_bid", -amount, "out", map[string]any{"auction": auctionID})
		s.saveCompanyLocked(company)
		s.saveStateLocked()
		return map[string]any{"auction": a, "status": "bid_placed"}, nil
	}
	return nil, fmt.Errorf("auction not found")
}

func (s *Service) processAuctionDeadlines() {
	now := s.now().UTC()
	for i := range s.State.Auctions {
		a := &s.State.Auctions[i]
		if a.Status != "open" {
			continue
		}
		endsAt, err := time.Parse(time.RFC3339, a.EndsAt)
		if err != nil || now.Before(endsAt) {
			continue
		}
		// Auction ended, award to highest bidder
		if a.HighestBidder == 0 {
			a.Status = "expired"
			continue
		}
		winner := s.getCompanyLocked(a.HighestBidder)
		if winner != nil {
			winner.PlacedBuildings = append(winner.PlacedBuildings, map[string]any{
				"id":          a.ItemID,
				"kind":        2,
				"level":       1,
				"baseCost":    10000,
				"busy":        false,
				"x":           0,
				"y":           0,
				"placedAt":    now.Format(time.RFC3339),
				"fromAuction": true,
			})
			s.addLedger("auction_win", -a.CurrentBid, "out", map[string]any{"auction": a.ID, "item": a.ItemID})
		}
		a.Status = "awarded"
	}
}
