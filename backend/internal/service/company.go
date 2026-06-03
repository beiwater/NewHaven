package service

import (
	"fmt"
	"go-sim-api/internal/formula"
	"math"
)

func (s *Service) CompanyProfile(companyID int) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return map[string]any{"error": "company not found"}
	}

	return map[string]any{
		"authCompany": map[string]any{
			"companyId":        company.ID,
			"company":          company.Name,
			"money":            company.Money,
			"simBoosts":        3,
			"exchangedToday":   5000.0,
			"maxTags":          5,
			"displayCaseSlots": 3,
		},
		"authUser":    map[string]any{"playerId": "dev-player", "isModerator": false, "supporter": false},
		"levelInfo":   map[string]any{"level": company.Level, "xp": 12800, "inTutorial": false},
		"unlocks":     FeatureUnlockPayload(company.Level),
		"preferences": map[string]any{"theme": "System"},
	}
}

func (s *Service) CompaniesByPlayer(playerID string) []map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]map[string]any, 0, len(s.State.Companies))
	for _, c := range s.State.Companies {
		out = append(out, map[string]any{"companyId": c.ID, "company": c.Name, "playerId": playerID, "level": c.Level})
	}
	return out
}

func (s *Service) COOandCTOSkill() (coo float64, cto float64) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, e := range s.State.Executives {
		role := fmt.Sprint(e["role"])
		skill := floatFromAny(e["skill"])
		if role == "COO" {
			coo = skill
		}
		if role == "CTO" {
			cto = skill
		}
	}
	return
}

func (s *Service) FinancialStatements(companyID int) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return map[string]any{"error": "company not found"}
	}

	revenue := 0.0
	expenses := 0.0
	operating := 0.0
	investing := 0.0
	financing := 0.0
	for _, e := range s.State.Ledger {
		kind := e.Kind
		amt := math.Abs(e.Amount)
		dir := e.Direction
		sign := 1.0
		if dir == "out" {
			sign = -1.0
		}
		switch kind {
		case "market_trade", "gov_contract_reward", "bond_interest_income":
			revenue += amt
		case "market_fee", "market_take_buy", "bond_interest_expense", "gov_bid_deposit":
			expenses += amt
		}
		switch kind {
		case "market_trade", "market_take_buy", "production_input", "production_output", "gov_contract_reward":
			operating += amt * sign
		case "bond_buy", "bond_call":
			investing += amt * sign
		case "bond_issue", "bond_interest_income", "bond_interest_expense":
			financing += amt * sign
		}
	}
	assets := company.Money
	for _, v := range company.Inventory {
		assets += float64(v) * 10
	}
	liabilities := 0.0
	for _, b := range s.State.Bonds {
		if b.IssuerCompanyID == company.ID {
			liabilities += float64(b.Amount) * formula.BondFaceValue
		}
	}
	return map[string]any{
		"incomeStatement":   map[string]any{"revenue": revenue, "expenses": expenses, "netIncome": revenue - expenses},
		"cashflowStatement": map[string]any{"operating": operating, "investing": investing, "financing": financing},
		"balanceSheet":      map[string]any{"assets": assets, "liabilities": liabilities, "equity": assets - liabilities},
	}
}
