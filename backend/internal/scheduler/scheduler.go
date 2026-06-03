package scheduler

import (
	"context"
	"go-sim-api/internal/model"
	"log"
	"time"
)

type GameService interface {
	SettleAllBonds() map[string]any
	CheckMarketLock(resourceID int)
	ResourcesWithMarket() []int
	ResetDailyMarket()
	RunBotMarketCycle()
	AwardGovernmentContracts() []model.GovContract
	ResolveGovernmentDefaults() []model.GovContract
	CleanupOrders()
	SaveAll()
	RefreshDailyOrders()
	RunAllProductionJobs()
}

type Scheduler struct {
	svc    GameService
	ticker *time.Ticker
	ctx    context.Context
	cancel context.CancelFunc
}

func New(svc GameService) *Scheduler {
	ctx, cancel := context.WithCancel(context.Background())
	return &Scheduler{
		svc:    svc,
		ticker: time.NewTicker(60 * time.Second),
		ctx:    ctx,
		cancel: cancel,
	}
}

func (s *Scheduler) Start() {
	log.Println("[scheduler] started (tick every 60s)")
	go func() {
		for {
			select {
			case <-s.ctx.Done():
				return
			case <-s.ticker.C:
				s.tick()
			}
		}
	}()
}

func (s *Scheduler) Stop() {
	s.ticker.Stop()
	s.cancel()
}

func (s *Scheduler) tick() {
	log.Println("[scheduler] tick start")

	// 1. Bond interest settlement
	result := s.svc.SettleAllBonds()
	if defaults, ok := result["defaults"].([]any); ok && len(defaults) > 0 {
		log.Printf("[scheduler] bond defaults: %d", len(defaults))
	}

	// 2. Government contract awards
	s.svc.AwardGovernmentContracts()

	// 3. Government default resolution
	s.svc.ResolveGovernmentDefaults()

	// 4. Bot market cycle
	s.svc.RunBotMarketCycle()
	// 5. Check market locks and deploy national team if needed
	for _, rid := range s.svc.ResourcesWithMarket() {
		s.svc.CheckMarketLock(rid)
	}

	// 6. Order cleanup (remove zero-remaining orders)
	s.svc.CleanupOrders()

	// 7. Refresh production jobs
	s.svc.RunAllProductionJobs()

	// 8. Refresh daily orders
	s.svc.RefreshDailyOrders()

	// 9. Persistent save
	s.svc.SaveAll()

	log.Println("[scheduler] tick end")
}
