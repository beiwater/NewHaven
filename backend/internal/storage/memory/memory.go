package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/beiwater/NewHaven/backend/internal/domain/auth"
	"github.com/beiwater/NewHaven/backend/internal/domain/chat"
	"github.com/beiwater/NewHaven/backend/internal/domain/company"
	"github.com/beiwater/NewHaven/backend/internal/domain/finance"
	"github.com/beiwater/NewHaven/backend/internal/domain/market"
	"github.com/beiwater/NewHaven/backend/internal/domain/production"
	"github.com/beiwater/NewHaven/backend/internal/domain/research"
	"github.com/beiwater/NewHaven/backend/internal/domain/social"
	"github.com/beiwater/NewHaven/backend/internal/domain/warehouse"
	"github.com/beiwater/NewHaven/backend/internal/storage"
)

// Store is an in-memory implementation of storage.Storage.
// Thread-safe via sync.RWMutex per domain.
// Data is NOT persisted across restarts.
type chatRoomData struct {
	sync.RWMutex
	rooms    map[string]*chat.ChatRoom
	messages map[string][]chat.Message
	nextID   int64
	readAt   map[string]int64
}

type Store struct {
	mu                  sync.RWMutex
	snapshotPath        string
	players             map[int]*auth.Player
	byUser              map[string]*auth.Player
	companies           map[int]*company.Company
	byPlayer            map[int]*company.Company
	orders              map[string]*market.MarketOrder
	ordersByRequest     map[string]*market.MarketOrder
	trades              []market.Trade
	takeOrderExecutions map[string]*market.TakeOrderExecution
	tickers             map[int]*market.Ticker
	jobs                map[string]*production.ProductionJob
	jobsByRequest       map[string]*production.ProductionJob
	ledger              []finance.LedgerEntry
	bonds               map[string]*finance.Bond
	holdings            []finance.BondHolding
	companyResearch     map[string]*research.ResourceResearch
	messages            []social.Message
	notifs              []social.Notification
	warehouses          map[int]*warehouse.Warehouse
	chatData            chatRoomData

	buildings map[string]*company.Building

	// Per-domain ID counters (independent, no longer shared)
	nextPlayerID  int
	nextCompanyID int
	nextLedgerID  int64
	nextMessageID int64
	nextNotifID   int

	// Per-domain locks for high-volume silos (ledger, messages, notifications)
	// These operate independently from s.mu, so Bot ledger writes don't block player registration.
	ledgerMu sync.Mutex
	msgMu    sync.Mutex
	notifMu  sync.Mutex
}

func New() *Store {
	return &Store{
		players:             make(map[int]*auth.Player),
		byUser:              make(map[string]*auth.Player),
		companies:           make(map[int]*company.Company),
		byPlayer:            make(map[int]*company.Company),
		orders:              make(map[string]*market.MarketOrder),
		ordersByRequest:     make(map[string]*market.MarketOrder),
		takeOrderExecutions: make(map[string]*market.TakeOrderExecution),
		tickers:             make(map[int]*market.Ticker),
		jobs:                make(map[string]*production.ProductionJob),
		jobsByRequest:       make(map[string]*production.ProductionJob),
		bonds:               make(map[string]*finance.Bond),
		buildings:           make(map[string]*company.Building),
		warehouses:          make(map[int]*warehouse.Warehouse),
		chatData: chatRoomData{
			rooms:    make(map[string]*chat.ChatRoom),
			messages: make(map[string][]chat.Message),
			nextID:   1,
			readAt:   make(map[string]int64),
		},
		companyResearch: make(map[string]*research.ResourceResearch),
		nextPlayerID:    1,
		nextCompanyID:   1,
		nextLedgerID:    1,
		nextMessageID:   1,
		nextNotifID:     1,
		snapshotPath:    envOrDefault("SIM_API_SNAPSHOT_PATH", "data/snapshot.json"),
	}
}

func (s *Store) Close() error { return nil }

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// snapshotPath implements file-based snapshot persistence.

func (s *Store) SaveSnapshot(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	snap := s.collectSnapshot()
	data, err := json.MarshalIndent(snap, "", "  ")
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(s.snapshotPath), 0755); err != nil {
		return err
	}
	return os.WriteFile(s.snapshotPath, data, 0644)
}

func (s *Store) LoadSnapshot(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	data, err := os.ReadFile(s.snapshotPath)
	if err != nil {
		return err
	}
	var snap storage.GameSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return err
	}
	s.applySnapshot(&snap)
	return nil
}

