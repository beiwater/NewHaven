package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
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
		writeAppErr(w, err)
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
		writeAppErr(w, err)
		return
	}

	writeSuccess(w, 200, resp)
}

// handleClaimProduction claims produced resources from a production job.
func (h *ProductionHandler) handleClaimProduction(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	jobID := chi.URLParam(r, "jobId")
	if jobID == "" {
		writeErr(w, 400, ErrorValidation, "jobId path parameter is required", nil)
		return
	}

	resp, err := h.svc.ClaimProduction(r.Context(), companyID, jobID)
	if err != nil {
		writeAppErr(w, err)
		return
	}

	writeSuccess(w, 200, resp)
}

// handleListClaimableJobs returns claimable production jobs.
func (h *ProductionHandler) handleListClaimableJobs(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	resp, err := h.svc.ListClaimableJobs(r.Context(), companyID)
	if err != nil {
		writeAppErr(w, err)
		return
	}

	writeSuccess(w, 200, resp)
}
