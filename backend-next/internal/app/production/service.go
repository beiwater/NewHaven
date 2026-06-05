package production

import (
	"context"

	openapi "github.com/newhaven/backend-next/internal/generated/openapi"
	"github.com/newhaven/backend-next/internal/storage"
)

// Service is the production application use case.
type Service struct {
	production storage.ProductionStorage
}

// NewService creates a new production service.
func NewService(production storage.ProductionStorage) *Service {
	return &Service{production: production}
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
