package finance

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/beiwater/NewHaven/backend/internal/apperr"
	"github.com/beiwater/NewHaven/backend/internal/config"
	openapi "github.com/beiwater/NewHaven/backend/internal/generated/openapi"
	"github.com/beiwater/NewHaven/backend/internal/platform"
	"github.com/beiwater/NewHaven/backend/internal/storage"
)

// Service provides financial and bond views.
type Service struct {
	finance   storage.FinanceStorage
	companies storage.CompanyStorage
	clock     platform.Clock
	idgen     *platform.IDGen
	gameCfg   *config.GameConfig
}

// NewService creates a new finance service.
func NewService(finance storage.FinanceStorage, companies storage.CompanyStorage, clock platform.Clock, idgen *platform.IDGen, gameCfg *config.GameConfig) *Service {
	if idgen == nil {
		idgen = platform.NewIDGen()
	}
	if gameCfg == nil {
		gameCfg = &config.GameConfig{
			BondFaceValue:   5000,
			BondMinInterest: 0.5,
			BondMaxInterest: 2.0,
		}
	}
	return &Service{
		finance:   finance,
		companies: companies,
		clock:     clock,
		idgen:     idgen,
		gameCfg:   gameCfg,
	}
}

// GetRecentCashflow returns the most recent cashflow entries with signed moneyDelta.
func (s *Service) GetRecentCashflow(ctx context.Context, companyID int) (*openapi.RecentCashflowResponse, error) {
	company, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "company not found", err)
	}

	entries, err := s.finance.GetLedgerEntries(ctx, companyID, 100)
	if err != nil {
		return nil, err
	}

	now := s.clock.Now().UTC()
	data := make([]openapi.CashflowEntry, 0, len(entries))
	var oldestPulled time.Time

	for i, e := range entries {
		delta := math.Abs(e.Amount)
		if e.Direction == "out" {
			delta = -delta
		}
		delta32 := float32(delta)

		var at time.Time
		if e.CreatedAt != "" {
			at, _ = time.Parse(time.RFC3339, e.CreatedAt)
		}
		if at.IsZero() {
			at = now
		}
		if i == len(entries)-1 {
			oldestPulled = at
		}

		entry := openapi.CashflowEntry{
			Kind:       &e.Kind,
			MoneyDelta: &delta32,
			At:         &at,
		}
		data = append(data, entry)
	}

	if oldestPulled.IsZero() {
		oldestPulled = now
	}

	money := float32(company.Money)
	return &openapi.RecentCashflowResponse{
		Data:         &data,
		OldestPulled: &oldestPulled,
		Money:        &money,
	}, nil
}

// GetIncomeStatement computes revenue, expenses, and net income from the ledger.
func (s *Service) GetIncomeStatement(ctx context.Context, companyID int) (*openapi.IncomeStatementResponse, error) {
	if _, err := s.companies.GetCompany(ctx, companyID); err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "company not found", err)
	}

	entries, err := s.finance.GetLedgerEntries(ctx, companyID, 1000)
	if err != nil {
		return nil, err
	}

	var revenue, expenses float64
	for _, e := range entries {
		if e.Direction == "in" {
			revenue += math.Abs(e.Amount)
		} else {
			expenses += math.Abs(e.Amount)
		}
	}
	netIncome := revenue - expenses

	rev32 := float32(revenue)
	exp32 := float32(expenses)
	net32 := float32(netIncome)

	return &openapi.IncomeStatementResponse{
		Revenue:   &rev32,
		Expenses:  &exp32,
		NetIncome: &net32,
	}, nil
}

// GetBalanceSheet returns assets (company money) and liabilities/equity.
func (s *Service) GetBalanceSheet(ctx context.Context, companyID int) (*openapi.BalanceSheetResponse, error) {
	company, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "company not found", err)
	}

	assets := float32(company.Money)
	liabilities := float32(0)
	equity := assets

	return &openapi.BalanceSheetResponse{
		Assets:      &assets,
		Liabilities: &liabilities,
		Equity:      &equity,
	}, nil
}

// GetCashflowStatement categorizes ledger entries into operating/investing/financing.
func (s *Service) GetCashflowStatement(ctx context.Context, companyID int) (*openapi.CashflowStatementResponse, error) {
	if _, err := s.companies.GetCompany(ctx, companyID); err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "company not found", err)
	}

	entries, err := s.finance.GetLedgerEntries(ctx, companyID, 1000)
	if err != nil {
		return nil, err
	}

	var operating, investing, financing float64
	for _, e := range entries {
		delta := math.Abs(e.Amount)
		if e.Direction == "out" {
			delta = -delta
		}
		switch e.Kind {
		case "bond_issue", "bond_buy", "bond_call", "bond_interest_income", "bond_interest_expense":
			financing += delta
		case "buy_building", "building_upgrade", "demolish_building", "warehouse_upgrade", "research_start", "research_complete", "slot_upgrade":
			investing += delta
		default:
			operating += delta
		}
	}

	op32 := float32(operating)
	inv32 := float32(investing)
	fin32 := float32(financing)

	return &openapi.CashflowStatementResponse{
		Operating: &op32,
		Investing: &inv32,
		Financing: &fin32,
	}, nil
}

// GetPastFinances returns a daily net series computed from the ledger,
// falling back to a deterministic sample series when there are no timestamped entries.
func (s *Service) GetPastFinances(ctx context.Context, companyID int) (*openapi.PastFinancesResponse, error) {
	if _, err := s.companies.GetCompany(ctx, companyID); err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "company not found", err)
	}

	entries, err := s.finance.GetLedgerEntries(ctx, companyID, 1000)
	if err != nil {
		return nil, err
	}

	// Accumulate net by date (using date portion of CreatedAt).
	dayNet := make(map[string]float64)
	hasTimestamped := false
	for _, e := range entries {
		if e.CreatedAt == "" {
			continue
		}
		ts, parseErr := time.Parse(time.RFC3339, e.CreatedAt)
		if parseErr != nil {
			continue
		}
		hasTimestamped = true
		date := ts.Format("2006-01-02")
		delta := math.Abs(e.Amount)
		if e.Direction == "out" {
			delta = -delta
		}
		dayNet[date] += delta
	}

	// Build sorted series.
	type dayPoint struct {
		date string
		net  float64
	}
	points := make([]dayPoint, 0, len(dayNet))
	if hasTimestamped {
		for d, n := range dayNet {
			points = append(points, dayPoint{date: d, net: n})
		}
		sort.Slice(points, func(i, j int) bool {
			return points[i].date < points[j].date
		})
	}

	// Fallback if no data.
	if len(points) == 0 {
		d1 := "2026-05-28"
		n1 := float32(890.2)
		d2 := "2026-05-29"
		n2 := float32(1022.4)
		series := []openapi.PastFinancePoint{
			{Date: &d1, Net: &n1},
			{Date: &d2, Net: &n2},
		}
		return &openapi.PastFinancesResponse{
			Series: &series,
		}, nil
	}

	series := make([]openapi.PastFinancePoint, len(points))
	for i, p := range points {
		date := p.date
		net := float32(p.net)
		series[i] = openapi.PastFinancePoint{
			Date: &date,
			Net:  &net,
		}
	}

	return &openapi.PastFinancesResponse{
		Series: &series,
	}, nil
}
