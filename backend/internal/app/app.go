package app

import (
	"context"
	"log/slog"
	"path/filepath"

	"github.com/beiwater/NewHaven/backend/internal/app/auth"
	"github.com/beiwater/NewHaven/backend/internal/app/building"
	"github.com/beiwater/NewHaven/backend/internal/app/company"
	"github.com/beiwater/NewHaven/backend/internal/app/executive"
	"github.com/beiwater/NewHaven/backend/internal/app/finance"
	"github.com/beiwater/NewHaven/backend/internal/app/market"
	"github.com/beiwater/NewHaven/backend/internal/app/production"
	"github.com/beiwater/NewHaven/backend/internal/app/research"
	"github.com/beiwater/NewHaven/backend/internal/app/terminal"
	"github.com/beiwater/NewHaven/backend/internal/app/warehouse"
	"github.com/beiwater/NewHaven/backend/internal/catalog"
	"github.com/beiwater/NewHaven/backend/internal/config"
	"github.com/beiwater/NewHaven/backend/internal/httpapi"
	"github.com/beiwater/NewHaven/backend/internal/platform"
	"github.com/beiwater/NewHaven/backend/internal/storage"
	"github.com/beiwater/NewHaven/backend/internal/storage/memory"
	"github.com/beiwater/NewHaven/backend/internal/storage/postgres"
)

// App holds all shared infrastructure, domain services, and HTTP handlers.
type App struct {
	// Shared infrastructure
	Config    *config.Config
	Storage   storage.Storage
	Clock     platform.Clock
	IDGen     *platform.IDGen
	Logger    *platform.Logger
	Resources map[int]*catalog.ResourceEntry
	Buildings map[int]*catalog.BuildingEntry
	Economy   map[int]*catalog.EconomyModelEntry

	// Domain services
	AuthService       *auth.Service
	CompanyService    *company.Service
	WarehouseService  *warehouse.Service
	BuildingService   *building.Service
	ProductionService *production.Service
	MarketService     *market.Service
	ResearchService   *research.Service
	ExecutiveService  *executive.Service
	FinanceService    *finance.Service
	TerminalService   *terminal.Service

	// Scheduler dependencies
	PostgresStore *postgres.Store
	SaveAll       func(ctx context.Context) error

	// HTTP handlers
	CompanyHandler     *httpapi.CompanyHandler
	AuthHandler        *httpapi.AuthHandler
	WarehouseHandler   *httpapi.WarehouseHandler
	BuildingHandler    *httpapi.BuildingHandler
	ProductionHandler  *httpapi.ProductionHandler
	MarketHandler      *httpapi.MarketHandler
	FinanceHandler     *httpapi.FinanceHandler
	BondHandler        *httpapi.BondHandler
	PlayerHandler      *httpapi.PlayerHandler
	SocialHandler      *httpapi.SocialHandler
	ChatHandler        *httpapi.ChatHandler
	ContractHandler    *httpapi.ContractHandler
	ResearchHandler    *httpapi.ResearchHandler
	ExecutiveHandler   *httpapi.ExecutiveHandler
	LeaderboardHandler *httpapi.LeaderboardHandler
	AdminHandler       *httpapi.AdminHandler
	ReportHandler      *httpapi.ReportHandler
}

