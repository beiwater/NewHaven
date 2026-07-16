package httpapi

import (
	"encoding/json"
	"net/http"

	openapi "github.com/beiwater/NewHaven/backend/internal/generated/openapi"
	"github.com/go-chi/chi/v5"

	"github.com/beiwater/NewHaven/backend/internal/app/production"
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

// handleStartProductionV1 keeps the current frontend busy endpoint working
// while routing all behavior through the backend production service.
func (h *ProductionHandler) handleStartProductionV1(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	var req struct {
		Kind      int     `json:"kind"`
		Amount    int     `json:"amount"`
		Quality   *int    `json:"quality,omitempty"`
		RequestID *string `json:"requestId,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, ErrorBadRequest, "invalid request body", nil)
		return
	}
	buildingID := chi.URLParam(r, "buildingId")
	if buildingID == "" || req.Kind == 0 || req.Amount < 1 {
		writeErr(w, 400, ErrorValidation, "buildingId, kind, and amount (min 1) are required", nil)
		return
	}

	resp, err := h.svc.StartProduction(r.Context(), companyID, &openapi.StartProductionRequest{
		BuildingId: buildingID,
		ResourceId: req.Kind,
		Quantity:   req.Amount,
		Quality:    req.Quality,
		RequestId:  req.RequestID,
	})
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

func (h *ProductionHandler) handleProductionQueue(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}
	resp, err := h.svc.ProductionQueue(r.Context(), companyID)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeSuccess(w, 200, resp)
}

func (h *ProductionHandler) handleProductionOptions(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}
	buildingID := chi.URLParam(r, "buildingId")
	if buildingID == "" {
		writeErr(w, 400, ErrorBadRequest, "buildingId required", nil)
		return
	}
	resp, err := h.svc.ProductionOptions(r.Context(), companyID, buildingID)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeSuccess(w, 200, resp)
}

func (h *ProductionHandler) handleCancelProduction(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}
	var req struct {
		JobID string `json:"jobId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.JobID == "" {
		writeErr(w, 400, ErrorBadRequest, "jobId required", nil)
		return
	}
	resp, err := h.svc.CancelProductionJob(r.Context(), companyID, req.JobID)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeSuccess(w, 200, resp)
}

func (h *ProductionHandler) handleClaimAll(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, 401, ErrorUnauthorized, "company not authenticated", nil)
		return
	}
	resp, err := h.svc.ClaimAll(r.Context(), companyID)
	if err != nil {
		writeAppErr(w, err)
		return
	}
	writeSuccess(w, 200, resp)
}
