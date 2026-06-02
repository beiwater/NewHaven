package service

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"go-sim-api/internal/aml"
	"go-sim-api/internal/config"
	"go-sim-api/internal/model"
)

func (s *Service) replaceBotOrders(resourceID, quality, qty, kind int) {
	toRemove := int(math.Ceil(float64(qty) * s.Cfg.Game.BotReplacementRate))
	if toRemove <= 0 {
		return
	}
	removed := 0
	for i := len(s.State.Orders) - 1; i >= 0; i-- {
		o := &s.State.Orders[i]
		if o.ResourceID != resourceID || o.Quality != quality || o.Kind != kind || o.Status != "open" {
			continue
		}
		botIDs := map[int]bool{s.Cfg.Game.Bot1ID: true, s.Cfg.Game.Bot2ID: true}
		if !botIDs[o.CompanyID] {
			continue
		}
		if o.CompanyID == s.State.NationalTeamID {
			continue // Don't remove national team orders
		}
		// Reduce this bot order
		reduce := o.Remaining
		if reduce > toRemove-removed {
			reduce = toRemove - removed
		}
		o.Remaining -= reduce
		o.Quantity -= reduce
		if o.Remaining <= 0 {
			o.Status = "filled"
		}
		removed += reduce
		if removed >= toRemove {
			break
		}
	}
	s.saveOrdersLocked()
}

func (s *Service) CheckMarketLock(resourceID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sellQty := 0
	for _, o := range s.State.Orders {
		if o.ResourceID == resourceID && o.Kind == 0 && o.Status == "open" {
			sellQty += o.Remaining
		}
	}
	buyQty := 0
	for _, o := range s.State.Orders {
		if o.ResourceID == resourceID && o.Kind == 1 && o.Status == "open" {
			buyQty += o.Remaining
		}
	}
	totalYesterday := s.State.YesterdayVolume[resourceID] / 100
	if totalYesterday <= 0 {
		totalYesterday = 1000
	}
	threshold := s.Cfg.Game.MarketLockThreshold
	sellRatio := float64(sellQty) / totalYesterday
	buyRatio := float64(buyQty) / totalYesterday

	// 1. Price cap: market sold out on sell side
	if sellRatio < threshold && !s.State.MarketLocked[resourceID] {
		s.State.MarketLocked[resourceID] = true
		s.addLedger("market_lock_sell", 0, "out", map[string]any{"resourceId": resourceID, "sellRatio": sellRatio})
		s.deployNationalTeam(resourceID)
	}
	// 2. Price floor: market dumped on buy side
	if buyRatio < threshold && !s.State.MarketLocked[resourceID] {
		s.State.MarketLocked[resourceID] = true
		s.addLedger("market_lock_buy", 0, "out", map[string]any{"resourceId": resourceID, "buyRatio": buyRatio})
		s.deployNationalTeam(resourceID)
	}
	// 3. Price floor: price dropped too low (10% below daily low)
	lowPrice := s.State.DailyLowPrice[resourceID]
	lastPrice := s.State.LastTradePrice[resourceID]
	if lowPrice > 0 && lastPrice > 0 && lastPrice < lowPrice*0.9 && !s.State.MarketLocked[resourceID] {
		s.State.MarketLocked[resourceID] = true
		s.addLedger("market_floor", lastPrice, "out", map[string]any{"resourceId": resourceID, "lowPrice": lowPrice, "lastPrice": lastPrice})
		s.deployNationalTeam(resourceID)
	}
}

func (s *Service) deployNationalTeam(resourceID int) {
	s.State.NationalTeamActive[resourceID] = true
	yestVol := s.State.YesterdayVolume[resourceID]
	todayVol := s.State.DailyTradeVolume[resourceID]
	avgVol := (yestVol + todayVol) / 2
	if avgVol <= 0 {
		avgVol = 10000
	}
	volume := int(math.Ceil(avgVol * s.Cfg.Game.NationalTeamVolumePct))
	highPrice := s.State.YesterdayHighPrice[resourceID]
	if highPrice <= 0 {
		highPrice = s.Cfg.Game.BotOrderBase + float64(resourceID%7)
	}
	ntPrice := highPrice * s.Cfg.Game.NationalTeamPricePct
	floorPrice := highPrice * 0.80 // floor: 80% of yesterday high
	if s.State.NationalTeamID == 0 {
		s.State.NationalTeamID = 999999
	}
	now := time.Now().UTC().Format(time.RFC3339)

	// Supply at capped price (sell)
	if sellQty := volume / 2; sellQty > 0 {
		sellPrice := math.Round(ntPrice*100) / 100
		s.State.Orders = append(s.State.Orders, model.MarketOrder{
			ID:         uniqueMarketID("nt-sell", resourceID),
			ResourceID: resourceID, Kind: 0, Price: sellPrice,
			Quality: 0, Quantity: sellQty, Remaining: sellQty,
			CompanyID: s.State.NationalTeamID, CreatedAt: now, Status: "open",
		})
		s.addLedger("national_team_sell", 0, "out", map[string]any{
			"resourceId": resourceID, "qty": sellQty, "price": sellPrice, "reason": "price_cap",
		})
	}
	// Demand at floor price (buy)
	if buyQty := volume / 2; buyQty > 0 {
		buyPrice := math.Round(floorPrice*100) / 100
		s.State.Orders = append(s.State.Orders, model.MarketOrder{
			ID:         uniqueMarketID("nt-buy", resourceID),
			ResourceID: resourceID, Kind: 1, Price: buyPrice,
			Quality: 0, Quantity: buyQty, Remaining: buyQty,
			CompanyID: s.State.NationalTeamID, CreatedAt: now, Status: "open",
		})
		s.addLedger("national_team_buy", 0, "out", map[string]any{
			"resourceId": resourceID, "qty": buyQty, "price": buyPrice, "reason": "price_floor",
		})
	}
	s.saveOrdersLocked()
}

