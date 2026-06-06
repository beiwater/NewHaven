package market

import (
	"context"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/newhaven/backend-next/internal/catalog"
	"github.com/newhaven/backend-next/internal/config"
	"github.com/newhaven/backend-next/internal/domain/finance"
	domainmarket "github.com/newhaven/backend-next/internal/domain/market"
	"github.com/newhaven/backend-next/internal/formula"
	openapi "github.com/newhaven/backend-next/internal/generated/openapi"
	"github.com/newhaven/backend-next/internal/platform"
	"github.com/newhaven/backend-next/internal/storage"
)

// Service is the market application use case.
type Service struct {
	mu        sync.Mutex
	market    storage.MarketStorage
	companies storage.CompanyStorage
	finance   storage.FinanceStorage
	resources map[int]*catalog.ResourceEntry
	cfg       *config.GameConfig
	clock     platform.Clock
	idgen     *platform.IDGen
}

// NewService creates a new market service.
func NewService(market storage.MarketStorage, companies storage.CompanyStorage, finance storage.FinanceStorage, resources map[int]*catalog.ResourceEntry, cfg *config.GameConfig, clock platform.Clock, idgen *platform.IDGen) *Service {
	return &Service{
		market:    market,
		companies: companies,
		finance:   finance,
		resources: resources,
		cfg:       cfg,
		clock:     clock,
		idgen:     idgen,
	}
}

// ListResources returns market-tradable resource definitions.
func (s *Service) ListResources(ctx context.Context) (*openapi.ResourcesResponse, error) {
	dtos := make([]openapi.ResourceDefinition, 0)
	for _, r := range s.resources {
		if r.DbLetter <= 0 {
			continue
		}
		if r.IsResearch {
			continue
		}
		if !r.IsExchangeTradable {
			continue
		}
		rid := r.DbLetter
		producedFrom := make(map[string]int)
		for k, v := range r.ProducedFrom {
			producedFrom[fmt.Sprintf("%d", k)] = v
		}
		dto := openapi.ResourceDefinition{
			ResourceId:         &rid,
			Name:               &r.Name,
			ProducedFrom:       &producedFrom,
			ProducedPerHourRaw: &r.ProducedPerHourRaw,
			UnitsSoldAnHour:    &r.UnitsSoldAnHour,
			HasEconomyModel:    &r.HasEconomyModel,
		}
		dtos = append(dtos, dto)
	}
	sort.Slice(dtos, func(i, j int) bool {
		return valueOrZero(dtos[i].ResourceId) < valueOrZero(dtos[j].ResourceId)
	})
	return &openapi.ResourcesResponse{
		Resources: &dtos,
	}, nil
}

// GetMarketTicker returns ticker data for a resource, falling back to a synthetic series.
func (s *Service) GetMarketTicker(ctx context.Context, resourceID int) (*openapi.MarketTickerResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Try reading from storage first.
	ticker, err := s.market.GetTicker(ctx, resourceID)
	if err == nil && ticker != nil {
		series := make([]openapi.MarketTickerPoint, 48)
		now := s.clock.Now().UTC().Truncate(time.Hour)
		for i := 47; i >= 0; i-- {
			ts := now.Add(-time.Duration(i) * time.Hour)
			price32 := float32(ticker.LastPrice)
			series[47-i] = openapi.MarketTickerPoint{
				Price: &price32,
				Time:  &ts,
			}
		}
		return &openapi.MarketTickerResponse{
			Resource: &resourceID,
			Series:   &series,
		}, nil
	}

	// Fallback: synthesize deterministic series from catalog BasePrice.
	basePrice := s.basePriceForResource(resourceID)
	if basePrice <= 0 {
		basePrice = 20.0 + float64(resourceID%11)*3.0
	}

	series := make([]openapi.MarketTickerPoint, 48)
	now := s.clock.Now().UTC().Truncate(time.Hour)
	for i := 47; i >= 0; i-- {
		hour := now.Add(-time.Duration(i) * time.Hour)
		h := hour.Unix() / 3600
		wave := math.Sin(float64(h+int64(resourceID*17))*0.37)*0.025 +
			math.Cos(float64(h+int64(resourceID*31))*0.11)*0.015
		price := float32(math.Round(basePrice*(1+wave)*100) / 100)
		series[47-i] = openapi.MarketTickerPoint{
			Price: &price,
			Time:  &hour,
		}
	}

	return &openapi.MarketTickerResponse{
		Resource: &resourceID,
		Series:   &series,
	}, nil
}

