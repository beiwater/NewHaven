package market

import (
	"context"
	"errors"
	"sort"
	"time"

	"github.com/beiwater/NewHaven/backend/internal/apperr"
	"github.com/beiwater/NewHaven/backend/internal/domain/company"
	"github.com/beiwater/NewHaven/backend/internal/domain/finance"
	domainmarket "github.com/beiwater/NewHaven/backend/internal/domain/market"
	"github.com/beiwater/NewHaven/backend/internal/formula"
	openapi "github.com/beiwater/NewHaven/backend/internal/generated/openapi"
	"github.com/beiwater/NewHaven/backend/internal/storage"
)

type tradeResult struct {
	fill  int
	cost  float64
	trade *domainmarket.Trade
}

// TakeOrder takes (buys from) market sell orders up to the requested quantity and maxPrice.
func (s *Service) TakeOrder(ctx context.Context, companyID int, req *openapi.TakeOrderRequest) (*openapi.TakeOrderResponse, error) {
	if req == nil {
		return nil, apperr.BadRequest("request is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.resources[req.Resource]; !ok {
		return nil, apperr.NotFoundf("resource %d not found", req.Resource)
	}

	if req.Quantity <= 0 {
		return nil, apperr.BadRequest("quantity must be positive")
	}

	if req.MaxPrice <= 0 {
		return nil, apperr.BadRequest("maxPrice must be positive")
	}

	if req.Quality != 0 {
		return nil, apperr.BadRequest("non-zero quality not supported in this phase")
	}
	requestID, err := normalizeRequestID(req.RequestId)
	if err != nil {
		return nil, err
	}

	taker, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "company not found", err)
	}

	// Collect candidate sell orders.
	allOrders, err := s.market.GetOrdersByResource(ctx, req.Resource)
	if err != nil {
		return nil, err
	}

	sellOrders := make([]*domainmarket.MarketOrder, 0)
	for i := range allOrders {
		o := &allOrders[i]
		if o.IsBuy {
			continue
		}
		if o.Quality != req.Quality {
			continue
		}
		if o.Status != domainmarket.StatusOpen && o.Status != domainmarket.StatusPartial {
			continue
		}
		if o.Remaining() <= 0 {
			continue
		}
		if o.Price > float64(req.MaxPrice) {
			continue
		}
		sellOrders = append(sellOrders, o)
	}

	// Sort by price ascending, then CreatedAt ascending, then ID.
	sort.Slice(sellOrders, func(i, j int) bool {
		if sellOrders[i].Price != sellOrders[j].Price {
			return sellOrders[i].Price < sellOrders[j].Price
		}
		if sellOrders[i].CreatedAt != sellOrders[j].CreatedAt {
			return sellOrders[i].CreatedAt < sellOrders[j].CreatedAt
		}
		return sellOrders[i].ID < sellOrders[j].ID
	})

	execution, replay, err := s.reserveTakeOrderExecution(ctx, companyID, requestID, req)
	if err != nil {
		return nil, err
	}
	if replay != nil {
		return replay, nil
	}

	var results []tradeResult
	need := req.Quantity
	now := s.clock.Now().UTC()

	for _, sell := range sellOrders {
		if need <= 0 {
			break
		}
		fill := need
		if sell.Remaining() < fill {
			fill = sell.Remaining()
		}
		fee := formula.ExchangeFee(fill, sell.Price, s.exchangeFeePct())
		cost := float64(fill)*sell.Price + fee

		if taker.Money < cost {
			break
		}
		result, err := s.executeFillLocked(ctx, taker, sell, companyID, req.Resource, req.Quality, fill, cost, fee, now)
		if err != nil {
			if execution != nil {
				if completeErr := s.completeTakeOrderExecution(ctx, execution, results); completeErr != nil {
					return nil, completeErr
				}
			}
			return nil, err
		}
		results = append(results, *result)
		need -= fill
	}

	if execution != nil {
		if err := s.completeTakeOrderExecution(ctx, execution, results); err != nil {
			return nil, err
		}
	}
	return buildTakeOrderResponse(results), nil
}

