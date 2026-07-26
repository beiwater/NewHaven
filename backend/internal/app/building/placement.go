package building

import (
	"context"

	"github.com/beiwater/NewHaven/backend/internal/apperr"
	domain "github.com/beiwater/NewHaven/backend/internal/domain/company"
	openapi "github.com/beiwater/NewHaven/backend/internal/generated/openapi"
)

// -- Service methods --

func (s *Service) PlaceBuilding(ctx context.Context, companyID int, req *openapi.PlaceBuildingRequest) (*openapi.PlaceBuildingResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "no company found for player", err)
	}

	// Find the building (must be unplaced: no mapId set)
	for i, b := range company.Buildings {
		if b.ID != req.BuildingId {
			continue
		}
		if b.MapID != "" {
			return nil, apperr.Conflict("building already placed")
		}
		// Validate target map position via legacy placement logic
		placeMapID := ""
		if req.MapId != nil {
			placeMapID = *req.MapId
		}
		placeSlotID := ""
		if req.SlotId != nil {
			placeSlotID = *req.SlotId
		}
		placeX := 0
		if req.X != nil {
			placeX = *req.X
		}
		placeY := 0
		if req.Y != nil {
			placeY = *req.Y
		}
		mapID, slotID, x, y, err := validateMapPlacement(company.Buildings, "", company.Level, placeMapID, placeSlotID, placeX, placeY)
		if err != nil {
			return nil, apperr.BadRequestf("invalid placement: %v", err)
		}
		origMapID := company.Buildings[i].MapID
		origSlotID := company.Buildings[i].SlotID
		origX := company.Buildings[i].X
		origY := company.Buildings[i].Y
		company.Buildings[i].MapID = mapID
		company.Buildings[i].SlotID = slotID
		company.Buildings[i].X = x
		company.Buildings[i].Y = y

		if err := s.companies.UpdateCompany(ctx, company); err != nil {
			company.Buildings[i].MapID = origMapID
			company.Buildings[i].SlotID = origSlotID
			company.Buildings[i].X = origX
			company.Buildings[i].Y = origY
			return nil, apperr.Internalf("save company: %v", err)
		}

		dto := s.buildingToDTO(&company.Buildings[i])
		status := "placed"
		return &openapi.PlaceBuildingResponse{
			Building: &dto,
			Status:   &status,
		}, nil
	}
	return nil, apperr.NotFoundf("building %s not found", req.BuildingId)
}

func (s *Service) MoveBuilding(ctx context.Context, companyID int, req *openapi.MoveBuildingRequest) (*openapi.MoveBuildingResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "no company found for player", err)
	}

	for i, b := range company.Buildings {
		if b.ID != req.BuildingId {
			continue
		}
		if b.UpgradeTargetLevel > 0 || b.UpgradeCompletesAt != "" {
			return nil, apperr.Conflict("building cannot be moved while it is being upgraded")
		}
		// Validate target map position via legacy placement logic
		moveMapID := ""
		if req.MapId != nil {
			moveMapID = *req.MapId
		}
		moveSlotID := ""
		if req.SlotId != nil {
			moveSlotID = *req.SlotId
		}
		moveX := 0
		if req.X != nil {
			moveX = *req.X
		}
		moveY := 0
		if req.Y != nil {
			moveY = *req.Y
		}
		mapID, slotID, x, y, err := validateMapPlacement(company.Buildings, req.BuildingId, company.Level, moveMapID, moveSlotID, moveX, moveY)
		if err != nil {
			return nil, apperr.BadRequestf("invalid move target: %v", err)
		}
		// Save originals before mutation for rollback
		origMapID := company.Buildings[i].MapID
		origSlotID := company.Buildings[i].SlotID
		origX := company.Buildings[i].X
		origY := company.Buildings[i].Y

		company.Buildings[i].MapID = mapID
		company.Buildings[i].SlotID = slotID
		company.Buildings[i].X = x
		company.Buildings[i].Y = y

		if err := s.companies.UpdateCompany(ctx, company); err != nil {
			company.Buildings[i].MapID = origMapID
			company.Buildings[i].SlotID = origSlotID
			company.Buildings[i].X = origX
			company.Buildings[i].Y = origY
			return nil, apperr.Internalf("save company: %v", err)
		}

		dto := s.buildingToDTO(&company.Buildings[i])
		status := "moved"
		return &openapi.MoveBuildingResponse{
			Building: &dto,
			Status:   &status,
		}, nil
	}
	return nil, apperr.NotFoundf("building %s not found", req.BuildingId)
}

func (s *Service) DemolishBuilding(ctx context.Context, companyID int, req *openapi.DemolishBuildingRequest) (*openapi.DemolishBuildingResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "no company found for player", err)
	}

	for i, b := range company.Buildings {
		if b.ID != req.BuildingId {
			continue
		}
		if b.UpgradeTargetLevel > 0 || b.UpgradeCompletesAt != "" {
			return nil, apperr.Conflict("building cannot be demolished while it is being upgraded")
		}
		// Refund = 50% of base cost
		baseCost := float64(b.BuildingID) * 5000
		if entry, ok := s.buildings[b.BuildingID]; ok && entry.BaseCost > 0 {
			baseCost = float64(entry.BaseCost)
		}
		refund := baseCost * 0.5
		// Save originals before mutation for rollback
		origBuildings := make([]domain.Building, len(company.Buildings))
		copy(origBuildings, company.Buildings)

		company.Buildings = append(company.Buildings[:i], company.Buildings[i+1:]...)

		if err := s.companies.UpdateCompany(ctx, company); err != nil {
			company.Buildings = origBuildings
			return nil, apperr.Internalf("save company: %v", err)
		}
		// Credit the refund atomically after the building is removed. If it fails
		// (company vanished), restore the building so the player isn't charged for
		// nothing.
		if _, err := s.companies.AdjustMoney(ctx, companyID, refund, false); err != nil {
			company.Buildings = origBuildings
			_ = s.companies.UpdateCompany(ctx, company)
			return nil, apperr.Internalf("credit refund: %v", err)
		}

		status := "demolished"
		return &openapi.DemolishBuildingResponse{
			Refund: &refund,
			Status: &status,
		}, nil
	}
	return nil, apperr.NotFoundf("building %s not found", req.BuildingId)
}

// StashBuilding removes a building from the map but keeps it in inventory.
func (s *Service) StashBuilding(ctx context.Context, companyID int, buildingID string) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "no company found for player", err)
	}

	for i, b := range company.Buildings {
		if b.ID != buildingID {
			continue
		}
		if b.UpgradeTargetLevel > 0 || b.UpgradeCompletesAt != "" {
			return nil, apperr.Conflict("building cannot be stashed while it is being upgraded")
		}
		if b.MapID == "" {
			return nil, apperr.BadRequest("building is already stashed")
		}

		origMapID := company.Buildings[i].MapID
		origSlotID := company.Buildings[i].SlotID
		origX := company.Buildings[i].X
		origY := company.Buildings[i].Y

		company.Buildings[i].MapID = ""
		company.Buildings[i].SlotID = ""
		company.Buildings[i].X = 0
		company.Buildings[i].Y = 0

		if err := s.companies.UpdateCompany(ctx, company); err != nil {
			company.Buildings[i].MapID = origMapID
			company.Buildings[i].SlotID = origSlotID
			company.Buildings[i].X = origX
			company.Buildings[i].Y = origY
			return nil, apperr.Internalf("save company: %v", err)
		}

		return map[string]any{"ok": true, "buildingId": buildingID, "status": "stashed"}, nil
	}
	return nil, apperr.NotFoundf("building %s not found", buildingID)
}
