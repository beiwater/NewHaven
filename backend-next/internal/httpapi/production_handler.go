package httpapi

import (
	"net/http"

	"github.com/newhaven/backend-next/internal/app/production"
)

// ProductionHandler handles production-related HTTP endpoints.
type ProductionHandler struct {
	svc *production.Service
}

// NewProductionHandler creates a new ProductionHandler.
func NewProductionHandler(svc *production.Service) *ProductionHandler {
	return &ProductionHandler{svc: svc}
}

// handleListProductionJobs returns production jobs for the authenticated company.
func (h *ProductionHandler) handleListProductionJobs(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	resp, err := h.svc.ListProductionJobs(r.Context(), companyID)
	if err != nil {
		writeErr(w, 500, ErrorInternal, "failed to list production jobs", nil)
		return
	}

	writeSuccess(w, 200, resp)
}