func (s *Service) reserveTakeOrderExecution(ctx context.Context, companyID int, requestID string, req *openapi.TakeOrderRequest) (*domainmarket.TakeOrderExecution, *openapi.TakeOrderResponse, error) {
	if requestID == "" {
		return nil, nil, nil
	}
	existing, err := s.market.GetTakeOrderExecution(ctx, companyID, requestID)
	if err != nil {
		return nil, nil, apperr.Internalf("find idempotent take order: %v", err)
	}
	if existing != nil {
		return takeOrderExecutionResult(existing, req)
	}

	execution := &domainmarket.TakeOrderExecution{
		ClientRequestID: requestID,
		CompanyID:       companyID,
		ResourceID:      req.Resource,
		Quantity:        req.Quantity,
		Quality:         req.Quality,
		MaxPrice:        float64(req.MaxPrice),
		CreatedAt:       s.clock.Now().UTC().Format(time.RFC3339),
	}
	if err := s.market.CreateTakeOrderExecution(ctx, execution); err != nil {
		if errors.Is(err, storage.ErrAlreadyExists) {
			existing, findErr := s.market.GetTakeOrderExecution(ctx, companyID, requestID)
			if findErr != nil {
				return nil, nil, apperr.Internalf("find concurrent take order: %v", findErr)
			}
			if existing != nil {
				return takeOrderExecutionResult(existing, req)
			}
		}
		return nil, nil, apperr.Internalf("reserve take order request: %v", err)
	}
	return execution, nil, nil
}

func takeOrderExecutionResult(execution *domainmarket.TakeOrderExecution, req *openapi.TakeOrderRequest) (*domainmarket.TakeOrderExecution, *openapi.TakeOrderResponse, error) {
	if execution.ResourceID != req.Resource ||
		execution.Quantity != req.Quantity ||
		execution.Quality != req.Quality ||
		execution.MaxPrice != float64(req.MaxPrice) {
		return nil, nil, apperr.Conflict("requestId was already used for a different immediate market purchase")
	}
	if !execution.Completed {
		return nil, nil, apperr.Conflict("this immediate market purchase is still being processed")
	}
	return nil, takeOrderExecutionResponse(execution), nil
}

func (s *Service) completeTakeOrderExecution(ctx context.Context, execution *domainmarket.TakeOrderExecution, results []tradeResult) error {
	execution.Completed = true
	execution.Trades = make([]domainmarket.Trade, 0, len(results))
	for _, result := range results {
		execution.Trades = append(execution.Trades, *result.trade)
	}
	if err := s.market.UpdateTakeOrderExecution(ctx, execution); err != nil {
		return apperr.Internalf("complete take order request: %v", err)
	}
	return nil
}

func takeOrderExecutionResponse(execution *domainmarket.TakeOrderExecution) *openapi.TakeOrderResponse {
	results := make([]tradeResult, 0, len(execution.Trades))
	for i := range execution.Trades {
		trade := &execution.Trades[i]
		results = append(results, tradeResult{
			fill:  trade.Quantity,
			cost:  float64(trade.Quantity)*trade.Price + trade.BuyerFee,
			trade: trade,
		})
	}
	return buildTakeOrderResponse(results)
}

