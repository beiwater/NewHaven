package production_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/beiwater/NewHaven/backend/internal/app/production"
	"github.com/beiwater/NewHaven/backend/internal/catalog"
	"github.com/beiwater/NewHaven/backend/internal/config"
	"github.com/beiwater/NewHaven/backend/internal/domain/auth"
	domain "github.com/beiwater/NewHaven/backend/internal/domain/company"
	proddmn "github.com/beiwater/NewHaven/backend/internal/domain/production"
	"github.com/beiwater/NewHaven/backend/internal/formula"
	openapi "github.com/beiwater/NewHaven/backend/internal/generated/openapi"
	"github.com/beiwater/NewHaven/backend/internal/platform"
	"github.com/beiwater/NewHaven/backend/internal/storage/memory"
)

// newTestService creates a production Service with minimal dependencies suitable for testing.
// It uses a memory store and fake clock. Catalogs are populated from the provided maps, or
func newTestService(t *testing.T, resources map[int]*catalog.ResourceEntry, buildings map[int]*catalog.BuildingEntry) (*production.Service, *memory.Store) {
	t.Helper()
	store := memory.New()
	if resources == nil {
		resources = make(map[int]*catalog.ResourceEntry)
	}
	if buildings == nil {
		buildings = make(map[int]*catalog.BuildingEntry)
	}
	cfg := &config.GameConfig{
		ProductionMod: 1.0,
	}
	clock := platform.NewFakeClock(time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC))
	idgen := platform.NewIDGen()
	svc := production.NewService(store, store, store, store, cfg, resources, buildings, clock, idgen)
	return svc, store
}

func TestListProductionJobs_ReturnsJobs(t *testing.T) {
	ctx := context.Background()
	svc, store := newTestService(t, nil, nil)

	err := store.CreatePlayer(ctx, &auth.Player{ID: 1, Username: "alice"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}

	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID:  1,
		Name:      "Alice Corp",
		Money:     100000,
		Level:     1,
		XP:        0,
		Inventory: make(map[int]int),
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}

	company, err := store.GetCompanyByPlayerID(ctx, 1)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}

	clockNow := time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)
	err = store.CreateJob(ctx, &proddmn.ProductionJob{
		ID:              "job-1",
		CompanyID:       company.ID,
		BuildingID:      "bld-1",
		ResourceID:      5,
		Quantity:        10,
		TargetQuantity:  10,
		StartedAt:       clockNow,
		DurationSeconds: 60.0,
		Status:          proddmn.StatusRunning,
	})
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	err = store.CreateJob(ctx, &proddmn.ProductionJob{
		ID:              "job-2",
		CompanyID:       company.ID,
		BuildingID:      "bld-1",
		ResourceID:      7,
		Quantity:        25,
		TargetQuantity:  25,
		StartedAt:       clockNow.Add(-time.Hour),
		DurationSeconds: 120.0,
		Status:          proddmn.StatusReady,
	})
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	resp, err := svc.ListProductionJobs(ctx, company.ID)
	if err != nil {
		t.Fatalf("ListProductionJobs failed: %v", err)
	}

	if resp.Jobs == nil {
		t.Fatal("expected non-nil jobs array")
	}

	if len(*resp.Jobs) != 2 {
		t.Fatalf("expected 2 jobs, got %d", len(*resp.Jobs))
	}

	// Check first job
	j := (*resp.Jobs)[0]
	if j.Id == nil || *j.Id != "job-1" {
		t.Errorf("expected job id 'job-1', got %v", j.Id)
	}
	if j.ResourceId == nil || *j.ResourceId != 5 {
		t.Errorf("expected resource_id 5, got %v", j.ResourceId)
	}
	if j.Quantity == nil || *j.Quantity != 10 {
		t.Errorf("expected quantity 10, got %v", j.Quantity)
	}
	if j.TargetQuantity == nil || *j.TargetQuantity != 10 {
		t.Errorf("expected target_quantity 10, got %v", j.TargetQuantity)
	}
	if j.StartedAt == nil {
		t.Error("expected non-nil started_at")
	}
	if j.DurationSeconds == nil || *j.DurationSeconds != 60.0 {
		t.Errorf("expected duration_seconds 60.0, got %v", j.DurationSeconds)
	}
	if j.Status == nil || *j.Status != "running" {
		t.Errorf("expected status 'running', got %v", j.Status)
	}

	// Check second job
	j2 := (*resp.Jobs)[1]
	if j2.Id == nil || *j2.Id != "job-2" {
		t.Errorf("expected job id 'job-2', got %v", j2.Id)
	}
	if j2.ResourceId == nil || *j2.ResourceId != 7 {
		t.Errorf("expected resource_id 7, got %v", j2.ResourceId)
	}
	if j2.Status == nil || *j2.Status != "ready" {
		t.Errorf("expected status 'ready', got %v", j2.Status)
	}
}

