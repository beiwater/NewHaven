package httpapi

import (
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/newhaven/backend-next/internal/config"
)

func NewRouter(cfg *config.Config, authHandler *AuthHandler, companyHandler *CompanyHandler, warehouseHandler *WarehouseHandler, buildingHandler *BuildingHandler, productionHandler *ProductionHandler, marketHandler *MarketHandler, financeHandler *FinanceHandler, bondHandler *BondHandler) *chi.Mux {
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
		r.Post("/register", authHandler.handleRegister)
		r.Post("/login", authHandler.handleLogin)
	})
	// Bond routes (authenticated)
	r.With(AuthRequired(cfg.JWTSigningKey)).Get("/api/bonds/", bondHandler.handleListBonds)
	r.With(AuthRequired(cfg.JWTSigningKey)).Post("/api/bonds/", bondHandler.handleCreateBond)
	r.With(AuthRequired(cfg.JWTSigningKey)).Get("/api/bonds/{bondId}/", bondHandler.handleGetBond)

	// API v2 domain routes (authenticated)
	r.Route("/api/v2", func(r chi.Router) {
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/companies/me/warehouse/", warehouseHandler.handleGetMyWarehouse)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/companies/me/warehouse/upgrade/", warehouseHandler.handleUpgradeWarehouse)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/companies/me/cashflow/recent/", financeHandler.handleRecentCashflow)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/companies/me/income-statement/", financeHandler.handleIncomeStatement)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/companies/me/balance-sheet/", financeHandler.handleBalanceSheet)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/companies/me/cashflow-statement/", financeHandler.handleCashflowStatement)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/market-order/", marketHandler.handleCreateOrder)
		r.With(AuthRequired(cfg.JWTSigningKey)).Delete("/market-order/cancel/{orderId}/", marketHandler.handleCancelOrder)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/market-order/take/", marketHandler.handleTakeOrder)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/players/me/companies/", companyHandler.handleListMyCompanies)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/production/jobs/", productionHandler.handleListProductionJobs)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/buildings/market/", buildingHandler.handleBuildingMarket)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/buildings/buy/", buildingHandler.handleBuyBuilding)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/buildings/place/", buildingHandler.handlePlaceBuilding)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/buildings/move/", buildingHandler.handleMoveBuilding)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/buildings/demolish/", buildingHandler.handleDemolishBuilding)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/companies/me/buildings/", buildingHandler.handleListMyBuildingsV2)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/production/start/", productionHandler.handleStartProduction)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/production/claim/{jobId}/", productionHandler.handleClaimProduction)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/production/claimable/", productionHandler.handleListClaimableJobs)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/production/queue/", productionHandler.handleProductionQueue)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/production/cancel/", productionHandler.handleCancelProduction)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/production/claim-all/", productionHandler.handleClaimAll)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/buildings/{buildingId}/production-options/", productionHandler.handleProductionOptions)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/companies/me/bonds/owned/", bondHandler.handleGetOwnedBonds)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/companies/me/bonds/sold/", bondHandler.handleGetSoldBonds)
	})

	// API v3 domain routes (authenticated)
	r.Route("/api/v3", func(r chi.Router) {
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/companies/me/buildings/", buildingHandler.handleListMyBuildings)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/resources/", marketHandler.handleResources)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/companies/me/past-finances/", financeHandler.handlePastFinances)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/market-ticker/{resourceId}/", marketHandler.handleMarketTicker)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/market-depth/{resourceId}/{quality}/", marketHandler.handleMarketDepth)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/market/{resourceId}/{quality}/", marketHandler.handleMarketOrders)
	})

	// API v1 routes (authenticated)
	r.Route("/api/v1", func(r chi.Router) {
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/buildings/{buildingId}/upgrade/", buildingHandler.handleUpgradeBuilding)
	})

	return r
}