// executeFillLocked fills a single sell order for a taker. Called under s.mu.
// Returns the fill result or error. On error, all mutations are rolled back.
func (s *Service) executeFillLocked(
	ctx context.Context,
	taker *company.Company,
	sell *domainmarket.MarketOrder,
	companyID int,
	resourceID int,
	quality int,
	fill int,
	cost float64,
	fee float64,
	now time.Time,
) (*tradeResult, error) {
	// Deduct from taker atomically (funds check + debit under the store lock).
	if _, err := s.companies.AdjustMoney(ctx, companyID, -cost, true); err != nil {
		if err == storage.ErrInsufficientFunds {
			return nil, apperr.InsufficientFunds("insufficient funds")
		}
		return nil, apperr.Internalf("debit taker company: %v", err)
	}

	// Add inventory to taker.
	if err := s.companies.UpdateInventory(ctx, companyID, resourceID, fill); err != nil {
		// Rollback taker money.
		_, _ = s.companies.AdjustMoney(ctx, companyID, cost, false)
		return nil, apperr.Internalf("add inventory to taker: %v", err)
	}

	// Credit seller if different company.
	sellerProceeds := float64(fill) * sell.Price
	sellerCredited := false
	if sell.CompanyID != companyID {
		if _, err := s.companies.AdjustMoney(ctx, sell.CompanyID, sellerProceeds, false); err != nil {
			// Rollback: undo taker money + inventory.
			_, _ = s.companies.AdjustMoney(ctx, companyID, cost, false)
			_ = s.companies.UpdateInventory(ctx, companyID, resourceID, -fill)
			return nil, apperr.Internalf("credit seller: %v", err)
		}
		sellerCredited = true
	}

	// Update sell order.
	originalFilledQuantity := sell.FilledQuantity
	originalStatus := sell.Status
	sell.FilledQuantity += fill
	if sell.Remaining() == 0 {
		sell.Status = domainmarket.StatusFilled
	} else {
		sell.Status = domainmarket.StatusPartial
	}
	if err := s.market.UpdateOrder(ctx, sell); err != nil {
		// Rollback: undo taker, seller, inventory.
		_, _ = s.companies.AdjustMoney(ctx, companyID, cost, false)
		_ = s.companies.UpdateInventory(ctx, companyID, resourceID, -fill)
		if sellerCredited {
			_, _ = s.companies.AdjustMoney(ctx, sell.CompanyID, -sellerProceeds, false)
		}
		sell.FilledQuantity = originalFilledQuantity
		sell.Status = originalStatus
		return nil, apperr.Internalf("update sell order: %v", err)
	}

	// Record trade.
	trade := &domainmarket.Trade{
		ID:          s.idgen.Next("trade"),
		BuyOrderID:  "take-" + s.idgen.NanoID(),
		SellOrderID: sell.ID,
		ResourceID:  resourceID,
		Quality:     quality,
		Quantity:    fill,
		Price:       sell.Price,
		BuyerFee:    fee,
		CreatedAt:   now.Format(time.RFC3339),
	}
	if err := s.market.SaveTrade(ctx, trade); err != nil {
		// Rollback everything.
		_, _ = s.companies.AdjustMoney(ctx, companyID, cost, false)
		_ = s.companies.UpdateInventory(ctx, companyID, resourceID, -fill)
		if sellerCredited {
			_, _ = s.companies.AdjustMoney(ctx, sell.CompanyID, -sellerProceeds, false)
		}
		sell.FilledQuantity = originalFilledQuantity
		sell.Status = originalStatus
		_ = s.market.UpdateOrder(ctx, sell)
		return nil, apperr.Internalf("save trade: %v", err)
	}

	// Update ticker.
	ticker, tickErr := s.market.GetTicker(ctx, resourceID)
	if tickErr != nil {
		ticker = &domainmarket.Ticker{
			ResourceID: resourceID,
		}
	}
	ticker.LastPrice = sell.Price
	ticker.Volume24h += float64(fill) * sell.Price
	ticker.UpdatedAt = now.Format(time.RFC3339)
	_ = s.market.UpdateTicker(ctx, ticker)

	// Append buyer ledger entry.
	_ = s.finance.AppendLedgerEntry(ctx, &finance.LedgerEntry{
		CompanyID: companyID,
		Kind:      "market_take_buy",
		Amount:    cost,
		Direction: "out",
		Metadata: map[string]any{
			"resourceId":  resourceID,
			"quantity":    fill,
			"price":       sell.Price,
			"fee":         fee,
			"tradeId":     trade.ID,
			"sellOrderId": sell.ID,
		},
	})

	return &tradeResult{fill: fill, cost: cost, trade: trade}, nil
}

// buildTakeOrderResponse constructs the response DTO from fill results.
func buildTakeOrderResponse(results []tradeResult) *openapi.TakeOrderResponse {
	amountBought := 0
	moneyDelta := float32(0.0)
	tradeDTOs := make([]openapi.TradeDTO, len(results))
	for i, r := range results {
		amountBought += r.fill
		moneyDelta -= float32(r.cost)
		tID := r.trade.ID
		tResourceID := r.trade.ResourceID
		tQuality := r.trade.Quality
		tQty := r.trade.Quantity
		tPrice := float32(r.trade.Price)
		tBuyOrderID := r.trade.BuyOrderID
		tSellOrderID := r.trade.SellOrderID
		// Parse CreatedAt string to time.Time.
		var tCreatedAt time.Time
		if r.trade.CreatedAt != "" {
			tCreatedAt, _ = time.Parse(time.RFC3339, r.trade.CreatedAt)
		}
		tradeDTOs[i] = openapi.TradeDTO{
			Id:          &tID,
			ResourceId:  &tResourceID,
			Quality:     &tQuality,
			Quantity:    &tQty,
			Price:       &tPrice,
			BuyOrderId:  &tBuyOrderID,
			SellOrderId: &tSellOrderID,
			CreatedAt:   &tCreatedAt,
		}
	}

	return &openapi.TakeOrderResponse{
		AmountBought: &amountBought,
		Trades:       &tradeDTOs,
		MoneyDelta:   &moneyDelta,
	}
}
