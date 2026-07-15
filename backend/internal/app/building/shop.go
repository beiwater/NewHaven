package building

import (
	"context"
	"fmt"

	"github.com/beiwater/NewHaven/backend/internal/apperr"
	domain "github.com/beiwater/NewHaven/backend/internal/domain/company"
	"github.com/beiwater/NewHaven/backend/internal/formula"
	openapi "github.com/beiwater/NewHaven/backend/internal/generated/openapi"
)

func (s *Service) BuyBuilding(ctx context.Context, companyID int, buildingID string) (*openapi.BuyBuildingResponse, error) {
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

	// Check building slot limit
	maxBuildings := 2 + company.Level/2
	if len(company.Buildings) >= maxBuildings {
		return nil, apperr.BadRequestf("building limit reached: %d/%d slots used", len(company.Buildings), maxBuildings)
	}

	cost := *item.Cost
	if company.Money < cost {
		return nil, apperr.InsufficientFunds(fmt.Sprintf("not enough money: need %.0f, have %.0f", cost, company.Money))
	}

	origMoney := company.Money
	company.Money -= cost

	b := &domain.Building{
		ID:         s.idgen.Next("b"),
		BuildingID: kind,
		Kind:       kind,
		Name:       *item.Name,
		Level:      1,
		MapID:      "",
		SlotID:     "",
		X:          0,
		Y:          0,
		RobotCount: 0,
	}

	// Initialize empty shelves for retail buildings
	if entry, ok := s.buildings[kind]; ok && entry.Type == "retail" {
		b.Shelves = []domain.ShelfItem{}
	}
	company.Buildings = append(company.Buildings, *b)

	if err := s.companies.UpdateCompany(ctx, company); err != nil {
		company.Money = origMoney
		company.Buildings = company.Buildings[:len(company.Buildings)-1]
		return nil, apperr.Internalf("save company: %v", err)
	}

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
	}, nil
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
