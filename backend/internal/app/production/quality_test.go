package production_test

import (
	"context"
	"testing"
	"time"

	"github.com/beiwater/NewHaven/backend/internal/app/production"
	"github.com/beiwater/NewHaven/backend/internal/apperr"
	"github.com/beiwater/NewHaven/backend/internal/catalog"
	"github.com/beiwater/NewHaven/backend/internal/config"
	"github.com/beiwater/NewHaven/backend/internal/domain/auth"
	domain "github.com/beiwater/NewHaven/backend/internal/domain/company"
	proddmn "github.com/beiwater/NewHaven/backend/internal/domain/production"
	openapi "github.com/beiwater/NewHaven/backend/internal/generated/openapi"
	"github.com/beiwater/NewHaven/backend/internal/platform"
	"github.com/beiwater/NewHaven/backend/internal/storage/memory"
)

func qualityWarehouseAmount(t *testing.T, store *memory.Store, companyID, resourceID, quality int) int {
	t.Helper()
	warehouse, err := store.GetWarehouse(context.Background(), companyID)
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range warehouse.Items {
		if item.ResourceID == resourceID && item.Quality == quality {
			return item.Amount
		}
	}
	return 0
}

func TestStartProduction_ReservesLowerQualityInputsPerCompany(t *testing.T) {
	ctx := context.Background()
	resources := map[int]*catalog.ResourceEntry{
		3: {ID: 3, Name: "Flour", ProducedPerHourRaw: 36, ProducedFrom: map[int]int{1: 2}},
	}
	buildings := map[int]*catalog.BuildingEntry{
		3: {ID: 3, Name: "Mill", Produces: []int{3}},
	}
	svc, store := newTestService(t, resources, buildings)

	createCompany := func(playerID int, buildingID string) int {
		if err := store.CreatePlayer(ctx, &auth.Player{ID: playerID, Username: buildingID}); err != nil {
			t.Fatal(err)
		}
		if err := store.CreateCompany(ctx, &domain.Company{
			PlayerID:  playerID,
			Name:      buildingID,
			Inventory: map[int]int{1: 100},
			Buildings: []domain.Building{{ID: buildingID, BuildingID: 3, Kind: 3, Level: 1}},
		}); err != nil {
			t.Fatal(err)
		}
		company, err := store.GetCompanyByPlayerID(ctx, playerID)
		if err != nil {
			t.Fatal(err)
		}
		return company.ID
	}
	companyA := createCompany(301, "quality-mill-a")
	companyB := createCompany(302, "quality-mill-b")
	quality := 1
	requestID := "quality-run"
	req := &openapi.StartProductionRequest{
		BuildingId: "quality-mill-a",
		ResourceId: 3,
		Quality:    &quality,
		Quantity:   5,
		RequestId:  &requestID,
	}
	resp, err := svc.StartProduction(ctx, companyA, req)
	if err != nil {
		t.Fatalf("StartProduction: %v", err)
	}
	if got := qualityWarehouseAmount(t, store, companyA, 1, 0); got != 80 {
		t.Fatalf("company A Q0 Grain = %d, want 80 after 2 recipe units × 2 quality multiplier × 5 output", got)
	}
	if got := qualityWarehouseAmount(t, store, companyB, 1, 0); got != 100 {
		t.Fatalf("company B Q0 Grain changed across account boundary: %d", got)
	}
	job, err := store.GetJob(ctx, *resp.Job.Id)
	if err != nil {
		t.Fatal(err)
	}
	if job.Quality != 1 || len(job.ConsumedInputs) != 1 || job.ConsumedInputs[0].Quality != 0 || job.ConsumedInputs[0].Quantity != 20 {
		t.Fatalf("quality reservation was not persisted: %+v", job)
	}

	changedQuality := 2
	changed := *req
	changed.Quality = &changedQuality
	if _, err := svc.StartProduction(ctx, companyA, &changed); !apperr.HasKind(err, apperr.KindConflict) {
		t.Fatalf("requestId reused with another quality error = %v, want conflict", err)
	}

	// Request IDs and reservations are company-scoped: another player may use
	// the same client-generated ID without seeing or replaying company A's job.
	companyBRequest := *req
	companyBRequest.BuildingId = "quality-mill-b"
	if _, err := svc.StartProduction(ctx, companyB, &companyBRequest); err != nil {
		t.Fatalf("company B should be able to use the same requestId: %v", err)
	}
	if got := qualityWarehouseAmount(t, store, companyB, 1, 0); got != 80 {
		t.Fatalf("company B Q0 Grain = %d, want its own 20-unit reservation", got)
	}
}