// GetSnapshotData returns a copy of the current game state as a GameSnapshot.
// The caller should hold s.mu (read or write).
func (s *Store) collectSnapshot() *storage.GameSnapshot {
	snap := &storage.GameSnapshot{
		Players:             make(map[int]*auth.Player, len(s.players)),
		ByUser:              make(map[string]*auth.Player, len(s.byUser)),
		Companies:           make(map[int]*company.Company, len(s.companies)),
		Orders:              make(map[string]*market.MarketOrder, len(s.orders)),
		TakeOrderExecutions: make(map[string]*market.TakeOrderExecution, len(s.takeOrderExecutions)),
		Tickers:             make(map[int]*market.Ticker, len(s.tickers)),
		Jobs:                make(map[string]*production.ProductionJob, len(s.jobs)),
		Bonds:               make(map[string]*finance.Bond, len(s.bonds)),
		Warehouses:          make(map[int]*warehouse.Warehouse, len(s.warehouses)),
	}
	for k, v := range s.players {
		snap.Players[k] = v
	}
	for k, v := range s.byUser {
		snap.ByUser[k] = v
	}
	for k, v := range s.companies {
		snap.Companies[k] = v
	}
	for k, v := range s.orders {
		snap.Orders[k] = v
	}
	for k, v := range s.takeOrderExecutions {
		snap.TakeOrderExecutions[k] = v
	}
	for k, v := range s.tickers {
		snap.Tickers[k] = v
	}
	for k, v := range s.jobs {
		snap.Jobs[k] = v
	}
	for k, v := range s.bonds {
		snap.Bonds[k] = v
	}
	for k, v := range s.warehouses {
		snap.Warehouses[k] = v
	}
	snap.Trades = append([]market.Trade(nil), s.trades...)
	snap.Ledger = append([]finance.LedgerEntry(nil), s.ledger...)
	snap.Holdings = append([]finance.BondHolding(nil), s.holdings...)
	if s.companyResearch != nil {
		rrCopy := make(map[string]research.ResourceResearch, len(s.companyResearch))
		for k, v := range s.companyResearch {
			rrCopy[k] = *v
		}
		snap.CompanyResearch = rrCopy
	}
	snap.Messages = append([]social.Message(nil), s.messages...)
	snap.Notifs = append([]social.Notification(nil), s.notifs...)
	snap.NextID = s.nextCompanyID
	return snap
}

// applySnapshot replaces all store state from a GameSnapshot.
// The caller must hold s.mu write lock.
func (s *Store) applySnapshot(snap *storage.GameSnapshot) {
	s.players = snap.Players
	if s.players == nil {
		s.players = make(map[int]*auth.Player)
	}
	s.byUser = snap.ByUser
	if s.byUser == nil {
		s.byUser = make(map[string]*auth.Player)
	}
	s.companies = snap.Companies
	if s.companies == nil {
		s.companies = make(map[int]*company.Company)
	}
	// rebuild byPlayer map from companies
	s.byPlayer = make(map[int]*company.Company, len(s.companies))
	for _, c := range s.companies {
		s.byPlayer[c.PlayerID] = c
	}
	s.orders = snap.Orders
	if s.orders == nil {
		s.orders = make(map[string]*market.MarketOrder)
	}
	s.ordersByRequest = make(map[string]*market.MarketOrder)
	for _, order := range s.orders {
		if order.ClientRequestID != "" {
			s.ordersByRequest[marketRequestKey(order.CompanyID, order.ClientRequestID)] = order
		}
	}
	if snap.Trades != nil {
		s.trades = snap.Trades
	} else {
		s.trades = nil
	}
	s.takeOrderExecutions = snap.TakeOrderExecutions
	if s.takeOrderExecutions == nil {
		s.takeOrderExecutions = make(map[string]*market.TakeOrderExecution)
	}
	s.tickers = snap.Tickers
	if s.tickers == nil {
		s.tickers = make(map[int]*market.Ticker)
	}
	s.jobs = snap.Jobs
	if s.jobs == nil {
		s.jobs = make(map[string]*production.ProductionJob)
	}
	s.jobsByRequest = make(map[string]*production.ProductionJob)
	for _, job := range s.jobs {
		if job.ClientRequestID != "" {
			s.jobsByRequest[productionRequestKey(job.CompanyID, job.ClientRequestID)] = job
		}
	}
	if snap.Ledger != nil {
		s.ledger = snap.Ledger
	} else {
		s.ledger = nil
	}
	s.bonds = snap.Bonds
	if s.bonds == nil {
		s.bonds = make(map[string]*finance.Bond)
	}
	if snap.Holdings != nil {
		s.holdings = snap.Holdings
	} else {
		s.holdings = nil
	}
	if snap.CompanyResearch != nil {
		s.companyResearch = make(map[string]*research.ResourceResearch, len(snap.CompanyResearch))
		for k, v := range snap.CompanyResearch {
			v := v
			s.companyResearch[k] = &v
		}
	} else {
		s.companyResearch = make(map[string]*research.ResourceResearch)
	}
	if snap.Messages != nil {
		s.messages = snap.Messages
	} else {
		s.messages = nil
	}
	if snap.Notifs != nil {
		s.notifs = snap.Notifs
	} else {
		s.notifs = nil
	}
	s.warehouses = snap.Warehouses
	if s.warehouses == nil {
		s.warehouses = make(map[int]*warehouse.Warehouse)
	}
	s.nextCompanyID = snap.NextID
	if s.nextCompanyID <= 0 {
		s.nextCompanyID = 1
	}
	// Legacy fallback: other counters start from a reasonable base
	s.nextPlayerID = 1
	s.nextLedgerID = 1
	s.nextMessageID = 1
	s.nextNotifID = 1
}

