package building

import (
	"context"
	"fmt"

	"github.com/beiwater/NewHaven/backend/internal/apperr"
	"github.com/beiwater/NewHaven/backend/internal/catalog"
	domain "github.com/beiwater/NewHaven/backend/internal/domain/company"
	openapi "github.com/beiwater/NewHaven/backend/internal/generated/openapi"
)

// StockShelf moves items from the company warehouse into a retail building's shelf.
// If price is nil, the existing shelf price is kept (or 0 for new items — the retail
// tick refreshes from market price). If priceLock is set via SetShelfPrice, the price
// is frozen; otherwise the retail tick updates it from the market ticker.
func (s *Service) StockShelf(ctx context.Context, companyID int, buildingID string, resourceID int, quantity int, price *float64) (*openapi.ShelfActionResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	company, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "company not found", err)
	}

	building, entry, err := s.findRetailBuilding(company, buildingID)
	if err != nil {
		return nil, err
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

	// Find or create shelf slot
	shelfIdx := -1
	for i := range building.Shelves {
		if building.Shelves[i].ResourceID == resourceID {
			shelfIdx = i
			break
		}
	}

	if shelfIdx == -1 {
		// New shelf: check slot count
		if len(building.Shelves) >= maxSlots {
			return nil, apperr.BadRequestf("shelf limit reached (%d/%d)", len(building.Shelves), maxSlots)
		}
		// Determine starting price
		var startPrice float64
		if price != nil {
			startPrice = *price
		}
		maxQty := s.shelfCapacity(entry, building.Level)
		building.Shelves = append(building.Shelves, domain.ShelfItem{
			ResourceID: resourceID,
			Quantity:   0,
			MaxQty:     maxQty,
			Price:      startPrice,
			PriceLock:  price != nil,
		})
		shelfIdx = len(building.Shelves) - 1
	}

	shelf := &building.Shelves[shelfIdx]

	// Check max quantity
	if shelf.Quantity+quantity > shelf.MaxQty {
		return nil, apperr.BadRequestf("shelf capacity exceeded: %d + %d > %d", shelf.Quantity, quantity, shelf.MaxQty)
	}

	// Deduct from warehouse inventory
	if err := s.companies.UpdateInventory(ctx, companyID, resourceID, -quantity); err != nil {
		return nil, apperr.WrapMsg(apperr.KindBadRequest, "insufficient warehouse inventory", err)
	}

	// Add to shelf
	shelf.Quantity += quantity

	// Update price if provided
	if price != nil {
		shelf.Price = *price
		shelf.PriceLock = true
	}

	if err := s.companies.UpdateCompany(ctx, company); err != nil {
		// Rollback inventory
		_ = s.companies.UpdateInventory(ctx, companyID, resourceID, quantity)
		return nil, apperr.Internal("failed to save company after stock")
	}

	return &openapi.ShelfActionResponse{
		Shelf: s.shelfToDTO(shelf),
	}, nil
}

// UnstockShelf moves items from a retail building's shelf back to the warehouse.
func (s *Service) UnstockShelf(ctx context.Context, companyID int, buildingID string, resourceID int, quantity int) (*openapi.ShelfActionResponse, error) {
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

	shelfIdx := -1
	for i := range building.Shelves {
		if building.Shelves[i].ResourceID == resourceID {
			shelfIdx = i
			break
		}
	}
	if shelfIdx == -1 {
		return nil, apperr.NotFoundf("resource %d not found on shelves", resourceID)
	}

	shelf := &building.Shelves[shelfIdx]
	if quantity > shelf.Quantity {
		return nil, apperr.BadRequestf("not enough items on shelf: have %d, want %d", shelf.Quantity, quantity)
	}

	shelf.Quantity -= quantity

	// Add back to warehouse
	if err := s.companies.UpdateInventory(ctx, companyID, resourceID, quantity); err != nil {
		return nil, apperr.Internal("failed to return items to warehouse")
	}

	// Remove shelf if empty
	if shelf.Quantity <= 0 {
		building.Shelves = append(building.Shelves[:shelfIdx], building.Shelves[shelfIdx+1:]...)
	}

	if err := s.companies.UpdateCompany(ctx, company); err != nil {
		return nil, apperr.Internal("failed to save company after unstock")
	}

	// Return the updated shelf (or nil if removed)
	if shelf.Quantity > 0 {
		return &openapi.ShelfActionResponse{
			Shelf: s.shelfToDTO(shelf),
		}, nil
	}
	return &openapi.ShelfActionResponse{
		Shelf: nil,
	}, nil
}

// SetShelfPrice updates the price and lock status for a shelf item.
func (s *Service) SetShelfPrice(ctx context.Context, companyID int, buildingID string, resourceID int, price float64, lock bool) (*openapi.ShelfActionResponse, error) {
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
			building.Shelves[i].Price = price
			building.Shelves[i].PriceLock = lock
			if err := s.companies.UpdateCompany(ctx, company); err != nil {
				return nil, apperr.Internal("failed to save company after price update")
			}
			return &openapi.ShelfActionResponse{
				Shelf: s.shelfToDTO(&building.Shelves[i]),
			}, nil
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
	qty := sh.Quantity
	maxQty := sh.MaxQty
	price := sh.Price
	priceLock := sh.PriceLock
	revenue := sh.Revenue
	return &openapi.ShelfItem{
		ResourceId: &rid,
		Quantity:   &qty,
		MaxQty:     &maxQty,
		Price:      &price,
		PriceLock:  &priceLock,
		Revenue:    &revenue,
	}
}

var _ = fmt.Sprintf // keep fmt import available
