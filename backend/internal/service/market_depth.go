package service

import "sort"

type depthLevel struct {
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
	Qty      int     `json:"qty"`
}

// OrderBookDepth returns top 5 buy and sell orders aggregated by price
func (s *Service) OrderBookDepth(resourceID, quality int) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	buys := map[float64]int{}
	sells := map[float64]int{}
	for _, o := range s.State.Orders {
		if o.ResourceID != resourceID || o.Quality != quality || o.Remaining <= 0 {
			continue
		}
		if o.Kind == 1 {
			buys[o.Price] += o.Remaining
		} else {
			sells[o.Price] += o.Remaining
		}
	}
	// Sort and take top 5
	buyPrices := sortKeysDesc(buys)
	sellPrices := sortKeysAsc(sells)
	if len(buyPrices) > 5 {
		buyPrices = buyPrices[:5]
	}
	if len(sellPrices) > 5 {
		sellPrices = sellPrices[:5]
	}
	buyLevels := make([]depthLevel, 0, len(buyPrices))
	for _, p := range buyPrices {
		buyLevels = append(buyLevels, depthLevel{Price: p, Quantity: buys[p], Qty: buys[p]})
	}
	sellLevels := make([]depthLevel, 0, len(sellPrices))
	for _, p := range sellPrices {
		sellLevels = append(sellLevels, depthLevel{Price: p, Quantity: sells[p], Qty: sells[p]})
	}
	return map[string]any{"buys": buyLevels, "sells": sellLevels}
}

func sortKeysDesc(m map[float64]int) []float64 {
	k := make([]float64, 0, len(m))
	for i := range m {
		k = append(k, i)
	}
	sort.Slice(k, func(a, b int) bool { return k[a] > k[b] })
	return k
}

func sortKeysAsc(m map[float64]int) []float64 {
	k := make([]float64, 0, len(m))
	for i := range m {
		k = append(k, i)
	}
	sort.Slice(k, func(a, b int) bool { return k[a] < k[b] })
	return k
}
