package httpapi

import (
	"net/http"

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
