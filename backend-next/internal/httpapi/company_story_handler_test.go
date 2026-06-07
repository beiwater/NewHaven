package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/newhaven/backend-next/internal/app"
	"github.com/newhaven/backend-next/internal/app/company"
	"github.com/newhaven/backend-next/internal/config"
	"github.com/newhaven/backend-next/internal/httpapi"
	"github.com/newhaven/backend-next/internal/storage/memory"
)

func TestCompanyProfileAndStoryProgressRoutes(t *testing.T) {
	cfg := &config.Config{JWTSigningKey: "test-secret"}
	store := memory.New()
	a := app.New(cfg, store, nil, nil, nil)
	companyHandler := httpapi.NewCompanyHandler(company.NewService(store, a.Logger, 0), a.MarketService)
	mux := httpapi.NewRouter(cfg, &httpapi.RouterHandlers{Auth: httpapi.NewAuthHandler(a.AuthService), Company: companyHandler})

	reg := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(`{"username":"storyuser","password":"secret123"}`))
	mux.ServeHTTP(reg, req)
	var registered apiResponse
	if err := json.Unmarshal(reg.Body.Bytes(), &registered); err != nil {
		t.Fatal(err)
	}
	var authData struct {
		Token     string `json:"token"`
		CompanyID int    `json:"company_id"`
	}
	if err := json.Unmarshal(registered.Data, &authData); err != nil {
		t.Fatal(err)
	}

	profile := httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v3/companies/"+strconv.Itoa(authData.CompanyID)+"/", nil)
	req.Header.Set("Authorization", "Bearer "+authData.Token)
	mux.ServeHTTP(profile, req)
	if profile.Code != http.StatusOK {
		t.Fatalf("profile returned %d: %s", profile.Code, profile.Body.String())
	}

	update := httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodPatch, "/api/v2/companies/me/story-progress/", strings.NewReader(
		`{"storyId":"chapter1Arrival","stepId":"intro-bells","status":"in_progress"}`,
	))
	req.Header.Set("Authorization", "Bearer "+authData.Token)
	mux.ServeHTTP(update, req)
	if update.Code != http.StatusOK {
		t.Fatalf("story update returned %d: %s", update.Code, update.Body.String())
	}
}
