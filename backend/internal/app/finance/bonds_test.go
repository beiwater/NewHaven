package finance_test

import (
	"context"
	"math"
	"testing"
	"time"

	appfinance "github.com/beiwater/NewHaven/backend/internal/app/finance"
	"github.com/beiwater/NewHaven/backend/internal/config"
	"github.com/beiwater/NewHaven/backend/internal/platform"
	"github.com/beiwater/NewHaven/backend/internal/storage/memory"
)

// newTestSvcWithClock returns a service, store, and a FakeClock for time control.
func newTestSvcWithClock() (*appfinance.Service, *memory.Store, *platform.FakeClock) {
	store := memory.New()
	clock := platform.NewFakeClock(time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC))
	idgen := platform.NewIDGen()
	cfg := &config.GameConfig{BondFaceValue: 5000, BondMinInterest: 0.5, BondMaxInterest: 2.0}
	svc := appfinance.NewService(store, store, clock, idgen, cfg)
	return svc, store, clock
}

// bondInterest computes the interest paid per holder: Floor(FaceValue * qty * rate).
func bondInterest(qty int) float64 {
	return math.Floor(5000.0 * float64(qty) * 0.012)
}

// totalMoney sums every company's balance — the invariant a closed economy must
// preserve across bond operations (no minting, no destruction).
func totalMoney(t *testing.T, store *memory.Store) float64 {
	t.Helper()
	cos, err := store.GetAllCompanies(context.Background())
	if err != nil {
		t.Fatalf("GetAllCompanies: %v", err)
	}
	sum := 0.0
	for _, c := range cos {
		sum += c.Money
	}
	return sum
}

func money(t *testing.T, store *memory.Store, id int) float64 {
	t.Helper()
	c, err := store.GetCompany(context.Background(), id)
	if err != nil {
		t.Fatalf("GetCompany %d: %v", id, err)
	}
	return c.Money
}

// TestCreateBond_DoesNotMintMoney is the regression test for the money-mint
// exploit: issuing a bond must not credit the issuer anything up front.
func TestCreateBond_DoesNotMintMoney(t *testing.T) {
	ctx := context.Background()
	svc, store, _ := newTestSvcWithClock()
	issuerID := newTestCompany(t, store, 501, "mint-issuer", 1_000_000)

	before := money(t, store, issuerID)
	systemBefore := totalMoney(t, store)

	// Issue a large bond — under the old bug this would have minted 5000 * 5000
	// = 25,000,000 into the issuer's account instantly.
	if _, err := svc.CreateBond(ctx, issuerID, 5000, 1.2); err != nil {
		t.Fatalf("CreateBond: %v", err)
	}

	if got := money(t, store, issuerID); got != before {
		t.Errorf("issuer money changed on create: %g -> %g (must not mint)", before, got)
	}
	if got := totalMoney(t, store); got != systemBefore {
		t.Errorf("total money changed on create: %g -> %g", systemBefore, got)
	}
}

// TestBuyBond_TransfersBuyerToIssuer verifies capital is raised from the buyer
// (buyer debited, issuer credited by the same amount, system total unchanged).
func TestBuyBond_TransfersBuyerToIssuer(t *testing.T) {
	ctx := context.Background()
	svc, store, _ := newTestSvcWithClock()
	issuerID := newTestCompany(t, store, 601, "buy-issuer", 1_000_000)
	buyerID := newTestCompany(t, store, 602, "buy-buyer", 1_000_000)

	resp, err := svc.CreateBond(ctx, issuerID, 5, 1.2)
	if err != nil {
		t.Fatalf("CreateBond: %v", err)
	}
	bondID := *resp.Bond.Id

	issuerBefore := money(t, store, issuerID)
	buyerBefore := money(t, store, buyerID)
	systemBefore := totalMoney(t, store)

	if _, err := svc.BuyBond(ctx, buyerID, bondID, 2); err != nil {
		t.Fatalf("BuyBond: %v", err)
	}

	cost := 2 * 5000.0
	if got := money(t, store, buyerID); got != buyerBefore-cost {
		t.Errorf("buyer money = %g; want %g", got, buyerBefore-cost)
	}
	if got := money(t, store, issuerID); got != issuerBefore+cost {
		t.Errorf("issuer money = %g; want %g", got, issuerBefore+cost)
	}
	if got := totalMoney(t, store); got != systemBefore {
		t.Errorf("total money changed on buy: %g -> %g", systemBefore, got)
	}
}

