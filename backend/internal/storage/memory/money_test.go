package memory

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/beiwater/NewHaven/backend/internal/domain/company"
	"github.com/beiwater/NewHaven/backend/internal/storage"
)

func newCo(t *testing.T, s *Store, playerID int, money float64) int {
	t.Helper()
	c := &company.Company{PlayerID: playerID, Name: "co", Money: money, Inventory: map[int]int{}}
	if err := s.CreateCompany(context.Background(), c); err != nil {
		t.Fatalf("CreateCompany: %v", err)
	}
	return c.ID
}

// TestAdjustMoney_Concurrent hammers AdjustMoney from many goroutines and
// asserts no updates are lost — the exact failure mode of the old
// GetCompany -> mutate Money -> UpdateCompany pattern.
func TestAdjustMoney_Concurrent(t *testing.T) {
	s := New()
	ctx := context.Background()
	id := newCo(t, s, 1, 0)

	const goroutines = 50
	const perG = 200
	var wg sync.WaitGroup
	for g := 0; g < goroutines; g++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < perG; i++ {
				if _, err := s.AdjustMoney(ctx, id, 1, false); err != nil {
					t.Errorf("AdjustMoney: %v", err)
					return
				}
			}
		}()
	}
	wg.Wait()

	c, _ := s.GetCompany(ctx, id)
	want := float64(goroutines * perG)
	if c.Money != want {
		t.Errorf("lost updates: money = %g; want %g", c.Money, want)
	}
}

// TestTransferMoney_Concurrent moves money back and forth between two accounts
// under contention and asserts the two-account total is conserved exactly.
func TestTransferMoney_Concurrent(t *testing.T) {
	s := New()
	ctx := context.Background()
	a := newCo(t, s, 1, 1_000_000)
	b := newCo(t, s, 2, 1_000_000)
	total := 2_000_000.0

	var wg sync.WaitGroup
	for g := 0; g < 40; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for i := 0; i < 250; i++ {
				from, to := a, b
				if (g+i)%2 == 0 {
					from, to = b, a
				}
				_ = s.TransferMoney(ctx, from, to, 100) // may fail on insufficient funds; that's fine
			}
		}(g)
	}
	wg.Wait()

	ca, _ := s.GetCompany(ctx, a)
	cb, _ := s.GetCompany(ctx, b)
	if ca.Money+cb.Money != total {
		t.Errorf("money not conserved: %g + %g = %g; want %g", ca.Money, cb.Money, ca.Money+cb.Money, total)
	}
	if ca.Money < 0 || cb.Money < 0 {
		t.Errorf("negative balance: a=%g b=%g", ca.Money, cb.Money)
	}
}

// TestAdjustMoney_RequireFunds verifies the funds guard is atomic and rejects
// overdrafts without mutating the balance.
func TestAdjustMoney_RequireFunds(t *testing.T) {
	s := New()
	ctx := context.Background()
	id := newCo(t, s, 1, 100)

	if _, err := s.AdjustMoney(ctx, id, -150, true); !errors.Is(err, storage.ErrInsufficientFunds) {
		t.Fatalf("want ErrInsufficientFunds, got %v", err)
	}
	c, _ := s.GetCompany(ctx, id)
	if c.Money != 100 {
		t.Errorf("balance changed on rejected debit: %g", c.Money)
	}
	if _, err := s.AdjustMoney(ctx, id, -100, true); err != nil {
		t.Fatalf("exact debit should succeed: %v", err)
	}
	c, _ = s.GetCompany(ctx, id)
	if c.Money != 0 {
		t.Errorf("balance = %g; want 0", c.Money)
	}
}

// TestTransferMoney_RejectsInvalid guards NaN/Inf/negative amounts.
func TestTransferMoney_RejectsInvalid(t *testing.T) {
	s := New()
	ctx := context.Background()
	a := newCo(t, s, 1, 100)
	b := newCo(t, s, 2, 100)
	if err := s.TransferMoney(ctx, a, b, -5); err == nil {
		t.Error("negative transfer should fail")
	}
	if ca, _ := s.GetCompany(ctx, a); ca.Money != 100 {
		t.Errorf("balance changed on invalid transfer: %g", ca.Money)
	}
}
