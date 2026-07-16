package formula_test

import (
	"testing"

	"github.com/beiwater/NewHaven/backend/internal/formula"
)

func TestRetailQualitySpeedMultiplier(t *testing.T) {
	cases := []struct {
		quality int
		want    float64
	}{
		{-1, 1.00},
		{0, 1.00},
		{1, 1.02},
		{11, 1.22},
		{12, 1.24},
		{13, 1.24},
	}
	for _, tc := range cases {
		if got := formula.RetailQualitySpeedMultiplier(tc.quality); got != tc.want {
			t.Fatalf("Q%d multiplier = %.2f, want %.2f", tc.quality, got, tc.want)
		}
	}
}
