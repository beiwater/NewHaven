package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/newhaven/backend-next/internal/apperr"
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

// featureUnlockLevels defines which features unlock at which player level.
var featureUnlockLevels = map[string]int{
	"map":         1,
	"build":       1,
	"warehouse":   1,
	"leaderboard": 1,
	"market":      2,
	"contracts":   3,
	"research":    4,
	"executives":  5,
	"finance":     6,
}

// buildingUnlockLevels defines which building IDs unlock at which player level.
var buildingUnlockLevels = map[int]int{
	1:  1,  // Farm
	2:  2,  // Barn
	3:  2,  // Mill
	4:  3,  // Kitchen
	5:  4,  // Bakery
	6:  5,  // Market Stall
	7:  6,  // Cafe
	8:  7,  // Food Truck
	9:  8,  // Restaurant
	10: 9,  // Trading Hub
	11: 10, // Warehouse
	12: 11, // Shop
}

func xpToNextLevel(level int) int {
	if level < 1 {
		return 100
	}
	return level * 100
}

func buildingSlotsForLevel(level int) int {
	return 2 + level/2
}

// featureUnlocksAtLevel returns which features are unlocked and their unlock levels.
func featureUnlocksAtLevel(level int) map[string]any {
	features := make(map[string]bool)
	featureLevels := make(map[string]int)
	for f, lvl := range featureUnlockLevels {
		featureLevels[f] = lvl
		features[f] = level >= lvl
	}
	return map[string]any{
		"features":      features,
		"featureLevels": featureLevels,
	}
}

// buildingUnlocksAtLevel returns which building IDs are unlocked at this level.
func buildingUnlocksAtLevel(level int) []int {
	var ids []int
	for bid, lvl := range buildingUnlockLevels {
		if lvl <= level {
			ids = append(ids, bid)
		}
	}
	return ids
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
		"xpToNextLevel": xpToNextLevel(c.Level),
		"buildingSlots": buildingSlotsForLevel(c.Level),
		"buildingsUsed": len(c.Buildings),
		"unlocks":       featureUnlocksAtLevel(c.Level),
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
			"endsAt":     "2026-06-07T00:00:00Z",
			"multiplier": 2.0,
		})
	default:
		writeErr(w, http.StatusMethodNotAllowed, ErrorBadRequest, "method not allowed", nil)
	}
}
