package service

import (
	"testing"
	"time"

	"go-sim-api/internal/formula"
	"go-sim-api/internal/model"
)

func newBondTestService() *Service {
	s := newCoreTestService()
	s.mu.Lock()
	s.State.Bonds = []model.Bond{
		{ID: "bond-1", Amount: 100, Interest: 0.012, PurchasedAt: time.Now().UTC().Format(time.RFC3339),
			MissedPayments: 0, InterestCollected: 0.0, RatingWhenPurchased: "A",
			IssuerCompanyID: 900001, OwnerCompanyID: 1, RestructurePct: 0.0},
		{ID: "bond-2", Amount: 50, Interest: 0.015, PurchasedAt: time.Now().UTC().Format(time.RFC3339),
			MissedPayments: 0, InterestCollected: 0.0, RatingWhenPurchased: "A",
			IssuerCompanyID: 1, OwnerCompanyID: 900001, RestructurePct: 0.0},
	}
	s.mu.Unlock()
	return s
}

func TestBondMarketView(t *testing.T) {
	s := newBondTestService()
	view := s.BondMarketView()
	if len(view) == 0 {
		t.Fatal("expected bonds in market view")
	}
	if len(view) != 2 {
		t.Fatalf("expected 2 bonds, got %d", len(view))
	}
	for _, b := range view {
		if b["id"] == nil || b["amount"] == nil || b["interest"] == nil {
			t.Errorf("bond missing required fields: %v", b)
		}
	}
}

func TestBondRatingGroup(t *testing.T) {
	s := newBondTestService()
	tests := []struct {
		rating string
		want   string
	}{
		{"AAA", "AAA to AA"},
		{"AA+", "AAA to AA"},
		{"AA", "AAA to AA"},
		{"AA-", "AA- to A"},
		{"A+", "AA- to A"},
		{"A", "AA- to A"},
		{"A-", "A- to BBB"},
		{"BBB+", "A- to BBB"},
		{"BBB", "A- to BBB"},
		{"BBB-", "BBB- to BB"},
		{"BB+", "BBB- to BB"},
		{"BB", "BBB- to BB"},
		{"BB-", "BB- to B"},
		{"B+", "BB- to B"},
		{"B", "BB- to B"},
		{"B-", "B- to C"},
		{"C", "B- to C"},
		{"D", "D to D"},
		{"", "D to D"},
	}
	for _, tc := range tests {
		got := s.BondRatingGroup(tc.rating)
		if got != tc.want {
			t.Errorf("BondRatingGroup(%q) = %q, want %q", tc.rating, got, tc.want)
		}
	}
}

func TestSettleBondInterest(t *testing.T) {
	s := newBondTestService()
	s.mu.Lock()
	s.State.Companies[0].Money = 500000
	s.mu.Unlock()
	result := s.SettleBondInterest(s.State.Companies[0].ID)
	income, _ := result["dailyBondIncome"].(float64)
	expense, _ := result["dailyBondExpense"].(float64)
	if income <= 0 {
		t.Errorf("expected bond income > 0, got %.2f", income)
	}
	if expense <= 0 {
		t.Errorf("expected bond expense > 0, got %.2f", expense)
	}
	// bond-1: issuer=bot, owner=player → income
	// DailyBondInterest(100, 1.2) = floor(100*50*1.2) = 6000
	if income != 6000 {
		t.Errorf("expected income 6000, got %.0f", income)
	}
	// bond-2: issuer=player, owner=bot → expense
	// DailyBondInterest(50, 1.5) = floor(50*50*1.5) = 3750
	if expense != 3750 {
		t.Errorf("expected expense 3750, got %.0f", expense)
	}
}

func TestSettleBondInterestDefaultsWhenInsolvent(t *testing.T) {
	s := newBondTestService()
	// bond-1: issuer=bot, owner=player → income 6000
	// bond-2 (modified): issuer=player, owner=bot, amount=100, interest=0.02 → expense 10000
	// net = 6000 - 10000 = -4000, company has 1000 < 4000 → defaults
	s.mu.Lock()
	s.State.Bonds[1] = model.Bond{
		ID: "bond-2", Amount: 100, Interest: 0.02,
		PurchasedAt:    time.Now().UTC().Format(time.RFC3339),
		MissedPayments: 0, InterestCollected: 0.0, RatingWhenPurchased: "A",
		IssuerCompanyID: 1, OwnerCompanyID: 900001, RestructurePct: 0.0,
	}
	s.State.Companies[0].Money = 1000
	s.mu.Unlock()
	result := s.SettleBondInterest(s.State.Companies[0].ID)
	defaults, _ := result["defaults"].([]map[string]any)
	if len(defaults) == 0 {
		t.Fatal("expected defaults when company cannot afford interest")
	}
	found := false
	for _, d := range defaults {
		if d["id"] == "bond-2" {
			found = true
			missed, _ := d["missed_payments"].(int)
			if missed != 1 {
				t.Errorf("expected 1 missed payment, got %d", missed)
			}
			restr, _ := d["restructure_percentage"].(int)
			if restr != 10 {
				t.Errorf("expected restructure 10, got %d", restr)
			}
		}
	}
	if !found {
		t.Error("bond-2 should be in defaults")
	}
	if s.State.Companies[0].Money != 0 {
		t.Errorf("company money should be 0 after default, got %.0f", s.State.Companies[0].Money)
	}
}

