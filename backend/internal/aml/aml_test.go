package aml

import (
	"fmt"
	"testing"
	"time"
)

func TestLargeTransactionFlagged(t *testing.T) {
	aml := New(true)
	aml.RecordTransaction(Transaction{
		ID: "t1", FromID: 1, ToID: 2,
		Amount: 100000, Type: "market_trade", Timestamp: time.Now(),
	})
	if len(aml.Flags()) != 1 {
		t.Errorf("expected 1 flag, got %d", len(aml.Flags()))
	}
	if aml.Flags()[0].Severity != "medium" {
		t.Errorf("expected medium severity, got %s", aml.Flags()[0].Severity)
	}
}

func TestLargeTransactionNotFlaggedBelowThreshold(t *testing.T) {
	aml := New(true)
	aml.RecordTransaction(Transaction{
		ID: "t1", FromID: 1, ToID: 2,
		Amount: 10000, Type: "market_trade", Timestamp: time.Now(),
	})
	if len(aml.Flags()) != 0 {
		t.Errorf("expected 0 flags for amount below threshold, got %d", len(aml.Flags()))
	}
}

func TestRapidTradeDetection(t *testing.T) {
	aml := New(true)
	now := time.Now()
	for i := 0; i < 5; i++ {
		aml.RecordTransaction(Transaction{
			ID:        fmt.Sprintf("t%d", i),
			FromID:    1,
			ToID:      2,
			Amount:    1000,
			Type:      "market_trade",
			Timestamp: now.Add(time.Duration(i) * time.Second),
		})
	}
	ok, msg := aml.CheckRapidTrades(1)
	if ok {
		t.Errorf("expected rapid trade rejection, got ok: %s", msg)
	}
}

func TestRapidTradeOKBelowLimit(t *testing.T) {
	aml := New(true)
	now := time.Now()
	for i := 0; i < 3; i++ {
		aml.RecordTransaction(Transaction{
			ID:        fmt.Sprintf("t%d", i),
			FromID:    1,
			ToID:      2,
			Amount:    1000,
			Type:      "market_trade",
			Timestamp: now.Add(time.Duration(i) * time.Second),
		})
	}
	ok, msg := aml.CheckRapidTrades(1)
	if !ok {
		t.Errorf("expected rapid trade ok (3 trades within threshold), got fail: %s", msg)
	}
}

func TestRoundTripNotSelf(t *testing.T) {
	aml := New(true)
	detected := aml.DetectRoundTrip(1, 2, 50000, 10)
	if detected {
		t.Errorf("expected no round trip detection for different players")
	}
}

func TestRoundTripDetection(t *testing.T) {
	aml := New(true)
	now := time.Now()
	aml.RecordTransaction(Transaction{
		ID: "t1", FromID: 1, ToID: 2,
		Amount: 50000, ResourceID: 10, Type: "market_trade",
		Timestamp: now,
	})
	aml.RecordTransaction(Transaction{
		ID: "t2", FromID: 3, ToID: 1,
		Amount: 52000, ResourceID: 10, Type: "market_trade",
		Timestamp: now.Add(30 * time.Second),
	})
	detected := aml.DetectRoundTrip(1, 1, 52000, 10)
	if !detected {
		t.Errorf("expected round trip detection for same player trading same resource")
	}
}

func TestFlagsSnapshot(t *testing.T) {
	aml := New(true)
	aml.RecordTransaction(Transaction{
		ID: "t1", FromID: 1, ToID: 2,
		Amount: 100000, Type: "market_trade", Timestamp: time.Now(),
	})
	flags1 := aml.Flags()
	flags2 := aml.Flags()
	flags1[0].Detail = "tampered"
	if aml.Flags()[0].Detail == "tampered" {
		t.Errorf("Flags() should return a copy")
	}
	if len(flags1) != len(flags2) {
		t.Errorf("snapshot length mismatch")
	}
}

func TestClear(t *testing.T) {
	aml := New(true)
	aml.RecordTransaction(Transaction{
		ID: "t1", FromID: 1, ToID: 2,
		Amount: 100000, Type: "market_trade", Timestamp: time.Now(),
	})
	aml.Clear()
	if len(aml.Flags()) != 0 {
		t.Errorf("expected 0 flags after clear")
	}
	ok, _ := aml.CheckRapidTrades(1)
	if !ok {
		t.Errorf("expected ok after clear")
	}
}
