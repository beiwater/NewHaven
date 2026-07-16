package market

import (
	"context"
	"log/slog"
	"math"
	"strconv"
	"sync"
	"time"

	"github.com/beiwater/NewHaven/backend/internal/catalog"
	domain "github.com/beiwater/NewHaven/backend/internal/domain/company"
	"github.com/beiwater/NewHaven/backend/internal/domain/finance"
	"github.com/beiwater/NewHaven/backend/internal/formula"
)

// NPC retail constants.
const (
	defaultStoreSize = 1
	defaultWeather   = 1.06
	defaultAccel     = 1.0
)

// retailShelf represents a single shelf sale to process.
type retailShelf struct {
	buildingID    string
	buildingKind  int
	buildingLevel int
	resourceID    int
	quantity      int
	price         float64
	priceLocked   bool
	salesModPct   float64
}

type retailBuilding struct {
	kind  int
	level int
}

// ProcessRetailSales iterates all bot/NPC companies and sells their goods.
// NPC companies still sell from inventory (they don't use shelves).
// Real player companies are identified by their positive PlayerID and skipped;
// LastRetailAt is settlement state, not a safe account-type discriminator.
func (s *Service) ProcessRetailSales(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.economy) == 0 {
		slog.Debug("[retail] no economy model loaded, skipping")
		return nil
	}

	companies, err := s.companies.GetAllCompanies(ctx)
	if err != nil {
		return err
	}

	now := s.clock.Now().UTC()
	var skipped int

	for _, company := range companies {
		if company == nil {
			continue
		}
		if company.PlayerID > 0 {
			skipped++
			continue
		}
		s.processNPCRetail(ctx, company, now)
	}

	if skipped > 0 {
		slog.Debug("[retail] skipped player companies on scheduler tick", "count", skipped)
	}
	return nil
}

// processNPCRetail sells NPC inventory items (legacy behavior — NPCs don't use shelves).
func (s *Service) processNPCRetail(ctx context.Context, company *domain.Company, now time.Time) {
	if len(company.Inventory) == 0 {
		return
	}
	for resourceID, qty := range company.Inventory {
		if qty <= 0 {
			continue
		}
		eco, ok := s.economy[resourceID]
		if !ok {
			continue
		}
		price := s.salePriceForResource(ctx, resourceID)
		if price <= 0 {
			continue
		}
		sold, earned := computeSale(eco, price, price, qty, defaultStoreSize, 0, 60)
		if sold <= 0 {
			continue
		}
		s.applyNPCSale(ctx, company, resourceID, sold, earned, price, now)
	}
}