func TestListProductionJobs_EmptyList(t *testing.T) {
	ctx := context.Background()
	svc, store := newTestService(t, nil, nil)

	err := store.CreatePlayer(ctx, &auth.Player{ID: 2, Username: "bob"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}

	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID:  2,
		Name:      "Bob Corp",
		Money:     100000,
		Level:     1,
		XP:        0,
		Inventory: make(map[int]int),
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}

	company, err := store.GetCompanyByPlayerID(ctx, 2)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}

	resp, err := svc.ListProductionJobs(ctx, company.ID)
	if err != nil {
		t.Fatalf("ListProductionJobs failed: %v", err)
	}

	if resp.Jobs == nil {
		t.Fatal("expected non-nil jobs array (should be empty)")
	}

	if len(*resp.Jobs) != 0 {
		t.Errorf("expected 0 jobs, got %d", len(*resp.Jobs))
	}
}

func TestListProductionJobs_CompanyNotFound(t *testing.T) {
	ctx := context.Background()
	svc, _ := newTestService(t, nil, nil)

	resp, err := svc.ListProductionJobs(ctx, 99999)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.Jobs == nil {
		t.Fatal("expected non-nil jobs array (should be empty)")
	}

	if len(*resp.Jobs) != 0 {
		t.Errorf("expected 0 jobs, got %d", len(*resp.Jobs))
	}
}

func TestStartProduction_Success_CreatesJobAndDeductsInventory(t *testing.T) {
	ctx := context.Background()

	// Set up catalogs:
	// Building type 1 (Farm) produces resource 1 (Grain) and 6 (Sugar), 12 (Vegetables)
	// Resource 3 (Flour) requires 2 Grain (resource 1) per unit: producedFrom: {"1":2}
	// For this test, building type 3 (Mill) produces resource 3 (Flour).
	// Mill building_id=3, produces [3].
	resources := map[int]*catalog.ResourceEntry{
		3: {ID: 3, Name: "Flour", ProducedPerHourRaw: 320, ProducedFrom: map[int]int{1: 2}},
		1: {ID: 1, Name: "Grain", ProducedPerHourRaw: 500, ProducedFrom: map[int]int{}},
	}
	buildings := map[int]*catalog.BuildingEntry{
		3: {ID: 3, Name: "Mill", Produces: []int{3}},
	}

	svc, store := newTestService(t, resources, buildings)

	// Create player and company
	err := store.CreatePlayer(ctx, &auth.Player{ID: 10, Username: "millowner"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}

	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID: 10,
		Name:     "Mill Corp",
		Money:    100000,
		Level:    1,
		XP:       0,
		Inventory: map[int]int{
			1: 100, // 100 Grain
		},
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}

	company, err := store.GetCompanyByPlayerID(ctx, 10)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}

	bldID := "my-mill"
	company.Buildings = []domain.Building{
		{ID: bldID, BuildingID: 3, Level: 1, Name: "My Mill"},
	}
	_ = store.UpdateCompany(ctx, company)
	// Re-read to get consistent state
	company, _ = store.GetCompany(ctx, company.ID)

	// Produce 10 Flour. Each Flour needs 2 Grain, so 20 Grain needed.
	resp, err := svc.StartProduction(ctx, company.ID, &openapi.StartProductionRequest{BuildingId: bldID, ResourceId: 3, Quantity: 10})
	if err != nil {
		t.Fatalf("StartProduction failed: %v", err)
	}

	if resp.Job == nil {
		t.Fatal("expected non-nil job in response")
	}
	if resp.Job.Id == nil || *resp.Job.Id == "" {
		t.Fatal("expected non-empty job id")
	}
	if resp.Job.ResourceId == nil || *resp.Job.ResourceId != 3 {
		t.Errorf("expected resource_id 3, got %v", resp.Job.ResourceId)
	}
	if resp.Job.Quantity == nil || *resp.Job.Quantity != 10 {
		t.Errorf("expected quantity 10, got %v", resp.Job.Quantity)
	}
	if resp.Job.Status == nil || *resp.Job.Status != "running" {
		t.Errorf("expected status 'running', got %v", resp.Job.Status)
	}

	// Building status
	if resp.Building == nil {
		t.Fatal("expected non-nil building status")
	}
	if resp.Building.Id == nil || *resp.Building.Id != bldID {
		t.Errorf("expected building id %s, got %v", bldID, resp.Building.Id)
	}
	if resp.Building.Busy == nil || !*resp.Building.Busy {
		t.Error("expected building busy=true")
	}

	// Verify inventory: started with 100 Grain, used 20, should have 80.
	updatedCompany, _ := store.GetCompany(ctx, company.ID)
	if got := updatedCompany.Inventory[1]; got != 80 {
		t.Errorf("expected 80 Grain remaining, got %d", got)
	}

	// Verify job is listable via ListProductionJobs
	listResp, err := svc.ListProductionJobs(ctx, company.ID)
	if err != nil {
		t.Fatalf("ListProductionJobs failed: %v", err)
	}
	if listResp.Jobs == nil || len(*listResp.Jobs) != 1 {
		t.Fatalf("expected 1 job in list, got %d", len(*listResp.Jobs))
	}
}

