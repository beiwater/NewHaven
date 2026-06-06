package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/newhaven/backend-next/internal/apperr"
	"github.com/newhaven/backend-next/internal/domain/research"
	"github.com/newhaven/backend-next/internal/storage"
)

// ResearchHandler handles research project endpoints.
type ResearchHandler struct {
	companies storage.CompanyStorage
	research  storage.ResearchStorage
	clock     func() time.Time
}

// NewResearchHandler creates a new ResearchHandler.
func NewResearchHandler(companies storage.CompanyStorage, research storage.ResearchStorage) *ResearchHandler {
	return &ResearchHandler{
		companies: companies,
		research:  research,
		clock:     time.Now,
	}
}

// handleListResearch returns the catalog of all research projects with per-company status.
func (h *ResearchHandler) handleListResearch(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "company not authenticated", nil)
		return
	}
	projects, err := h.research.GetProjects(r.Context())
	if err != nil {
		writeAppErr(w, apperr.WrapMsg(apperr.KindInternal, "failed to load research projects", err))
		return
	}
	progress, err := h.research.GetCompanyProgress(r.Context(), companyID)
	if err != nil {
		writeAppErr(w, apperr.WrapMsg(apperr.KindInternal, "failed to load research progress", err))
		return
	}

	// Build progress map keyed by project ID.
	type progEntry struct {
		status      string
		startedAt   string
		completesAt string
	}
	progressMap := make(map[string]progEntry)
	for _, p := range progress {
		entry := progEntry{status: "in_progress", startedAt: p.StartedAt}
		if p.CompletedAt != "" {
			entry.status = "completed"
		}
		progressMap[p.ProjectID] = entry
	}

	result := make([]map[string]any, 0, len(projects))
	for _, p := range projects {
		prog, hasProgress := progressMap[p.ID]
		status := "available"
		startedAt := ""
		completesAt := ""
		if hasProgress {
			status = prog.status
			startedAt = prog.startedAt
			if prog.status == "in_progress" {
				start, parseErr := time.Parse(time.RFC3339, prog.startedAt)
				if parseErr == nil {
					completesAt = start.Add(time.Duration(p.DurationSeconds) * time.Second).Format(time.RFC3339)
				}
			}
		}
		item := map[string]any{
			"id":              p.ID,
			"name":            p.Name,
			"category":        p.Category,
			"cost":            p.Cost,
			"durationSeconds": p.DurationSeconds,
			"status":          status,
			"progress":        0,
			"prerequisites":   p.Prerequisites,
			"effects":         p.Effects,
			"startedAt":       startedAt,
			"completesAt":     completesAt,
		}
		if status == "completed" {
			item["progress"] = 100
		}
		result = append(result, item)
	}

	writeSuccess(w, http.StatusOK, map[string]any{
		"projects": result,
	})
}

// handleResearchProgress returns the company's in-progress research.
func (h *ResearchHandler) handleResearchProgress(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "company not authenticated", nil)
		return
	}
	progress, err := h.research.GetCompanyProgress(r.Context(), companyID)
	if err != nil {
		writeAppErr(w, apperr.WrapMsg(apperr.KindInternal, "failed to load research progress", err))
		return
	}

	projects := make([]map[string]any, 0, len(progress))
	for _, p := range progress {
		status := "in_progress"
		progPct := 0.0
		now := h.clock().UTC()
		start, parseErr := time.Parse(time.RFC3339, p.StartedAt)
		if p.CompletedAt != "" {
			status = "completed"
			progPct = 100.0
		} else if parseErr == nil {
			elapsed := now.Sub(start).Seconds()
			proj, _ := h.research.GetProjects(r.Context())
			for _, pr := range proj {
				if pr.ID == p.ProjectID && pr.DurationSeconds > 0 {
					progPct = (elapsed / pr.DurationSeconds) * 100.0
					if progPct > 100.0 {
						progPct = 100.0
					}
					break
				}
			}
		}
		projects = append(projects, map[string]any{
			"projectId":   p.ProjectID,
			"status":      status,
			"progress":    progPct,
			"startedAt":   p.StartedAt,
			"completedAt": p.CompletedAt,
		})
	}

	writeSuccess(w, http.StatusOK, map[string]any{
		"projects": projects,
	})
}

// handleStartResearch starts a new research project.
func (h *ResearchHandler) handleStartResearch(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "company not authenticated", nil)
		return
	}
	var req struct {
		ProjectID string `json:"projectId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ProjectID == "" {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "projectId is required", nil)
		return
	}

	// Verify project exists.
	projects, err := h.research.GetProjects(r.Context())
	if err != nil {
		writeAppErr(w, apperr.WrapMsg(apperr.KindInternal, "failed to load research projects", err))
		return
	}
	var foundProject *research.Project
	for _, p := range projects {
		if p.ID == req.ProjectID {
			foundProject = &p
			break
		}
	}
	if foundProject == nil {
		writeErr(w, http.StatusNotFound, ErrorNotFound, "research project not found", nil)
		return
	}

	// Check if already started.
	existing, _ := h.research.GetCompanyProgress(r.Context(), companyID)
	for _, p := range existing {
		if p.ProjectID == req.ProjectID && p.CompletedAt == "" {
			writeErr(w, http.StatusBadRequest, ErrorConflict, "research project already in progress", nil)
			return
		}
	}

	now := h.clock().UTC().Format(time.RFC3339)

	// Persist progress.
	progress := &research.CompanyProgress{
		CompanyID:   companyID,
		ProjectID:   req.ProjectID,
		StartedAt:   now,
		CompletedAt: "",
	}
	if err := h.research.SaveProgress(r.Context(), progress); err != nil {
		writeAppErr(w, apperr.WrapMsg(apperr.KindInternal, "failed to save research progress", err))
		return
	}

	completesAt := h.clock().UTC().Add(time.Duration(foundProject.DurationSeconds) * time.Second).Format(time.RFC3339)
	writeSuccess(w, http.StatusOK, map[string]any{
		"project": map[string]any{
			"id":              foundProject.ID,
			"name":            foundProject.Name,
			"status":          "in_progress",
			"progress":        0,
			"startedAt":       now,
			"completesAt":     completesAt,
			"durationSeconds": foundProject.DurationSeconds,
		},
		"status": "started",
	})
}

// handleCompleteResearch marks a research project as complete.
func (h *ResearchHandler) handleCompleteResearch(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "company not authenticated", nil)
		return
	}
	projectID := chi.URLParam(r, "projectId")
	if projectID == "" {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "projectId is required", nil)
		return
	}

	progress, err := h.research.GetCompanyProgress(r.Context(), companyID)
	if err != nil {
		writeAppErr(w, apperr.WrapMsg(apperr.KindInternal, "failed to load research progress", err))
		return
	}
	var found bool
	for _, p := range progress {
		if p.ProjectID == projectID && p.CompletedAt == "" {
			found = true
			break
		}
	}
	if !found {
		writeErr(w, http.StatusBadRequest, ErrorConflict, "research project not in progress or already completed", nil)
		return
	}

	now := h.clock().UTC().Format(time.RFC3339)
	writeSuccess(w, http.StatusOK, map[string]any{
		"ok":              true,
		"projectId":       projectID,
		"patentsGained":   1,
		"qualityImproved": 0,
		"completedAt":     now,
	})
}
