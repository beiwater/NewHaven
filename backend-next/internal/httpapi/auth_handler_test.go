package httpapi_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/newhaven/backend-next/internal/app"
	"github.com/newhaven/backend-next/internal/config"
	"github.com/newhaven/backend-next/internal/httpapi"
	"github.com/newhaven/backend-next/internal/storage/memory"
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
	a := app.New(cfg, store)
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

	// Verify token field exists in data.
	var data map[string]any
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	token, ok := data["token"]
	if !ok {
		t.Fatal("expected token field in response data")
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestRegister_MissingFields(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store)
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

func TestLogin_Handler(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store)
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
	token, ok := data["token"]
	if !ok {
		t.Fatal("expected token in response data")
	}
	if token == "" {
		t.Error("expected non-empty token")
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store)
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
	a := app.New(cfg, store)
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
