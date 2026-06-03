package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"go-sim-api/internal/middleware"
	"go-sim-api/internal/service"
)

type Handler struct {
	svc       *service.Service
	jwtSecret string
}

func New(svc *service.Service, jwtSecret string) *Handler {
	return &Handler{svc: svc, jwtSecret: jwtSecret}
}

func (h *Handler) companyID(r *http.Request) int {
	if cid, ok := r.Context().Value(middleware.CompanyIDKey).(int); ok {
		return cid
	}
	return 0
}

func (h *Handler) playerID(r *http.Request) int {
	if pid, ok := r.Context().Value(middleware.PlayerIDKey).(int); ok {
		return pid
	}
	return 0
}

func (h *Handler) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		// Public endpoints: no auth required
		if path == "/healthz" || path == "/readyz" || path == "/api/register" || path == "/api/login" || path == "/api/csrf/" {
			next(w, r)
			return
		}
		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			writeErr(w, 401, "unauthorized")
			return
		}
		token := strings.TrimPrefix(auth, "Bearer ")
		playerID, companyID, err := middleware.ParseJWT(token, h.jwtSecret)
		if err != nil {
			writeErr(w, 401, "invalid token: "+err.Error())
			return
		}
		ctx := context.WithValue(r.Context(), middleware.PlayerIDKey, playerID)
		ctx = context.WithValue(ctx, middleware.CompanyIDKey, companyID)
		next(w, r.WithContext(ctx))
	}
}

func (h *Handler) Register(mux *http.ServeMux) {
	h.RegisterAuth(mux)

	h.RegisterCompany(mux)
	h.RegisterFinancial(mux)
	h.RegisterPlayer(mux)
	h.RegisterMarket(mux)
	h.RegisterProduction(mux)
	h.RegisterProductionQueue(mux)
	h.RegisterBond(mux)
	h.RegisterGovernment(mux)
	h.RegisterMessage(mux)
	h.RegisterRecipe(mux)
	h.RegisterDev(mux)
	h.RegisterAerospace(mux)
	h.RegisterLeaderboard(mux)
	h.RegisterHealth(mux)
	h.RegisterBuildingShop(mux)
	h.RegisterOrder(mux)
}

func writeJSON(w http.ResponseWriter, s int, p any) {
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(s)
	_ = json.NewEncoder(w).Encode(p)
}

func writeErr(w http.ResponseWriter, s int, m string) {
	writeJSON(w, s, map[string]any{"error": m})
}

func tailInt(p string) (int, error) {
	parts := strings.Split(strings.Trim(p, "/"), "/")
	if len(parts) == 0 {
		return 0, strconv.ErrSyntax
	}
	return strconv.Atoi(parts[len(parts)-1])
}
