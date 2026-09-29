package formula_test

import (
	"testing"

	"github.com/beiwater/NewHaven/backend/internal/formula"
)

func TestCTOProductionMultiplier(t *testing.T) {
	if got := formula.CTOProductionMultiplier(0); got != 1 {
		t.Fatalf("zero CTO multiplier = %v, want 1", got)
	}
	if got := formula.CTOProductionMultiplier(50); got != 2 {
		t.Fatalf("50-point CTO multiplier = %v, want 2", got)
	}
	if got := formula.CTOProductionMultiplier(1_000); got != 3 {
		t.Fatalf("CTO multiplier cap = %v, want 3", got)
	}
}

func TestCMOSalesBonusPct(t *testing.T) {
	if got := formula.CMOSalesBonusPct(40); got != 20 {
		t.Fatalf("CMO bonus = %v, want 20", got)
	}
	if got := formula.CMOSalesBonusPct(1_000); got != 50 {
		t.Fatalf("CMO bonus cap = %v, want 50", got)
	}
}
