package production_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/beiwater/NewHaven/backend/internal/app/production"
	"github.com/beiwater/NewHaven/backend/internal/catalog"
	"github.com/beiwater/NewHaven/backend/internal/config"
	"github.com/beiwater/NewHaven/backend/internal/domain/auth"
	domain "github.com/beiwater/NewHaven/backend/internal/domain/company"
	proddmn "github.com/beiwater/NewHaven/backend/internal/domain/production"
	"github.com/beiwater/NewHaven/backend/internal/platform"
)

func createProductionCompany(t *testing.T, store interface {
	CreatePlayer(context.Context, *auth.Player) error
	CreateCompany(context.Context, *domain.Company) error
	GetCompanyByPlayerID(context.Context, int) (*domain.Company, error)
}, playerID int, inventory map[int]int) *domain.Company {
	t.Helper()
	ctx := context.Background()
	if err := store.CreatePlayer(ctx, &auth.Player{ID: playerID, Username: "producer"}); err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	companyBuildings := []domain.Building{
		{ID: "bld-101-1", BuildingID: 1, Kind: 1, Name: "Bakery", Level: 1, MapID: "map_1", SlotID: "slot_a1", X: 5, Y: 10},
	}
	if err := store.CreateCompany(ctx, &domain.Company{
		PlayerID:  playerID,
		Name:      "Production Corp",
		Money:     100000,
		Level:     1,
		Inventory: inventory,
		Buildings: companyBuildings,
	}); err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}
	company, err := store.GetCompanyByPlayerID(ctx, playerID)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}
	return company
}

func TestProductionQueueRefreshesAndGroupsJobs(t *testing.T) {
	ctx := context.Background()
	svc, store := newTestService(t, nil, nil)
	company := createProductionCompany(t, store, 101, map[int]int{})

	if err := store.CreateJob(ctx, &proddmn.ProductionJob{
		ID:              "queue-job",
		CompanyID:       company.ID,
		BuildingID:      company.Buildings[0].ID,
		ResourceID:      1,
		Quantity:        10,
		TargetQuantity:  10,
		StartedAt:       time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC),
		DurationSeconds: 60,
		Status:          proddmn.StatusRunning,
	}); err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	resp, err := svc.ProductionQueue(ctx, company.ID)
	if err != nil {
		t.Fatalf("ProductionQueue: %v", err)
	}
	if resp.InUse == nil || *resp.InUse != 1 {
		t.Fatalf("inUse = %v, want 1", resp.InUse)
	}
	if resp.MaxSlots == nil || *resp.MaxSlots != len(company.Buildings) {
		t.Fatalf("maxSlots = %v, want %d", resp.MaxSlots, len(company.Buildings))
	}
	jobs := (*resp.ByBuilding)[company.Buildings[0].ID]
	if len(jobs) != 1 || jobs[0].Status == nil || *jobs[0].Status != "ready" {
		t.Fatalf("queue jobs = %+v, want one ready job", jobs)
	}
}

func TestCancelProductionRefundsHalfInputsOnceAndKeepsTombstone(t *testing.T) {
	ctx := context.Background()
	resources := map[int]*catalog.ResourceEntry{
		3: {ID: 3, ProducedFrom: map[int]int{1: 2}},
	}
	svc, store := newTestService(t, resources, nil)
	secondSvc := production.NewService(
		store,
		store,
		store,
		store,
		&config.GameConfig{ProductionMod: 1},
		resources,
		nil,
		platform.NewFakeClock(time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)),
		platform.NewIDGen(),
	)
	company := createProductionCompany(t, store, 102, map[int]int{1: 10})
	requestID := "cancel-original-start"

	if err := store.CreateJob(ctx, &proddmn.ProductionJob{
		ID:              "cancel-job",
		CompanyID:       company.ID,
		BuildingID:      company.Buildings[0].ID,
		ResourceID:      3,
		Quantity:        6,
		TargetQuantity:  6,
		StartedAt:       time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC),
		Status:          proddmn.StatusRunning,
		ClientRequestID: requestID,
	}); err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	services := []*production.Service{svc, secondSvc}
	errs := make([]error, len(services))
	var wg sync.WaitGroup
	for i := range services {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			resp, err := services[i].CancelProductionJob(ctx, company.ID, "cancel-job")
			errs[i] = err
			if err == nil && (resp.Status == nil || *resp.Status != "cancelled") {
				errs[i] = fmt.Errorf("status = %v, want cancelled", resp.Status)
			}
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("concurrent cancel %d: %v", i, err)
		}
	}
	company, _ = store.GetCompany(ctx, company.ID)
	if company.Inventory[1] != 16 {
		t.Fatalf("inventory = %d, want 16", company.Inventory[1])
	}
	job, err := store.GetJob(ctx, "cancel-job")
	if err != nil || job.Status != proddmn.StatusCancelled {
		t.Fatalf("cancel tombstone = %+v, err=%v", job, err)
	}
	byRequest, err := store.GetJobByClientRequestID(ctx, company.ID, requestID)
	if err != nil || byRequest == nil || byRequest.Status != proddmn.StatusCancelled {
		t.Fatalf("cancel request tombstone was not retained: job=%+v err=%v", byRequest, err)
	}
}

func TestClaimAllRefreshesNewlyCompletedJobs(t *testing.T) {
	ctx := context.Background()
	svc, store := newTestService(t, nil, nil)
	company := createProductionCompany(t, store, 103, map[int]int{})

	if err := store.CreateJob(ctx, &proddmn.ProductionJob{
		ID:              "claim-all-job",
		CompanyID:       company.ID,
		BuildingID:      company.Buildings[0].ID,
		ResourceID:      7,
		Quantity:        4,
		TargetQuantity:  4,
		StartedAt:       time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC),
		DurationSeconds: 60,
		Status:          proddmn.StatusRunning,
		ClaimableAmount: 0,
	}); err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	resp, err := svc.ClaimAll(ctx, company.ID)
	if err != nil {
		t.Fatalf("ClaimAll: %v", err)
	}
	if resp.Total == nil || *resp.Total != 1 {
		t.Fatalf("total = %v, want 1", resp.Total)
	}
	company, _ = store.GetCompany(ctx, company.ID)
	if company.Inventory[7] != 4 {
		t.Fatalf("inventory = %d, want 4", company.Inventory[7])
	}
}
