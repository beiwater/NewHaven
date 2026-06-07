package market

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/newhaven/backend-next/internal/domain/company"
	domainmarket "github.com/newhaven/backend-next/internal/domain/market"
)

// botConfig holds tuning parameters for bot market activity.
type botConfig struct {
	enabled          bool
	spreadPct        float64
	orderSizeMin     int
	orderSizeMax     int
	resourcesPerTick int
}

var botCfg = botConfig{
	enabled:          true,
	spreadPct:        0.05,
	orderSizeMin:     50,
	orderSizeMax:     500,
	resourcesPerTick: 3,
}

// EnsureBotCompanies creates the bot player and company needed for market liquidity.
// Must be called once before RunBotCycle. Idempotent.
func (s *Service) EnsureBotCompanies(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.botCompanyID != 0 {
		return nil // already initialized
	}

	const botPlayerID = -900001 // sentinel; no real player uses this

	// Check if already created by a previous run.
	existing, err := s.companies.GetCompanyByPlayerID(ctx, botPlayerID)
	if err == nil && existing != nil {
		s.botCompanyID = existing.ID
		// Top up inventory for any new resources.
		for rid := range s.resources {
			if existing.Inventory == nil {
				existing.Inventory = make(map[int]int)
			}
			existing.Inventory[rid] = 999999
		}
		_ = s.companies.UpdateCompany(ctx, existing)
		return nil
	}

	bot := &company.Company{
		PlayerID:  botPlayerID,
		Name:      "Atlas Trading Bot",
		Money:     5_000_000,
		Level:     99,
		XP:        0,
		Inventory: make(map[int]int),
	}
	for rid := range s.resources {
		bot.Inventory[rid] = 999999
	}
	if err := s.companies.CreateCompany(ctx, bot); err != nil {
		return fmt.Errorf("create bot company: %w", err)
	}
	s.botCompanyID = bot.ID
	slog.Info("bot company created", "company_id", s.botCompanyID)
	return nil
}

// RunBotCycle generates NPC buy/sell orders to maintain market liquidity.
func (s *Service) RunBotCycle(ctx context.Context) error {
	if !botCfg.enabled {
		return nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if s.botCompanyID == 0 {
		slog.Warn("[bot] company not initialized, skipping cycle")
		return nil
	}

	var ids []int
	for id := range s.resources {
		ids = append(ids, id)
	}
	if len(ids) == 0 {
		return nil
	}
	sort.Ints(ids)

	rng := rand.New(rand.NewSource(s.clock.Now().UnixNano()))
	n := botCfg.resourcesPerTick
	if n > len(ids) {
		n = len(ids)
	}
	rng.Shuffle(len(ids), func(i, j int) { ids[i], ids[j] = ids[j], ids[i] })
	selected := ids[:n]

	now := s.clock.Now().UTC()

	for _, resourceID := range selected {
		midPrice := s.estimateMidPrice(ctx, resourceID, rng)
		if midPrice <= 0 {
			continue
		}

		existing, _ := s.market.GetOrdersByResource(ctx, resourceID)
		botOrderCount := 0
		for _, o := range existing {
			if o.CompanyID == s.botCompanyID && o.Remaining() > 0 {
				botOrderCount++
			}
		}
		if botOrderCount >= 4 {
			continue
		}

		buyPrice := midPrice * (1 - botCfg.spreadPct)
		buyQty := botCfg.orderSizeMin + rng.Intn(botCfg.orderSizeMax-botCfg.orderSizeMin+1)
		buyOrder := &domainmarket.MarketOrder{
			ID:         s.idgen.Next("bot-buy"),
			CompanyID:  s.botCompanyID,
			ResourceID: resourceID,
			IsBuy:      true,
			Price:      buyPrice,
			Quantity:   buyQty,
			Quality:    0,
			Status:     domainmarket.StatusOpen,
			CreatedAt:  now.Format(time.RFC3339),
		}

		sellPrice := midPrice * (1 + botCfg.spreadPct)
		sellQty := botCfg.orderSizeMin + rng.Intn(botCfg.orderSizeMax-botCfg.orderSizeMin+1)
		sellOrder := &domainmarket.MarketOrder{
			ID:         s.idgen.Next("bot-sell"),
			CompanyID:  s.botCompanyID,
			ResourceID: resourceID,
			IsBuy:      false,
			Price:      sellPrice,
			Quantity:   sellQty,
			Quality:    0,
			Status:     domainmarket.StatusOpen,
			CreatedAt:  now.Format(time.RFC3339),
		}

		_ = s.market.CreateOrder(ctx, buyOrder)
		_ = s.market.CreateOrder(ctx, sellOrder)
		s.matchNewBuyOrder(ctx, buyOrder)
		s.matchNewSellOrder(ctx, sellOrder)

		slog.Debug("bot orders placed",
			"resource", resourceID,
			"buy_qty", buyQty,
			"buy_price", buyPrice,
			"sell_qty", sellQty,
			"sell_price", sellPrice,
		)
	}
	return nil
}

// MatchAllOrders iterates over all open orders and attempts to match them.
func (s *Service) MatchAllOrders(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var ids []int
	for id := range s.resources {
		ids = append(ids, id)
	}
	sort.Ints(ids)

	for _, resourceID := range ids {
		allOrders, err := s.market.GetOrdersByResource(ctx, resourceID)
		if err != nil {
			continue
		}
		for i := range allOrders {
			if allOrders[i].IsBuy && allOrders[i].Remaining() > 0 && allOrders[i].Status == domainmarket.StatusOpen {
				s.matchNewBuyOrder(ctx, &allOrders[i])
			}
		}
		allOrders, _ = s.market.GetOrdersByResource(ctx, resourceID)
		for i := range allOrders {
			if !allOrders[i].IsBuy && allOrders[i].Remaining() > 0 && allOrders[i].Status == domainmarket.StatusOpen {
				s.matchNewSellOrder(ctx, &allOrders[i])
			}
		}
	}
	return nil
}

// RefreshDailyOrders generates fresh daily orders (stub).
func (s *Service) RefreshDailyOrders(_ context.Context) error {
	return nil
}

// CleanupMarket removes stale orders (filled/cancelled/expired) and refreshes daily orders.
func (s *Service) CleanupMarket(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return nil
}

// estimateMidPrice estimates the mid-market price for a resource.
func (s *Service) estimateMidPrice(ctx context.Context, resourceID int, rng *rand.Rand) float64 {
	orders, err := s.market.GetOrdersByResource(ctx, resourceID)
	if err != nil {
		return 0
	}
	var lowestSell, highestBuy float64
	lowestSell = math.MaxFloat64
	for _, o := range orders {
		if o.Remaining() <= 0 {
			continue
		}
		if o.IsBuy && o.Price > highestBuy {
			highestBuy = o.Price
		}
		if !o.IsBuy && o.Price < lowestSell {
			lowestSell = o.Price
		}
	}
	if highestBuy > 0 && lowestSell < math.MaxFloat64 {
		return (highestBuy + lowestSell) / 2
	}
	if highestBuy > 0 {
		return highestBuy * 1.1
	}
	if lowestSell < math.MaxFloat64 {
		return lowestSell * 0.9
	}
	ticker, err := s.market.GetTicker(ctx, resourceID)
	if err == nil && ticker != nil {
		return float64(ticker.LastPrice)
	}
	basePrice := 10.0 + float64(resourceID)*5.0
	jitter := 1.0 + (rng.Float64()-0.5)*0.2
	return basePrice * jitter
}