// TestBuyBond_RejectsOversubscription verifies a buyer cannot purchase more
// units than remain unissued.
func TestBuyBond_RejectsOversubscription(t *testing.T) {
	ctx := context.Background()
	svc, store, _ := newTestSvcWithClock()
	issuerID := newTestCompany(t, store, 701, "over-issuer", 1_000_000)
	buyerID := newTestCompany(t, store, 702, "over-buyer", 100_000_000)

	resp, err := svc.CreateBond(ctx, issuerID, 3, 1.2)
	if err != nil {
		t.Fatalf("CreateBond: %v", err)
	}
	bondID := *resp.Bond.Id

	if _, err := svc.BuyBond(ctx, buyerID, bondID, 4); err == nil {
		t.Fatal("BuyBond(4) on a 3-unit issue should fail")
	}
	// Buying exactly the supply is fine; a further unit is then rejected.
	if _, err := svc.BuyBond(ctx, buyerID, bondID, 3); err != nil {
		t.Fatalf("BuyBond(3): %v", err)
	}
	if _, err := svc.BuyBond(ctx, buyerID, bondID, 1); err == nil {
		t.Fatal("BuyBond(1) after fully subscribed should fail")
	}
}

// TestBuyBond_RejectsSelfBuy verifies an issuer cannot buy their own bond.
func TestBuyBond_RejectsSelfBuy(t *testing.T) {
	ctx := context.Background()
	svc, store, _ := newTestSvcWithClock()
	issuerID := newTestCompany(t, store, 801, "self-issuer", 1_000_000)

	resp, err := svc.CreateBond(ctx, issuerID, 5, 1.2)
	if err != nil {
		t.Fatalf("CreateBond: %v", err)
	}
	if _, err := svc.BuyBond(ctx, issuerID, *resp.Bond.Id, 1); err == nil {
		t.Fatal("issuer buying own bond should fail")
	}
}

// TestSettleBondInterest_ConservesMoney verifies the issuer pays exactly what
// the holder receives and the system total is unchanged.
func TestSettleBondInterest_ConservesMoney(t *testing.T) {
	ctx := context.Background()
	svc, store, _ := newTestSvcWithClock()

	issuerID := newTestCompany(t, store, 101, "bond-issuer", 1_000_000)
	holderID := newTestCompany(t, store, 102, "bond-holder", 1_000_000)

	resp, err := svc.CreateBond(ctx, issuerID, 5, 1.2)
	if err != nil {
		t.Fatalf("CreateBond: %v", err)
	}
	bondID := *resp.Bond.Id

	if _, err := svc.BuyBond(ctx, holderID, bondID, 2); err != nil {
		t.Fatalf("BuyBond: %v", err)
	}

	holderStart := money(t, store, holderID)
	issuerStart := money(t, store, issuerID)
	systemBefore := totalMoney(t, store)

	expectedInterest := bondInterest(2) // Floor(5000 * 2 * 0.012) = 120
	result, err := svc.SettleBondInterest(ctx)
	if err != nil {
		t.Fatalf("SettleBondInterest: %v", err)
	}
	if got := result["settledCount"].(int); got != 1 {
		t.Errorf("settledCount = %d; want 1", got)
	}

	if got := money(t, store, holderID); got != holderStart+expectedInterest {
		t.Errorf("holder money = %g; want %g", got, holderStart+expectedInterest)
	}
	if got := money(t, store, issuerID); got != issuerStart-expectedInterest {
		t.Errorf("issuer money = %g; want %g", got, issuerStart-expectedInterest)
	}
	if got := totalMoney(t, store); got != systemBefore {
		t.Errorf("total money changed on settle: %g -> %g", systemBefore, got)
	}
}

