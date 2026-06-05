package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/newhaven/backend-next/internal/app"
	"github.com/newhaven/backend-next/internal/app/company"
	"github.com/newhaven/backend-next/internal/config"
	"github.com/newhaven/backend-next/internal/httpapi"
	"github.com/newhaven/backend-next/internal/storage/memory"
)

func TestListMyCompanies_NoToken_401(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store)
	companySvc := company.NewService(store, a.Logger)
	companyHandler := httpapi.NewCompanyHandler(companySvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, authHandler, companyHandler, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/players/me/companies/", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp apiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error == nil {
		t.Fatal("expected error, got nil")
	}
	if resp.Error.Code != "UNAUTHORIZED" {
		t.Errorf("expected code UNAUTHORIZED, got %s", resp.Error.Code)
	}
}

func TestListMyCompanies_WithToken_200(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store)
	companySvc := company.NewService(store, a.Logger)
	companyHandler := httpapi.NewCompanyHandler(companySvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, authHandler, companyHandler, nil, nil, nil)

	// First, register a user to get a valid token
	registerBody := `{"username":"testuser","password":"secret123"}`
	regReq := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(registerBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	mux.ServeHTTP(regW, regReq)

	if regW.Code != http.StatusOK {
		t.Fatalf("register failed: %d; body: %s", regW.Code, regW.Body.String())
	}

	var regResp apiResponse
	if err := json.Unmarshal(regW.Body.Bytes(), &regResp); err != nil {
		t.Fatalf("unmarshal register response: %v", err)
	}
	if regResp.Error != nil {
		t.Fatalf("register returned error: %+v", *regResp.Error)
	}
	var regData map[string]any
	if err := json.Unmarshal(regResp.Data, &regData); err != nil {
		t.Fatalf("unmarshal register data: %v", err)
	}
	token, ok := regData["token"]
	if !ok || token == "" {
		t.Fatal("register did not return a token")
	}

	// Now hit the companies endpoint with the token
	req := httptest.NewRequest(http.MethodGet, "/api/v2/players/me/companies/", nil)
	req.Header.Set("Authorization", "Bearer "+token.(string))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp apiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", *resp.Error)
	}

	// Verify companies list in data
	var data map[string]any
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	companies, ok := data["companies"]
	if !ok {
		t.Fatal("expected companies field in response data")
	}
	companiesList, ok := companies.([]any)
	if !ok {
		t.Fatalf("expected companies to be array, got %T", companies)
	}
	if len(companiesList) == 0 {
		t.Fatal("expected at least one company")
	}

	companyData := companiesList[0].(map[string]any)
	if name, ok := companyData["name"]; !ok || name == "" {
		t.Error("expected company with a name")
	}
}

func TestListMyCompanies_ReturnsEnvelope(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store)
	companySvc := company.NewService(store, a.Logger)
	companyHandler := httpapi.NewCompanyHandler(companySvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, authHandler, companyHandler, nil, nil, nil)

	// Register and login
	registerBody := `{"username":"envtest","password":"secret123"}`
	regReq := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(registerBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	mux.ServeHTTP(regW, regReq)

	if regW.Code != http.StatusOK {
		t.Fatalf("register failed: %d", regW.Code)
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

	req := httptest.NewRequest(http.MethodGet, "/api/v2/players/me/companies/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	// Verify the envelope structure
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}

	if _, ok := raw["data"]; !ok {
		t.Error("response missing 'data' field")
	}
	if _, ok := raw["error"]; !ok {
		t.Error("response missing 'error' field")
	}
	if string(raw["error"]) != "null" {
		t.Errorf("expected error null, got %s", string(raw["error"]))
	}
}
