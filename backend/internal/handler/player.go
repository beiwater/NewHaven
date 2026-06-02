package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"go-sim-api/internal/service"
)

func (h *Handler) RegisterPlayer(mux *http.ServeMux) {
	mux.HandleFunc("/api/v2/companies/me/achievements/", h.withAuth(h.handleAchievements))
	mux.HandleFunc("/api/v2/no-cache/companies/me/achievements/", h.withAuth(h.handleNoCacheAchievements))
	mux.HandleFunc("/api/v2/no-cache/companies/achievements/", h.withAuth(h.handleDeleteAchievements))
	// SimBoost
	mux.HandleFunc("/api/v2/players/simboosts-use/", h.withAuth(h.handleSimboostsUse))
	mux.HandleFunc("/api/v2/players/simboosts/", h.withAuth(h.handleSimboostTypes))
	mux.HandleFunc("/api/v2/players/unlocked-hqs/", h.withAuth(h.handleUnlockedHQs))
	mux.HandleFunc("/api/v2/players/devices/", h.withAuth(h.handlePlayerDevices))
	// Level / XP
	mux.HandleFunc("/api/v2/players/me/level/", h.withAuth(h.handleLevel))
	mux.HandleFunc("/api/v2/players/me/xp/", h.withAuth(h.handlePlayerAddXP))
	mux.HandleFunc("/api/v2/players/me/level-rewards/", h.withAuth(h.handlePlayerLevelRewards))
	mux.HandleFunc("/api/v2/players/me/offline-income/", h.withAuth(h.handleOfflineIncome))
}

func (h *Handler) handleAchievements(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, h.svc.Snapshot().Achievements)
}

func (h *Handler) handleNoCacheAchievements(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, h.svc.Snapshot().Achievements)
}

func (h *Handler) handleDeleteAchievements(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodDelete {
		writeJSON(w, 200, map[string]any{"ok": true})
		return
	}
	writeErr(w, 405, "method not allowed")
}

func (h *Handler) handleSimboostsUse(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, 200, h.svc.SimBoostsUse())
	case http.MethodPost:
		var req struct {
			BoostID string `json:"boostId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.BoostID == "" {
			writeErr(w, 400, "boostId required")
			return
		}
		resp, err := h.svc.UseSimBoost(req.BoostID)
		if err != nil {
			writeErr(w, 400, err.Error())
			return
		}
		writeJSON(w, 200, resp)
	default:
		writeErr(w, 405, "method not allowed")
	}
}

func (h *Handler) handleUnlockedHQs(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, []string{"classic", "city", "industrial"})
}

func (h *Handler) handlePlayerDevices(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, []any{})
}

func (h *Handler) handleSimboostTypes(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, map[string]any{"boosts": h.svc.SimBoostTypes()})
}

func (h *Handler) handleLevel(w http.ResponseWriter, r *http.Request) {
	st := h.svc.Snapshot()
	c := st.GetCompany(h.companyID(r))
	writeJSON(w, 200, map[string]any{
		"level": c.Level, "currentXp": st.XP, "xpToNextLevel": st.XpToNextLevel,
		"buildingSlots": service.BuildingSlotsForLevel(c.Level), "buildingsUsed": len(c.PlacedBuildings) + len(c.UnplacedBuildings),
		"unlocks": service.FeatureUnlockPayload(c.Level),
	})
}

func (h *Handler) handlePlayerAddXP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "method not allowed")
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "invalid json")
		return
	}
	amount := intFromAny(body["xp"])
	if amount <= 0 {
		amount = 100
	}
	writeJSON(w, 200, h.svc.AddXP(amount))
}

func (h *Handler) handlePlayerLevelRewards(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v2/players/me/level-rewards/")
	path = strings.Trim(path, "/")
	level, err := strconv.Atoi(path)
	if err != nil || level < 0 || level > 60 {
		writeErr(w, 400, "invalid level")
		return
	}
	writeJSON(w, 200, h.svc.LevelRewards(level))
}

func (h *Handler) handleOfflineIncome(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, 200, h.svc.CalculateOfflineIncome(h.companyID(r)))
}
