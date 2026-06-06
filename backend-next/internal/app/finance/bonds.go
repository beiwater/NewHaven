package finance

import (
	"context"
	"sort"
	"time"

	"github.com/newhaven/backend-next/internal/apperr"
	domainfinance "github.com/newhaven/backend-next/internal/domain/finance"
	"github.com/newhaven/backend-next/internal/formula"
	openapi "github.com/newhaven/backend-next/internal/generated/openapi"
)

// --- Bond helpers ---

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
		return nil, apperr.Internalf("get active bonds: %v", err)
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
		return nil, apperr.NotFoundf("bond %s not found", bondID)
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
		return nil, apperr.Validation("amount must be > 0")
	}
	if interestPct < float32(s.bondMinInterest()) || interestPct > float32(s.bondMaxInterest()) {
		return nil, apperr.Validationf("interest must be between %.1f and %.1f", s.bondMinInterest(), s.bondMaxInterest())
	}

	company, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "company not found", err)
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
		return nil, apperr.Internalf("update company: %v", err)
	}

	// Store bond.
	if err := s.finance.CreateBond(ctx, b); err != nil {
		company.Money -= credit                     // rollback
		_ = s.companies.UpdateCompany(ctx, company) // best-effort
		return nil, apperr.Internalf("create bond: %v", err)
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
		return nil, apperr.Internalf("get holdings: %v", err)
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
		return nil, apperr.Internalf("get issued bonds: %v", err)
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
