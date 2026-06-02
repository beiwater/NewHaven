package handler

import (
	"net/http"
	"strconv"
)

func (h *Handler) RegisterLeaderboard(mux *http.ServeMux) {
	mux.HandleFunc("/api/v2/leaderboard/", h.withAuth(h.handleLeaderboard))
}

func (h *Handler) handleLeaderboard(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, 405, "method not allowed")
		return
	}

	q := r.URL.Query()
	sortBy := q.Get("sort")
	if sortBy == "" {
		sortBy = "net_worth"
	}

	page, _ := strconv.Atoi(q.Get("page"))
	if page < 1 {
		page = 1
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	result := h.svc.Leaderboard(sortBy, page, limit)
	writeJSON(w, 200, result)
}
