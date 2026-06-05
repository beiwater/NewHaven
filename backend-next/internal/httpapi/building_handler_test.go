package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/newhaven/backend-next/internal/app"
	"github.com/newhaven/backend-next/internal/app/building"
	"github.com/newhaven/backend-next/internal/config"
	"github.com/newhaven/backend-next/internal/httpapi"
	"github.com/newhaven/backend-next/internal/storage/memory"
)

func TestListMyBuildings_NoToken_401(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store)
	buildingSvc := building.NewService(store)
	buildingHandler := httpapi.NewBuildingHandler(buildingSvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, buildingHandler, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v3/companies/me/buildings/", nil)
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

func TestListMyBuildings_WithToken_200(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store)
	buildingSvc := building.NewService(store)
	buildingHandler := httpapi.NewBuildingHandler(buildingSvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, buildingHandler, nil, nil, nil)

	// Register a user to get a valid token
	registerBody := `{"username":"blduser","password":"secret123"}`
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

	// Now hit the buildings endpoint with the token
	req := httptest.NewRequest(http.MethodGet, "/api/v3/companies/me/buildings/", nil)
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

	// Verify buildings data in response
	var data map[string]any
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}

	buildings, ok := data["buildings"]
	if !ok {
		t.Fatal("expected buildings field in response data")
	}
	buildingsList, ok := buildings.([]any)
	if !ok {
		t.Fatalf("expected buildings to be array, got %T", buildings)
	}
	if len(buildingsList) != 2 {
		t.Fatalf("expected 2 buildings, got %d", len(buildingsList))
	}

	first := buildingsList[0].(map[string]any)
	if name, ok := first["name"]; !ok || name == "" {
		t.Error("expected building with a name")
	}
	if id, ok := first["id"]; !ok || id == "" {
		t.Error("expected building with an id")
	}
}

func TestListMyBuildings_EmptyArray_200(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store)
	buildingSvc := building.NewService(store)
	buildingHandler := httpapi.NewBuildingHandler(buildingSvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, buildingHandler, nil, nil, nil)

	// Register a user to get a valid token
	registerBody := `{"username":"emptyman","password":"secret123"}`
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

	// Now hit the buildings endpoint with the token
	req := httptest.NewRequest(http.MethodGet, "/api/v3/companies/me/buildings/", nil)
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

	var data map[string]any
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}

	buildings, ok := data["buildings"]
	if !ok {
		t.Fatal("expected buildings field in response data")
	}
	buildingsList, ok := buildings.([]any)
	if !ok {
		t.Fatalf("expected buildings to be array, got %T", buildings)
	}
	if len(buildingsList) != 2 {
		t.Errorf("expected 2 buildings, got %d", len(buildingsList))
	}
}
