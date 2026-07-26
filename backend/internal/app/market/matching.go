package market

import (
	"context"
	"sort"
	"time"

	"github.com/beiwater/NewHaven/backend/internal/domain/finance"
	domainmarket "github.com/beiwater/NewHaven/backend/internal/domain/market"
	"github.com/beiwater/NewHaven/backend/internal/formula"
)

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
	var undo []func()
	var err error
	defer func() {
		if r := recover(); r != nil {
			for i := len(undo) - 1; i >= 0; i-- {
				undo[i]()
			}
			panic(r)
		}
		if err != nil {
			for i := len(undo) - 1; i >= 0; i-- {
				undo[i]()
			}
		}
	}()

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
	if err = s.companies.UpdateInventory(ctx, buyerID, buyOrder.ResourceID, fill); err != nil {
		return
	}
	undo = append(undo, func() {
		_ = s.companies.UpdateInventory(ctx, buyerID, buyOrder.ResourceID, -fill)
	})

	// Refund buyer excess limit: fill * (buyOrder.Price - execPrice) if buy.Price > execPrice.
	if buyOrder.Price > execPrice {
		// Refund the reserved excess. The buyer had already reserved buyOrder.Price * quantity.
		// For this fill, only execPrice * fill was spent, so refund (buyOrder.Price - execPrice) * fill.
		// AdjustMoney applies the credit atomically under the store lock, so it
		// cannot be lost to a concurrent settlement mutating the same balance.
		excessRefund := (buyOrder.Price - execPrice) * float64(fill)
		if _, err = s.companies.AdjustMoney(ctx, buyerID, excessRefund, false); err != nil {
			return
		}
		undo = append(undo, func() {
			_, _ = s.companies.AdjustMoney(ctx, buyerID, -excessRefund, false)
		})
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

	revenue := float64(fill) * execPrice
	net := revenue - fee
	if _, err = s.companies.AdjustMoney(ctx, sellerID, net, false); err != nil {
		return
	}
	undo = append(undo, func() {
		_, _ = s.companies.AdjustMoney(ctx, sellerID, -net, false)
	})

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

	oldBuyFilled := buyOrder.FilledQuantity
	oldBuyStatus := buyOrder.Status
	buyOrder.FilledQuantity += fill
	if buyOrder.Remaining() == 0 {
		buyOrder.Status = domainmarket.StatusFilled
	} else {
		buyOrder.Status = domainmarket.StatusPartial
	}
	if err = s.market.UpdateOrder(ctx, buyOrder); err != nil {
		return
	}
	undo = append(undo, func() {
		buyOrder.FilledQuantity = oldBuyFilled
		buyOrder.Status = oldBuyStatus
		_ = s.market.UpdateOrder(ctx, buyOrder)
	})

	oldSellFilled := sellOrder.FilledQuantity
	oldSellStatus := sellOrder.Status
	sellOrder.FilledQuantity += fill
	if sellOrder.Remaining() == 0 {
		sellOrder.Status = domainmarket.StatusFilled
	} else {
		sellOrder.Status = domainmarket.StatusPartial
	}
	if err = s.market.UpdateOrder(ctx, sellOrder); err != nil {
		return
	}
	undo = append(undo, func() {
		sellOrder.FilledQuantity = oldSellFilled
		sellOrder.Status = oldSellStatus
		_ = s.market.UpdateOrder(ctx, sellOrder)
	})

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
	if err = s.market.SaveTrade(ctx, trade); err != nil {
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
