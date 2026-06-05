package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/newhaven/backend-next/internal/app"
	"github.com/newhaven/backend-next/internal/app/warehouse"
	"github.com/newhaven/backend-next/internal/config"
	"github.com/newhaven/backend-next/internal/httpapi"
	"github.com/newhaven/backend-next/internal/storage/memory"
)

func TestGetMyWarehouse_NoToken_401(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store)
	warehouseSvc := warehouse.NewService(store, store, a.Logger)
	warehouseHandler := httpapi.NewWarehouseHandler(warehouseSvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, authHandler, nil, warehouseHandler)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/companies/me/warehouse/", nil)
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

func TestGetMyWarehouse_WithToken_200(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store)
	warehouseSvc := warehouse.NewService(store, store, a.Logger)
	warehouseHandler := httpapi.NewWarehouseHandler(warehouseSvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, authHandler, nil, warehouseHandler)

	// Register a user to get a valid token
	registerBody := `{"username":"whuser","password":"secret123"}`
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

	// Now hit the warehouse endpoint with the token
	req := httptest.NewRequest(http.MethodGet, "/api/v2/companies/me/warehouse/", nil)
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

	// Verify warehouse data in response
	var data map[string]any
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}

	companyID, ok := data["company_id"]
	if !ok {
		t.Fatal("expected company_id field in response data")
	}
	companyIDFloat, ok := companyID.(float64)
	if !ok || companyIDFloat <= 0 {
		t.Fatalf("expected positive company_id, got %v (%T)", companyID, companyID)
	}

	capacity, ok := data["capacity"]
	if !ok {
		t.Fatal("expected capacity field")
	}
	if capacity.(float64) != 1000 {
		t.Errorf("expected capacity 1000, got %v", capacity)
	}

	usedCapacity, ok := data["used_capacity"]
	if !ok {
		t.Fatal("expected used_capacity field")
	}
	if usedCapacity.(float64) != 0 {
		t.Errorf("expected used_capacity 0, got %v", usedCapacity)
	}

	items, ok := data["items"]
	if !ok {
		t.Fatal("expected items field in response data")
	}
	itemsList, ok := items.([]any)
	if !ok {
		t.Fatalf("expected items to be array, got %T", items)
	}
	if itemsList == nil {
		t.Fatal("expected items to be non-nil array")
	}
	if len(itemsList) != 0 {
		t.Errorf("expected empty items array, got %d items", len(itemsList))
	}
}

func TestGetMyWarehouse_ReturnsItemsArrayNotNull(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store)
	warehouseSvc := warehouse.NewService(store, store, a.Logger)
	warehouseHandler := httpapi.NewWarehouseHandler(warehouseSvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, authHandler, nil, warehouseHandler)

	// Register a user
	registerBody := `{"username":"itemsuser","password":"secret123"}`
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

	req := httptest.NewRequest(http.MethodGet, "/api/v2/companies/me/warehouse/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(w.Body.Bytes(), &raw); err != nil {
		t.Fatalf("unmarshal envelope: %v", err)
	}

	// Parse the data field to check items
	var data map[string]json.RawMessage
	if err := json.Unmarshal(raw["data"], &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}

	itemsRaw, ok := data["items"]
	if !ok {
		t.Fatal("response data missing 'items' field")
	}

	var items []any
	if err := json.Unmarshal(itemsRaw, &items); err != nil {
		t.Fatalf("unmarshal items: %v", err)
	}

	if items == nil {
		t.Fatal("expected items to be non-nil array, got nil")
	}
}
