package company_test

import (
	"context"
	"errors"
	"log/slog"
	"testing"

	"github.com/beiwater/NewHaven/backend/internal/app/company"
	"github.com/beiwater/NewHaven/backend/internal/domain/auth"
	domain "github.com/beiwater/NewHaven/backend/internal/domain/company"
	"github.com/beiwater/NewHaven/backend/internal/platform"
	"github.com/beiwater/NewHaven/backend/internal/storage/memory"
)

func TestListMyCompanies_ReturnsCompany(t *testing.T) {
	ctx := context.Background()
	store := memory.New()

	err := store.CreatePlayer(ctx, &auth.Player{ID: 1, Username: "alice"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}
	err = store.CreateCompany(ctx, &domain.Company{
		PlayerID: 1,
		Name:     "Alice Co",
		Money:    1000.50,
		Level:    5,
	})
	if err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}

	logger := platform.NewLogger(slog.Default())
	svc := company.NewService(store, logger, 0)

	resp, err := svc.ListMyCompanies(ctx, 1)
	if err != nil {
		t.Fatalf("ListMyCompanies failed: %v", err)
	}

	if resp.Companies == nil {
		t.Fatal("expected non-nil companies")
	}
	if len(*resp.Companies) != 1 {
		t.Fatalf("expected 1 company, got %d", len(*resp.Companies))
	}

	got := (*resp.Companies)[0]
	if got.Id == nil || *got.Id <= 0 {
		t.Errorf("expected positive id, got %v", got.Id)
	}
	if got.Name == nil || *got.Name != "Alice Co" {
		t.Errorf("expected name 'Alice Co', got %v", got.Name)
	}
	if got.Money == nil || *got.Money != 1000.5 {
		t.Errorf("expected money 1000.5, got %v", got.Money)
	}
	if got.Level == nil || *got.Level != 5 {
		t.Errorf("expected level 5, got %v", got.Level)
	}
}

func TestListMyCompanies_NoCompany(t *testing.T) {
	ctx := context.Background()
	store := memory.New()

	err := store.CreatePlayer(ctx, &auth.Player{ID: 2, Username: "bob"})
	if err != nil {
		t.Fatalf("CreatePlayer: %v", err)
	}

	logger := platform.NewLogger(slog.Default())
	svc := company.NewService(store, logger, 0)

	_, err = svc.ListMyCompanies(ctx, 2)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, company.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}