// GetSnapshotData returns a snapshot of the current game state for external callers (e.g. PG store).
func (s *Store) GetSnapshotData() *storage.GameSnapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.collectSnapshot()
}

// LoadFromSnapshot populates the store from a snapshot (e.g. loaded from PG).
func (s *Store) LoadFromSnapshot(snap *storage.GameSnapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.applySnapshot(snap)
}

// --- PlayerStorage ---
func (s *Store) CreatePlayer(_ context.Context, p *auth.Player) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p.Username = normalizeUsername(p.Username)
	if _, ok := s.byUser[p.Username]; ok {
		return fmt.Errorf("%w: username", storage.ErrAlreadyExists)
	}
	p.ID = s.nextPlayerID
	s.nextPlayerID++
	p.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	s.players[p.ID] = p
	s.byUser[p.Username] = p
	return nil
}

func (s *Store) DeletePlayer(_ context.Context, id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.players[id]
	if !ok {
		return fmt.Errorf("player not found")
	}
	delete(s.players, id)
	delete(s.byUser, p.Username)
	return nil
}

func (s *Store) GetPlayerByUsername(_ context.Context, username string) (*auth.Player, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.byUser[normalizeUsername(username)]
	if !ok {
		return nil, fmt.Errorf("player not found")
	}
	return p, nil
}

func normalizeUsername(username string) string {
	return strings.ToLower(strings.TrimSpace(username))
}

func (s *Store) GetPlayerByID(_ context.Context, id int) (*auth.Player, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.players[id]
	if !ok {
		return nil, fmt.Errorf("player not found")
	}
	return p, nil
}

// --- CompanyStorage ---
func (s *Store) CreateCompany(_ context.Context, c *company.Company) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	c.ID = s.nextCompanyID + 1000000
	s.nextCompanyID++
	c.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	s.companies[c.ID] = c
	s.byPlayer[c.PlayerID] = c
	// Auto-create default warehouse for new company
	s.warehouses[c.ID] = &warehouse.Warehouse{
		CompanyID:    c.ID,
		Capacity:     1000,
		UsedCapacity: 0,
		Items:        []warehouse.Item{},
	}
	// Initialize empty buildings slice for new company (only if not already set)
	if c.Buildings == nil {
		c.Buildings = []company.Building{}
	}
	return nil
}

func (s *Store) GetCompany(_ context.Context, id int) (*company.Company, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.companies[id]
	if !ok {
		return nil, fmt.Errorf("company not found")
	}
	return c, nil
}

func (s *Store) GetCompanyByPlayerID(_ context.Context, playerID int) (*company.Company, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.byPlayer[playerID]
	if !ok {
		return nil, fmt.Errorf("company not found")
	}
	return c, nil
}

func (s *Store) GetAllCompanies(_ context.Context) ([]*company.Company, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*company.Company, 0, len(s.companies))
	for _, c := range s.companies {
		result = append(result, c)
	}
	return result, nil
}

// PurchaseBuilding atomically applies a building purchase and its money
// deduction. The request ID is scoped to a company, so retries cannot create
// duplicate buildings while unrelated players remain independent.
func (s *Store) PurchaseBuilding(_ context.Context, companyID int, building company.Building, cost float64, maxBuildings int) (*company.Building, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.companies[companyID]
	if !ok {
		return nil, false, fmt.Errorf("company %d not found", companyID)
	}
	if building.PurchaseRequestID != "" {
		for i := range c.Buildings {
			existing := &c.Buildings[i]
			if existing.PurchaseRequestID != building.PurchaseRequestID {
				continue
			}
			if existing.PurchaseCatalogItemID != building.PurchaseCatalogItemID {
				return nil, false, storage.ErrIdempotencyConflict
			}
			copy := *existing
			return &copy, true, nil
		}
	}
	if len(c.Buildings) >= maxBuildings {
		return nil, false, storage.ErrLimitReached
	}
	if c.Money < cost {
		return nil, false, storage.ErrInsufficientFunds
	}
	c.Money -= cost
	c.Buildings = append(c.Buildings, building)
	created := c.Buildings[len(c.Buildings)-1]
	return &created, false, nil
}

func (s *Store) UpdateCompany(_ context.Context, c *company.Company) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.companies[c.ID] = c
	return nil
}

// StartBuildingUpgrade reserves the upgrade cost and records construction in
// one critical section. A second tab or server instance therefore cannot
// charge the same player twice or begin two upgrades for one building.
func (s *Store) StartBuildingUpgrade(_ context.Context, companyID int, buildingID string, expectedLevel, targetLevel int, cost float64, startedAt, completesAt string) (*company.Building, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.companies[companyID]
	if !ok {
		return nil, fmt.Errorf("company %d not found", companyID)
	}
	for i := range c.Buildings {
		b := &c.Buildings[i]
		if b.ID != buildingID {
			continue
		}
		if b.Level != expectedLevel || b.UpgradeTargetLevel != 0 || b.UpgradeCompletesAt != "" {
			return nil, storage.ErrStateConflict
		}
		if c.Money < cost {
			return nil, storage.ErrInsufficientFunds
		}
		c.Money -= cost
		b.UpgradeTargetLevel = targetLevel
		b.UpgradeStartedAt = startedAt
		b.UpgradeCompletesAt = completesAt
		copy := *b
		return &copy, nil
	}
	return nil, fmt.Errorf("building %s not found", buildingID)
}

