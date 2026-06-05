package production_test

import (
	"context"
	"testing"
	"time"

	"github.com/newhaven/backend-next/internal/app/production"
	"github.com/newhaven/backend-next/internal/domain/auth"
	domain "github.com/newhaven/backend-next/internal/domain/company"
	proddmn "github.com/newhaven/backend-next/internal/domain/production"
	"github.com/newhaven/backend-next/internal/storage/memory"
)

func TestListProductionJobs_ReturnsJobs(t *testing.T) {
	ctx := context.Background()
	store := memory.New()

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

	now := time.Now()
	err = store.CreateJob(ctx, &proddmn.ProductionJob{
		ID:              "job-1",
		CompanyID:       company.ID,
		BuildingID:      "bld-1",
		ResourceID:      5,
		Quantity:        10,
		TargetQuantity:  10,
		StartedAt:       now,
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
		StartedAt:       now.Add(-time.Hour),
		DurationSeconds: 120.0,
		Status:          proddmn.StatusReady,
	})
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	svc := production.NewService(store)

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
	store := memory.New()

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

	svc := production.NewService(store)

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
	store := memory.New()

	svc := production.NewService(store)

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
