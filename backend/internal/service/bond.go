package service

import (
	"fmt"
	"time"

	"go-sim-api/internal/anticheat"
	"go-sim-api/internal/formula"
	"go-sim-api/internal/model"
)

func (s *Service) BondMarketView() []map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]map[string]any, 0, len(s.State.Bonds))
	for _, b := range s.State.Bonds {
		interest := b.Interest * 100
		out = append(out, map[string]any{
			"id": b.ID, "amount": b.Amount, "interest": b.Interest,
			"purchased_at": b.PurchasedAt, "missed_payments": b.MissedPayments,
			"interestCollected": b.InterestCollected, "ratingWhenPurchased": b.RatingWhenPurchased,
			"dailyInterest":  formula.DailyBondInterest(b.Amount, interest),
			"periodInterest": formula.PeriodBondInterest(b.Amount, interest),
		})
	}
	return out
}
func (s *Service) IssueOrAdjustBond(companyID int, amount int, interestPct float64) (map[string]any, error) {
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
	s.AC.RecordAction(pid, anticheat.ActBondIssue, fmt.Sprintf("amount=%d interest=%.2f%%", amount, interestPct))
	s.SD.RecordAction(pid)
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be > 0")
	}
	if interestPct < s.Cfg.Game.BondMinInterest || interestPct > s.Cfg.Game.BondMaxInterest {
		return nil, fmt.Errorf("interest must be between 0.5 and 2.0")
	}
	interest := interestPct / 100.0
	now := s.now().UTC().Format(time.RFC3339)
	b := model.Bond{
		ID:                  fmt.Sprintf("bond-%d", s.now().UnixNano()),
		Amount:              amount,
		Interest:            interest,
		PurchasedAt:         now,
		MissedPayments:      0,
		InterestCollected:   0.0,
		RatingWhenPurchased: "A",
		IssuerCompanyID:     company.ID,
		OwnerCompanyID:      0,
		CallableAfter:       s.now().UTC().Add(14 * 24 * time.Hour).Format(time.RFC3339),
		RestructurePct:      0.0,
	}
	s.State.Bonds = append([]model.Bond{b}, s.State.Bonds...)
	company.Money += float64(amount) * formula.BondFaceValue
	s.addLedger("bond_issue", float64(amount)*formula.BondFaceValue, "in", map[string]any{"bondId": b.ID})
	s.saveCompanyLocked(company)
	return map[string]any{
		"id": b.ID, "amount": b.Amount, "interest": b.Interest,
		"purchased_at": b.PurchasedAt, "missed_payments": b.MissedPayments,
		"interestCollected": b.InterestCollected, "ratingWhenPurchased": b.RatingWhenPurchased,
		"issuerCompanyId": b.IssuerCompanyID, "ownerCompanyId": b.OwnerCompanyID,
		"callableAfter": b.CallableAfter, "restructure_percentage": b.RestructurePct,
	}, nil
}
func (s *Service) BuyBond(companyID int, bondID string, amount int) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be > 0")
	}
	for i := range s.State.Bonds {
		b := &s.State.Bonds[i]
		if b.ID != bondID {
			continue
		}
		if amount > b.Amount {
			return nil, fmt.Errorf("insufficient bond amount")
		}
		if b.IssuerCompanyID == company.ID {
			return nil, fmt.Errorf("cannot buy own bonds")
		}
		maxPct := 0.01
		if company.Level >= 15 {
			maxPct = 0.03
		}
		if company.Level >= 20 {
			maxPct = 0.05
		}
		capUnits := max(1, int(float64(b.Amount)*maxPct))
		if amount > capUnits {
			return nil, fmt.Errorf("buy amount exceeds level cap (%d units)", capUnits)
		}
		cost := float64(amount) * formula.BondFaceValue
		if company.Money < cost {
			return nil, fmt.Errorf("not enough cash")
		}
		company.Money -= cost
		s.addLedger("bond_buy", cost, "out", map[string]any{"bondId": bondID, "amount": amount})
		b.Amount -= amount
		owned := model.Bond{
			ID:                  fmt.Sprintf("owned-%d", s.now().UnixNano()),
			Amount:              amount,
			Interest:            b.Interest,
			PurchasedAt:         s.now().UTC().Format(time.RFC3339),
			MissedPayments:      0,
			InterestCollected:   0.0,
			RatingWhenPurchased: b.RatingWhenPurchased,
			IssuerCompanyID:     b.IssuerCompanyID,
			OwnerCompanyID:      company.ID,
			CallableAfter:       s.now().UTC().Add(14 * 24 * time.Hour).Format(time.RFC3339),
			RestructurePct:      0.0,
		}
		s.State.Bonds = append([]model.Bond{owned}, s.State.Bonds...)
		s.saveStateLocked()
		return map[string]any{
			"id": owned.ID, "amount": owned.Amount, "interest": owned.Interest,
			"purchased_at": owned.PurchasedAt, "missed_payments": owned.MissedPayments,
			"interestCollected": owned.InterestCollected, "ratingWhenPurchased": owned.RatingWhenPurchased,
			"issuerCompanyId": owned.IssuerCompanyID, "ownerCompanyId": owned.OwnerCompanyID,
			"callableAfter": owned.CallableAfter, "restructure_percentage": owned.RestructurePct,
		}, nil
	}
	return nil, fmt.Errorf("bond not found")
}
func (s *Service) BondRatingGroup(rating string) string {
	switch rating {
	case "AAA", "AA+", "AA":
		return "AAA to AA"
	case "AA-", "A+", "A":
		return "AA- to A"
	case "A-", "BBB+", "BBB":
		return "A- to BBB"
	case "BBB-", "BB+", "BB":
		return "BBB- to BB"
	case "BB-", "B+", "B":
		return "BB- to B"
	case "B-", "C":
		return "B- to C"
	default:
		return "D to D"
	}
}
func (s *Service) CallBond(companyID int, bondID string, amount int) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be > 0")
	}
	for i := range s.State.Bonds {
		b := &s.State.Bonds[i]
		if b.ID != bondID {
			continue
		}
		callableAfter, _ := time.Parse(time.RFC3339, b.CallableAfter)
		if s.now().UTC().Before(callableAfter) {
			return nil, fmt.Errorf("bond is still locked for 14 days")
		}
		if amount > b.Amount {
			return nil, fmt.Errorf("insufficient amount")
		}
		cashNeed := float64(amount) * formula.BondFaceValue
		if company.Money < cashNeed {
			return nil, fmt.Errorf("not enough cash to call bond")
		}
		company.Money -= cashNeed
		s.addLedger("bond_call", cashNeed, "out", map[string]any{"bondId": bondID, "amount": amount})
		b.Amount -= amount
		s.saveStateLocked()
		return map[string]any{"id": bondID, "called": amount, "cashDelta": -cashNeed}, nil
	}
	return nil, fmt.Errorf("bond not found")
}
func (s *Service) SettleBondInterest(companyID int) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return map[string]any{"dailyBondIncome": 0.0, "dailyBondExpense": 0.0, "defaults": []map[string]any{}, "money": 0.0}
	}
	totalIncome := 0.0
	totalExpense := 0.0
	defaults := []map[string]any{}
	for i := range s.State.Bonds {
		b := &s.State.Bonds[i]
		amount := b.Amount
		interest := b.Interest * 100
		daily := formula.DailyBondInterest(amount, interest)
		issuer := b.IssuerCompanyID
		owner := b.OwnerCompanyID
		if owner == company.ID && issuer != company.ID {
			totalIncome += daily
			b.InterestCollected += daily
			continue
		}
		if issuer == company.ID && owner != 0 && owner != company.ID {
			totalExpense += daily
			s.addLedger("bond_interest_expense", daily, "out", map[string]any{"bondId": b.ID})
		}
	}
	net := totalIncome - totalExpense
	if net >= 0 || company.Money >= -net {
		company.Money += net
	} else {
		company.Money = 0
		for i := range s.State.Bonds {
			b := &s.State.Bonds[i]
			issuer := b.IssuerCompanyID
			owner := b.OwnerCompanyID
			if issuer == company.ID && owner != 0 && owner != company.ID {
				b.MissedPayments++
				restructure := min(80, int(b.RestructurePct+10))
				b.RestructurePct = float64(restructure)
				defaults = append(defaults, map[string]any{
					"id": b.ID, "missed_payments": b.MissedPayments, "restructure_percentage": restructure,
				})
				s.addLedger("bond_default", 0, "out", map[string]any{"bondId": b.ID, "missed": b.MissedPayments})
			}
		}
	}
	if totalIncome > 0 {
		s.addLedger("bond_interest_income", totalIncome, "in", map[string]any{})
	}
	s.saveStateLocked()
	return map[string]any{
		"dailyBondIncome": totalIncome, "dailyBondExpense": totalExpense,
		"defaults": defaults, "money": company.Money,
	}
}
