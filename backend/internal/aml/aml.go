package aml

import (
	"fmt"
	"sync"
	"time"
)

type Transaction struct {
	ID         string    `json:"id"`
	FromID     int       `json:"fromId"`
	ToID       int       `json:"toId"`
	Amount     float64   `json:"amount"`
	ResourceID int       `json:"resourceId"`
	Type       string    `json:"type"` // market_trade, bond_payment, gov_contract, auction
	Timestamp  time.Time `json:"timestamp"`
}

type AMLFlag struct {
	ID        string    `json:"id"`
	PlayerID  int       `json:"playerId"`
	Type      string    `json:"type"`
	Severity  string    `json:"severity"`
	Detail    string    `json:"detail"`
	Amount    float64   `json:"amount"`
	Timestamp time.Time `json:"timestamp"`
}

type AMLSystem struct {
	mu           sync.Mutex
	transactions []Transaction
	flags        []AMLFlag
	nextID       int
	thresholds   map[string]float64
	enabled      bool
}

func New(enabled bool) *AMLSystem {
	return &AMLSystem{
		enabled: enabled,
		thresholds: map[string]float64{
			"large_trade":    50000, // trades over this amount flagged
			"rapid_sequence": 3,     // more than 3 trades in 60s flagged
			"round_trip":     10000, // round-trip self-trade amount flag
		},
	}
}

func (a *AMLSystem) RecordTransaction(t Transaction) {
	if !a.enabled {
		return
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.transactions = append(a.transactions, t)
	if len(a.transactions) > 10000 {
		a.transactions = a.transactions[len(a.transactions)-5000:]
	}
	// Check large transactions
	if t.Amount > a.thresholds["large_trade"] {
		a.nextID++
		a.flags = append(a.flags, AMLFlag{
			ID:        fmt.Sprintf("aml-%d", a.nextID),
			PlayerID:  t.FromID,
			Type:      "large_transaction",
			Severity:  "medium",
			Amount:    t.Amount,
			Detail:    fmt.Sprintf("large %s of %.2f (threshold %.0f)", t.Type, t.Amount, a.thresholds["large_trade"]),
			Timestamp: time.Now(),
		})
	}
}

func (a *AMLSystem) CheckRapidTrades(playerID int) (bool, string) {
	if !a.enabled {
		return true, ""
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-60 * time.Second)
	count := 0
	for _, t := range a.transactions {
		if (t.FromID == playerID || t.ToID == playerID) && t.Timestamp.After(cutoff) {
			count++
		}
	}
	if count > int(a.thresholds["rapid_sequence"]) {
		return false, fmt.Sprintf("rapid trades: %d in 60s (threshold %.0f)", count, a.thresholds["rapid_sequence"])
	}
	return true, ""
}

func (a *AMLSystem) DetectRoundTrip(fromID, toID int, amount float64, resourceID int) bool {
	if !a.enabled {
		return false
	}
	if fromID != toID {
		return false
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-10 * time.Minute)
	match := 0
	for _, t := range a.transactions {
		if t.Timestamp.Before(cutoff) {
			continue
		}
		// Same player bought AND sold same resource recently
		if t.ResourceID == resourceID && (t.FromID == fromID || t.ToID == fromID) {
			match++
		}
	}
	if match >= 2 {
		a.nextID++
		a.flags = append(a.flags, AMLFlag{
			ID:        fmt.Sprintf("aml-%d", a.nextID),
			PlayerID:  fromID,
			Type:      "round_trip",
			Severity:  "high",
			Amount:    amount,
			Detail:    fmt.Sprintf("player %d round-trip trading resource %d (%d trades in 10min)", fromID, resourceID, match),
			Timestamp: time.Now(),
		})
		return true
	}
	return false
}

func (a *AMLSystem) Flags() []AMLFlag {
	a.mu.Lock()
	defer a.mu.Unlock()
	out := make([]AMLFlag, len(a.flags))
	copy(out, a.flags)
	return out
}

func (a *AMLSystem) Clear() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.transactions = nil
	a.flags = nil
}