// TestSettleBondInterest_Idempotent verifies that calling SettleBondInterest
// twice within 24 hours does not double-pay.
func TestSettleBondInterest_Idempotent(t *testing.T) {
	ctx := context.Background()
	svc, store, _ := newTestSvcWithClock()

	issuerID := newTestCompany(t, store, 201, "idem-issuer", 1_000_000)
	holderID := newTestCompany(t, store, 202, "idem-holder", 1_000_000)

	resp, err := svc.CreateBond(ctx, issuerID, 5, 1.2)
	if err != nil {
		t.Fatalf("CreateBond: %v", err)
	}
	if _, err := svc.BuyBond(ctx, holderID, *resp.Bond.Id, 2); err != nil {
		t.Fatalf("BuyBond: %v", err)
	}

	if _, err := svc.SettleBondInterest(ctx); err != nil {
		t.Fatalf("first SettleBondInterest: %v", err)
	}
	holderAfterFirst := money(t, store, holderID)
	issuerAfterFirst := money(t, store, issuerID)

	// Second settlement immediately (same clock time) — should be skipped.
	result, err := svc.SettleBondInterest(ctx)
	if err != nil {
		t.Fatalf("second SettleBondInterest: %v", err)
	}
	if got := result["settledCount"].(int); got != 0 {
		t.Errorf("second call settledCount = %d; want 0 (should skip)", got)
	}
	if got := money(t, store, holderID); got != holderAfterFirst {
		t.Errorf("holder money changed on second call: %g -> %g", holderAfterFirst, got)
	}
	if got := money(t, store, issuerID); got != issuerAfterFirst {
		t.Errorf("issuer money changed on second call: %g -> %g", issuerAfterFirst, got)
	}
}

// TestSettleBondInterest_MultipleHolders verifies each holder is paid correctly
// and the issuer's expense equals the sum of holder incomes (no rounding leak).
func TestSettleBondInterest_MultipleHolders(t *testing.T) {
	ctx := context.Background()
	svc, store, _ := newTestSvcWithClock()

	issuerID := newTestCompany(t, store, 301, "multi-issuer", 1_000_000)
	holderA := newTestCompany(t, store, 302, "multi-holder-a", 1_000_000)
	holderB := newTestCompany(t, store, 303, "multi-holder-b", 1_000_000)

	resp, err := svc.CreateBond(ctx, issuerID, 10, 1.2)
	if err != nil {
		t.Fatalf("CreateBond: %v", err)
	}
	bondID := *resp.Bond.Id

	if _, err := svc.BuyBond(ctx, holderA, bondID, 3); err != nil {
		t.Fatalf("BuyBond holderA: %v", err)
	}
	if _, err := svc.BuyBond(ctx, holderB, bondID, 4); err != nil {
		t.Fatalf("BuyBond holderB: %v", err)
	}

	holderAStart := money(t, store, holderA)
	holderBStart := money(t, store, holderB)
	issuerStart := money(t, store, issuerID)
	systemBefore := totalMoney(t, store)

	result, err := svc.SettleBondInterest(ctx)
	if err != nil {
		t.Fatalf("SettleBondInterest: %v", err)
	}
	if got := result["settledCount"].(int); got != 2 {
		t.Errorf("settledCount = %d; want 2", got)
	}

	interestA := bondInterest(3) // 180
	interestB := bondInterest(4) // 240
	if got := money(t, store, holderA); got != holderAStart+interestA {
		t.Errorf("holderA money = %g; want %g", got, holderAStart+interestA)
	}
	if got := money(t, store, holderB); got != holderBStart+interestB {
		t.Errorf("holderB money = %g; want %g", got, holderBStart+interestB)
	}
	// Issuer pays exactly interestA + interestB — not a separately-floored total.
	if got := money(t, store, issuerID); got != issuerStart-(interestA+interestB) {
		t.Errorf("issuer money = %g; want %g", got, issuerStart-(interestA+interestB))
	}
	if got := totalMoney(t, store); got != systemBefore {
		t.Errorf("total money changed on settle: %g -> %g", systemBefore, got)
	}
}