// GetMarketDepth returns aggregated buy/sell depth for a resource and quality.
func (s *Service) GetMarketDepth(ctx context.Context, resourceID int, quality int) (*openapi.MarketDepthResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	orders, err := s.market.GetOrdersByResource(ctx, resourceID)
	if err != nil {
		return nil, err
	}

	buysMap := make(map[float64]int)
	sellsMap := make(map[float64]int)

	for _, o := range orders {
		if o.Quality != quality {
			continue
		}
		if o.Status != domainmarket.StatusOpen && o.Status != domainmarket.StatusPartial {
			continue
		}
		remaining := o.Remaining()
		if remaining <= 0 {
			continue
		}
		if o.IsBuy {
			buysMap[o.Price] += remaining
		} else {
			sellsMap[o.Price] += remaining
		}
	}

	buyPrices := sortFloat64KeysDesc(buysMap)
	sellPrices := sortFloat64KeysAsc(sellsMap)
	if len(buyPrices) > 5 {
		buyPrices = buyPrices[:5]
	}
	if len(sellPrices) > 5 {
		sellPrices = sellPrices[:5]
	}

	buys := make([]openapi.MarketDepthLevel, 0, len(buyPrices))
	for _, p := range buyPrices {
		q := buysMap[p]
		p32 := float32(p)
		level := openapi.MarketDepthLevel{
			Price:    &p32,
			Quantity: &q,
			Qty:      &q,
		}
		buys = append(buys, level)
	}

	sells := make([]openapi.MarketDepthLevel, 0, len(sellPrices))
	for _, p := range sellPrices {
		q := sellsMap[p]
		p32 := float32(p)
		level := openapi.MarketDepthLevel{
			Price:    &p32,
			Quantity: &q,
			Qty:      &q,
		}
		sells = append(sells, level)
	}

	return &openapi.MarketDepthResponse{
		Buys:  &buys,
		Sells: &sells,
	}, nil
}

// ListMarketOrders returns orders for a resource and quality as DTOs.
func (s *Service) ListMarketOrders(ctx context.Context, resourceID int, quality int) (*openapi.MarketOrderListResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	orders, err := s.market.GetOrdersByResource(ctx, resourceID)
	if err != nil {
		return nil, err
	}

	dtos := make([]openapi.MarketOrderDTO, 0)
	sort.Slice(orders, func(i, j int) bool {
		if orders[i].IsBuy != orders[j].IsBuy {
			return !orders[i].IsBuy
		}
		if orders[i].Price != orders[j].Price {
			if orders[i].IsBuy {
				return orders[i].Price > orders[j].Price
			}
			return orders[i].Price < orders[j].Price
		}
		return orders[i].ID < orders[j].ID
	})
	for _, o := range orders {
		if o.Quality != quality {
			continue
		}
		kind := 0
		if o.IsBuy {
			kind = 1
		}
		kindVal := openapi.MarketOrderDTOKind(kind)

		remaining := o.Remaining()

		// Parse CreatedAt string to time.Time.
		var createdAt time.Time
		if o.CreatedAt != "" {
			createdAt, _ = time.Parse(time.RFC3339, o.CreatedAt)
		}

		statusStr := string(o.Status)

		dto := openapi.MarketOrderDTO{
			Id:         &o.ID,
			ResourceId: &o.ResourceID,
			Kind:       &kindVal,
			Price:      float32Ptr(o.Price),
			Quality:    &o.Quality,
			Quantity:   &o.Quantity,
			Remaining:  &remaining,
			CompanyId:  &o.CompanyID,
			CreatedAt:  &createdAt,
			Status:     &statusStr,
		}
		dtos = append(dtos, dto)
	}

	return &openapi.MarketOrderListResponse{
		Orders: &dtos,
	}, nil
}