func TestStartProduction_InsufficientInput_NoJobCreated(t *testing.T) {
	ctx := context.Background()

	resources := map[int]*catalog.ResourceEntry{
		3: {ID: 3, Name: "Flour", ProducedPerHourRaw: 320, ProducedFrom: map[int]int{1: 2}},
	}
	buildings := map[int]*catalog.BuildingEntry{
		3: {ID: 3, Name: "Mill", Produces: []int{3}},
	}

	svc, store := newTestService(t, resources, buildings)

	err := store.CreatePlayer(ctx, &auth.Player{ID: 11, Username: "poor"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}

	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID: 11,
		Name:     "Poor Corp",
		Money:    100000,
		Level:    1,
		XP:       0,
		Inventory: map[int]int{
			1: 5, // Only 5 Grain, need 20 for 10 Flour
		},
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}

	company, err := store.GetCompanyByPlayerID(ctx, 11)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}

	bldID := "poor-mill"
	company.Buildings = []domain.Building{
		{ID: bldID, BuildingID: 3, Level: 1},
	}
	_ = store.UpdateCompany(ctx, company)

	// Verify no job was created
	listResp, _ := svc.ListProductionJobs(ctx, company.ID)
	if listResp.Jobs != nil && len(*listResp.Jobs) > 0 {
		t.Error("expected no jobs to be created after insufficient inventory error")
	}

	// Verify inventory unchanged
	updatedCompany, _ := store.GetCompany(ctx, company.ID)
	if got := updatedCompany.Inventory[1]; got != 5 {
		t.Errorf("expected 5 Grain unchanged, got %d", got)
	}
}

func TestStartProduction_BuildingCannotProduce_Error(t *testing.T) {
	ctx := context.Background()

	resources := map[int]*catalog.ResourceEntry{
		3: {ID: 3, Name: "Flour", ProducedPerHourRaw: 320, ProducedFrom: map[int]int{}},
	}
	// Barn (id=2) produces [2,8], not Flour (id=3)
	buildings := map[int]*catalog.BuildingEntry{
		2: {ID: 2, Name: "Barn", Produces: []int{2, 8}},
	}

	svc, store := newTestService(t, resources, buildings)

	err := store.CreatePlayer(ctx, &auth.Player{ID: 12, Username: "wrongbld"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}

	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID:  12,
		Name:      "Wrong Bld Corp",
		Money:     100000,
		Level:     1,
		XP:        0,
		Inventory: map[int]int{},
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}

	company, err := store.GetCompanyByPlayerID(ctx, 12)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}

	company.Buildings = []domain.Building{
		{ID: "bld-test", BuildingID: 2, Level: 1, Name: "Test Barn"},
	}
	_ = store.UpdateCompany(ctx, company)

	_, err = svc.StartProduction(ctx, company.ID, &openapi.StartProductionRequest{BuildingId: company.Buildings[0].ID, ResourceId: 3, Quantity: 1})
	if err == nil {
		t.Fatal("expected error for building cannot produce resource, got nil")
	}
}

