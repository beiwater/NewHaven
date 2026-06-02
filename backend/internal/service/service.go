package service

import (
	"context"
	"fmt"
	"go-sim-api/internal/aml"
	"go-sim-api/internal/anticheat"
	"go-sim-api/internal/config"
	"go-sim-api/internal/data"
	"go-sim-api/internal/model"
	"go-sim-api/internal/storage"
	"sync"
	"time"
)

type Service struct {
	AML   *aml.AMLSystem
	mu    sync.Mutex
	Cfg   *config.Config
	Data  *data.StaticData
	Store storage.Storage
	AC    *anticheat.AntiCheat
	SD    *anticheat.ScriptDetector
	State model.GameState
}

// now returns SimulatedAt if set, otherwise real time.
func (s *Service) now() time.Time {
	if s.State.SimulatedAt != "" {
		t, err := time.Parse(time.RFC3339, s.State.SimulatedAt)
		if err == nil {
			return t
		}
	}
	return time.Now()
}

func (s *Service) SetSimulatedAt(t string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.State.SimulatedAt = t
}
func New(d *data.StaticData, cfg *config.Config, st storage.Storage) *Service {
	now := time.Now().UTC().Format(time.RFC3339)

	// Try to load state from storage
	if st != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if loaded, err := st.LoadState(ctx); err == nil && loaded != nil && len(loaded.Companies) > 0 {
			s := &Service{
				Cfg:   cfg,
				Data:  d,
				Store: st,
				AC:    anticheat.New(cfg.ACEnabled),
				AML:   aml.New(cfg.AMLEnabled),
				SD:    anticheat.NewScriptDetector(cfg.ScriptDetectEnabled),
				State: *loaded,
			}
			s.ensureRuntimeState()
			return s
		}
	}

	mainCompany := model.Company{
		ID: cfg.Game.CompanyID, Name: cfg.Game.CompanyName, Money: cfg.Game.StartMoney, Level: cfg.Game.StartLevel,
		Inventory:         map[int]int{1: 2000, 2: 1800, 3: 900, 8: 300, 9: 120},
		PlacedBuildings:   []map[string]any{{"id": "b-1", "kind": 2, "level": 3, "busy": false, "baseCost": 10000, "x": 2, "y": 1, "placedAt": now}},
		UnplacedBuildings: []map[string]any{},
		WarehouseLevel:    0,
	}
	botA := model.Company{ID: cfg.Game.Bot1ID, Name: cfg.Game.Bot1Name, Money: cfg.Game.BotMoney, Level: cfg.Game.BotLevel, Inventory: botInventoryFromConfig(cfg, 24000)}
	botB := model.Company{ID: cfg.Game.Bot2ID, Name: cfg.Game.Bot2Name, Money: cfg.Game.BotMoney, Level: cfg.Game.BotLevel, Inventory: botInventoryFromConfig(cfg, 18000)}
	return &Service{
		Cfg:   cfg,
		Data:  d,
		Store: st,
		AC:    anticheat.New(cfg.ACEnabled),
		AML:   aml.New(cfg.AMLEnabled),
		SD:    anticheat.NewScriptDetector(cfg.ScriptDetectEnabled),
		State: model.GameState{
			Players:             []model.Player{},
			NextPlayerID:        1,
			Companies:           []model.Company{mainCompany, botA, botB},
			CSRFToken:           cfg.CSRFToken,
			Messages:            []model.Message{{ID: "m-1", Chatroom: "N", Body: "Welcome to Sim API clone.", At: now}},
			Bonds:               []model.Bond{{ID: "bond-1", Amount: 5000, Interest: 0.012, PurchasedAt: now, MissedPayments: 0, InterestCollected: 2.4, RatingWhenPurchased: "A", IssuerCompanyID: cfg.Game.Bot1ID, OwnerCompanyID: cfg.Game.CompanyID, RestructurePct: 0.0}},
			ContractsIn:         []map[string]any{},
			ContractsOut:        []map[string]any{},
			Achievements:        []map[string]any{{"id": "a-1", "code": "first_trade", "progress": 1, "goal": 1, "completed": true}, {"id": "a-2", "code": "production_novice", "progress": 12, "goal": 100, "completed": false}},
			Notifications:       []model.Notification{{ID: "n-1", Title: "Market Update", Body: "Bots are providing liquidity.", Read: false}},
			PlayerPreferences:   map[string]any{"soundEnabled": true, "buildingAnimationsEnabled": true, "autoDisableAnimations": false},
			BotCompanies:        []model.Company{botA, botB},
			MarketTicks:         map[int][]map[string]any{},
			ProductionJobs:      []model.ProductionJob{},
			GovernmentContracts: []model.GovContract{{ID: "gov-1", ResourceID: 8, Quality: 0, Quantity: 500, MaxPrice: 12.4, DepositRate: 0.1, Status: "open", Bids: []map[string]any{}, WinnerCompanyID: 0}},
			Executives:          []map[string]any{{"id": "ex-coo-1", "role": "COO", "skill": 8}, {"id": "ex-cto-1", "role": "CTO", "skill": 7}},
			XP:                  0,
			XpToNextLevel:       100,
			MarketPressure:      map[int]float64{},
			UnlockedRecipes:     map[int]bool{},
			ResearchedQuality:   map[int]int{},
			ResearchProjects: []model.ResearchProject{
				{
					ID: "research-project-29", Name: "Plant Research",
					Building:     "Plant Research Center",
					ResourceCost: map[int]int{24: 200, 25: 150},
					CashCost:     50000, DurationHours: 6,
					QualityResourceID: 3, // improves quality of Apples
					Status:            "available", Progress: 0,
				},
				{
					ID: "research-project-30", Name: "Energy Research",
					Building:     "Physics Lab",
					ResourceCost: map[int]int{24: 300, 25: 200},
					CashCost:     80000, DurationHours: 8,
					QualityResourceID: 11, // improves quality of Ice Cream
					Status:            "in_progress", Progress: 45,
					StartedAt: now, CompletesAt: time.Now().UTC().Add(8 * time.Hour).Format(time.RFC3339),
				},
				{
					ID: "research-project-31", Name: "Mining Research",
					Building:     "Physics Lab",
					ResourceCost: map[int]int{24: 400, 25: 300},
					CashCost:     100000, DurationHours: 10,
					QualityResourceID: 10, // improves quality of Minerals
					Status:            "in_progress", Progress: 12,
					StartedAt: now, CompletesAt: time.Now().UTC().Add(10 * time.Hour).Format(time.RFC3339),
				},
				{
					ID: "research-project-32", Name: "Chemical Research",
					Building:     "Physics Lab",
					ResourceCost: map[int]int{24: 500, 25: 400},
					CashCost:     120000, DurationHours: 12,
					QualityResourceID: 7, // improves quality of Chemicals
					Status:            "available", Progress: 0,
				},
			},
			Auctions: []model.Auction{
				{ID: "auction-1", Item: "Industrial Warehouse", ItemID: "b-auction-1", StartingBid: 50000, CurrentBid: 50000, HighestBidder: 0, EndsAt: time.Now().UTC().Add(48 * time.Hour).Format(time.RFC3339), Status: "open", CreatedAt: now},
				{ID: "auction-2", Item: "Farmland Expansion", ItemID: "b-auction-2", StartingBid: 25000, CurrentBid: 25000, HighestBidder: 0, EndsAt: time.Now().UTC().Add(72 * time.Hour).Format(time.RFC3339), Status: "open", CreatedAt: now},
			},
			Ledger:             []model.LedgerEntry{{ID: "l-0", At: now, Kind: "opening_balance", Amount: cfg.Game.StartMoney, Direction: "in"}},
			LastActiveAt:       now,
			ProcessedRequests:  map[string]map[string]any{},
			DailyTradeVolume:   map[int]float64{},
			DailyTradeQty:      map[int]int{},
			DailyHighPrice:     map[int]float64{},
			DailyLowPrice:      map[int]float64{},
			YesterdayVolume:    map[int]float64{},
			YesterdayHighPrice: map[int]float64{},
			YesterdayClose:     map[int]float64{},
			LastTradePrice:     map[int]float64{},
			MarketLocked:       map[int]bool{},
			NationalTeamActive: map[int]bool{},
		},
	}
}