func (s *Service) ResetDailyMarket() {
	s.mu.Lock()
	defer s.mu.Unlock()
	for k, v := range s.State.DailyHighPrice {
		s.State.YesterdayHighPrice[k] = v
	}
	for k, v := range s.State.DailyTradeVolume {
		s.State.YesterdayVolume[k] = v
	}
	for k, v := range s.State.LastTradePrice {
		s.State.YesterdayClose[k] = v
	}
	s.State.DailyTradeVolume = map[int]float64{}
	s.State.DailyTradeQty = map[int]int{}
	s.State.DailyHighPrice = map[int]float64{}
	s.State.DailyLowPrice = map[int]float64{}
	s.State.MarketLocked = map[int]bool{}
	s.State.NationalTeamActive = map[int]bool{}
}

func (s *Service) RunBotMarketCycle() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !s.Cfg.BotEnabled {
		return
	}

	cycleTime := s.now().UTC().Truncate(time.Hour)
	cycleKey := cycleTime.Format(time.RFC3339)
	resources := s.marketResourceIDsLocked()
	botIDs := map[int]bool{s.Cfg.Game.Bot1ID: true, s.Cfg.Game.Bot2ID: true}
	if s.State.LastBotCycleAt == cycleKey && s.hasOpenBotOrdersForResourcesLocked(resources, botIDs) {
		return
	}
	s.State.LastBotCycleAt = cycleKey
	s.ensureBotLiquidityLocked(resources)

	now := cycleTime.Format(time.RFC3339)
	hour := cycleTime.Hour()
	cycleVol := 1.0 + s.Cfg.Game.BotCycleAmplitude*math.Sin(float64(hour)/24.0*2*math.Pi)
	spread := s.Cfg.Game.BotSpread

	filtered := make([]model.MarketOrder, 0, len(s.State.Orders))
	for _, o := range s.State.Orders {
		if !botIDs[o.CompanyID] {
			filtered = append(filtered, o)
		}
	}
	s.State.Orders = filtered

	for _, resourceID := range resources {
		prices := s.ComputeChainPrice(resourceID)
		basePrice := prices["processorPrice"] // bots operate at processor/wholesale level
		buyBase := prices["producerPrice"]

		for _, bot := range s.State.BotCompanies {
			inv := bot.Inventory[resourceID]
			target := s.Cfg.Game.BotOrderQty * 50
			pressure := float64(target-inv) / float64(target)
			buyQty := max(s.Cfg.Game.BotOrderQty/3, s.Cfg.Game.BotOrderQty/2+rand.Intn(s.Cfg.Game.BotOrderQty/2)-int(math.Max(0, pressure*float64(s.Cfg.Game.BotOrderQty/3))))
			sellQty := max(s.Cfg.Game.BotOrderQty/3, s.Cfg.Game.BotOrderQty/2+rand.Intn(s.Cfg.Game.BotOrderQty/2)-int(math.Max(0, -pressure*float64(s.Cfg.Game.BotOrderQty/3))))
			buy := model.MarketOrder{
				ID:         uniqueMarketID("bot-buy", bot.ID, resourceID),
				ResourceID: resourceID, Kind: 1, Price: math.Round(buyBase*cycleVol*(1-spread)*100) / 100,
				Quality: 0, Quantity: buyQty, Remaining: buyQty, CompanyID: bot.ID, CreatedAt: now, Status: "open",
			}
			sellInv := min(sellQty, max(0, inv))
			sell := model.MarketOrder{
				ID:         uniqueMarketID("bot-sell", bot.ID, resourceID),
				ResourceID: resourceID, Kind: 0, Price: math.Round(basePrice*cycleVol*(1+spread)*100) / 100,
				Quality: 0, Quantity: int(sellInv), Remaining: int(sellInv), CompanyID: bot.ID, CreatedAt: now, Status: "open",
			}
			orders := []model.MarketOrder{buy}
			if sellInv > 0 {
				orders = append(orders, sell)
			}
			s.State.Orders = append(orders, s.State.Orders...)
		}
		s.recordBotTradeLocked(resourceID, basePrice, cycleVol, now)
	}
	if len(s.State.Orders) > s.Cfg.Game.MaxBotOrders {
		s.State.Orders = s.State.Orders[:s.Cfg.Game.MaxBotOrders]
	}
	s.saveOrdersLocked()
	s.saveTradesLocked()
	s.saveStateLocked()
}

