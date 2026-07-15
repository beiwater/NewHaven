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

// bondInterest computes the interest paid by SettleBondInterest: Floor(FaceValue * qty * rate).
func bondInterest(qty int) float64 {
	return math.Floor(5000.0 * float64(qty) * 0.012)
}

// TestSettleBondInterest_ComputesCorrectInterest verifies the formula
// produces the expected interest amount.
func TestSettleBondInterest_ComputesCorrectInterest(t *testing.T) {
	ctx := context.Background()
	svc, store, _ := newTestSvcWithClock()

	// Issuer company; CreateBond also credits the issuer (amount * FaceValue).
	issuerID := newTestCompany(t, store, 101, "bond-issuer", 1_000_000)
	// Holder company starts with enough cash to buy bonds.
	holderID := newTestCompany(t, store, 102, "bond-holder", 1_000_000)

	// Create a bond: 5 units at 1.2% interest.
	// CreateBond stores interest as decimal: 1.2 / 100 = 0.012.
	bondResp, err := svc.CreateBond(ctx, issuerID, 5, 1.2)
	if err != nil {
		t.Fatalf("CreateBond: %v", err)
	}
	bondID := *bondResp.Bond.Id

	// Issuer gets credited 5 * 5000 = 25,000 on create.
	issuerAfterCreate, err := store.GetCompany(nil, issuerID)
	if err != nil {
		t.Fatalf("GetCompany issuer: %v", err)
	}
	issuerStartMoney := issuerAfterCreate.Money // 1,025,000

	// Holder buys 2 units of the bond (costs 2 * 5000 = 10,000).
	_, err = svc.BuyBond(ctx, holderID, bondID, 2)
	if err != nil {
		t.Fatalf("BuyBond: %v", err)
	}

	holderBefore, err := store.GetCompany(nil, holderID)
	if err != nil {
		t.Fatalf("GetCompany holder: %v", err)
	}
	holderStartMoney := holderBefore.Money // 1,000,000 - 10,000 = 990,000

	// Settle bond interest.
	expectedInterest := bondInterest(2) // Floor(5000 * 2 * 0.012) = 120
	result, err := svc.SettleBondInterest(ctx)
	if err != nil {
		t.Fatalf("SettleBondInterest: %v", err)
	}
	settledCount, ok := result["settledCount"].(int)
	if !ok {
		t.Fatal("settledCount missing or not int")
	}
	if settledCount != 1 {
		t.Errorf("settledCount = %d; want 1", settledCount)
	}

	// Verify holder received correct interest.
	holderAfter, err := store.GetCompany(nil, holderID)
	if err != nil {
		t.Fatalf("GetCompany holder: %v", err)
	}
	wantHolderMoney := holderStartMoney + expectedInterest
	if holderAfter.Money != wantHolderMoney {
		t.Errorf("holder money = %g; want %g", holderAfter.Money, wantHolderMoney)
	}

	// Verify issuer paid correct interest.
	issuerAfter, err := store.GetCompany(nil, issuerID)
	if err != nil {
		t.Fatalf("GetCompany issuer: %v", err)
	}
	wantIssuerMoney := issuerStartMoney - expectedInterest
	if issuerAfter.Money != wantIssuerMoney {
		t.Errorf("issuer money = %g; want %g", issuerAfter.Money, wantIssuerMoney)
	}

	// Verify holder got a ledger entry.
	holderLedger, err := store.GetLedgerEntries(nil, holderID, 100)
	if err != nil {
		t.Fatalf("GetLedgerEntries: %v", err)
	}
	foundIncome := false
	for _, e := range holderLedger {
		if e.Kind == "bond_interest_income" {
			foundIncome = true
			if e.Amount != expectedInterest {
				t.Errorf("holder ledger amount = %g; want %g", e.Amount, expectedInterest)
			}
			break
		}
	}
	if !foundIncome {
		t.Error("holder has no bond_interest_income ledger entry")
	}

	// Verify issuer got a ledger entry.
	issuerLedger, err := store.GetLedgerEntries(nil, issuerID, 100)
	if err != nil {
		t.Fatalf("GetLedgerEntries: %v", err)
	}
	foundExpense := false
	for _, e := range issuerLedger {
		if e.Kind == "bond_interest_expense" {
			foundExpense = true
			if e.Amount != expectedInterest {
				t.Errorf("issuer ledger amount = %g; want %g", e.Amount, expectedInterest)
			}
			break
		}
	}
	if !foundExpense {
		t.Error("issuer has no bond_interest_expense ledger entry")
	}
}

