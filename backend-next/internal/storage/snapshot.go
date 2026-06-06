package storage

import (
	"github.com/newhaven/backend-next/internal/domain/auth"
	"github.com/newhaven/backend-next/internal/domain/company"
	"github.com/newhaven/backend-next/internal/domain/finance"
	"github.com/newhaven/backend-next/internal/domain/market"
	"github.com/newhaven/backend-next/internal/domain/production"
	"github.com/newhaven/backend-next/internal/domain/research"
	"github.com/newhaven/backend-next/internal/domain/social"
	"github.com/newhaven/backend-next/internal/domain/warehouse"
)

// GameSnapshot captures the complete game state for persistence.
type GameSnapshot struct {
	Players     map[int]*auth.Player                `json:"players"`
	ByUser      map[string]*auth.Player              `json:"by_user"`
	Companies   map[int]*company.Company             `json:"companies"`
	Orders      map[string]*market.MarketOrder       `json:"orders"`
	Trades      []market.Trade                       `json:"trades"`
	Tickers     map[int]*market.Ticker               `json:"tickers"`
	Jobs        map[string]*production.ProductionJob `json:"jobs"`
	Ledger      []finance.LedgerEntry                `json:"ledger"`
	Bonds       map[string]*finance.Bond             `json:"bonds"`
	Holdings    []finance.BondHolding                `json:"holdings"`
	CompanyResearch map[string]research.ResourceResearch `json:"companyResearch"` // key: "companyID:resourceID"
	Messages    []social.Message                     `json:"messages"`
	Notifs      []social.Notification                `json:"notifications"`
	Warehouses  map[int]*warehouse.Warehouse         `json:"warehouses"`
	NextID      int                                  `json:"nextId"`
}