func TestClaimProduction_CreditsOnlyTheProducedQuality(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	clock := platform.NewFakeClock(time.Date(2026, 7, 16, 12, 0, 0, 0, time.UTC))
	svc := production.NewService(store, store, store, store, &config.GameConfig{ProductionMod: 1}, nil, nil, clock, platform.NewIDGen())
	if err := store.CreatePlayer(ctx, &auth.Player{ID: 303, Username: "quality-claimer"}); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateCompany(ctx, &domain.Company{
		PlayerID:  303,
		Name:      "Quality Claimer",
		Money:     100000,
		Buildings: []domain.Building{{ID: "quality-farm", BuildingID: 1, Kind: 1, Level: 1}},
	}); err != nil {
		t.Fatal(err)
	}
	company, _ := store.GetCompanyByPlayerID(ctx, 303)
	if err := store.CreateJob(ctx, &proddmn.ProductionJob{
		ID: "quality-output", CompanyID: company.ID, BuildingID: "quality-farm",
		ResourceID: 1, Quality: 7, Quantity: 10, TargetQuantity: 10,
		StartedAt: clock.Now().Add(-time.Hour), DurationSeconds: 60, Status: proddmn.StatusRunning,
	}); err != nil {
		t.Fatal(err)
	}
	resp, err := svc.ClaimProduction(ctx, company.ID, "quality-output")
	if err != nil {
		t.Fatalf("ClaimProduction: %v", err)
	}
	if resp.Quality == nil || *resp.Quality != 7 {
		t.Fatalf("claim quality = %v, want Q7", resp.Quality)
	}
	if got := qualityWarehouseAmount(t, store, company.ID, 1, 7); got != 10 {
		t.Fatalf("Q7 output = %d, want 10", got)
	}
	if got := qualityWarehouseAmount(t, store, company.ID, 1, 0); got != 0 {
		t.Fatalf("claim polluted Q0 inventory with %d units", got)
	}
}

func TestStartProduction_RefinesRawGoodsFromPreviousQuality(t *testing.T) {
	ctx := context.Background()
	resources := map[int]*catalog.ResourceEntry{
		1: {ID: 1, Name: "Grain", ProducedPerHourRaw: 60},
	}
	buildings := map[int]*catalog.BuildingEntry{
		1: {ID: 1, Name: "Farm", Produces: []int{1}},
	}
	svc, store := newTestService(t, resources, buildings)
	if err := store.CreatePlayer(ctx, &auth.Player{ID: 305, Username: "quality-refiner"}); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateCompany(ctx, &domain.Company{
		PlayerID:  305,
		Name:      "Quality Refiner",
		Buildings: []domain.Building{{ID: "quality-farm", BuildingID: 1, Kind: 1, Level: 1}},
	}); err != nil {
		t.Fatal(err)
	}
	company, _ := store.GetCompanyByPlayerID(ctx, 305)
	if err := store.UpdateInventoryQuality(ctx, company.ID, 1, 2, 20); err != nil {
		t.Fatal(err)
	}
	quality := 3
	resp, err := svc.StartProduction(ctx, company.ID, &openapi.StartProductionRequest{
		BuildingId: "quality-farm", ResourceId: 1, Quality: &quality, Quantity: 6,
	})
	if err != nil {
		t.Fatalf("StartProduction raw Q3: %v", err)
	}
	if got := qualityWarehouseAmount(t, store, company.ID, 1, 2); got != 8 {
		t.Fatalf("Q2 Grain = %d, want 8 after refining 12 units", got)
	}
	job, _ := store.GetJob(ctx, *resp.Job.Id)
	if len(job.ConsumedInputs) != 1 || job.ConsumedInputs[0] != (proddmn.InventoryStack{ResourceID: 1, Quality: 2, Quantity: 12}) {
		t.Fatalf("raw refinement reservation = %+v", job.ConsumedInputs)
	}
}

func TestCancelProduction_RefundsTheReservedInputQuality(t *testing.T) {
	ctx := context.Background()
	svc, store := newTestService(t, nil, nil)
	if err := store.CreatePlayer(ctx, &auth.Player{ID: 304, Username: "quality-cancel"}); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateCompany(ctx, &domain.Company{PlayerID: 304, Name: "Quality Cancel", Money: 100000}); err != nil {
		t.Fatal(err)
	}
	company, _ := store.GetCompanyByPlayerID(ctx, 304)
	if err := store.UpdateInventoryQuality(ctx, company.ID, 1, 4, 20); err != nil {
		t.Fatal(err)
	}
	if err := store.UpdateInventoryQuality(ctx, company.ID, 1, 4, -20); err != nil {
		t.Fatal(err)
	}
	if err := store.CreateJob(ctx, &proddmn.ProductionJob{
		ID: "quality-cancel-job", CompanyID: company.ID, ResourceID: 3, Quality: 5,
		Quantity: 5, TargetQuantity: 5, StartedAt: time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC),
		DurationSeconds: 3600, Status: proddmn.StatusRunning,
		ConsumedInputs: []proddmn.InventoryStack{{ResourceID: 1, Quality: 4, Quantity: 20}},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.CancelProductionJob(ctx, company.ID, "quality-cancel-job"); err != nil {
		t.Fatalf("CancelProductionJob: %v", err)
	}
	if got := qualityWarehouseAmount(t, store, company.ID, 1, 4); got != 10 {
		t.Fatalf("Q4 refund = %d, want 10", got)
	}
}
