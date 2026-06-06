package httpapi_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/newhaven/backend-next/internal/app"
	"github.com/newhaven/backend-next/internal/app/production"
	"github.com/newhaven/backend-next/internal/catalog"
	"github.com/newhaven/backend-next/internal/config"
	proddmn "github.com/newhaven/backend-next/internal/domain/production"
	"github.com/newhaven/backend-next/internal/httpapi"
	"github.com/newhaven/backend-next/internal/platform"
	"github.com/newhaven/backend-next/internal/storage/memory"
)

// newProductionSvc creates a Service with empty catalogs and fake clock,
// using the provided store (or creating a new one if nil).
func newProductionSvc(store *memory.Store) *production.Service {
	if store == nil {
		store = memory.New()
	}
	cfg := &config.GameConfig{ProductionMod: 1.0}
	clock := platform.NewFakeClock(time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC))
	idgen := platform.NewIDGen()
	return production.NewService(store, store, store, cfg,
		make(map[int]*catalog.ResourceEntry),
		make(map[int]*catalog.BuildingEntry),
		clock, idgen)
}

func TestListProductionJobs_NoToken_401(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store, nil, nil)
	productionSvc := newProductionSvc(store)
	productionHandler := httpapi.NewProductionHandler(productionSvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, &httpapi.RouterHandlers{Auth: authHandler, Production: productionHandler})

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
	a := app.New(cfg, store, nil, nil)
	productionSvc := newProductionSvc(store)
	productionHandler := httpapi.NewProductionHandler(productionSvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, &httpapi.RouterHandlers{Auth: authHandler, Production: productionHandler})

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
}

func TestListProductionJobs_EmptyJobsIsArray_200(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store, nil, nil)
	productionSvc := newProductionSvc(store)
	productionHandler := httpapi.NewProductionHandler(productionSvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, &httpapi.RouterHandlers{Auth: authHandler, Production: productionHandler})

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

	req := httptest.NewRequest(http.MethodGet, "/api/v2/production/jobs/", nil)
	req.Header.Set("Authorization", "Bearer "+token.(string))
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

// --- StartProduction handler tests ---

