package warehouse

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/newhaven/backend-next/internal/apperr"
	"github.com/newhaven/backend-next/internal/config"
	domain "github.com/newhaven/backend-next/internal/domain/warehouse"
	"github.com/newhaven/backend-next/internal/generated/openapi"
	"github.com/newhaven/backend-next/internal/platform"
	"github.com/newhaven/backend-next/internal/storage"
)

var ErrNotFound = errors.New("warehouse not found")

// Service is the warehouse application use case.
type Service struct {
	mu         sync.Mutex
	warehouses storage.WarehouseStorage
	companies  storage.CompanyStorage
	cfg        *config.GameConfig
	logger     *platform.Logger
}

// NewService creates a new warehouse service.
func NewService(warehouses storage.WarehouseStorage, companies storage.CompanyStorage, cfg *config.GameConfig, logger *platform.Logger) *Service {
	return &Service{
		warehouses: warehouses,
		companies:  companies,
		cfg:        cfg,
		logger:     logger,
	}
}

// GetMyWarehouse returns the warehouse for the given company.
func (s *Service) GetMyWarehouse(ctx context.Context, companyID int) (*openapi.GetMyWarehouseData, error) {
	// Verify company exists first
	_, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "warehouse not found", errors.Join(ErrNotFound, err))
	}

	w, err := s.warehouses.GetWarehouse(ctx, companyID)
	if err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "warehouse not found", errors.Join(ErrNotFound, err))
	}

	return toWarehouseData(w), nil
}

func (s *Service) UpgradeWarehouse(ctx context.Context, companyID int) (*openapi.WarehouseUpgradeResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "warehouse not found", errors.Join(ErrNotFound, err))
	}

	warehouseLevel := company.WarehouseLevel
	upgradeCost := 25000.0
	if s.cfg != nil && s.cfg.WarehouseUpgradeCost > 0 {
		upgradeCost = s.cfg.WarehouseUpgradeCost
	}
	cost := float64(warehouseLevel+1) * upgradeCost

	if company.Money < cost {
		return nil, apperr.InsufficientFunds(fmt.Sprintf("not enough money: need %.0f, have %.0f", cost, company.Money))
	}

	origMoney := company.Money
	origLevel := company.WarehouseLevel
	company.Money -= cost
	company.WarehouseLevel++

	// Calculate new capacity
	baseCap := 1000
	if s.cfg != nil && s.cfg.WarehouseBaseCap > 0 {
		baseCap = s.cfg.WarehouseBaseCap
	}
	capacity := (company.WarehouseLevel + 2) * baseCap

	// Update the warehouse storage
	w, whErr := s.warehouses.GetWarehouse(ctx, companyID)
	if whErr != nil {
		company.Money = origMoney
		company.WarehouseLevel = origLevel
		return nil, apperr.Internalf("get warehouse: %v", whErr)
	}

	origCapacity := w.Capacity
	w.Capacity = capacity

	// Update both in a single atomic-like step with rollback
	if err := s.companies.UpdateCompany(ctx, company); err != nil {
		company.Money = origMoney
		company.WarehouseLevel = origLevel
		w.Capacity = origCapacity
		return nil, apperr.Internalf("save company: %v", err)
	}

	if err := s.warehouses.UpdateWarehouse(ctx, w); err != nil {
		// Rollback company changes
		company.Money = origMoney
		company.WarehouseLevel = origLevel
		w.Capacity = origCapacity
		_ = s.companies.UpdateCompany(ctx, company)
		return nil, apperr.Internalf("update warehouse: %v", err)
	}

	return &openapi.WarehouseUpgradeResponse{
		Level:    &company.WarehouseLevel,
		Capacity: &capacity,
		Cost:     &cost,
	}, nil
}

func toWarehouseData(w *domain.Warehouse) *openapi.GetMyWarehouseData {
	companyID := w.CompanyID
	capacity := w.Capacity
	used := w.UsedCapacity

	var items []openapi.WarehouseItem
	if len(w.Items) > 0 {
		items = make([]openapi.WarehouseItem, len(w.Items))
		for i, item := range w.Items {
			rid := item.ResourceID
			rname := item.ResourceName
			q := item.Quality
			a := item.Amount
			items[i] = openapi.WarehouseItem{
				ResourceId:   &rid,
				ResourceName: &rname,
				Quality:      &q,
				Amount:       &a,
			}
		}
	} else {
		items = []openapi.WarehouseItem{}
	}

	return &openapi.GetMyWarehouseData{
		CompanyId:    &companyID,
		Capacity:     &capacity,
		UsedCapacity: &used,
		Items:        &items,
	}
}
