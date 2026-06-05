package httpapi

import (
	"errors"
	"net/http"

	"github.com/newhaven/backend-next/internal/app/company"
)

// CompanyHandler handles company-related HTTP endpoints.
type CompanyHandler struct {
	svc *company.Service
}

// NewCompanyHandler creates a new CompanyHandler.
func NewCompanyHandler(svc *company.Service) *CompanyHandler {
	return &CompanyHandler{svc: svc}
}

// handleListMyCompanies returns the companies for the authenticated player.
func (h *CompanyHandler) handleListMyCompanies(w http.ResponseWriter, r *http.Request) {
	playerID, ok := PlayerIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "player not authenticated", nil)
		return
	}

	resp, err := h.svc.ListMyCompanies(r.Context(), playerID)
	if err != nil {
		if errors.Is(err, company.ErrNotFound) {
			writeErr(w, 404, ErrorNotFound, "no company found for player", nil)
			return
		}
		writeErr(w, 500, ErrorInternal, "failed to list companies", nil)
		return
	}

	writeSuccess(w, 200, resp)
}