func TestStartProduction_NoToken_401(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store, nil, nil)
	productionSvc := newProductionSvc(store)
	productionHandler := httpapi.NewProductionHandler(productionSvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, &httpapi.RouterHandlers{Auth: authHandler, Production: productionHandler})

	body := `{"building_id":"bld-1","resource_id":3,"quantity":10}`
	req := httptest.NewRequest(http.MethodPost, "/api/v2/production/start/", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestStartProduction_InvalidBody_400(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store, nil, nil)
	productionSvc := newProductionSvc(store)
	productionHandler := httpapi.NewProductionHandler(productionSvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, &httpapi.RouterHandlers{Auth: authHandler, Production: productionHandler})

	// Register to get token
	regBody := `{"username":"startinv","password":"secret123"}`
	regReq := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	mux.ServeHTTP(regW, regReq)

	var regResp apiResponse
	if err := json.Unmarshal(regW.Body.Bytes(), &regResp); err != nil {
		t.Fatalf("unmarshal register: %v", err)
	}
	var regData map[string]any
	if err := json.Unmarshal(regResp.Data, &regData); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	token := regData["token"].(string)

	// Send invalid JSON body
	req := httptest.NewRequest(http.MethodPost, "/api/v2/production/start/", strings.NewReader(`{bad json`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestStartProduction_Success_200(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store, nil, nil)

	// Set up catalog with Mill producing Flour from Grain
	resources := map[int]*catalog.ResourceEntry{
		3: {ID: 3, Name: "Flour", ProducedPerHourRaw: 320, ProducedFrom: map[int]int{1: 2}},
	}
	buildings := map[int]*catalog.BuildingEntry{
		3: {ID: 3, Name: "Mill", Produces: []int{3}},
	}
	clock := platform.NewFakeClock(time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC))
	idgen := platform.NewIDGen()
	productionSvc := production.NewService(store, store, store, &config.GameConfig{ProductionMod: 1.0},
		resources, buildings, clock, idgen)
	productionHandler := httpapi.NewProductionHandler(productionSvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, &httpapi.RouterHandlers{Auth: authHandler, Production: productionHandler})

	// Register to get token
	regBody := `{"username":"startsuccess","password":"secret123"}`
	regReq := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	mux.ServeHTTP(regW, regReq)

	var regResp apiResponse
	if err := json.Unmarshal(regW.Body.Bytes(), &regResp); err != nil {
		t.Fatalf("unmarshal register: %v", err)
	}
	var regData map[string]any
	if err := json.Unmarshal(regResp.Data, &regData); err != nil {
		t.Fatalf("unmarshal register data: %v", err)
	}
	token := regData["token"].(string)
	companyID := int(regData["company_id"].(float64))

	// Add Grain to inventory for production
	err := store.UpdateInventory(nil, companyID, 1, 100)
	if err != nil {
		t.Fatalf("seed inventory: %v", err)
	}

	// Change first building to Mill (type 3) to match our catalog
	company, err := store.GetCompany(nil, companyID)
	if err != nil {
		t.Fatalf("get company: %v", err)
	}
	bldID := company.Buildings[0].ID
	company.Buildings[0].BuildingID = 3
	company.Buildings[0].Level = 1
	company.Buildings[0].Name = "My Mill"
	_ = store.UpdateCompany(nil, company)

	// Send start production request
	body := `{"building_id":"` + bldID + `","resource_id":3,"quantity":5}`
	req := httptest.NewRequest(http.MethodPost, "/api/v2/production/start/", strings.NewReader(body))
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
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", *resp.Error)
	}

	var data map[string]any
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}

	// Verify job in response
	jobObj, ok := data["job"]
	if !ok {
		t.Fatal("response data missing 'job' field")
	}
	job, ok := jobObj.(map[string]any)
	if !ok {
		t.Fatalf("expected job object, got %T", jobObj)
	}
	if job["id"] == nil || job["id"] == "" {
		t.Error("expected non-empty job id")
	}
	if job["status"] != "running" {
		t.Errorf("expected status 'running', got %v", job["status"])
	}

	// Verify building status
	bldObj, ok := data["building"]
	if !ok {
		t.Fatal("response data missing 'building' field")
	}
	bld, ok := bldObj.(map[string]any)
	if !ok {
		t.Fatalf("expected building object, got %T", bldObj)
	}
	if bld["busy"] != true {
		t.Errorf("expected building busy true, got %v", bld["busy"])
	}

	// Verify job is listed via GET /production/jobs/
	getReq := httptest.NewRequest(http.MethodGet, "/api/v2/production/jobs/", nil)
	getReq.Header.Set("Authorization", "Bearer "+token)
	getW := httptest.NewRecorder()
	mux.ServeHTTP(getW, getReq)

	if getW.Code != http.StatusOK {
		t.Fatalf("expected 200 for list, got %d; body: %s", getW.Code, getW.Body.String())
	}

	var listResp apiResponse
	if err := json.Unmarshal(getW.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("unmarshal list response: %v", err)
	}
	var listData map[string]any
	if err := json.Unmarshal(listResp.Data, &listData); err != nil {
		t.Fatalf("unmarshal list data: %v", err)
	}
	jobsArr := listData["jobs"].([]any)
	if len(jobsArr) != 1 {
		t.Fatalf("expected 1 job in list, got %d", len(jobsArr))
	}
}

// --- ClaimProduction handler tests ---

func TestClaimProduction_NoToken_401(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store, nil, nil)
	productionSvc := newProductionSvc(store)
	productionHandler := httpapi.NewProductionHandler(productionSvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, &httpapi.RouterHandlers{Auth: authHandler, Production: productionHandler})

	req := httptest.NewRequest(http.MethodPost, "/api/v2/production/claim/some-job/", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestClaimProduction_NoToken_Claimable(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store, nil, nil)
	productionSvc := newProductionSvc(store)
	productionHandler := httpapi.NewProductionHandler(productionSvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, &httpapi.RouterHandlers{Auth: authHandler, Production: productionHandler})

	req := httptest.NewRequest(http.MethodGet, "/api/v2/production/claimable/", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestClaimClaimable_Empty_200(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store, nil, nil)
	productionSvc := newProductionSvc(store)
	productionHandler := httpapi.NewProductionHandler(productionSvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, &httpapi.RouterHandlers{Auth: authHandler, Production: productionHandler})

	// Register user
	regBody := `{"username":"cl-empt","password":"secret123"}`
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
	token := regData["token"].(string)

	req := httptest.NewRequest(http.MethodGet, "/api/v2/production/claimable/", nil)
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

	jobsRaw, ok := data["jobs"]
	if !ok {
		t.Fatal("response data missing 'jobs' field")
	}
	var jobs []any
	if err := json.Unmarshal(jobsRaw, &jobs); err != nil {
		t.Fatalf("unmarshal jobs: %v", err)
	}
	if jobs == nil {
		t.Fatal("expected non-nil jobs array (should be empty)")
	}
	if len(jobs) != 0 {
		t.Errorf("expected 0 claimable jobs, got %d", len(jobs))
	}
}

func TestClaimProduction_InvalidJobId_400(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store, nil, nil)
	productionSvc := newProductionSvc(store)
	productionHandler := httpapi.NewProductionHandler(productionSvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, &httpapi.RouterHandlers{Auth: authHandler, Production: productionHandler})

	// Register to get token
	regBody := `{"username":"cl-inv","password":"secret123"}`
	regReq := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(regBody))
	regReq.Header.Set("Content-Type", "application/json")
	regW := httptest.NewRecorder()
	mux.ServeHTTP(regW, regReq)

	var regResp apiResponse
	if err := json.Unmarshal(regW.Body.Bytes(), &regResp); err != nil {
		t.Fatalf("unmarshal register: %v", err)
	}
	var regData map[string]any
	if err := json.Unmarshal(regResp.Data, &regData); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}
	token := regData["token"].(string)

	// Claim a non-existent job
	req := httptest.NewRequest(http.MethodPost, "/api/v2/production/claim/nonexistent/", nil)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for nonexistent job, got %d; body: %s", w.Code, w.Body.String())
	}
}

