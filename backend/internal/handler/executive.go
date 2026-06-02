package handler

import (
	"encoding/json"
	"net/http"
	"strings"
)

func (h *Handler) handleExecSearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "method not allowed")
		return
	}
	var body map[string]any
	scope := "market"
	if err := json.NewDecoder(r.Body).Decode(&body); err == nil {
		if s, ok := body["scope"].(string); ok {
			scope = s
		}
	}
	var execs []map[string]any
	switch scope {
	case "mine":
		execs = h.svc.MyExecutives()
	default:
		execs = h.svc.ExecutiveCatalog()
	}
	writeJSON(w, 200, map[string]any{"executives": execs, "total": len(execs)})
}

func (h *Handler) handleExecRecruit(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "method not allowed")
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "invalid json")
		return
	}
	execID, _ := body["executiveId"].(string)
	if execID == "" {
		execID = "exec-1"
	}
	writeJSON(w, 200, h.svc.RecruitExecutive(execID))
}

func (h *Handler) handleExecTrain(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "method not allowed")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v2/executives/train/")
	id = strings.Trim(id, "/")
	if id == "" {
		writeErr(w, 400, "missing executive id")
		return
	}
	writeJSON(w, 200, h.svc.TrainExecutive(id))
}

func (h *Handler) handleExecPoach(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "method not allowed")
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "invalid json")
		return
	}
	writeJSON(w, 200, h.svc.PoachExecutive(body))
}

func (h *Handler) handleExecOffers(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		writeJSON(w, 200, map[string]any{"offers": h.svc.IncomingOffers()})
	case http.MethodPost:
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, 400, "invalid json")
			return
		}
		writeJSON(w, 200, h.svc.RespondToOffer(body))
	default:
		writeErr(w, 405, "method not allowed")
	}
}

func (h *Handler) handleV3ExecutiveByID(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v3/executives/")
	id = strings.Trim(id, "/")
	if id == "" {
		writeErr(w, 404, "not found")
		return
	}
	if r.Method != http.MethodGet {
		writeErr(w, 405, "method not allowed")
		return
	}
	detail, err := h.svc.ExecutiveDetail(id)
	if err != nil {
		writeErr(w, 404, err.Error())
		return
	}
	writeJSON(w, 200, detail)
}
