package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/beiwater/NewHaven/backend/internal/app"
	"github.com/beiwater/NewHaven/backend/internal/config"
	"github.com/beiwater/NewHaven/backend/internal/httpapi"
	"github.com/beiwater/NewHaven/backend/internal/storage/memory"
)

// apiResponse is the standard envelope returned by the HTTP layer.
type apiResponse struct {
	Data  json.RawMessage `json:"data"`
	Error *apiErr         `json:"error"`
}

type apiErr struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func TestRegister_Handler(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store, nil, nil, nil)
	handler := httpapi.NewAuthHandler(a.AuthService)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	body := `{"username":"alice","password":"secret123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
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
	if len(resp.Data) == 0 {
		t.Fatal("expected non-empty data")
	}

	// Verify all expected fields in response data.
	var data map[string]any
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	for _, k := range []string{"token", "player_id", "company_id", "username"} {
		if _, ok := data[k]; !ok {
			t.Errorf("expected %q field in response data", k)
		}
	}
	if token, _ := data["token"].(string); token == "" {
		t.Error("expected non-empty token")
	}
}

func TestRegister_MissingFields(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store, nil, nil, nil)
	handler := httpapi.NewAuthHandler(a.AuthService)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Empty JSON body — both username and password missing.
	body := `{}`
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp apiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error == nil {
		t.Fatal("expected error in response")
	}
	if resp.Error.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code VALIDATION_ERROR, got %q", resp.Error.Code)
	}
}

func TestRegister_Duplicate(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store, nil, nil, nil)
	handler := httpapi.NewAuthHandler(a.AuthService)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	body := `{"username":"dupuser","password":"secret123"}`

	// First register — should succeed.
	req1 := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader([]byte(body)))
	req1.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req1)
	if w.Code != http.StatusOK {
		t.Fatalf("first register: expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	// Register again with same credentials — should get CONFLICT.
	req2 := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader([]byte(body)))
	req2.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req2)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("duplicate: expected 400, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp apiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error == nil {
		t.Fatal("expected error in response")
	}
	if resp.Error.Code != "CONFLICT" {
		t.Errorf("expected code CONFLICT, got %q", resp.Error.Code)
	}
	if resp.Error.Message == "" {
		t.Error("expected non-empty error message")
	}
	if string(resp.Data) != "null" {
		t.Error("expected null data on error")
	}
}

func TestLogin_Handler(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store, nil, nil, nil)
	handler := httpapi.NewAuthHandler(a.AuthService)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// First register.
	regBody := `{"username":"bob","password":"hunter2"}`
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader([]byte(regBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("register: expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	// Then login with same credentials.
	loginBody := `{"username":"bob","password":"hunter2"}`
	req = httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader([]byte(loginBody)))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("login: expected 200, got %d; body: %s", w.Code, w.Body.String())
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
	for _, k := range []string{"token", "player_id", "company_id", "username"} {
		if _, ok := data[k]; !ok {
			t.Errorf("expected %q field in response data", k)
		}
	}
	if token, _ := data["token"].(string); token == "" {
		t.Error("expected non-empty token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store, nil, nil, nil)
	handler := httpapi.NewAuthHandler(a.AuthService)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	// Register first.
	regBody := `{"username":"carol","password":"correctpw"}`
	req := httptest.NewRequest(http.MethodPost, "/api/register", bytes.NewReader([]byte(regBody)))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("register: expected 200, got %d; body: %s", w.Code, w.Body.String())
	}

	// Login with wrong password.
	loginBody := `{"username":"carol","password":"wrongpw"}`
	req = httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader([]byte(loginBody)))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d; body: %s", w.Code, w.Body.String())
	}

	var resp apiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Error == nil {
		t.Fatal("expected error in response")
	}
	if resp.Error.Code != "UNAUTHORIZED" {
		t.Errorf("expected code UNAUTHORIZED, got %q", resp.Error.Code)
	}
}

func TestLogin_NonExistentUser(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store, nil, nil, nil)
	handler := httpapi.NewAuthHandler(a.AuthService)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	body := `{"username":"nobody","password":"anything"}`
	req := httptest.NewRequest(http.MethodPost, "/api/login", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
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
		t.Fatal("expected error in response")
	}
	if resp.Error.Code != "UNAUTHORIZED" {
		t.Errorf("expected code UNAUTHORIZED, got %q", resp.Error.Code)
	}
}
