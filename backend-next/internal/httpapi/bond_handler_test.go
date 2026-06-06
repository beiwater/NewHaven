package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/newhaven/backend-next/internal/app"
	appfinance "github.com/newhaven/backend-next/internal/app/finance"
	"github.com/newhaven/backend-next/internal/config"
	"github.com/newhaven/backend-next/internal/httpapi"
	"github.com/newhaven/backend-next/internal/platform"
	"github.com/newhaven/backend-next/internal/storage/memory"
)

func newBondSvc(store *memory.Store) *appfinance.Service {
	clock := platform.NewFakeClock(time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC))
	idgen := platform.NewIDGen()
	cfg := &config.GameConfig{BondFaceValue: 5000, BondMinInterest: 0.5, BondMaxInterest: 2.0}
	return appfinance.NewService(store, store, clock, idgen, cfg)
}

func registerBondTestToken(t *testing.T, mux http.Handler, username string) string {
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

// GET /api/bonds/ - no token -> 401
func TestListBonds_NoToken_401(t *testing.T) {
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	store := memory.New()
	a := app.New(cfg, store)
	bondHandler := httpapi.NewBondHandler(newBondSvc(store))
	authHandler := httpapi.NewAuthHandler(a.AuthService)
	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, nil, nil, nil, bondHandler, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/bonds/", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d; body: %s", w.Code, w.Body.String())
	}
}

// GET /api/bonds/ - with token -> 200, empty list
func TestListBonds_WithToken_200(t *testing.T) {
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	store := memory.New()
	a := app.New(cfg, store)
	bondHandler := httpapi.NewBondHandler(newBondSvc(store))
	authHandler := httpapi.NewAuthHandler(a.AuthService)
	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, nil, nil, nil, bondHandler, nil, nil, nil, nil, nil, nil)

	token := registerBondTestToken(t, mux, "listbondsuser")

	req := httptest.NewRequest(http.MethodGet, "/api/bonds/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp apiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	var data map[string]any
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	bonds, ok := data["bonds"].([]any)
	if !ok {
		t.Fatal("expected bonds array in response")
	}
	if len(bonds) != 0 {
		t.Errorf("expected empty bonds, got %d", len(bonds))
	}
}

// POST /api/bonds/ - invalid JSON -> 400
func TestCreateBond_InvalidJSON_400(t *testing.T) {
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	store := memory.New()
	a := app.New(cfg, store)
	bondHandler := httpapi.NewBondHandler(newBondSvc(store))
	authHandler := httpapi.NewAuthHandler(a.AuthService)
	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, nil, nil, nil, bondHandler, nil, nil, nil, nil, nil, nil)

	token := registerBondTestToken(t, mux, "createbondinvalidjson")

	body := `{invalid json`
	req := httptest.NewRequest(http.MethodPost, "/api/bonds/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d; body: %s", w.Code, w.Body.String())
	}
}

// POST /api/bonds/ - invalid payload (missing fields) -> 400
func TestCreateBond_InvalidPayload_400(t *testing.T) {
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	store := memory.New()
	a := app.New(cfg, store)
	bondHandler := httpapi.NewBondHandler(newBondSvc(store))
	authHandler := httpapi.NewAuthHandler(a.AuthService)
	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, nil, nil, nil, bondHandler, nil, nil, nil, nil, nil, nil)

	token := registerBondTestToken(t, mux, "createbondinvalidpay")

	body := `{"amount": 0, "interest": 0}`
	req := httptest.NewRequest(http.MethodPost, "/api/bonds/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d; body: %s", w.Code, w.Body.String())
	}
}

// POST /api/bonds/ - success
func TestCreateBond_Success_200(t *testing.T) {
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	store := memory.New()
	a := app.New(cfg, store)
	bondHandler := httpapi.NewBondHandler(newBondSvc(store))
	authHandler := httpapi.NewAuthHandler(a.AuthService)
	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, nil, nil, nil, bondHandler, nil, nil, nil, nil, nil, nil)

	token := registerBondTestToken(t, mux, "createbondsuccess")

	body := `{"amount": 5, "interest": 1.2}`
	req := httptest.NewRequest(http.MethodPost, "/api/bonds/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp apiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	var data map[string]any
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	bond, ok := data["bond"].(map[string]any)
	if !ok {
		t.Fatal("expected bond object in response")
	}
	if id, ok := bond["id"].(string); !ok || id == "" {
		t.Error("expected non-empty bond id")
	}
}

// GET /api/bonds/{bondId}/ - not found -> 404
func TestGetBond_NotFound_404(t *testing.T) {
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	store := memory.New()
	a := app.New(cfg, store)
	bondHandler := httpapi.NewBondHandler(newBondSvc(store))
	authHandler := httpapi.NewAuthHandler(a.AuthService)
	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, nil, nil, nil, bondHandler, nil, nil, nil, nil, nil, nil)

	token := registerBondTestToken(t, mux, "getbondnotfound")

	req := httptest.NewRequest(http.MethodGet, "/api/bonds/nonexistent-id/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d; body: %s", w.Code, w.Body.String())
	}
}

// GET /api/bonds/{bondId}/ - success
func TestGetBond_Success_200(t *testing.T) {
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	store := memory.New()
	a := app.New(cfg, store)
	bondHandler := httpapi.NewBondHandler(newBondSvc(store))
	authHandler := httpapi.NewAuthHandler(a.AuthService)
	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, nil, nil, nil, bondHandler, nil, nil, nil, nil, nil, nil)

	token := registerBondTestToken(t, mux, "getbondsuccess")

	// First create a bond
	createBody := `{"amount": 3, "interest": 1.0}`
	createReq := httptest.NewRequest(http.MethodPost, "/api/bonds/", strings.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.Header.Set("Authorization", "Bearer "+token)
	createW := httptest.NewRecorder()
	mux.ServeHTTP(createW, createReq)
	if createW.Code != http.StatusOK {
		t.Fatalf("create bond failed: %d; body: %s", createW.Code, createW.Body.String())
	}
	var createResp apiResponse
	json.Unmarshal(createW.Body.Bytes(), &createResp)
	var createData map[string]any
	json.Unmarshal(createResp.Data, &createData)
	bondData := createData["bond"].(map[string]any)
	bondID := bondData["id"].(string)

	// Now get it by ID
	req := httptest.NewRequest(http.MethodGet, "/api/bonds/"+bondID+"/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
}

// GET /api/v2/companies/me/bonds/owned/ - success
func TestGetOwnedBonds_Success_200(t *testing.T) {
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	store := memory.New()
	a := app.New(cfg, store)
	bondHandler := httpapi.NewBondHandler(newBondSvc(store))
	authHandler := httpapi.NewAuthHandler(a.AuthService)
	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, nil, nil, nil, bondHandler, nil, nil, nil, nil, nil, nil)

	token := registerBondTestToken(t, mux, "ownedbonds")

	req := httptest.NewRequest(http.MethodGet, "/api/v2/companies/me/bonds/owned/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
}

// GET /api/v2/companies/me/bonds/sold/ - success
func TestGetSoldBonds_Success_200(t *testing.T) {
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	store := memory.New()
	a := app.New(cfg, store)
	bondHandler := httpapi.NewBondHandler(newBondSvc(store))
	authHandler := httpapi.NewAuthHandler(a.AuthService)
	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, nil, nil, nil, bondHandler, nil, nil, nil, nil, nil, nil)

	token := registerBondTestToken(t, mux, "soldbonds")

	req := httptest.NewRequest(http.MethodGet, "/api/v2/companies/me/bonds/sold/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}
}
