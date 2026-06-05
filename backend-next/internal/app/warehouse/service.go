package warehouse

import (
	"context"
	"errors"

	domain "github.com/newhaven/backend-next/internal/domain/warehouse"
	"github.com/newhaven/backend-next/internal/generated/openapi"
	"github.com/newhaven/backend-next/internal/platform"
	"github.com/newhaven/backend-next/internal/storage"
)

var ErrNotFound = errors.New("warehouse not found")

// Service is the warehouse application use case.
type Service struct {
	warehouses storage.WarehouseStorage
	companies  storage.CompanyStorage
	logger     *platform.Logger
}

// NewService creates a new warehouse service.
func NewService(warehouses storage.WarehouseStorage, companies storage.CompanyStorage, logger *platform.Logger) *Service {
	return &Service{
		warehouses: warehouses,
		companies:  companies,
		logger:     logger,
	}
}

// GetMyWarehouse returns the warehouse for the given company.
func (s *Service) GetMyWarehouse(ctx context.Context, companyID int) (*openapi.GetMyWarehouseData, error) {
	// Verify company exists first
	_, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, errors.Join(ErrNotFound, err)
	}

	w, err := s.warehouses.GetWarehouse(ctx, companyID)
	if err != nil {
		return nil, errors.Join(ErrNotFound, err)
	}

	return toWarehouseData(w), nil
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
