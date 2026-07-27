package storage

import (
	"github.com/beiwater/NewHaven/backend/internal/domain/auth"
	"github.com/beiwater/NewHaven/backend/internal/domain/chat"
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
	NextID              int                                   `json:"nextId"` // next company ID

	// Private chat state. Previously absent from the snapshot entirely, so every
	// direct-message room and its history was silently discarded on restart.
	ChatRooms         map[string]*chat.ChatRoom `json:"chatRooms,omitempty"`
	ChatMessages      map[string][]chat.Message `json:"chatMessages,omitempty"`
	ChatReadAt        map[string]int64          `json:"chatReadAt,omitempty"`
	NextChatMessageID int64                     `json:"nextChatMessageId,omitempty"`

	// Remaining ID counters. Without these, a restored store restarted every
	// counter at 1 and immediately reissued IDs that existing rows already held
	// — a new registration would take over player 1's identity, and ledger,
	// message, and notification IDs would collide. Omitted in snapshots written
	// before this field existed, so restore derives them from the data instead.
	NextPlayerID  int   `json:"nextPlayerId,omitempty"`
	NextLedgerID  int64 `json:"nextLedgerId,omitempty"`
	NextMessageID int64 `json:"nextMessageId,omitempty"`
	NextNotifID   int   `json:"nextNotifId,omitempty"`
}
