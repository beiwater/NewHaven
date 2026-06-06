package httpapi

import (
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/newhaven/backend-next/internal/domain/company"
	"github.com/newhaven/backend-next/internal/storage"
)

// ExecutiveHandler handles executive/HR endpoints.
type ExecutiveHandler struct {
	companies storage.CompanyStorage
	clock     func() time.Time
}

// NewExecutiveHandler creates a new ExecutiveHandler.
func NewExecutiveHandler(companies storage.CompanyStorage) *ExecutiveHandler {
	return &ExecutiveHandler{
		companies: companies,
		clock:     time.Now,
	}
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
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "company not authenticated", nil)
		return
	}

	var req struct {
		Scope string `json:"scope,omitempty"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		req = struct{ Scope string `json:"scope,omitempty"` }{}
	}
	if req.Scope == "" {
		req.Scope = "market"
	}

	if req.Scope == "mine" {
		c, err := h.companies.GetCompany(r.Context(), companyID)
		if err != nil {
			writeAppErr(w, err)
			return
		}
		writeSuccess(w, http.StatusOK, map[string]any{
			"executives": c.Executives,
			"total":      len(c.Executives),
		})
		return
	}

	// Generate market executives (consistent for the current hour).
	count := 6
	now := h.clock()
	seed := now.Unix() / 3600
	rng := rand.New(rand.NewSource(seed))

	execs := make([]map[string]any, 0, count)
	for i := 0; i < count; i++ {
		nameIdx := rng.Intn(len(executiveNames))
		titleIdx := rng.Intn(len(executiveTitles))
		level := rng.Intn(10) + 1
		rarity := pickRarity(rng.Intn(100))
		id := fmt.Sprintf("exec-market-%d-%d", i, now.UnixMilli())
		execs = append(execs, map[string]any{
			"id":              id,
			"name":            executiveNames[nameIdx],
			"title":           executiveTitles[titleIdx],
			"level":           level,
			"rarity":          rarity,
			"stage":           stageAtLevel(level),
			"salary":          salaryAtLevel(level),
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
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "company not authenticated", nil)
		return
	}
	var req struct {
		ExecutiveID string `json:"executiveId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.ExecutiveID == "" {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "executiveId is required", nil)
		return
	}

	c, err := h.companies.GetCompany(r.Context(), companyID)
	if err != nil {
		writeAppErr(w, err)
		return
	}

	// Build the executive from the market data (reconstruct via ID seed).
	now := h.clock()
	seed := now.Unix() / 3600
	rng := rand.New(rand.NewSource(seed))
	nameIdx := rng.Intn(len(executiveNames))
	titleIdx := rng.Intn(len(executiveTitles))
	level := rng.Intn(10) + 1
	rarity := pickRarity(rng.Intn(100))

	exec := company.Executive{
		ID:              req.ExecutiveID,
		Name:            executiveNames[nameIdx],
		Title:           executiveTitles[titleIdx],
		Level:           level,
		Rarity:          rarity,
		Stage:           stageAtLevel(level),
		Salary:          salaryAtLevel(level),
		ProductionBonus: productionBonusAtLevel(level),
		SalesBonus:      salesBonusAtLevel(level),
		MgmtDiscount:    mgmtDiscountAtLevel(level),
		Morale:          100,
	}

	c.Executives = append(c.Executives, exec)
	if err := h.companies.UpdateCompany(r.Context(), c); err != nil {
		writeAppErr(w, err)
		return
	}

	writeSuccess(w, http.StatusOK, map[string]any{
		"ok":        true,
		"executive": exec,
	})
}

// handleTrainExecutive levels up an executive.
func (h *ExecutiveHandler) handleTrainExecutive(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "company not authenticated", nil)
		return
	}
	executiveID := chi.URLParam(r, "executiveId")
	if executiveID == "" {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "executiveId is required", nil)
		return
	}

	c, err := h.companies.GetCompany(r.Context(), companyID)
	if err != nil {
		writeAppErr(w, err)
		return
	}

	for i := range c.Executives {
		if c.Executives[i].ID == executiveID {
			c.Executives[i].Level++
			newLvl := c.Executives[i].Level
			c.Executives[i].Stage = stageAtLevel(newLvl)
			c.Executives[i].Salary = salaryAtLevel(newLvl)
			c.Executives[i].ProductionBonus = productionBonusAtLevel(newLvl)
			c.Executives[i].SalesBonus = salesBonusAtLevel(newLvl)
			c.Executives[i].MgmtDiscount = mgmtDiscountAtLevel(newLvl)

			if err := h.companies.UpdateCompany(r.Context(), c); err != nil {
				writeAppErr(w, err)
				return
			}
			writeSuccess(w, http.StatusOK, map[string]any{
				"ok":        true,
				"executive": c.Executives[i],
			})
			return
		}
	}
	writeErr(w, http.StatusNotFound, ErrorNotFound, "executive not found", nil)
}

// handleGetExecutiveDetail returns details for a single executive.
func (h *ExecutiveHandler) handleGetExecutiveDetail(w http.ResponseWriter, r *http.Request) {
	companyID, ok := CompanyIDFromCtx(r.Context())
	if !ok {
		writeErr(w, http.StatusUnauthorized, ErrorUnauthorized, "company not authenticated", nil)
		return
	}
	executiveID := chi.URLParam(r, "id")
	if executiveID == "" {
		writeErr(w, http.StatusBadRequest, ErrorValidation, "executive id is required", nil)
		return
	}

	c, err := h.companies.GetCompany(r.Context(), companyID)
	if err != nil {
		writeAppErr(w, err)
		return
	}

	for _, exec := range c.Executives {
		if exec.ID == executiveID {
			writeSuccess(w, http.StatusOK, exec)
			return
		}
	}
	writeErr(w, http.StatusNotFound, ErrorNotFound, "executive not found", nil)
}

// --- Helper functions matching frontend executive formulas ---

func productionBonusAtLevel(level int) float64 {
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
	switch {
	case level >= 10:
		return "Executive VP"
	case level >= 8:
		return "Director"
	case level >= 6:
		return "Senior Manager"
	case level >= 4:
		return "Manager"
	default:
		return "Associate"
	}
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
