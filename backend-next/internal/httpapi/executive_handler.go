package httpapi

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
)

// ExecutiveHandler handles executive/HR endpoints.
type ExecutiveHandler struct {
	clock     func() time.Time
	recruited map[int][]map[string]any
}

// NewExecutiveHandler creates a new ExecutiveHandler.
func NewExecutiveHandler() *ExecutiveHandler {
	return &ExecutiveHandler{
		clock:     time.Now,
		recruited: make(map[int][]map[string]any),
	}
}

// executiveData represents a simplified executive for API responses.
type executiveData struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Title           string  `json:"title"`
	Level           int     `json:"level"`
	Rarity          string  `json:"rarity"`
	Stage           string  `json:"stage"`
	Salary          float64 `json:"salary"`
	ProductionBonus float64 `json:"productionBonus"`
	SalesBonus      float64 `json:"salesBonus"`
	MgmtDiscount    float64 `json:"mgmtDiscount"`
	Morale          int     `json:"morale,omitempty"`
}

var executiveNames = []string{
	"Alice Chen", "Bob Martinez", "Carol Wu", "David Park",
	"Elena Rossi", "Frank Okafor", "Grace Kim", "Henry Dubois",
	"Iris Johansson", "James Tanaka", "Kate Schmidt", "Leo Andersen",
}

var executiveTitles = []string{
	"Operations Manager", "Supply Chain Director", "Quality Assurance Lead",
	"Process Engineer", "Logistics Coordinator", "Production Supervisor",
	"Procurement Specialist", "Warehouse Manager", "Distribution Planner",
}

var rarityWeights = []struct {
	name   string
	weight int
}{
	{"Common", 50},
	{"Rare", 30},
	{"Epic", 15},
	{"Legendary", 5},
}

// handleSearchExecutives searches/refreshes the executive market.
func (h *ExecutiveHandler) handleSearchExecutives(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Scope string `json:"scope,omitempty"` // "mine" to list owned executives
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		// Empty body is fine - default to market search.
		req = struct {
			Scope string `json:"scope,omitempty"`
		}{}
	}

	if req.Scope == "mine" {
		companyID, ok := CompanyIDFromCtx(r.Context())
		if !ok {
			writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "company not authenticated", nil)
			return
		}
		list := h.recruited[companyID]
		if list == nil {
			list = []map[string]any{}
		}
		writeSuccess(w, http.StatusOK, map[string]any{
			"executives": list,
			"total":      len(list),
		})
		return
	}

	// Generate a set of market executives.
	count := 6
	execs := make([]map[string]any, 0, count)
	rng := h.clock().UnixNano()
	for i := 0; i < count; i++ {
		rng = (rng*1103515245 + 12345) & 0x7fffffff
		nameIdx := int(rng) % len(executiveNames)
		titleIdx := (int(rng) / len(executiveNames)) % len(executiveTitles)
		level := (int(rng) % 10) + 1

		rarity := pickRarity(int(rng / 100))
		salary := salaryAtLevel(level)
		stage := stageAtLevel(level)

		id := fmt.Sprintf("exec-%d-%d", i, h.clock().UnixMilli())
		execs = append(execs, map[string]any{
			"id":              id,
			"name":            executiveNames[nameIdx],
			"title":           executiveTitles[titleIdx],
			"level":           level,
			"rarity":          rarity,
			"stage":           stage,
			"salary":          salary,
			"productionBonus": productionBonusAtLevel(level),
			"salesBonus":      salesBonusAtLevel(level),
			"mgmtDiscount":    mgmtDiscountAtLevel(level),
		})
	}

	writeSuccess(w, http.StatusOK, map[string]any{
		"executives":      execs,
		"total":           count,
		"refreshCooldown": "09:00:00",
	})
}

