// Package service implements the simulation service layer, including
// order book management, trade execution, and market matching logic.
package service

import (
	"fmt"
	"sort"
	"time"

	"go-sim-api/internal/anticheat"
	"go-sim-api/internal/formula"
	"go-sim-api/internal/model"
)

func (s *Service) CreateOrder(companyID int, resourceID, kind, quality, quantity int, price float64) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}
	pid := company.ID
	if ok, msg := s.AC.CheckRateLimit(pid); !ok {
		return nil, fmt.Errorf("cheat detected: %s", msg)
	}
	s.AC.RecordAction(pid, anticheat.ActCreateOrder, fmt.Sprintf("res=%d price=%.2f qty=%d", resourceID, price, quantity))
	s.SD.RecordAction(pid)

	if quantity <= 0 || price <= 0 || (kind != 0 && kind != 1) {
		return nil, fmt.Errorf("invalid order payload")
	}
	if !formula.IsValidTick(price) {
		return nil, fmt.Errorf("price does not match tick size")
	}
	total := float64(quantity) * price
	if kind == 1 && company.Money < total {
		return nil, fmt.Errorf("not enough cash")
	}
	if kind == 0 && s.inventoryGet(company, resourceID, quality) < quantity {
		return nil, fmt.Errorf("not enough inventory")
	}
	// Check if market is locked for this resource
	if s.State.MarketLocked[resourceID] {
		return nil, fmt.Errorf("market locked: price cap in effect")
	}
	if kind == 1 {
		company.Money -= total
	} else {
		s.inventorySub(company, resourceID, quality, quantity)
	}

	order := model.MarketOrder{
		ID:         uniqueMarketID("order", company.ID, resourceID, kind),
		ResourceID: resourceID, Kind: kind, Price: price, Quality: quality,
		Quantity: quantity, Remaining: quantity, Status: "open", CompanyID: company.ID, CreatedAt: s.now().UTC().Format(time.RFC3339),
	}
	s.State.Orders = append([]model.MarketOrder{order}, s.State.Orders...)
	if kind == 1 {
		s.addLedger("market_buy_reserve", total, "out", map[string]any{"resourceId": resourceID, "orderId": order.ID})
	}
	s.matchLimitOrders(companyID, resourceID, quality)
	s.saveOrdersLocked()
	s.saveCompanyLocked(company)
	// Player order replaces bot orders
	if kind == 0 { // player selling -> reduces bot sell orders
		s.replaceBotOrders(resourceID, quality, quantity, 0)
	} else { // player buying -> reduces bot buy orders
		s.replaceBotOrders(resourceID, quality, quantity, 1)
	}
	return map[string]any{"order": order}, nil
}

func (s *Service) CancelOrder(companyID int, orderID string) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}
	pid := company.ID
	if ok, msg := s.AC.CheckRateLimit(pid); !ok {
		return nil, fmt.Errorf("cheat detected: %s", msg)
	}
	if ok, msg := s.AC.CheckQuickCancel(pid, anticheat.ActCancelOrder); !ok {
		return nil, fmt.Errorf("cheat detected: %s", msg)
	}
	s.AC.RecordAction(pid, anticheat.ActCancelOrder, fmt.Sprintf("order=%s", orderID))
	s.SD.RecordAction(pid)

	for i := range s.State.Orders {
		o := &s.State.Orders[i]
		if o.ID != orderID {
			continue
		}
		if o.Remaining <= 0 {
			return nil, fmt.Errorf("order already filled")
		}
		if o.Status != "" && o.Status != "open" {
			return nil, fmt.Errorf("order not cancellable")
		}
		if o.Kind == 1 {
			company.Money += float64(o.Remaining) * o.Price
			s.addLedger("market_buy_refund", float64(o.Remaining)*o.Price, "in", map[string]any{"orderId": o.ID})
		} else {
			s.inventoryAdd(company, o.ResourceID, o.Quality, o.Remaining)
		}
		o.Remaining = 0
		o.Status = "cancelled"
		s.saveOrdersLocked()
		s.saveCompanyLocked(company)
		return map[string]any{"id": o.ID, "status": "cancelled"}, nil
	}
	return nil, fmt.Errorf("order not found")
}

