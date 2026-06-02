package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func (h *Handler) RegisterMarket(mux *http.ServeMux) {
	mux.HandleFunc("/api/v3/market-ticker/", h.withAuth(h.handleMarketTicker))
	mux.HandleFunc("/api/v3/market/", h.withAuth(h.handleMarketOrderbook))
	mux.HandleFunc("/api/v2/market-order/", h.withAuth(h.handleCreateOrder))
	mux.HandleFunc("/api/v2/market-order/cancel/", h.withAuth(h.handleCancelOrder))
	mux.HandleFunc("/api/market/buy/orders/", h.withAuth(h.handleMarketBuyOrders))
	mux.HandleFunc("/api/v2/market-order/take/", h.withAuth(h.handleTakeOrder))
	mux.HandleFunc("/api/v2/weather/", h.withAuth(h.handleWeather))
	mux.HandleFunc("/api/v3/market-depth/", h.withAuth(h.handleMarketDepth))
	mux.HandleFunc("/api/v2/production-modifiers/", h.withAuth(h.handleProductionModifiers))
	mux.HandleFunc("/api/v3/resources/", h.withAuth(h.handleResources))
	mux.HandleFunc("/api/v3/resources-info/", h.withAuth(h.handleResourceInfo))
}

func (h *Handler) handleMarketTicker(w http.ResponseWriter, r *http.Request) {
	id, err := tailInt(r.URL.Path)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	h.svc.RunBotMarketCycle()
	writeJSON(w, 200, h.svc.BuildTicker(id))
}

func (h *Handler) handleMarketOrderbook(w http.ResponseWriter, r *http.Request) {
	h.svc.RunBotMarketCycle()
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v3/market/"), "/"), "/")
	if len(parts) != 2 {
		writeErr(w, 400, "expect /api/v3/market/{resource}/{quality}/")
		return
	}
	resID, _ := strconv.Atoi(parts[0])
	quality, _ := strconv.Atoi(parts[1])
	out := []any{}
	for _, o := range h.svc.Snapshot().Orders {
		if o.ResourceID == resID && o.Quality == quality {
			out = append(out, o)
		}
	}
	writeJSON(w, 200, out)
}

type CreateOrderRequest struct {
	RequestID  string  `json:"requestId"`
	ResourceID int     `json:"resourceId"`
	Kind       int     `json:"kind"`
	Quality    int     `json:"quality"`
	Quantity   int     `json:"quantity"`
	Price      float64 `json:"price"`
}

func (h *Handler) handleCreateOrder(w http.ResponseWriter, r *http.Request) {
	if cached, ok := h.svc.CheckIdempotent(requestIDFromBody(r)); ok {
		writeJSON(w, 200, cached)
		return
	}
	var req CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, "invalid json")
		return
	}
	result, err := h.svc.CreateOrder(h.companyID(r), req.ResourceID, req.Kind, req.Quality, req.Quantity, req.Price)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	h.svc.MarkIdempotent(req.RequestID, result)
	writeJSON(w, 200, result)
}

func (h *Handler) handleCancelOrder(w http.ResponseWriter, r *http.Request) {
	if cached, ok := h.svc.CheckIdempotent(requestIDFromBody(r)); ok {
		writeJSON(w, 200, cached)
		return
	}
	if r.Method != http.MethodDelete {
		writeErr(w, 405, "method not allowed")
		return
	}
	orderID := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v2/market-order/cancel/"), "/")
	resp, err := h.svc.CancelOrder(h.companyID(r), orderID)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	h.svc.MarkIdempotent(requestIDFromBody(r), resp)
	writeJSON(w, 200, resp)
}

func (h *Handler) handleMarketBuyOrders(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "method not allowed")
		return
	}
	parts := strings.Split(strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/market/buy/orders/"), "/"), "/")
	if len(parts) < 2 {
		writeErr(w, 400, "expect /api/market/buy/orders/{resource}/{quality}")
		return
	}
	resID, _ := strconv.Atoi(parts[0])
	quality, _ := strconv.Atoi(parts[1])
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "invalid json")
		return
	}
	resp, err := h.svc.CreateOrder(h.companyID(r), resID, 1, quality, intFromAny(body["amount"]), floatFromAny(body["price"]))
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, resp)
}

