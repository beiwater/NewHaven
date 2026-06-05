package building

import (
	"context"

	openapi "github.com/newhaven/backend-next/internal/generated/openapi"
	"github.com/newhaven/backend-next/internal/storage"
)

// Service is the building application use case.
type Service struct {
	companies storage.CompanyStorage
}

// NewService creates a new building service.
func NewService(companies storage.CompanyStorage) *Service {
	return &Service{companies: companies}
}

// ListMyBuildings returns all buildings for the given company.
func (s *Service) ListMyBuildings(ctx context.Context, companyID int) (*openapi.BuildingListResponse, error) {
	buildings, err := s.companies.GetBuildings(ctx, companyID)
	if err != nil {
		return nil, err
	}

	dtos := make([]openapi.BuildingDTO, 0, len(buildings))
	for _, b := range buildings {
		id := b.ID
		buildingID := b.BuildingID
		name := b.Name
		level := b.Level
		mapID := b.MapID
		slotID := b.SlotID
		x := b.X
		y := b.Y
		robotCount := b.RobotCount
		dtos = append(dtos, openapi.BuildingDTO{
			Id:         &id,
			BuildingId: &buildingID,
			Name:       &name,
			Level:      &level,
			MapId:      &mapID,
			SlotId:     &slotID,
			X:          &x,
			Y:          &y,
			RobotCount: &robotCount,
		})
	}

	return &openapi.BuildingListResponse{
		Buildings: &dtos,
	}, nil
}
