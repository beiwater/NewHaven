package market

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"math/rand"
	"sort"
	"time"

	"github.com/beiwater/NewHaven/backend/internal/domain/company"
	domainmarket "github.com/beiwater/NewHaven/backend/internal/domain/market"
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
	spreadPct:        0.03,
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
// Places up to 3 buy + 3 sell levels per selected resource with dynamic spread,
// inventory feedback, stale order cleanup, and random drift.
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

	ids := s.tradableResourceIDs()
	if len(ids) == 0 {
		return nil
	}
	sort.Ints(ids)

	rng := rand.New(rand.NewSource(s.clock.Now().UnixNano()))
	n := botCfg.resourcesPerTick
	if n > len(ids) {
		n = len(ids)
	}
	// Round-robin selection — ensures every resource gets regular coverage.
	tickNum := s.clock.Now().Unix() / 60
	startIdx := (int(tickNum) * n) % len(ids)
	selected := make([]int, n)
	for i := 0; i < n; i++ {
		selected[i] = ids[(startIdx+i)%len(ids)]
	}

	now := s.clock.Now().UTC()

	for _, resourceID := range selected {
		s.processResourceCycle(ctx, rng, now, resourceID)
	}
	return nil
}

// EnsureMarketLiquidity gives every exchange-tradable resource a two-sided
// bot book. It is safe to call whenever the market is opened: existing healthy
// bot levels count against the target, so the call replenishes gaps without
// multiplying orders on every page visit.
func (s *Service) EnsureMarketLiquidity(ctx context.Context) error {
	if !botCfg.enabled {
		return nil
	}
	if err := s.EnsureBotCompanies(ctx); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	ids := s.tradableResourceIDs()
	rng := rand.New(rand.NewSource(s.clock.Now().UnixNano()))
	now := s.clock.Now().UTC()
	for _, resourceID := range ids {
		s.processResourceCycle(ctx, rng, now, resourceID)
	}
	return nil
}

func (s *Service) tradableResourceIDs() []int {
	ids := make([]int, 0, len(s.resources))
	for _, resource := range s.resources {
		if resource == nil || resource.DbLetter <= 0 || resource.IsResearch || !resource.IsExchangeTradable {
			continue
		}
		ids = append(ids, resource.DbLetter)
	}
	sort.Ints(ids)
	return ids
}

// processResourceCycle handles a single resource's bot cycle iteration.
func (s *Service) processResourceCycle(ctx context.Context, rng *rand.Rand, now time.Time, resourceID int) {
	// Step 1: Gather market data
	allOrders, err := s.market.GetOrdersByResource(ctx, resourceID)
	if err != nil {
		return
	}

	fairPrice, spread := s.calculateFairPrice(ctx, rng, resourceID, allOrders)
	netPos := s.calculateBotNetPosition(allOrders, s.botCompanyID)
	wantedBuy, wantedSell := s.cancelStaleBotOrders(ctx, allOrders, fairPrice, spread)
	if wantedBuy == 0 && wantedSell == 0 {
		return
	}
	basePrice, drift, inventoryOffset := s.calculateBotBasePrice(rng, resourceID, fairPrice, spread, netPos, float64(s.clock.Now().Unix()))
	s.placeBotOrderLevels(ctx, rng, now, resourceID, basePrice, spread, wantedBuy, wantedSell, fairPrice, netPos, drift, inventoryOffset)
}

