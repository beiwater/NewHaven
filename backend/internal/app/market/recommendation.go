package market

import (
	"context"
	"math"

	domainmarket "github.com/beiwater/NewHaven/backend/internal/domain/market"
	"github.com/beiwater/NewHaven/backend/internal/formula"
)

type priceRecommendation struct {
	Fair float64
	Buy  float64
	Sell float64
}

// recommendedPrices derives player-facing defaults from the live order book.
// A catalog base price is the final fallback, so an empty market never sends a
// player back to a hard-coded $10 quote.
func (s *Service) recommendedPrices(ctx context.Context, resourceID int) priceRecommendation {
	orders, _ := s.market.GetOrdersByResource(ctx, resourceID)
	bestBid := 0.0
	bestAsk := math.MaxFloat64
	for _, order := range orders {
		if order.Status != domainmarket.StatusOpen && order.Status != domainmarket.StatusPartial {
			continue
		}
		if order.Remaining() <= 0 || order.Quality != 0 {
			continue
		}
		if order.IsBuy && order.Price > bestBid {
			bestBid = order.Price
		}
		if !order.IsBuy && order.Price < bestAsk {
			bestAsk = order.Price
		}
	}

	fair := 0.0
	if bestBid > 0 && bestAsk < math.MaxFloat64 {
		fair = (bestBid + bestAsk) / 2
	}
	if fair <= 0 {
		if ticker, err := s.market.GetTicker(ctx, resourceID); err == nil && ticker != nil && ticker.LastPrice > 0 {
			fair = ticker.LastPrice
		}
	}
	if fair <= 0 {
		fair = s.basePriceForResource(resourceID)
	}
	if fair <= 0 {
		fair = 20 + float64(resourceID%11)*3
	}

	buy := fair
	if bestAsk < math.MaxFloat64 {
		buy = bestAsk
	}
	sell := fair
	if bestBid > 0 {
		sell = bestBid
	}
	return priceRecommendation{
		Fair: snapToMarketTick(fair),
		Buy:  snapToMarketTick(buy),
		Sell: snapToMarketTick(sell),
	}
}

func snapToMarketTick(price float64) float64 {
	if price <= 0 || math.IsNaN(price) || math.IsInf(price, 0) {
		return 0.001
	}
	step := formula.TickStep(price)
	price = math.Round(price/step) * step
	// Rounding can cross a price band, so normalize once more using the final
	// band's tick size.
	step = formula.TickStep(price)
	return math.Round(price/step) * step
}
