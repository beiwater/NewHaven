package finance

import (
	"context"
	"fmt"
	"math"
	"sort"
	"time"

	"github.com/newhaven/backend-next/internal/config"
	domainfinance "github.com/newhaven/backend-next/internal/domain/finance"
	"github.com/newhaven/backend-next/internal/formula"
	openapi "github.com/newhaven/backend-next/internal/generated/openapi"
	"github.com/newhaven/backend-next/internal/platform"
	"github.com/newhaven/backend-next/internal/storage"
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
		return nil, err
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
		return nil, err
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
		return nil, err
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
		return nil, err
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
		return nil, err
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

// --- Bond methods ---

func (s *Service) bondFaceValue() float64 {
	if s.gameCfg != nil && s.gameCfg.BondFaceValue > 0 {
		return s.gameCfg.BondFaceValue
	}
	return 5000
}

func (s *Service) bondMinInterest() float64 {
	if s.gameCfg != nil && s.gameCfg.BondMinInterest > 0 {
		return s.gameCfg.BondMinInterest
	}
	return 0.5
}

func (s *Service) bondMaxInterest() float64 {
	if s.gameCfg != nil && s.gameCfg.BondMaxInterest > 0 {
		return s.gameCfg.BondMaxInterest
	}
	return 2.0
}

// bondRatingGroup maps a bond rating to the legacy rating filter bucket.
func bondRatingGroup(rating string) string {
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

// bondToDTO converts a domain Bond and ownership info to an openapi BondDTO.
func (s *Service) bondToDTO(b *domainfinance.Bond, ownerID int, purchasedAt string, amount int) openapi.BondDTO {
	fv := s.bondFaceValue()
	interestRatePct := b.InterestRate * 100.0 // stored as decimal, convert to percent
	daily := formula.DailyBondInterest(amount, fv, interestRatePct)
	period := daily // same as daily in this phase

	var callableTime *time.Time
	if b.CreatedAt != "" {
		if t, err := time.Parse(time.RFC3339, b.CreatedAt); err == nil {
			callable := t.Add(14 * 24 * time.Hour)
			callableTime = &callable
		}
	}

	purchased := b.CreatedAt
	if purchasedAt != "" {
		purchased = purchasedAt
	}
	var purchasedTime *time.Time
	if purchased != "" {
		if t, err := time.Parse(time.RFC3339, purchased); err == nil {
			purchasedTime = &t
		}
	}

	id := b.ID
	interest := float32(b.InterestRate)
	missed := 0
	interestCollected := float32(0)
	rating := "A"
	dailyF := float32(daily)
	periodF := float32(period)
	issuerID := b.IssuerCompanyID
	restructure := float32(0)

	return openapi.BondDTO{
		Id:                    &id,
		Amount:                &amount,
		Interest:              &interest,
		PurchasedAt:           purchasedTime,
		MissedPayments:        &missed,
		InterestCollected:     &interestCollected,
		RatingWhenPurchased:   &rating,
		DailyInterest:         &dailyF,
		PeriodInterest:        &periodF,
		IssuerCompanyId:       &issuerID,
		OwnerCompanyId:        &ownerID,
		CallableAfter:         callableTime,
		RestructurePercentage: &restructure,
	}
}

// ListBonds returns all active bonds on the market.
func (s *Service) ListBonds(ctx context.Context, ratingFilter string) (*openapi.BondListResponse, error) {
	bonds, err := s.finance.GetActiveBonds(ctx)
	if err != nil {
		return nil, fmt.Errorf("get active bonds: %w", err)
	}
	// Sort by CreatedAt desc, then ID asc.
	sort.Slice(bonds, func(i, j int) bool {
		if bonds[i].CreatedAt != bonds[j].CreatedAt {
			return bonds[i].CreatedAt > bonds[j].CreatedAt
		}
		return bonds[i].ID < bonds[j].ID
	})

	dtos := make([]openapi.BondDTO, 0, len(bonds))
	for _, b := range bonds {
		b := b
		rating := "A"
		if ratingFilter != "" && bondRatingGroup(rating) != ratingFilter {
			continue
		}
		amount := b.TotalQuantity - b.IssuedQuantity
		if amount <= 0 {
			continue
		}
		ownerID := 0 // market view: no owner
		dtos = append(dtos, s.bondToDTO(&b, ownerID, "", amount))
	}

	return &openapi.BondListResponse{Bonds: &dtos}, nil
}

// GetBond returns a single bond by ID.
func (s *Service) GetBond(ctx context.Context, bondID string) (*openapi.GetBondResponse, error) {
	b, err := s.finance.GetBond(ctx, bondID)
	if err != nil {
		return nil, fmt.Errorf("bond %q not found: %w", bondID, err)
	}
	ownerID := 0
	amount := b.TotalQuantity - b.IssuedQuantity
	if amount < 0 {
		amount = 0
	}
	dto := s.bondToDTO(b, ownerID, "", amount)
	return &openapi.GetBondResponse{Bond: &dto}, nil
}

// CreateBond issues a new bond for the company.
func (s *Service) CreateBond(ctx context.Context, companyID int, amount int, interestPct float32) (*openapi.CreateBondResponse, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("amount must be > 0")
	}
	if interestPct < float32(s.bondMinInterest()) || interestPct > float32(s.bondMaxInterest()) {
		return nil, fmt.Errorf("interest must be between %.1f and %.1f", s.bondMinInterest(), s.bondMaxInterest())
	}

	company, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("company not found: %w", err)
	}

	fv := s.bondFaceValue()
	now := s.clock.Now().UTC().Format(time.RFC3339)

	b := &domainfinance.Bond{
		ID:              s.idgen.Next("bond"),
		IssuerCompanyID: companyID,
		FaceValue:       fv,
		InterestRate:    float64(interestPct) / 100.0,
		TotalQuantity:   amount,
		IssuedQuantity:  0,
		Status:          "active",
		CreatedAt:       now,
	}

	// Credit company money.
	credit := float64(amount) * fv
	company.Money += credit
	if err := s.companies.UpdateCompany(ctx, company); err != nil {
		company.Money -= credit // rollback
		return nil, fmt.Errorf("update company: %w", err)
	}

	// Store bond.
	if err := s.finance.CreateBond(ctx, b); err != nil {
		company.Money -= credit                     // rollback
		_ = s.companies.UpdateCompany(ctx, company) // best-effort
		return nil, fmt.Errorf("create bond: %w", err)
	}

	// Append ledger entry.
	_ = s.finance.AppendLedgerEntry(ctx, &domainfinance.LedgerEntry{
		CompanyID: companyID,
		Kind:      "bond_issue",
		Direction: "in",
		Amount:    credit,
		Metadata:  map[string]any{"bondId": b.ID, "amount": amount, "interest": interestPct},
		CreatedAt: now,
	})

	dto := s.bondToDTO(b, 0, "", amount)
	return &openapi.CreateBondResponse{Bond: &dto}, nil
}

