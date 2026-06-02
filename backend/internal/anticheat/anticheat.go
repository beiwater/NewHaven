package anticheat

import (
	"fmt"
	"sync"
	"time"
)

type ActionType string

const (
	ActCreateOrder ActionType = "create_order"
	ActCancelOrder ActionType = "cancel_order"
	ActTakeOrder   ActionType = "take_order"
	ActClaimProd   ActionType = "claim_production"
	ActStartProd   ActionType = "start_production"
	ActGovBid      ActionType = "gov_bid"
	ActBondIssue   ActionType = "bond_issue"
	ActAuctionBid  ActionType = "auction_bid"
)

type Alert struct {
	ID        string    `json:"id"`
	PlayerID  int       `json:"playerId"`
	Type      string    `json:"type"`
	Detail    string    `json:"detail"`
	Severity  string    `json:"severity"` // low|medium|high|critical
	Timestamp time.Time `json:"timestamp"`
}

type ActionRecord struct {
	PlayerID  int
	Action    ActionType
	Detail    string
	Timestamp time.Time
}

type AntiCheat struct {
	mu      sync.Mutex
	actions map[int][]ActionRecord // playerID → action history
	alerts  []Alert
	nextID  int
	enabled bool

	// Config
	MaxActionsPerMinute int
	WashTradeWindow     time.Duration
	MinCancelDelay      time.Duration
}

func New(enabled bool) *AntiCheat {
	return &AntiCheat{
		actions:             map[int][]ActionRecord{},
		alerts:              []Alert{},
		enabled:             enabled,
		MaxActionsPerMinute: 30, // max 30 API calls/minute
		WashTradeWindow:     5 * time.Minute,
		MinCancelDelay:      2 * time.Second, // minimum time between order and cancel
	}
}

func (ac *AntiCheat) RecordAction(playerID int, action ActionType, detail string) {
	if !ac.enabled {
		return
	}
	ac.mu.Lock()
	defer ac.mu.Unlock()
	now := time.Now()
	ac.actions[playerID] = append(ac.actions[playerID], ActionRecord{
		PlayerID: playerID, Action: action, Detail: detail, Timestamp: now,
	})
	// Trim old records (keep last 5 minutes)
	cutoff := now.Add(-5 * time.Minute)
	var recent []ActionRecord
	for _, r := range ac.actions[playerID] {
		if r.Timestamp.After(cutoff) {
			recent = append(recent, r)
		}
	}
	ac.actions[playerID] = recent
}

func (ac *AntiCheat) CheckRateLimit(playerID int) (bool, string) {
	if !ac.enabled {
		return true, ""
	}
	ac.mu.Lock()
	defer ac.mu.Unlock()
	now := time.Now()
	cutoff := now.Add(-1 * time.Minute)
	count := 0
	for _, r := range ac.actions[playerID] {
		if r.Timestamp.After(cutoff) {
			count++
		}
	}
	if count > ac.MaxActionsPerMinute {
		return false, fmt.Sprintf("rate limit: %d actions in 1 min (max %d)", count, ac.MaxActionsPerMinute)
	}
	return true, ""
}

func (ac *AntiCheat) CheckQuickCancel(playerID int, action ActionType) (bool, string) {
	if !ac.enabled {
		return true, ""
	}
	if action != ActCancelOrder {
		return true, ""
	}
	ac.mu.Lock()
	defer ac.mu.Unlock()
	now := time.Now()
	for _, r := range ac.actions[playerID] {
		if r.Action == ActCreateOrder && now.Sub(r.Timestamp) < ac.MinCancelDelay {
			return false, fmt.Sprintf("order cancelled too quickly (%.0fs < min %.0fs)", now.Sub(r.Timestamp).Seconds(), ac.MinCancelDelay.Seconds())
		}
	}
	return true, ""
}

func (ac *AntiCheat) AddAlert(playerID int, atype, detail, severity string) {
	if !ac.enabled {
		return
	}
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.nextID++
	ac.alerts = append(ac.alerts, Alert{
		ID:       fmt.Sprintf("alert-%d", ac.nextID),
		PlayerID: playerID, Type: atype, Detail: detail,
		Severity: severity, Timestamp: time.Now(),
	})
}

func (ac *AntiCheat) Alerts() []Alert {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	out := make([]Alert, len(ac.alerts))
	copy(out, ac.alerts)
	return out
}

func (ac *AntiCheat) DetectWashTrade(playerID int, resourceID int, orders []any) bool {
	if !ac.enabled {
		return false
	}
	// Wash trading: same player buying AND selling same resource at similar prices
	ac.mu.Lock()
	defer ac.mu.Unlock()
	buyCount := 0
	sellCount := 0
	for _, o := range orders {
		if m, ok := o.(map[string]any); ok {
			cid := 0
			if v, ok := m["companyId"].(float64); ok {
				cid = int(v)
			}
			kind := 0
			if v, ok := m["kind"].(float64); ok {
				kind = int(v)
			}
			rid := 0
			if v, ok := m["resourceId"].(float64); ok {
				rid = int(v)
			}
			if cid == playerID && rid == resourceID {
				if kind == 1 {
					buyCount++
				} else {
					sellCount++
				}
			}
		}
	}
	if buyCount >= 2 && sellCount >= 2 {
		ac.nextID++
		ac.alerts = append(ac.alerts, Alert{
			ID:       fmt.Sprintf("alert-%d", ac.nextID),
			PlayerID: playerID, Type: "wash_trade",
			Detail:   fmt.Sprintf("player %d has %d buy and %d sell orders on resource %d", playerID, buyCount, sellCount, resourceID),
			Severity: "medium", Timestamp: time.Now(),
		})
		return true
	}
	return false
}

func (ac *AntiCheat) Clear() {
	ac.mu.Lock()
	defer ac.mu.Unlock()
	ac.actions = map[int][]ActionRecord{}
	ac.alerts = nil
}
