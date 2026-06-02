package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"go-sim-api/internal/formula"
)

func (h *Handler) RegisterCompany(mux *http.ServeMux) {
	mux.HandleFunc("/api/csrf/", h.withAuth(h.handleCSRF))
	mux.HandleFunc("/api/v2/players/me/companies/", h.withAuth(h.handlePlayersMeCompanies))
	mux.HandleFunc("/api/v2/players/", h.withAuth(h.handlePlayersByID))
	mux.HandleFunc("/api/v2/companies/me/buildings/", h.withAuth(h.handleCompaniesMeBuildings))
	mux.HandleFunc("/api/v2/companies/me/administration-overhead/", h.withAuth(h.handleAdminOverhead))
	mux.HandleFunc("/api/v3/companies/", h.withAuth(h.handleV3Companies))
	mux.HandleFunc("/api/v2/companies/", h.withAuth(h.handleV2Companies))
	// Executive recruitment
	mux.HandleFunc("/api/v2/executives/search/", h.withAuth(h.handleExecSearch))
	mux.HandleFunc("/api/v2/executives/recruit/", h.withAuth(h.handleExecRecruit))
	mux.HandleFunc("/api/v2/executives/train/", h.withAuth(h.handleExecTrain))
	mux.HandleFunc("/api/v3/executives/poach/", h.withAuth(h.handleExecPoach))
	mux.HandleFunc("/api/v3/executives/offers/", h.withAuth(h.handleExecOffers))
	mux.HandleFunc("/api/v3/executives/", h.withAuth(h.handleV3ExecutiveByID))
	// Building auctions
	mux.HandleFunc("/api/v2/companies/me/auctions/", h.withAuth(h.handleMyAuctions))
	mux.HandleFunc("/api/v2/auctions/", h.withAuth(h.handleAuctions))
	mux.HandleFunc("/api/v2/companies/me/warehouse/", h.withAuth(h.handleWarehouse))
}

func (h *Handler) handleCSRF(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, map[string]any{"csrfToken": h.svc.Snapshot().CSRFToken})
}

func (h *Handler) handlePlayersMeCompanies(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, h.svc.CompaniesByPlayer("dev-player"))
}

func (h *Handler) handlePlayersByID(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost && strings.Contains(r.URL.Path, "/preferences/") {
		h.handleSavePreferences(w, r)
		return
	}
	if strings.Contains(r.URL.Path, "/companies/") {
		writeJSON(w, 200, h.svc.CompaniesByPlayer("external-player"))
		return
	}
	writeErr(w, 404, "not found")
}

type SavePreferencesRequest struct {
	Preferences map[string]any `json:"-"`
}

func (h *Handler) handleSavePreferences(w http.ResponseWriter, r *http.Request) {
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeErr(w, 400, "invalid json")
		return
	}
	writeJSON(w, 200, h.svc.UpdatePreferences(body))
}

func (h *Handler) handleCompaniesMeBuildings(w http.ResponseWriter, r *http.Request) {
	snap := h.svc.Snapshot()
	company := snap.GetCompany(h.companyID(r))
	buildings := make([]map[string]any, 0, len(company.PlacedBuildings)+len(company.UnplacedBuildings))
	for _, b := range company.PlacedBuildings {
		copy := cloneMap(b)
		copy["placed"] = true
		buildings = append(buildings, copy)
	}
	for _, b := range company.UnplacedBuildings {
		copy := cloneMap(b)
		copy["placed"] = false
		buildings = append(buildings, copy)
	}
	writeJSON(w, 200, buildings)
}

func (h *Handler) handleAdminOverhead(w http.ResponseWriter, _ *http.Request) {
	coo, _ := h.svc.COOandCTOSkill()
	base := h.svc.Cfg.Game.AdminOverheadBase
	writeJSON(w, 200, map[string]any{
		"baseOverhead": base,
		"cooSkill":     coo,
		"multiplier":   formula.AdminOverheadWithCOO(base, coo),
	})
}

