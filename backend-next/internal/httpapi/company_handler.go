package httpapi

import (
	"log/slog"
	"net/http"

	"github.com/newhaven/backend-next/internal/app/company"
	"github.com/newhaven/backend-next/internal/app/market"
)
// CompanyHandler handles company-related HTTP endpoints.
type CompanyHandler struct {
	svc       *company.Service
	marketSvc *market.Service
}

// NewCompanyHandler creates a new CompanyHandler.
func NewCompanyHandler(svc *company.Service, marketSvc *market.Service) *CompanyHandler {
	return &CompanyHandler{svc: svc, marketSvc: marketSvc}
}

// handleListMyCompanies returns the companies for the authenticated player.
func (h *CompanyHandler) handleListMyCompanies(w http.ResponseWriter, r *http.Request) {
	playerID, ok := PlayerIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "player not authenticated", nil)
		return
	}

	// Catch up retail sales before returning company data.
	companyID, ok := CompanyIDFromCtx(r.Context())
	if ok {
		if err := h.marketSvc.CatchUpPlayerRetail(r.Context(), companyID); err != nil {
			slog.Warn("retail catch-up failed", "company_id", companyID, "error", err)
		}
	}

	resp, err := h.svc.ListMyCompanies(r.Context(), playerID)
	if err != nil {
		writeAppErr(w, err)
		return
	}

	writeSuccess(w, 200, resp)
}

// handleCompleteTutorial marks the company's tutorial as complete and catches up retail.
func (h *CompanyHandler) handleCompleteTutorial(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	// Catch up retail before completing tutorial.
	if err := h.marketSvc.CatchUpPlayerRetail(r.Context(), companyID); err != nil {
		slog.Warn("retail catch-up failed", "company_id", companyID, "error", err)
	}

	writeSuccess(w, 200, map[string]any{"ok": true})
}
