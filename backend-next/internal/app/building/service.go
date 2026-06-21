package building

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/newhaven/backend-next/internal/apperr"
	"github.com/newhaven/backend-next/internal/catalog"
	"github.com/newhaven/backend-next/internal/config"
	domain "github.com/newhaven/backend-next/internal/domain/company"
	openapi "github.com/newhaven/backend-next/internal/generated/openapi"
	"github.com/newhaven/backend-next/internal/platform"
	"github.com/newhaven/backend-next/internal/storage"
)

// Service is the building application use case.
type Service struct {
	mu        sync.Mutex
	companies storage.CompanyStorage
	buildings map[int]*catalog.BuildingEntry
	cfg       *config.GameConfig
	clock     platform.Clock
	idgen     *platform.IDGen
}

func NewService(companies storage.CompanyStorage, buildings map[int]*catalog.BuildingEntry, cfg *config.GameConfig, clock platform.Clock, idgen *platform.IDGen) *Service {
	return &Service{
		companies: companies,
		buildings: buildings,
		cfg:       cfg,
		clock:     clock,
		idgen:     idgen,
	}
}

// ListMyBuildings returns all buildings for the given company.
func (s *Service) ListMyBuildings(ctx context.Context, companyID int) (*openapi.BuildingListResponse, error) {
	company, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "no company found for player", err)
	}
	dtos := make([]openapi.BuildingDTO, 0, len(company.Buildings))
	for _, b := range company.Buildings {
		dtos = append(dtos, s.buildingToDTO(&b))
	}
	return &openapi.BuildingListResponse{Buildings: &dtos}, nil
}

// BuildingMarket returns the list of buildings available for purchase.
func (s *Service) BuildingMarket(ctx context.Context) ([]openapi.BuildingMarketItem, error) {
	items := make([]openapi.BuildingMarketItem, 0, len(s.buildings))
	keys := make([]int, 0, len(s.buildings))
	for k := range s.buildings {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	for _, key := range keys {
		entry := s.buildings[key]
		cost := float64(entry.BaseCost)
		if cost <= 0 {
			baseCost := 50000.0
			if s.cfg != nil && s.cfg.BaseBuildingCost > 0 {
				baseCost = s.cfg.BaseBuildingCost
			}
			cost = baseCost + float64(entry.Kind)*10000
		}
		id := fmt.Sprintf("b-shop-%d", entry.Kind)
		unlockLevel := buildingUnlockLevel(entry.Kind)
		produces := entry.Produces
		if produces == nil {
			produces = []int{}
		}
		isRetail := entry.Type == "retail"
		item := openapi.BuildingMarketItem{
			Id:          &id,
			Name:        &entry.Name,
			Kind:        &entry.Kind,
			Cost:        &cost,
			UnlockLevel: &unlockLevel,
			Description: &entry.Description,
			Produces:    &produces,
			IsRetail:    &isRetail,
		}
		items = append(items, item)
	}
	return items, nil
}

func (s *Service) buildingToDTO(b *domain.Building) openapi.BuildingDTO {
	id := b.ID
	buildingID := b.BuildingID
	name := b.Name
	level := b.Level
	mapID := b.MapID
	slotID := b.SlotID
	x := b.X
	y := b.Y
	robotCount := b.RobotCount
	placed := b.MapID != ""
	isRetail := false
	var produces []int
	if entry, ok := s.buildings[b.Kind]; ok {
		isRetail = entry.Type == "retail"
		produces = entry.Produces
	}
	var shelfDTOs []openapi.ShelfItem
	if len(b.Shelves) > 0 {
		shelfDTOs = make([]openapi.ShelfItem, 0, len(b.Shelves))
		for _, sh := range b.Shelves {
			resourceID := sh.ResourceID
			qty := sh.Quantity
			maxQty := sh.MaxQty
			price := sh.Price
			priceLock := sh.PriceLock
			revenue := sh.Revenue
			shelfDTOs = append(shelfDTOs, openapi.ShelfItem{
				ResourceId: &resourceID,
				Quantity:   &qty,
				MaxQty:     &maxQty,
				Price:      &price,
				PriceLock:  &priceLock,
				Revenue:    &revenue,
			})
		}
	}
	return openapi.BuildingDTO{
		Id:         &id,
		BuildingId: &buildingID,
		Name:       &name,
		Level:      &level,
		MapId:      &mapID,
		SlotId:     &slotID,
		X:          &x,
		Y:          &y,
		RobotCount: &robotCount,
		Placed:     &placed,
		IsRetail:   &isRetail,
		Produces:   &produces,
		Shelves:    &shelfDTOs,
	}
}

func buildingUnlockLevel(kind int) int {
	m := map[int]int{
		1: 1, 2: 2, 3: 2, 4: 3, 5: 4,
		6: 5, 7: 6, 8: 7, 9: 8, 10: 9,
		11: 10, 12: 11,
	}
	if lvl, ok := m[kind]; ok {
		return lvl
	}
	return 1
}
