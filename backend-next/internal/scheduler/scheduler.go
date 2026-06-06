package scheduler

import (
	"context"
	"log/slog"
	"time"
)

// Scheduler runs background economy maintenance tasks on a ticker.
type Scheduler struct {
	svc         BotService
	settleBonds func(ctx context.Context) error
	ticker      *time.Ticker
	done        chan struct{}
}

// New creates a Scheduler that ticks every 60 seconds.
// settleBonds is an optional function called on each tick to settle bond interest.
func New(svc BotService, settleBonds func(ctx context.Context) error) *Scheduler {
	return &Scheduler{
		svc:         svc,
		settleBonds: settleBonds,
		ticker:      time.NewTicker(60 * time.Second),
		done:        make(chan struct{}),
	}
}

// Start begins the background tick loop in a goroutine.
func (s *Scheduler) Start() {
	slog.Info("[scheduler] started (tick every 60s)")
	go func() {
		for {
			select {
			case <-s.done:
				return
			case <-s.ticker.C:
				s.tick()
			}
		}
	}()
}

// Stop signals the tick loop to shut down and waits for it to finish.
func (s *Scheduler) Stop() {
	s.ticker.Stop()
	close(s.done)
}

func (s *Scheduler) tick() {
	ctx := context.Background()
	slog.Debug("[scheduler] tick start")

	// 1. Bill production jobs
	if err := s.svc.RefreshDailyOrders(ctx); err != nil {
		slog.Warn("[scheduler] RefreshDailyOrders", "error", err)
	}

	// 2. Bot market cycle – generate NPC orders for liquidity
	if err := s.svc.RunBotCycle(ctx); err != nil {
		slog.Warn("[scheduler] RunBotCycle", "error", err)
	}

	// 3. Match all open orders
	if err := s.svc.MatchAllOrders(ctx); err != nil {
		slog.Warn("[scheduler] MatchAllOrders", "error", err)
	}

	// 4. Cleanup stale orders
	if err := s.svc.CleanupMarket(ctx); err != nil {
		slog.Warn("[scheduler] CleanupMarket", "error", err)
	}

	// 5. Settle bond interest
	if s.settleBonds != nil {
		if err := s.settleBonds(ctx); err != nil {
			slog.Warn("[scheduler] settleBonds", "error", err)
		}
	}
	slog.Debug("[scheduler] tick end")
}