func (s *Service) marketResourceIDsLocked() []int {
	resources := []int{8, 9, 10, 11, 12}
	if s.Cfg.Game.BotResources == "" {
		return resources
	}
	parts := strings.Split(s.Cfg.Game.BotResources, ",")
	resources = make([]int, 0, len(parts))
	seen := map[int]bool{}
	for _, p := range parts {
		id, err := strconv.Atoi(strings.TrimSpace(p))
		if err == nil && id > 0 && !seen[id] {
			resources = append(resources, id)
			seen[id] = true
		}
	}
	return resources
}

func (s *Service) ensureBotCompaniesLocked() {
	if len(s.State.BotCompanies) == 0 {
		s.State.BotCompanies = []model.Company{
			{ID: s.Cfg.Game.Bot1ID, Name: s.Cfg.Game.Bot1Name, Money: s.Cfg.Game.BotMoney, Level: s.Cfg.Game.BotLevel, Inventory: map[int]int{}},
			{ID: s.Cfg.Game.Bot2ID, Name: s.Cfg.Game.Bot2Name, Money: s.Cfg.Game.BotMoney, Level: s.Cfg.Game.BotLevel, Inventory: map[int]int{}},
		}
	}
	for i := range s.State.BotCompanies {
		if s.State.BotCompanies[i].Inventory == nil {
			s.State.BotCompanies[i].Inventory = map[int]int{}
		}
	}
}

func (s *Service) ensureBotLiquidityLocked(resources []int) {
	s.ensureBotCompaniesLocked()
	for i := range s.State.BotCompanies {
		bot := &s.State.BotCompanies[i]
		if bot.Money < s.Cfg.Game.BotMoney/2 {
			bot.Money = s.Cfg.Game.BotMoney
		}
		for _, resourceID := range resources {
			if bot.Inventory[resourceID] < s.Cfg.Game.BotOrderQty*25 {
				bot.Inventory[resourceID] = s.Cfg.Game.BotOrderQty * 80
			}
		}
		if company := s.getCompanyLocked(bot.ID); company != nil {
			company.Money = bot.Money
			if company.Inventory == nil {
				company.Inventory = map[int]int{}
			}
			for _, resourceID := range resources {
				company.Inventory[resourceID] = bot.Inventory[resourceID]
			}
		}
	}
}

func (s *Service) hasOpenBotOrdersForResourcesLocked(resources []int, botIDs map[int]bool) bool {
	seen := map[int]bool{}
	for _, order := range s.State.Orders {
		if !botIDs[order.CompanyID] || order.Remaining <= 0 || order.Status == "filled" || order.Status == "cancelled" {
			continue
		}
		seen[order.ResourceID] = true
	}
	for _, resourceID := range resources {
		if !seen[resourceID] {
			return false
		}
	}
	return len(resources) > 0
}

func (s *Service) recordBotTradeLocked(resourceID int, basePrice, cycleVol float64, at string) {
	if len(s.State.BotCompanies) < 2 {
		return
	}
	qty := max(1, s.Cfg.Game.BotOrderQty/4+rand.Intn(max(1, s.Cfg.Game.BotOrderQty/4)))
	price := math.Round(basePrice*cycleVol*(0.99+rand.Float64()*0.02)*100) / 100
	buyer := s.State.BotCompanies[0]
	seller := s.State.BotCompanies[1]
	trade := model.Trade{
		ID:         uniqueMarketID("bot-trade", resourceID),
		ResourceID: resourceID, Quality: 0, Quantity: qty, Price: price,
		BuyOrderID:  fmt.Sprintf("bot-cycle-buy-%d", resourceID),
		SellOrderID: fmt.Sprintf("bot-cycle-sell-%d", resourceID),
		CreatedAt:   at,
	}
	s.State.Trades = append([]model.Trade{trade}, s.State.Trades...)
	if len(s.State.Trades) > 1000 {
		s.State.Trades = s.State.Trades[:1000]
	}
	s.State.BotCompanies[0].Inventory[resourceID] += qty
	s.State.BotCompanies[1].Inventory[resourceID] = max(0, s.State.BotCompanies[1].Inventory[resourceID]-qty)
	s.updateTradeMarketData(resourceID, qty, price)
	s.AML.RecordTransaction(aml.Transaction{
		ID: trade.ID, FromID: buyer.ID, ToID: seller.ID,
		Amount: float64(qty) * price, ResourceID: resourceID,
		Type: "bot_market_trade", Timestamp: s.now(),
	})
}

func botInventoryFromConfig(cfg *config.Config, qty int) map[int]int {
	inventory := map[int]int{}
	for _, part := range strings.Split(cfg.Game.BotResources, ",") {
		id, err := strconv.Atoi(strings.TrimSpace(part))
		if err == nil && id > 0 {
			inventory[id] = qty
		}
	}
	if len(inventory) == 0 {
		inventory[8] = qty
		inventory[9] = qty
	}
	return inventory
}
