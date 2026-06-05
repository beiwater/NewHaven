package memory

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/newhaven/backend-next/internal/domain/auth"
	"github.com/newhaven/backend-next/internal/domain/company"
	"github.com/newhaven/backend-next/internal/domain/finance"
	"github.com/newhaven/backend-next/internal/domain/market"
	"github.com/newhaven/backend-next/internal/domain/production"
	"github.com/newhaven/backend-next/internal/domain/research"
	"github.com/newhaven/backend-next/internal/domain/social"
	"github.com/newhaven/backend-next/internal/domain/warehouse"
)

// Store is an in-memory implementation of storage.Storage.
// Thread-safe via sync.RWMutex per domain.
// Data is NOT persisted across restarts.
type Store struct {
	mu         sync.RWMutex
	players    map[int]*auth.Player
	byUser     map[string]*auth.Player
	companies  map[int]*company.Company
	byPlayer   map[int]*company.Company
	orders     map[string]*market.MarketOrder
	trades     []market.Trade
	tickers    map[int]*market.Ticker
	jobs       map[string]*production.ProductionJob
	ledger     []finance.LedgerEntry
	bonds      map[string]*finance.Bond
	holdings   []finance.BondHolding
	research   []research.Project
	progress   []research.CompanyProgress
	messages   []social.Message
	notifs     []social.Notification
	warehouses map[int]*warehouse.Warehouse
	buildings  map[string]*company.Building
	nextID     int
}

func New() *Store {
	return &Store{
		players:    make(map[int]*auth.Player),
		byUser:     make(map[string]*auth.Player),
		companies:  make(map[int]*company.Company),
		byPlayer:   make(map[int]*company.Company),
		orders:     make(map[string]*market.MarketOrder),
		tickers:    make(map[int]*market.Ticker),
		jobs:       make(map[string]*production.ProductionJob),
		bonds:      make(map[string]*finance.Bond),
		buildings:  make(map[string]*company.Building),
		warehouses: make(map[int]*warehouse.Warehouse),
		nextID:     1,
	}
}

func (s *Store) Close() error { return nil }

// --- PlayerStorage ---

func (s *Store) CreatePlayer(_ context.Context, p *auth.Player) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.byUser[p.Username]; ok {
		return fmt.Errorf("username already exists")
	}
	p.ID = s.nextID
	s.nextID++
	p.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	s.players[p.ID] = p
	s.byUser[p.Username] = p
	return nil
}

func (s *Store) GetPlayerByUsername(_ context.Context, username string) (*auth.Player, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.byUser[username]
	if !ok {
		return nil, fmt.Errorf("player not found")
	}
	return p, nil
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
	c.ID = s.nextID + 1000000
	s.nextID++
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
	// Auto-create sample buildings for new company
	bld1 := company.Building{
		ID:         fmt.Sprintf("bld-%d-1", c.ID),
		BuildingID: 1,
		Kind:       1,
		Name:       "Bakery",
		Level:      1,
		MapID:      "map_1",
		SlotID:     "slot_a1",
		X:          5,
		Y:          10,
	}
	bld2 := company.Building{
		ID:         fmt.Sprintf("bld-%d-2", c.ID),
		BuildingID: 2,
		Kind:       2,
		Name:       "Workshop",
		Level:      1,
		MapID:      "map_1",
		SlotID:     "slot_b1",
		X:          15,
		Y:          20,
	}
	c.Buildings = []company.Building{bld1, bld2}
	s.buildings[bld1.ID] = &c.Buildings[0]
	s.buildings[bld2.ID] = &c.Buildings[1]
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

func (s *Store) UpdateCompany(_ context.Context, c *company.Company) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.companies[c.ID] = c
	return nil
}

func (s *Store) SaveBuilding(_ context.Context, b *company.Building) error { return nil }
func (s *Store) RemoveBuilding(_ context.Context, buildingID string) error { return nil }
func (s *Store) UpdateInventory(_ context.Context, companyID int, resourceID int, delta int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
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

// --- MarketStorage ---

func (s *Store) CreateOrder(_ context.Context, o *market.MarketOrder) error {
	s.mu.Lock()
	defer s.mu.Unlock()
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

func (s *Store) UpdateOrder(_ context.Context, o *market.MarketOrder) error {
	s.mu.Lock()
	defer s.mu.Unlock()
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
	return j, nil
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
	s.jobs[j.ID] = j
	return nil
}

// --- FinanceStorage ---

func (s *Store) AppendLedgerEntry(_ context.Context, e *finance.LedgerEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	e.ID = int64(s.nextID)
	s.nextID++
	s.ledger = append(s.ledger, *e)
	return nil
}

func (s *Store) GetLedgerEntries(_ context.Context, companyID int, limit int) ([]finance.LedgerEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
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

func (s *Store) GetProjects(_ context.Context) ([]research.Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.research, nil
}

func (s *Store) GetCompanyProgress(_ context.Context, companyID int) ([]research.CompanyProgress, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var result []research.CompanyProgress
	for _, p := range s.progress {
		if p.CompanyID == companyID {
			result = append(result, p)
		}
	}
	return result, nil
}

func (s *Store) SaveProgress(_ context.Context, p *research.CompanyProgress) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.progress = append(s.progress, *p)
	return nil
}

// --- SocialStorage ---

func (s *Store) SaveMessage(_ context.Context, m *social.Message) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	m.ID = int64(s.nextID)
	s.nextID++
	s.messages = append(s.messages, *m)
	return nil
}

func (s *Store) GetMessages(_ context.Context, _ string, limit int) ([]social.Message, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n := len(s.messages)
	if limit > 0 && limit < n {
		n = limit
	}
	return s.messages[len(s.messages)-n:], nil
}

func (s *Store) CreateNotification(_ context.Context, n *social.Notification) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	n.ID = s.nextID
	s.nextID++
	s.notifs = append(s.notifs, *n)
	return nil
}

func (s *Store) GetNotifications(_ context.Context, companyID int, limit int) ([]social.Notification, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
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
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.notifs {
		if s.notifs[i].ID == notificationID {
			s.notifs[i].Read = true
			break
		}
	}
	return nil
}
