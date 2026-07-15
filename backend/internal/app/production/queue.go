package production

import (
	"context"
	"fmt"
	"sort"

	"github.com/beiwater/NewHaven/backend/internal/apperr"
	proddmn "github.com/beiwater/NewHaven/backend/internal/domain/production"
	openapi "github.com/beiwater/NewHaven/backend/internal/generated/openapi"
)

// ProductionQueue returns the production queue overview grouped by building.
// Returns byBuilding (map of buildingID -> jobs), inUse (count), maxSlots (total capacity).
func (s *Service) ProductionQueue(ctx context.Context, companyID int) (*openapi.ProductionQueueResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "company not found", err)
	}

	if err := s.refreshJobStatuses(ctx, companyID); err != nil {
		return nil, apperr.Internalf("refresh production queue: %v", err)
	}
	jobs, err := s.production.GetJobsByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}

	byBuilding := make(map[string][]openapi.ProductionJobDTO)
	inUse := 0
	for _, j := range jobs {
		if j.Status == proddmn.StatusClaimed {
			continue
		}
		inUse++
		// Convert to DTO
		jobID := j.ID
		buildingID := j.BuildingID
		resourceID := j.ResourceID
		quantity := j.Quantity
		targetQty := j.TargetQuantity
		dur := float32(j.DurationSeconds)
		claimedAmt := j.ClaimedAmount
		claimableAmt := j.ClaimableAmount
		status := openapi.ProductionJobDTOStatus(j.Status)
		dto := openapi.ProductionJobDTO{
			Id:              &jobID,
			BuildingId:      &buildingID,
			ResourceId:      &resourceID,
			Quantity:        &quantity,
			TargetQuantity:  &targetQty,
			StartedAt:       &j.StartedAt,
			DurationSeconds: &dur,
			ClaimedAmount:   &claimedAmt,
			ClaimableAmount: &claimableAmt,
			Status:          &status,
		}
		byBuilding[j.BuildingID] = append(byBuilding[j.BuildingID], dto)
	}

	// Calculate max slots
	baseSlots := 3
	if s.cfg != nil && s.cfg.BaseProductionSlots > 0 {
		baseSlots = s.cfg.BaseProductionSlots
	}
	maxSlots := baseSlots * len(company.Buildings)
	if maxSlots < 1 {
		maxSlots = 1
	}

	return &openapi.ProductionQueueResponse{
		ByBuilding: &byBuilding,
		InUse:      &inUse,
		MaxSlots:   &maxSlots,
	}, nil
}

// ProductionOptions returns what resources a building can produce.
func (s *Service) ProductionOptions(ctx context.Context, companyID int, buildingID string) ([]openapi.ResourceDefinition, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "company not found", err)
	}

	// Find the building
	var buildingKind int
	for _, b := range company.Buildings {
		if b.ID == buildingID {
			buildingKind = b.BuildingID
			break
		}
	}
	if buildingKind == 0 {
		return nil, apperr.NotFoundf("building %s not found", buildingID)
	}

	entry, ok := s.buildings[buildingKind]
	if !ok {
		return nil, apperr.NotFoundf("building type %d not found", buildingKind)
	}

	// Return resources this building can produce
	var result []openapi.ResourceDefinition
	for _, pid := range entry.Produces {
		res, ok := s.resources[pid]
		if !ok {
			continue
		}

		rid := res.DbLetter
		name := res.Name
		rate := res.ProducedPerHourRaw
		sold := res.UnitsSoldAnHour
		eco := res.HasEconomyModel
		producedFrom := make(map[string]int)
		for k, v := range res.ProducedFrom {
			producedFrom[fmt.Sprintf("%d", k)] = v
		}

		result = append(result, openapi.ResourceDefinition{
			ResourceId:         &rid,
			Name:               &name,
			ProducedFrom:       &producedFrom,
			ProducedPerHourRaw: &rate,
			UnitsSoldAnHour:    &sold,
			HasEconomyModel:    &eco,
		})
	}

	if result == nil {
		result = []openapi.ResourceDefinition{}
	}

	sort.Slice(result, func(i, j int) bool {
		return valueOrZero(result[i].ResourceId) < valueOrZero(result[j].ResourceId)
	})

	return result, nil
}

// CancelProductionJob cancels a running production job and refunds 50% of inputs.
func (s *Service) CancelProductionJob(ctx context.Context, companyID int, jobID string) (*openapi.CancelJobResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, err := s.production.GetJob(ctx, jobID)
	if err != nil {
		return nil, apperr.NotFoundf("job %s not found", jobID)
	}
	if job.CompanyID != companyID {
		return nil, apperr.NotFoundf("job %s not found", jobID)
	}
	if job.Status == proddmn.StatusClaimed {
		return nil, apperr.Conflict("job already claimed")
	}

	// Refund 50% of inputs based on the resource recipe.
	refunds := make(map[int]int)
	resEntry, ok := s.resources[job.ResourceID]
	if ok {
		for resourceID, amountPerUnit := range resEntry.ProducedFrom {
			refundAmount := (amountPerUnit * job.Quantity) / 2
			if refundAmount > 0 {
				if err := s.companies.UpdateInventory(ctx, companyID, resourceID, refundAmount); err != nil {
					rollbackInventory(ctx, s.companies, companyID, refunds)
					return nil, apperr.Internalf("refund input %d: %v", resourceID, err)
				}
				refunds[resourceID] = refundAmount
			}
		}
	}

	if err := s.production.DeleteJob(ctx, job.ID); err != nil {
		rollbackInventory(ctx, s.companies, companyID, refunds)
		return nil, apperr.Internalf("delete cancelled job: %v", err)
	}

	status := "cancelled"
	return &openapi.CancelJobResponse{
		JobId:  &jobID,
		Status: &status,
	}, nil
}

// ClaimAll claims all claimable jobs for the company.
// Returns partial success semantics: claimed array, errors array, total count.
func (s *Service) ClaimAll(ctx context.Context, companyID int) (*openapi.ClaimAllResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.refreshJobStatuses(ctx, companyID); err != nil {
		return nil, apperr.Internalf("refresh claimable jobs: %v", err)
	}
	jobs, err := s.production.GetJobsByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}

	claimed := make([]interface{}, 0)
	errs := make([]string, 0)

	for _, j := range jobs {
		if j.ClaimableAmount <= 0 || j.Status == proddmn.StatusClaimed {
			continue
		}
		resp, err := s.claimProductionLocked(ctx, companyID, j.ID)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		claimed = append(claimed, map[string]any{
			"jobId":         *resp.JobId,
			"status":        string(*resp.Status),
			"claimedAmount": *resp.ClaimedAmount,
			"xp":            *resp.Xp,
		})
	}

	total := len(claimed)

	return &openapi.ClaimAllResponse{
		Claimed: &claimed,
		Errors:  &errs,
		Total:   &total,
	}, nil
}

func rollbackInventory(ctx context.Context, companies interface {
	UpdateInventory(context.Context, int, int, int) error
}, companyID int, changes map[int]int) {
	for resourceID, amount := range changes {
		_ = companies.UpdateInventory(ctx, companyID, resourceID, -amount)
	}
}

func valueOrZero(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}