// CreateOrder creates a new market order (buy or sell).
func (s *Service) CreateOrder(ctx context.Context, companyID int, req *openapi.CreateOrderRequestFrontend) (*openapi.CreateOrderResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate resource exists in catalog.
	if _, ok := s.resources[req.ResourceId]; !ok {
		return nil, fmt.Errorf("resource %d not found", req.ResourceId)
	}

	// Validate kind.
	if req.Kind != 0 && req.Kind != 1 {
		return nil, fmt.Errorf("kind must be 0 (sell) or 1 (buy)")
	}

	// Validate quantity.
	if req.Quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}

	// Validate price.
	if req.Price <= 0 {
		return nil, fmt.Errorf("price must be positive")
	}

	// Validate quality; backend-next only supports quality 0.
	if req.Quality != 0 {
		return nil, fmt.Errorf("non-zero quality not supported in this phase")
	}

	isBuy := req.Kind == 1
	var reservedCompanyID int
	var originalMoney float64

	// Pre-check and reserve funds/inventory.
	if isBuy {
		total := float64(req.Price) * float64(req.Quantity)
		company, err := s.companies.GetCompany(ctx, companyID)
		if err != nil {
			return nil, fmt.Errorf("company lookup: %w", err)
		}
		if company.Money < total {
			return nil, fmt.Errorf("insufficient funds")
		}
		reservedCompanyID = company.ID
		originalMoney = company.Money
		company.Money -= total
		if err := s.companies.UpdateCompany(ctx, company); err != nil {
			return nil, fmt.Errorf("update company: %w", err)
		}
	} else {
		if err := s.companies.UpdateInventory(ctx, companyID, req.ResourceId, -req.Quantity); err != nil {
			return nil, fmt.Errorf("reserve inventory: %w", err)
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
		return nil, fmt.Errorf("create order: %w", err)
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
		return nil, err // will be wrapped by handler as not-found
	}

	if order.CompanyID != companyID {
		return nil, fmt.Errorf("order not found")
	}

	if order.Status == domainmarket.StatusFilled || order.Status == domainmarket.StatusCancelled || order.Remaining() <= 0 {
		return nil, fmt.Errorf("order already settled")
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
			return nil, fmt.Errorf("company lookup: %w", err)
		}
		originalMoney = company.Money
		company.Money += refund
		if err := s.companies.UpdateCompany(ctx, company); err != nil {
			return nil, fmt.Errorf("update company: %w", err)
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
			return nil, fmt.Errorf("return inventory: %w", err)
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
		return nil, fmt.Errorf("update order: %w", err)
	}

	statusStr := "cancelled"
	return &openapi.CancelOrderResponse{
		Id:     &order.ID,
		Status: &statusStr,
	}, nil
}

