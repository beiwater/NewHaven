package httpapi

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/newhaven/backend-next/internal/app/company"
)

func (h *CompanyHandler) handleGetCompanyProfile(w http.ResponseWriter, r *http.Request) {
	playerID, playerOK := PlayerIDFromCtx(r.Context())
	companyID, companyOK := CompanyIDFromCtx(r.Context())
	if !playerOK || !companyOK {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}
	requestedCompanyID, err := strconv.Atoi(chi.URLParam(r, "companyId"))
	if err != nil {
		writeErr(w, 400, ErrorValidation, "invalid company id", nil)
		return
	}
	// Catch up retail sales since last login before returning profile.
	if err := h.marketSvc.CatchUpPlayerRetail(r.Context(), companyID); err != nil {
		slog.Warn("retail catch-up failed", "company_id", companyID, "error", err)
	}
	resp, err := h.svc.GetProfile(r.Context(), playerID, companyID, requestedCompanyID)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeSuccess(w, 200, resp)
}

func (h *CompanyHandler) handleUpdateStoryProgress(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	// Catch up retail before updating story progress.
	if err := h.marketSvc.CatchUpPlayerRetail(r.Context(), companyID); err != nil {
		slog.Warn("retail catch-up failed", "company_id", companyID, "error", err)
	}

	var req company.StoryProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, ErrorBadRequest, "invalid request body", nil)
		return
	}
	resp, err := h.svc.UpdateStoryProgress(r.Context(), companyID, req)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeSuccess(w, 200, resp)
}
