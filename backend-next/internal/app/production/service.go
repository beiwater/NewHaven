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
	"github.com/newhaven/backend-next/internal/domain/finance"
	proddmn "github.com/newhaven/backend-next/internal/domain/production"
	"github.com/newhaven/backend-next/internal/platform"
)

// MaxDurationSeconds is the cap for any production job (48 hours).
const MaxDurationSeconds = 48 * 3600

// Service is the production application use case.
type Service struct {
	production storage.ProductionStorage
	companies  storage.CompanyStorage
	finance    storage.FinanceStorage
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
	finance storage.FinanceStorage,
	cfg *config.GameConfig,
	resources map[int]*catalog.ResourceEntry,
	buildings map[int]*catalog.BuildingEntry,
	clock platform.Clock,
	idgen *platform.IDGen,
) *Service {
	return &Service{
		production: production,
		companies:  companies,
		finance:    finance,
		cfg:        cfg,
		resources:  resources,
		buildings:  buildings,
		clock:      clock,
		idgen:      idgen,
	}
}

// refreshJobStatuses recomputes Status, ClaimableAmount for all of a company's jobs
// based on the current clock time, and persists any changes.
func (s *Service) refreshJobStatuses(ctx context.Context, companyID int) error {
	jobs, err := s.production.GetJobsByCompany(ctx, companyID)
	if err != nil {
		return err
	}
	now := s.clock.Now()
	for i := range jobs {
		j := &jobs[i]
		if j.Status == proddmn.StatusClaimed {
			if j.ClaimableAmount != 0 {
				j.ClaimableAmount = 0
				if err := s.production.UpdateJob(ctx, j); err != nil {
					return fmt.Errorf("refresh claimed job %s: %w", j.ID, err)
				}
			}
			continue
		}
		// Compute claimable amount.
		elapsed := now.Sub(j.StartedAt).Seconds()
		totalSecs := j.DurationSeconds
		claimable := 0
		if totalSecs > 0 {
			if elapsed >= totalSecs {
				// Job is fully complete.
				claimable = j.TargetQuantity - j.ClaimedAmount
			} else if elapsed > 0 {
				// Partial completion.
				produced := int(math.Floor(elapsed / totalSecs * float64(j.TargetQuantity)))
				claimable = produced - j.ClaimedAmount
			}
		}
		if claimable < 0 {
			claimable = 0
		}
		j.ClaimableAmount = claimable

		if claimable > 0 && elapsed >= totalSecs {
			j.Status = proddmn.StatusReady
		} else if j.Status != proddmn.StatusRunning {
			j.Status = proddmn.StatusRunning
		}

		if err := s.production.UpdateJob(ctx, j); err != nil {
			return fmt.Errorf("refresh job %s: %w", j.ID, err)
		}
	}
	return nil
}

