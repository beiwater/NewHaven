package service

import (
	"testing"

	"go-sim-api/internal/model"
)

func TestPlaceGovernmentBid(t *testing.T) {
	s := newCoreTestService()
	s.mu.Lock()
	s.State.GovernmentContracts = []model.GovContract{
		{ID: "gov-1", ResourceID: 8, Quality: 0, Quantity: 500,
			MaxPrice: 12.4, DepositRate: 0.1, Status: "open",
			Bids: []map[string]any{}, WinnerCompanyID: 0},
	}
	s.State.Companies[0].Money = 50000
	s.mu.Unlock()

	result, err := s.PlaceGovernmentBid(s.State.Companies[0].ID, "gov-1", 11.5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
}

func TestPlaceGovernmentBid_NotOpen(t *testing.T) {
	s := newCoreTestService()
	s.State.GovernmentContracts = []model.GovContract{
		{ID: "gov-1", Status: "awarded", Bids: []map[string]any{}},
	}
	_, err := s.PlaceGovernmentBid(s.State.Companies[0].ID, "gov-1", 11.5)
	if err == nil {
		t.Fatal("expected error for closed contract")
	}
}

func TestAwardGovernmentContracts(t *testing.T) {
	s := newCoreTestService()
	s.mu.Lock()
	s.State.GovernmentContracts = []model.GovContract{
		{ID: "gov-1", ResourceID: 8, Quantity: 500, MaxPrice: 12.4,
			DepositRate: 0.1, Status: "open",
			Bids: []map[string]any{
				{"companyId": 1, "unitPrice": 11.0, "deposit": 5000.0},
				{"companyId": 900001, "unitPrice": 12.0, "deposit": 5000.0},
			},
			WinnerCompanyID: 0},
	}
	s.State.Companies[0].Money = 100000
	s.mu.Unlock()

	result := s.AwardGovernmentContracts()
	if len(result) == 0 {
		t.Fatal("expected awarded contracts")
	}
}