// CompleteBuildingUpgrade only applies the requested construction once. The
// expected completion timestamp prevents a stale poll from completing a newer
// upgrade after a snapshot restore or later construction cycle.
func (s *Store) CompleteBuildingUpgrade(_ context.Context, companyID int, buildingID string, expectedTargetLevel int, expectedCompletesAt string) (*company.Building, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.companies[companyID]
	if !ok {
		return nil, false, fmt.Errorf("company %d not found", companyID)
	}
	for i := range c.Buildings {
		b := &c.Buildings[i]
		if b.ID != buildingID {
			continue
		}
		if b.UpgradeTargetLevel != expectedTargetLevel || b.UpgradeCompletesAt != expectedCompletesAt {
			return nil, false, nil
		}
		b.Level = expectedTargetLevel
		b.UpgradeTargetLevel = 0
		b.UpgradeStartedAt = ""
		b.UpgradeCompletesAt = ""
		copy := *b
		return &copy, true, nil
	}
	return nil, false, fmt.Errorf("building %s not found", buildingID)
}

func (s *Store) SaveBuilding(_ context.Context, b *company.Building) error { return nil }
func (s *Store) RemoveBuilding(_ context.Context, buildingID string) error { return nil }
func (s *Store) UpdateInventory(_ context.Context, companyID int, resourceID int, delta int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.updateInventoryLocked(companyID, resourceID, delta)
}

func (s *Store) updateInventoryLocked(companyID int, resourceID int, delta int) error {
	c, ok := s.companies[companyID]
	if !ok {
		return fmt.Errorf("company %d not found", companyID)
	}
	w, ok := s.warehouses[companyID]
	if !ok {
		return fmt.Errorf("warehouse for company %d not found", companyID)
	}
	cur := 0
	if c.Inventory != nil {
		cur = c.Inventory[resourceID]
	}
	newVal := cur + delta
	if newVal < 0 {
		return fmt.Errorf("insufficient inventory: resource %d has %d, need %d more", resourceID, cur, -delta)
	}
	if c.Inventory == nil {
		c.Inventory = make(map[int]int)
	}
	c.Inventory[resourceID] = newVal

	// Keep warehouse items consistent for quality 0.
	found := false
	for i := range w.Items {
		if w.Items[i].ResourceID == resourceID && w.Items[i].Quality == 0 {
			if newVal > 0 {
				// Preserve existing ResourceName if present.
				w.Items[i].Amount = newVal
			} else {
				// Remove zero-amount item.
				w.Items = append(w.Items[:i], w.Items[i+1:]...)
			}
			found = true
			break
		}
	}
	if !found && newVal > 0 {
		w.Items = append(w.Items, warehouse.Item{
			ResourceID:   resourceID,
			ResourceName: "",
			Quality:      0,
			Amount:       newVal,
		})
	}

	// Recompute UsedCapacity as the sum of item amounts.
	total := 0
	for _, item := range w.Items {
		total += item.Amount
	}
	w.UsedCapacity = total

	return nil
}

func (s *Store) GetBuildings(_ context.Context, companyID int) ([]company.Building, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	c, ok := s.companies[companyID]
	if !ok {
		return nil, fmt.Errorf("company %d not found", companyID)
	}
	return c.Buildings, nil
}

// --- WarehouseStorage ---

func (s *Store) GetWarehouse(_ context.Context, companyID int) (*warehouse.Warehouse, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	w, ok := s.warehouses[companyID]
	if !ok {
		return nil, fmt.Errorf("warehouse not found")
	}
	return w, nil
}

func (s *Store) UpdateWarehouse(_ context.Context, w *warehouse.Warehouse) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.warehouses == nil {
		s.warehouses = make(map[int]*warehouse.Warehouse)
	}
	s.warehouses[w.CompanyID] = w
	return nil
}

// --- MarketStorage ---

func (s *Store) CreateOrder(_ context.Context, o *market.MarketOrder) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if o.ClientRequestID != "" {
		key := marketRequestKey(o.CompanyID, o.ClientRequestID)
		if _, exists := s.ordersByRequest[key]; exists {
			return fmt.Errorf("%w: market request", storage.ErrAlreadyExists)
		}
		s.ordersByRequest[key] = o
	}
	s.orders[o.ID] = o
	return nil
}

func (s *Store) GetOrder(_ context.Context, orderID string) (*market.MarketOrder, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	o, ok := s.orders[orderID]
	if !ok {
		return nil, fmt.Errorf("order not found")
	}
	return o, nil
}