// ListProductionJobs returns all production jobs for the given company,
// with their status refreshed against the current clock.
func (s *Service) ListProductionJobs(ctx context.Context, companyID int) (*openapi.ProductionJobListResponse, error) {
	// Refresh computed fields before returning.
	if err := s.refreshJobStatuses(ctx, companyID); err != nil {
		return nil, err
	}

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
		claimedAmt := j.ClaimedAmount
		claimableAmt := j.ClaimableAmount
		status := openapi.ProductionJobDTOStatus(j.Status)

		dtos = append(dtos, openapi.ProductionJobDTO{
			Id:              &id,
			ResourceId:      &resourceID,
			Quantity:        &quantity,
			TargetQuantity:  &targetQuantity,
			StartedAt:       &startedAt,
			DurationSeconds: &durationSeconds,
			ClaimedAmount:   &claimedAmt,
			ClaimableAmount: &claimableAmt,
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
	claimedAmt := job.ClaimedAmount
	claimableAmt := job.ClaimableAmount
	status := openapi.ProductionJobDTOStatus(job.Status)

	jobDTO := &openapi.ProductionJobDTO{
		Id:              &jobID,
		ResourceId:      &resourceID,
		Quantity:        &quantity,
		TargetQuantity:  &targetQty,
		StartedAt:       &startedAt,
		DurationSeconds: &dur,
		ClaimedAmount:   &claimedAmt,
		ClaimableAmount: &claimableAmt,
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

// ClaimProduction claims available produced resources from a production job.
// It validates ownership, checks claimable amount, adds output to inventory,
// awards incremental XP, and appends a finance ledger entry.
func (s *Service) ClaimProduction(ctx context.Context, companyID int, jobID string) (*openapi.ClaimProductionResponse, error) {
	// Refresh job statuses so time-based computations are current.
	if err := s.refreshJobStatuses(ctx, companyID); err != nil {
		return nil, fmt.Errorf("refresh statuses: %w", err)
	}

	job, err := s.production.GetJob(ctx, jobID)
	if err != nil {
		return nil, fmt.Errorf("job not found")
	}
	if job.CompanyID != companyID {
		// Do not leak existence of other companies' jobs.
		return nil, fmt.Errorf("job not found")
	}
	if job.Status == proddmn.StatusClaimed {
		return nil, fmt.Errorf("job already claimed")
	}
	if job.ClaimableAmount <= 0 {
		return nil, fmt.Errorf("nothing to claim yet")
	}

	claimAmount := job.ClaimableAmount

	// Add produced resource to inventory.
	if err := s.companies.UpdateInventory(ctx, companyID, job.ResourceID, claimAmount); err != nil {
		return nil, fmt.Errorf("add output to inventory: %w", err)
	}

	// Update job fields.
	job.ClaimedAmount += claimAmount
	if job.ClaimedAmount >= job.TargetQuantity {
		job.ClaimedAmount = job.TargetQuantity
		job.ClaimableAmount = 0
		job.Status = proddmn.StatusClaimed
	} else {
		job.ClaimableAmount = 0
		// Recompute remaining claimable after update.
	}

	if err := s.production.UpdateJob(ctx, job); err != nil {
		return nil, fmt.Errorf("update job: %w", err)
	}

	// Award incremental production XP.
	xpEarned := s.productionXPForClaim(job, claimAmount)
	if xpEarned > 0 {
		company, err := s.companies.GetCompany(ctx, companyID)
		if err != nil {
			return nil, fmt.Errorf("get company for XP: %w", err)
		}
		company.XP += int64(xpEarned)
		if err := s.companies.UpdateCompany(ctx, company); err != nil {
			return nil, fmt.Errorf("save company XP: %w", err)
		}
		// Persist XP awarded on job for future incremental calculation.
		job.XPAwarded += xpEarned
		if err := s.production.UpdateJob(ctx, job); err != nil {
			return nil, fmt.Errorf("update job xp: %w", err)
		}
	}

	// Append finance ledger entry.
	isPartial := job.Status != proddmn.StatusClaimed
	metadata := map[string]any{
		"resourceId": job.ResourceID,
		"jobId":      job.ID,
		"partial":    isPartial,
	}
	if s.finance != nil {
		if err := s.finance.AppendLedgerEntry(ctx, &finance.LedgerEntry{
			CompanyID: companyID,
			Kind:      "production_output",
			Amount:    float64(claimAmount),
			Direction: "in",
			Metadata:  metadata,
		}); err != nil {
			return nil, fmt.Errorf("append production ledger: %w", err)
		}
	}
	output := map[string]int{fmt.Sprintf("%d", job.ResourceID): claimAmount}
	remaining := job.TargetQuantity - job.ClaimedAmount
	if remaining < 0 {
		remaining = 0
	}
	claimRespStatus := openapi.ClaimProductionResponseStatus(job.Status)

	resp := &openapi.ClaimProductionResponse{
		JobId:         &job.ID,
		Status:        &claimRespStatus,
		Output:        &output,
		ClaimedAmount: &job.ClaimedAmount,
		Remaining:     &remaining,
		Xp:            &xpEarned,
	}

	// Look up current level for response.
	company, err := s.companies.GetCompany(ctx, companyID)
	if err == nil {
		lvl := company.Level
		resp.Level = &lvl
	}

	return resp, nil
}

// ListClaimableJobs returns jobs with claimable amount > 0 for the given company.
func (s *Service) ListClaimableJobs(ctx context.Context, companyID int) (*openapi.ClaimableJobListResponse, error) {
	// Refresh statuses so claimable amounts are current.
	if err := s.refreshJobStatuses(ctx, companyID); err != nil {
		return nil, err
	}

	jobs, err := s.production.GetJobsByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}

	dtos := make([]openapi.ClaimableJobDTO, 0)
	for _, j := range jobs {
		if j.ClaimableAmount <= 0 {
			continue
		}
		jobID := j.ID
		buildingID := j.BuildingID
		resourceID := j.ResourceID
		totalAmt := j.TargetQuantity
		claimedAmt := j.ClaimedAmount
		claimableAmt := j.ClaimableAmount

		dtos = append(dtos, openapi.ClaimableJobDTO{
			JobId:           &jobID,
			BuildingId:      &buildingID,
			ResourceId:      &resourceID,
			TotalAmount:     &totalAmt,
			ClaimedAmount:   &claimedAmt,
			ClaimableAmount: &claimableAmt,
		})
	}

	return &openapi.ClaimableJobListResponse{
		Jobs: &dtos,
	}, nil
}

// productionXPForClaim computes the incremental XP earned for this claim.
// Total XP per job is 10, awarded proportionally to the fraction claimed.
func (s *Service) productionXPForClaim(job *proddmn.ProductionJob, claimAmount int) int {
	if job == nil || job.TargetQuantity <= 0 || claimAmount <= 0 {
		return 0
	}
	totalXP := 10
	// Fraction of total job completed including this claim.
	completedFraction := float64(job.ClaimedAmount) / float64(job.TargetQuantity)
	if completedFraction > 1.0 {
		completedFraction = 1.0
	}
	totalEarned := int(math.Floor(completedFraction * float64(totalXP)))
	if totalEarned > totalXP {
		totalEarned = totalXP
	}
	xp := totalEarned - job.XPAwarded
	if xp < 0 {
		return 0
	}
	return xp
}
