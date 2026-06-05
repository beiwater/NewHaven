package httpapi

import (
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"github.com/newhaven/backend-next/internal/config"
)

func NewRouter(cfg *config.Config) *chi.Mux {
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

	// API v2 (retained compatibility)
	r.Route("/api/v2", func(r chi.Router) {
		_ = cfg // will use when wiring domain routes
	})

	// API v3 (retained)
	r.Route("/api/v3", func(r chi.Router) {
	})

	// Future: mounts for each migrated domain
	// r.Mount("/api/v2/companies", companyHandler)
	// r.Mount("/api/v3/market", marketHandler)

	return r
}
