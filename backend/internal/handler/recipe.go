package handler

import (
	"net/http"
	"strconv"
	"strings"
)

func (h *Handler) RegisterRecipe(mux *http.ServeMux) {
	mux.HandleFunc("/api/v2/recipes/", h.withAuth(h.handleRecipes))
}

func (h *Handler) handleRecipes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v2/recipes/")
	path = strings.Trim(path, "/")
	if path != "" {
		id, err := strconv.Atoi(path)
		if err != nil {
			writeErr(w, 400, "invalid resource id")
			return
		}
		writeJSON(w, 200, h.svc.RecipeDetail(id))
		return
	}
	writeJSON(w, 200, h.svc.RecipeList())
}
