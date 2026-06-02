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

func TestPlaceGovernmentBid_InvalidPrice(t *testing.T) {
	s := newCoreTestService()
	// The service rejects non-positive prices
	_, err := s.PlaceGovernmentBid(s.State.Companies[0].ID, "gov-1", 0)
	if err == nil {
		t.Fatal("expected error for zero price")
	}
}

func TestPlaceGovernmentBid_ContractNotFound(t *testing.T) {
	s := newCoreTestService()
	_, err := s.PlaceGovernmentBid(s.State.Companies[0].ID, "nonexistent", 10.0)
	if err == nil {
		t.Fatal("expected error for nonexistent contract")
	}
}

func TestDeliverGovernmentContract(t *testing.T) {
	s := newCoreTestService()
	s.mu.Lock()
	s.State.GovernmentContracts = []model.GovContract{
		{ID: "gov-1", ResourceID: 8, Quality: 0, Quantity: 500,
			MaxPrice: 12.4, DepositRate: 0.1, Status: "open",
			Bids: []map[string]any{
				{"companyId": 1, "unitPrice": 11.0, "deposit": 5000.0},
			},
			WinnerCompanyID: 0},
	}
	s.State.Companies[0].Inventory[8] = 2000
	s.mu.Unlock()

	// Award first
	result := s.AwardGovernmentContracts()
	if len(result) == 0 {
		t.Fatal("expected awarded contract")
	}

	if _, err := s.DeliverGovernmentContract(s.State.Companies[0].ID, "gov-1"); err != nil {
		t.Fatalf("DeliverGovernmentContract: %v", err)
	}
}

func TestDeliverGovernmentContract_InsufficientInventory(t *testing.T) {
	s := newCoreTestService()
	s.State.GovernmentContracts = []model.GovContract{
		{ID: "gov-1", ResourceID: 8, Quality: 0, Quantity: 500,
			MaxPrice: 12.4, DepositRate: 0.1, Status: "awarded",
			Bids: []map[string]any{
				{"companyId": 1, "unitPrice": 11.0, "deposit": 5000.0},
			},
			WinnerCompanyID: 1},
	}
	// Set inventory to 0
	s.State.Companies[0].Inventory[8] = 0

	_, err := s.DeliverGovernmentContract(s.State.Companies[0].ID, "gov-1")
	if err == nil {
		t.Fatal("expected error for insufficient inventory")
	}
}

func TestContractsByCompany(t *testing.T) {
	s := newCoreTestService()
	companyID := s.State.Companies[0].ID
	s.State.GovernmentContracts = []model.GovContract{
		{ID: "gov-1", Status: "awarded", WinnerCompanyID: companyID, Bids: []map[string]any{}},
		{ID: "gov-2", Status: "delivered", WinnerCompanyID: 900001, Bids: []map[string]any{}},
		{ID: "gov-3", Status: "open", WinnerCompanyID: companyID, Bids: []map[string]any{}},
	}
	contracts := s.ContractsByCompany(s.State.Companies[0].ID)
	if len(contracts) != 2 {
		t.Errorf("expected 2 contracts for company %d, got %d: %+v", companyID, len(contracts), contracts)
	}
}

func TestContractsByCompany_NoContracts(t *testing.T) {
	s := newCoreTestService()
	s.State.GovernmentContracts = nil
	contracts := s.ContractsByCompany(s.State.Companies[0].ID)
	if len(contracts) > 0 {
		t.Errorf("expected no contracts, got %v", contracts)
	}
}
