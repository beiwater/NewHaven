package finance_test

import (
	"context"
	"testing"
	"time"

	appfinance "github.com/newhaven/backend-next/internal/app/finance"
	domainauth "github.com/newhaven/backend-next/internal/domain/auth"
	domaincompany "github.com/newhaven/backend-next/internal/domain/company"
	domainfinance "github.com/newhaven/backend-next/internal/domain/finance"
	"github.com/newhaven/backend-next/internal/platform"
	"github.com/newhaven/backend-next/internal/storage/memory"
)

func newTestSvc() (*appfinance.Service, *memory.Store) {
	store := memory.New()
	clock := platform.NewFakeClock(time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC))
	svc := appfinance.NewService(store, store, clock)
	return svc, store
}

func newTestCompany(t *testing.T, store *memory.Store, playerID int, username string, money float64) int {
	t.Helper()
	err := store.CreatePlayer(nil, &domainauth.Player{ID: playerID, Username: username, PasswordHash: "hash"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	err = store.CreateCompany(nil, &domaincompany.Company{
		PlayerID:  playerID,
		Name:      username + " Corp",
		Money:     money,
		Level:     1,
		XP:        0,
		Inventory: make(map[int]int),
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}
	c, err := store.GetCompanyByPlayerID(nil, playerID)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}
	return c.ID
}

func addLedger(t *testing.T, store *memory.Store, cid int, kind string, direction string, amount float64, createdAt string) {
	t.Helper()
	err := store.AppendLedgerEntry(nil, &domainfinance.LedgerEntry{
		CompanyID: cid,
		Kind:      kind,
		Amount:    amount,
		Direction: direction,
		CreatedAt: createdAt,
	})
	if err != nil {
		t.Fatalf("AppendLedgerEntry: %v", err)
	}
}

func TestRecentCashflow_ConvertsDirection(t *testing.T) {
	ctx := context.Background()
	svc, store := newTestSvc()
	cid := newTestCompany(t, store, 1, "testco", 10000)

	addLedger(t, store, cid, "market_trade", "in", 500, "2026-06-07T11:00:00Z")
	addLedger(t, store, cid, "market_fee", "out", 20, "2026-06-07T11:01:00Z")
	addLedger(t, store, cid, "market_take_buy", "out", 200, "2026-06-07T11:02:00Z")

	resp, err := svc.GetRecentCashflow(ctx, cid)
	if err != nil {
		t.Fatalf("GetRecentCashflow: %v", err)
	}
	if resp.Data == nil || len(*resp.Data) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(*resp.Data))
	}

	e0 := (*resp.Data)[0]
	if e0.MoneyDelta == nil || *e0.MoneyDelta != -200 {
		t.Errorf("expected moneyDelta -200, got %v", e0.MoneyDelta)
	}
	e1 := (*resp.Data)[1]
	if e1.MoneyDelta == nil || *e1.MoneyDelta != -20 {
		t.Errorf("expected moneyDelta -20, got %v", e1.MoneyDelta)
	}
	e2 := (*resp.Data)[2]
	if e2.MoneyDelta == nil || *e2.MoneyDelta != 500 {
		t.Errorf("expected moneyDelta 500, got %v", e2.MoneyDelta)
	}

	if resp.Money == nil || *resp.Money != 10000 {
		t.Errorf("expected money 10000, got %v", resp.Money)
	}
	if resp.OldestPulled == nil || resp.OldestPulled.IsZero() {
		t.Errorf("expected non-zero oldestPulled")
	}
}

func TestRecentCashflow_EmptyLedger(t *testing.T) {
	ctx := context.Background()
	svc, store := newTestSvc()
	cid := newTestCompany(t, store, 2, "empty", 5000)

	resp, err := svc.GetRecentCashflow(ctx, cid)
	if err != nil {
		t.Fatalf("GetRecentCashflow: %v", err)
	}
	if resp.Data == nil || len(*resp.Data) != 0 {
		t.Errorf("expected empty data")
	}
	if resp.OldestPulled == nil || resp.OldestPulled.IsZero() {
		t.Errorf("expected non-zero oldestPulled")
	}
}