// TestSettleBondInterest_Idempotent verifies that calling SettleBondInterest
// twice within 24 hours does not double-pay.
func TestSettleBondInterest_Idempotent(t *testing.T) {
	ctx := context.Background()
	svc, store, _ := newTestSvcWithClock()

	issuerID := newTestCompany(t, store, 201, "idem-issuer", 1_000_000)
	holderID := newTestCompany(t, store, 202, "idem-holder", 1_000_000)

	bondResp, err := svc.CreateBond(ctx, issuerID, 5, 1.2)
	if err != nil {
		t.Fatalf("CreateBond: %v", err)
	}
	bondID := *bondResp.Bond.Id

	_, err = svc.BuyBond(ctx, holderID, bondID, 2)
	if err != nil {
		t.Fatalf("BuyBond: %v", err)
	}

	// First settlement.
	_, err = svc.SettleBondInterest(ctx)
	if err != nil {
		t.Fatalf("first SettleBondInterest: %v", err)
	}

	holderAfterFirst, err := store.GetCompany(nil, holderID)
	if err != nil {
		t.Fatalf("GetCompany: %v", err)
	}
	issuerAfterFirst, err := store.GetCompany(nil, issuerID)
	if err != nil {
		t.Fatalf("GetCompany: %v", err)
	}

	// Count ledger entries after first settlement.
	holderLedger1, _ := store.GetLedgerEntries(nil, holderID, 100)
	incomeCount1 := 0
	for _, e := range holderLedger1 {
		if e.Kind == "bond_interest_income" {
			incomeCount1++
		}
	}
	issuerLedger1, _ := store.GetLedgerEntries(nil, issuerID, 100)
	expenseCount1 := 0
	for _, e := range issuerLedger1 {
		if e.Kind == "bond_interest_expense" {
			expenseCount1++
		}
	}

	// Second settlement immediately (same clock time) — should be skipped.
	result, err := svc.SettleBondInterest(ctx)
	if err != nil {
		t.Fatalf("second SettleBondInterest: %v", err)
	}
	settledCount, ok := result["settledCount"].(int)
	if !ok {
		t.Fatal("settledCount missing or not int")
	}
	if settledCount != 0 {
		t.Errorf("second call settledCount = %d; want 0 (should skip)", settledCount)
	}

	// Money should be unchanged.
	holderAfterSecond, _ := store.GetCompany(nil, holderID)
	if holderAfterSecond.Money != holderAfterFirst.Money {
		t.Errorf("holder money changed on second call: %g → %g", holderAfterFirst.Money, holderAfterSecond.Money)
	}
	issuerAfterSecond, _ := store.GetCompany(nil, issuerID)
	if issuerAfterSecond.Money != issuerAfterFirst.Money {
		t.Errorf("issuer money changed on second call: %g → %g", issuerAfterFirst.Money, issuerAfterSecond.Money)
	}

	// No new ledger entries.
	holderLedger2, _ := store.GetLedgerEntries(nil, holderID, 100)
	incomeCount2 := 0
	for _, e := range holderLedger2 {
		if e.Kind == "bond_interest_income" {
			incomeCount2++
		}
	}
	if incomeCount2 != incomeCount1 {
		t.Errorf("holder income ledger entries increased: %d → %d", incomeCount1, incomeCount2)
	}
	issuerLedger2, _ := store.GetLedgerEntries(nil, issuerID, 100)
	expenseCount2 := 0
	for _, e := range issuerLedger2 {
		if e.Kind == "bond_interest_expense" {
			expenseCount2++
		}
	}
	if expenseCount2 != expenseCount1 {
		t.Errorf("issuer expense ledger entries increased: %d → %d", expenseCount1, expenseCount2)
	}
}