// CatchUpPlayerRetail computes retail sales for a player company since its last
// settlement. Only sells from retail building shelves.
func (s *Service) CatchUpPlayerRetail(ctx context.Context, companyID int) error {
	// Profile data is polled frequently and the same account may be open in
	// multiple tabs. Serialize the read/compute/write settlement so two requests
	// cannot sell the same shelf interval twice. The lock is per company so one
	// busy player cannot stall retail settlement for every other account.
	lockValue, _ := s.retailLocks.LoadOrStore(companyID, &sync.Mutex{})
	companyLock := lockValue.(*sync.Mutex)
	companyLock.Lock()
	defer companyLock.Unlock()

	company, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return err
	}

	now := s.clock.Now().UTC()
	legacyBatchLocked := s.lockLegacySaleBatches(ctx, company)

	if company.LastRetailAt == "" {
		company.LastRetailAt = now.Add(-60 * time.Second).Format(time.RFC3339)
		return s.companies.UpdateCompany(ctx, company)
	}

	lastSettle, err := time.Parse(time.RFC3339, company.LastRetailAt)
	if err != nil {
		return err
	}
	elapsedSeconds := now.Sub(lastSettle).Seconds()
	if elapsedSeconds <= 0 {
		if legacyBatchLocked {
			return s.companies.UpdateCompany(ctx, company)
		}
		return nil
	}

	if len(s.economy) == 0 {
		slog.Debug("[retail] no economy model loaded for catch-up")
		company.LastRetailAt = now.Format(time.RFC3339)
		company.RetailCarry = nil
		return s.companies.UpdateCompany(ctx, company)
	}

	shelves := s.collectPlayerShelves(company)
	nextCarry := make(map[string]float64, len(shelves))
	activeSecondsByBuilding := make(map[string]float64, len(shelves))
	activeBuildings := make(map[string]retailBuilding, len(shelves))
	for _, sh := range shelves {
		if sh.quantity <= 0 {
			continue
		}
		activeBuildings[sh.buildingID] = retailBuilding{kind: sh.buildingKind, level: sh.buildingLevel}
		activeSeconds := elapsedSeconds
		qty := sh.quantity
		eco, ok := s.economy[sh.resourceID]
		if !ok {
			activeSecondsByBuilding[sh.buildingID] = math.Max(activeSecondsByBuilding[sh.buildingID], activeSeconds)
			continue
		}
		price := sh.price
		if !sh.priceLocked {
			if mp := s.salePriceForResource(ctx, sh.resourceID); mp > 0 {
				price = mp
			}
		}
		if price <= 0 {
			activeSecondsByBuilding[sh.buildingID] = math.Max(activeSecondsByBuilding[sh.buildingID], activeSeconds)
			continue
		}
		recommendedPrice := s.salePriceForResource(ctx, sh.resourceID)
		unitsPerHour := retailUnitsPerHour(eco, recommendedPrice, price, sh.buildingLevel, sh.salesModPct)
		activeSeconds = retailActiveSeconds(unitsPerHour, qty, elapsedSeconds, company.RetailCarry[sh.carryKey()])
		activeSecondsByBuilding[sh.buildingID] = math.Max(activeSecondsByBuilding[sh.buildingID], activeSeconds)
		key := sh.carryKey()
		sold, earned, carry := computeSaleWithCarryAtRate(unitsPerHour, price, qty, elapsedSeconds, company.RetailCarry[key])
		if sold < qty && carry > 0 {
			nextCarry[key] = carry
		}
		if sold <= 0 {
			continue
		}
		s.applyPlayerShelfSale(ctx, company, sh.buildingID, sh.resourceID, sold, earned, price, now)
	}

	for buildingID, activeSeconds := range activeSecondsByBuilding {
		if activeSeconds <= 0 {
			continue
		}
		building := activeBuildings[buildingID]
		workers := formula.BuildingWorkerCount(building.kind, building.level)
		hourlyWage := formula.BuildingHourlyWage(building.kind, building.level)
		payroll := hourlyWage * activeSeconds / 3600
		if payroll <= 0 {
			continue
		}
		company.Money -= payroll
		logRetailPayroll(ctx, s.finance, company.ID, buildingID, workers, hourlyWage, activeSeconds, payroll, now)
	}

	company.RetailCarry = nextCarry
	company.LastRetailAt = now.Format(time.RFC3339)
	return s.companies.UpdateCompany(ctx, company)
}

// lockLegacySaleBatches migrates shelves created before sale batches became
// immutable. Their current positive price (or today's market fallback) becomes
// the committed batch price, so a deployment cannot leave old sales floating.
func (s *Service) lockLegacySaleBatches(ctx context.Context, company *domain.Company) bool {
	changed := false
	for i := range company.Buildings {
		building := &company.Buildings[i]
		entry, ok := s.buildings[building.Kind]
		if !ok || entry.Type != "retail" {
			continue
		}
		for j := range building.Shelves {
			shelf := &building.Shelves[j]
			if shelf.Quantity <= 0 || shelf.PriceLock {
				continue
			}
			if shelf.Price <= 0 {
				shelf.Price = s.salePriceForResource(ctx, shelf.ResourceID)
			}
			if shelf.Price > 0 {
				shelf.PriceLock = true
				changed = true
			}
		}
	}
	return changed
}

// collectPlayerShelves gathers all non-empty shelves from the company's retail buildings.
func (s *Service) collectPlayerShelves(company *domain.Company) []retailShelf {
	salesModPct := aggregateSalesBonus(company.Executives)
	var shelves []retailShelf

	for i := range company.Buildings {
		b := &company.Buildings[i]
		entry, ok := s.buildings[b.Kind]
		if !ok || entry.Type != "retail" {
			continue
		}
		for _, shelf := range b.Shelves {
			if shelf.Quantity <= 0 {
				continue
			}
			shelves = append(shelves, retailShelf{
				buildingID:    b.ID,
				buildingKind:  b.Kind,
				buildingLevel: b.Level,
				resourceID:    shelf.ResourceID,
				quantity:      shelf.Quantity,
				price:         shelf.Price,
				priceLocked:   shelf.PriceLock,
				salesModPct:   salesModPct,
			})
		}
	}
	return shelves
}

