package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/newhaven/backend-next/internal/app"
	"github.com/newhaven/backend-next/internal/app/finance"
	"github.com/newhaven/backend-next/internal/config"
	"github.com/newhaven/backend-next/internal/httpapi"
	"github.com/newhaven/backend-next/internal/platform"
	"github.com/newhaven/backend-next/internal/storage/memory"
)

func newFinanceSvc(store *memory.Store) *finance.Service {
	clock := platform.NewFakeClock(time.Date(2026, 6, 7, 12, 0, 0, 0, time.UTC))
	idgen := platform.NewIDGen()
	cfg := &config.GameConfig{BondFaceValue: 5000, BondMinInterest: 0.5, BondMaxInterest: 2.0}
	return finance.NewService(store, store, clock, idgen, cfg)
}

func registerFinanceTestToken(t *testing.T, mux http.Handler, username string) string {
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

func TestFinanceRecentCashflow_NoToken_401(t *testing.T) {
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	store := memory.New()
	a := app.New(cfg, store)
	financeHandler := httpapi.NewFinanceHandler(newFinanceSvc(store))
	authHandler := httpapi.NewAuthHandler(a.AuthService)
	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, nil, nil, financeHandler, nil, nil, nil, nil, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/companies/me/cashflow/recent/", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestFinanceRecentCashflow_Success_200(t *testing.T) {
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	store := memory.New()
	a := app.New(cfg, store)
	financeHandler := httpapi.NewFinanceHandler(newFinanceSvc(store))
	authHandler := httpapi.NewAuthHandler(a.AuthService)
	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, nil, nil, financeHandler, nil, nil, nil, nil, nil, nil, nil)
	token := registerFinanceTestToken(t, mux, "fin1")

	req := httptest.NewRequest(http.MethodGet, "/api/v2/companies/me/cashflow/recent/", nil)
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
	if data["data"] == nil {
		t.Fatal("response missing 'data' field")
	}
	if data["money"] == nil {
		t.Fatal("response missing 'money' field")
	}
	if data["oldestPulled"] == nil {
		t.Fatal("response missing 'oldestPulled' field")
	}
}

func TestFinanceIncomeStatement_Success_200(t *testing.T) {
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	store := memory.New()
	a := app.New(cfg, store)
	financeHandler := httpapi.NewFinanceHandler(newFinanceSvc(store))
	authHandler := httpapi.NewAuthHandler(a.AuthService)
	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, nil, nil, financeHandler, nil, nil, nil, nil, nil, nil, nil)
	token := registerFinanceTestToken(t, mux, "fin2")

	req := httptest.NewRequest(http.MethodGet, "/api/v2/companies/me/income-statement/", nil)
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
	if data["revenue"] == nil {
		t.Fatal("response missing 'revenue' field")
	}
	if data["expenses"] == nil {
		t.Fatal("response missing 'expenses' field")
	}
	if data["netIncome"] == nil {
		t.Fatal("response missing 'netIncome' field")
	}
}

func TestFinanceBalanceSheet_Success_200(t *testing.T) {
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	store := memory.New()
	a := app.New(cfg, store)
	financeHandler := httpapi.NewFinanceHandler(newFinanceSvc(store))
	authHandler := httpapi.NewAuthHandler(a.AuthService)
	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, nil, nil, financeHandler, nil, nil, nil, nil, nil, nil, nil)
	token := registerFinanceTestToken(t, mux, "fin3")

	req := httptest.NewRequest(http.MethodGet, "/api/v2/companies/me/balance-sheet/", nil)
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
	if data["assets"] == nil {
		t.Fatal("response missing 'assets' field")
	}
	if data["equity"] == nil {
		t.Fatal("response missing 'equity' field")
	}
}

func TestFinanceCashflowStatement_Success_200(t *testing.T) {
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	store := memory.New()
	a := app.New(cfg, store)
	financeHandler := httpapi.NewFinanceHandler(newFinanceSvc(store))
	authHandler := httpapi.NewAuthHandler(a.AuthService)
	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, nil, nil, financeHandler, nil, nil, nil, nil, nil, nil, nil)
	token := registerFinanceTestToken(t, mux, "fin4")

	req := httptest.NewRequest(http.MethodGet, "/api/v2/companies/me/cashflow-statement/", nil)
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
	if data["operating"] == nil {
		t.Fatal("response missing 'operating' field")
	}
}

func TestFinancePastFinances_Success_200(t *testing.T) {
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	store := memory.New()
	a := app.New(cfg, store)
	financeHandler := httpapi.NewFinanceHandler(newFinanceSvc(store))
	authHandler := httpapi.NewAuthHandler(a.AuthService)
	mux := httpapi.NewRouter(cfg, authHandler, nil, nil, nil, nil, nil, financeHandler, nil, nil, nil, nil, nil, nil, nil)
	token := registerFinanceTestToken(t, mux, "fin5")

	req := httptest.NewRequest(http.MethodGet, "/api/v3/companies/me/past-finances/", nil)
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
	if data["series"] == nil {
		t.Fatal("response missing 'series' field")
	}
}
