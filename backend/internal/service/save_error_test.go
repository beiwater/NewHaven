package service

import (
	"context"
	"errors"
	"testing"

	"go-sim-api/internal/config"
	"go-sim-api/internal/data"
	"go-sim-api/internal/model"
	"go-sim-api/internal/storage"
)

// errStorage is a fake storage that returns a fixed error from every Save method.
type errStorage struct {
	storage.NoopStorage
	err error
}

func (e *errStorage) SaveState(_ context.Context, _ *model.GameState) error {
	return e.err
}

func (e *errStorage) SaveCompany(_ context.Context, _ *model.Company) error {
	return e.err
}

func (e *errStorage) SaveOrders(_ context.Context, _ []model.MarketOrder) error {
	return e.err
}

func (e *errStorage) SaveTrades(_ context.Context, _ []model.Trade) error {
	return e.err
}

// TestSaveStateLockedRecordsError verifies that when storage.SaveState fails,
// the error is recorded in LastSaveError.
func TestSaveStateLockedRecordsError(t *testing.T) {
	fakeErr := errors.New("postgres connection refused")
	store := &errStorage{err: fakeErr}
	s := New(&data.StaticData{Resources: []map[string]any{}, EconomyModel: map[string]any{"models": map[string]any{}}}, config.DefaultTestConfig(), store)

	// Call saveStateLocked — error should be recorded
	s.saveStateLocked()
	if s.LastSaveError() == nil {
		t.Fatal("expected LastSaveError to be set after failed save")
	}
	if !errors.Is(s.LastSaveError(), fakeErr) {
		t.Errorf("LastSaveError = %v, want %v", s.LastSaveError(), fakeErr)
	}
}

// TestSaveCompanyLockedRecordsError verifies Company save errors.
func TestSaveCompanyLockedRecordsError(t *testing.T) {
	fakeErr := errors.New("timeout writing company")
	store := &errStorage{err: fakeErr}
	s := New(&data.StaticData{Resources: []map[string]any{}, EconomyModel: map[string]any{"models": map[string]any{}}}, config.DefaultTestConfig(), store)

	s.saveCompanyLocked(&model.Company{ID: 1})
	if s.LastSaveError() == nil {
		t.Fatal("expected LastSaveError after failed company save")
	}
	if !errors.Is(s.LastSaveError(), fakeErr) {
		t.Errorf("LastSaveError = %v, want %v", s.LastSaveError(), fakeErr)
	}
}

// TestSaveOrdersLockedRecordsError verifies Orders save errors.
func TestSaveOrdersLockedRecordsError(t *testing.T) {
	fakeErr := errors.New("orders table locked")
	store := &errStorage{err: fakeErr}
	s := New(&data.StaticData{Resources: []map[string]any{}, EconomyModel: map[string]any{"models": map[string]any{}}}, config.DefaultTestConfig(), store)

	s.saveOrdersLocked()
	if s.LastSaveError() == nil {
		t.Fatal("expected LastSaveError after failed orders save")
	}
	if !errors.Is(s.LastSaveError(), fakeErr) {
		t.Errorf("LastSaveError = %v, want %v", s.LastSaveError(), fakeErr)
	}
}

// TestSaveTradesLockedRecordsError verifies Trades save errors.
func TestSaveTradesLockedRecordsError(t *testing.T) {
	fakeErr := errors.New("trades insert failed")
	store := &errStorage{err: fakeErr}
	s := New(&data.StaticData{Resources: []map[string]any{}, EconomyModel: map[string]any{"models": map[string]any{}}}, config.DefaultTestConfig(), store)

	s.saveTradesLocked()
	if s.LastSaveError() == nil {
		t.Fatal("expected LastSaveError after failed trades save")
	}
	if !errors.Is(s.LastSaveError(), fakeErr) {
		t.Errorf("LastSaveError = %v, want %v", s.LastSaveError(), fakeErr)
	}
}

// TestSaveSuccessClearsError verifies that a successful save clears LastSaveError.
func TestSaveSuccessClearsError(t *testing.T) {
	store := &storage.NoopStorage{}
	s := New(&data.StaticData{Resources: []map[string]any{}, EconomyModel: map[string]any{"models": map[string]any{}}}, config.DefaultTestConfig(), store)

	// First set an error
	s.setSaveError(errors.New("previous error"))
	if s.LastSaveError() == nil {
		t.Fatal("expected error to be set before save")
	}

	// Successful save should clear it
	s.saveStateLocked()
	if s.LastSaveError() != nil {
		t.Errorf("LastSaveError should be nil after successful save, got %v", s.LastSaveError())
	}
}

// TestSaveStateLockedNilStoreDoesNotSetError verifies that when Store is nil,
// LastSaveError remains unchanged.
func TestSaveStateLockedNilStoreDoesNotSetError(t *testing.T) {
	s := New(&data.StaticData{Resources: []map[string]any{}, EconomyModel: map[string]any{"models": map[string]any{}}}, config.DefaultTestConfig(), nil)

	// Save on nil store should not touch LastSaveError
	s.setSaveError(nil)
	s.saveStateLocked()
	if s.LastSaveError() != nil {
		t.Errorf("expected nil LastSaveError for nil store, got %v", s.LastSaveError())
	}
}