// TakeOrder takes (buys from) market sell orders up to the requested quantity and maxPrice.
func (s *Service) TakeOrder(ctx context.Context, companyID int, req *openapi.TakeOrderRequest) (*openapi.TakeOrderResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("request is required")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.resources[req.Resource]; !ok {
		return nil, fmt.Errorf("resource %d not found", req.Resource)
	}

	if req.Quantity <= 0 {
		return nil, fmt.Errorf("quantity must be positive")
	}

	if req.MaxPrice <= 0 {
		return nil, fmt.Errorf("maxPrice must be positive")
	}

	if req.Quality != 0 {
		return nil, fmt.Errorf("non-zero quality not supported in this phase")
	}

	taker, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("company lookup: %w", err)
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

	type tradeResult struct {
		fill  int
		cost  float64
		trade *domainmarket.Trade
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
			// Cannot afford the full cost of this fill; stop as per legacy behavior.
			break
		}

		// Deduct from taker.
		takerOriginalMoney := taker.Money
		taker.Money -= cost
		if err := s.companies.UpdateCompany(ctx, taker); err != nil {
			taker.Money = takerOriginalMoney
			return nil, fmt.Errorf("update taker company: %w", err)
		}

		// Add inventory to taker.
		if err := s.companies.UpdateInventory(ctx, companyID, req.Resource, fill); err != nil {
			// Rollback taker money.
			taker.Money = takerOriginalMoney
			_ = s.companies.UpdateCompany(ctx, taker)
			return nil, fmt.Errorf("add inventory to taker: %w", err)
		}

		// Credit seller if different company.
		var sellerOriginalMoney float64
		sellerCredited := false
		seller, err := s.companies.GetCompany(ctx, sell.CompanyID)
		if err == nil && sell.CompanyID != companyID {
			sellerOriginalMoney = seller.Money
			seller.Money += float64(fill) * sell.Price
			if err := s.companies.UpdateCompany(ctx, seller); err != nil {
				// Rollback: undo taker money + inventory.
				taker.Money = takerOriginalMoney
				_ = s.companies.UpdateCompany(ctx, taker)
				_ = s.companies.UpdateInventory(ctx, companyID, req.Resource, -fill)
				return nil, fmt.Errorf("credit seller: %w", err)
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
			taker.Money = takerOriginalMoney
			_ = s.companies.UpdateCompany(ctx, taker)
			_ = s.companies.UpdateInventory(ctx, companyID, req.Resource, -fill)
			if sellerCredited {
				seller.Money = sellerOriginalMoney
				_ = s.companies.UpdateCompany(ctx, seller)
			}
			sell.FilledQuantity = originalFilledQuantity
			sell.Status = originalStatus
			return nil, fmt.Errorf("update sell order: %w", err)
		}

		// Record trade.
		trade := &domainmarket.Trade{
			ID:          s.idgen.Next("trade"),
			BuyOrderID:  "take-" + s.idgen.NanoID(),
			SellOrderID: sell.ID,
			ResourceID:  req.Resource,
			Quality:     req.Quality,
			Quantity:    fill,
			Price:       sell.Price,
			BuyerFee:    fee,
			CreatedAt:   now.Format(time.RFC3339),
		}
		if err := s.market.SaveTrade(ctx, trade); err != nil {
			// Rollback everything.
			taker.Money = takerOriginalMoney
			_ = s.companies.UpdateCompany(ctx, taker)
			_ = s.companies.UpdateInventory(ctx, companyID, req.Resource, -fill)
			if sellerCredited {
				seller.Money = sellerOriginalMoney
				_ = s.companies.UpdateCompany(ctx, seller)
			}
			sell.FilledQuantity = originalFilledQuantity
			sell.Status = originalStatus
			_ = s.market.UpdateOrder(ctx, sell)
			return nil, fmt.Errorf("save trade: %w", err)
		}

		// Update ticker.
		ticker, tickErr := s.market.GetTicker(ctx, req.Resource)
		if tickErr != nil {
			ticker = &domainmarket.Ticker{
				ResourceID: req.Resource,
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
				"resourceId":  req.Resource,
				"quantity":    fill,
				"price":       sell.Price,
				"fee":         fee,
				"tradeId":     trade.ID,
				"sellOrderId": sell.ID,
			},
		})

		results = append(results, tradeResult{fill: fill, cost: cost, trade: trade})
		need -= fill
	}

	// Build response.
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
	}, nil
}

// matchNewBuyOrder attempts to fill a newly-created buy order against existing sell orders.
// Called under s.mu. Mutates order in-place.
func (s *Service) matchNewBuyOrder(ctx context.Context, order *domainmarket.MarketOrder) {
	if !order.IsBuy || order.Remaining() <= 0 {
		return
	}

	allOrders, err := s.market.GetOrdersByResource(ctx, order.ResourceID)
	if err != nil {
		return
	}

	candidates := make([]*domainmarket.MarketOrder, 0)
	for i := range allOrders {
		o := &allOrders[i]
		if o.IsBuy {
			continue
		}
		if o.CompanyID == order.CompanyID {
			continue // same company
		}
		if o.Quality != order.Quality {
			continue
		}
		if o.Status != domainmarket.StatusOpen && o.Status != domainmarket.StatusPartial {
			continue
		}
		if o.Remaining() <= 0 {
			continue
		}
		if o.Price > order.Price {
			continue // sell price too high for this buy limit
		}
		candidates = append(candidates, o)
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Price != candidates[j].Price {
			return candidates[i].Price < candidates[j].Price
		}
		if candidates[i].CreatedAt != candidates[j].CreatedAt {
			return candidates[i].CreatedAt < candidates[j].CreatedAt
		}
		return candidates[i].ID < candidates[j].ID
	})

	for _, sell := range candidates {
		if order.Remaining() <= 0 {
			break
		}
		s.executeMatchFill(ctx, order, sell, nil, nil)
	}
}

