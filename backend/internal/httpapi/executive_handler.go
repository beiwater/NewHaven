package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	executiveapp "github.com/beiwater/NewHaven/backend/internal/app/executive"
	"github.com/beiwater/NewHaven/backend/internal/domain/company"
)

// ExecutiveHandler is deliberately a thin HTTP boundary. Candidate generation,
// money movement and cross-tab safety live in the executive application service.
type ExecutiveHandler struct {
	service *executiveapp.Service
}

func NewExecutiveHandler(service *executiveapp.Service) *ExecutiveHandler {
	return &ExecutiveHandler{service: service}
}

// handleSearchExecutives lists either the hourly candidate market or the
// authenticated company's own roster.
func (h *ExecutiveHandler) handleSearchExecutives(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "company not authenticated", nil)
		return
	}
	var req struct {
		Scope string `json:"scope,omitempty"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Scope == "mine" {
		executives, err := h.service.MyExecutives(r.Context(), companyID)
		if err != nil {
			writeAppErr(w, err)
			return
		}
		writeSuccess(w, http.StatusOK, map[string]any{"executives": executives, "total": len(executives)})
		return
	}
	candidates := h.service.MarketCandidates()
	writeSuccess(w, http.StatusOK, map[string]any{
		"executives":      candidates,
		"total":           len(candidates),
		"refreshCooldown": "01:00:00",
	})
}

func (h *ExecutiveHandler) handleRecruitExecutive(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "company not authenticated", nil)
		return
	}
	var req struct {
		ExecutiveID string `json:"executiveId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ExecutiveID == "" {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "executiveId is required", nil)
		return
	}
	executive, cost, err := h.service.Recruit(r.Context(), companyID, req.ExecutiveID)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, map[string]any{"ok": true, "executive": executive, "cost": cost})
}

func (h *ExecutiveHandler) handleTrainExecutive(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "company not authenticated", nil)
		return
	}
	executiveID := chi.URLParam(r, "executiveId")
	if executiveID == "" {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "executiveId is required", nil)
		return
	}
	executive, cost, err := h.service.Train(r.Context(), companyID, executiveID)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, map[string]any{"ok": true, "executive": executive, "cost": cost})
}

func (h *ExecutiveHandler) handleAssignExecutivePosition(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "company not authenticated", nil)
		return
	}
	executiveID := chi.URLParam(r, "executiveId")
	var req struct {
		Position company.ExecutivePosition `json:"position"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || executiveID == "" {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "position and executiveId are required", nil)
		return
	}
	executive, err := h.service.AssignPosition(r.Context(), companyID, executiveID, req.Position)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, map[string]any{"ok": true, "executive": executive})
}

func (h *ExecutiveHandler) handleGetExecutiveDetail(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "company not authenticated", nil)
		return
	}
	executiveID := chi.URLParam(r, "id")
	if executiveID == "" {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "executive id is required", nil)
		return
	}
	executive, err := h.service.Detail(r.Context(), companyID, executiveID)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, executive)
}
