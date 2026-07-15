package building

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/beiwater/NewHaven/backend/internal/apperr"
	domain "github.com/beiwater/NewHaven/backend/internal/domain/company"
	"github.com/beiwater/NewHaven/backend/internal/formula"
	openapi "github.com/beiwater/NewHaven/backend/internal/generated/openapi"
	"github.com/beiwater/NewHaven/backend/internal/storage"
)

func (s *Service) BuyBuilding(ctx context.Context, companyID int, buildingID string, requestIDValue ...*string) (*openapi.BuyBuildingResponse, error) {
	var rawRequestID *string
	if len(requestIDValue) > 0 {
		rawRequestID = requestIDValue[0]
	}
	requestID, err := normalizeBuildingPurchaseRequestID(rawRequestID)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	company, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "no company found for player", err)
	}

	// Find the building in the market catalog
	items, err := s.BuildingMarket(ctx)
	if err != nil {
		return nil, apperr.Internal("failed to load building market")
	}
	var item *openapi.BuildingMarketItem
	for _, mi := range items {
		if mi.Id != nil && *mi.Id == buildingID {
			item = &mi
			break
		}
	}
	if item == nil {
		return nil, apperr.NotFoundf("building %s not found in market", buildingID)
	}

	kind := *item.Kind
	if company.Level < buildingUnlockLevel(kind) {
		return nil, apperr.BadRequestf("building unlocks at level %d", buildingUnlockLevel(kind))
	}

	// Money and slot checks happen in the same storage critical section as the
	// purchase, preventing concurrent requests from overwriting each other.
	maxBuildings := 2 + company.Level/2
	cost := *item.Cost

	b := domain.Building{
		ID:                    s.idgen.Next("b"),
		BuildingID:            kind,
		Kind:                  kind,
		Name:                  *item.Name,
		Level:                 1,
		MapID:                 "",
		SlotID:                "",
		X:                     0,
		Y:                     0,
		RobotCount:            0,
		PurchaseRequestID:     requestID,
		PurchaseCatalogItemID: buildingID,
		PurchaseCost:          cost,
	}

	// Initialize empty shelves for retail buildings
	if entry, ok := s.buildings[kind]; ok && entry.Type == "retail" {
		b.Shelves = []domain.ShelfItem{}
	}

	purchased, replayed, err := s.companies.PurchaseBuilding(ctx, companyID, b, cost, maxBuildings)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrIdempotencyConflict):
			return nil, apperr.Conflict("requestId was already used for a different building purchase")
		case errors.Is(err, storage.ErrLimitReached):
			return nil, apperr.BadRequestf("building limit reached: %d/%d slots used", len(company.Buildings), maxBuildings)
		case errors.Is(err, storage.ErrInsufficientFunds):
			return nil, apperr.InsufficientFunds(fmt.Sprintf("not enough money: need %.0f, have %.0f", cost, company.Money))
		default:
			return nil, apperr.Internalf("save building purchase: %v", err)
		}
	}
	if replayed {
		cost = purchased.PurchaseCost
	}
	return buyBuildingResponse(purchased, cost), nil
}

func normalizeBuildingPurchaseRequestID(value *string) (string, error) {
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

func buyBuildingResponse(b *domain.Building, cost float64) *openapi.BuyBuildingResponse {
	placed := b.MapID != ""
	return &openapi.BuyBuildingResponse{
		Building: &openapi.BuildingDTO{
			Id:         &b.ID,
			BuildingId: &b.BuildingID,
			Name:       &b.Name,
			Level:      &b.Level,
			MapId:      &b.MapID,
			SlotId:     &b.SlotID,
			X:          &b.X,
			Y:          &b.Y,
			RobotCount: &b.RobotCount,
			Placed:     &placed,
		},
		Cost: &cost,
	}
}

func (s *Service) UpgradeBuilding(ctx context.Context, companyID int, buildingID string) (*openapi.UpgradeBuildingResponse, error) {
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
		currLevel := b.Level
		if currLevel <= 0 {
			currLevel = 1
		}
		nextLevel := currLevel + 1
		baseCost := float64(b.BuildingID) * 5000
		// Look up catalog entry for base cost
		if entry, ok := s.buildings[b.BuildingID]; ok && entry.BaseCost > 0 {
			baseCost = float64(entry.BaseCost)
		}
		// Upgrade consumption scales with the building's current size. Charging
		// against nextLevel made every upgrade one full level too expensive.
		cost := formula.UpgradeCost(baseCost, currLevel)

		if company.Money < cost {
			return nil, apperr.InsufficientFunds(fmt.Sprintf("need %.0f to upgrade to level %d, have %.0f", cost, nextLevel, company.Money))
		}

		origMoney := company.Money
		origLevel := company.Buildings[i].Level
		company.Money -= cost
		company.Buildings[i].Level = nextLevel

		if err := s.companies.UpdateCompany(ctx, company); err != nil {
			company.Money = origMoney
			company.Buildings[i].Level = origLevel
			return nil, apperr.Internalf("save company: %v", err)
		}

		return &openapi.UpgradeBuildingResponse{
			BuildingId:       &buildingID,
			OldLevel:         &currLevel,
			NewLevel:         &nextLevel,
			Cost:             &cost,
			OutputMultiplier: &nextLevel,
		}, nil
	}
	return nil, apperr.NotFoundf("building %s not found", buildingID)
}
