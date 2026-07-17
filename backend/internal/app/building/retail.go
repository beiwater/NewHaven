package building

import (
	"context"
	"errors"
	"math"

	"github.com/beiwater/NewHaven/backend/internal/apperr"
	"github.com/beiwater/NewHaven/backend/internal/catalog"
	domain "github.com/beiwater/NewHaven/backend/internal/domain/company"
	"github.com/beiwater/NewHaven/backend/internal/formula"
	openapi "github.com/beiwater/NewHaven/backend/internal/generated/openapi"
	"github.com/beiwater/NewHaven/backend/internal/storage"
)

// StockShelf moves items from the company warehouse into a retail building's shelf.
// A stock action starts an immutable sale batch. Its price and quantity remain
// committed until demand sells the batch out; players then start a fresh batch.
func (s *Service) StockShelf(ctx context.Context, companyID int, buildingID string, resourceID, quality, quantity int, price float64) (*openapi.ShelfActionResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if quantity <= 0 {
		return nil, apperr.BadRequest("sale quantity must be positive")
	}
	if price <= 0 || math.IsNaN(price) || math.IsInf(price, 0) {
		return nil, apperr.BadRequest("sale price must be positive")
	}
	if !formula.ValidProductQuality(quality) {
		return nil, apperr.BadRequestf("quality must be between Q%d and Q%d", formula.MinProductQuality, formula.MaxProductQuality)
	}

	retailStore, ok := s.companies.(storage.RetailShelfStorage)
	if !ok {
		return nil, apperr.Internal("retail storage does not support atomic sale batches")
	}
	building, err := retailStore.GetRetailBuilding(ctx, companyID, buildingID)
	if err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "retail building not found", err)
	}
	entry, ok := s.buildings[building.Kind]
	if !ok {
		return nil, apperr.NotFoundf("building kind %d not found in catalog", building.Kind)
	}
	if entry.Type != "retail" {
		return nil, apperr.BadRequestf("building %q is not a retail building", building.Name)
	}
	if building.UpgradeTargetLevel > 0 || building.UpgradeCompletesAt != "" {
		return nil, apperr.Conflict("building is under construction; start sales after the upgrade finishes")
	}

	// Calculate max shelf slots for this building
	maxSlots := entry.RetailSlots + (building.Level-1)*entry.SlotPerLevel
	if maxSlots <= 0 {
		return nil, apperr.BadRequestf("building %q does not support retail", building.Name)
	}

	// Check if this resource is allowed in this building
	if !s.buildingCanSell(entry, resourceID) {
		return nil, apperr.BadRequestf("building %q cannot sell resource %d", building.Name, resourceID)
	}

	// A resource already on a shelf is an active batch. Restocking it would let
	// a player change the committed quantity while the sale is running.
	for i := range building.Shelves {
		if building.Shelves[i].ResourceID == resourceID {
			return nil, apperr.Conflict("sale already active; wait until this batch sells out")
		}
	}
	if len(building.Shelves) >= maxSlots {
		return nil, apperr.BadRequestf("shelf limit reached (%d/%d)", len(building.Shelves), maxSlots)
	}
	maxQty := s.shelfCapacity(entry, building.Level)
	if quantity > maxQty {
		return nil, apperr.BadRequestf("shelf capacity exceeded: %d > %d", quantity, maxQty)
	}

	shelf, err := retailStore.StockRetailShelf(ctx, companyID, buildingID, building.Level, domain.ShelfItem{
		ResourceID: resourceID,
		Quality:    quality,
		Quantity:   quantity,
		MaxQty:     maxQty,
		Price:      price,
		PriceLock:  true,
	}, maxSlots)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrAlreadyExists):
			return nil, apperr.Conflict("sale already active; wait until this batch sells out")
		case errors.Is(err, storage.ErrLimitReached):
			return nil, apperr.Conflict("shelf capacity changed; refresh and try again")
		case errors.Is(err, storage.ErrStateConflict):
			return nil, apperr.Conflict("building state changed; refresh and try again")
		case errors.Is(err, storage.ErrInsufficientInventory):
			return nil, apperr.WrapMsg(apperr.KindBadRequest, "insufficient warehouse inventory", err)
		default:
			return nil, apperr.Internalf("stock sale batch: %v", err)
		}
	}

	return &openapi.ShelfActionResponse{
		Shelf: s.shelfToDTO(shelf),
	}, nil
}

// UnstockShelf moves items from a retail building's shelf back to the warehouse.
func (s *Service) UnstockShelf(ctx context.Context, companyID int, buildingID string, resourceID int, _ int) (*openapi.ShelfActionResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	company, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "company not found", err)
	}

	building, _, err := s.findRetailBuilding(company, buildingID)
	if err != nil {
		return nil, err
	}

	for i := range building.Shelves {
		if building.Shelves[i].ResourceID == resourceID {
			return nil, apperr.Conflict("active sale batches cannot be cancelled; wait until the batch sells out")
		}
	}
	return nil, apperr.NotFoundf("resource %d not found on shelves", resourceID)
}

// SetShelfPrice updates the price and lock status for a shelf item.
func (s *Service) SetShelfPrice(ctx context.Context, companyID int, buildingID string, resourceID int, _ float64, _ bool) (*openapi.ShelfActionResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	company, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "company not found", err)
	}

	building, _, err := s.findRetailBuilding(company, buildingID)
	if err != nil {
		return nil, err
	}

	for i := range building.Shelves {
		if building.Shelves[i].ResourceID == resourceID {
			return nil, apperr.Conflict("active sale price is locked until the batch sells out")
		}
	}

	return nil, apperr.NotFoundf("resource %d not found on shelves", resourceID)
}

// --- helpers ---

func (s *Service) findRetailBuilding(company *domain.Company, buildingID string) (*domain.Building, *catalog.BuildingEntry, error) {
	for i := range company.Buildings {
		b := &company.Buildings[i]
		if b.ID == buildingID {
			entry, ok := s.buildings[b.Kind]
			if !ok {
				return nil, nil, apperr.NotFoundf("building kind %d not found in catalog", b.Kind)
			}
			if entry.Type != "retail" {
				return nil, nil, apperr.BadRequestf("building %q is not a retail building", b.Name)
			}
			return b, entry, nil
		}
	}
	return nil, nil, apperr.NotFoundf("building %q not found", buildingID)
}

func (s *Service) buildingCanSell(entry *catalog.BuildingEntry, resourceID int) bool {
	for _, pid := range entry.Produces {
		if pid == resourceID {
			return true
		}
	}
	return false
}

func (s *Service) shelfCapacity(entry *catalog.BuildingEntry, level int) int {
	base := entry.RetailSlots * 50
	perLevel := entry.SlotPerLevel * 25
	return base + (level-1)*perLevel
}

func (s *Service) shelfToDTO(sh *domain.ShelfItem) *openapi.ShelfItem {
	rid := sh.ResourceID
	quality := sh.Quality
	qty := sh.Quantity
	maxQty := sh.MaxQty
	price := sh.Price
	priceLock := sh.PriceLock
	revenue := sh.Revenue
	return &openapi.ShelfItem{
		ResourceId: &rid,
		Quality:    &quality,
		Quantity:   &qty,
		MaxQty:     &maxQty,
		Price:      &price,
		PriceLock:  &priceLock,
		Revenue:    &revenue,
	}
}
