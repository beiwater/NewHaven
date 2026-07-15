package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/beiwater/NewHaven/backend/internal/app/research"
	"github.com/beiwater/NewHaven/backend/internal/apperr"
)

// ResearchHandler handles per-resource research endpoints.
type ResearchHandler struct {
	svc *research.Service
}

// NewResearchHandler creates a new ResearchHandler.
func NewResearchHandler(svc *research.Service) *ResearchHandler {
	return &ResearchHandler{svc: svc}
}

// handleListResearch returns all researchable resources with their current levels.
func (h *ResearchHandler) handleListResearch(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	items, err := h.svc.ListResearch(r.Context(), companyID)
	if err != nil {
		writeAppErr(w, apperr.WrapMsg(apperr.KindInternal, "failed to load research", err))
		return
	}

	writeSuccess(w, http.StatusOK, map[string]any{
		"research": items,
	})
}

// handleLevelUp pays money and increases a resource's research level by 1.
func (h *ResearchHandler) handleLevelUp(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	var req struct {
		ResourceID int `json:"resourceId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ResourceID <= 0 {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "resourceId is required", nil)
		return
	}

	resp, err := h.svc.LevelUp(r.Context(), companyID, req.ResourceID)
	if err != nil {
		switch {
		case err == research.ErrMaxLevel:
			writeErr(w, http.StatusBadRequest, ErrorConflict, "research already at max level", nil)
		case err == research.ErrInsufficientFunds:
			writeErr(w, http.StatusBadRequest, ErrorInsufficientFunds, "insufficient funds for research", nil)
		default:
			writeAppErr(w, apperr.WrapMsg(apperr.KindInternal, "research level-up failed", err))
		}
		return
	}

	writeSuccess(w, http.StatusOK, map[string]any{
		"research": resp,
	})
}