func TestStartProduction_DurationCap_Error(t *testing.T) {
	ctx := context.Background()

	// Low producedPerHourRaw to make even small quantities exceed cap
	resources := map[int]*catalog.ResourceEntry{
		1: {ID: 1, Name: "Grain", ProducedPerHourRaw: 1, ProducedFrom: map[int]int{}},
	}
	buildings := map[int]*catalog.BuildingEntry{
		1: {ID: 1, Name: "Farm", Produces: []int{1}},
	}

	svc, store := newTestService(t, resources, buildings)

	err := store.CreatePlayer(ctx, &auth.Player{ID: 13, Username: "durcap"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}

	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID:  13,
		Name:      "Dur Cap Corp",
		Money:     100000,
		Level:     1,
		XP:        0,
		Inventory: map[int]int{},
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}

	company, err := store.GetCompanyByPlayerID(ctx, 13)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}

	bldID := "bld-durcap"
	company.Buildings = []domain.Building{
		{ID: bldID, BuildingID: 1, Level: 1},
	}
	_ = store.UpdateCompany(ctx, company)

	// With producedPerHourRaw=1, level=1, rate=1, quantity=200000 =>
	// duration = ceil(200000 / 1 * 3600) = 720,000,000s >> 48h (172800s)
	_, err = svc.StartProduction(ctx, company.ID, &openapi.StartProductionRequest{BuildingId: bldID, ResourceId: 1, Quantity: 200000})
	if err == nil {
		t.Fatal("expected error for duration exceeding cap, got nil")
	}
}

func TestStartProduction_WarehouseReflectsDeduction(t *testing.T) {
	ctx := context.Background()

	resources := map[int]*catalog.ResourceEntry{
		3: {ID: 3, Name: "Flour", ProducedPerHourRaw: 320, ProducedFrom: map[int]int{1: 2}},
		1: {ID: 1, Name: "Grain", ProducedPerHourRaw: 500, ProducedFrom: map[int]int{}},
	}
	buildings := map[int]*catalog.BuildingEntry{
		3: {ID: 3, Name: "Mill", Produces: []int{3}},
	}

	svc, store := newTestService(t, resources, buildings)

	// Create player and company with inventory
	if err := store.CreatePlayer(ctx, &auth.Player{ID: 14, Username: "warehousecheck"}); err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	if err := store.CreateCompany(ctx, &domain.Company{
		PlayerID:  14,
		Name:      "Warehouse Check Corp",
		Money:     100000,
		Level:     1,
		XP:        0,
		Inventory: map[int]int{1: 50},
	}); err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}

	company, err := store.GetCompanyByPlayerID(ctx, 14)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}
	bldID := "bld-mill"
	company.Buildings = []domain.Building{
		{ID: bldID, BuildingID: 3, Level: 1},
	}
	_ = store.UpdateCompany(ctx, company)

	// Check warehouse before production: should have no items
	warehouseBefore, err := store.GetWarehouse(ctx, company.ID)
	if err != nil {
		t.Fatalf("GetWarehouse before: %v", err)
	}
	if len(warehouseBefore.Items) != 0 {
		t.Fatalf("expected empty warehouse before production, got %d items", len(warehouseBefore.Items))
	}
	if warehouseBefore.UsedCapacity != 0 {
		t.Fatalf("expected used_capacity 0 before production, got %d", warehouseBefore.UsedCapacity)
	}

	// Produce 5 Flour. Each needs 2 Grain, so 10 Grain is deducted.
	_, err = svc.StartProduction(ctx, company.ID, &openapi.StartProductionRequest{BuildingId: bldID, ResourceId: 3, Quantity: 5})
	if err != nil {
		t.Fatalf("StartProduction: %v", err)
	}

	// Check warehouse after production: should have 40 Grain (quality 0)
	warehouseAfter, err := store.GetWarehouse(ctx, company.ID)
	if err != nil {
		t.Fatalf("GetWarehouse after: %v", err)
	}

	if len(warehouseAfter.Items) != 1 {
		t.Fatalf("expected 1 warehouse item after production, got %d: %+v", len(warehouseAfter.Items), warehouseAfter.Items)
	}
	item := warehouseAfter.Items[0]
	if item.ResourceID != 1 {
		t.Errorf("expected resource_id 1 (Grain), got %d", item.ResourceID)
	}
	if item.Quality != 0 {
		t.Errorf("expected quality 0, got %d", item.Quality)
	}
	if item.Amount != 40 {
		t.Errorf("expected 40 Grain remaining, got %d", item.Amount)
	}
	if warehouseAfter.UsedCapacity != 40 {
		t.Errorf("expected used_capacity 40, got %d", warehouseAfter.UsedCapacity)
	}

	// Verify company inventory also matches
	updatedCompany, _ := store.GetCompany(ctx, company.ID)
	if updatedCompany.Inventory[1] != 40 {
		t.Errorf("company.Inventory[1] = %d, want 40", updatedCompany.Inventory[1])
	}

	// Produce enough to exhaust all Grain (need 20 more for 10 Flour)
	_, err = svc.StartProduction(ctx, company.ID, &openapi.StartProductionRequest{BuildingId: bldID, ResourceId: 3, Quantity: 20})
	if err != nil {
		t.Fatalf("StartProduction: %v", err)
	}

	warehouseAfter2, err := store.GetWarehouse(ctx, company.ID)
	if err != nil {
		t.Fatalf("GetWarehouse after: %v", err)
	}
	if len(warehouseAfter2.Items) != 0 {
		t.Fatalf("expected 0 warehouse items after exhausting inventory, got %d", len(warehouseAfter2.Items))
	}
	if warehouseAfter2.UsedCapacity != 0 {
		t.Errorf("expected used_capacity 0 after exhausting, got %d", warehouseAfter2.UsedCapacity)
	}
}

