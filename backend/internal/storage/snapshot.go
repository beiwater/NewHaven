package storage

import (
	"github.com/beiwater/NewHaven/backend/internal/domain/auth"
	"github.com/beiwater/NewHaven/backend/internal/domain/company"
	"github.com/beiwater/NewHaven/backend/internal/domain/finance"
	"github.com/beiwater/NewHaven/backend/internal/domain/market"
	"github.com/beiwater/NewHaven/backend/internal/domain/production"
	"github.com/beiwater/NewHaven/backend/internal/domain/research"
	"github.com/beiwater/NewHaven/backend/internal/domain/social"
	"github.com/beiwater/NewHaven/backend/internal/domain/warehouse"
)

// GameSnapshot captures the complete game state for persistence.
type GameSnapshot struct {
	Players             map[int]*auth.Player                  `json:"players"`
	ByUser              map[string]*auth.Player               `json:"by_user"`
	Companies           map[int]*company.Company              `json:"companies"`
	Orders              map[string]*market.MarketOrder        `json:"orders"`
	Trades              []market.Trade                        `json:"trades"`
	TakeOrderExecutions map[string]*market.TakeOrderExecution `json:"take_order_executions,omitempty"`
	Tickers             map[int]*market.Ticker                `json:"tickers"`
	Jobs                map[string]*production.ProductionJob  `json:"jobs"`
	Ledger              []finance.LedgerEntry                 `json:"ledger"`
	Bonds               map[string]*finance.Bond              `json:"bonds"`
	Holdings            []finance.BondHolding                 `json:"holdings"`
	CompanyResearch     map[string]research.ResourceResearch  `json:"companyResearch"` // key: "companyID:resourceID"
	Messages            []social.Message                      `json:"messages"`
	Notifs              []social.Notification                 `json:"notifications"`
	Warehouses          map[int]*warehouse.Warehouse          `json:"warehouses"`
	NextID              int                                   `json:"nextId"`
}

// NextAvailablePlayerID derives the first unused player ID from persisted
// accounts so loading an older snapshot cannot reuse an existing identity.
func (s *GameSnapshot) NextAvailablePlayerID() int {
	nextID := 1
	for playerID := range s.Players {
		nextID = max(nextID, playerID+1)
	}
	return nextID
}

// NextAvailableLedgerID derives the sequence cursor for append-only finance
// records. Snapshots intentionally remain backward compatible and need not
// persist a separate counter.
func (s *GameSnapshot) NextAvailableLedgerID() int64 {
	var nextID int64 = 1
	for _, entry := range s.Ledger {
		nextID = max(nextID, entry.ID+1)
	}
	return nextID
}

// NextAvailableMessageID prevents chat history identities from being reused
// after a snapshot restore.
func (s *GameSnapshot) NextAvailableMessageID() int64 {
	var nextID int64 = 1
	for _, message := range s.Messages {
		nextID = max(nextID, message.ID+1)
	}
	return nextID
}

// NextAvailableNotificationID prevents a newly created notification from
// colliding with a persisted notification after restart.
func (s *GameSnapshot) NextAvailableNotificationID() int {
	nextID := 1
	for _, notification := range s.Notifs {
		nextID = max(nextID, notification.ID+1)
	}
	return nextID
}
