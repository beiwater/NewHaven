package httpapi

import (
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/newhaven/backend-next/internal/config"
)

func NewRouter(cfg *config.Config, authHandler *AuthHandler, companyHandler *CompanyHandler, warehouseHandler *WarehouseHandler, buildingHandler *BuildingHandler) *chi.Mux {
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
	})

	// API v3 domain routes (authenticated)
	r.Route("/api/v3", func(r chi.Router) {
		r.With(AuthRequired(cfg.JWTSigningKey)).Get("/companies/me/buildings/", buildingHandler.handleListMyBuildings)
	})

	return r
}