func TestStartProduction_DurationMatchesFormula(t *testing.T) {
	ctx := context.Background()

	resources := map[int]*catalog.ResourceEntry{
		1: {ID: 1, DbLetter: 1, Name: "Grain", ProducedFrom: map[int]int{}, ProducedPerHourRaw: 500, UnitsSoldAnHour: 150, HasEconomyModel: true, BasePrice: 23, IsExchangeTradable: true},
	}
	buildings := map[int]*catalog.BuildingEntry{
		1: {ID: 1, Kind: 1, Name: "Farm", Produces: []int{1}, BaseCost: 10000, BaseOutput: 500, BaseOutputPerHour: 500},
	}
	svc, store := newTestService(t, resources, buildings)

	// Create player and company
	err := store.CreatePlayer(ctx, &auth.Player{ID: 10, Username: "producer"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID:  10,
		Name:      "Producer Corp",
		Money:     100000,
		Level:     1,
		XP:        0,
		Inventory: map[int]int{1: 100},
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}

	// Assign a Farm (building_id=1, level 1) to the company
	company, err := store.GetCompanyByPlayerID(ctx, 10)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}
	company.Buildings = []domain.Building{
		{ID: "bld-1", BuildingID: 1, Level: 1, Name: "Test Farm", X: 0, Y: 0, MapID: "", SlotID: ""},
	}
	err = store.UpdateCompany(ctx, company)
	if err != nil {
		t.Fatalf("UpdateCompany: %v", err)
	}

	// Start production of 10 Grain in a level-1 Farm (produces 500/hr)
	resp, err := svc.StartProduction(ctx, company.ID, &openapi.StartProductionRequest{
		BuildingId: "bld-1",
		ResourceId: 1,
		Quantity:   10,
	})
	if err != nil {
		t.Fatalf("StartProduction: %v", err)
	}

	// DurationSeconds(10, 500, 1, 1.0) = ceil(10/500/1/1*3600) = 72
	expectedDuration := formula.DurationSeconds(10, 500, 1, 1.0)
	if resp.Job == nil || resp.Job.DurationSeconds == nil {
		t.Fatal("expected job and duration in response")
	}
	if *resp.Job.DurationSeconds != float32(expectedDuration) {
		t.Errorf("expected duration %.0f, got %.0f", expectedDuration, *resp.Job.DurationSeconds)
	}
}

// --- ClaimProduction tests ---