func (s *Store) GetOrderByClientRequestID(_ context.Context, companyID int, requestID string) (*market.MarketOrder, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	order, ok := s.ordersByRequest[marketRequestKey(companyID, requestID)]
	if !ok {
		return nil, nil
	}
	return order, nil
}

// CancelMarketOrder atomically refunds the unfilled reservation and marks the
// order cancelled. The replay flag lets callers keep HTTP cancellation
// idempotent without writing a duplicate refund ledger entry.
func (s *Store) CancelMarketOrder(_ context.Context, companyID int, orderID string) (*market.MarketOrder, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	order, ok := s.orders[orderID]
	if !ok || order.CompanyID != companyID {
		return nil, false, fmt.Errorf("order not found")
	}
	if order.Status == market.StatusCancelled {
		copy := *order
		return &copy, true, nil
	}
	remaining := order.Remaining()
	if order.Status == market.StatusFilled || remaining <= 0 {
		return nil, false, storage.ErrAlreadySettled
	}
	company, ok := s.companies[companyID]
	if !ok {
		return nil, false, fmt.Errorf("company %d not found", companyID)
	}
	if order.IsBuy {
		company.Money += order.Price * float64(remaining)
	} else if err := s.updateInventoryLocked(companyID, order.ResourceID, remaining); err != nil {
		return nil, false, err
	}
	order.Status = market.StatusCancelled
	copy := *order
	return &copy, false, nil
}

func marketRequestKey(companyID int, requestID string) string {
	return fmt.Sprintf("%d:%s", companyID, requestID)
}

func (s *Store) UpdateOrder(_ context.Context, o *market.MarketOrder) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, ok := s.orders[o.ID]; ok && existing.Status == market.StatusCancelled && o.Status != market.StatusCancelled {
		return storage.ErrStateConflict
	}
	s.orders[o.ID] = o
	return nil
}

func (s *Store) GetOpenOrders(_ context.Context, resourceID int, isBuy bool) ([]market.MarketOrder, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []market.MarketOrder
	for _, o := range s.orders {
		if o.ResourceID == resourceID && o.IsBuy == isBuy && o.Status == market.StatusOpen {
			result = append(result, *o)
		}
	}
	return result, nil
}

func (s *Store) GetOrdersByCompany(_ context.Context, companyID int) ([]market.MarketOrder, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []market.MarketOrder
	for _, o := range s.orders {
		if o.CompanyID == companyID {
			result = append(result, *o)
		}
	}
	return result, nil
}

func (s *Store) GetOrdersByResource(_ context.Context, resourceID int) ([]market.MarketOrder, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []market.MarketOrder
	for _, o := range s.orders {
		if o.ResourceID == resourceID {
			result = append(result, *o)
		}
	}
	return result, nil
}

func (s *Store) SaveTrade(_ context.Context, t *market.Trade) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.trades = append(s.trades, *t)
	return nil
}

func (s *Store) GetTrades(_ context.Context, _ int, limit int) ([]market.Trade, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n := len(s.trades)
	if limit > 0 && limit < n {
		n = limit
	}
	return s.trades[len(s.trades)-n:], nil
}

func (s *Store) CreateTakeOrderExecution(_ context.Context, execution *market.TakeOrderExecution) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := marketRequestKey(execution.CompanyID, execution.ClientRequestID)
	if _, exists := s.takeOrderExecutions[key]; exists {
		return fmt.Errorf("%w: take order request", storage.ErrAlreadyExists)
	}
	copy := *execution
	copy.Trades = append([]market.Trade(nil), execution.Trades...)
	s.takeOrderExecutions[key] = &copy
	return nil
}

func (s *Store) GetTakeOrderExecution(_ context.Context, companyID int, requestID string) (*market.TakeOrderExecution, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	execution, ok := s.takeOrderExecutions[marketRequestKey(companyID, requestID)]
	if !ok {
		return nil, nil
	}
	copy := *execution
	copy.Trades = append([]market.Trade(nil), execution.Trades...)
	return &copy, nil
}

func (s *Store) UpdateTakeOrderExecution(_ context.Context, execution *market.TakeOrderExecution) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := marketRequestKey(execution.CompanyID, execution.ClientRequestID)
	if _, exists := s.takeOrderExecutions[key]; !exists {
		return fmt.Errorf("take order execution not found")
	}
	copy := *execution
	copy.Trades = append([]market.Trade(nil), execution.Trades...)
	s.takeOrderExecutions[key] = &copy
	return nil
}

func (s *Store) GetTicker(_ context.Context, resourceID int) (*market.Ticker, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tickers[resourceID]
	if !ok {
		return nil, fmt.Errorf("ticker not found")
	}
	return t, nil
}

func (s *Store) UpdateTicker(_ context.Context, t *market.Ticker) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tickers[t.ResourceID] = t
	return nil
}

func (s *Store) GetTickers(_ context.Context) ([]market.Ticker, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []market.Ticker
	for _, t := range s.tickers {
		result = append(result, *t)
	}
	return result, nil
}

