package httpapi

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/newhaven/backend-next/internal/app/market"
)

// MarketHandler handles market-related HTTP endpoints.
type MarketHandler struct {
	svc *market.Service
}

// NewMarketHandler creates a new MarketHandler.
func NewMarketHandler(svc *market.Service) *MarketHandler {
	return &MarketHandler{svc: svc}
}

// handleResources returns the list of market-tradable resources.
func (h *MarketHandler) handleResources(w http.ResponseWriter, r *http.Request) {
	_, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	resp, err := h.svc.ListResources(r.Context())
	if err != nil {
		writeErr(w, 500, ErrorInternal, "failed to list resources", nil)
		return
	}

	writeSuccess(w, 200, resp)
}

// handleMarketTicker returns ticker data for a resource.
func (h *MarketHandler) handleMarketTicker(w http.ResponseWriter, r *http.Request) {
	_, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	resourceIDStr := chi.URLParam(r, "resourceId")
	resourceID, err := strconv.Atoi(resourceIDStr)
	if err != nil || resourceID <= 0 {
		writeErr(w, 400, ErrorValidation, "invalid resourceId", nil)
		return
	}

	resp, err := h.svc.GetMarketTicker(r.Context(), resourceID)
	if err != nil {
		writeErr(w, 500, ErrorInternal, "failed to get market ticker", nil)
		return
	}

	writeSuccess(w, 200, resp)
}

// handleMarketDepth returns market depth for a resource and quality.
func (h *MarketHandler) handleMarketDepth(w http.ResponseWriter, r *http.Request) {
	_, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	resourceIDStr := chi.URLParam(r, "resourceId")
	resourceID, err := strconv.Atoi(resourceIDStr)
	if err != nil || resourceID <= 0 {
		writeErr(w, 400, ErrorValidation, "invalid resourceId", nil)
		return
	}

	qualityStr := chi.URLParam(r, "quality")
	quality, err := strconv.Atoi(qualityStr)
	if err != nil || quality < 0 {
		writeErr(w, 400, ErrorValidation, "invalid quality", nil)
		return
	}

	resp, err := h.svc.GetMarketDepth(r.Context(), resourceID, quality)
	if err != nil {
		writeErr(w, 500, ErrorInternal, "failed to get market depth", nil)
		return
	}

	writeSuccess(w, 200, resp)
}

// handleMarketOrders returns market orders for a resource and quality.
func (h *MarketHandler) handleMarketOrders(w http.ResponseWriter, r *http.Request) {
	_, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	resourceIDStr := chi.URLParam(r, "resourceId")
	resourceID, err := strconv.Atoi(resourceIDStr)
	if err != nil || resourceID <= 0 {
		writeErr(w, 400, ErrorValidation, "invalid resourceId", nil)
		return
	}

	qualityStr := chi.URLParam(r, "quality")
	quality, err := strconv.Atoi(qualityStr)
	if err != nil || quality < 0 {
		writeErr(w, 400, ErrorValidation, "invalid quality", nil)
		return
	}

	resp, err := h.svc.ListMarketOrders(r.Context(), resourceID, quality)
	if err != nil {
		writeErr(w, 500, ErrorInternal, "failed to list market orders", nil)
		return
	}

	writeSuccess(w, 200, resp)
}