func (s *Service) ensureRuntimeState() {
	if s.State.Companies == nil {
		s.State.Companies = []model.Company{}
	}
	for i := range s.State.Companies {
		c := &s.State.Companies[i]
		if c.Inventory == nil {
			c.Inventory = map[int]int{}
		}
		if c.QualityInventory == nil {
			c.QualityInventory = map[string]int{}
		}
	}
	if s.State.Trades == nil {
		s.State.Trades = []model.Trade{}
	}
	if s.State.ProcessedRequests == nil {
		s.State.ProcessedRequests = map[string]map[string]any{}
	}
	if s.State.MarketPressure == nil {
		s.State.MarketPressure = map[int]float64{}
	}
	if s.State.DailyTradeVolume == nil {
		s.State.DailyTradeVolume = map[int]float64{}
	}
	if s.State.DailyTradeQty == nil {
		s.State.DailyTradeQty = map[int]int{}
	}
	if s.State.DailyHighPrice == nil {
		s.State.DailyHighPrice = map[int]float64{}
	}
	if s.State.DailyLowPrice == nil {
		s.State.DailyLowPrice = map[int]float64{}
	}
	if s.State.YesterdayVolume == nil {
		s.State.YesterdayVolume = map[int]float64{}
	}
	if s.State.YesterdayHighPrice == nil {
		s.State.YesterdayHighPrice = map[int]float64{}
	}
	if s.State.YesterdayClose == nil {
		s.State.YesterdayClose = map[int]float64{}
	}
	if s.State.LastTradePrice == nil {
		s.State.LastTradePrice = map[int]float64{}
	}
	if s.State.MarketLocked == nil {
		s.State.MarketLocked = map[int]bool{}
	}
	if s.State.NationalTeamActive == nil {
		s.State.NationalTeamActive = map[int]bool{}
	}
	s.ensureBotCompaniesLocked()
}

