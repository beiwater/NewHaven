package research_test

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"testing"

	appresearch "github.com/beiwater/NewHaven/backend/internal/app/research"
	"github.com/beiwater/NewHaven/backend/internal/catalog"
	"github.com/beiwater/NewHaven/backend/internal/config"
	"github.com/beiwater/NewHaven/backend/internal/domain/auth"
	"github.com/beiwater/NewHaven/backend/internal/domain/company"
	"github.com/beiwater/NewHaven/backend/internal/platform"
	"github.com/beiwater/NewHaven/backend/internal/storage/memory"
)

func newResearchService(store *memory.Store, resources map[int]*catalog.ResourceEntry) *appresearch.Service {
	return appresearch.NewService(
		store,
		store,
		store,
		resources,
		&config.GameConfig{ResearchBaseCost: 1000, ResearchCostGrowth: 1.2},
		platform.NewLogger(slog.Default()),
	)
}

func createResearchCompany(t *testing.T, store *memory.Store, playerID int, money float64) int {
	t.Helper()
	ctx := context.Background()
	if err := store.CreatePlayer(ctx, &auth.Player{ID: playerID, Username: fmt.Sprintf("researcher-%d", playerID)}); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateCompany(ctx, &company.Company{PlayerID: playerID, Name: "Research Company", Money: money}); err != nil {
		t.Fatal(err)
	}
	created, err := store.GetCompanyByPlayerID(ctx, playerID)
	if err != nil {
		t.Fatal(err)
	}
	return created.ID
}

func TestListResearch_DefaultsToQualityZeroAndUsesProductTierCost(t *testing.T) {
	t.Parallel()
	store := memory.New()
	service := newResearchService(store, map[int]*catalog.ResourceEntry{
		3:  {ID: 3, Name: "Flour", Tier: 2, ProducedPerHourRaw: 100},
		99: {ID: 99, Name: "Patent", IsResearch: true, ProducedPerHourRaw: 100},
	})
	companyID := createResearchCompany(t, store, 401, 10000)

	items, err := service.ListResearch(context.Background(), companyID)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 1 {
		t.Fatalf("research items = %d, want 1 product: %+v", len(items), items)
	}
	item := items[0]
	if item.ResourceID != 3 || item.MaxQuality != 0 || item.NextQuality != 1 || item.NextCost != 2000 {
		t.Fatalf("unexpected Q0 research state: %+v", item)
	}
	if item.SalesSpeedBonus != 0 || item.NextSalesSpeedPct != 2 {
		t.Fatalf("unexpected retail quality effect: %+v", item)
	}
}

func TestUnlockQuality_IsAtomicIdempotentAndCompanyScoped(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := memory.New()
	resources := map[int]*catalog.ResourceEntry{
		1: {ID: 1, Name: "Grain", Tier: 1, ProducedPerHourRaw: 100},
	}
	companyA := createResearchCompany(t, store, 402, 10000)
	companyB := createResearchCompany(t, store, 403, 10000)
	services := []*appresearch.Service{newResearchService(store, resources), newResearchService(store, resources)}

	var wg sync.WaitGroup
	errorsSeen := make(chan error, len(services))
	responses := make(chan bool, len(services))
	for _, service := range services {
		wg.Add(1)
		go func(service *appresearch.Service) {
			defer wg.Done()
			response, err := service.UnlockQuality(ctx, companyA, 1, 1)
			if err == nil {
				responses <- response.Charged
			}
			errorsSeen <- err
		}(service)
	}
	wg.Wait()
	close(errorsSeen)
	close(responses)
	for err := range errorsSeen {
		if err != nil {
			t.Fatalf("concurrent Q1 unlock: %v", err)
		}
	}
	charged := 0
	for didCharge := range responses {
		if didCharge {
			charged++
		}
	}
	if charged != 1 {
		t.Fatalf("charged responses = %d, want exactly 1", charged)
	}

	updatedA, _ := store.GetCompany(ctx, companyA)
	updatedB, _ := store.GetCompany(ctx, companyB)
	if updatedA.Money != 9000 {
		t.Fatalf("company A money = %g, want one $1000 charge", updatedA.Money)
	}
	if updatedB.Money != 10000 {
		t.Fatalf("company B money changed across account boundary: %g", updatedB.Money)
	}
	researchA, _ := store.GetResourceResearch(ctx, companyA, 1)
	researchB, _ := store.GetResourceResearch(ctx, companyB, 1)
	if researchA == nil || researchA.Level != 1 || researchB != nil {
		t.Fatalf("research crossed company boundary: A=%+v B=%+v", researchA, researchB)
	}
	entries, _ := store.GetLedgerEntries(ctx, companyA, 10)
	if len(entries) != 1 || entries[0].Kind != "quality_research" || entries[0].Amount != 1000 {
		t.Fatalf("quality research ledger = %+v", entries)
	}

	if _, err := services[0].UnlockQuality(ctx, companyB, 1, 2); err != appresearch.ErrResearchSequence {
		t.Fatalf("skipped Q1 error = %v, want sequence error", err)
	}
	updatedB, _ = store.GetCompany(ctx, companyB)
	if updatedB.Money != 10000 {
		t.Fatalf("failed skip charged company B: %g", updatedB.Money)
	}
}

func TestUnlockQuality_InsufficientFundsDoesNotAdvance(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := memory.New()
	service := newResearchService(store, map[int]*catalog.ResourceEntry{
		9: {ID: 9, Name: "Pizza", Tier: 4, ProducedPerHourRaw: 100},
	})
	companyID := createResearchCompany(t, store, 404, 100)

	if _, err := service.UnlockQuality(ctx, companyID, 9, 1); err != appresearch.ErrInsufficientFunds {
		t.Fatalf("UnlockQuality error = %v, want insufficient funds", err)
	}
	research, _ := store.GetResourceResearch(ctx, companyID, 9)
	company, _ := store.GetCompany(ctx, companyID)
	if research != nil || company.Money != 100 {
		t.Fatalf("failed research mutated state: research=%+v money=%g", research, company.Money)
	}
}