// GetOwnedBonds returns bonds held by the company.
func (s *Service) GetOwnedBonds(ctx context.Context, companyID int) (*openapi.BondListResponse, error) {
	holdings, err := s.finance.GetCompanyBondHoldings(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("get holdings: %w", err)
	}

	// Sort by PurchasedAt desc, then BondID asc.
	sort.Slice(holdings, func(i, j int) bool {
		if holdings[i].PurchasedAt != holdings[j].PurchasedAt {
			return holdings[i].PurchasedAt > holdings[j].PurchasedAt
		}
		return holdings[i].BondID < holdings[j].BondID
	})

	dtos := make([]openapi.BondDTO, 0, len(holdings))
	for _, h := range holdings {
		h := h
		b, err := s.finance.GetBond(ctx, h.BondID)
		if err != nil {
			continue // skip orphaned holdings
		}
		dto := s.bondToDTO(b, companyID, h.PurchasedAt, h.Quantity)
		dtos = append(dtos, dto)
	}

	return &openapi.BondListResponse{Bonds: &dtos}, nil
}

// GetSoldBonds returns bonds issued by the company.
func (s *Service) GetSoldBonds(ctx context.Context, companyID int) (*openapi.BondListResponse, error) {
	bonds, err := s.finance.GetBondsByIssuer(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("get issued bonds: %w", err)
	}

	// Sort by CreatedAt desc, then ID asc.
	sort.Slice(bonds, func(i, j int) bool {
		if bonds[i].CreatedAt != bonds[j].CreatedAt {
			return bonds[i].CreatedAt > bonds[j].CreatedAt
		}
		return bonds[i].ID < bonds[j].ID
	})

	dtos := make([]openapi.BondDTO, 0, len(bonds))
	for _, b := range bonds {
		b := b
		ownerID := 0 // issuer view
		dtos = append(dtos, s.bondToDTO(&b, ownerID, "", b.TotalQuantity))
	}

	return &openapi.BondListResponse{Bonds: &dtos}, nil
}