func (h *Handler) handleV3Companies(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeErr(w, 405, "method not allowed")
		return
	}
	writeJSON(w, 200, h.svc.CompanyProfile(h.companyID(r)))
}

func (h *Handler) handleV2Companies(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/v2/companies/")
	switch {
	case strings.Contains(path, "/collectibles/"):
		h.handleCompanyCollectibles(w, r)
	case strings.Contains(path, "/game-notifications/"):
		h.handleCompanyNotifications(w, r)
	case strings.Contains(path, "/market-orders/"):
		h.handleCompanyMarketOrders(w, r)
	case strings.Contains(path, "/certificates/"):
		h.handleCompanyCertificates(w, r)
	case strings.Contains(path, "/display-case/"):
		h.handleCompanyDisplayCase(w, r)
	case strings.Contains(path, "/former-executives/"):
		h.handleCompanyFormerExecutives(w, r)
	case strings.Contains(path, "/royalties/"):
		h.handleCompanyRoyalties(w, r)
	case strings.Contains(path, "/egg-collection/"):
		h.handleCompanyEggCollection(w, r)
	case strings.Contains(path, "/tags/"):
		h.handleCompanyTags(w, r)
	default:
		writeErr(w, 404, "unknown path")
	}
}

func (h *Handler) handleCompanyCollectibles(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, []map[string]any{{"id": 1, "name": "Golden Spatula"}})
}

func (h *Handler) handleCompanyNotifications(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, h.svc.Snapshot().Notifications)
}

func (h *Handler) handleCompanyMarketOrders(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, h.svc.Snapshot().Orders)
}

func (h *Handler) handleCompanyCertificates(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, []any{})
}

func (h *Handler) handleCompanyDisplayCase(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, []any{})
}

func (h *Handler) handleCompanyFormerExecutives(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, []any{})
}

func (h *Handler) handleCompanyRoyalties(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, []any{})
}

func (h *Handler) handleCompanyEggCollection(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, 200, []any{})
}

func (h *Handler) handleCompanyTags(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPatch {
		writeJSON(w, 200, map[string]any{"ok": true})
		return
	}
	writeJSON(w, 200, []string{"industry-food", "market-maker"})
}

func (h *Handler) handleMyAuctions(w http.ResponseWriter, r *http.Request) {
	auctions, err := h.svc.MyAuctionList(h.companyID(r))
	if err != nil {
		writeErr(w, 500, err.Error())
		return
	}
	writeJSON(w, 200, map[string]any{"auctions": auctions})
}

func (h *Handler) handleWarehouse(w http.ResponseWriter, r *http.Request) {
	snap := h.svc.Snapshot()
	c := snap.GetCompany(h.companyID(r))
	inv := make([]map[string]any, 0)
	for rid, qty := range c.Inventory {
		name := ""
		for _, rr := range h.svc.Data.Resources {
			if intFromAny(rr["dbLetter"]) == rid {
				name = fmt.Sprint(rr["name"])
				break
			}
		}
		inv = append(inv, map[string]any{
			"resourceId": rid, "name": name, "quantity": qty,
			"quality": 0, "estimatedValue": 0,
		})
	}
	if c.QualityInventory != nil {
		for key, qty := range c.QualityInventory {
			parts := strings.Split(key, "_")
			if len(parts) == 2 {
				rid, _ := strconv.Atoi(parts[0])
				qual, _ := strconv.Atoi(parts[1])
				name := ""
				for _, rr := range h.svc.Data.Resources {
					if intFromAny(rr["dbLetter"]) == rid {
						name = fmt.Sprint(rr["name"])
						break
					}
				}
				inv = append(inv, map[string]any{
					"resourceId": rid, "name": name, "quantity": qty,
					"quality": qual, "estimatedValue": 0,
				})
			}
		}
	}
	writeJSON(w, 200, map[string]any{
		"inventory": inv, "capacity": 2000, "used": len(c.Inventory) + len(c.QualityInventory),
	})
}
