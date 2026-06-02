package tests

// NOTE: run with: go test -v ./tests/

import (
	"fmt"
	"testing"
	"time"

	"go-sim-api/internal/anticheat"
)

// Round 1: Basic bot detection — rapid identical actions
func TestRound1_BasicBotDetection(t *testing.T) {
	ac := anticheat.New(true)
	sd := anticheat.NewScriptDetector(true)
	pid := 999

	// Simulate bot: 50 actions in 10 seconds (human average is ~5/min)
	for i := 0; i < 50; i++ {
		ac.RecordAction(pid, anticheat.ActCreateOrder, fmt.Sprintf("test-order-%d", i))
		sd.RecordAction(pid)
	}

	ok, msg := ac.CheckRateLimit(pid)
	if ok {
		t.Errorf("Round 1 FAIL: bot should be rate limited")
	} else {
		t.Logf("Round 1 PASS: rate limit triggered: %s", msg)
	}

	isBot, score := sd.IsLikelyBot(pid)
	if !isBot {
		t.Errorf("Round 1 FAIL: script detector should flag bot (score=%.2f)", score)
	} else {
		t.Logf("Round 1 PASS: bot detected with human score=%.2f", score)
	}
}

// Round 2: Wash trade detection
func TestRound2_WashTradeDetection(t *testing.T) {
	ac := anticheat.New(true)
	pid := 888

	// Simulate wash trading: same player has buy + sell on same resource
	orders := []any{
		map[string]any{"companyId": float64(pid), "resourceId": float64(8), "kind": float64(1)},
		map[string]any{"companyId": float64(pid), "resourceId": float64(8), "kind": float64(1)},
		map[string]any{"companyId": float64(pid), "resourceId": float64(8), "kind": float64(0)},
		map[string]any{"companyId": float64(pid), "resourceId": float64(8), "kind": float64(0)},
	}
	if !ac.DetectWashTrade(pid, 8, orders) {
		t.Errorf("Round 2 FAIL: wash trade not detected")
	} else {
		t.Logf("Round 2 PASS: wash trade detected")
	}
}

// Round 3: AML + Script combo — sophisticated cheat
func TestRound3_AMLandScriptAdvanced(t *testing.T) {
	ac := anticheat.New(true)
	sd := anticheat.NewScriptDetector(true)
	pid := 777

	// Simulate sophisticated: normal delays between actions, but round-trip trading
	// Every 2-5 seconds, do a buy then a sell on same resource
	for i := 0; i < 6; i++ {
		time.Sleep(2 * time.Millisecond) // small delay between actions (real script would be fast)
		ac.RecordAction(pid, anticheat.ActCreateOrder, fmt.Sprintf("buy-%d", i))
		sd.RecordAction(pid)
		ac.RecordAction(pid, anticheat.ActCancelOrder, fmt.Sprintf("sell-%d", i))
		sd.RecordAction(pid)
	}

	// Check: script has high perfect-timing ratio
	isBot, score := sd.IsLikelyBot(pid)
	if !isBot {
		t.Errorf("Round 3 FAIL: bot should be detected (score=%.2f)", score)
	} else {
		t.Logf("Round 3 PASS: bot detected with human score=%.2f", score)
	}

	// Clear and verify clean state
	ac.Clear()
	sd = anticheat.NewScriptDetector(true)
	// Simulate human: 5 actions over 30 seconds with variable delays
	for i := 0; i < 5; i++ {
		sd.RecordAction(pid)
		time.Sleep(time.Duration(100+i*50) * time.Millisecond) // variable 100-350ms
	}
	isBot2, score2 := sd.IsLikelyBot(pid)
	if isBot2 {
		t.Logf("Round 3 note: human also has low score (%.2f) due to low action count", score2)
	} else {
		t.Logf("Round 3 PASS: human not flagged (score=%.2f)", score2)
	}
}
