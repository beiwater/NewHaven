package app

import (
	"context"
	"log/slog"

	"github.com/newhaven/backend-next/internal/app/auth"
	"github.com/newhaven/backend-next/internal/app/building"
	"github.com/newhaven/backend-next/internal/app/company"
	"github.com/newhaven/backend-next/internal/app/finance"
	"github.com/newhaven/backend-next/internal/app/market"
	"github.com/newhaven/backend-next/internal/app/production"
	"github.com/newhaven/backend-next/internal/app/research"
	"github.com/newhaven/backend-next/internal/app/warehouse"
	"github.com/newhaven/backend-next/internal/catalog"
	"github.com/newhaven/backend-next/internal/config"
	"github.com/newhaven/backend-next/internal/httpapi"
	"github.com/newhaven/backend-next/internal/platform"
	"github.com/newhaven/backend-next/internal/storage"
	"github.com/newhaven/backend-next/internal/storage/memory"
	"github.com/newhaven/backend-next/internal/storage/postgres"
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
	FinanceService    *finance.Service

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
	ContractHandler    *httpapi.ContractHandler
	ResearchHandler    *httpapi.ResearchHandler
	ExecutiveHandler   *httpapi.ExecutiveHandler
	LeaderboardHandler *httpapi.LeaderboardHandler
	AdminHandler      *httpapi.AdminHandler
}

// New creates a fully wired App with all services and handlers constructed.
func New(cfg *config.Config, st storage.Storage, resources map[int]*catalog.ResourceEntry, buildings map[int]*catalog.BuildingEntry, economy map[int]*catalog.EconomyModelEntry) *App {
	clock := platform.RealClock{}
	idgen := platform.NewIDGen()
	logger := platform.NewLogger(slog.Default())

	// Auth
	authSvc := auth.NewService(st, st, clock, idgen, logger, cfg.JWTSigningKey)

	// Market
	marketSvc := market.NewService(st, st, st, resources, economy, cfg.Game, clock, idgen)
	marketHandler := httpapi.NewMarketHandler(marketSvc)

	// Company
	companySvc := company.NewService(st, logger)
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
	researchSvc := research.NewService(st, st, resources, cfg.Game, logger)
	researchHandler := httpapi.NewResearchHandler(researchSvc)

	// Finance
	financeSvc := finance.NewService(st, st, clock, idgen, cfg.Game)
	financeHandler := httpapi.NewFinanceHandler(financeSvc)
	bondHandler := httpapi.NewBondHandler(financeSvc)
	playerHandler := httpapi.NewPlayerHandler(st)
	contractHandler := httpapi.NewContractHandler()
	executiveHandler := httpapi.NewExecutiveHandler(st)
	leaderboardHandler := httpapi.NewLeaderboardHandler(st)
	adminHandler := httpapi.NewAdminHandler(st)
	socialHandler := httpapi.NewSocialHandler(st, st, cfg.Game.MaxMessageLength)
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
		Config:        cfg,
		Storage:       st,
		Clock:         clock,
		IDGen:         idgen,
		Logger:        logger,
		Resources:     resources,
		Buildings:     buildings,
		Economy:       economy,

		AuthService:       authSvc,
		CompanyService:    companySvc,
		WarehouseService:  warehouseSvc,
		BuildingService:   buildingSvc,
		ProductionService: productionSvc,
		MarketService:     marketSvc,
		FinanceService:    financeSvc,
		ResearchService:   researchSvc,

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
		ContractHandler:    contractHandler,
		ResearchHandler:    researchHandler,
		ExecutiveHandler:   executiveHandler,
		LeaderboardHandler: leaderboardHandler,
		AdminHandler:      adminHandler,
	}
}
