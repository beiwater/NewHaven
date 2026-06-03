package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (h *Handler) RegisterGovernment(mux *http.ServeMux) {
	mux.HandleFunc("/api/v3/government-orders/", h.withAuth(h.handleGovernmentOrders))
	mux.HandleFunc("/api/v3/government-orders/bid/", h.withAuth(h.handleGovernmentBid))
	mux.HandleFunc("/api/v3/government-orders/award/", h.withAuth(h.handleGovernmentAward))
	mux.HandleFunc("/api/v3/government-orders/deliver/", h.withAuth(h.handleGovernmentDeliver))
	mux.HandleFunc("/api/v3/government-orders/resolve-defaults/", h.withAuth(h.handleGovernmentResolveDefaults))
}

func (h *Handler) handleGovernmentOrders(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, h.svc.Snapshot().GovernmentContracts)
}

func (h *Handler) handleGovernmentBid(w http.ResponseWriter, r *http.Request) {
	if cached, ok := h.svc.CheckIdempotent(requestIDFromBody(r)); ok {
		writeJSON(w, 200, cached)
		return
	}
	if r.Method != http.MethodPost {
		writeErr(w, 405, "method not allowed")
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "invalid json")
		return
	}
	resp, err := h.svc.PlaceGovernmentBid(h.companyID(r), fmt.Sprint(body["contractId"]), floatFromAny(body["unitPrice"]))
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	h.svc.MarkIdempotent(requestIDFromBody(r), resp)
	writeJSON(w, 200, resp)
}

func (h *Handler) handleGovernmentAward(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "method not allowed")
		return
	}
	writeJSON(w, 200, h.svc.AwardGovernmentContracts())
}

func (h *Handler) handleGovernmentDeliver(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "method not allowed")
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "invalid json")
		return
	}
	resp, err := h.svc.DeliverGovernmentContract(h.companyID(r), fmt.Sprint(body["contractId"]))
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, resp)
}

func (h *Handler) handleGovernmentResolveDefaults(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "method not allowed")
		return
	}
	writeJSON(w, 200, h.svc.ResolveGovernmentDefaults())
}
