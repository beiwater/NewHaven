package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/newhaven/backend-next/internal/app"
	"github.com/newhaven/backend-next/internal/app/market"
	"github.com/newhaven/backend-next/internal/catalog"
	"github.com/newhaven/backend-next/internal/config"
	"github.com/newhaven/backend-next/internal/httpapi"
	"github.com/newhaven/backend-next/internal/platform"
	"github.com/newhaven/backend-next/internal/storage/memory"
)

func newMarketSvc(store *memory.Store) *market.Service {
	resources := map[int]*catalog.ResourceEntry{
		1: {ID: 1, DbLetter: 1, Name: "Grain", IsExchangeTradable: true, IsResearch: false, ProducedPerHourRaw: 500, UnitsSoldAnHour: 150, HasEconomyModel: true, BasePrice: 23},
		2: {ID: 2, DbLetter: 2, Name: "Dairy Milk", IsExchangeTradable: true, IsResearch: false, ProducedPerHourRaw: 420, UnitsSoldAnHour: 130, HasEconomyModel: true, BasePrice: 28},
	}
	clock := platform.NewFakeClock(time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC))
	return market.NewService(store, resources, clock)
}

func registerMarketTestToken(t *testing.T, mux http.Handler, username string) string {
	t.Helper()
	regBody := `{"username":"` + username + `","password":"secret123"}`
	regReq := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	mux.ServeHTTP(regW, regReq)
	if regW.Code != http.StatusOK {
		t.Fatalf("register failed: %d; body: %s", regW.Code, regW.Body.String())
	}
	var regResp apiResponse
	if err := json.Unmarshal(regW.Body.Bytes(), &regResp); err != nil {
		t.Fatalf("unmarshal register: %v", err)
	}
	var regData map[string]any
	if err := json.Unmarshal(regResp.Data, &regData); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	return regData["token"].(string)
}