func (s *Service) addXP(company *model.Company, amount int) {
	s.State.XP += amount
	for s.State.XP >= s.State.XpToNextLevel {
		s.State.XP -= s.State.XpToNextLevel
		if company != nil {
			company.Level++
			s.State.XpToNextLevel = company.Level * 100
		}
		s.addLedger("level_up", 0, "in", map[string]any{"newLevel": s.State.XpToNextLevel})
	}
}

// getCompanyLocked returns a pointer to the company with the given ID.
// Caller MUST hold s.mu. Returns nil if not found.
func (s *Service) getCompanyLocked(id int) *model.Company {
	for i := range s.State.Companies {
		if s.State.Companies[i].ID == id {
			return &s.State.Companies[i]
		}
	}
	return nil
}
func intFromAny(v any) int {
	switch t := v.(type) {
	case int:
		return t
	case float64:
		return int(t)
	default:
		return 0
	}
}

// Scheduler-global: settle for all companies.
func (s *Service) SettleAllBonds() map[string]any {
	result := map[string]any{"dailyBondIncome": 0.0, "dailyBondExpense": 0.0, "defaults": []map[string]any{}}
	for i := range s.State.Companies {
		cid := s.State.Companies[i].ID
		r := s.SettleBondInterest(cid)
		result["dailyBondIncome"] = result["dailyBondIncome"].(float64) + r["dailyBondIncome"].(float64)
		result["dailyBondExpense"] = result["dailyBondExpense"].(float64) + r["dailyBondExpense"].(float64)
	}
	return result
}

func (s *Service) RunAllProductionJobs() {
	for i := range s.State.Companies {
		s.refreshProductionJobs(s.State.Companies[i].ID)
	}
}
func floatFromAny(v any) float64 {
	switch t := v.(type) {

	case float64:
		return t
	case int:
		return float64(t)
	default:
		return 0
	}
}

func (s *Service) inventoryKey(resID, quality int) string {
	return fmt.Sprintf("%d_%d", resID, quality)
}
func (s *Service) inventoryGet(company *model.Company, resID, quality int) int {
	if quality <= 0 {
		return company.Inventory[resID]
	}
	if company.QualityInventory == nil {
		return 0
	}
	return company.QualityInventory[s.inventoryKey(resID, quality)]
}
func (s *Service) inventoryAdd(company *model.Company, resID, quality, qty int) {
	if qty <= 0 {
		return
	}
	if quality <= 0 {
		company.Inventory[resID] += qty
		return
	}
	if company.QualityInventory == nil {
		company.QualityInventory = map[string]int{}
	}
	company.QualityInventory[s.inventoryKey(resID, quality)] += qty
}
func (s *Service) inventorySub(company *model.Company, resID, quality, qty int) bool {
	if qty <= 0 {
		return false
	}
	cur := s.inventoryGet(company, resID, quality)
	if cur < qty {
		return false
	}
	if quality <= 0 {
		company.Inventory[resID] -= qty
	} else {
		company.QualityInventory[s.inventoryKey(resID, quality)] -= qty
	}
	return true
}
func (s *Service) ResourcesWithMarket() []int {
	return []int{8, 9, 10, 11, 12}
}