// calculateFairPrice computes the fair price and dynamic spread from ticker/order book data.
func (s *Service) calculateFairPrice(
	ctx context.Context,
	rng *rand.Rand,
	resourceID int,
	allOrders []domainmarket.MarketOrder,
) (fairPrice float64, spread float64) {
	ticker, _ := s.market.GetTicker(ctx, resourceID)

	// Calculate fairPrice
	if ticker != nil && ticker.LastPrice > 0 {
		fairPrice = ticker.LastPrice
	} else {
		var highestBuy, lowestSell float64
		lowestSell = math.MaxFloat64
		for _, o := range allOrders {
			if o.Status != domainmarket.StatusOpen && o.Status != domainmarket.StatusPartial {
				continue
			}
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
			fairPrice = (highestBuy + lowestSell) / 2
		} else if highestBuy > 0 {
			fairPrice = highestBuy * 1.05
		} else if lowestSell < math.MaxFloat64 {
			fairPrice = lowestSell * 0.95
		} else {
			fairPrice = s.basePriceForResource(resourceID)
			if fairPrice <= 0 {
				fairPrice = 20.0 + float64(resourceID%11)*3.0
			}
			jitter := 1.0 + (rng.Float64()-0.5)*0.1
			fairPrice *= jitter
		}
	}

	// Dynamic spread
	spread = 0.06
	if ticker != nil && ticker.Volume24h > 0 {
		spread = 0.03
	}
	return
}

// calculateBotNetPosition computes net position from filled quantities of bot orders.
func (s *Service) calculateBotNetPosition(allOrders []domainmarket.MarketOrder, botCompanyID int) int {
	netPos := 0
	for _, o := range allOrders {
		if o.CompanyID != botCompanyID {
			continue
		}
		if o.IsBuy {
			netPos += o.FilledQuantity
		} else {
			netPos -= o.FilledQuantity
		}
	}
	return netPos
}

// cancelStaleBotOrders cancels bot orders too far from fair price and returns remaining slot budgets.
func (s *Service) cancelStaleBotOrders(ctx context.Context, allOrders []domainmarket.MarketOrder, fairPrice float64, spread float64) (wantedBuy, wantedSell int) {
	wantedBuy = 3
	wantedSell = 3
	for i := range allOrders {
		o := &allOrders[i]
		if o.CompanyID != s.botCompanyID {
			continue
		}
		if o.Status != domainmarket.StatusOpen && o.Status != domainmarket.StatusPartial {
			continue
		}
		if o.Remaining() <= 0 {
			continue
		}
		// Check if this order's price is too far from fair price
		if math.Abs(o.Price-fairPrice)/fairPrice > 3.0*spread {
			s.cancelBotOrderLocked(ctx, o)
			continue
		}
		// Valid bot order — counts against budget
		if o.IsBuy {
			wantedBuy--
		} else {
			wantedSell--
		}
	}
	if wantedBuy < 0 {
		wantedBuy = 0
	}
	if wantedSell < 0 {
		wantedSell = 0
	}
	return
}

// calculateBotBasePrice computes the base price with sine-wave cycle, inventory feedback, and random drift.
func (s *Service) calculateBotBasePrice(rng *rand.Rand, resourceID int, fairPrice float64, spread float64, netPos int, nowSec float64) (basePrice float64, drift float64, inventoryOffset float64) {
	// Step 4: Market cycle wave (sine wave superposition, stateless, time-driven)
	phaseA := float64(resourceID) * 1.7 // per-resource phase
	phaseB := float64(resourceID) * 2.3
	periodA := 1200.0 // 20 min primary
	periodB := 420.0  // 7 min secondary
	waveA := math.Sin(2*math.Pi*nowSec/periodA + phaseA)
	waveB := math.Sin(2*math.Pi*nowSec/periodB + phaseB)
	cycleOffset := 0.03 * (0.65*waveA + 0.35*waveB) // ±3% max, blended

	// Step 5: Inventory feedback and random drift
	inventoryOffset = -float64(netPos) * 0.002
	drift = (rng.Float64() - 0.5) * fairPrice * 0.01
	basePrice = fairPrice*(1+cycleOffset) + inventoryOffset + drift
	if basePrice < 0.01 {
		basePrice = 0.01
	}
	return
}

