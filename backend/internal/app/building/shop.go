package building

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/beiwater/NewHaven/backend/internal/apperr"
	domain "github.com/beiwater/NewHaven/backend/internal/domain/company"
	proddmn "github.com/beiwater/NewHaven/backend/internal/domain/production"
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
	if err := s.settleCompletedUpgrades(ctx, companyID); err != nil {
		return nil, apperr.Internalf("complete building upgrades: %v", err)
	}
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
		if b.UpgradeTargetLevel > 0 || b.UpgradeCompletesAt != "" {
			return nil, apperr.Conflict("this building is already being upgraded")
		}
		if len(b.Shelves) > 0 {
			return nil, apperr.Conflict("wait for active sale batches to sell out before upgrading this building")
		}
		if jobs, ok := s.companies.(storage.ProductionStorage); ok {
			buildingJobs, err := jobs.GetJobsByBuilding(ctx, buildingID)
			if err != nil {
				return nil, apperr.Internalf("load building production jobs: %v", err)
			}
			for _, job := range buildingJobs {
				if job.CompanyID == companyID && job.Status != proddmn.StatusClaimed && job.Status != proddmn.StatusCancelled {
					return nil, apperr.Conflict("collect or cancel the current production run before upgrading this building")
				}
			}
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

		duration := formula.UpgradeDuration(b.BuildingID, currLevel)
		startedAt := time.Now().UTC()
		if s.clock != nil {
			startedAt = s.clock.Now().UTC()
		}
		completesAt := startedAt.Add(duration)
		if upgrades, ok := s.companies.(storage.BuildingUpgradeStorage); ok {
			if _, err := upgrades.StartBuildingUpgrade(ctx, companyID, buildingID, currLevel, nextLevel, cost, startedAt.Format(time.RFC3339), completesAt.Format(time.RFC3339)); err != nil {
				switch {
				case errors.Is(err, storage.ErrInsufficientFunds):
					return nil, apperr.InsufficientFunds(fmt.Sprintf("need %.0f to upgrade to level %d", cost, nextLevel))
				case errors.Is(err, storage.ErrStateConflict):
					return nil, apperr.Conflict("this building changed or is already being upgraded")
				default:
					return nil, apperr.Internalf("start building upgrade: %v", err)
				}
			}
		} else {
			// Compatibility for legacy adapters and failure-injection test doubles.
			if company.Money < cost {
				return nil, apperr.InsufficientFunds(fmt.Sprintf("need %.0f to upgrade to level %d, have %.0f", cost, nextLevel, company.Money))
			}
			origMoney := company.Money
			company.Money -= cost
			company.Buildings[i].UpgradeTargetLevel = nextLevel
			company.Buildings[i].UpgradeStartedAt = startedAt.Format(time.RFC3339)
			company.Buildings[i].UpgradeCompletesAt = completesAt.Format(time.RFC3339)
			if err := s.companies.UpdateCompany(ctx, company); err != nil {
				company.Money = origMoney
				company.Buildings[i].UpgradeTargetLevel = 0
				company.Buildings[i].UpgradeStartedAt = ""
				company.Buildings[i].UpgradeCompletesAt = ""
				return nil, apperr.Internalf("save company: %v", err)
			}
		}
		status := "upgrading"

		return &openapi.UpgradeBuildingResponse{
			BuildingId:         &buildingID,
			OldLevel:           &currLevel,
			NewLevel:           &nextLevel,
			Cost:               &cost,
			OutputMultiplier:   &nextLevel,
			Status:             &status,
			UpgradeCompletesAt: &completesAt,
		}, nil
	}
	return nil, apperr.NotFoundf("building %s not found", buildingID)
}
