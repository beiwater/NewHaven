package service

import (
	"fmt"
	"slices"
	"time"

	"go-sim-api/internal/aml"
	"go-sim-api/internal/anticheat"
	"go-sim-api/internal/formula"
	"go-sim-api/internal/model"
)

func (s *Service) TakeOrder(companyID int, resourceID, quantity, quality int, maxPrice float64) (map[string]any, error) {
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
	s.AC.RecordAction(pid, anticheat.ActTakeOrder, fmt.Sprintf("res=%d qty=%d maxPrice=%.2f", resourceID, quantity, maxPrice))
	s.SD.RecordAction(pid)

	if quantity <= 0 || maxPrice <= 0 {
		return nil, fmt.Errorf("invalid payload")
	}
	sells := make([]*model.MarketOrder, 0)
	for i := range s.State.Orders {
		o := &s.State.Orders[i]
		if o.ResourceID == resourceID && o.Quality == quality && o.Kind == 0 && o.Remaining > 0 && o.Price <= maxPrice {
			sells = append(sells, o)
		}
	}
	slices.SortFunc(sells, func(a, b *model.MarketOrder) int {
		if a.Price < b.Price {
			return -1
		}
		if a.Price > b.Price {
			return 1
		}
		return 0
	})
	need := quantity
	bought := 0
	spent := 0.0
	trades := []model.Trade{}
	for _, sell := range sells {
		if need == 0 {
			break
		}
		fill, cost, trade := s.executeTakeFill(company, sell, resourceID, quality, need)
		if trade == nil {
			break
		}
		need -= fill
		bought += fill
		spent += cost
		trades = append(trades, *trade)
	}
	s.saveOrdersLocked()
	s.saveCompanyLocked(company)
	s.saveTradesLocked()
	s.State.MarketPressure[resourceID] += float64(bought) * 0.02
	if s.State.MarketPressure[resourceID] > 1 {
		s.State.MarketPressure[resourceID] = 1
	}
	return map[string]any{
		"amountBought": bought, "trades": trades, "moneyDelta": -spent,
	}, nil
}

// executeTakeFill fulfills a single fill against a sell order.
// Returns (fill, cost, trade). Returns nil trade when money runs out.
// Caller holds s.mu.
func (s *Service) executeTakeFill(company *model.Company, sell *model.MarketOrder, resourceID, quality, need int) (int, float64, *model.Trade) {
	fill := min(need, sell.Remaining)
	fee := formula.ExchangeFee(fill, sell.Price, s.Cfg.Game.ExchangeFeePct)
	cost := float64(fill)*sell.Price + fee
	if company.Money < cost {
		return 0, 0, nil
	}
	company.Money -= cost
	s.addLedger("market_take_buy", cost, "out", map[string]any{"resourceId": resourceID, "qty": fill})
	s.inventoryAdd(company, resourceID, quality, fill)
	sell.Remaining -= fill
	sellCompany := s.getCompanyLocked(sell.CompanyID)
	if sellCompany != nil && sellCompany.ID != company.ID {
		sellCompany.Money += float64(fill) * sell.Price
	}
	if sell.Remaining == 0 {
		sell.Status = "filled"
	}
	s.addXP(company, fill/10+1)
	trade := &model.Trade{
		ID:         fmt.Sprintf("trade-%d", s.now().UnixNano()),
		ResourceID: resourceID, Quality: quality, Quantity: fill, Price: sell.Price,
		BuyOrderID: "direct-take", SellOrderID: sell.ID, CreatedAt: s.now().UTC().Format(time.RFC3339),
	}
	s.State.Trades = append([]model.Trade{*trade}, s.State.Trades...)
	s.AML.RecordTransaction(aml.Transaction{
		ID: trade.ID, FromID: company.ID,
		ToID: sell.CompanyID, Amount: float64(fill) * sell.Price,
		ResourceID: resourceID, Type: "market_trade",
		Timestamp: s.now(),
	})
	if ok, msg := s.AML.CheckRapidTrades(company.ID); !ok {
		s.AC.AddAlert(company.ID, "rapid_trading", msg, "low")
	}
	s.AML.DetectRoundTrip(company.ID, sell.CompanyID, float64(fill)*sell.Price, resourceID)
	s.updateTradeMarketData(resourceID, fill, sell.Price)
	return fill, cost, trade
}

// updateTradeMarketData tracks daily market stats after a trade.
func (s *Service) updateTradeMarketData(resourceID, qty int, price float64) {
	s.State.DailyTradeVolume[resourceID] += float64(qty) * price
	s.State.DailyTradeQty[resourceID] += qty
	if price > s.State.DailyHighPrice[resourceID] {
		s.State.DailyHighPrice[resourceID] = price
	}
	if s.State.DailyLowPrice == nil {
		s.State.DailyLowPrice = map[int]float64{}
	}
	if s.State.DailyLowPrice[resourceID] == 0 || price < s.State.DailyLowPrice[resourceID] {
		s.State.DailyLowPrice[resourceID] = price
	}
	s.State.LastTradePrice[resourceID] = price
}
