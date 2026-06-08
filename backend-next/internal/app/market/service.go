package market

import (
	"context"
	"fmt"
	"github.com/newhaven/backend-next/internal/catalog"
	"github.com/newhaven/backend-next/internal/config"
	domainmarket "github.com/newhaven/backend-next/internal/domain/market"
	openapi "github.com/newhaven/backend-next/internal/generated/openapi"
	"github.com/newhaven/backend-next/internal/platform"
	"github.com/newhaven/backend-next/internal/storage"
	"math"
	"sort"
	"sync"
	"time"
)

// Service is the market application use case.
type Service struct {
	mu           sync.Mutex
	market       storage.MarketStorage
	companies    storage.CompanyStorage
	finance      storage.FinanceStorage
	resources    map[int]*catalog.ResourceEntry
	buildings    map[int]*catalog.BuildingEntry
	economy      map[int]*catalog.EconomyModelEntry
	cfg          *config.GameConfig
	clock        platform.Clock
	idgen        *platform.IDGen
	botCompanyID int // set by EnsureBotCompanies, used by RunBotCycle
}

func NewService(market storage.MarketStorage, companies storage.CompanyStorage, finance storage.FinanceStorage, resources map[int]*catalog.ResourceEntry, buildings map[int]*catalog.BuildingEntry, economy map[int]*catalog.EconomyModelEntry, cfg *config.GameConfig, clock platform.Clock, idgen *platform.IDGen) *Service {
	return &Service{
		market:    market,
		companies: companies,
		finance:   finance,
		resources: resources,
		buildings: buildings,
		economy:   economy,
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

// GetTickers returns ticker data for all market-tradable resources.
func (s *Service) GetTickers(ctx context.Context) ([]openapi.MarketTickerResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := s.clock.Now().UTC().Truncate(time.Hour)

	// Collect all exchange-tradable resource IDs.
	resourceIDs := make([]int, 0, len(s.resources))
	for _, r := range s.resources {
		if r.DbLetter <= 0 || r.IsResearch || !r.IsExchangeTradable {
			continue
		}
		resourceIDs = append(resourceIDs, r.DbLetter)
	}
	sort.Ints(resourceIDs)

	tickers := make([]openapi.MarketTickerResponse, 0, len(resourceIDs))
	for _, rid := range resourceIDs {
		tickers = append(tickers, s.buildTickerResponse(ctx, rid, now))
	}

	return tickers, nil
}

// buildTickerResponse builds a single ticker response without locking.
// Caller must hold s.mu.
func (s *Service) buildTickerResponse(ctx context.Context, resourceID int, now time.Time) openapi.MarketTickerResponse {
	// Try reading from storage first.
	ticker, err := s.market.GetTicker(ctx, resourceID)
	if err == nil && ticker != nil {
		series := make([]openapi.MarketTickerPoint, 48)
		for i := 47; i >= 0; i-- {
			ts := now.Add(-time.Duration(i) * time.Hour)
			price32 := float32(ticker.LastPrice)
			series[47-i] = openapi.MarketTickerPoint{
				Price: &price32,
				Time:  &ts,
			}
		}
		return openapi.MarketTickerResponse{
			Resource: &resourceID,
			Series:   &series,
		}
	}

	// Fallback: synthesize deterministic series from catalog BasePrice.
	basePrice := s.basePriceForResource(resourceID)
	if basePrice <= 0 {
		basePrice = 20.0 + float64(resourceID%11)*3.0
	}

	series := make([]openapi.MarketTickerPoint, 48)
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

	return openapi.MarketTickerResponse{
		Resource: &resourceID,
		Series:   &series,
	}
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

// ListMyOrders returns all open/partial orders for a company across all resources.
func (s *Service) ListMyOrders(ctx context.Context, companyID int) (*openapi.MarketOrderListResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	orders, err := s.market.GetOrdersByCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}

	dtos := make([]openapi.MarketOrderDTO, 0)
	for i := range orders {
		o := orders[i]
		if o.Status != domainmarket.StatusOpen && o.Status != domainmarket.StatusPartial {
			continue
		}
		remaining := o.Remaining()
		if remaining <= 0 {
			continue
		}
		kind := 0
		if o.IsBuy {
			kind = 1
		}
		kindVal := openapi.MarketOrderDTOKind(kind)

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
	for i := range orders {
		o := orders[i]
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
		kind := 0
		if o.IsBuy {
			kind = 1
		}
		kindVal := openapi.MarketOrderDTOKind(kind)


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
