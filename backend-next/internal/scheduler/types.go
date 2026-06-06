package scheduler

import "context"

// BotService defines the game services the scheduler needs to run the
// bot-driven economy and background maintenance tasks.
type BotService interface {
	// RunBotCycle generates NPC buy/sell orders to maintain market liquidity.
	RunBotCycle(ctx context.Context) error
	// MatchAllOrders attempts to cross open buy/sell orders across all resources.
	MatchAllOrders(ctx context.Context) error
	// RefreshDailyOrders generates a new set of daily supply orders.
	RefreshDailyOrders(ctx context.Context) error
	// CleanupMarket removes stale/filled orders and refreshes daily orders.
	CleanupMarket(ctx context.Context) error
}
