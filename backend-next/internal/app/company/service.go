package company

import (
	"context"
	"errors"
	"github.com/newhaven/backend-next/internal/apperr"

	"github.com/newhaven/backend-next/internal/domain/company"
	"github.com/newhaven/backend-next/internal/generated/openapi"
	"github.com/newhaven/backend-next/internal/platform"
	"github.com/newhaven/backend-next/internal/storage"
)

var ErrNotFound = errors.New("company not found")

// Service is the company application use case.
type Service struct {
	companies storage.CompanyStorage
	logger    *platform.Logger
}

// NewService creates a new company service.
func NewService(companies storage.CompanyStorage, logger *platform.Logger) *Service {
	return &Service{companies: companies, logger: logger}
}

// ListMyCompanies returns the companies owned by the given player.
func (s *Service) ListMyCompanies(ctx context.Context, playerID int) (*openapi.MyCompaniesResponse, error) {
	c, err := s.companies.GetCompanyByPlayerID(ctx, playerID)
	if err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "no company found for player", errors.Join(ErrNotFound, err))
	}

	return &openapi.MyCompaniesResponse{
		Companies: &[]openapi.CompanySummary{
			toSummary(c),
		},
	}, nil
}

func toSummary(c *company.Company) openapi.CompanySummary {
	id := c.ID
	name := c.Name
	money := float32(c.Money)
	level := c.Level
	return openapi.CompanySummary{
		Id:    &id,
		Name:  &name,
		Money: &money,
		Level: &level,
	}
}
