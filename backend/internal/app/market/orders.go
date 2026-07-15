package market

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/beiwater/NewHaven/backend/internal/apperr"
	"github.com/beiwater/NewHaven/backend/internal/domain/finance"
	domainmarket "github.com/beiwater/NewHaven/backend/internal/domain/market"
	"github.com/beiwater/NewHaven/backend/internal/formula"
	openapi "github.com/beiwater/NewHaven/backend/internal/generated/openapi"
	"github.com/beiwater/NewHaven/backend/internal/storage"
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
	if !formula.IsValidTick(float64(req.Price)) {
		return nil, apperr.BadRequestf("price must use a %.3f market tick", formula.TickStep(float64(req.Price)))
	}

	// Validate quality; backend only supports quality 0.
	if req.Quality != 0 {
		return nil, apperr.BadRequest("non-zero quality not supported in this phase")
	}

	isBuy := req.Kind == 1
	requestID, err := normalizeRequestID(req.RequestId)
	if err != nil {
		return nil, err
	}
	if requestID != "" {
		existing, err := s.findOrderByRequestID(ctx, companyID, requestID)
		if err != nil {
			return nil, err
		}
		if existing != nil {
			if !sameCreateOrder(existing, req, isBuy) {
				return nil, apperr.Conflict("requestId was already used for a different market order")
			}
			return createOrderResponse(existing), nil
		}
	}

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
		ID:              s.idgen.Next("order"),
		ClientRequestID: requestID,
		CompanyID:       companyID,
		ResourceID:      req.ResourceId,
		IsBuy:           isBuy,
		Price:           float64(req.Price),
		Quantity:        req.Quantity,
		FilledQuantity:  0,
		Quality:         0,
		Status:          domainmarket.StatusOpen,
		CreatedAt:       now.Format(time.RFC3339),
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

	return createOrderResponse(order), nil
}

func normalizeRequestID(value *string) (string, error) {
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

func (s *Service) findOrderByRequestID(ctx context.Context, companyID int, requestID string) (*domainmarket.MarketOrder, error) {
	order, err := s.market.GetOrderByClientRequestID(ctx, companyID, requestID)
	if err != nil {
		return nil, apperr.Internalf("find idempotent market order: %v", err)
	}
	return order, nil
}

func sameCreateOrder(order *domainmarket.MarketOrder, req *openapi.CreateOrderRequestFrontend, isBuy bool) bool {
	return order.ResourceID == req.ResourceId &&
		order.IsBuy == isBuy &&
		order.Price == float64(req.Price) &&
		order.Quantity == req.Quantity &&
		order.Quality == req.Quality
}

func createOrderResponse(order *domainmarket.MarketOrder) *openapi.CreateOrderResponse {
	remaining := order.Remaining()
	kindVal := openapi.MarketOrderDTOKind(1)
	if !order.IsBuy {
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
	}
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

	cancelled, replayed, err := s.market.CancelMarketOrder(ctx, companyID, orderID)
	if errors.Is(err, storage.ErrAlreadySettled) || replayed {
		return nil, apperr.Conflict("order already settled")
	}
	if err != nil {
		return nil, apperr.Internalf("cancel order: %v", err)
	}

	if cancelled.IsBuy {
		refund := cancelled.Price * float64(cancelled.Remaining())
		_ = s.finance.AppendLedgerEntry(ctx, &finance.LedgerEntry{
			CompanyID: companyID,
			Kind:      "market_buy_refund",
			Amount:    refund,
			Direction: "in",
			Metadata: map[string]any{
				"orderId": cancelled.ID,
			},
		})
	}

	statusStr := "cancelled"
	return &openapi.CancelOrderResponse{
		Id:     &cancelled.ID,
		Status: &statusStr,
	}, nil
}
