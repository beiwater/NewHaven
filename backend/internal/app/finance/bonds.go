package finance

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/beiwater/NewHaven/backend/internal/apperr"
	domainfinance "github.com/beiwater/NewHaven/backend/internal/domain/finance"
	"github.com/beiwater/NewHaven/backend/internal/formula"
	openapi "github.com/beiwater/NewHaven/backend/internal/generated/openapi"
	"github.com/beiwater/NewHaven/backend/internal/storage"
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

	if _, err := s.companies.GetCompany(ctx, companyID); err != nil {
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

	// Issuing a bond does NOT credit the issuer up front. Capital is raised only
	// as buyers purchase units (see BuyBond, which transfers buyer -> issuer).
	// Crediting the full face value here minted money: an issuer could float a
	// huge bond, pocket amount*faceValue instantly, never sell a unit (so owe no
	// interest and never call), and keep the cash — an unbounded money exploit.
	if err := s.finance.CreateBond(ctx, b); err != nil {
		return nil, apperr.Internalf("create bond: %v", err)
	}

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

// BuyBond purchases a quantity of an existing bond on the secondary market.
func (s *Service) BuyBond(ctx context.Context, companyID int, bondID string, amount int) (map[string]any, error) {
	if amount <= 0 {
		return nil, apperr.BadRequest("amount must be positive")
	}

	b, err := s.finance.GetBond(ctx, bondID)
	if err != nil {
		return nil, apperr.NotFoundf("bond %s not found", bondID)
	}
	if b.Status != "active" {
		return nil, apperr.BadRequest("bond is not active")
	}
	if b.IssuerCompanyID == companyID {
		return nil, apperr.BadRequest("cannot buy your own bond")
	}
	// Cannot subscribe to more than the remaining unissued supply. Without this
	// cap a buyer could purchase unlimited units of a small issue, forcing the
	// issuer to pay interest (and, on call, principal) far beyond what they
	// floated.
	remaining := b.TotalQuantity - b.IssuedQuantity
	if amount > remaining {
		return nil, apperr.BadRequestf("only %d units remaining", remaining)
	}

	cost := b.FaceValue * float64(amount)
	// Atomic buyer -> issuer transfer: the buyer lends capital to the issuer.
	// (Previously the buyer's cash was destroyed and the issuer credited nothing,
	// because the issuer had already been paid in full at CreateBond time.)
	if err := s.companies.TransferMoney(ctx, companyID, b.IssuerCompanyID, cost); err != nil {
		if err == storage.ErrInsufficientFunds {
			return nil, apperr.BadRequest("insufficient funds")
		}
		return nil, apperr.Internalf("transfer funds: %v", err)
	}

	now := s.clock.Now().UTC()
	holding := &domainfinance.BondHolding{
		BondID:      bondID,
		CompanyID:   companyID,
		Quantity:    amount,
		PurchasedAt: now.Format(time.RFC3339),
	}
	if err := s.finance.CreateBondHolding(ctx, holding); err != nil {
		_ = s.companies.TransferMoney(ctx, b.IssuerCompanyID, companyID, cost) // rollback
		return nil, apperr.Internalf("create holding: %v", err)
	}

	b.IssuedQuantity += amount
	if err := s.finance.UpdateBond(ctx, b); err != nil {
		return nil, apperr.Internalf("update bond: %v", err)
	}

	nowStr := now.Format(time.RFC3339)
	s.finance.AppendLedgerEntry(ctx, &domainfinance.LedgerEntry{
		CompanyID: companyID,
		Kind:      "bond_buy", Amount: cost, Direction: "out",
		CreatedAt: nowStr,
		Metadata:  map[string]any{"bond_id": bondID, "quantity": amount},
	})
	// Mirror the capital raised on the issuer's ledger so their cashflow reflects
	// the incoming funds.
	s.finance.AppendLedgerEntry(ctx, &domainfinance.LedgerEntry{
		CompanyID: b.IssuerCompanyID,
		Kind:      "bond_issue", Amount: cost, Direction: "in",
		CreatedAt: nowStr,
		Metadata:  map[string]any{"bond_id": bondID, "quantity": amount, "buyer": companyID},
	})

	return map[string]any{"ok": true}, nil
}

// CallBond redeems a bond before maturity (issuer action).
func (s *Service) CallBond(ctx context.Context, companyID int, bondID string, amount int) (map[string]any, error) {
	if amount <= 0 {
		return nil, apperr.BadRequest("amount must be positive")
	}

	b, err := s.finance.GetBond(ctx, bondID)
	if err != nil {
		return nil, apperr.NotFoundf("bond %s not found", bondID)
	}
	if b.IssuerCompanyID != companyID {
		return nil, apperr.Forbidden("only the issuer can call this bond")
	}
	if b.Status != "active" {
		return nil, apperr.BadRequest("bond is not active")
	}

	now := s.clock.Now().UTC()
	holdings, err := s.finance.GetBondHoldings(ctx, bondID)
	if err != nil {
		return nil, apperr.Internalf("get holdings: %v", err)
	}

	// Compute the total principal the issuer must repay, and verify they can
	// afford it BEFORE paying anyone. Otherwise a partial call would repay some
	// holders, leave the issuer's balance negative, and desync the ledger.
	totalPayback := 0.0
	for _, h := range holdings {
		totalPayback += b.FaceValue * float64(h.Quantity)
	}
	if totalPayback > 0 {
		issuer, err := s.companies.GetCompany(ctx, companyID)
		if err != nil {
			return nil, apperr.NotFoundf("company %d not found", companyID)
		}
		if issuer.Money < totalPayback {
			return nil, apperr.BadRequest("insufficient funds to call bond")
		}
	}

	for _, h := range holdings {
		payback := b.FaceValue * float64(h.Quantity)
		// Atomic issuer -> holder repayment of principal.
		if err := s.companies.TransferMoney(ctx, companyID, h.CompanyID, payback); err != nil {
			continue
		}
		s.finance.AppendLedgerEntry(ctx, &domainfinance.LedgerEntry{
			CompanyID: h.CompanyID,
			Kind:      "bond_call", Amount: payback, Direction: "in",
			CreatedAt: now.Format(time.RFC3339),
		})
		s.finance.AppendLedgerEntry(ctx, &domainfinance.LedgerEntry{
			CompanyID: companyID,
			Kind:      "bond_call", Amount: payback, Direction: "out",
			CreatedAt: now.Format(time.RFC3339),
		})
	}

	b.Status = "called"
	if err := s.finance.UpdateBond(ctx, b); err != nil {
		return nil, apperr.Internalf("update bond: %v", err)
	}

	return map[string]any{"ok": true}, nil
}

// SettleBondInterest pays interest to all bond holders.
func (s *Service) SettleBondInterest(ctx context.Context) (map[string]any, error) {
	bonds, err := s.finance.GetActiveBonds(ctx)
	if err != nil {
		return nil, apperr.Internalf("get active bonds: %v", err)
	}

	now := s.clock.Now().UTC()
	settledCount := 0

	for _, b := range bonds {
		b := b
		// Skip if already settled within the last 24 hours.
		if b.LastSettledAt != "" {
			if t, err := time.Parse(time.RFC3339, b.LastSettledAt); err == nil && now.Sub(t) < 24*time.Hour {
				continue
			}
		}
		holdings, err := s.finance.GetBondHoldings(ctx, b.ID)
		if err != nil {
			continue
		}

		// The issuer must pay each holder exactly what that holder receives, so
		// interest is transferred atomically holder by holder and the issuer's
		// total expense is the *sum of the per-holder floored amounts*. Charging
		// the issuer a separately-floored total (Floor(fv*issued*rate)) did not
		// reconcile with the sum of holder credits and destroyed money on every
		// tick.
		for _, h := range holdings {
			interest := math.Floor(b.FaceValue * float64(h.Quantity) * b.InterestRate)
			if interest <= 0 {
				continue
			}
			if err := s.companies.TransferMoney(ctx, b.IssuerCompanyID, h.CompanyID, interest); err != nil {
				// Issuer can't cover this holder's coupon (or a company is
				// missing); skip this holder rather than minting money.
				continue
			}
			s.finance.AppendLedgerEntry(ctx, &domainfinance.LedgerEntry{
				CompanyID: h.CompanyID,
				Kind:      "bond_interest_income", Amount: interest, Direction: "in",
				CreatedAt: now.Format(time.RFC3339),
			})
			s.finance.AppendLedgerEntry(ctx, &domainfinance.LedgerEntry{
				CompanyID: b.IssuerCompanyID,
				Kind:      "bond_interest_expense", Amount: interest, Direction: "out",
				CreatedAt: now.Format(time.RFC3339),
			})
			settledCount++
		}

		// Persist the settlement timestamp for idempotency.
		b.LastSettledAt = now.Format(time.RFC3339)
		if err := s.finance.UpdateBond(ctx, &b); err != nil {
			continue
		}
	}

	return map[string]any{"ok": true, "settledCount": settledCount}, nil
}
