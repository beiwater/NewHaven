package auth_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"
	"time"

	"github.com/newhaven/backend-next/internal/app/auth"
	domain "github.com/newhaven/backend-next/internal/domain/auth"
	companydomain "github.com/newhaven/backend-next/internal/domain/company"
	"github.com/newhaven/backend-next/internal/platform"
	"github.com/newhaven/backend-next/internal/storage"
	"github.com/newhaven/backend-next/internal/storage/memory"
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
