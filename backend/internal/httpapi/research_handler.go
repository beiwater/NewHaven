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

// handleListResearch returns all products with their current quality ceiling.
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

// handleUnlockQuality pays money and unlocks one explicit target quality.
// The target makes a client retry safe: replaying Qn never advances to Q(n+1).
func (h *ResearchHandler) handleUnlockQuality(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	var req struct {
		ResourceID    int `json:"resourceId"`
		TargetQuality int `json:"targetQuality"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ResourceID <= 0 || req.TargetQuality <= 0 {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "resourceId and targetQuality are required", nil)
		return
	}

	resp, err := h.svc.UnlockQuality(r.Context(), companyID, req.ResourceID, req.TargetQuality)
	if err != nil {
		switch {
		case err == research.ErrMaxLevel:
			writeErr(w, http.StatusBadRequest, ErrorConflict, "target quality must be between Q1 and Q12", nil)
		case err == research.ErrInsufficientFunds:
			writeErr(w, http.StatusBadRequest, ErrorInsufficientFunds, "insufficient funds for quality research", nil)
		case err == research.ErrResearchSequence:
			writeErr(w, http.StatusConflict, ErrorConflict, "quality research must unlock the next Q level", nil)
		case err == research.ErrResourceNotFound:
			writeErr(w, http.StatusNotFound, ErrorNotFound, "researchable resource not found", nil)
		default:
			writeAppErr(w, apperr.WrapMsg(apperr.KindInternal, "quality research failed", err))
		}
		return
	}

	writeSuccess(w, http.StatusOK, map[string]any{
		"research": resp,
	})
}
