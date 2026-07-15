package httpapi_test

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beiwater/NewHaven/backend/internal/app"
	"github.com/beiwater/NewHaven/backend/internal/config"
	"github.com/beiwater/NewHaven/backend/internal/httpapi"
	"github.com/beiwater/NewHaven/backend/internal/storage/memory"
)

func TestRegisterRejectsInvalidAndUnknownFields(t *testing.T) {
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	store := memory.New()
	handler := httpapi.NewAuthHandler(app.New(cfg, store, nil, nil, nil).AuthService)
	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	for _, body := range []string{
		`{"username":"ab","password":"secret123"}`,
		`{"username":"valid-user","password":"short"}`,
		`{"username":"valid-user","password":"secret123","admin":true}`,
		`{"username":"valid-user","password":"secret123"} {}`,
	} {
		req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, req)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 for %s, got %d: %s", body, w.Code, w.Body.String())
		}
	}
}
