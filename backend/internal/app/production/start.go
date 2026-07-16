package production

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/beiwater/NewHaven/backend/internal/apperr"
	"github.com/beiwater/NewHaven/backend/internal/catalog"
	domain "github.com/beiwater/NewHaven/backend/internal/domain/company"
	proddmn "github.com/beiwater/NewHaven/backend/internal/domain/production"
	"github.com/beiwater/NewHaven/backend/internal/formula"
	openapi "github.com/beiwater/NewHaven/backend/internal/generated/openapi"
	"github.com/beiwater/NewHaven/backend/internal/storage"
)

// StartProduction starts a new production job for the given company.
// It validates the building/resource, deducts required input inventory,
// calculates duration, creates a running production job, and returns the result.
func (s *Service) StartProduction(ctx context.Context, companyID int, req *openapi.StartProductionRequest) (*openapi.StartProductionResponse, error) {
	if req == nil {
		return nil, apperr.BadRequest("request is required")
	}
	requestID, err := normalizeProductionRequestID(req.RequestId)
	if err != nil {
		return nil, err
	}
	quality := 0
	if req.Quality != nil {
		quality = *req.Quality
	}
	if !formula.ValidProductQuality(quality) {
		return nil, apperr.Validationf("quality must be between Q%d and Q%d", formula.MinProductQuality, formula.MaxProductQuality)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	company, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "company not found", err)
	}
	if requestID != "" {
		existing, err := s.production.GetJobByClientRequestID(ctx, companyID, requestID)
		if err != nil {
			return nil, apperr.Internalf("find idempotent production job: %v", err)
		}
		if existing != nil {
			if existing.Status == proddmn.StatusCancelled {
				return nil, apperr.Conflict("this production request was cancelled; start again with a new requestId")
			}
			if !sameStartProduction(existing, req) {
				return nil, apperr.Conflict("requestId was already used for a different production run")
			}
			return startProductionResponse(existing), nil
		}
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
	if building.UpgradeTargetLevel > 0 || building.UpgradeCompletesAt != "" {
		completesAt, parseErr := time.Parse(time.RFC3339, building.UpgradeCompletesAt)
		if parseErr == nil && !s.clock.Now().Before(completesAt) {
			if upgrades, ok := s.companies.(storage.BuildingUpgradeStorage); ok {
				if _, _, err := upgrades.CompleteBuildingUpgrade(ctx, companyID, building.ID, building.UpgradeTargetLevel, building.UpgradeCompletesAt); err != nil {
					return nil, apperr.Internalf("complete building upgrade: %v", err)
				}
				company, err = s.companies.GetCompany(ctx, companyID)
				if err != nil {
					return nil, apperr.WrapMsg(apperr.KindNotFound, "company not found", err)
				}
				building = nil
				for i, candidate := range company.Buildings {
					if candidate.ID == req.BuildingId {
						building = &company.Buildings[i]
						break
					}
				}
			}
		}
		if building == nil || building.UpgradeTargetLevel > 0 || building.UpgradeCompletesAt != "" {
			return nil, apperr.Conflict("this building is under construction; wait for the upgrade to finish")
		}
	}

	// A building is a single production line: it may produce only one item at a
	// time. A finished run remains assigned until the player collects or cancels
	// it, so input and output states cannot overlap invisibly.
	existingJobs, err := s.production.GetJobsByBuilding(ctx, building.ID)
	if err != nil {
		return nil, apperr.Internalf("load production jobs for building %s: %v", building.ID, err)
	}
	for _, job := range existingJobs {
		if job.CompanyID == companyID && job.Status != proddmn.StatusClaimed && job.Status != proddmn.StatusCancelled {
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
	if quality > 0 {
		researchState, err := s.research.GetResourceResearch(ctx, companyID, req.ResourceId)
		if err != nil {
			return nil, apperr.Internalf("load quality research: %v", err)
		}
		maxQuality := 0
		if researchState != nil {
			maxQuality = researchState.Level
		}
		if maxQuality > formula.MaxProductQuality {
			maxQuality = formula.MaxProductQuality
		}
		if quality > maxQuality {
			return nil, apperr.Conflict("Q%d is locked for this product; research it before starting production", quality)
		}
	}

	inputs := qualityProductionInputs(resEntry, req.ResourceId, req.Quantity, quality)

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

	effectiveMod := prodMod
	// A CTO is the only executive effect that changes a production run. The
	// assigned executive's Science skill is resolved server-side so a client
	// cannot invent a speed bonus or borrow another company's executive.
	ctoSkill := domain.ActiveExecutiveSkill(company.Executives, domain.ExecutivePositionCTO)
	effectiveMod *= formula.CTOProductionMultiplier(ctoSkill)

	duration := formula.DurationSeconds(req.Quantity, resEntry.ProducedPerHourRaw, level, effectiveMod)
	durationSeconds := int(duration)

	// Duration cap: 48h validation.
	if durationSeconds > MaxDurationSeconds {
		return nil, apperr.Validationf("production duration %ds exceeds maximum %ds; reduce quantity", durationSeconds, MaxDurationSeconds)
	}

	// Create the production job.
	now := s.clock.Now()
	job := &proddmn.ProductionJob{
		ID:              s.idgen.Next("prod"),
		ClientRequestID: requestID,
		CompanyID:       companyID,
		BuildingID:      req.BuildingId,
		ResourceID:      req.ResourceId,
		Quality:         quality,
		Quantity:        req.Quantity,
		TargetQuantity:  req.Quantity,
		ConsumedInputs:  inputs,
		StartedAt:       now,
		DurationSeconds: float64(durationSeconds),
		Status:          proddmn.StatusRunning,
	}

	if err := s.production.StartProductionJob(ctx, job, inputs); err != nil {
		if errors.Is(err, storage.ErrAlreadyExists) {
			if requestID != "" {
				existing, findErr := s.production.GetJobByClientRequestID(ctx, companyID, requestID)
				if findErr == nil && existing != nil && sameStartProduction(existing, req) {
					return startProductionResponse(existing), nil
				}
			}
			return nil, apperr.Conflict("this building is already producing; collect or cancel its current run first")
		}
		if errors.Is(err, storage.ErrInsufficientInventory) {
			return nil, apperr.InsufficientInventory(err.Error())
		}
		return nil, apperr.Internalf("create job: %v", err)
	}

	return startProductionResponse(job), nil
}

func normalizeProductionRequestID(value *string) (string, error) {
	if value == nil {
		return "", nil
	}
	requestID := strings.TrimSpace(*value)
	if requestID == "" {
		return "", apperr.BadRequest("requestId cannot be empty")
	}
	if len(requestID) > 128 {
		return "", apperr.BadRequest("requestId cannot exceed 128 characters")
	}
	return requestID, nil
}

func sameStartProduction(job *proddmn.ProductionJob, req *openapi.StartProductionRequest) bool {
	quality := 0
	if req.Quality != nil {
		quality = *req.Quality
	}
	return job.BuildingID == req.BuildingId &&
		job.ResourceID == req.ResourceId &&
		job.Quality == quality &&
		job.TargetQuantity == req.Quantity
}

func qualityProductionInputs(resource *catalog.ResourceEntry, resourceID, quantity, quality int) []proddmn.InventoryStack {
	if resource == nil || quantity <= 0 {
		return nil
	}
	inputQuality := 0
	multiplier := 1
	if quality > 0 {
		inputQuality = quality - 1
		multiplier = formula.QualityInputMultiplier
	}
	inputs := make([]proddmn.InventoryStack, 0, len(resource.ProducedFrom))
	for inputResourceID, amountPerUnit := range resource.ProducedFrom {
		needed := amountPerUnit * quantity * multiplier
		if needed > 0 {
			inputs = append(inputs, proddmn.InventoryStack{ResourceID: inputResourceID, Quality: inputQuality, Quantity: needed})
		}
	}
	// A raw resource has no upstream recipe. Higher-quality raw output is a
	// refinement run that converts two units of the previous quality into one.
	if quality > 0 && len(inputs) == 0 {
		inputs = append(inputs, proddmn.InventoryStack{ResourceID: resourceID, Quality: inputQuality, Quantity: quantity * formula.QualityInputMultiplier})
	}
	return inputs
}

func startProductionResponse(job *proddmn.ProductionJob) *openapi.StartProductionResponse {
	// Build response DTOs.
	jobID := job.ID
	buildingID := job.BuildingID
	resourceID := job.ResourceID
	quality := job.Quality
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
		Quality:         &quality,
		Quantity:        &quantity,
		TargetQuantity:  &targetQty,
		StartedAt:       &startedAt,
		DurationSeconds: &dur,
		ClaimedAmount:   &claimedAmt,
		ClaimableAmount: &claimableAmt,
		Status:          &status,
	}

	buildingID = job.BuildingID
	busy := true
	buildingStatus := &openapi.BuildingProductionStatus{
		Id:    &buildingID,
		Busy:  &busy,
		JobId: &jobID,
	}

	return &openapi.StartProductionResponse{
		Job:      jobDTO,
		Building: buildingStatus,
	}
}
