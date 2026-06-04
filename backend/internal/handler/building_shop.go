package handler

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) RegisterBuildingShop(mux *http.ServeMux) {
	mux.HandleFunc("/api/v2/buildings/market/", h.withAuth(h.handleBuildingMarket))
	mux.HandleFunc("/api/v2/buildings/buy/", h.withAuth(h.handleBuyBuilding))
	mux.HandleFunc("/api/v2/buildings/place/", h.withAuth(h.handlePlaceBuilding))
	mux.HandleFunc("/api/v2/buildings/move/", h.withAuth(h.handleMoveBuilding))
	mux.HandleFunc("/api/v2/buildings/demolish/", h.withAuth(h.handleDemolishBuilding))
	mux.HandleFunc("/api/v2/companies/me/warehouse/upgrade/", h.withAuth(h.handleWarehouseUpgrade))
}

func (h *Handler) handleBuildingMarket(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, h.svc.BuildingMarket())
}

type BuyBuildingReq struct {
	RequestID  string `json:"requestId"`
	BuildingID string `json:"buildingId"`
}

func (h *Handler) handleBuyBuilding(w http.ResponseWriter, r *http.Request) {
	if cached, ok := h.svc.CheckIdempotent(requestIDFromBody(r)); ok {
		writeJSON(w, 200, cached)
		return
	}
	var req BuyBuildingReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.BuildingID == "" {
		writeErr(w, 400, "buildingId required")
		return
	}
	resp, err := h.svc.BuyBuilding(h.companyID(r), req.BuildingID)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	h.svc.MarkIdempotent(req.RequestID, resp)
	writeJSON(w, 200, resp)
}

type PlaceBuildingReq struct {
	BuildingID string `json:"buildingId"`
	MapID      string `json:"mapId"`
	SlotID     string `json:"slotId"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
}

func (h *Handler) handlePlaceBuilding(w http.ResponseWriter, r *http.Request) {
	var req PlaceBuildingReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, "invalid request")
		return
	}
	resp, err := h.svc.PlaceBuilding(h.companyID(r), req.BuildingID, req.MapID, req.SlotID, req.X, req.Y)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, resp)
}

type MoveBuildingReq struct {
	BuildingID string `json:"buildingId"`
	MapID      string `json:"mapId"`
	SlotID     string `json:"slotId"`
	X          int    `json:"x"`
	Y          int    `json:"y"`
}

func (h *Handler) handleMoveBuilding(w http.ResponseWriter, r *http.Request) {
	var req MoveBuildingReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, "invalid request")
		return
	}
	resp, err := h.svc.MoveBuilding(h.companyID(r), req.BuildingID, req.MapID, req.SlotID, req.X, req.Y)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, resp)
}

type DemolishBuildingReq struct {
	BuildingID string `json:"buildingId"`
}

func (h *Handler) handleDemolishBuilding(w http.ResponseWriter, r *http.Request) {
	var req DemolishBuildingReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.BuildingID == "" {
		writeErr(w, 400, "buildingId required")
		return
	}
	resp, err := h.svc.DemolishBuilding(h.companyID(r), req.BuildingID)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, resp)
}

func (h *Handler) handleWarehouseUpgrade(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.WarehouseUpgrade(h.companyID(r))
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, resp)
}