// handleRecruitExecutive recruits an executive from the market.
func (h *ExecutiveHandler) handleRecruitExecutive(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ExecutiveID string `json:"executiveId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ExecutiveID == "" {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "executiveId is required", nil)
		return
	}

	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	exec := generateExecutiveData(req.ExecutiveID, h.clock().UnixMilli())
	h.recruited[companyID] = append(h.recruited[companyID], exec)

	writeSuccess(w, http.StatusOK, map[string]any{
		"ok":        true,
		"executive": exec,
	})
}

// handleTrainExecutive levels up an executive.
func (h *ExecutiveHandler) handleTrainExecutive(w http.ResponseWriter, r *http.Request) {
	executiveID := chi.URLParam(r, "executiveId")
	if executiveID == "" {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "executiveId is required", nil)
		return
	}

	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	execs := h.recruited[companyID]
	var exec map[string]any
	for _, e := range execs {
		if e["id"] == executiveID {
			exec = e
			break
		}
	}
	if exec == nil {
		writeErr(w, http.StatusNotFound, ErrorNotFound, "executive not found", nil)
		return
	}

	// Increment level and recalculate bonuses.
	level := int(exec["level"].(float64)) + 1
	exec["level"] = level
	exec["stage"] = stageAtLevel(level)
	exec["salary"] = salaryAtLevel(level)
	exec["productionBonus"] = productionBonusAtLevel(level)
	exec["salesBonus"] = salesBonusAtLevel(level)
	exec["mgmtDiscount"] = mgmtDiscountAtLevel(level)

	writeSuccess(w, http.StatusOK, map[string]any{
		"ok":        true,
		"executive": exec,
	})
}

// handleGetExecutiveDetail returns details for a single executive.
func (h *ExecutiveHandler) handleGetExecutiveDetail(w http.ResponseWriter, r *http.Request) {
	executiveID := chi.URLParam(r, "id")
	if executiveID == "" {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "executive id is required", nil)
		return
	}

	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	execs := h.recruited[companyID]
	for _, exec := range execs {
		if exec["id"] == executiveID {
			writeSuccess(w, http.StatusOK, exec)
			return
		}
	}

	writeErr(w, http.StatusNotFound, ErrorNotFound, "executive not found", nil)
}

// --- Helper functions matching frontend executive formulas ---

func productionBonusAtLevel(level int) float64 {
	// Lv1 base: 2%, increment decays by 0.12 per level.
	pct := 2.0
	inc := 2.0
	for i := 1; i < level; i++ {
		inc -= 0.12
		if inc < 0 {
			inc = 0
		}
		pct += inc
	}
	return pct
}

func salesBonusAtLevel(level int) float64 {
	// Lv1 base: 4%, decays slower.
	pct := 4.0
	inc := 4.0
	for i := 1; i < level; i++ {
		inc -= 0.2
		if inc < 0 {
			inc = 0
		}
		pct += inc
	}
	return pct
}

func mgmtDiscountAtLevel(level int) float64 {
	// Lv1 base: 2%, decays moderately.
	pct := 2.0
	inc := 2.0
	for i := 1; i < level; i++ {
		inc -= 0.15
		if inc < 0 {
			inc = 0
		}
		pct += inc
	}
	return pct
}

func salaryAtLevel(level int) float64 {
	return 600 + 80*math.Pow(float64(level), 1.3)
}

func stageAtLevel(level int) string {
	if level >= 10 {
		return "Executive VP"
	}
	if level >= 8 {
		return "Director"
	}
	if level >= 6 {
		return "Senior Manager"
	}
	if level >= 4 {
		return "Manager"
	}
	return "Associate"
}

func pickRarity(seed int) string {
	total := 0
	for _, r := range rarityWeights {
		total += r.weight
	}
	roll := seed % total
	cum := 0
	for _, r := range rarityWeights {
		cum += r.weight
		if roll < cum {
			return r.name
		}
	}
	return "Common"
}

// generateExecutiveData creates an executive map from an ID and seed.
func generateExecutiveData(id string, seed int64) map[string]any {
	rng := seed
	rng = (rng*1103515245 + 12345) & 0x7fffffff
	nameIdx := int(rng) % len(executiveNames)
	titleIdx := (int(rng) / len(executiveNames)) % len(executiveTitles)
	level := (int(rng) % 10) + 1

	rarity := pickRarity(int(rng / 100))
	salary := salaryAtLevel(level)
	stage := stageAtLevel(level)

	return map[string]any{
		"id":              id,
		"name":            executiveNames[nameIdx],
		"title":           executiveTitles[titleIdx],
		"level":           level,
		"rarity":          rarity,
		"stage":           stage,
		"salary":          salary,
		"productionBonus": productionBonusAtLevel(level),
		"salesBonus":      salesBonusAtLevel(level),
		"mgmtDiscount":    mgmtDiscountAtLevel(level),
	}
}