func (s *Service) matchLimitOrders(companyID int, resourceID, quality int) {
	buys, sells := s.collectOrders(resourceID, quality)
	for _, buy := range buys {
		for _, sell := range sells {
			if buy.Remaining <= 0 || sell.Remaining <= 0 {
				continue
			}
			if buy.Price < sell.Price {
				continue
			}
			fill := min(buy.Remaining, sell.Remaining)
			s.executeMatch(companyID, buy, sell, fill, resourceID)
		}
	}
}

func (s *Service) collectOrders(resourceID, quality int) (buys, sells []*model.MarketOrder) {
	for i := range s.State.Orders {
		o := &s.State.Orders[i]
		if o.ResourceID != resourceID || o.Quality != quality || o.Remaining <= 0 || o.Status == "filled" || o.Status == "cancelled" {
			continue
		}
		if o.Kind == 1 {
			buys = append(buys, o)
		} else {
			sells = append(sells, o)
		}
	}
	sort.Slice(buys, func(i, j int) bool { return buys[i].Price > buys[j].Price })
	sort.Slice(sells, func(i, j int) bool { return sells[i].Price < sells[j].Price })
	return
}

func (s *Service) executeMatch(companyID int, buy, sell *model.MarketOrder, fill int, resourceID int) {
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return
	}
	execPrice := sell.Price
	fee := formula.ExchangeFee(fill, execPrice, s.Cfg.Game.ExchangeFeePct)
	buy.Remaining -= fill
	sell.Remaining -= fill
	if buy.Remaining == 0 {
		buy.Status = "filled"
	}
	if sell.Remaining == 0 {
		sell.Status = "filled"
	}
	buyCompany := s.getCompanyLocked(buy.CompanyID)
	sellCompany := s.getCompanyLocked(sell.CompanyID)
	if buyCompany != nil {
		if buyCompany.ID == company.ID {
			s.inventoryAdd(company, resourceID, buy.Quality, fill)
			company.Money += float64(fill) * (buy.Price - execPrice)
		} else {
			if buyCompany.Inventory == nil {
				buyCompany.Inventory = map[int]int{}
			}
			buyCompany.Inventory[resourceID] += fill
		}
	}
	if sellCompany != nil {
		if sellCompany.ID == company.ID {
			company.Money += float64(fill)*execPrice - fee
		} else {
			sellCompany.Money += float64(fill)*execPrice - fee
		}
	}
	s.State.MarketPressure[resourceID] += 0.05 // each match adds slight buy pressure
	if s.State.MarketPressure[resourceID] > 1 {
		s.State.MarketPressure[resourceID] = 1
	}
	s.addLedger("market_trade", float64(fill)*execPrice, "in", map[string]any{"tradeQty": fill, "resourceId": resourceID})
	s.addLedger("market_fee", fee, "out", map[string]any{"tradeQty": fill, "resourceId": resourceID})
	s.State.Trades = append([]model.Trade{{
		ID:         fmt.Sprintf("trade-%d", s.now().UnixNano()),
		ResourceID: resourceID, Quality: buy.Quality, Quantity: fill, Price: execPrice,
		BuyOrderID: buy.ID, SellOrderID: sell.ID, CreatedAt: s.now().UTC().Format(time.RFC3339),
	}}, s.State.Trades...)
	s.addXP(company, fill/10+1)
	s.saveOrdersLocked()
	s.saveCompanyLocked(company)
	s.saveTradesLocked()
	// Track market data for competition system
	s.State.DailyTradeVolume[resourceID] += float64(fill) * execPrice
	s.State.DailyTradeQty[resourceID] += fill
	if execPrice > s.State.DailyHighPrice[resourceID] {
		s.State.DailyHighPrice[resourceID] = execPrice
	}
	if s.State.DailyLowPrice == nil {
		s.State.DailyLowPrice = map[int]float64{}
	}
	if s.State.DailyLowPrice[resourceID] == 0 || execPrice < s.State.DailyLowPrice[resourceID] {
		s.State.DailyLowPrice[resourceID] = execPrice
	}
	s.State.LastTradePrice[resourceID] = execPrice
}