func TestClaimProduction_Success_200(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store, nil, nil)
	productionSvc := newProductionSvc(store)
	productionHandler := httpapi.NewProductionHandler(productionSvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, &httpapi.RouterHandlers{Auth: authHandler, Production: productionHandler})

	// Register user
	regBody := `{"username":"cl-succ","password":"secret123"}`
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
	token := regData["token"].(string)
	companyID := int(regData["company_id"].(float64))

	// Seed a completed job for this company
	startedAt := time.Date(2026, 6, 5, 10, 0, 0, 0, time.UTC)
	jobID := "handler-claim-job"
	err := store.CreateJob(nil, &proddmn.ProductionJob{
		ID:              jobID,
		CompanyID:       companyID,
		BuildingID:      "bld-1",
		ResourceID:      3,
		Quantity:        10,
		TargetQuantity:  10,
		StartedAt:       startedAt,
		DurationSeconds: 60.0,
		ClaimedAmount:   0,
		ClaimableAmount: 10,
		Status:          proddmn.StatusReady,
	})
	if err != nil {
		t.Fatalf("CreateJob: %v", err)
	}

	// POST to claim endpoint
	req := httptest.NewRequest(http.MethodPost, "/api/v2/production/claim/"+jobID+"/", nil)
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
	if resp.Error != nil {
		t.Fatalf("unexpected error: %+v", *resp.Error)
	}

	var data map[string]any
	if err := json.Unmarshal(resp.Data, &data); err != nil {
		t.Fatalf("unmarshal data: %v", err)
	}

	if data["job_id"] != jobID {
		t.Errorf("expected job_id %s, got %v", jobID, data["job_id"])
	}
	if data["status"] != "claimed" {
		t.Errorf("expected status 'claimed', got %v", data["status"])
	}
	if data["xp"] == nil || data["xp"].(float64) <= 0 {
		t.Errorf("expected positive xp, got %v", data["xp"])
	}
}

// --- ProductionQueue handler tests ---

func TestProductionQueue_NoToken_401(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store, nil, nil)
	productionSvc := newProductionSvc(store)
	productionHandler := httpapi.NewProductionHandler(productionSvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, &httpapi.RouterHandlers{Auth: authHandler, Production: productionHandler})

	req := httptest.NewRequest(http.MethodGet, "/api/v2/production/queue/", nil)
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

// --- ProductionOptions handler tests ---

func TestProductionOptions_NoToken_401(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store, nil, nil)
	productionSvc := newProductionSvc(store)
	productionHandler := httpapi.NewProductionHandler(productionSvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, &httpapi.RouterHandlers{Auth: authHandler, Production: productionHandler})

	req := httptest.NewRequest(http.MethodGet, "/api/v2/buildings/b1/production-options/", nil)
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

// --- CancelProduction handler tests ---

func TestCancelProduction_NoToken_401(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store, nil, nil)
	productionSvc := newProductionSvc(store)
	productionHandler := httpapi.NewProductionHandler(productionSvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, &httpapi.RouterHandlers{Auth: authHandler, Production: productionHandler})

	req := httptest.NewRequest(http.MethodPost, "/api/v2/production/cancel/", nil)
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

// --- ClaimAll handler tests ---

func TestClaimAll_NoToken_401(t *testing.T) {
	cfg := &config.Config{
		JWTSigningKey: "test-secret",
	}
	store := memory.New()
	a := app.New(cfg, store, nil, nil)
	productionSvc := newProductionSvc(store)
	productionHandler := httpapi.NewProductionHandler(productionSvc)
	authHandler := httpapi.NewAuthHandler(a.AuthService)

	mux := httpapi.NewRouter(cfg, &httpapi.RouterHandlers{Auth: authHandler, Production: productionHandler})

	req := httptest.NewRequest(http.MethodPost, "/api/v2/production/claim-all/", nil)
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
