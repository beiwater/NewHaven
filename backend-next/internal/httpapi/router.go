package httpapi

import (
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/newhaven/backend-next/internal/config"
)

func NewRouter(cfg *config.Config, authHandler *AuthHandler, companyHandler *CompanyHandler, warehouseHandler *WarehouseHandler, buildingHandler *BuildingHandler, productionHandler *ProductionHandler, marketHandler *MarketHandler) *chi.Mux {
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

	// API v2 domain routes (authenticated)
	r.Route("/api/v2", func(r chi.Router) {
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/companies/me/warehouse/", warehouseHandler.handleGetMyWarehouse)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/players/me/companies/", companyHandler.handleListMyCompanies)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/production/jobs/", productionHandler.handleListProductionJobs)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/production/start/", productionHandler.handleStartProduction)
		r.With(AuthRequired(cfg.JWTSigningKey)).Post("/production/claim/{jobId}/", productionHandler.handleClaimProduction)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/production/claimable/", productionHandler.handleListClaimableJobs)
	})

	// API v3 domain routes (authenticated)
	r.Route("/api/v3", func(r chi.Router) {
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/companies/me/buildings/", buildingHandler.handleListMyBuildings)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/resources/", marketHandler.handleResources)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/market-ticker/{resourceId}/", marketHandler.handleMarketTicker)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/market-depth/{resourceId}/{quality}/", marketHandler.handleMarketDepth)
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/market/{resourceId}/{quality}/", marketHandler.handleMarketOrders)
	})

	return r
}
