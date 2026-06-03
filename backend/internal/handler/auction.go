package handler

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (h *Handler) handleAuctions(w http.ResponseWriter, r *http.Request) {
	remainder := strings.TrimPrefix(r.URL.Path, "/api/v2/auctions/")
	remainder = strings.Trim(remainder, "/")
	if remainder == "" {
		// GET /api/v2/auctions/ -> list
		list := h.svc.AvailableAuctionList()
		writeJSON(w, 200, map[string]any{"auctions": list, "total": len(list)})
		return
	}
	parts := strings.SplitN(remainder, "/", 2)
	auctionID := parts[0]
	if len(parts) == 1 && auctionID != "" {
		// GET /api/v2/auctions/{id}/
		detail, err := h.svc.AuctionDetail(auctionID)
		if err != nil {
			writeErr(w, 404, err.Error())
			return
		}
		writeJSON(w, 200, detail)
		return
	}
	if len(parts) == 2 && parts[1] == "bid" {
		// POST /api/v2/auctions/{id}/bid/
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
		result, err := h.svc.PlaceAuctionBid(h.companyID(r), auctionID, floatFromAny(body["amount"]))
		if err != nil {
			writeErr(w, 400, err.Error())
			return
		}
		h.svc.MarkIdempotent(requestIDFromBody(r), result)
		writeJSON(w, 200, result)
		return
	}
	writeErr(w, 404, "not found")
}