type TakeOrderRequest struct {
	RequestID string  `json:"requestId"`
	Resource  int     `json:"resource"`
	Quantity  int     `json:"quantity"`
	Quality   int     `json:"quality"`
	MaxPrice  float64 `json:"maxPrice"`
}

func (h *Handler) handleTakeOrder(w http.ResponseWriter, r *http.Request) {
	if cached, ok := h.svc.CheckIdempotent(requestIDFromBody(r)); ok {
		writeJSON(w, 200, cached)
		return
	}
	var body TakeOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "invalid json")
		return
	}
	resp, err := h.svc.TakeOrder(h.companyID(r), body.Resource, body.Quantity, body.Quality, body.MaxPrice)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	h.svc.MarkIdempotent(body.RequestID, resp)
	writeJSON(w, 200, resp)
}

func (h *Handler) handleWeather(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]any{
		"id": 0, "realm": 0,
		"since":                  time.Now().Add(-2 * time.Hour).UTC().Format(time.RFC3339),
		"until":                  time.Now().Add(4 * time.Hour).UTC().Format(time.RFC3339),
		"sellingSpeedMultiplier": h.svc.Cfg.Game.WeatherSpeedMult,
	})
}

func (h *Handler) handleProductionModifiers(w http.ResponseWriter, _ *http.Request) {
	modifiers := make([]map[string]any, 0, len(h.svc.Data.Resources))
	for _, it := range h.svc.Data.Resources {
		rid := intFromAny(it["dbLetter"])
		if rid <= 0 {
			continue
		}
		modifiers = append(modifiers, map[string]any{
			"resource": rid,
			"modifier": h.svc.Cfg.Game.ProductionMod,
		})
	}
	writeJSON(w, 200, map[string]any{"resourceProductionModifiers": modifiers})
}

func (h *Handler) handleMarketDepth(w http.ResponseWriter, r *http.Request) {
	h.svc.RunBotMarketCycle()
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	// path: api/v3/market-depth/{resource}/{quality}/
	if len(parts) < 4 {
		writeErr(w, 400, "path should be /api/v3/market-depth/{resource}/{quality}/")
		return
	}
	resID, _ := strconv.Atoi(parts[len(parts)-2])
	qual, _ := strconv.Atoi(parts[len(parts)-1])
	writeJSON(w, 200, h.svc.OrderBookDepth(resID, qual))
}

func (h *Handler) handleResources(w http.ResponseWriter, r *http.Request) {
	path := strings.Trim(strings.TrimPrefix(r.URL.Path, "/api/v3/resources/"), "/")
	if path == "" {
		out := make([]map[string]any, 0, len(h.svc.Data.Resources))
		for _, it := range h.svc.Data.Resources {
			rid := intFromAny(it["dbLetter"])
			if rid <= 0 {
				continue
			}
			if isResearch := boolFromAny(it["isResearch"]); isResearch {
				continue
			}
			if tradable := boolFromAny(it["isExchangeTradable"]); !tradable {
				continue
			}
			out = append(out, map[string]any{
				"resourceId":         rid,
				"name":               it["name"],
				"producedFrom":       it["producedFrom"],
				"producedPerHourRaw": it["producedPerHourRaw"],
				"unitsSoldAnHour":    it["unitsSoldAnHour"],
				"hasEconomyModel":    it["hasEconomyModel"],
			})
		}
		writeJSON(w, 200, map[string]any{"resources": out})
		return
	}
	id, err := strconv.Atoi(path)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	for _, it := range h.svc.Data.Resources {
		if intFromAny(it["dbLetter"]) == id {
			writeJSON(w, 200, it)
			return
		}
	}
	writeErr(w, 404, "resource not found")
}

func (h *Handler) handleResourceInfo(w http.ResponseWriter, r *http.Request) {
	id, err := tailInt(r.URL.Path)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, h.svc.ResourceInfo(id))
}
