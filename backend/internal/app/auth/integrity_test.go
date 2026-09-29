package auth_test

import (
	"context"
	"errors"
	"log/slog"
	"path/filepath"
	"testing"
	"time"

	"github.com/beiwater/NewHaven/backend/internal/app/auth"
	domain "github.com/beiwater/NewHaven/backend/internal/domain/auth"
	companydomain "github.com/beiwater/NewHaven/backend/internal/domain/company"
	"github.com/beiwater/NewHaven/backend/internal/platform"
	"github.com/beiwater/NewHaven/backend/internal/storage"
	"github.com/beiwater/NewHaven/backend/internal/storage/memory"
)

type failingCompanyStore struct {
	storage.CompanyStorage
}

func (s failingCompanyStore) CreateCompany(context.Context, *companydomain.Company) error {
	return errors.New("injected company failure")
}

func newAuthService(players storage.PlayerStorage, companies storage.CompanyStorage) *auth.Service {
	return auth.NewService(
		players,
		companies,
		nil,
		platform.NewFakeClock(time.Date(2026, 6, 6, 12, 0, 0, 0, time.UTC)),
		platform.NewIDGen(),
		platform.NewLogger(slog.Default()),
		"test-secret",
		"",
	)
}

func TestRegisterNormalizesUsernameAndLogin(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	svc := newAuthService(store, store)

	resp, err := svc.Register(ctx, &domain.RegisterRequest{Username: " Alice ", Password: "secret123"})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Username != "alice" {
		t.Fatalf("expected normalized username, got %q", resp.Username)
	}
	login, err := svc.Login(ctx, &domain.LoginRequest{Username: "ALICE", Password: "secret123"})
	if err != nil {
		t.Fatalf("case-insensitive login failed: %v", err)
	}
	if login.Username != "alice" {
		t.Fatalf("login response returned non-canonical username %q", login.Username)
	}
	if _, err := svc.Register(ctx, &domain.RegisterRequest{Username: "ALICE", Password: "another123"}); !errors.Is(err, auth.ErrUsernameTaken) {
		t.Fatalf("expected normalized duplicate rejection, got %v", err)
	}
}

func TestRegisterRejectsInvalidCredentials(t *testing.T) {
	store := memory.New()
	svc := newAuthService(store, store)
	for _, req := range []domain.RegisterRequest{
		{Username: "ab", Password: "secret123"},
		{Username: "space name", Password: "secret123"},
		{Username: "valid-name", Password: "short"},
		{Username: "valid-name", Password: "secret123", Email: "not-an-email"},
	} {
		if _, err := svc.Register(context.Background(), &req); err == nil {
			t.Fatalf("expected validation error for %+v", req)
		}
	}
}

func TestRegisterRollsBackPlayerWhenCompanyCreationFails(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	svc := newAuthService(store, failingCompanyStore{CompanyStorage: store})
	_, err := svc.Register(ctx, &domain.RegisterRequest{Username: "rollback", Password: "secret123"})
	if err == nil {
		t.Fatal("expected company creation failure")
	}
	if _, err := store.GetPlayerByUsername(ctx, "rollback"); err == nil {
		t.Fatal("player remained after failed company creation")
	}
}

func TestRegisterAfterSnapshotRestorePreservesExistingAccountCompany(t *testing.T) {
	// Given
	t.Setenv("SIM_API_SNAPSHOT_PATH", filepath.Join(t.TempDir(), "snapshot.json"))
	ctx := context.Background()
	store := memory.New()
	svc := newAuthService(store, store)
	original, err := svc.Register(ctx, &domain.RegisterRequest{Username: "original", Password: "secret123"})
	if err != nil {
		t.Fatal(err)
	}
	if err := store.SaveSnapshot(ctx); err != nil {
		t.Fatal(err)
	}

	restored := memory.New()
	if err := restored.LoadSnapshot(ctx); err != nil {
		t.Fatal(err)
	}
	restoredService := newAuthService(restored, restored)
	newcomer, err := restoredService.Register(ctx, &domain.RegisterRequest{Username: "newcomer", Password: "secret123"})
	if err != nil {
		t.Fatal(err)
	}
	if newcomer.CompanyID == original.CompanyID {
		t.Fatal("test requires distinct companies")
	}

	// When
	login, err := restoredService.Login(ctx, &domain.LoginRequest{Username: "original", Password: "secret123"})
	if err != nil {
		t.Fatal(err)
	}

	// Then
	if login.CompanyID != original.CompanyID {
		t.Fatalf("original account returned to company %d after snapshot restore, want %d", login.CompanyID, original.CompanyID)
	}
}
