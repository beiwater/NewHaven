package httpapi

import (
	"net/http"

	"github.com/newhaven/backend-next/internal/app/warehouse"
)

// WarehouseHandler handles warehouse-related HTTP endpoints.
type WarehouseHandler struct {
	svc *warehouse.Service
}

// NewWarehouseHandler creates a new WarehouseHandler.
func NewWarehouseHandler(svc *warehouse.Service) *WarehouseHandler {
	return &WarehouseHandler{svc: svc}
}

// handleGetMyWarehouse returns the warehouse for the authenticated company.
func (h *WarehouseHandler) handleGetMyWarehouse(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	resp, err := h.svc.GetMyWarehouse(r.Context(), companyID)
	if err != nil {
		writeAppErr(w, err)
		return
	}

	writeSuccess(w, 200, resp)
}