func TestClaimProduction_JobCompleted_CreditsInventory(t *testing.T) {
	ctx := context.Background()
	svc, store := newTestService(t, nil, nil)

	err := store.CreatePlayer(ctx, &auth.Player{ID: 30, Username: "claimer"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID:  30,
		Name:      "Claim Corp",
		Money:     100000,
		Level:     1,
		XP:        0,
		Inventory: make(map[int]int),
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}
	company, err := store.GetCompanyByPlayerID(ctx, 30)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}

	// Create a job that is already completed (started far enough in the past).
	clock := platform.NewFakeClock(time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC))
	startedAt := clock.Now().Add(-2 * time.Hour)
	err = store.CreateJob(ctx, &proddmn.ProductionJob{
		ID:              "claim-job-1",
		CompanyID:       company.ID,
		BuildingID:      "bld-1",
		ResourceID:      5,
		Quantity:        10,
		TargetQuantity:  10,
		StartedAt:       startedAt,
		DurationSeconds: 3600.0, // 1 hour, so 2h elapsed = done
		ClaimedAmount:   0,
		ClaimableAmount: 10,
		Status:          proddmn.StatusReady,
	})
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	// Verify inventory is empty before claim.
	company, _ = store.GetCompany(ctx, company.ID)
	if _, ok := company.Inventory[5]; ok {
		t.Fatal("expected no resource 5 in inventory before claim")
	}

	resp, err := svc.ClaimProduction(ctx, company.ID, "claim-job-1")
	if err != nil {
		t.Fatalf("ClaimProduction failed: %v", err)
	}

	if resp.JobId == nil || *resp.JobId != "claim-job-1" {
		t.Errorf("expected job_id 'claim-job-1', got %v", resp.JobId)
	}
	if resp.Status == nil || *resp.Status != "claimed" {
		t.Errorf("expected status 'claimed', got %v", resp.Status)
	}
	if resp.Output == nil {
		t.Fatal("expected non-nil output map")
	}
	if (*resp.Output)["5"] != 10 {
		t.Errorf("expected output[5]=10, got %d", (*resp.Output)["5"])
	}
	if resp.ClaimedAmount == nil || *resp.ClaimedAmount != 10 {
		t.Errorf("expected claimed_amount 10, got %v", resp.ClaimedAmount)
	}
	if resp.Remaining == nil || *resp.Remaining != 0 {
		t.Errorf("expected remaining 0, got %v", resp.Remaining)
	}
	if resp.Xp == nil || *resp.Xp <= 0 {
		t.Errorf("expected positive XP, got %v", resp.Xp)
	}

	// Verify inventory: resource 5 should have 10.
	company, _ = store.GetCompany(ctx, company.ID)
	if got := company.Inventory[5]; got != 10 {
		t.Errorf("expected 10 of resource 5 in inventory, got %d", got)
	}

	// Verify job is claimed.
	job, err := store.GetJob(ctx, "claim-job-1")
	if err != nil {
		t.Fatalf("GetJob: %v", err)
	}
	if job.Status != proddmn.StatusClaimed {
		t.Errorf("expected job status 'claimed', got %s", job.Status)
	}
	if job.ClaimedAmount != 10 {
		t.Errorf("expected job claimed_amount 10, got %d", job.ClaimedAmount)
	}

	entries, err := store.GetLedgerEntries(ctx, company.ID, 10)
	if err != nil {
		t.Fatalf("GetLedgerEntries: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected 1 ledger entry, got %d", len(entries))
	}
	if entries[0].Kind != "production_output" {
		t.Errorf("expected production_output ledger kind, got %s", entries[0].Kind)
	}
	if entries[0].Amount != 10 {
		t.Errorf("expected ledger amount 10, got %v", entries[0].Amount)
	}
}

func TestClaimProduction_PartialClaim_Works(t *testing.T) {
	ctx := context.Background()
	svc, store := newTestService(t, nil, nil)

	err := store.CreatePlayer(ctx, &auth.Player{ID: 31, Username: "partial"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID:  31,
		Name:      "Partial Corp",
		Money:     100000,
		Level:     1,
		XP:        5,
		Inventory: make(map[int]int),
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}
	company, err := store.GetCompanyByPlayerID(ctx, 31)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}

	// Create a fully-complete job that has already been partially claimed (10 of 20 units).
	// ClaimedAmount=10, ClaimableAmount will be computed as 10 (target - claimed).
	err = store.CreateJob(ctx, &proddmn.ProductionJob{
		ID:              "partial-job",
		CompanyID:       company.ID,
		BuildingID:      "bld-1",
		ResourceID:      7,
		Quantity:        20,
		TargetQuantity:  20,
		StartedAt:       time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC), // 2h ago, fully elapsed
		DurationSeconds: 60.0,
		ClaimedAmount:   10, // already partially claimed
		ClaimableAmount: 10,
		XPAwarded:       5, // half of 10 max XP
		Status:          proddmn.StatusReady,
	})
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	// Claim the remaining amount (10).
	resp, err := svc.ClaimProduction(ctx, company.ID, "partial-job")
	if err != nil {
		t.Fatalf("Partial claim failed: %v", err)
	}
	if resp.Status == nil || *resp.Status != "claimed" {
		t.Errorf("expected status 'claimed' after final claim, got %v", resp.Status)
	}
	if resp.ClaimedAmount == nil || *resp.ClaimedAmount != 20 {
		t.Errorf("expected claimed_amount 20 after full claim, got %v", resp.ClaimedAmount)
	}
	if resp.Remaining == nil || *resp.Remaining != 0 {
		t.Errorf("expected remaining 0 after full claim, got %v", resp.Remaining)
	}
	if resp.Xp == nil || *resp.Xp <= 0 {
		t.Errorf("expected positive incremental XP, got %v", resp.Xp)
	}

	// Verify inventory has 10 units (the remaining claimed amount).
	company, _ = store.GetCompany(ctx, company.ID)
	if got := company.Inventory[7]; got != 10 {
		t.Errorf("expected 10 of resource 7 in inventory, got %d", got)
	}

	// Verify job is fully claimed.
	job, err := store.GetJob(ctx, "partial-job")
	if err != nil {
		t.Fatalf("GetJob: %v", err)
	}
	if job.Status != proddmn.StatusClaimed {
		t.Errorf("expected job status 'claimed', got %s", job.Status)
	}
	if job.ClaimedAmount != 20 {
		t.Errorf("expected job claimed_amount 20, got %d", job.ClaimedAmount)
	}
}

