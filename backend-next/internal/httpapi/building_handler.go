package httpapi

import (
	"encoding/json"
	"github.com/go-chi/chi/v5"
	openapi "github.com/newhaven/backend-next/internal/generated/openapi"
	"net/http"
	"strconv"

	"github.com/newhaven/backend-next/internal/app/building"
)

// BuildingHandler handles building-related HTTP endpoints.
type BuildingHandler struct {
	svc *building.Service
}

// NewBuildingHandler creates a new BuildingHandler.
func NewBuildingHandler(svc *building.Service) *BuildingHandler {
	return &BuildingHandler{svc: svc}
}

// handleListMyBuildings returns buildings for the authenticated company.
func (h *BuildingHandler) handleListMyBuildings(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	resp, err := h.svc.ListMyBuildings(r.Context(), companyID)
	if err != nil {
		writeAppErr(w, err)
		return
	}

	writeSuccess(w, 200, resp)
}

// handleListMyBuildingsV2 is a transitional /api/v2/companies/me/buildings/ endpoint.
// It currently delegates to handleListMyBuildings which returns snake_case BuildingDTO.
// TODO: migrate BuildingDTO to camelCase for v2 frontend compatibility.
func (h *BuildingHandler) handleListMyBuildingsV2(w http.ResponseWriter, r *http.Request) {
	h.handleListMyBuildings(w, r)
}

// handleListCompanyBuildings returns buildings for a specific company (by companyId).
func (h *BuildingHandler) handleListCompanyBuildings(w http.ResponseWriter, r *http.Request) {
	companyID, err := strconv.Atoi(chi.URLParam(r, "companyId"))
	if err != nil {
		writeErr(w, 400, ErrorValidation, "invalid company id", nil)
		return
	}

	resp, err := h.svc.ListMyBuildings(r.Context(), companyID)
	if err != nil {
		writeAppErr(w, err)
		return
	}

	writeSuccess(w, 200, resp)
}

// handleBuildingMarket returns the building market list wrapped in the standard envelope.
func (h *BuildingHandler) handleBuildingMarket(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}
	_ = companyID
	resp, err := h.svc.BuildingMarket(r.Context())
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeSuccess(w, 200, resp)
}

// handleBuyBuilding purchases a building from the market.
func (h *BuildingHandler) handleBuyBuilding(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	var req openapi.BuyBuildingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, ErrorBadRequest, "invalid request body", nil)
		return
	}

	resp, err := h.svc.BuyBuilding(r.Context(), companyID, req.BuildingId)
	if err != nil {
		writeAppErr(w, err)
		return
	}

	writeSuccess(w, 200, resp)
}

// handlePlaceBuilding places a building on the map.
func (h *BuildingHandler) handlePlaceBuilding(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	var req openapi.PlaceBuildingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, ErrorBadRequest, "invalid request body", nil)
		return
	}

	resp, err := h.svc.PlaceBuilding(r.Context(), companyID, &req)
	if err != nil {
		writeAppErr(w, err)
		return
	}

	writeSuccess(w, 200, resp)
}

// handleMoveBuilding moves a building to a new position.
func (h *BuildingHandler) handleMoveBuilding(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	var req openapi.MoveBuildingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, ErrorBadRequest, "invalid request body", nil)
		return
	}

	resp, err := h.svc.MoveBuilding(r.Context(), companyID, &req)
	if err != nil {
		writeAppErr(w, err)
		return
	}

	writeSuccess(w, 200, resp)
}

// handleDemolishBuilding demolishes a building and refunds part of the cost.
func (h *BuildingHandler) handleDemolishBuilding(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	var req openapi.DemolishBuildingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, ErrorBadRequest, "invalid request body", nil)
		return
	}

	resp, err := h.svc.DemolishBuilding(r.Context(), companyID, &req)
	if err != nil {
		writeAppErr(w, err)
		return
	}

	writeSuccess(w, 200, resp)
}

// handleStashBuilding stashes a placed building back into inventory.
func (h *BuildingHandler) handleStashBuilding(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "company not authenticated", nil)
		return
	}
	var req struct {
		BuildingID string `json:"buildingId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.BuildingID == "" {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "buildingId is required", nil)
		return
	}
	resp, err := h.svc.StashBuilding(r.Context(), companyID, req.BuildingID)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeSuccess(w, http.StatusOK, resp)
}

// handleUpgradeBuilding upgrades a building to the next level.
func (h *BuildingHandler) handleUpgradeBuilding(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	buildingID := chi.URLParam(r, "buildingId")
	if buildingID == "" {
		writeErr(w, 400, ErrorBadRequest, "building ID is required", nil)
		return
	}

	resp, err := h.svc.UpgradeBuilding(r.Context(), companyID, buildingID)
	if err != nil {
		writeAppErr(w, err)
		return
	}

	writeSuccess(w, 200, resp)
}

// handleStockShelf stocks items into a retail building's shelf.
func (h *BuildingHandler) handleStockShelf(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	buildingID := chi.URLParam(r, "buildingId")
	if buildingID == "" {
		writeErr(w, 400, ErrorBadRequest, "building ID is required", nil)
		return
	}

	var req openapi.StockShelfRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, ErrorBadRequest, "invalid request body", nil)
		return
	}

	resp, err := h.svc.StockShelf(r.Context(), companyID, buildingID, req.ResourceId, req.Quantity, req.Price)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeSuccess(w, 200, resp)
}

// handleUnstockShelf removes items from a retail building's shelf back to warehouse.
func (h *BuildingHandler) handleUnstockShelf(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	buildingID := chi.URLParam(r, "buildingId")
	if buildingID == "" {
		writeErr(w, 400, ErrorBadRequest, "building ID is required", nil)
		return
	}

	var req openapi.UnstockShelfRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, ErrorBadRequest, "invalid request body", nil)
		return
	}

	resp, err := h.svc.UnstockShelf(r.Context(), companyID, buildingID, req.ResourceId, req.Quantity)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeSuccess(w, 200, resp)
}

// handleSetShelfPrice sets the price for a shelf item.
func (h *BuildingHandler) handleSetShelfPrice(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	buildingID := chi.URLParam(r, "buildingId")
	if buildingID == "" {
		writeErr(w, 400, ErrorBadRequest, "building ID is required", nil)
		return
	}

	var req openapi.SetShelfPriceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, ErrorBadRequest, "invalid request body", nil)
		return
	}

	lock := false
	if req.Lock != nil {
		lock = *req.Lock
	}
	resp, err := h.svc.SetShelfPrice(r.Context(), companyID, buildingID, req.ResourceId, req.Price, lock)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeSuccess(w, 200, resp)
}
