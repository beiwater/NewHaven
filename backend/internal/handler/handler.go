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

type Handler struct{ svc *service.Service }

func New(svc *service.Service) *Handler { return &Handler{svc: svc} }
func (h *Handler) companyID(r *http.Request) int {
	if cid, ok := r.Context().Value(middleware.CompanyIDKey).(int); ok {
		return cid
	}
	return 0
}

func (h *Handler) withAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
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
		_, companyID, ok := h.svc.ValidateToken(token)
		if !ok {
			writeErr(w, 401, "invalid token")
			return
		}
		ctx := context.WithValue(r.Context(), middleware.CompanyIDKey, companyID)
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
