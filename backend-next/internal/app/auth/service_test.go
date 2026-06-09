package auth_test

import (
	"context"
	"log/slog"
	"testing"
	"time"

	"github.com/newhaven/backend-next/internal/app/auth"
	domain "github.com/newhaven/backend-next/internal/domain/auth"
	"github.com/newhaven/backend-next/internal/platform"
	"github.com/newhaven/backend-next/internal/storage/memory"
)

func TestRegister_Success(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	clock := platform.NewFakeClock(time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC))
	idgen := platform.NewIDGen()
	logger := platform.NewLogger(slog.Default())

	svc := auth.NewService(store, store, clock, idgen, logger, "test-secret", "")

	resp, err := svc.Register(ctx, &domain.RegisterRequest{
		Username: "alice",
		Password: "secret123",
		Name:     "Alice",
		Gender:   "female",
		Email:    "alice@example.com",
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
	if resp.PlayerID <= 0 {
		t.Errorf("expected positive player_id, got %d", resp.PlayerID)
	}
	if resp.CompanyID <= 0 {
		t.Errorf("expected positive company_id, got %d", resp.CompanyID)
	}
	if resp.Username != "alice" {
		t.Errorf("expected username 'alice', got %q", resp.Username)
	}
}

func TestRegister_DuplicateUsername(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	clock := platform.NewFakeClock(time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC))
	idgen := platform.NewIDGen()
	logger := platform.NewLogger(slog.Default())

	svc := auth.NewService(store, store, clock, idgen, logger, "test-secret", "")

	_, err := svc.Register(ctx, &domain.RegisterRequest{
		Username: "alice",
		Password: "secret123",
	})
	if err != nil {
		t.Fatalf("first Register failed: %v", err)
	}

	_, err = svc.Register(ctx, &domain.RegisterRequest{
		Username: "alice",
		Password: "anotherPass",
	})
	if err == nil {
		t.Fatal("expected error on duplicate username, got nil")
	}
	if err != auth.ErrUsernameTaken {
		t.Fatalf("expected ErrUsernameTaken, got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	clock := platform.NewFakeClock(time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC))
	idgen := platform.NewIDGen()
	logger := platform.NewLogger(slog.Default())

	svc := auth.NewService(store, store, clock, idgen, logger, "test-secret", "")

	_, err := svc.Register(ctx, &domain.RegisterRequest{
		Username: "bob",
		Password: "hunter2",
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	resp, err := svc.Login(ctx, &domain.LoginRequest{
		Username: "bob",
		Password: "hunter2",
	})
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}

	if resp.Token == "" {
		t.Error("expected non-empty token")
	}
	if resp.Username != "bob" {
		t.Errorf("expected username 'bob', got %q", resp.Username)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	clock := platform.NewFakeClock(time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC))
	idgen := platform.NewIDGen()
	logger := platform.NewLogger(slog.Default())

	svc := auth.NewService(store, store, clock, idgen, logger, "test-secret", "")

	_, err := svc.Register(ctx, &domain.RegisterRequest{
		Username: "carol",
		Password: "correctpw",
	})
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	_, err = svc.Login(ctx, &domain.LoginRequest{
		Username: "carol",
		Password: "wrongpw",
	})
	if err == nil {
		t.Fatal("expected error on wrong password, got nil")
	}
	if err != auth.ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestDevBootstrap(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	clock := platform.NewFakeClock(time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC))
	idgen := platform.NewIDGen()
	logger := platform.NewLogger(slog.Default())

	svc := auth.NewService(store, store, clock, idgen, logger, "test-secret", "")

	// First call should create the dev user.
	if err := svc.DevBootstrap(ctx); err != nil {
		t.Fatalf("first DevBootstrap failed: %v", err)
	}

	// Verify dev user exists.
	player, err := store.GetPlayerByUsername(ctx, "dev")
	if err != nil {
		t.Fatalf("dev player not found after bootstrap: %v", err)
	}
	if player.Username != "dev" {
		t.Errorf("expected username 'dev', got %q", player.Username)
	}

	// Second call should be a no-op.
	if err := svc.DevBootstrap(ctx); err != nil {
		t.Fatalf("second DevBootstrap failed: %v", err)
	}
}