// --- ProductionStorage ---

func (s *Store) CreateJob(_ context.Context, j *production.ProductionJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, existing := range s.jobs {
		if existing.CompanyID == j.CompanyID && existing.BuildingID == j.BuildingID && existing.Status != production.StatusClaimed && existing.Status != production.StatusCancelled {
			return fmt.Errorf("%w: production line", storage.ErrAlreadyExists)
		}
	}
	if j.ClientRequestID != "" {
		key := productionRequestKey(j.CompanyID, j.ClientRequestID)
		if _, exists := s.jobsByRequest[key]; exists {
			return fmt.Errorf("%w: production request", storage.ErrAlreadyExists)
		}
		s.jobsByRequest[key] = j
	}
	s.jobs[j.ID] = j
	return nil
}

func (s *Store) GetJob(_ context.Context, jobID string) (*production.ProductionJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	j, ok := s.jobs[jobID]
	if !ok {
		return nil, fmt.Errorf("job not found")
	}
	copy := *j
	return &copy, nil
}

// ClaimProductionOutput atomically moves a claimable amount into the owning
// company's inventory, settles accrued payroll, awards XP, and advances the
// job state. Keeping these mutations under one lock prevents two service
// instances from duplicating output or charging the same active seconds.
func (s *Store) ClaimProductionOutput(_ context.Context, companyID int, jobID string, expectedClaimAmount int, xpEarned int, payroll production.PayrollSettlement) (*production.ProductionJob, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[jobID]
	if !ok || job.CompanyID != companyID {
		return nil, fmt.Errorf("job not found")
	}
	if job.Status == production.StatusClaimed || job.Status == production.StatusCancelled {
		return nil, storage.ErrAlreadySettled
	}
	if job.ClaimableAmount <= 0 {
		return nil, storage.ErrNothingToClaim
	}
	if expectedClaimAmount <= 0 || job.ClaimableAmount != expectedClaimAmount {
		return nil, storage.ErrStateConflict
	}
	if !validProductionPayroll(job, payroll) {
		return nil, storage.ErrStateConflict
	}
	if err := s.updateInventoryLocked(companyID, job.ResourceID, expectedClaimAmount); err != nil {
		return nil, err
	}
	company := s.companies[companyID]
	if company == nil {
		return nil, fmt.Errorf("company %d not found", companyID)
	}
	company.Money -= payroll.Amount
	company.XP += int64(xpEarned)
	job.PayrollSettledSeconds = payroll.SettledSeconds
	job.ClaimedAmount += expectedClaimAmount
	if job.ClaimedAmount >= job.TargetQuantity {
		job.ClaimedAmount = job.TargetQuantity
		job.Status = production.StatusClaimed
	}
	job.ClaimableAmount = 0
	job.XPAwarded += xpEarned
	copy := *job
	return &copy, nil
}

// CancelProductionJob atomically applies every input refund and records a
// cancellation tombstone. The tombstone keeps the original request ID alive,
// preventing delayed start retries from resurrecting a cancelled run.
func (s *Store) CancelProductionJob(_ context.Context, companyID int, jobID string, refunds map[int]int, payroll production.PayrollSettlement) (*production.ProductionJob, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[jobID]
	if !ok || job.CompanyID != companyID {
		return nil, false, fmt.Errorf("job not found")
	}
	if job.Status == production.StatusClaimed {
		return nil, false, storage.ErrAlreadySettled
	}
	if job.Status == production.StatusCancelled {
		copy := *job
		return &copy, true, nil
	}
	if _, ok := s.companies[companyID]; !ok {
		return nil, false, fmt.Errorf("company %d not found", companyID)
	}
	if !validProductionPayroll(job, payroll) {
		return nil, false, storage.ErrStateConflict
	}
	if _, ok := s.warehouses[companyID]; !ok {
		return nil, false, fmt.Errorf("warehouse for company %d not found", companyID)
	}
	for resourceID, amount := range refunds {
		if amount <= 0 {
			continue
		}
		if err := s.updateInventoryLocked(companyID, resourceID, amount); err != nil {
			return nil, false, err
		}
	}
	s.companies[companyID].Money -= payroll.Amount
	job.PayrollSettledSeconds = payroll.SettledSeconds
	job.Status = production.StatusCancelled
	job.ClaimableAmount = 0
	copy := *job
	return &copy, false, nil
}

func validProductionPayroll(job *production.ProductionJob, payroll production.PayrollSettlement) bool {
	if payroll.Amount < 0 || payroll.SettledSeconds < payroll.ExpectedSeconds {
		return false
	}
	return math.Abs(job.PayrollSettledSeconds-payroll.ExpectedSeconds) < 0.000001
}

func (s *Store) GetJobByClientRequestID(_ context.Context, companyID int, requestID string) (*production.ProductionJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	job, ok := s.jobsByRequest[productionRequestKey(companyID, requestID)]
	if !ok {
		return nil, nil
	}
	return job, nil
}

func productionRequestKey(companyID int, requestID string) string {
	return fmt.Sprintf("%d:%s", companyID, requestID)
}

