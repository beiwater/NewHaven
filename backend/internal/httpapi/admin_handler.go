package httpapi

import (
	"net/http"

	"github.com/beiwater/NewHaven/backend/internal/apperr"
	"github.com/beiwater/NewHaven/backend/internal/storage"
)

// AdminHandler handles admin/dev endpoints.
type AdminHandler struct {
	st storage.Storage
}

// NewAdminHandler creates a new AdminHandler.
func NewAdminHandler(st storage.Storage) *AdminHandler {
	return &AdminHandler{st: st}
}

// handleSaveSnapshot persists the current game state to the snapshot file (or PG if configured).
func (h *AdminHandler) handleSaveSnapshot(w http.ResponseWriter, r *http.Request) {
	if err := h.st.SaveSnapshot(r.Context()); err != nil {
		writeAppErr(w, apperr.WrapMsg(apperr.KindInternal, "save snapshot failed", err))
		return
	}
	writeSuccess(w, http.StatusOK, map[string]string{"status": "saved"})
}

// handleLoadSnapshot reloads game state from the snapshot file (or PG if configured).
func (h *AdminHandler) handleLoadSnapshot(w http.ResponseWriter, r *http.Request) {
	if err := h.st.LoadSnapshot(r.Context()); err != nil {
		writeAppErr(w, apperr.WrapMsg(apperr.KindInternal, "load snapshot failed", err))
		return
	}
	writeSuccess(w, http.StatusOK, map[string]string{"status": "loaded"})
}