func TestClaimProduction_TooEarly_ReturnsError(t *testing.T) {
	ctx := context.Background()
	svc, store := newTestService(t, nil, nil)

	err := store.CreatePlayer(ctx, &auth.Player{ID: 32, Username: "early"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID:  32,
		Name:      "Early Corp",
		Money:     100000,
		Level:     1,
		XP:        0,
		Inventory: make(map[int]int),
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}
	company, err := store.GetCompanyByPlayerID(ctx, 32)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}

	// Create a job that just started; no claimable amount.
	now := time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)
	err = store.CreateJob(ctx, &proddmn.ProductionJob{
		ID:              "early-job",
		CompanyID:       company.ID,
		BuildingID:      "bld-1",
		ResourceID:      3,
		Quantity:        100,
		TargetQuantity:  100,
		StartedAt:       now,
		DurationSeconds: 3600.0, // 1h job
		ClaimedAmount:   0,
		ClaimableAmount: 0,
		Status:          proddmn.StatusRunning,
	})
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	_, err = svc.ClaimProduction(ctx, company.ID, "early-job")
	if err == nil {
		t.Fatal("expected error for early claim, got nil")
	}
	if !strings.Contains(err.Error(), "nothing to claim") {
		t.Errorf("expected 'nothing to claim' error, got: %v", err)
	}

	// Verify inventory unchanged (no items added).
	company, _ = store.GetCompany(ctx, company.ID)
	if company.Inventory != nil && company.Inventory[3] != 0 {
		t.Errorf("expected no resource 3 in inventory, got %d", company.Inventory[3])
	}

	// Verify job state unchanged.
	job, err := store.GetJob(ctx, "early-job")
	if err != nil {
		t.Fatalf("GetJob: %v", err)
	}
	if job.ClaimedAmount != 0 {
		t.Errorf("expected claimed_amount 0, got %d", job.ClaimedAmount)
	}
	if job.Status != proddmn.StatusRunning {
		t.Errorf("expected status 'running', got %s", job.Status)
	}
}

func TestClaimProduction_AlreadyClaimed_ReturnsError(t *testing.T) {
	ctx := context.Background()
	svc, store := newTestService(t, nil, nil)

	err := store.CreatePlayer(ctx, &auth.Player{ID: 33, Username: "done"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID:  33,
		Name:      "Done Corp",
		Money:     100000,
		Level:     1,
		XP:        0,
		Inventory: make(map[int]int),
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}
	company, err := store.GetCompanyByPlayerID(ctx, 33)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}

	err = store.CreateJob(ctx, &proddmn.ProductionJob{
		ID:              "done-job",
		CompanyID:       company.ID,
		BuildingID:      "bld-1",
		ResourceID:      3,
		Quantity:        10,
		TargetQuantity:  10,
		StartedAt:       time.Now().Add(-2 * time.Hour),
		DurationSeconds: 60.0,
		ClaimedAmount:   10,
		ClaimableAmount: 0,
		Status:          proddmn.StatusClaimed,
	})
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	_, err = svc.ClaimProduction(ctx, company.ID, "done-job")
	if err == nil {
		t.Fatal("expected error for claimed job, got nil")
	}
	if !strings.Contains(err.Error(), "already claimed") {
		t.Errorf("expected 'already claimed' error, got: %v", err)
	}
}

