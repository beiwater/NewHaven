package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

func (h *Handler) RegisterBond(mux *http.ServeMux) {
	mux.HandleFunc("/api/bonds/settle-interest/", h.withAuth(h.handleSettleBondInterest))
	mux.HandleFunc("/api/bonds/", h.withAuth(h.handleBonds))
	mux.HandleFunc("/api/v2/companies/me/bonds/owned/", h.withAuth(h.handleBondsOwned))
	mux.HandleFunc("/api/v2/companies/me/bonds/sold/", h.withAuth(h.handleBondsSold))
}

func (h *Handler) handleSettleBondInterest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "method not allowed")
		return
	}
	writeJSON(w, 200, h.svc.SettleBondInterest(h.companyID(r)))
}

func (h *Handler) handleBonds(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) >= 3 && parts[2] != "" {
		switch r.Method {
		case http.MethodGet:
			h.handleBondsDetail(w, r)
		case http.MethodPatch:
			h.handleBondsBuy(w, r)
		case http.MethodPut:
			h.handleBondsCall(w, r)
		default:
			writeErr(w, 405, "method not allowed")
		}
		return
	}
	switch r.Method {
	case http.MethodGet:
		h.handleBondsList(w, r)
	case http.MethodPatch:
		h.handleBondsCreate(w, r)
	default:
		writeErr(w, 405, "method not allowed")
	}
}

func (h *Handler) handleBondsList(w http.ResponseWriter, r *http.Request) {
	ratingFilter := r.URL.Query().Get("rating")
	if ratingFilter == "" {
		writeJSON(w, 200, h.svc.BondMarketView())
		return
	}
	filtered := []map[string]any{}
	for _, b := range h.svc.BondMarketView() {
		group := h.svc.BondRatingGroup(fmt.Sprint(b["ratingWhenPurchased"]))
		if group == ratingFilter {
			filtered = append(filtered, b)
		}
	}
	writeJSON(w, 200, filtered)
}

func (h *Handler) handleBondsDetail(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	bondID := parts[2]
	for _, b := range h.svc.BondMarketView() {
		if fmt.Sprint(b["id"]) == bondID {
			writeJSON(w, 200, b)
			return
		}
	}
	writeErr(w, 404, "bond not found")
}

func (h *Handler) handleBondsCreate(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "invalid json")
		return
	}
	resp, err := h.svc.IssueOrAdjustBond(h.companyID(r), intFromAny(body["amount"]), floatFromAny(body["interest"]))
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, resp)
}

func (h *Handler) handleBondsBuy(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	bondID := parts[2]
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "invalid json")
		return
	}
	resp, err := h.svc.BuyBond(h.companyID(r), bondID, intFromAny(body["amount"]))
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, resp)
}

func (h *Handler) handleBondsCall(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	bondID := parts[2]
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "invalid json")
		return
	}
	resp, err := h.svc.CallBond(h.companyID(r), bondID, intFromAny(body["amount"]))
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, resp)
}

func (h *Handler) handleBondsOwned(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, h.svc.BondMarketView())
}

func (h *Handler) handleBondsSold(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, []any{})
}