// TestSettleBondInterest_MultipleHolders verifies that interest is paid
// correctly to each holder and the issuer pays once.
func TestSettleBondInterest_MultipleHolders(t *testing.T) {
	ctx := context.Background()
	svc, store, _ := newTestSvcWithClock()

	issuerID := newTestCompany(t, store, 301, "multi-issuer", 1_000_000)
	holderA := newTestCompany(t, store, 302, "multi-holder-a", 1_000_000)
	holderB := newTestCompany(t, store, 303, "multi-holder-b", 1_000_000)

	bondResp, err := svc.CreateBond(ctx, issuerID, 10, 1.2)
	if err != nil {
		t.Fatalf("CreateBond: %v", err)
	}
	bondID := *bondResp.Bond.Id

	_, err = svc.BuyBond(ctx, holderA, bondID, 3)
	if err != nil {
		t.Fatalf("BuyBond holderA: %v", err)
	}
	_, err = svc.BuyBond(ctx, holderB, bondID, 4)
	if err != nil {
		t.Fatalf("BuyBond holderB: %v", err)
	}
	// 3 + 4 = 7 issued, the issuer still holds 3 (unsold).
	// Only holders who actually hold via BondHolding get payments.
	// Issuer pays on total IssuedQuantity (7).

	// Capture money as local float64 — GetCompany returns a pointer that
	// is mutated in-place by the service, so we must snapshot before settlement.
	holderABefore, _ := store.GetCompany(nil, holderA)
	holderAMoney := holderABefore.Money
	holderBBefore, _ := store.GetCompany(nil, holderB)
	holderBMoney := holderBBefore.Money
	issuerBefore, _ := store.GetCompany(nil, issuerID)
	issuerMoney := issuerBefore.Money

	result, err := svc.SettleBondInterest(ctx)
	if err != nil {
		t.Fatalf("SettleBondInterest: %v", err)
	}
	settledCount, _ := result["settledCount"].(int)
	// Two holders each with one holding = 2 settlements.
	if settledCount != 2 {
		t.Errorf("settledCount = %d; want 2", settledCount)
	}

	// Holder A has 3 units: receives Floor(5000 * 3 * 0.012) = 180.
	interestA := bondInterest(3)
	holderAAfter, _ := store.GetCompany(nil, holderA)
	wantA := holderAMoney + interestA
	if holderAAfter.Money != wantA {
		t.Errorf("holderA money = %g; want %g", holderAAfter.Money, wantA)
	}

	// Holder B has 4 units: receives Floor(5000 * 4 * 0.012) = 240.
	interestB := bondInterest(4)
	holderBAfter, _ := store.GetCompany(nil, holderB)
	wantB := holderBMoney + interestB
	if holderBAfter.Money != wantB {
		t.Errorf("holderB money = %g; want %g", holderBAfter.Money, wantB)
	}

	// Issuer pays on total 7 units: Floor(5000 * 7 * 0.012) = 420.
	totalInterest := bondInterest(7)
	issuerAfter, _ := store.GetCompany(nil, issuerID)
	wantIssuer := issuerMoney - totalInterest
	if issuerAfter.Money != wantIssuer {
		t.Errorf("issuer money = %g; want %g", issuerAfter.Money, wantIssuer)
	}
}

// TestSettleBondInterest_SkipsThenResettlesAfter24h verifies that a bond already
// settled within 24 hours is skipped, and a bond settled >24h ago is settled again.
func TestSettleBondInterest_SkipsThenResettlesAfter24h(t *testing.T) {
	ctx := context.Background()
	svc, store, clock := newTestSvcWithClock()

	issuerID := newTestCompany(t, store, 401, "skip-issuer", 1_000_000)
	holderID := newTestCompany(t, store, 402, "skip-holder", 1_000_000)

	bondResp, err := svc.CreateBond(ctx, issuerID, 5, 1.2)
	if err != nil {
		t.Fatalf("CreateBond: %v", err)
	}
	bondID := *bondResp.Bond.Id

	_, err = svc.BuyBond(ctx, holderID, bondID, 2)
	if err != nil {
		t.Fatalf("BuyBond: %v", err)
	}
	holder, _ := store.GetCompany(nil, holderID)
	holderStart := holder.Money
	issuer, _ := store.GetCompany(nil, issuerID)
	issuerStart := issuer.Money

	// First settlement.
	result1, err := svc.SettleBondInterest(ctx)
	if err != nil {
		t.Fatalf("first SettleBondInterest: %v", err)
	}
	if result1["settledCount"].(int) != 1 {
		t.Fatalf("first call settledCount = %d; want 1", result1["settledCount"])
	}

	// Advance clock by 25 hours — beyond the 24h idempotency window.
	clock.Advance(25 * time.Hour)

	// Second settlement — should process again.
	result2, err := svc.SettleBondInterest(ctx)
	if err != nil {
		t.Fatalf("second SettleBondInterest: %v", err)
	}
	if result2["settledCount"].(int) != 1 {
		t.Errorf("second call settledCount = %d; want 1 (should re-settle after 25h)", result2["settledCount"])
	}

	interest := bondInterest(2) // 120 per settlement
	holderAfter, _ := store.GetCompany(nil, holderID)
	issuerAfter, _ := store.GetCompany(nil, issuerID)

	// Both payments should have been applied (2 * 120 = 240).
	wantHolder := holderStart + 2*interest
	wantIssuer := issuerStart - 2*interest
	if holderAfter.Money != wantHolder {
		t.Errorf("holder money = %g; want %g (two payments)", holderAfter.Money, wantHolder)
	}
	if issuerAfter.Money != wantIssuer {
		t.Errorf("issuer money = %g; want %g (two payments)", issuerAfter.Money, wantIssuer)
	}
}
