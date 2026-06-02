package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go-sim-api/internal/formula"
)

func (h *Handler) RegisterDev(mux *http.ServeMux) {
	mux.HandleFunc("/api/dev/ledger/", h.withAuth(h.handleDevLedger))
	mux.HandleFunc("/api/dev/formulas/production/", h.withAuth(h.handleDevFormulasProduction))
	mux.HandleFunc("/api/dev/formulas/retail/", h.withAuth(h.handleDevFormulasRetail))
	mux.HandleFunc("/api/dev/formulas/retail-season-weather/", h.withAuth(h.handleDevFormulasRetailSeasonWeather))
	mux.HandleFunc("/api/v4/", h.withAuth(h.handleV4))
	mux.HandleFunc("/api/v3/contracts-incoming/", h.withAuth(h.handleContractsIncoming))
	mux.HandleFunc("/api/v3/contracts-outgoing/me/", h.withAuth(h.handleContractsOutgoing))
	mux.HandleFunc("/api/v2/contracts-history-incoming/", h.withAuth(h.handleContractsHistoryIncoming))
	mux.HandleFunc("/api/v2/contracts-history-outgoing/", h.withAuth(h.handleContractsHistoryOutgoing))
	mux.HandleFunc("/api/v2/warehouse-contracts-summary/", h.withAuth(h.handleWarehouseContractsSummary))
	// Research system
	mux.HandleFunc("/api/v2/research/", h.withAuth(h.handleResearch))
	mux.HandleFunc("/api/v2/research/start/", h.withAuth(h.handleStartResearch))
	mux.HandleFunc("/api/v2/research/progress/", h.withAuth(h.handleResearchProgress))
	mux.HandleFunc("/api/dev/time/", h.withAuth(h.handleDevTime))
	mux.HandleFunc("/api/v2/research/complete/", h.withAuth(h.handleCompleteResearch))
}

func (h *Handler) handleDevLedger(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]any{"entries": h.svc.Snapshot().Ledger})
}

func (h *Handler) handleDevFormulasProduction(w http.ResponseWriter, _ *http.Request) {
	baseOutput := 500.0                          // Farm Lv1 base output
	speedBonus := 10.0                            // +10% speed
	level := 3
	ph := formula.OutputPerHour(baseOutput, speedBonus, level)
	dur := formula.ProductionDurationSeconds(100, 3600/baseOutput, level, 1.0)
	writeJSON(w, 200, map[string]any{
		"baseOutputPerHour": baseOutput,
		"speedBonusPct":     speedBonus,
		"level":             level,
		"outputPerHour":     ph,
		"durationSec":       dur,
	})
}

func (h *Handler) handleDevFormulasRetail(w http.ResponseWriter, _ *http.Request) {
	u := formula.UnitsSoldPerHour(1.0, 0.02, 10, 151.8, 100, 13.5, 4, 1.1, 0, 1, 1, 1.06)
	writeJSON(w, 200, map[string]any{"unitsPerHour": u})
}

func (h *Handler) handleDevFormulasRetailSeasonWeather(w http.ResponseWriter, _ *http.Request) {
	month := int(time.Now().UTC().Month())
	saturation := 1.0
	if seasons, ok := h.svc.Data.ResourceLookups["seasons"].(map[string]any); ok && len(seasons) > 0 {
		saturation = 1.05
	}
	switch month {
	case 12:
		saturation = 1.25
	case 8:
		saturation = 1.15
	}
	weather := 1.06
	u := formula.UnitsSoldPerHour(1.0, 0.02, 10, 151.8, 100, 13.5, 4, saturation, 0, 1, 1, weather)
	writeJSON(w, 200, map[string]any{
		"month": month, "seasonalSaturation": saturation, "weatherMultiplier": weather, "unitsPerHour": u,
	})
}

func (h *Handler) handleV4(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, "/resources-retail-info/") {
		// Return real retail info for all resources with economy model
		items := make([]map[string]any, 0)
		for _, res := range h.svc.Data.Resources {
			rid := intFromAny(res["dbLetter"])
			if rid <= 0 || !boolFromAny(res["hasEconomyModel"]) {
				continue
			}
			info := h.svc.ResourceInfo(rid)
			items = append(items, map[string]any{
				"resourceId":       rid,
				"name":             res["name"],
				"recommendedPrice": info["recommendedPrice"],
				"unitsSoldPerHour": info["unitsSoldPerHour"],
				"producedPerHour":  info["producedPerHour"],
			})
		}
		writeJSON(w, 200, map[string]any{"items": items})
		return
	}
	if strings.HasPrefix(r.URL.Path, "/api/v4/payment-packages/") {
		writeJSON(w, 200, map[string]any{"packages": []map[string]any{{"id": "starter", "priceUSD": 4.99, "simCash": 5000}}})
		return
	}
	writeErr(w, 404, "not found")
}

func (h *Handler) handleContractsIncoming(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, h.svc.Snapshot().ContractsIn)
}

func (h *Handler) handleContractsOutgoing(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, h.svc.Snapshot().ContractsOut)
}

func (h *Handler) handleContractsHistoryIncoming(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, h.svc.Snapshot().ContractsIn)
}

func (h *Handler) handleContractsHistoryOutgoing(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, h.svc.Snapshot().ContractsOut)
}

func (h *Handler) handleWarehouseContractsSummary(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]any{"incoming": 0, "outgoing": 0})
}

func (h *Handler) handleResearch(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]any{"projects": h.svc.ResearchProjects()})
}

func (h *Handler) handleStartResearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "method not allowed")
		return
	}
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "invalid json")
		return
	}
	projectID, _ := body["projectId"].(string)
	if projectID == "" {
		projectID = "research-project-29"
	}
	resp, err := h.svc.StartResearch(h.companyID(r), projectID)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, resp)
}

func (h *Handler) handleResearchProgress(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, h.svc.ResearchProgress())
}

func (h *Handler) handleCompleteResearch(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeErr(w, 405, "method not allowed")
		return
	}
	id := strings.TrimPrefix(r.URL.Path, "/api/v2/research/complete/")
	id = strings.Trim(id, "/")
	if id == "" {
		writeErr(w, 400, "missing project id")
		return
	}
	resp, err := h.svc.CompleteResearch(id)
	if err != nil {
		writeErr(w, 400, err.Error())
		return
	}
	writeJSON(w, 200, resp)
}

func (h *Handler) handleDevTime(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		st := h.svc.Snapshot()
		writeJSON(w, 200, map[string]any{
			"realTime":     time.Now().UTC().Format(time.RFC3339),
			"simulatedAt":  st.SimulatedAt,
			"effectiveNow": h.svc.Now().Format(time.RFC3339),
		})
	case http.MethodPost:
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			writeErr(w, 400, "invalid json")
			return
		}
		if t, ok := body["time"]; ok {
			h.svc.SetSimulatedAt(fmt.Sprint(t))
		} else {
			h.svc.SetSimulatedAt(time.Now().UTC().Add(8 * time.Hour).Format(time.RFC3339)) // +8h default
		}
		writeJSON(w, 200, map[string]any{"simulatedAt": h.svc.Snapshot().SimulatedAt})
	case http.MethodDelete:
		h.svc.SetSimulatedAt("")
		writeJSON(w, 200, map[string]any{"status": "reset"})
	}
}
