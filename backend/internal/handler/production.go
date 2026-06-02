package handler

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (h *Handler) RegisterProduction(mux *http.ServeMux) {
	mux.HandleFunc("/api/v1/buildings/", h.withAuth(h.handleV1Buildings))
	mux.HandleFunc("/api/v2/buildings/", h.withAuth(h.handleV2Buildings))
	mux.HandleFunc("/api/v2/production/jobs/", h.withAuth(h.handleProductionJobs))
	mux.HandleFunc("/api/v2/production/claim/", h.withAuth(h.handleClaimProduction))
	mux.HandleFunc("/api/v2/production/claimable/", h.withAuth(h.handleClaimable))
	mux.HandleFunc("/api/v2/production/claim-all/", h.withAuth(h.handleClaimAll))
}

func (h *Handler) handleV2Buildings(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v2/buildings/"), "/")
	parts := strings.Split(path, "/")
	if len(parts) == 2 && parts[1] == "production-options" {
		writeJSON(w, 200, h.svc.ProductionOptions(h.companyID(r), parts[0]))
		return
	}
	writeErr(w, 404, "unknown building action")
}

func (h *Handler) handleV1Buildings(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	switch {
	case strings.HasSuffix(strings.TrimSuffix(path, "/"), "/busy"):
		h.handleBuildingBusy(w, r)
	case strings.HasSuffix(strings.TrimSuffix(path, "/"), "/upgrade"):
		h.handleBuildingUpgrade(w, r)
	default:
		writeErr(w, 404, "unknown building action")
	}
}

func (h *Handler) handleBuildingBusy(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !strings.HasSuffix(r.URL.Path, "/busy/") {
		writeErr(w, 404, "not found")
		return
	}
	body := map[string]any{}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	path := strings.TrimSuffix(r.URL.Path, "/")
	path = strings.TrimSuffix(path, "/busy")
	id := strings.TrimPrefix(path, "/api/v1/buildings/")
	resp := h.svc.StartBuildingProduction(h.companyID(r), id, body)
	if errMsg, ok := resp["error"].(string); ok && errMsg != "" {
		writeErr(w, 400, errMsg)
		return
	}
	writeJSON(w, 200, resp)
}
func (h *Handler) handleBuildingUpgrade(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "method not allowed")
		return
	}
	path := strings.TrimSuffix(r.URL.Path, "/")
	path = strings.TrimSuffix(path, "/upgrade")
	id := strings.TrimPrefix(path, "/api/v1/buildings/")
	resp, err := h.svc.UpgradeBuilding(h.companyID(r), id)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, resp)
}

func (h *Handler) handleProductionJobs(w http.ResponseWriter, r *http.Request) {
	h.svc.RefreshProductionJobs(h.companyID(r))
	writeJSON(w, 200, h.svc.Snapshot().ProductionJobs)
}

func (h *Handler) handleClaimProduction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "method not allowed")
		return
	}
	if cached, ok := h.svc.CheckIdempotent(requestIDFromBody(r)); ok {
		writeJSON(w, 200, cached)
		return
	}
	jobID := strings.TrimPrefix(r.URL.Path, "/api/v2/production/claim/")
	jobID = strings.Trim(jobID, "/")
	result, err := h.svc.ClaimProduction(h.companyID(r), jobID)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	h.svc.MarkIdempotent(requestIDFromBody(r), result)
	writeJSON(w, 200, result)
}

func (h *Handler) handleClaimable(w http.ResponseWriter, r *http.Request) {
	claimable := make([]map[string]any, 0)
	h.svc.RefreshProductionJobs(h.companyID(r))
	for _, j := range h.svc.Snapshot().ProductionJobs {
		if j.Status == "ready" {
			claimable = append(claimable, map[string]any{
				"jobId":      j.ID,
				"buildingId": j.BuildingID,
				"resourceId": j.ResourceID,
				"amount":     j.Amount,
				"quality":    j.Quality,
			})
		}
	}
	writeJSON(w, 200, claimable)
}

func (h *Handler) handleClaimAll(w http.ResponseWriter, r *http.Request) {
	h.svc.RefreshProductionJobs(h.companyID(r))
	claimed := make([]map[string]any, 0)
	errors := make([]string, 0)
	for _, j := range h.svc.Snapshot().ProductionJobs {
		if j.Status == "ready" {
			if result, err := h.svc.ClaimProduction(h.companyID(r), j.ID); err != nil {
				errors = append(errors, err.Error())
			} else {
				claimed = append(claimed, result)
			}
		}
	}
	writeJSON(w, 200, map[string]any{
		"claimed": claimed,
		"errors":  errors,
		"total":   len(claimed),
	})
}