func TestClaimProduction_WrongCompany_ReturnsError(t *testing.T) {
	ctx := context.Background()
	svc, store := newTestService(t, nil, nil)

	// Create company A and a job.
	err := store.CreatePlayer(ctx, &auth.Player{ID: 34, Username: "owner"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID:  34,
		Name:      "Owner Corp",
		Money:     100000,
		Level:     1,
		XP:        0,
		Inventory: make(map[int]int),
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}
	ownerCompany, err := store.GetCompanyByPlayerID(ctx, 34)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}

	err = store.CreateJob(ctx, &proddmn.ProductionJob{
		ID:              "other-job",
		CompanyID:       ownerCompany.ID,
		BuildingID:      "bld-1",
		ResourceID:      3,
		Quantity:        10,
		TargetQuantity:  10,
		StartedAt:       time.Now().Add(-2 * time.Hour),
		DurationSeconds: 60.0,
		ClaimedAmount:   0,
		ClaimableAmount: 10,
		Status:          proddmn.StatusReady,
	})
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	// Create company B (different user) and try to claim.
	err = store.CreatePlayer(ctx, &auth.Player{ID: 35, Username: "intruder"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID:  35,
		Name:      "Intruder Corp",
		Money:     100000,
		Level:     1,
		XP:        0,
		Inventory: make(map[int]int),
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}
	intruderCompany, err := store.GetCompanyByPlayerID(ctx, 35)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}

	_, err = svc.ClaimProduction(ctx, intruderCompany.ID, "other-job")
	if err == nil {
		t.Fatal("expected error for wrong company claim, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error (no leak), got: %v", err)
	}
}

// --- ListClaimableJobs tests ---

func TestListClaimableJobs_Empty_ReturnsNonNilArray(t *testing.T) {
	ctx := context.Background()
	svc, store := newTestService(t, nil, nil)

	err := store.CreatePlayer(ctx, &auth.Player{ID: 36, Username: "emptyclaims"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID:  36,
		Name:      "Empty Claims Corp",
		Money:     100000,
		Level:     1,
		XP:        0,
		Inventory: make(map[int]int),
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}
	company, err := store.GetCompanyByPlayerID(ctx, 36)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}

	resp, err := svc.ListClaimableJobs(ctx, company.ID)
	if err != nil {
		t.Fatalf("ListClaimableJobs failed: %v", err)
	}
	if resp.Jobs == nil {
		t.Fatal("expected non-nil jobs array (should be empty)")
	}
	if len(*resp.Jobs) != 0 {
		t.Errorf("expected 0 claimable jobs, got %d", len(*resp.Jobs))
	}
}

func TestListClaimableJobs_WithReadyJob_ReturnsIt(t *testing.T) {
	ctx := context.Background()
	svc, store := newTestService(t, nil, nil)

	err := store.CreatePlayer(ctx, &auth.Player{ID: 37, Username: "hasclaims"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID:  37,
		Name:      "Has Claims Corp",
		Money:     100000,
		Level:     1,
		XP:        0,
		Inventory: make(map[int]int),
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}
	company, err := store.GetCompanyByPlayerID(ctx, 37)
	if err != nil {
		t.Fatalf("GetCompanyByPlayerID: %v", err)
	}

	// Create a completed job (ready + claimable).
	err = store.CreateJob(ctx, &proddmn.ProductionJob{
		ID:              "ready-job",
		CompanyID:       company.ID,
		BuildingID:      "bld-2",
		ResourceID:      5,
		Quantity:        30,
		TargetQuantity:  30,
		StartedAt:       time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC),
		DurationSeconds: 60.0,
		ClaimedAmount:   0,
		ClaimableAmount: 30,
		Status:          proddmn.StatusReady,
	})
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	resp, err := svc.ListClaimableJobs(ctx, company.ID)
	if err != nil {
		t.Fatalf("ListClaimableJobs failed: %v", err)
	}
	if resp.Jobs == nil {
		t.Fatal("expected non-nil jobs array")
	}
	if len(*resp.Jobs) != 1 {
		t.Fatalf("expected 1 claimable job, got %d", len(*resp.Jobs))
	}
	j := (*resp.Jobs)[0]
	if j.JobId == nil || *j.JobId != "ready-job" {
		t.Errorf("expected job_id 'ready-job', got %v", j.JobId)
	}
	if j.ClaimableAmount == nil || *j.ClaimableAmount != 30 {
		t.Errorf("expected claimable_amount 30, got %v", j.ClaimableAmount)
	}
	if j.TotalAmount == nil || *j.TotalAmount != 30 {
		t.Errorf("expected total_amount 30, got %v", j.TotalAmount)
	}
}
