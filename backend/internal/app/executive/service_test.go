package executive_test

import (
	"context"
	"sync"
	"testing"
	"time"

	executive "github.com/beiwater/NewHaven/backend/internal/app/executive"
	"github.com/beiwater/NewHaven/backend/internal/domain/company"
	"github.com/beiwater/NewHaven/backend/internal/platform"
	"github.com/beiwater/NewHaven/backend/internal/storage/memory"
)

func newExecutiveTestService(t *testing.T) (*executive.Service, *memory.Store, *platform.FakeClock) {
	t.Helper()
	clock := platform.NewFakeClock(time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC))
	store := memory.New()
	return executive.NewService(store, store, clock), store, clock
}

func createExecutiveTestCompany(t *testing.T, store *memory.Store, playerID int, money float64) int {
	t.Helper()
	c := &company.Company{PlayerID: playerID, Name: "executive test", Money: money, Inventory: map[int]int{}}
	if err := store.CreateCompany(context.Background(), c); err != nil {
		t.Fatalf("create company: %v", err)
	}
	return c.ID
}

func TestRecruitUsesExactCandidateAndScopesReplaysToCompany(t *testing.T) {
	svc, store, _ := newExecutiveTestService(t)
	companyA := createExecutiveTestCompany(t, store, 1, 1_000_000)
	companyB := createExecutiveTestCompany(t, store, 2, 1_000_000)
	candidate := svc.MarketCandidates()[3]

	hired, cost, err := svc.Recruit(context.Background(), companyA, candidate.ID)
	if err != nil {
		t.Fatalf("recruit candidate: %v", err)
	}
	if hired.ID != candidate.ID || hired.Specialty != candidate.Specialty || hired.Name != candidate.Name {
		t.Fatalf("recruit did not preserve selected candidate: got %#v want %#v", hired, candidate.Executive)
	}
	company, _ := store.GetCompany(context.Background(), companyA)
	if want := 1_000_000.0 - float64(cost); company.Money != want {
		t.Fatalf("company A money = %v, want %v", company.Money, want)
	}

	if _, _, err := svc.Recruit(context.Background(), companyA, candidate.ID); err == nil {
		t.Fatal("expected duplicate candidate replay to fail")
	}
	company, _ = store.GetCompany(context.Background(), companyA)
	if want := 1_000_000.0 - float64(cost); company.Money != want {
		t.Fatalf("duplicate replay changed money: got %v want %v", company.Money, want)
	}

	// Candidate IDs are intentionally scoped by the hiring company. Hiring the
	// same hourly offer in another account cannot change the first account.
	if _, _, err := svc.Recruit(context.Background(), companyB, candidate.ID); err != nil {
		t.Fatalf("second company recruit: %v", err)
	}
	company, _ = store.GetCompany(context.Background(), companyA)
	if len(company.Executives) != 1 {
		t.Fatalf("company A roster changed by company B: %d executives", len(company.Executives))
	}
}

func TestRecruitConcurrentReplayChargesOnce(t *testing.T) {
	svc, store, _ := newExecutiveTestService(t)
	companyID := createExecutiveTestCompany(t, store, 1, 1_000_000)
	candidate := svc.MarketCandidates()[0]

	var wg sync.WaitGroup
	results := make(chan error, 2)
	for range 2 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, _, err := svc.Recruit(context.Background(), companyID, candidate.ID)
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	successes := 0
	for err := range results {
		if err == nil {
			successes++
		}
	}
	if successes != 1 {
		t.Fatalf("concurrent recruitment successes = %d, want 1", successes)
	}
	company, _ := store.GetCompany(context.Background(), companyID)
	if len(company.Executives) != 1 {
		t.Fatalf("roster length = %d, want 1", len(company.Executives))
	}
	if want := 1_000_000.0 - float64(candidate.RecruitCost); company.Money != want {
		t.Fatalf("money = %v, want %v", company.Money, want)
	}
}

func TestTrainAndAssignmentAreCompanyIsolated(t *testing.T) {
	svc, store, _ := newExecutiveTestService(t)
	companyA := createExecutiveTestCompany(t, store, 1, 1_000_000)
	companyB := createExecutiveTestCompany(t, store, 2, 1_000_000)
	candidate := svc.MarketCandidates()[1]
	hired, _, err := svc.Recruit(context.Background(), companyA, candidate.ID)
	if err != nil {
		t.Fatalf("recruit: %v", err)
	}

	if _, err := svc.AssignPosition(context.Background(), companyB, hired.ID, company.ExecutivePositionCTO); err == nil {
		t.Fatal("expected cross-company assignment to be hidden")
	}
	beforeLevel, beforeSkills := hired.Level, hired.Skills
	updated, cost, err := svc.Train(context.Background(), companyA, hired.ID)
	if err != nil {
		t.Fatalf("train: %v", err)
	}
	if updated.Level != beforeLevel+1 {
		t.Fatalf("level = %d, want %d", updated.Level, beforeLevel+1)
	}
	if updated.Skills.Management < beforeSkills.Management+1 || updated.Skills.Science < beforeSkills.Science+1 {
		t.Fatalf("training did not improve broad skills: before=%+v after=%+v", beforeSkills, updated.Skills)
	}
	companyAfter, _ := store.GetCompany(context.Background(), companyA)
	if want := 1_000_000.0 - float64(candidate.RecruitCost) - float64(cost); companyAfter.Money != want {
		t.Fatalf("money = %v, want %v", companyAfter.Money, want)
	}
}