// matchNewSellOrder attempts to fill a newly-created sell order against existing buy orders.
// Called under s.mu. Mutates order in-place.
func (s *Service) matchNewSellOrder(ctx context.Context, order *domainmarket.MarketOrder) {
	if order.IsBuy || order.Remaining() <= 0 {
		return
	}

	allOrders, err := s.market.GetOrdersByResource(ctx, order.ResourceID)
	if err != nil {
		return
	}

	candidates := make([]*domainmarket.MarketOrder, 0)
	for i := range allOrders {
		o := &allOrders[i]
		if !o.IsBuy {
			continue
		}
		if o.CompanyID == order.CompanyID {
			continue // same company
		}
		if o.Quality != order.Quality {
			continue
		}
		if o.Status != domainmarket.StatusOpen && o.Status != domainmarket.StatusPartial {
			continue
		}
		if o.Remaining() <= 0 {
			continue
		}
		if o.Price < order.Price {
			continue // buy price too low for this sell limit
		}
		candidates = append(candidates, o)
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Price != candidates[j].Price {
			return candidates[i].Price > candidates[j].Price
		}
		if candidates[i].CreatedAt != candidates[j].CreatedAt {
			return candidates[i].CreatedAt < candidates[j].CreatedAt
		}
		return candidates[i].ID < candidates[j].ID
	})

	for _, buy := range candidates {
		if order.Remaining() <= 0 {
			break
		}
		s.executeMatchFill(ctx, nil, nil, order, buy)
	}
}

