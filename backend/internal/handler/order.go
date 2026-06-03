package handler

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) RegisterOrder(mux *http.ServeMux) {
	mux.HandleFunc("/api/v2/orders/daily/", h.withAuth(h.handleDailyOrders))
	mux.HandleFunc("/api/v2/orders/daily/complete/", h.withAuth(h.handleCompleteOrder))
	mux.HandleFunc("/api/v2/orders/daily/claim/", h.withAuth(h.handleClaimOrder))
}

func (h *Handler) handleDailyOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, 405, "method not allowed")
		return
	}
	writeJSON(w, 200, h.svc.DailyOrders())
}

func (h *Handler) handleCompleteOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "method not allowed")
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "invalid json")
		return
	}
	orderID, _ := body["orderId"].(string)
	if orderID == "" {
		writeErr(w, 400, "orderId required")
		return
	}
	resp, err := h.svc.CompleteDailyOrder(h.companyID(r), orderID)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, resp)
}

func (h *Handler) handleClaimOrder(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "method not allowed")
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "invalid json")
		return
	}
	orderID, _ := body["orderId"].(string)
	if orderID == "" {
		writeErr(w, 400, "orderId required")
		return
	}
	resp, err := h.svc.ClaimDailyOrderReward(h.companyID(r), orderID)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, resp)
}
