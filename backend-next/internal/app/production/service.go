package production

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/newhaven/backend-next/internal/catalog"
	"github.com/newhaven/backend-next/internal/config"
	proddmn "github.com/newhaven/backend-next/internal/domain/production"
	openapi "github.com/newhaven/backend-next/internal/generated/openapi"
	"github.com/newhaven/backend-next/internal/platform"
	"github.com/newhaven/backend-next/internal/storage"
)

// MaxDurationSeconds is the cap for any production job (48 hours).
const MaxDurationSeconds = 48 * 3600

// Service is the production application use case.
type Service struct {
	mu         sync.Mutex
	production storage.ProductionStorage
	companies  storage.CompanyStorage
	finance    storage.FinanceStorage
	research   storage.ResearchStorage
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
	research storage.ResearchStorage,
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
		research:   research,
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
	s.mu.Lock()
	defer s.mu.Unlock()
	// Refresh computed fields before returning.
	if err := s.refreshJobStatuses(ctx, companyID); err != nil {
		return nil, err
	}

	jobs, err := s.production.GetJobsByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}
	sort.Slice(jobs, func(i, j int) bool {
		return jobs[i].ID < jobs[j].ID
	})

	dtos := make([]openapi.ProductionJobDTO, 0, len(jobs))
	for _, j := range jobs {
		id := j.ID
		buildingID := j.BuildingID
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
			BuildingId:      &buildingID,
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

// ListClaimableJobs returns jobs with claimable amount > 0 for the given company.
func (s *Service) ListClaimableJobs(ctx context.Context, companyID int) (*openapi.ClaimableJobListResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
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
