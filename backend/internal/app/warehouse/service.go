package warehouse

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/beiwater/NewHaven/backend/internal/apperr"
	"github.com/beiwater/NewHaven/backend/internal/config"
	domain "github.com/beiwater/NewHaven/backend/internal/domain/warehouse"
	"github.com/beiwater/NewHaven/backend/internal/generated/openapi"
	"github.com/beiwater/NewHaven/backend/internal/platform"
	"github.com/beiwater/NewHaven/backend/internal/storage"
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

	// Fetch the warehouse before charging so a fetch failure costs nothing.
	w, whErr := s.warehouses.GetWarehouse(ctx, companyID)
	if whErr != nil {
		return nil, apperr.Internalf("get warehouse: %v", whErr)
	}

	// Charge atomically (funds check + debit happen together under the store
	// lock, so money can't be lost to a concurrent settlement).
	if _, err := s.companies.AdjustMoney(ctx, companyID, -cost, true); err != nil {
		if err == storage.ErrInsufficientFunds {
			return nil, apperr.InsufficientFunds(fmt.Sprintf("not enough money: need %.0f, have %.0f", cost, company.Money))
		}
		return nil, apperr.Internalf("charge warehouse upgrade: %v", err)
	}

	// Calculate new capacity
	baseCap := 1000
	if s.cfg != nil && s.cfg.WarehouseBaseCap > 0 {
		baseCap = s.cfg.WarehouseBaseCap
	}
	origLevel := company.WarehouseLevel
	origCapacity := w.Capacity
	company.WarehouseLevel++
	capacity := (company.WarehouseLevel + 2) * baseCap

	// Persist the level bump first (money is authoritative in the store and
	// untouched by UpdateCompany), refunding the charge on any failure. Capacity
	// is only mutated right before its own persist so an earlier failure leaves
	// the warehouse untouched.
	if err := s.companies.UpdateCompany(ctx, company); err != nil {
		company.WarehouseLevel = origLevel
		_, _ = s.companies.AdjustMoney(ctx, companyID, cost, false)
		return nil, apperr.Internalf("save company: %v", err)
	}

	w.Capacity = capacity
	if err := s.warehouses.UpdateWarehouse(ctx, w); err != nil {
		company.WarehouseLevel = origLevel
		w.Capacity = origCapacity
		_ = s.companies.UpdateCompany(ctx, company)
		_, _ = s.companies.AdjustMoney(ctx, companyID, cost, false)
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
