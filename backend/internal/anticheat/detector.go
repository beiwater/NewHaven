package anticheat

import (
	"math"
	"sync"
	"time"
)

// ScriptProfile stores timing data for a single player to evaluate
// whether their actions follow human-like patterns.
type ScriptProfile struct {
	PlayerID      int
	HumanScore    float64 // 0=bot, 1=human
	ActionCount   int
	PerfectTiming int // actions with <50ms delay between consecutive actions
	AvgDelayMs    float64
	LastAction    time.Time
}

// ScriptDetector monitors action timing across players and computes
// a human-likeness score based on delay variance.
type ScriptDetector struct {
	mu       sync.Mutex
	profiles map[int]*ScriptProfile
	enabled  bool
}

// NewScriptDetector returns an initialized ScriptDetector.
func NewScriptDetector(enabled bool) *ScriptDetector {
	return &ScriptDetector{enabled: enabled, profiles: map[int]*ScriptProfile{}}
}

// RecordAction records a timed action for the given player and updates
// their human-likeness score. Consecutive actions with <50ms delay are
// counted as "perfect timing" and reduce the human score.
func (sd *ScriptDetector) RecordAction(playerID int) {
	if !sd.enabled {
		return
	}
	sd.mu.Lock()
	defer sd.mu.Unlock()
	now := time.Now()
	p, ok := sd.profiles[playerID]
	if !ok {
		p = &ScriptProfile{PlayerID: playerID, HumanScore: 1.0}
		sd.profiles[playerID] = p
	}
	if !p.LastAction.IsZero() {
		delay := now.Sub(p.LastAction).Milliseconds()
		p.AvgDelayMs = (p.AvgDelayMs*float64(p.ActionCount) + float64(delay)) / float64(p.ActionCount+1)
		if delay < 50 { // < 50ms between actions → suspicious
			p.PerfectTiming++
		}
	}
	p.ActionCount++
	p.LastAction = now

	// Calculate human score
	// Humans have variable delays, bots have consistent fast delays
	if p.ActionCount > 10 {
		ratio := float64(p.PerfectTiming) / float64(p.ActionCount)
		p.HumanScore = 1.0 - math.Min(ratio*3, 1.0)
	}
}

// IsLikelyBot returns whether the player is likely a bot based on their
// human score. Returns false if insufficient data (<10 actions).
func (sd *ScriptDetector) IsLikelyBot(playerID int) (bool, float64) {
	if !sd.enabled {
		return false, 1.0
	}
	sd.mu.Lock()
	defer sd.mu.Unlock()
	p, ok := sd.profiles[playerID]
	if !ok || p.ActionCount < 10 {
		return false, 1.0 // not enough data
	}
	return p.HumanScore < 0.3, p.HumanScore
}
