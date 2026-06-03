package handler

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) RegisterAerospace(mux *http.ServeMux) {
	mux.HandleFunc("/api/v2/aerospace/projects/", h.withAuth(h.handleAerospaceProjects))
	mux.HandleFunc("/api/v2/aerospace/projects/create/", h.withAuth(h.handleAerospaceCreateProject))
	mux.HandleFunc("/api/v2/aerospace/launches/", h.withAuth(h.handleAerospaceLaunches))
	mux.HandleFunc("/api/v2/aerospace/launch/", h.withAuth(h.handleAerospaceLaunch))
	mux.HandleFunc("/api/v2/aerospace/components/", h.withAuth(h.handleAerospaceComponents))
}

func (h *Handler) handleAerospaceProjects(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]any{"projects": h.svc.RocketProjects()})
}

func (h *Handler) handleAerospaceCreateProject(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "method not allowed")
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "invalid json")
		return
	}
	writeJSON(w, 200, h.svc.CreateRocketProject(body))
}

func (h *Handler) handleAerospaceLaunches(w http.ResponseWriter, _ *http.Request) {
	launches := h.svc.LaunchHistory()
	writeJSON(w, 200, map[string]any{"launches": launches, "total": len(launches)})
}

func (h *Handler) handleAerospaceLaunch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "method not allowed")
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "invalid json")
		return
	}
	writeJSON(w, 200, h.svc.LaunchRocket(body))
}

func (h *Handler) handleAerospaceComponents(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]any{"components": h.svc.AvailableComponents()})
}