func TestMarketResources_NoToken_401(t *testing.T) {
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	store := memory.New()
	a := app.New(cfg, store)
	marketHandler := httpapi.NewMarketHandler(newMarketSvc(store))
	authHandler := httpapi.NewAuthHandler(a.AuthService)
	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, nil, marketHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v3/resources/", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestMarketResources_WithToken_200(t *testing.T) {
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	store := memory.New()
	a := app.New(cfg, store)
	marketHandler := httpapi.NewMarketHandler(newMarketSvc(store))
	authHandler := httpapi.NewAuthHandler(a.AuthService)
	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, nil, marketHandler)

	regBody := `{"username":"resuser","password":"secret123"}`
	regReq := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	mux.ServeHTTP(regW, regReq)
	if regW.Code != http.StatusOK {
		t.Fatalf("register failed: %d; body: %s", regW.Code, regW.Body.String())
	}
	var regResp apiResponse
	if err := json.Unmarshal(regW.Body.Bytes(), &regResp); err != nil {
		t.Fatalf("unmarshal register: %v", err)
	}
	var regData map[string]any
	if err := json.Unmarshal(regResp.Data, &regData); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	token := regData["token"].(string)

	req := httptest.NewRequest(http.MethodGet, "/api/v3/resources/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	var data map[string]json.RawMessage
	if err := json.Unmarshal(raw["data"], &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	resArr, ok := data["resources"]
	if !ok {
		t.Fatal("response missing 'resources' field")
	}
	var resources []any
	if err := json.Unmarshal(resArr, &resources); err != nil {
		t.Fatalf("unmarshal resources: %v", err)
	}
	if len(resources) != 2 {
		t.Errorf("expected 2 resources, got %d", len(resources))
	}
}

func TestMarketTicker_NoToken_401(t *testing.T) {
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	store := memory.New()
	a := app.New(cfg, store)
	marketHandler := httpapi.NewMarketHandler(newMarketSvc(store))
	authHandler := httpapi.NewAuthHandler(a.AuthService)
	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, nil, marketHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v3/market-ticker/1/", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestMarketTicker_WithToken_200(t *testing.T) {
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	store := memory.New()
	a := app.New(cfg, store)
	marketHandler := httpapi.NewMarketHandler(newMarketSvc(store))
	authHandler := httpapi.NewAuthHandler(a.AuthService)
	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, nil, marketHandler)

	regBody := `{"username":"tickeruser","password":"secret123"}`
	regReq := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	mux.ServeHTTP(regW, regReq)
	if regW.Code != http.StatusOK {
		t.Fatalf("register failed: %d; body: %s", regW.Code, regW.Body.String())
	}
	var regResp apiResponse
	if err := json.Unmarshal(regW.Body.Bytes(), &regResp); err != nil {
		t.Fatalf("unmarshal register: %v", err)
	}
	var regData map[string]any
	if err := json.Unmarshal(regResp.Data, &regData); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	token := regData["token"].(string)

	req := httptest.NewRequest(http.MethodGet, "/api/v3/market-ticker/1/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	var data map[string]json.RawMessage
	if err := json.Unmarshal(raw["data"], &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if data["resource"] == nil {
		t.Fatal("response missing 'resource' field")
	}
	if data["series"] == nil {
		t.Fatal("response missing 'series' field")
	}
}

func TestMarketTicker_InvalidResource_400(t *testing.T) {
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	store := memory.New()
	a := app.New(cfg, store)
	marketHandler := httpapi.NewMarketHandler(newMarketSvc(store))
	authHandler := httpapi.NewAuthHandler(a.AuthService)
	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, nil, marketHandler)
	token := registerMarketTestToken(t, mux, "tickerbad")

	req := httptest.NewRequest(http.MethodGet, "/api/v3/market-ticker/not-a-number/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestMarketDepth_WithToken_200(t *testing.T) {
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	store := memory.New()
	a := app.New(cfg, store)
	marketHandler := httpapi.NewMarketHandler(newMarketSvc(store))
	authHandler := httpapi.NewAuthHandler(a.AuthService)
	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, nil, marketHandler)

	regBody := `{"username":"depthuser","password":"secret123"}`
	regReq := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	mux.ServeHTTP(regW, regReq)
	if regW.Code != http.StatusOK {
		t.Fatalf("register failed: %d; body: %s", regW.Code, regW.Body.String())
	}
	var regResp apiResponse
	if err := json.Unmarshal(regW.Body.Bytes(), &regResp); err != nil {
		t.Fatalf("unmarshal register: %v", err)
	}
	var regData map[string]any
	if err := json.Unmarshal(regResp.Data, &regData); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	token := regData["token"].(string)

	req := httptest.NewRequest(http.MethodGet, "/api/v3/market-depth/1/0/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	var data map[string]json.RawMessage
	if err := json.Unmarshal(raw["data"], &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if data["buys"] == nil {
		t.Fatal("response missing 'buys' field")
	}
	if data["sells"] == nil {
		t.Fatal("response missing 'sells' field")
	}
}

func TestMarketDepth_InvalidQuality_400(t *testing.T) {
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	store := memory.New()
	a := app.New(cfg, store)
	marketHandler := httpapi.NewMarketHandler(newMarketSvc(store))
	authHandler := httpapi.NewAuthHandler(a.AuthService)
	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, nil, marketHandler)
	token := registerMarketTestToken(t, mux, "depthbad")

	req := httptest.NewRequest(http.MethodGet, "/api/v3/market-depth/1/-1/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestMarketOrders_WithToken_200(t *testing.T) {
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	store := memory.New()
	a := app.New(cfg, store)
	marketHandler := httpapi.NewMarketHandler(newMarketSvc(store))
	authHandler := httpapi.NewAuthHandler(a.AuthService)
	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, nil, marketHandler)

	regBody := `{"username":"orduser","password":"secret123"}`
	regReq := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	mux.ServeHTTP(regW, regReq)
	if regW.Code != http.StatusOK {
		t.Fatalf("register failed: %d; body: %s", regW.Code, regW.Body.String())
	}
	var regResp apiResponse
	if err := json.Unmarshal(regW.Body.Bytes(), &regResp); err != nil {
		t.Fatalf("unmarshal register: %v", err)
	}
	var regData map[string]any
	if err := json.Unmarshal(regResp.Data, &regData); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	token := regData["token"].(string)

	req := httptest.NewRequest(http.MethodGet, "/api/v3/market/1/0/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}
	var data map[string]json.RawMessage
	if err := json.Unmarshal(raw["data"], &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	if data["orders"] == nil {
		t.Fatal("response missing 'orders' field")
	}
}
