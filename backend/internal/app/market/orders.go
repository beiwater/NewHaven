package market

import (
	"context"
	"time"

	"github.com/beiwater/NewHaven/backend/internal/apperr"
	"github.com/beiwater/NewHaven/backend/internal/domain/finance"
	domainmarket "github.com/beiwater/NewHaven/backend/internal/domain/market"
	openapi "github.com/beiwater/NewHaven/backend/internal/generated/openapi"
)

// CreateOrder creates a new market order (buy or sell).
func (s *Service) CreateOrder(ctx context.Context, companyID int, req *openapi.CreateOrderRequestFrontend) (*openapi.CreateOrderResponse, error) {
	if req == nil {
		return nil, apperr.BadRequest("request is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate resource exists in catalog.
	if _, ok := s.resources[req.ResourceId]; !ok {
		return nil, apperr.NotFoundf("resource %d not found", req.ResourceId)
	}

	// Validate kind.
	if req.Kind != 0 && req.Kind != 1 {
		return nil, apperr.BadRequest("kind must be 0 (sell) or 1 (buy)")
	}

	// Validate quantity.
	if req.Quantity <= 0 {
		return nil, apperr.BadRequest("quantity must be positive")
	}

	// Validate price.
	if req.Price <= 0 {
		return nil, apperr.BadRequest("price must be positive")
	}

	// Validate quality; backend only supports quality 0.
	if req.Quality != 0 {
		return nil, apperr.BadRequest("non-zero quality not supported in this phase")
	}

	isBuy := req.Kind == 1
	var reservedCompanyID int
	var originalMoney float64

	// Pre-check and reserve funds/inventory.
	if isBuy {
		total := float64(req.Price) * float64(req.Quantity)
		company, err := s.companies.GetCompany(ctx, companyID)
		if err != nil {
			return nil, apperr.WrapMsg(apperr.KindNotFound, "company not found", err)
		}
		if company.Money < total {
			return nil, apperr.InsufficientFunds("insufficient funds")
		}
		reservedCompanyID = company.ID
		originalMoney = company.Money
		company.Money -= total
		if err := s.companies.UpdateCompany(ctx, company); err != nil {
			return nil, apperr.Internalf("update company: %v", err)
		}
	} else {
		if err := s.companies.UpdateInventory(ctx, companyID, req.ResourceId, -req.Quantity); err != nil {
			return nil, apperr.Internalf("reserve inventory: %v", err)
		}
	}

	// Create the order.
	now := s.clock.Now().UTC()
	order := &domainmarket.MarketOrder{
		ID:             s.idgen.Next("order"),
		CompanyID:      companyID,
		ResourceID:     req.ResourceId,
		IsBuy:          isBuy,
		Price:          float64(req.Price),
		Quantity:       req.Quantity,
		FilledQuantity: 0,
		Quality:        0,
		Status:         domainmarket.StatusOpen,
		CreatedAt:      now.Format(time.RFC3339),
	}

	if err := s.market.CreateOrder(ctx, order); err != nil {
		// Best-effort rollback.
		if isBuy {
			comp, rbErr := s.companies.GetCompany(ctx, reservedCompanyID)
			if rbErr == nil {
				comp.Money = originalMoney
				_ = s.companies.UpdateCompany(ctx, comp)
			}
		} else {
			_ = s.companies.UpdateInventory(ctx, companyID, req.ResourceId, req.Quantity)
		}
		return nil, apperr.Internalf("create order: %v", err)
	}

	// Append ledger entry for buy orders.
	if isBuy {
		total := float64(req.Price) * float64(req.Quantity)
		_ = s.finance.AppendLedgerEntry(ctx, &finance.LedgerEntry{
			CompanyID: companyID,
			Kind:      "market_buy_reserve",
			Amount:    total,
			Direction: "out",
			Metadata: map[string]any{
				"orderId":    order.ID,
				"resourceId": req.ResourceId,
				"quantity":   req.Quantity,
				"price":      float64(req.Price),
			},
		})
	}

	// Auto match the new order against existing opposite-side orders.
	if isBuy {
		s.matchNewBuyOrder(ctx, order)
	} else {
		s.matchNewSellOrder(ctx, order)
	}

	// Build response DTO.
	remaining := order.Remaining()
	kindVal := openapi.MarketOrderDTOKind(1)
	if !isBuy {
		kindVal = 0
	}
	var createdAt time.Time
	if order.CreatedAt != "" {
		createdAt, _ = time.Parse(time.RFC3339, order.CreatedAt)
	}
	statusStr := string(order.Status)
	orderDTO := openapi.MarketOrderDTO{
		Id:         &order.ID,
		ResourceId: &order.ResourceID,
		Kind:       &kindVal,
		Price:      float32Ptr(order.Price),
		Quality:    &order.Quality,
		Quantity:   &order.Quantity,
		Remaining:  &remaining,
		CompanyId:  &order.CompanyID,
		CreatedAt:  &createdAt,
		Status:     &statusStr,
	}

	return &openapi.CreateOrderResponse{
		Order: &orderDTO,
	}, nil
}

// CancelOrder cancels an existing open order and refunds the reserved funds/inventory.
func (s *Service) CancelOrder(ctx context.Context, companyID int, orderID string) (*openapi.CancelOrderResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	order, err := s.market.GetOrder(ctx, orderID)
	if err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "order not found", err)
	}

	if order.CompanyID != companyID {
		return nil, apperr.NotFound("order not found")
	}

	if order.Status == domainmarket.StatusFilled || order.Status == domainmarket.StatusCancelled || order.Remaining() <= 0 {
		return nil, apperr.Conflict("order already settled")
	}

	remaining := order.Remaining()
	originalFilledQuantity := order.FilledQuantity
	originalStatus := order.Status
	var originalMoney float64

	// Refund.
	if order.IsBuy {
		refund := order.Price * float64(remaining)
		company, err := s.companies.GetCompany(ctx, companyID)
		if err != nil {
			return nil, apperr.Internalf("company lookup: %v", err)
		}
		originalMoney = company.Money
		company.Money += refund
		if err := s.companies.UpdateCompany(ctx, company); err != nil {
			return nil, apperr.Internalf("update company: %v", err)
		}
		_ = s.finance.AppendLedgerEntry(ctx, &finance.LedgerEntry{
			CompanyID: companyID,
			Kind:      "market_buy_refund",
			Amount:    refund,
			Direction: "in",
			Metadata: map[string]any{
				"orderId": order.ID,
			},
		})
	} else {
		if err := s.companies.UpdateInventory(ctx, companyID, order.ResourceID, remaining); err != nil {
			return nil, apperr.Internalf("return inventory: %v", err)
		}
	}

	// Mark order cancelled.
	order.Status = domainmarket.StatusCancelled
	if err := s.market.UpdateOrder(ctx, order); err != nil {
		if order.IsBuy {
			company, rbErr := s.companies.GetCompany(ctx, companyID)
			if rbErr == nil {
				company.Money = originalMoney
				_ = s.companies.UpdateCompany(ctx, company)
			}
		} else {
			_ = s.companies.UpdateInventory(ctx, companyID, order.ResourceID, -remaining)
		}
		order.FilledQuantity = originalFilledQuantity
		order.Status = originalStatus
		return nil, apperr.Internalf("update order: %v", err)
	}

	statusStr := "cancelled"
	return &openapi.CancelOrderResponse{
		Id:     &order.ID,
		Status: &statusStr,
	}, nil
}
