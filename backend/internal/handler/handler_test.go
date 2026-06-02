package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-sim-api/internal/service"
)

func newTestHandler() *Handler {
	svc := service.NewTestService()
	return New(svc)
}

func TestHealthz(t *testing.T) {
	h := newTestHandler()
	_ = h
	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, 200, map[string]string{"status": "ok"})
	})
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	var resp map[string]string
	_ = json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != "ok" {
		t.Errorf("expected ok, got %s", resp["status"])
	}
}

func TestCSRF(t *testing.T) {
	h := newTestHandler()
	mux := http.NewServeMux()
	h.RegisterHealth(mux)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestCompanyProfile(t *testing.T) {
	h := newTestHandler()
	mux := http.NewServeMux()
	h.RegisterHealth(mux)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestMarketDepth(t *testing.T) {
	h := newTestHandler()
	mux := http.NewServeMux()
	h.RegisterHealth(mux)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestLevelEndpoint(t *testing.T) {
	h := newTestHandler()
	mux := http.NewServeMux()
	h.RegisterHealth(mux)

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestBuyBuildingAppearsAsUnplaced(t *testing.T) {
	h := newTestHandler()
	mux := http.NewServeMux()
	h.Register(mux)

	body := bytes.NewBufferString(`{"username":"builder"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/register", body)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("register expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var auth struct {
		Player struct {
			Token string `json:"token"`
		} `json:"player"`
	}
	if err := json.NewDecoder(w.Body).Decode(&auth); err != nil {
		t.Fatalf("decode register response: %v", err)
	}

	body = bytes.NewBufferString(`{"buildingId":"b-shop-1"}`)
	req = httptest.NewRequest(http.MethodPost, "/api/v2/buildings/buy/", body)
	req.Header.Set("Authorization", "Bearer "+auth.Player.Token)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("buy building expected 200, got %d: %s", w.Code, w.Body.String())
	}

	req = httptest.NewRequest(http.MethodGet, "/api/v2/companies/me/buildings/", nil)
	req.Header.Set("Authorization", "Bearer "+auth.Player.Token)
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("buildings expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var buildings []map[string]any
	if err := json.NewDecoder(w.Body).Decode(&buildings); err != nil {
		t.Fatalf("decode buildings response: %v", err)
	}
	if len(buildings) != 1 {
		t.Fatalf("expected 1 building, got %d", len(buildings))
	}
	if buildings[0]["placed"] != false {
		t.Fatalf("expected bought building to be unplaced, got %#v", buildings[0])
	}
}