func TestIssueOrAdjustBond(t *testing.T) {
	s := newBondTestService()
	s.mu.Lock()
	s.Cfg.Game.BondMinInterest = 0.5
	s.Cfg.Game.BondMaxInterest = 2.0
	initialMoney := s.State.Companies[0].Money
	s.mu.Unlock()

	result, err := s.IssueOrAdjustBond(s.State.Companies[0].ID, 10, 1.5)
	if err != nil {
		t.Fatalf("IssueOrAdjustBond() unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected bond result")
	}
	id, _ := result["id"].(string)
	if id == "" {
		t.Error("expected bond id")
	}
	s.mu.Lock()
	expectedAdd := float64(10) * formula.BondFaceValue
	if s.State.Companies[0].Money != initialMoney+expectedAdd {
		t.Errorf("money = %.0f, want %.0f", s.State.Companies[0].Money, initialMoney+expectedAdd)
	}
	s.mu.Unlock()
}

func TestIssueOrAdjustBondInvalidAmount(t *testing.T) {
	s := newBondTestService()
	_, err := s.IssueOrAdjustBond(s.State.Companies[0].ID, 0, 1.5)
	if err == nil {
		t.Fatal("expected error for amount <= 0")
	}
}

func TestIssueOrAdjustBondInvalidInterestRate(t *testing.T) {
	s := newBondTestService()
	_, err := s.IssueOrAdjustBond(s.State.Companies[0].ID, 10, 0.1)
	if err == nil {
		t.Fatal("expected error for interest below min")
	}
	_, err = s.IssueOrAdjustBond(s.State.Companies[0].ID, 10, 3.0)
	if err == nil {
		t.Fatal("expected error for interest above max")
	}
}

func TestBuyBond(t *testing.T) {
	s := newBondTestService()
	s.mu.Lock()
	s.State.Companies[0].Money = 500000
	s.mu.Unlock()

	_, err := s.BuyBond(s.State.Companies[0].ID, "bond-1", 1)
	if err != nil {
		t.Fatalf("BuyBond() unexpected error: %v", err)
	}
}

func TestBuyBondCannotBuyOwn(t *testing.T) {
	s := newBondTestService()
	s.mu.Lock()
	s.State.Companies[0].Money = 500000
	s.mu.Unlock()

	_, err := s.BuyBond(s.State.Companies[0].ID, "bond-2", 1)
	if err == nil {
		t.Fatal("expected error when buying own bond")
	}
}

func TestBuyBondInsufficientAmount(t *testing.T) {
	s := newBondTestService()
	s.mu.Lock()
	s.State.Companies[0].Money = 500000
	s.mu.Unlock()

	_, err := s.BuyBond(s.State.Companies[0].ID, "bond-1", 200)
	if err == nil {
		t.Fatal("expected error for insufficient bond amount")
	}
}

func TestBuyBondNotFound(t *testing.T) {
	s := newBondTestService()
	s.mu.Lock()
	s.State.Companies[0].Money = 500000
	s.mu.Unlock()

	_, err := s.BuyBond(s.State.Companies[0].ID, "nonexistent-bond", 1)
	if err == nil {
		t.Fatal("expected error for nonexistent bond")
	}
}

func TestCallBondNotFound(t *testing.T) {
	s := newBondTestService()
	s.mu.Lock()
	s.State.Companies[0].Money = 500000
	s.mu.Unlock()

	_, err := s.CallBond(s.State.Companies[0].ID, "nonexistent", 1)
	if err == nil {
		t.Fatal("expected error for nonexistent bond")
	}
}

func TestCallBondInvalidAmount(t *testing.T) {
	s := newBondTestService()
	s.mu.Lock()
	s.State.Companies[0].Money = 500000
	s.mu.Unlock()

	_, err := s.CallBond(s.State.Companies[0].ID, "bond-1", 0)
	if err == nil {
		t.Fatal("expected error for amount <= 0")
	}
}
