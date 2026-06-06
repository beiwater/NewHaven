package storage

import (
	"context"
	"errors"

	"github.com/newhaven/backend-next/internal/domain/auth"
	"github.com/newhaven/backend-next/internal/domain/company"
	"github.com/newhaven/backend-next/internal/domain/finance"
	"github.com/newhaven/backend-next/internal/domain/market"
	"github.com/newhaven/backend-next/internal/domain/production"
	"github.com/newhaven/backend-next/internal/domain/research"
	"github.com/newhaven/backend-next/internal/domain/social"
	"github.com/newhaven/backend-next/internal/domain/warehouse"
)

var ErrAlreadyExists = errors.New("already exists")

// PlayerStorage handles player/auth persistence.
type PlayerStorage interface {
	CreatePlayer(ctx context.Context, p *auth.Player) error
	DeletePlayer(ctx context.Context, id int) error
	GetPlayerByUsername(ctx context.Context, username string) (*auth.Player, error)
	GetPlayerByID(ctx context.Context, id int) (*auth.Player, error)
}

// CompanyStorage handles company persistence.
type CompanyStorage interface {
	CreateCompany(ctx context.Context, c *company.Company) error
	GetCompany(ctx context.Context, id int) (*company.Company, error)
	GetCompanyByPlayerID(ctx context.Context, playerID int) (*company.Company, error)
	GetAllCompanies(ctx context.Context) ([]*company.Company, error)
	UpdateCompany(ctx context.Context, c *company.Company) error
	SaveBuilding(ctx context.Context, b *company.Building) error
	RemoveBuilding(ctx context.Context, buildingID string) error
	GetBuildings(ctx context.Context, companyID int) ([]company.Building, error)
	UpdateInventory(ctx context.Context, companyID int, resourceID int, delta int) error
}

// MarketStorage handles order book and trade persistence.
type MarketStorage interface {
	CreateOrder(ctx context.Context, o *market.MarketOrder) error
	GetOrder(ctx context.Context, orderID string) (*market.MarketOrder, error)
	UpdateOrder(ctx context.Context, o *market.MarketOrder) error
	GetOpenOrders(ctx context.Context, resourceID int, isBuy bool) ([]market.MarketOrder, error)
	GetOrdersByCompany(ctx context.Context, companyID int) ([]market.MarketOrder, error)
	GetOrdersByResource(ctx context.Context, resourceID int) ([]market.MarketOrder, error)
	SaveTrade(ctx context.Context, t *market.Trade) error
	GetTrades(ctx context.Context, resourceID int, limit int) ([]market.Trade, error)
	GetTicker(ctx context.Context, resourceID int) (*market.Ticker, error)
	UpdateTicker(ctx context.Context, t *market.Ticker) error
	GetTickers(ctx context.Context) ([]market.Ticker, error)
}

// ProductionStorage handles production job persistence.
type ProductionStorage interface {
	CreateJob(ctx context.Context, j *production.ProductionJob) error
	GetJob(ctx context.Context, jobID string) (*production.ProductionJob, error)
	GetJobsByCompany(ctx context.Context, companyID int) ([]production.ProductionJob, error)
	GetJobsByBuilding(ctx context.Context, buildingID string) ([]production.ProductionJob, error)
	UpdateJob(ctx context.Context, j *production.ProductionJob) error
	DeleteJob(ctx context.Context, jobID string) error
}

// FinanceStorage handles ledger and bond persistence.
type FinanceStorage interface {
	AppendLedgerEntry(ctx context.Context, e *finance.LedgerEntry) error
	GetLedgerEntries(ctx context.Context, companyID int, limit int) ([]finance.LedgerEntry, error)
	CreateBond(ctx context.Context, b *finance.Bond) error
	GetBond(ctx context.Context, bondID string) (*finance.Bond, error)
	UpdateBond(ctx context.Context, b *finance.Bond) error
	GetActiveBonds(ctx context.Context) ([]finance.Bond, error)
	GetBondsByIssuer(ctx context.Context, companyID int) ([]finance.Bond, error)
	CreateBondHolding(ctx context.Context, h *finance.BondHolding) error
	GetBondHoldings(ctx context.Context, bondID string) ([]finance.BondHolding, error)
	GetCompanyBondHoldings(ctx context.Context, companyID int) ([]finance.BondHolding, error)
}

// ResearchStorage handles research persistence for per-resource levels.
type ResearchStorage interface {
	GetCompanyResearch(ctx context.Context, companyID int) ([]research.ResourceResearch, error)
	GetResourceResearch(ctx context.Context, companyID int, resourceID int) (*research.ResourceResearch, error)
	SaveResourceResearch(ctx context.Context, rr *research.ResourceResearch) error
}

// SocialStorage handles chat and notification persistence.
type SocialStorage interface {
	SaveMessage(ctx context.Context, m *social.Message) error
	GetMessages(ctx context.Context, channel string, limit int) ([]social.Message, error)
	CreateNotification(ctx context.Context, n *social.Notification) error
	GetNotifications(ctx context.Context, companyID int, limit int) ([]social.Notification, error)
	MarkNotificationRead(ctx context.Context, notificationID int) error
}

// WarehouseStorage handles warehouse persistence.
type WarehouseStorage interface {
	GetWarehouse(ctx context.Context, companyID int) (*warehouse.Warehouse, error)
	UpdateWarehouse(ctx context.Context, w *warehouse.Warehouse) error
}

// SnapshotStorage handles full game state save/load.
type SnapshotStorage interface {
	SaveSnapshot(ctx context.Context) error
	LoadSnapshot(ctx context.Context) error
}

// Storage combines all domain storage interfaces.
// Implementations can choose to implement all or compose from sub-stores.
type Storage interface {
	PlayerStorage
	CompanyStorage
	MarketStorage
	ProductionStorage
	FinanceStorage
	ResearchStorage
	SocialStorage
	WarehouseStorage
	SnapshotStorage
	Close() error
}