func (sh retailShelf) carryKey() string {
	return sh.buildingID + ":" + strconv.Itoa(sh.resourceID)
}

// salePriceForResource returns the current market price or base price.
func (s *Service) salePriceForResource(ctx context.Context, resourceID int) float64 {
	price := s.basePriceForResource(resourceID)
	ticker, err := s.market.GetTicker(ctx, resourceID)
	if err == nil && ticker.LastPrice > 0 {
		price = ticker.LastPrice
	}
	return price
}

// computeSale calculates units sold and revenue for one item.
func computeSale(eco *catalog.EconomyModelEntry, recommendedPrice, price float64, availableQty int, storeSize int, salesModPct float64, elapsedSeconds float64) (sold int, earned float64) {
	unitsSold := retailUnitsSold(eco, recommendedPrice, price, storeSize, salesModPct, elapsedSeconds)
	if unitsSold <= 0 {
		return 0, 0
	}
	sold = int(math.Floor(unitsSold))
	if sold > availableQty {
		sold = availableQty
	}
	if sold <= 0 {
		return 0, 0
	}
	return sold, float64(sold) * price
}

// computeSaleWithCarry carries fractional demand forward across short catch-ups.
// It is deliberately used only for player shelves; NPC scheduler ticks already
// run at a fixed one-minute cadence.
func computeSaleWithCarryAtRate(unitsPerHour, price float64, availableQty int, elapsedSeconds, carry float64) (sold int, earned, remainingCarry float64) {
	unitsSold := unitsPerHour*elapsedSeconds/3600 + carry
	if unitsSold <= 0 {
		return 0, 0, 0
	}
	sold = int(math.Floor(unitsSold))
	if sold > availableQty {
		return availableQty, float64(availableQty) * price, 0
	}
	remainingCarry = unitsSold - float64(sold)
	if sold <= 0 {
		return 0, 0, remainingCarry
	}
	return sold, float64(sold) * price, remainingCarry
}

func retailUnitsSold(eco *catalog.EconomyModelEntry, recommendedPrice, price float64, storeSize int, salesModPct float64, elapsedSeconds float64) float64 {
	return retailUnitsPerHour(eco, recommendedPrice, price, storeSize, salesModPct) * elapsedSeconds / 3600
}

func retailUnitsPerHour(eco *catalog.EconomyModelEntry, recommendedPrice, price float64, storeSize int, salesModPct float64) float64 {
	if eco == nil || price <= eco.ModeledProductionCostPerUnit {
		return 0
	}
	quality := 4.0
	saturation := 1.0
	accel := 1.0
	weather := 1.06
	if storeSize <= 0 {
		storeSize = 1
	}
	if recommendedPrice <= eco.ModeledProductionCostPerUnit {
		recommendedPrice = eco.ModeledProductionCostPerUnit * 1.25
	}

	unitsPerHour := formula.UnitsSoldPerHour(
		eco.BuildingKindModifier,
		eco.BuildingLevelsNeededPerUnitPerHour,
		eco.ModeledProductionCostPerUnit,
		eco.ModeledStoreWages,
		eco.ModeledUnitsSoldAnHour,
		recommendedPrice,
		quality,
		saturation,
		-salesModPct,
		1,
		accel,
		weather,
	)
	if unitsPerHour <= 0 {
		return 0
	}
	return unitsPerHour * float64(storeSize) * formula.RetailPriceSpeedMultiplier(price, recommendedPrice)
}

