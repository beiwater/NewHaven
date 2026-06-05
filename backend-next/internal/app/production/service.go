package production

import (
	"context"
	"fmt"
	"math"

	openapi "github.com/newhaven/backend-next/internal/generated/openapi"
	"github.com/newhaven/backend-next/internal/storage"

	"github.com/newhaven/backend-next/internal/catalog"
	"github.com/newhaven/backend-next/internal/config"
	domain "github.com/newhaven/backend-next/internal/domain/company"
	proddmn "github.com/newhaven/backend-next/internal/domain/production"
	"github.com/newhaven/backend-next/internal/platform"
)

// MaxDurationSeconds is the cap for any production job (48 hours).
const MaxDurationSeconds = 48 * 3600

// Service is the production application use case.
type Service struct {
	production storage.ProductionStorage
	companies  storage.CompanyStorage
	cfg        *config.GameConfig
	resources  map[int]*catalog.ResourceEntry
	buildings  map[int]*catalog.BuildingEntry
	clock      platform.Clock
	idgen      *platform.IDGen
}

// NewService creates a new production service.
func NewService(
	production storage.ProductionStorage,
	companies storage.CompanyStorage,
	cfg *config.GameConfig,
	resources map[int]*catalog.ResourceEntry,
	buildings map[int]*catalog.BuildingEntry,
	clock platform.Clock,
	idgen *platform.IDGen,
) *Service {
	return &Service{
		production: production,
		companies:  companies,
		cfg:        cfg,
		resources:  resources,
		buildings:  buildings,
		clock:      clock,
		idgen:      idgen,
	}
}

// ListProductionJobs returns all production jobs for the given company.
func (s *Service) ListProductionJobs(ctx context.Context, companyID int) (*openapi.ProductionJobListResponse, error) {
	jobs, err := s.production.GetJobsByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}

	dtos := make([]openapi.ProductionJobDTO, 0, len(jobs))
	for _, j := range jobs {
		id := j.ID
		resourceID := j.ResourceID
		quantity := j.Quantity
		targetQuantity := j.TargetQuantity
		startedAt := j.StartedAt
		durationSeconds := float32(j.DurationSeconds)
		status := openapi.ProductionJobDTOStatus(j.Status)

		dtos = append(dtos, openapi.ProductionJobDTO{
			Id:              &id,
			ResourceId:      &resourceID,
			Quantity:        &quantity,
			TargetQuantity:  &targetQuantity,
			StartedAt:       &startedAt,
			DurationSeconds: &durationSeconds,
			Status:          &status,
		})
	}

	return &openapi.ProductionJobListResponse{
		Jobs: &dtos,
	}, nil
}