// placeBotOrderLevels places buy and sell order levels for a resource.
func (s *Service) placeBotOrderLevels(ctx context.Context, rng *rand.Rand, now time.Time, resourceID int, basePrice float64, spread float64, wantedBuy, wantedSell int, fairPrice float64, netPos int, drift, inventoryOffset float64) {
	// Base size for level 1
	size1 := botCfg.orderSizeMin + rng.Intn(botCfg.orderSizeMax-botCfg.orderSizeMin+1)
	size2 := int(float64(size1) * 1.5)
	size3 := int(float64(size1) * 2.5)

	// Level prices
	buyPrices := [3]float64{
		basePrice * (1 - spread),
		basePrice * (1 - 1.5*spread),
		basePrice * (1 - 2.5*spread),
	}
	sellPrices := [3]float64{
		basePrice * (1 + spread),
		basePrice * (1 + 1.5*spread),
		basePrice * (1 + 2.5*spread),
	}
	buySizes := [3]int{size1, size2, size3}
	sellSizes := [3]int{size1, size2, size3}

	// Place buy levels
	for i := 0; i < 3 && i < wantedBuy; i++ {
		price := snapToMarketTick(buyPrices[i])
		if price < 0.01 {
			price = 0.01
		}
		size := buySizes[i]
		if size <= 0 {
			continue
		}
		buyOrder := &domainmarket.MarketOrder{
			ID:         s.idgen.Next("bot-buy"),
			CompanyID:  s.botCompanyID,
			ResourceID: resourceID,
			IsBuy:      true,
			Price:      price,
			Quantity:   size,
			Quality:    0,
			Status:     domainmarket.StatusOpen,
			CreatedAt:  now.Format(time.RFC3339),
		}
		if err := s.market.CreateOrder(ctx, buyOrder); err != nil {
			slog.Warn("[bot] create buy order failed", "resource", resourceID, "price", price, "error", err)
			continue
		}
		s.matchNewBuyOrder(ctx, buyOrder)
	}

	// Place sell levels
	for i := 0; i < 3 && i < wantedSell; i++ {
		price := snapToMarketTick(sellPrices[i])
		if price < 0.01 {
			price = 0.01
		}
		size := sellSizes[i]
		if size <= 0 {
			continue
		}
		sellOrder := &domainmarket.MarketOrder{
			ID:         s.idgen.Next("bot-sell"),
			CompanyID:  s.botCompanyID,
			ResourceID: resourceID,
			IsBuy:      false,
			Price:      price,
			Quantity:   size,
			Quality:    0,
			Status:     domainmarket.StatusOpen,
			CreatedAt:  now.Format(time.RFC3339),
		}
		if err := s.market.CreateOrder(ctx, sellOrder); err != nil {
			slog.Warn("[bot] create sell order failed", "resource", resourceID, "price", price, "error", err)
			continue
		}
		s.matchNewSellOrder(ctx, sellOrder)
	}

	slog.Debug("bot orders placed",
		"resource", resourceID,
		"fairPrice", fairPrice,
		"spread", spread,
		"netPos", netPos,
		"drift", drift,
		"inventoryOffset", inventoryOffset,
		"buyPrices", buyPrices,
		"sellPrices", sellPrices,
	)
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
	basePrice := s.basePriceForResource(resourceID)
	if basePrice <= 0 {
		basePrice = 20.0 + float64(resourceID%11)*3.0
	}
	jitter := 1.0 + (rng.Float64()-0.5)*0.2
	return basePrice * jitter
}

// cancelBotOrderLocked cancels a bot order and refunds reserved funds/inventory.
// Called under s.mu (from RunBotCycle). Does NOT acquire the lock.
func (s *Service) cancelBotOrderLocked(ctx context.Context, order *domainmarket.MarketOrder) {
	if order.Status == domainmarket.StatusFilled || order.Status == domainmarket.StatusCancelled {
		return
	}
	remaining := order.Remaining()
	order.Status = domainmarket.StatusCancelled
	if err := s.market.UpdateOrder(ctx, order); err != nil {
		slog.Warn("[bot] cancel order failed", "id", order.ID, "error", err)
		return
	}
	// Refund reserved funds/inventory.
	if order.IsBuy {
		refund := order.Price * float64(remaining)
		_, _ = s.companies.AdjustMoney(ctx, order.CompanyID, refund, false)
	} else {
		_ = s.companies.UpdateInventory(ctx, order.CompanyID, order.ResourceID, remaining)
	}
}
