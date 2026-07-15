package httpapi

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/beiwater/NewHaven/backend/internal/platform"
)

// sanitizePlainText strips HTML tags and trims whitespace.
// This prevents stored XSS if report descriptions are ever rendered in a UI.
func sanitizePlainText(s string) string {
	// Simple but effective: strip anything that looks like an HTML tag.
	var buf strings.Builder
	buf.Grow(len(s))
	inTag := false
	for _, r := range s {
		switch {
		case r == '<':
			inTag = true
		case r == '>':
			inTag = false
		case !inTag:
			buf.WriteRune(r)
		}
	}
	return strings.TrimSpace(buf.String())
}

// --- Types ---

type reportRequest struct {
	Category    string `json:"category"`
	Description string `json:"description"`
}

type reportRecord struct {
	ID          string `json:"id"`
	PlayerID    int    `json:"playerId"`
	CompanyID   int    `json:"companyId"`
	Category    string `json:"category"`
	Description string `json:"description"`
	At          string `json:"at"`
}

// ReportHandler handles bug report submissions.
type ReportHandler struct {
	logDir string
	idgen  *platform.IDGen
	clock  platform.Clock
}

// NewReportHandler creates a ReportHandler that writes reports to logDir.
func NewReportHandler(logDir string, idgen *platform.IDGen, clock platform.Clock) *ReportHandler {
	return &ReportHandler{logDir: logDir, idgen: idgen, clock: clock}
}

// handleSubmitReport receives a bug/feature report and persists it as JSON.
func (h *ReportHandler) handleSubmitReport(w http.ResponseWriter, r *http.Request) {
	var req reportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, ErrorBadRequest, "invalid request body", nil)
		return
	}

	// Validate category
	switch req.Category {
	case "bug", "feature", "feedback", "other":
		// valid
	default:
		writeErr(w, http.StatusBadRequest, ErrorValidation, "category must be one of: bug, feature, feedback, other", nil)
		return
	}

	if req.Description == "" {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "description is required", nil)
		return
	}

	if len(req.Description) > 2000 {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "description must be 2000 characters or fewer", nil)
		return
	}

	// Sanitize: strip HTML tags to prevent stored XSS
	req.Description = sanitizePlainText(req.Description)

	if req.Description == "" {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "description cannot be empty after filtering", nil)
		return
	}
	playerID, _ := PlayerIDFromCtx(r.Context())
	companyID, _ := CompanyIDFromCtx(r.Context())

	now := h.clock.Now()
	record := reportRecord{
		ID:          h.idgen.Next("report"),
		PlayerID:    playerID,
		CompanyID:   companyID,
		Category:    req.Category,
		Description: req.Description,
		At:          now.Format(time.RFC3339),
	}

	// Ensure log directory exists
	if err := os.MkdirAll(h.logDir, 0755); err != nil {
		writeErr(w, http.StatusInternalServerError, ErrorInternal, "failed to persist report", nil)
		return
	}

	filename := fmt.Sprintf("report-%s.json", record.ID)
	path := filepath.Join(h.logDir, filename)

	data, err := json.MarshalIndent(record, "", "  ")
	if err != nil {
		writeErr(w, http.StatusInternalServerError, ErrorInternal, "failed to serialize report", nil)
		return
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		writeErr(w, http.StatusInternalServerError, ErrorInternal, "failed to write report", nil)
		return
	}

	writeSuccess(w, http.StatusCreated, map[string]string{
		"id":     record.ID,
		"status": "submitted",
	})
}
