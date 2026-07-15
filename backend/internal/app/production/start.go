package production

import (
	"context"
	"fmt"
	"math"

	"github.com/beiwater/NewHaven/backend/internal/apperr"
	domain "github.com/beiwater/NewHaven/backend/internal/domain/company"
	proddmn "github.com/beiwater/NewHaven/backend/internal/domain/production"
	"github.com/beiwater/NewHaven/backend/internal/formula"
	openapi "github.com/beiwater/NewHaven/backend/internal/generated/openapi"
)

// StartProduction starts a new production job for the given company.
// It validates the building/resource, deducts required input inventory,
// calculates duration, creates a running production job, and returns the result.
func (s *Service) StartProduction(ctx context.Context, companyID int, req *openapi.StartProductionRequest) (*openapi.StartProductionResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "company not found", err)
	}

	// Validate building ownership.
	var building *domain.Building
	for i, b := range company.Buildings {
		if b.ID == req.BuildingId {
			building = &company.Buildings[i]
			break
		}
	}
	if building == nil {
		return nil, apperr.NotFoundf("building %s not found for company %d", req.BuildingId, companyID)
	}

	// A building is a single production line: it may produce only one item at a
	// time. A finished run remains assigned until the player collects or cancels
	// it, so input and output states cannot overlap invisibly.
	existingJobs, err := s.production.GetJobsByBuilding(ctx, building.ID)
	if err != nil {
		return nil, apperr.Internalf("load production jobs for building %s: %v", building.ID, err)
	}
	for _, job := range existingJobs {
		if job.CompanyID == companyID && job.Status != proddmn.StatusClaimed {
			return nil, apperr.Conflict("this building is already producing; collect or cancel its current run first")
		}
	}

	// Look up building type in static catalog.
	bldEntry, ok := s.buildings[building.BuildingID]
	if !ok {
		return nil, apperr.NotFoundf("building type %d not found in catalog", building.BuildingID)
	}

	// Verify building can produce the requested resource.
	canProduce := false
	for _, pid := range bldEntry.Produces {
		if pid == req.ResourceId {
			canProduce = true
			break
		}
	}
	if !canProduce {
		return nil, apperr.BadRequestf("building %q cannot produce resource %d", bldEntry.Name, req.ResourceId)
	}

	// Look up resource in static catalog.
	resEntry, ok := s.resources[req.ResourceId]
	if !ok {
		return nil, apperr.NotFoundf("resource %d not found in catalog", req.ResourceId)
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
	for _, inp := range inputs {
		var curAmt int
		if company.Inventory != nil {
			curAmt = company.Inventory[inp.resourceID]
		}
		if curAmt < inp.amount {
			return nil, apperr.InsufficientInventory(fmt.Sprintf("insufficient inventory: resource %d has %d, need %d", inp.resourceID, curAmt, inp.amount))
		}
	}

	// Calculate duration via governed formula.
	level := building.Level
	if level < 1 {
		level = 1
	}
	prodMod := s.cfg.ProductionMod
	if prodMod <= 0 {
		prodMod = 1.0
	}
	if resEntry.ProducedPerHourRaw <= 0 {
		return nil, apperr.Validationf("invalid production rate for resource %d", req.ResourceId)
	}

	// Apply research speed bonus (if any) to reduce production duration.
	researchLevel := 0
	if rr, err := s.research.GetResourceResearch(ctx, companyID, req.ResourceId); err == nil && rr != nil {
		researchLevel = rr.Level
	}
	effectiveMod := prodMod * formula.ResearchSpeedBonus(researchLevel)

	duration := formula.DurationSeconds(req.Quantity, resEntry.ProducedPerHourRaw, level, effectiveMod)
	durationSeconds := int(duration)

	// Duration cap: 48h validation.
	if durationSeconds > MaxDurationSeconds {
		return nil, apperr.Validationf("production duration %ds exceeds maximum %ds; reduce quantity", durationSeconds, MaxDurationSeconds)
	}

	// Deduct input resources from inventory via UpdateInventory (rejects negative final amounts).
	for _, inp := range inputs {
		if err := s.companies.UpdateInventory(ctx, companyID, inp.resourceID, -inp.amount); err != nil {
			return nil, apperr.Internalf("deduct input %d: %v", inp.resourceID, err)
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
		return nil, apperr.Internalf("create job: %v", err)
	}

	// Build response DTOs.
	jobID := job.ID
	buildingID := job.BuildingID
	resourceID := job.ResourceID
	quantity := job.Quantity
	targetQty := job.TargetQuantity
	startedAt := job.StartedAt
	dur := float32(job.DurationSeconds)
	claimedAmt := job.ClaimedAmount
	claimableAmt := job.ClaimableAmount
	status := openapi.ProductionJobDTOStatus(job.Status)

	jobDTO := &openapi.ProductionJobDTO{
		Id:              &jobID,
		BuildingId:      &buildingID,
		ResourceId:      &resourceID,
		Quantity:        &quantity,
		TargetQuantity:  &targetQty,
		StartedAt:       &startedAt,
		DurationSeconds: &dur,
		ClaimedAmount:   &claimedAmt,
		ClaimableAmount: &claimableAmt,
		Status:          &status,
	}

	buildingID = building.ID
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
