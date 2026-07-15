package httpapi

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/beiwater/NewHaven/backend/internal/domain/company"
	"github.com/beiwater/NewHaven/backend/internal/storage"
)

// LeaderboardHandler handles leaderboard endpoints.
type LeaderboardHandler struct {
	st storage.CompanyStorage
}

// NewLeaderboardHandler creates a new LeaderboardHandler.
func NewLeaderboardHandler(st storage.CompanyStorage) *LeaderboardHandler {
	return &LeaderboardHandler{st: st}
}

// handleLeaderboard returns the company leaderboard.
func (h *LeaderboardHandler) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	sortKey := r.URL.Query().Get("sort")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	if sortKey == "" {
		sortKey = "net_worth"
	}
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 || limit > 100 {
		limit = 10
	}

	companies, err := h.st.GetAllCompanies(r.Context())
	if err != nil {
		writeErr(w, 500, ErrorInternal, "failed to load companies", nil)
		return
	}

	sortCompanies(companies, sortKey)

	// Paginate
	start := (page - 1) * limit
	if start > len(companies) {
		start = len(companies)
	}
	end := start + limit
	if end > len(companies) {
		end = len(companies)
	}

	entries := make([]map[string]any, 0, end-start)
	for i, c := range companies[start:end] {
		entries = append(entries, map[string]any{
			"rank":        start + i + 1,
			"companyId":   c.ID,
			"companyName": c.Name,
			"level":       c.Level,
			"mainStat":    mainStat(c, sortKey),
		})
	}

	total := len(companies)
	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}

	writeSuccess(w, http.StatusOK, map[string]any{
		"entries":    entries,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": totalPages,
		"sort":       sortKey,
	})
}

func mainStat(c *company.Company, sortKey string) int64 {
	switch sortKey {
	case "net_worth":
		return int64(c.Money) + c.XP
	case "level":
		return int64(c.Level)
	case "production":
		total := int64(0)
		for _, b := range c.Buildings {
			total += int64(b.Level)
		}
		return total
	case "sales", "contracts":
		// Not tracked yet; fall back to wealth.
		return int64(c.Money) + c.XP
	default:
		return int64(c.Money) + c.XP
	}
}

func sortCompanies(companies []*company.Company, sortKey string) {
	sort.Slice(companies, func(i, j int) bool {
		a, b := companies[i], companies[j]
		switch sortKey {
		case "net_worth":
			return a.Money+float64(a.XP) > b.Money+float64(b.XP)
		case "level":
			if a.Level != b.Level {
				return a.Level > b.Level
			}
			return a.XP > b.XP
		default:
			return a.Money+float64(a.XP) > b.Money+float64(b.XP)
		}
	})
}
