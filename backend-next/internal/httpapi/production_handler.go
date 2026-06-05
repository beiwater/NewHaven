package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"

	openapi "github.com/newhaven/backend-next/internal/generated/openapi"

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

// handleStartProduction starts a new production job.
func (h *ProductionHandler) handleStartProduction(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	var req openapi.StartProductionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, ErrorBadRequest, "invalid request body", nil)
		return
	}

	if req.BuildingId == "" || req.ResourceId == 0 || req.Quantity < 1 {
		writeErr(w, 400, ErrorValidation, "building_id, resource_id, and quantity (min 1) are required", nil)
		return
	}

	resp, err := h.svc.StartProduction(r.Context(), companyID, &req)
	if err != nil {
		// Map known error types to appropriate HTTP status codes
		switch {
		case strings.Contains(err.Error(), "not found"):
			writeErr(w, 404, ErrorNotFound, err.Error(), nil)
		case strings.Contains(err.Error(), "cannot produce"):
			writeErr(w, 400, ErrorBadRequest, err.Error(), nil)
		case strings.Contains(err.Error(), "insufficient inventory"):
			writeErr(w, 400, ErrorInsufficientInv, err.Error(), nil)
		case strings.Contains(err.Error(), "exceeds maximum"):
			writeErr(w, 400, ErrorValidation, err.Error(), nil)
		case strings.Contains(err.Error(), "invalid production rate"):
			writeErr(w, 400, ErrorValidation, err.Error(), nil)
		default:
			writeErr(w, 500, ErrorInternal, "failed to start production", nil)
		}
		return
	}

	writeSuccess(w, 200, resp)
}