// New creates a fully wired App with all services and handlers constructed.
func New(cfg *config.Config, st storage.Storage, resources map[int]*catalog.ResourceEntry, buildings map[int]*catalog.BuildingEntry, economy map[int]*catalog.EconomyModelEntry) *App {
	clock := platform.RealClock{}
	idgen := platform.NewIDGen()
	logger := platform.NewLogger(slog.Default())

	// Ensure game config is non-nil (tests may pass nil)
	if cfg.Game == nil {
		cfg.Game = &config.GameConfig{
			MaxMessageLength:     500,
			BondFaceValue:        5000,
			ExchangeFeePct:       0.04,
			BotReplacementRate:   0.33,
			BondMinInterest:      0.5,
			BondMaxInterest:      2.0,
			ProductionMod:        1.0,
			AdminOverheadBase:    1.35,
			BaseBuildingCost:     50000,
			BaseProductionSlots:  3,
			WarehouseBaseCap:     1000,
			WarehouseUpgradeCost: 25000,
			BaseOutput:           100,
			MaxBuildings:         20,
			ResearchBaseCost:     1000,
			ResearchCostGrowth:   1.2,
			ResearchSpeedBonus:   0.002,
		}
	}

	// Auth
	authSvc := auth.NewService(st, st, st, clock, idgen, logger, cfg.JWTSigningKey, cfg.DevPassword)

	// Market
	marketSvc := market.NewService(st, st, st, resources, buildings, economy, cfg.Game, clock, idgen)
	marketHandler := httpapi.NewMarketHandler(marketSvc)

	// Company
	companySvc := company.NewService(st, logger, cfg.Game.NewbieLevelUpTo)
	companyHandler := httpapi.NewCompanyHandler(companySvc, marketSvc)
	authHandler := httpapi.NewAuthHandler(authSvc)

	// Warehouse
	warehouseSvc := warehouse.NewService(st, st, cfg.Game, logger)
	warehouseHandler := httpapi.NewWarehouseHandler(warehouseSvc)

	// Building
	buildingSvc := building.NewService(st, buildings, cfg.Game, clock, idgen)
	buildingHandler := httpapi.NewBuildingHandler(buildingSvc)
	// Production
	productionSvc := production.NewService(st, st, st, st, cfg.Game, resources, buildings, clock, idgen)
	productionHandler := httpapi.NewProductionHandler(productionSvc)

	// Research
	researchSvc := research.NewService(st, st, st, resources, cfg.Game, logger)
	researchHandler := httpapi.NewResearchHandler(researchSvc)

	// Finance
	financeSvc := finance.NewService(st, st, clock, idgen, cfg.Game)
	financeHandler := httpapi.NewFinanceHandler(financeSvc)
	bondHandler := httpapi.NewBondHandler(financeSvc)
	playerHandler := httpapi.NewPlayerHandler(st)
	contractHandler := httpapi.NewContractHandler()
	executiveSvc := executive.NewService(st, st, clock)
	executiveHandler := httpapi.NewExecutiveHandler(executiveSvc)
	leaderboardHandler := httpapi.NewLeaderboardHandler(st)
	adminHandler := httpapi.NewAdminHandler(st)
	terminalSvc := terminal.NewService(st, logger)
	socialHandler := httpapi.NewSocialHandler(st, st, cfg.Game.MaxMessageLength)
	chatHandler := httpapi.NewChatHandler(st, st, cfg.Game.MaxMessageLength, terminalSvc)
	logDir := filepath.Join(config.FindProjectRoot(), "log")
	reportHandler := httpapi.NewReportHandler(logDir, idgen, clock)
	// Snapshot persistence: file-based (memory) or PostgreSQL
	var pgStore *postgres.Store
	var saveAll func(ctx context.Context) error
	if cfg.DatabaseURL != "" {
		memStore, ok := st.(*memory.Store)
		if ok {
			var err error
			pgStore, err = postgres.New(context.Background(), cfg.DatabaseURL, memStore)
			if err != nil {
				slog.Warn("[app] postgres not available, skipping persistence", "error", err)
				pgStore = nil
			} else {
				if err := pgStore.LoadSnapshot(context.Background()); err != nil {
					slog.Info("[app] no existing snapshot in database", "error", err)
				}
				saveAll = pgStore.SaveSnapshot
			}
		} else {
			slog.Warn("[app] storage is not memory-backed, skipping postgres persistence")
		}
	}
	if saveAll == nil {
		// Memory-only mode: auto-save to data/snapshot.json via scheduler
		saveAll = st.SaveSnapshot
	}

	return &App{
		Config:    cfg,
		Storage:   st,
		Clock:     clock,
		IDGen:     idgen,
		Logger:    logger,
		Resources: resources,
		Buildings: buildings,
		Economy:   economy,

		AuthService:       authSvc,
		CompanyService:    companySvc,
		WarehouseService:  warehouseSvc,
		BuildingService:   buildingSvc,
		ProductionService: productionSvc,
		MarketService:     marketSvc,
		FinanceService:    financeSvc,
		ResearchService:   researchSvc,
		ExecutiveService:  executiveSvc,
		TerminalService:   terminalSvc,

		PostgresStore: pgStore,
		SaveAll:       saveAll,

		CompanyHandler:     companyHandler,
		AuthHandler:        authHandler,
		WarehouseHandler:   warehouseHandler,
		BuildingHandler:    buildingHandler,
		ProductionHandler:  productionHandler,
		MarketHandler:      marketHandler,
		FinanceHandler:     financeHandler,
		BondHandler:        bondHandler,
		PlayerHandler:      playerHandler,
		SocialHandler:      socialHandler,
		ChatHandler:        chatHandler,
		ContractHandler:    contractHandler,
		ResearchHandler:    researchHandler,
		ExecutiveHandler:   executiveHandler,
		LeaderboardHandler: leaderboardHandler,
		AdminHandler:       adminHandler,
		ReportHandler:      reportHandler,
	}
}
