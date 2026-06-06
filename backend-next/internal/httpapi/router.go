package httpapi

import (
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/newhaven/backend-next/internal/config"
)

type RouterHandlers struct {
	Auth        *AuthHandler
	Company     *CompanyHandler
	Warehouse   *WarehouseHandler
	Building    *BuildingHandler
	Production  *ProductionHandler
	Market      *MarketHandler
	Finance     *FinanceHandler
	Bond        *BondHandler
	Player      *PlayerHandler
	Social      *SocialHandler
	Contract    *ContractHandler
	Research    *ResearchHandler
	Executive   *ExecutiveHandler
	Leaderboard *LeaderboardHandler
}

func NewRouter(cfg *config.Config, h *RouterHandlers) *chi.Mux {
	r := chi.NewRouter()

	// Middleware stack (outermost first)
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(Logger)
	r.Use(chimw.Recoverer)
	r.Use(CORS)

	// Health
	r.Get("/healthz", handleHealthz)
	r.Get("/readyz", handleReadyz)

	// Auth routes (unauthenticated)
	r.Route("/api", func(r chi.Router) {
		r.Post("/register", h.Auth.handleRegister)
		r.Post("/login", h.Auth.handleLogin)
	})
	// Bond routes (authenticated)
	r.With(AuthRequired(cfg.JWTSigningKey)).Get("/api/bonds/", h.Bond.handleListBonds)
	r.With(AuthRequired(cfg.JWTSigningKey)).Post("/api/bonds/", h.Bond.handleCreateBond)
	r.With(AuthRequired(cfg.JWTSigningKey)).Get("/api/bonds/{bondId}/", h.Bond.handleGetBond)
	r.With(AuthRequired(cfg.JWTSigningKey)).Post("/api/bonds/{bondId}/call/", h.Bond.handleCallBond)
	r.With(AuthRequired(cfg.JWTSigningKey)).Post("/api/bonds/settle-interest/", h.Bond.handleSettleBondInterest)

	// Player routes (authenticated)
	r.With(AuthRequired(cfg.JWTSigningKey)).Get("/api/v2/players/me/level/", h.Player.handleLevel)
	r.With(AuthRequired(cfg.JWTSigningKey)).Get("/api/v2/players/simboosts/", h.Player.handleSimboostTypes)
	r.With(AuthRequired(cfg.JWTSigningKey)).Get("/api/v2/players/simboosts-use/", h.Player.handleSimboostsUse)
	r.With(AuthRequired(cfg.JWTSigningKey)).Post("/api/v2/players/simboosts-use/", h.Player.handleSimboostsUse)

	// Social routes (authenticated)
	r.With(AuthRequired(cfg.JWTSigningKey)).Get("/api/messages/", h.Social.handleMessages)
	r.With(AuthRequired(cfg.JWTSigningKey)).Post("/api/v2/message/", h.Social.handleV2Message)
	r.With(AuthRequired(cfg.JWTSigningKey)).Get("/api/v2/message/{messageId}/read/", h.Social.handleMarkRead)
	r.With(AuthRequired(cfg.JWTSigningKey)).Get("/api/v2/chatroom/", h.Social.handleChatroom)
	r.With(AuthRequired(cfg.JWTSigningKey)).Get("/api/v2/contacts/", h.Social.handleContacts)
	// API v2 domain routes (authenticated)
	r.Route("/api/v2", func(r chi.Router) {
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/companies/me/warehouse/", h.Warehouse.handleGetMyWarehouse)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/companies/me/warehouse/upgrade/", h.Warehouse.handleUpgradeWarehouse)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/companies/me/cashflow/recent/", h.Finance.handleRecentCashflow)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/companies/me/income-statement/", h.Finance.handleIncomeStatement)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/companies/me/balance-sheet/", h.Finance.handleBalanceSheet)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/companies/me/cashflow-statement/", h.Finance.handleCashflowStatement)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/market-order/", h.Market.handleCreateOrder)
		r.With(AuthRequired(cfg.JWTSigningKey)).Delete("/market-order/cancel/{orderId}/", h.Market.handleCancelOrder)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/market-order/take/", h.Market.handleTakeOrder)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/players/me/companies/", h.Company.handleListMyCompanies)
		r.With(AuthRequired(cfg.JWTSigningKey)).Patch("/companies/me/story-progress/", h.Company.handleUpdateStoryProgress)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/production/jobs/", h.Production.handleListProductionJobs)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/buildings/market/", h.Building.handleBuildingMarket)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/buildings/buy/", h.Building.handleBuyBuilding)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/buildings/place/", h.Building.handlePlaceBuilding)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/buildings/move/", h.Building.handleMoveBuilding)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/buildings/demolish/", h.Building.handleDemolishBuilding)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/buildings/stash/", h.Building.handleStashBuilding)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/companies/me/buildings/", h.Building.handleListMyBuildingsV2)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/production/start/", h.Production.handleStartProduction)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/production/claim/{jobId}/", h.Production.handleClaimProduction)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/production/claimable/", h.Production.handleListClaimableJobs)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/production/queue/", h.Production.handleProductionQueue)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/production/cancel/", h.Production.handleCancelProduction)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/production/claim-all/", h.Production.handleClaimAll)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/buildings/{buildingId}/production-options/", h.Production.handleProductionOptions)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/companies/me/bonds/owned/", h.Bond.handleGetOwnedBonds)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/companies/me/bonds/sold/", h.Bond.handleGetSoldBonds)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/bonds/{bondId}/buy/", h.Bond.handleBuyBond)
	})

	// API v3 domain routes (authenticated)
	r.Route("/api/v3", func(r chi.Router) {
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/companies/{companyId}/", h.Company.handleGetCompanyProfile)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/companies/me/buildings/", h.Building.handleListMyBuildings)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/resources/", h.Market.handleResources)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/companies/me/past-finances/", h.Finance.handlePastFinances)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/market-ticker/{resourceId}/", h.Market.handleMarketTicker)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/market-depth/{resourceId}/{quality}/", h.Market.handleMarketDepth)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/market/{resourceId}/{quality}/", h.Market.handleMarketOrders)
	})

	// API v1 routes (authenticated)
	r.Route("/api/v1", func(r chi.Router) {
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/buildings/{buildingId}/upgrade/", h.Building.handleUpgradeBuilding)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/buildings/{buildingId}/busy/", h.Production.handleStartProductionV1)
	})

	// Contract / daily orders routes (authenticated)
	r.Route("/api/v2/orders", func(r chi.Router) {
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/daily/", h.Contract.handleListDailyOrders)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/daily/complete/", h.Contract.handleCompleteDailyOrder)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/daily/claim/", h.Contract.handleClaimDailyOrder)
	})
	// Government contract routes (authenticated)
	r.Route("/api/v3/government-orders", func(r chi.Router) {
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/", h.Contract.handleListGovContracts)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/bid/", h.Contract.handleBidContract)
	})

	// Research routes (authenticated)
	r.Route("/api/v2/research", func(r chi.Router) {
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/", h.Research.handleListResearch)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/progress/", h.Research.handleResearchProgress)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/start/", h.Research.handleStartResearch)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/complete/{projectId}/", h.Research.handleCompleteResearch)
	})

	// Executive routes (authenticated)
	r.Route("/api/v2/executives", func(r chi.Router) {
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/search/", h.Executive.handleSearchExecutives)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/recruit/", h.Executive.handleRecruitExecutive)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/train/{executiveId}/", h.Executive.handleTrainExecutive)
	})
	r.With(AuthRequired(cfg.JWTSigningKey)).Get("/api/v3/executives/{id}/", h.Executive.handleGetExecutiveDetail)

	// Leaderboard routes (authenticated)
	r.With(AuthRequired(cfg.JWTSigningKey)).Get("/api/v2/leaderboard/", h.Leaderboard.handleLeaderboard)

	return r
}
