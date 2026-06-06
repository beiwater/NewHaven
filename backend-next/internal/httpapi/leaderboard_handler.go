package httpapi

import (
	"net/http"
	"strconv"
)

// LeaderboardHandler handles leaderboard endpoints.
type LeaderboardHandler struct {
}

// NewLeaderboardHandler creates a new LeaderboardHandler.
func NewLeaderboardHandler() *LeaderboardHandler {
	return &LeaderboardHandler{}
}

// handleLeaderboard returns the company leaderboard.
func (h *LeaderboardHandler) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	sort := r.URL.Query().Get("sort")
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	if sort == "" {
		sort = "net_worth"
	}
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	}
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 || limit > 100 {
		limit = 10
	}

	entries := []map[string]any{
		{
			"rank":        1,
			"companyId":   1000002,
			"companyName": "Dev Corp",
			"level":       100,
			"mainStat":    1000000000,
		},
		{
			"rank":        2,
			"companyId":   1000003,
			"companyName": "Alpha Industries",
			"level":       45,
			"mainStat":    45000000,
		},
		{
			"rank":        3,
			"companyId":   1000004,
			"companyName": "Beta Manufacturing",
			"level":       38,
			"mainStat":    32000000,
		},
		{
			"rank":        4,
			"companyId":   1000005,
			"companyName": "Gamma Trading Co",
			"level":       32,
			"mainStat":    28000000,
		},
		{
			"rank":        5,
			"companyId":   1000006,
			"companyName": "Delta Logistics",
			"level":       28,
			"mainStat":    22000000,
		},
	}
	total := 5
	totalPages := 1

	writeSuccess(w, http.StatusOK, map[string]any{
		"entries":    entries,
		"total":      total,
		"page":       page,
		"limit":      limit,
		"totalPages": totalPages,
		"sort":       sort,
	})
}
