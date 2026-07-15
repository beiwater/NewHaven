package httpapi_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/beiwater/NewHaven/backend/internal/app"
	"github.com/beiwater/NewHaven/backend/internal/config"
	domain "github.com/beiwater/NewHaven/backend/internal/domain/company"
	"github.com/beiwater/NewHaven/backend/internal/httpapi"
	"github.com/beiwater/NewHaven/backend/internal/storage/memory"
)

type isolationIdentity struct {
	Token     string
	PlayerID  int
	CompanyID int
}

func newIsolationRouter(t *testing.T) (*httpapi.RouterHandlers, http.Handler, *memory.Store) {
	t.Helper()
	cfg := &config.Config{JWTSigningKey: "isolation-test-secret"}
	store := memory.New()
	a := app.New(cfg, store, nil, nil, nil)
	handlers := &httpapi.RouterHandlers{
		Auth: a.AuthHandler, Company: a.CompanyHandler, Building: a.BuildingHandler,
		Player: a.PlayerHandler, Social: a.SocialHandler, Chat: a.ChatHandler,
		Admin: a.AdminHandler,
	}
	return handlers, httpapi.NewRouter(cfg, handlers), store
}

func registerIsolationPlayer(t *testing.T, mux http.Handler, username string) isolationIdentity {
	t.Helper()
	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(
		fmt.Sprintf(`{"username":%q,"password":"secret123"}`, username),
	))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("register %s: status=%d body=%s", username, w.Code, w.Body.String())
	}
	var envelope apiResponse
	if err := json.Unmarshal(w.Body.Bytes(), &envelope); err != nil {
		t.Fatalf("decode register response: %v", err)
	}
	var data struct {
		Token     string `json:"token"`
		PlayerID  int    `json:"player_id"`
		CompanyID int    `json:"company_id"`
	}
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		t.Fatalf("decode register data: %v", err)
	}
	return isolationIdentity(data)
}

func authorisedRequest(method, target, token, body string) *http.Request {
	request := httptest.NewRequest(method, target, strings.NewReader(body))
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")
	return request
}

func TestIsolation_AdminSnapshotRoutesAreNotExposed(t *testing.T) {
	_, mux, _ := newIsolationRouter(t)
	for _, path := range []string{"/api/admin/snapshot/save", "/api/admin/snapshot/load"} {
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodPost, path, nil))
		if w.Code != http.StatusNotFound {
			t.Fatalf("%s: expected 404, got %d; body=%s", path, w.Code, w.Body.String())
		}
	}
}

func TestIsolation_OtherCompanyBuildingsHideRetailState(t *testing.T) {
	_, mux, store := newIsolationRouter(t)
	alice := registerIsolationPlayer(t, mux, "isolation-alice")
	bob := registerIsolationPlayer(t, mux, "isolation-bob")

	company, err := store.GetCompany(context.Background(), bob.CompanyID)
	if err != nil {
		t.Fatal(err)
	}
	company.Buildings = []domain.Building{{
		ID: "bob-shop", BuildingID: 11, Kind: 11, Name: "Bob's Shop", Level: 3,
		MapID: "map-1", SlotID: "slot-1", X: 4, Y: 7, RobotCount: 9,
		Shelves: []domain.ShelfItem{{ResourceID: 101, Quantity: 88, Price: 999, Revenue: 123456}},
	}}
	if err := store.UpdateCompany(context.Background(), company); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, authorisedRequest(http.MethodGet,
		fmt.Sprintf("/api/v3/companies/%d/buildings/", bob.CompanyID), alice.Token, ""))
	if w.Code != http.StatusOK {
		t.Fatalf("expected safe public view, got %d; body=%s", w.Code, w.Body.String())
	}
	for _, privateField := range []string{"shelves", "revenue", "robot_count", "123456"} {
		if strings.Contains(w.Body.String(), privateField) {
			t.Fatalf("public building response leaked %q: %s", privateField, w.Body.String())
		}
	}
	if !strings.Contains(w.Body.String(), "Bob's Shop") {
		t.Fatalf("public building identity missing: %s", w.Body.String())
	}
}

func TestIsolation_PlayerCannotTargetAnotherPlayersPreferences(t *testing.T) {
	_, mux, _ := newIsolationRouter(t)
	alice := registerIsolationPlayer(t, mux, "prefs-alice")
	bob := registerIsolationPlayer(t, mux, "prefs-bob")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, authorisedRequest(http.MethodPost,
		fmt.Sprintf("/api/v2/players/%d/preferences/", bob.PlayerID), alice.Token, `{"music":false}`))
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for cross-player preference write, got %d; body=%s", w.Code, w.Body.String())
	}
}

func TestIsolation_LegacyPrivateChannelsAreDisabled(t *testing.T) {
	_, mux, _ := newIsolationRouter(t)
	alice := registerIsolationPlayer(t, mux, "legacy-chat-alice")
	bob := registerIsolationPlayer(t, mux, "legacy-chat-bob")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, authorisedRequest(http.MethodGet,
		fmt.Sprintf("/api/messages/?chatroom=C:%d", bob.CompanyID), alice.Token, ""))
	if w.Code != http.StatusGone {
		t.Fatalf("expected 410 for insecure legacy private channel, got %d; body=%s", w.Code, w.Body.String())
	}
}

func TestIsolation_ThirdCompanyCannotReadRoomMessages(t *testing.T) {
	_, mux, _ := newIsolationRouter(t)
	alice := registerIsolationPlayer(t, mux, "room-alice")
	bob := registerIsolationPlayer(t, mux, "room-bob")
	charlie := registerIsolationPlayer(t, mux, "room-charlie")

	w := httptest.NewRecorder()
	mux.ServeHTTP(w, authorisedRequest(http.MethodPost, "/api/v2/chat/room/", alice.Token,
		fmt.Sprintf(`{"companyId":%d}`, bob.CompanyID)))
	if w.Code != http.StatusOK {
		t.Fatalf("create room: status=%d body=%s", w.Code, w.Body.String())
	}
	var roomResponse struct {
		Room struct {
			ID string `json:"id"`
		} `json:"room"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &roomResponse); err != nil {
		t.Fatal(err)
	}

	w = httptest.NewRecorder()
	mux.ServeHTTP(w, authorisedRequest(http.MethodPost, "/api/v2/chat/room/send/", alice.Token,
		fmt.Sprintf(`{"roomId":%q,"body":"private merger plan"}`, roomResponse.Room.ID)))
	if w.Code != http.StatusOK {
		t.Fatalf("send room message: status=%d body=%s", w.Code, w.Body.String())
	}

	w = httptest.NewRecorder()
	mux.ServeHTTP(w, authorisedRequest(http.MethodGet,
		fmt.Sprintf("/api/v2/chat/room/%s/messages/", roomResponse.Room.ID), charlie.Token, ""))
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for third-party room read, got %d; body=%s", w.Code, w.Body.String())
	}
	if strings.Contains(w.Body.String(), "private merger plan") {
		t.Fatalf("private message leaked to third company: %s", w.Body.String())
	}
}