// retailActiveSeconds charges workers only for the time this shelf actually
// kept its building busy. A slow or overpriced batch remains active for the
// whole interval; a fast sell-out stops payroll at the exact sell-out point.
func retailActiveSeconds(unitsPerHour float64, quantity int, elapsedSeconds, carry float64) float64 {
	if elapsedSeconds <= 0 || quantity <= 0 {
		return 0
	}
	if unitsPerHour <= 0 {
		return elapsedSeconds
	}
	remainingDemand := float64(quantity) - math.Max(0, carry)
	if remainingDemand <= 0 {
		return 0
	}
	sellOutSeconds := remainingDemand * 3600 / unitsPerHour
	return math.Min(elapsedSeconds, math.Max(0, sellOutSeconds))
}

// applyNPCSale credits money, deducts inventory, logs.
func (s *Service) applyNPCSale(ctx context.Context, company *domain.Company, resourceID, sold int, earned, price float64, now time.Time) {
	if err := s.companies.UpdateInventory(ctx, company.ID, resourceID, -sold); err != nil {
		slog.Warn("[retail] NPC inventory deduction failed", "company", company.ID, "resource", resourceID, "error", err)
		return
	}
	company.Money += earned
	if err := s.companies.UpdateCompany(ctx, company); err != nil {
		slog.Warn("[retail] NPC money credit failed", "company", company.ID, "resource", resourceID, "error", err)
		_ = s.companies.UpdateInventory(ctx, company.ID, resourceID, sold)
		company.Money -= earned
		return
	}
	logSale(ctx, s.finance, company.ID, "retail_sale", earned, resourceID, sold, price, now)
}

// applyPlayerShelfSale deducts from shelf, credits money, logs.
func (s *Service) applyPlayerShelfSale(ctx context.Context, company *domain.Company, buildingID string, resourceID, sold int, earned, price float64, now time.Time) {
	if !deductFromShelf(company, buildingID, resourceID, sold, earned) {
		slog.Warn("[retail] player shelf disappeared before settlement", "company", company.ID, "building", buildingID, "resource", resourceID)
		return
	}

	company.Money += earned
	logSale(ctx, s.finance, company.ID, "retail_sale", earned, resourceID, sold, price, now)
}

// deductFromShelf decrements the matching shelf quantity (or removes if empty).
func deductFromShelf(company *domain.Company, buildingID string, resourceID, sold int, earned float64) bool {
	for i := range company.Buildings {
		b := &company.Buildings[i]
		if b.ID != buildingID {
			continue
		}
		for j := range b.Shelves {
			if b.Shelves[j].ResourceID == resourceID {
				b.Shelves[j].Revenue += earned
				if b.Shelves[j].Quantity > sold {
					b.Shelves[j].Quantity -= sold
				} else {
					b.Shelves = append(b.Shelves[:j], b.Shelves[j+1:]...)
				}
				return true
			}
		}
	}
	return false
}

// aggregateSalesBonus sums executive sales bonuses.
func aggregateSalesBonus(execs []domain.Executive) float64 {
	var total float64
	for _, ex := range execs {
		total += ex.SalesBonus
	}
	return total
}

// logSale appends a finance ledger entry.
func logSale(ctx context.Context, financeSvc financeWriter, companyID int, kind string, earned float64, resourceID, sold int, price float64, now time.Time) {
	_ = financeSvc.AppendLedgerEntry(ctx, &finance.LedgerEntry{
		CompanyID: companyID,
		Kind:      kind,
		Amount:    earned,
		Direction: "in",
		Metadata: map[string]any{
			"resourceId": resourceID,
			"quantity":   sold,
			"price":      price,
		},
		CreatedAt: now.Format(time.RFC3339),
	})
}

func logRetailPayroll(ctx context.Context, financeSvc financeWriter, companyID int, buildingID string, workers int, hourlyWage, activeSeconds, amount float64, now time.Time) {
	_ = financeSvc.AppendLedgerEntry(ctx, &finance.LedgerEntry{
		CompanyID: companyID,
		Kind:      "retail_wages",
		Amount:    amount,
		Direction: "out",
		Metadata: map[string]any{
			"buildingId":    buildingID,
			"workers":       workers,
			"hourlyWage":    hourlyWage,
			"activeSeconds": activeSeconds,
		},
		CreatedAt: now.Format(time.RFC3339),
	})
}

// Small interface to avoid importing the full storage package.
type financeWriter interface {
	AppendLedgerEntry(ctx context.Context, entry *finance.LedgerEntry) error
}