// StartProduction starts a new production job for the given company.
// It validates the building/resource, deducts required input inventory,
// calculates duration, creates a running production job, and returns the result.
func (s *Service) StartProduction(ctx context.Context, companyID int, req *openapi.StartProductionRequest) (*openapi.StartProductionResponse, error) {
	company, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("company lookup: %w", err)
	}

	// Validate building ownership
	var building *domain.Building
	for i, b := range company.Buildings {
		if b.ID == req.BuildingId {
			building = &company.Buildings[i]
			break
		}
	}
	if building == nil {
		return nil, fmt.Errorf("building %s not found for company %d", req.BuildingId, companyID)
	}

	// Look up building type in static catalog
	bldEntry, ok := s.buildings[building.BuildingID]
	if !ok {
		return nil, fmt.Errorf("building type %d not found in catalog", building.BuildingID)
	}

	// Verify building can produce the requested resource
	canProduce := false
	for _, pid := range bldEntry.Produces {
		if pid == req.ResourceId {
			canProduce = true
			break
		}
	}
	if !canProduce {
		return nil, fmt.Errorf("building %q cannot produce resource %d", bldEntry.Name, req.ResourceId)
	}

	// Look up resource in static catalog
	resEntry, ok := s.resources[req.ResourceId]
	if !ok {
		return nil, fmt.Errorf("resource %d not found in catalog", req.ResourceId)
	}

	// Calculate required input amounts from producedFrom recipe.
	type inputReq struct {
		resourceID int
		amount     int
	}
	var inputs []inputReq
	for rid, amountPerUnit := range resEntry.ProducedFrom {
		needed := int(math.Ceil(float64(amountPerUnit) * float64(req.Quantity)))
		inputs = append(inputs, inputReq{resourceID: rid, amount: needed})
	}

	// Pre-check inventory for sufficient inputs.
	// All checks before any deduction keeps the operation atomic.
	// Treat nil Inventory as empty; do not mutate the company pointer.
	for _, inp := range inputs {
		var curAmt int
		if company.Inventory != nil {
			curAmt = company.Inventory[inp.resourceID]
		}
		if curAmt < inp.amount {
			return nil, fmt.Errorf("insufficient inventory: resource %d has %d, need %d",
				inp.resourceID, curAmt, inp.amount)
		}
	}

	// Calculate duration.
	// Formula: max(30, ceil(quantity / (producedPerHourRaw * max(level,1) * ProductionMod) * 3600))
	// ProductionMod from game.json: higher value = faster production (divisor).
	level := building.Level
	if level < 1 {
		level = 1
	}
	prodMod := s.cfg.ProductionMod
	if prodMod <= 0 {
		prodMod = 1.0
	}
	if resEntry.ProducedPerHourRaw <= 0 {
		return nil, fmt.Errorf("invalid production rate for resource %d", req.ResourceId)
	}
	rate := float64(resEntry.ProducedPerHourRaw) * float64(level) * prodMod
	duration := math.Ceil(float64(req.Quantity) / rate * 3600)
	if duration < 30 {
		duration = 30
	}
	durationSeconds := int(duration)

	// Duration cap: 48h validation.
	if durationSeconds > MaxDurationSeconds {
		return nil, fmt.Errorf("production duration %ds exceeds maximum %ds; reduce quantity", durationSeconds, MaxDurationSeconds)
	}

	// Deduct input resources from inventory via UpdateInventory (rejects negative final amounts).
	for _, inp := range inputs {
		if err := s.companies.UpdateInventory(ctx, companyID, inp.resourceID, -inp.amount); err != nil {
			return nil, fmt.Errorf("deduct input %d: %w", inp.resourceID, err)
		}
	}

	// Create the production job.
	now := s.clock.Now()
	job := &proddmn.ProductionJob{
		ID:              s.idgen.Next("prod"),
		CompanyID:       companyID,
		BuildingID:      req.BuildingId,
		ResourceID:      req.ResourceId,
		Quantity:        req.Quantity,
		TargetQuantity:  req.Quantity,
		StartedAt:       now,
		DurationSeconds: float64(durationSeconds),
		Status:          proddmn.StatusRunning,
	}

	if err := s.production.CreateJob(ctx, job); err != nil {
		// Rollback inventory deduction on job creation failure.
		for _, inp := range inputs {
			_ = s.companies.UpdateInventory(ctx, companyID, inp.resourceID, inp.amount)
		}
		return nil, fmt.Errorf("create job: %w", err)
	}

	// Build response DTOs.
	jobID := job.ID
	resourceID := job.ResourceID
	quantity := job.Quantity
	targetQty := job.TargetQuantity
	startedAt := job.StartedAt
	dur := float32(job.DurationSeconds)
	status := openapi.ProductionJobDTOStatus(job.Status)

	jobDTO := &openapi.ProductionJobDTO{
		Id:              &jobID,
		ResourceId:      &resourceID,
		Quantity:        &quantity,
		TargetQuantity:  &targetQty,
		StartedAt:       &startedAt,
		DurationSeconds: &dur,
		Status:          &status,
	}

	buildingID := building.ID
	busy := true
	buildingStatus := &openapi.BuildingProductionStatus{
		Id:    &buildingID,
		Busy:  &busy,
		JobId: &jobID,
	}

	return &openapi.StartProductionResponse{
		Job:      jobDTO,
		Building: buildingStatus,
	}, nil
}
