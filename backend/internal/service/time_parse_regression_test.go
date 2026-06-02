package service

import (
	"strings"
	"testing"
	"time"

	"go-sim-api/internal/model"
)

func TestResearchProgressSkipsInvalidTimestamps(t *testing.T) {
	s := newCoreTestService()
	s.State.ResearchProjects = []model.ResearchProject{
		{
			ID:          "bad-research",
			Name:        "Broken Research",
			Status:      "in_progress",
			Progress:    25,
			StartedAt:   "not-a-time",
			CompletesAt: time.Now().UTC().Add(-time.Hour).Format(time.RFC3339),
		},
	}

	result := s.ResearchProgress()
	projects := result["projects"].([]model.ResearchProject)
	got := projects[0]
	if got.Status != "in_progress" {
		t.Fatalf("invalid timestamp should leave research in progress, got %q", got.Status)
	}
	if got.Progress != 25 {
		t.Fatalf("invalid timestamp should leave progress unchanged, got %d", got.Progress)
	}
}

func TestCalculateOfflineIncomeSkipsInvalidProductionJobTimes(t *testing.T) {
	s := newCoreTestService()
	companyID := s.State.Companies[0].ID
	startInventory := s.State.Companies[0].Inventory[3]
	now := time.Now().UTC()
	s.State.LastActiveAt = now.Add(-2 * time.Hour).Format(time.RFC3339)
	s.State.SimulatedAt = now.Format(time.RFC3339)
	s.State.Bonds = nil
	s.State.ProductionJobs = []model.ProductionJob{
		{
			ID:          "bad-job",
			ResourceID:  3,
			Amount:      10,
			Output:      map[int]int{3: 10},
			StartedAt:   "bad-start",
			CompletesAt: now.Add(-time.Hour).Format(time.RFC3339),
			Status:      "running",
		},
	}

	result := s.CalculateOfflineIncome(companyID)
	produced := result["produced"].(map[int]int)
	if len(produced) != 0 {
		t.Fatalf("invalid job timestamps should not produce items, got %v", produced)
	}
	if got := s.State.Companies[0].Inventory[3]; got != startInventory {
		t.Fatalf("inventory changed for invalid job: got %d, want %d", got, startInventory)
	}
}

func TestCallBondRejectsInvalidCallableAfter(t *testing.T) {
	s := newCoreTestService()
	companyID := s.State.Companies[0].ID
	s.State.Bonds = []model.Bond{
		{
			ID:              "bond-bad-time",
			Amount:          1,
			IssuerCompanyID: 999,
			OwnerCompanyID:  companyID,
			CallableAfter:   "not-a-time",
		},
	}

	_, err := s.CallBond(companyID, "bond-bad-time", 1)
	if err == nil {
		t.Fatal("expected invalid callableAfter error")
	}
	if !strings.Contains(err.Error(), "invalid callableAfter") {
		t.Fatalf("unexpected error: %v", err)
	}
}
