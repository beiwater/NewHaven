package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	appfinance "github.com/newhaven/backend-next/internal/app/finance"
	openapi "github.com/newhaven/backend-next/internal/generated/openapi"
)

// BondHandler handles bond-related HTTP endpoints.
type BondHandler struct {
	svc *appfinance.Service
}

// NewBondHandler creates a new BondHandler.
func NewBondHandler(svc *appfinance.Service) *BondHandler {
	return &BondHandler{svc: svc}
}

// handleListBonds returns all active bonds.
func (h *BondHandler) handleListBonds(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.ListBonds(r.Context(), r.URL.Query().Get("rating"))
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeSuccess(w, 200, resp)
}

// handleGetBond returns a single bond by ID.
func (h *BondHandler) handleGetBond(w http.ResponseWriter, r *http.Request) {
	bondID := chi.URLParam(r, "bondId")
	if bondID == "" {
		writeErr(w, 400, ErrorBadRequest, "missing bond ID", nil)
		return
	}

	resp, err := h.svc.GetBond(r.Context(), bondID)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeSuccess(w, 200, resp)
}

// handleCreateBond issues a new bond.
func (h *BondHandler) handleCreateBond(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	var req openapi.CreateBondRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, ErrorBadRequest, "invalid request body", nil)
		return
	}
	if req.Amount < 1 {
		writeErr(w, 400, ErrorValidation, "amount must be >= 1", nil)
		return
	}
	if req.Interest <= 0 {
		writeErr(w, 400, ErrorValidation, "interest must be positive", nil)
		return
	}

	resp, err := h.svc.CreateBond(r.Context(), companyID, req.Amount, req.Interest)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeSuccess(w, 200, resp)
}

// handleGetOwnedBonds returns bonds held by the company.
func (h *BondHandler) handleGetOwnedBonds(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	resp, err := h.svc.GetOwnedBonds(r.Context(), companyID)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeSuccess(w, 200, resp)
}

// handleGetSoldBonds returns bonds issued by the company.
func (h *BondHandler) handleGetSoldBonds(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	resp, err := h.svc.GetSoldBonds(r.Context(), companyID)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeSuccess(w, 200, resp)
}

// handleBuyBond buys a bond.
func (h *BondHandler) handleBuyBond(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}
	var req struct {
		BondID string `json:"bondId"`
		Amount int    `json:"amount"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.BondID == "" || req.Amount <= 0 {
		writeErr(w, 400, ErrorValidation, "bondId and amount are required", nil)
		return
	}
	resp, err := h.svc.BuyBond(r.Context(), companyID, req.BondID, req.Amount)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeSuccess(w, 200, resp)
}

// handleCallBond calls a bond (issuer only).
func (h *BondHandler) handleCallBond(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}
	bondID := chi.URLParam(r, "bondId")
	if bondID == "" {
		writeErr(w, 400, ErrorValidation, "bondId is required", nil)
		return
	}
	var req struct {
		Amount int `json:"amount"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)
	if req.Amount <= 0 {
		req.Amount = 1
	}
	resp, err := h.svc.CallBond(r.Context(), companyID, bondID, req.Amount)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeSuccess(w, 200, resp)
}

// handleSettleBondInterest pays interest to all bond holders.
func (h *BondHandler) handleSettleBondInterest(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}
	_ = companyID // any authenticated company can trigger
	resp, err := h.svc.SettleBondInterest(r.Context())
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeSuccess(w, 200, resp)
}