func (s *Store) GetJobsByCompany(_ context.Context, companyID int) ([]production.ProductionJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []production.ProductionJob
	for _, j := range s.jobs {
		if j.CompanyID == companyID {
			result = append(result, *j)
		}
	}
	return result, nil
}

func (s *Store) GetJobsByBuilding(_ context.Context, buildingID string) ([]production.ProductionJob, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []production.ProductionJob
	for _, j := range s.jobs {
		if j.BuildingID == buildingID {
			result = append(result, *j)
		}
	}
	return result, nil
}

func (s *Store) UpdateJob(_ context.Context, j *production.ProductionJob) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, ok := s.jobs[j.ID]; ok {
		if existing.ClaimedAmount > j.ClaimedAmount ||
			existing.XPAwarded > j.XPAwarded ||
			((existing.Status == production.StatusClaimed || existing.Status == production.StatusCancelled) && j.Status != existing.Status) {
			return storage.ErrStateConflict
		}
	}
	s.jobs[j.ID] = j
	if j.ClientRequestID != "" {
		s.jobsByRequest[productionRequestKey(j.CompanyID, j.ClientRequestID)] = j
	}
	return nil
}

func (s *Store) DeleteJob(_ context.Context, jobID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	job, ok := s.jobs[jobID]
	if !ok {
		return fmt.Errorf("job not found")
	}
	delete(s.jobs, jobID)
	if job.ClientRequestID != "" {
		delete(s.jobsByRequest, productionRequestKey(job.CompanyID, job.ClientRequestID))
	}
	return nil
}

// --- FinanceStorage ---
func (s *Store) AppendLedgerEntry(_ context.Context, e *finance.LedgerEntry) error {
	s.ledgerMu.Lock()
	defer s.ledgerMu.Unlock()
	e.ID = s.nextLedgerID
	s.nextLedgerID++
	s.ledger = append(s.ledger, *e)
	return nil
}
func (s *Store) GetLedgerEntries(_ context.Context, companyID int, limit int) ([]finance.LedgerEntry, error) {
	s.ledgerMu.Lock()
	defer s.ledgerMu.Unlock()
	var result []finance.LedgerEntry
	for i := len(s.ledger) - 1; i >= 0; i-- {
		if s.ledger[i].CompanyID == companyID {
			result = append(result, s.ledger[i])
			if limit > 0 && len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (s *Store) CreateBond(_ context.Context, b *finance.Bond) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bonds[b.ID] = b
	return nil
}

func (s *Store) GetBond(_ context.Context, bondID string) (*finance.Bond, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	b, ok := s.bonds[bondID]
	if !ok {
		return nil, fmt.Errorf("bond not found")
	}
	return b, nil
}

func (s *Store) UpdateBond(_ context.Context, b *finance.Bond) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.bonds[b.ID] = b
	return nil
}

func (s *Store) GetActiveBonds(_ context.Context) ([]finance.Bond, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []finance.Bond
	for _, b := range s.bonds {
		if b.Status == "active" {
			result = append(result, *b)
		}
	}
	return result, nil
}

func (s *Store) GetBondsByIssuer(_ context.Context, companyID int) ([]finance.Bond, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []finance.Bond
	for _, b := range s.bonds {
		if b.IssuerCompanyID == companyID {
			result = append(result, *b)
		}
	}
	return result, nil
}

func (s *Store) CreateBondHolding(_ context.Context, h *finance.BondHolding) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.holdings = append(s.holdings, *h)
	return nil
}

func (s *Store) GetBondHoldings(_ context.Context, bondID string) ([]finance.BondHolding, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []finance.BondHolding
	for _, h := range s.holdings {
		if h.BondID == bondID {
			result = append(result, h)
		}
	}
	return result, nil
}

func (s *Store) GetCompanyBondHoldings(_ context.Context, companyID int) ([]finance.BondHolding, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []finance.BondHolding
	for _, h := range s.holdings {
		if h.CompanyID == companyID {
			result = append(result, h)
		}
	}
	return result, nil
}

// --- ResearchStorage ---

func (s *Store) GetCompanyResearch(_ context.Context, companyID int) ([]research.ResourceResearch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []research.ResourceResearch
	prefix := fmt.Sprintf("%d:", companyID)
	for k, v := range s.companyResearch {
		if strings.HasPrefix(k, prefix) {
			result = append(result, *v)
		}
	}
	if result == nil {
		result = []research.ResourceResearch{}
	}
	return result, nil
}

func (s *Store) GetResourceResearch(_ context.Context, companyID int, resourceID int) (*research.ResourceResearch, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	key := fmt.Sprintf("%d:%d", companyID, resourceID)
	if rr, ok := s.companyResearch[key]; ok {
		return rr, nil
	}
	return nil, nil
}

func (s *Store) SaveResourceResearch(_ context.Context, rr *research.ResourceResearch) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := fmt.Sprintf("%d:%d", rr.CompanyID, rr.ResourceID)
	s.companyResearch[key] = rr
	return nil
}

