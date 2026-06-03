package handler

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) RegisterProductionQueue(mux *http.ServeMux) {
	mux.HandleFunc("/api/v2/production/queue/", h.withAuth(h.handleProductionQueue))
	mux.HandleFunc("/api/v2/production/slots/add/", h.withAuth(h.handleAddSlot))
	mux.HandleFunc("/api/v2/production/cancel/", h.withAuth(h.handleCancelProduction))
}

func (h *Handler) handleProductionQueue(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, h.svc.ProductionQueue(h.companyID(r)))
}

func (h *Handler) handleAddSlot(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.AddProductionSlot(h.companyID(r))
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, resp)
}

type CancelJobReq struct {
	JobID string `json:"jobId"`
}

func (h *Handler) handleCancelProduction(w http.ResponseWriter, r *http.Request) {
	var req CancelJobReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.JobID == "" {
		writeErr(w, 400, "jobId required")
		return
	}
	resp, err := h.svc.CancelProductionJob(h.companyID(r), req.JobID)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, resp)
}
