package httpapi

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/newhaven/backend-next/internal/apperr"
	"github.com/newhaven/backend-next/internal/formula"
	"github.com/newhaven/backend-next/internal/storage"
)

// PlayerHandler handles player-related HTTP endpoints.
type PlayerHandler struct {
	companies storage.CompanyStorage
}

// NewPlayerHandler creates a new PlayerHandler.
func NewPlayerHandler(companies storage.CompanyStorage) *PlayerHandler {
	return &PlayerHandler{companies: companies}
}

// handleLevel returns the authenticated player's level info.
func (h *PlayerHandler) handleLevel(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "company not authenticated", nil)
		return
	}
	c, err := h.companies.GetCompany(r.Context(), companyID)
	if err != nil {
		writeAppErr(w, apperr.WrapMsg(apperr.KindNotFound, "company not found", err))
		return
	}
	writeSuccess(w, http.StatusOK, map[string]any{
		"level":         c.Level,
		"currentXp":     c.XP,
		"xpToNextLevel": formula.XpToNextLevel(c.Level),
		"buildingSlots": formula.BuildingSlotsForLevel(c.Level),
		"buildingsUsed": len(c.Buildings),
		"unlocks":       formula.FeatureUnlocksAtLevel(c.Level),
	})
}

// handleSimboostTypes returns available simboost/powerup types.
func (h *PlayerHandler) handleSimboostTypes(w http.ResponseWriter, _ *http.Request) {
	writeSuccess(w, http.StatusOK, map[string]any{
		"boosts": []map[string]any{
			{"id": "speed", "name": "Speed Boost", "desc": "Faster production for 1 hour", "duration": 3600},
			{"id": "yield", "name": "Yield Boost", "desc": "+50% production output for 1 hour", "duration": 3600},
			{"id": "quality", "name": "Quality Boost", "desc": "Improve product quality for 1 hour", "duration": 3600},
		},
	})
}

// handleSimboostsUse returns active boosts and remaining uses (GET) or activates a boost (POST).
func (h *PlayerHandler) handleSimboostsUse(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeSuccess(w, http.StatusOK, map[string]any{
			"remaining": 3,
			"active":    []map[string]any{},
		})
	case http.MethodPost:
		var req struct {
			BoostID string `json:"boostId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.BoostID == "" {
			writeErr(w, http.StatusBadRequest, ErrorValidation, "boostId is required", nil)
			return
		}
		// Stub: accept any boost, return a simulated response
		writeSuccess(w, http.StatusOK, map[string]any{
			"boostId":    req.BoostID,
			"endsAt":     time.Now().Add(1 * time.Hour).Format(time.RFC3339),
			"multiplier": 1.5,
		})
	default:
		writeErr(w, http.StatusMethodNotAllowed, ErrorBadRequest, "method not allowed", nil)
	}
}