// --- SocialStorage ---

func (s *Store) SaveMessage(_ context.Context, m *social.Message) error {
	s.msgMu.Lock()
	defer s.msgMu.Unlock()
	m.ID = s.nextMessageID
	s.nextMessageID++
	s.messages = append(s.messages, *m)
	return nil
}

func (s *Store) GetMessages(_ context.Context, channel string, limit int) ([]social.Message, error) {
	s.msgMu.Lock()
	defer s.msgMu.Unlock()

	// Filter by channel if specified
	var filtered []social.Message
	if channel == "" {
		filtered = s.messages
	} else {
		filtered = make([]social.Message, 0, len(s.messages))
		for _, m := range s.messages {
			if m.Channel == channel {
				filtered = append(filtered, m)
			}
		}
	}

	n := len(filtered)
	if limit > 0 && limit < n {
		n = limit
	}
	if n == 0 {
		return []social.Message{}, nil
	}
	return filtered[len(filtered)-n:], nil
}

func (s *Store) CreateNotification(_ context.Context, n *social.Notification) error {
	s.notifMu.Lock()
	defer s.notifMu.Unlock()
	n.ID = s.nextNotifID
	s.nextNotifID++
	s.notifs = append(s.notifs, *n)
	return nil
}

func (s *Store) GetNotifications(_ context.Context, companyID int, limit int) ([]social.Notification, error) {
	s.notifMu.Lock()
	defer s.notifMu.Unlock()
	var result []social.Notification
	for i := len(s.notifs) - 1; i >= 0; i-- {
		if s.notifs[i].CompanyID == companyID {
			result = append(result, s.notifs[i])
			if limit > 0 && len(result) >= limit {
				break
			}
		}
	}
	return result, nil
}

func (s *Store) MarkNotificationRead(_ context.Context, notificationID int) error {
	s.notifMu.Lock()
	defer s.notifMu.Unlock()
	for i := range s.notifs {
		if s.notifs[i].ID == notificationID {
			s.notifs[i].Read = true
			break
		}
	}
	return nil
}

// --- ChatStorage ---

func (s *Store) GetOrCreateRoom(_ context.Context, companyID1, companyID2 int) (*chat.ChatRoom, error) {
	s.chatData.Lock()
	defer s.chatData.Unlock()

	a, b := min(companyID1, companyID2), max(companyID1, companyID2)
	roomID := fmt.Sprintf("p:%d-%d", a, b)

	if existing, ok := s.chatData.rooms[roomID]; ok {
		return existing, nil
	}

	now := time.Now().UTC().Format(time.RFC3339)
	room := &chat.ChatRoom{
		ID:           roomID,
		Participant1: a,
		Participant2: b,
		CreatedAt:    now,
	}
	s.chatData.rooms[roomID] = room
	s.chatData.messages[roomID] = []chat.Message{}
	return room, nil
}

func (s *Store) GetUserRooms(_ context.Context, companyID int) ([]*chat.ChatRoom, error) {
	s.chatData.RLock()
	defer s.chatData.RUnlock()

	var result []*chat.ChatRoom
	for _, r := range s.chatData.rooms {
		if r.Participant1 == companyID || r.Participant2 == companyID {
			result = append(result, r)
		}
	}

	// Sort by last message time DESC
	sort.Slice(result, func(i, j int) bool {
		return result[i].LastMessageAt > result[j].LastMessageAt
	})

	return result, nil
}

func (s *Store) GetRoomMessages(_ context.Context, roomID string, limit int) ([]chat.Message, error) {
	s.chatData.RLock()
	defer s.chatData.RUnlock()

	msgs := s.chatData.messages[roomID]
	n := len(msgs)
	if limit > 0 && limit < n {
		n = limit
	}
	if n == 0 {
		return []chat.Message{}, nil
	}
	return msgs[len(msgs)-n:], nil
}

func (s *Store) SaveRoomMessage(_ context.Context, msg *chat.Message) error {
	s.chatData.Lock()
	defer s.chatData.Unlock()

	msg.ID = s.chatData.nextID
	s.chatData.nextID++
	s.chatData.messages[msg.RoomID] = append(s.chatData.messages[msg.RoomID], *msg)

	// Update room's last message time
	if room, ok := s.chatData.rooms[msg.RoomID]; ok {
		room.LastMessageAt = msg.CreatedAt
		room.LastMessage = msg.Content
	}
	return nil
}

func (s *Store) MarkRoomRead(_ context.Context, roomID string, companyID, lastMessageID int64) error {
	s.chatData.Lock()
	defer s.chatData.Unlock()
	key := fmt.Sprintf("%s:%d", roomID, companyID)
	s.chatData.readAt[key] = lastMessageID
	return nil
}

func (s *Store) GetRoomReadStatus(_ context.Context, roomID string, companyID int) int64 {
	s.chatData.RLock()
	defer s.chatData.RUnlock()
	key := fmt.Sprintf("%s:%d", roomID, companyID)
	return s.chatData.readAt[key]
}