func TestIncomeStatement_AggregatesCorrectly(t *testing.T) {
	ctx := context.Background()
	svc, store := newTestSvc()
	cid := newTestCompany(t, store, 3, "inc", 0)

	addLedger(t, store, cid, "market_trade", "in", 1000, "")
	addLedger(t, store, cid, "market_trade", "in", 500, "")
	addLedger(t, store, cid, "market_fee", "out", 75, "")
	addLedger(t, store, cid, "buy_building", "out", 2000, "")

	resp, err := svc.GetIncomeStatement(ctx, cid)
	if err != nil {
		t.Fatalf("GetIncomeStatement: %v", err)
	}
	if resp.Revenue == nil || *resp.Revenue != 1500 {
		t.Errorf("expected revenue 1500, got %v", resp.Revenue)
	}
	if resp.Expenses == nil || *resp.Expenses != 2075 {
		t.Errorf("expected expenses 2075, got %v", resp.Expenses)
	}
	if resp.NetIncome == nil || *resp.NetIncome != -575 {
		t.Errorf("expected netIncome -575, got %v", resp.NetIncome)
	}
}

func TestBalanceSheet_UsesCompanyMoney(t *testing.T) {
	ctx := context.Background()
	svc, store := newTestSvc()
	cid := newTestCompany(t, store, 4, "bal", 25000)

	resp, err := svc.GetBalanceSheet(ctx, cid)
	if err != nil {
		t.Fatalf("GetBalanceSheet: %v", err)
	}
	if resp.Assets == nil || *resp.Assets != 25000 {
		t.Errorf("expected assets 25000, got %v", resp.Assets)
	}
	if resp.Liabilities == nil || *resp.Liabilities != 0 {
		t.Errorf("expected liabilities 0")
	}
	if resp.Equity == nil || *resp.Equity != 25000 {
		t.Errorf("expected equity 25000, got %v", resp.Equity)
	}
}

func TestCashflowStatement_Categorizes(t *testing.T) {
	ctx := context.Background()
	svc, store := newTestSvc()
	cid := newTestCompany(t, store, 5, "cf", 0)

	addLedger(t, store, cid, "market_trade", "in", 800, "")
	addLedger(t, store, cid, "market_fee", "out", 30, "")
	addLedger(t, store, cid, "buy_building", "out", 5000, "")
	addLedger(t, store, cid, "research_start", "out", 200, "")
	addLedger(t, store, cid, "bond_issue", "in", 10000, "")

	resp, err := svc.GetCashflowStatement(ctx, cid)
	if err != nil {
		t.Fatalf("GetCashflowStatement: %v", err)
	}
	if resp.Operating == nil || *resp.Operating != 770 {
		t.Errorf("expected operating 770, got %v", resp.Operating)
	}
	if resp.Investing == nil || *resp.Investing != -5200 {
		t.Errorf("expected investing -5200, got %v", resp.Investing)
	}
	if resp.Financing == nil || *resp.Financing != 10000 {
		t.Errorf("expected financing 10000, got %v", resp.Financing)
	}
}

func TestPastFinances_FallbackWhenNoTimestamps(t *testing.T) {
	ctx := context.Background()
	svc, store := newTestSvc()
	cid := newTestCompany(t, store, 6, "pf", 0)

	addLedger(t, store, cid, "market_trade", "in", 500, "")

	resp, err := svc.GetPastFinances(ctx, cid)
	if err != nil {
		t.Fatalf("GetPastFinances: %v", err)
	}
	if resp.Series == nil || len(*resp.Series) != 2 {
		t.Fatalf("expected 2 fallback series entries, got %d", len(*resp.Series))
	}
	if *(*resp.Series)[0].Date != "2026-05-28" {
		t.Errorf("expected first date 2026-05-28, got %v", *(*resp.Series)[0].Date)
	}
	if *(*resp.Series)[1].Net != float32(1022.4) {
		t.Errorf("expected second fallback net 1022.4, got %v", *(*resp.Series)[1].Net)
	}
}

func TestPastFinances_ComputesFromTimestamps(t *testing.T) {
	ctx := context.Background()
	svc, store := newTestSvc()
	cid := newTestCompany(t, store, 7, "pfts", 0)

	addLedger(t, store, cid, "market_trade", "in", 1000, "2026-06-01T12:00:00Z")
	addLedger(t, store, cid, "market_trade", "in", 500, "2026-06-01T14:00:00Z")
	addLedger(t, store, cid, "market_fee", "out", 40, "2026-06-02T10:00:00Z")

	resp, err := svc.GetPastFinances(ctx, cid)
	if err != nil {
		t.Fatalf("GetPastFinances: %v", err)
	}
	if resp.Series == nil || len(*resp.Series) != 2 {
		t.Fatalf("expected 2 series entries, got %d", len(*resp.Series))
	}
	if *(*resp.Series)[0].Net != 1500 {
		t.Errorf("expected day1 net 1500, got %v", *(*resp.Series)[0].Net)
	}
	if *(*resp.Series)[1].Net != -40 {
		t.Errorf("expected day2 net -40, got %v", *(*resp.Series)[1].Net)
	}
}