// executeMatchFill executes a single fill between a buy and a sell order.
// Called under s.mu. Mutates both orders in-place.
// At least one of newBuyOrder/newSellOrder must be non-nil (the newly created order).
func (s *Service) executeMatchFill(ctx context.Context, newBuyOrder *domainmarket.MarketOrder, existingSell *domainmarket.MarketOrder, newSellOrder *domainmarket.MarketOrder, existingBuy *domainmarket.MarketOrder) {
	// Determine which orders participate.
	var buyOrder, sellOrder *domainmarket.MarketOrder
	var buyerID, sellerID int

	if newBuyOrder != nil {
		buyOrder = newBuyOrder
		buyerID = newBuyOrder.CompanyID
		sellOrder = existingSell
		sellerID = existingSell.CompanyID
	} else {
		buyOrder = existingBuy
		buyerID = existingBuy.CompanyID
		sellOrder = newSellOrder
		sellerID = newSellOrder.CompanyID
	}

	execPrice := sellOrder.Price
	fill := buyOrder.Remaining()
	if sellOrder.Remaining() < fill {
		fill = sellOrder.Remaining()
	}
	fee := formula.ExchangeFee(fill, execPrice, s.exchangeFeePct())

	now := s.clock.Now().UTC()

	// -- Buyer side --

	// Buyer inventory.
	if err := s.companies.UpdateInventory(ctx, buyerID, buyOrder.ResourceID, fill); err != nil {
		return
	}

	// Refund buyer excess limit: fill * (buyOrder.Price - execPrice) if buy.Price > execPrice.
	if buyOrder.Price > execPrice {
		buyerCompany, err := s.companies.GetCompany(ctx, buyerID)
		if err != nil {
			return
		}
		// Refund the reserved excess. The buyer had already reserved buyOrder.Price * quantity.
		// For this fill, only execPrice * fill was spent, so refund (buyOrder.Price - execPrice) * fill.
		excessRefund := (buyOrder.Price - execPrice) * float64(fill)
		buyerCompany.Money += excessRefund
		if err := s.companies.UpdateCompany(ctx, buyerCompany); err != nil {
			return
		}
		_ = s.finance.AppendLedgerEntry(ctx, &finance.LedgerEntry{
			CompanyID: buyerID,
			Kind:      "market_buy_refund",
			Amount:    excessRefund,
			Direction: "in",
			Metadata: map[string]any{
				"orderId":   buyOrder.ID,
				"tradeQty":  fill,
				"execPrice": execPrice,
			},
		})
	}

	// -- Seller side --

	sellerCompany, err := s.companies.GetCompany(ctx, sellerID)
	if err != nil {
		return
	}
	revenue := float64(fill) * execPrice
	sellerCompany.Money += revenue - fee
	if err := s.companies.UpdateCompany(ctx, sellerCompany); err != nil {
		return
	}

	if fee > 0 {
		_ = s.finance.AppendLedgerEntry(ctx, &finance.LedgerEntry{
			CompanyID: sellerID,
			Kind:      "market_fee",
			Amount:    fee,
			Direction: "out",
			Metadata: map[string]any{
				"tradeQty":  fill,
				"execPrice": execPrice,
			},
		})
	}
	_ = s.finance.AppendLedgerEntry(ctx, &finance.LedgerEntry{
		CompanyID: sellerID,
		Kind:      "market_trade",
		Amount:    revenue,
		Direction: "in",
		Metadata: map[string]any{
			"buyOrderId":  buyOrder.ID,
			"sellOrderId": sellOrder.ID,
			"tradeQty":    fill,
			"execPrice":   execPrice,
		},
	})

	// -- Update orders --

	buyOrder.FilledQuantity += fill
	if buyOrder.Remaining() == 0 {
		buyOrder.Status = domainmarket.StatusFilled
	} else {
		buyOrder.Status = domainmarket.StatusPartial
	}
	if err := s.market.UpdateOrder(ctx, buyOrder); err != nil {
		return
	}

	sellOrder.FilledQuantity += fill
	if sellOrder.Remaining() == 0 {
		sellOrder.Status = domainmarket.StatusFilled
	} else {
		sellOrder.Status = domainmarket.StatusPartial
	}
	if err := s.market.UpdateOrder(ctx, sellOrder); err != nil {
		return
	}

	// -- Record trade --

	trade := &domainmarket.Trade{
		ID:          s.idgen.Next("trade"),
		BuyOrderID:  buyOrder.ID,
		SellOrderID: sellOrder.ID,
		ResourceID:  buyOrder.ResourceID,
		Quality:     buyOrder.Quality,
		Quantity:    fill,
		Price:       execPrice,
		BuyerFee:    0,
		SellerFee:   fee,
		CreatedAt:   now.Format(time.RFC3339),
	}
	if err := s.market.SaveTrade(ctx, trade); err != nil {
		return
	}

	// -- Update ticker --

	ticker, tickErr := s.market.GetTicker(ctx, buyOrder.ResourceID)
	if tickErr != nil {
		ticker = &domainmarket.Ticker{
			ResourceID: buyOrder.ResourceID,
		}
	}
	ticker.LastPrice = execPrice
	ticker.Volume24h += float64(fill) * execPrice
	ticker.UpdatedAt = now.Format(time.RFC3339)
	_ = s.market.UpdateTicker(ctx, ticker)
}

func (s *Service) exchangeFeePct() float64 {
	if s.cfg == nil {
		return 0.04
	}
	return s.cfg.ExchangeFeePct
}

// basePriceForResource looks up the catalog BasePrice for a given resource ID.
func (s *Service) basePriceForResource(resourceID int) float64 {
	if r, ok := s.resources[resourceID]; ok {
		return r.BasePrice
	}
	return 0
}

// --- Helpers ---

func float32Ptr(v float64) *float32 {
	r := float32(v)
	return &r
}

func valueOrZero(v *int) int {
	if v == nil {
		return 0
	}
	return *v
}

func sortFloat64KeysDesc(m map[float64]int) []float64 {
	keys := make([]float64, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] > keys[j] })
	return keys
}

func sortFloat64KeysAsc(m map[float64]int) []float64 {
	keys := make([]float64, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}
