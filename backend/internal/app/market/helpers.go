package market

import "sort"

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
