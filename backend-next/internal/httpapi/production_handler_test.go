package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/newhaven/backend-next/internal/app"
	"github.com/newhaven/backend-next/internal/app/production"
	"github.com/newhaven/backend-next/internal/config"
	"github.com/newhaven/backend-next/internal/httpapi"
	"github.com/newhaven/backend-next/internal/storage/memory"
)

func TestListProductionJobs_NoToken_401(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store)
	productionSvc := production.NewService(store)
	productionHandler := httpapi.NewProductionHandler(productionSvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, productionHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/production/jobs/", nil)
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

func TestListProductionJobs_WithToken_200(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store)
	productionSvc := production.NewService(store)
	productionHandler := httpapi.NewProductionHandler(productionSvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, productionHandler)

	// Register a user to get a valid token
	registerBody := `{"username":"produser","password":"secret123"}`
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

	// Now hit the production jobs endpoint with the token
	req := httptest.NewRequest(http.MethodGet, "/api/v2/production/jobs/", nil)
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

	// Verify jobs data in response
	var data map[string]any
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}

	jobs, ok := data["jobs"]
	if !ok {
		t.Fatal("expected jobs field in response data")
	}
	jobsList, ok := jobs.([]any)
	if !ok {
		t.Fatalf("expected jobs to be array, got %T", jobs)
	}
	if jobsList == nil {
		t.Fatal("expected jobs to be non-nil array")
	}
	// Dev bootstrap does not seed production jobs, so the list may be empty,
	// but the array must not be null.
}

func TestListProductionJobs_EmptyJobsIsArray_200(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store)
	productionSvc := production.NewService(store)
	productionHandler := httpapi.NewProductionHandler(productionSvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, productionHandler)

	// Register a user to get a valid token
	registerBody := `{"username":"emptyjobs","password":"secret123"}`
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

	// Now hit the production jobs endpoint with the token
	req := httptest.NewRequest(http.MethodGet, "/api/v2/production/jobs/", nil)
	req.Header.Set("Authorization", "Bearer "+token.(string))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	// Verify the raw JSON to ensure jobs is [] not null
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}

	var data map[string]json.RawMessage
	if err := json.Unmarshal(raw["data"], &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}

	jobsRaw, ok := data["jobs"]
	if !ok {
		t.Fatal("response data missing 'jobs' field")
	}

	var jobs []any
	if err := json.Unmarshal(jobsRaw, &jobs); err != nil {
		t.Fatalf("unmarshal jobs: %v", err)
	}

	if jobs == nil {
		t.Fatal("expected jobs to be non-nil array, got nil")
	}
}