// TestSettleBondInterest_SkipsThenResettlesAfter24h verifies the 24h window.
func TestSettleBondInterest_SkipsThenResettlesAfter24h(t *testing.T) {
	ctx := context.Background()
	svc, store, clock := newTestSvcWithClock()

	issuerID := newTestCompany(t, store, 401, "skip-issuer", 1_000_000)
	holderID := newTestCompany(t, store, 402, "skip-holder", 1_000_000)

	resp, err := svc.CreateBond(ctx, issuerID, 5, 1.2)
	if err != nil {
		t.Fatalf("CreateBond: %v", err)
	}
	if _, err := svc.BuyBond(ctx, holderID, *resp.Bond.Id, 2); err != nil {
		t.Fatalf("BuyBond: %v", err)
	}
	holderStart := money(t, store, holderID)
	issuerStart := money(t, store, issuerID)

	result1, err := svc.SettleBondInterest(ctx)
	if err != nil {
		t.Fatalf("first SettleBondInterest: %v", err)
	}
	if result1["settledCount"].(int) != 1 {
		t.Fatalf("first call settledCount = %d; want 1", result1["settledCount"])
	}

	clock.Advance(25 * time.Hour)

	result2, err := svc.SettleBondInterest(ctx)
	if err != nil {
		t.Fatalf("second SettleBondInterest: %v", err)
	}
	if result2["settledCount"].(int) != 1 {
		t.Errorf("second call settledCount = %d; want 1 (should re-settle after 25h)", result2["settledCount"])
	}

	interest := bondInterest(2) // 120 per settlement
	if got := money(t, store, holderID); got != holderStart+2*interest {
		t.Errorf("holder money = %g; want %g (two payments)", got, holderStart+2*interest)
	}
	if got := money(t, store, issuerID); got != issuerStart-2*interest {
		t.Errorf("issuer money = %g; want %g (two payments)", got, issuerStart-2*interest)
	}
}

// TestCallBond_RepaysPrincipal verifies calling a bond returns principal to
// holders, conserves money, and requires the issuer to have the funds.
func TestCallBond_RepaysPrincipal(t *testing.T) {
	ctx := context.Background()
	svc, store, _ := newTestSvcWithClock()

	issuerID := newTestCompany(t, store, 901, "call-issuer", 1_000_000)
	holderID := newTestCompany(t, store, 902, "call-holder", 1_000_000)

	resp, err := svc.CreateBond(ctx, issuerID, 5, 1.2)
	if err != nil {
		t.Fatalf("CreateBond: %v", err)
	}
	bondID := *resp.Bond.Id
	if _, err := svc.BuyBond(ctx, holderID, bondID, 2); err != nil {
		t.Fatalf("BuyBond: %v", err)
	}

	holderBefore := money(t, store, holderID)
	issuerBefore := money(t, store, issuerID)
	systemBefore := totalMoney(t, store)

	if _, err := svc.CallBond(ctx, issuerID, bondID, 2); err != nil {
		t.Fatalf("CallBond: %v", err)
	}

	principal := 2 * 5000.0
	if got := money(t, store, holderID); got != holderBefore+principal {
		t.Errorf("holder money = %g; want %g", got, holderBefore+principal)
	}
	if got := money(t, store, issuerID); got != issuerBefore-principal {
		t.Errorf("issuer money = %g; want %g", got, issuerBefore-principal)
	}
	if got := totalMoney(t, store); got != systemBefore {
		t.Errorf("total money changed on call: %g -> %g", systemBefore, got)
	}
}

// TestCallBond_RequiresFunds verifies a bankrupt issuer cannot call a bond and
// no holder is partially repaid.
func TestCallBond_RequiresFunds(t *testing.T) {
	ctx := context.Background()
	svc, store, _ := newTestSvcWithClock()

	// Issuer has just enough to be flush after selling, then we drain it below
	// the principal owed by having a large issue bought.
	issuerID := newTestCompany(t, store, 111, "broke-issuer", 0)
	holderID := newTestCompany(t, store, 112, "flush-holder", 1_000_000)

	resp, err := svc.CreateBond(ctx, issuerID, 100, 1.2)
	if err != nil {
		t.Fatalf("CreateBond: %v", err)
	}
	bondID := *resp.Bond.Id
	// Holder buys 100 units => issuer receives 500,000.
	if _, err := svc.BuyBond(ctx, holderID, bondID, 100); err != nil {
		t.Fatalf("BuyBond: %v", err)
	}
	// Interest settlement drains a little, but the issuer still has ~500k while
	// the principal owed is also 500k. Spend the issuer's cash so a call cannot
	// be covered: transfer issuer -> holder via a second purchase would need a
	// bond; instead directly assert by settling interest first (issuer pays out,
	// dropping below principal owed).
	if _, err := svc.SettleBondInterest(ctx); err != nil {
		t.Fatalf("SettleBondInterest: %v", err)
	}

	holderBefore := money(t, store, holderID)
	if _, err := svc.CallBond(ctx, issuerID, bondID, 100); err == nil {
		t.Fatal("CallBond should fail when issuer cannot cover principal")
	}
	if got := money(t, store, holderID); got != holderBefore {
		t.Errorf("holder money changed on failed call: %g -> %g (must be atomic)", holderBefore, got)
	}
}
