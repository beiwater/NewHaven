package market

import (
	"context"
	"log/slog"
	"math"
	"time"

	"github.com/newhaven/backend-next/internal/domain/finance"
	"github.com/newhaven/backend-next/internal/formula"
)

// ProcessRetailSales iterates all companies and sells their goods via the retail formula.
// For each company inventory item that has an economy model entry, it computes how many
// units are sold in the current 60-second tick, credits the company's money, and deducts
// inventory. Called by the scheduler every 60 seconds under s.mu.
func (s *Service) ProcessRetailSales(ctx context.Context) error {
	if len(s.economy) == 0 {
		slog.Debug("[retail] no economy model loaded, skipping")
		return nil
	}

	companies, err := s.companies.GetAllCompanies(ctx)
	if err != nil {
		return err
	}

	now := s.clock.Now().UTC()

	for _, company := range companies {
		if company == nil || len(company.Inventory) == 0 {
			continue
		}

		for resourceID, qty := range company.Inventory {
			if qty <= 0 {
				continue
			}

			eco, ok := s.economy[resourceID]
			if !ok {
				continue
			}

			// Get price: ticker lastPrice or resource basePrice
			price := s.basePriceForResource(resourceID)
			ticker, tickErr := s.market.GetTicker(ctx, resourceID)
			if tickErr == nil && ticker.LastPrice > 0 {
				price = ticker.LastPrice
			}
			if price <= 0 {
				continue
			}

			quality := 4.0       // middle quality (no per-inventory quality tracking yet)
			saturation := 1.0    // balanced market
			salesModPct := 0.0   // no modifier
			storeSize := 1       // default store size
			accel := 1.0         // default acceleration
			weather := 1.06      // default weather multiplier

			unitsPerHour := formula.UnitsSoldPerHour(
				eco.BuildingKindModifier,
				eco.BuildingLevelsNeededPerUnitPerHour,
				eco.ModeledProductionCostPerUnit,
				eco.ModeledStoreWages,
				eco.ModeledUnitsSoldAnHour,
				price,
				quality,
				saturation,
				salesModPct,
				storeSize,
				accel,
				weather,
			)

			// Scale from per-hour to per-tick (60 seconds)
			unitsSold := unitsPerHour * 60.0 / 3600.0
			if unitsSold <= 0 {
				continue
			}

			// Cap at available inventory
			sold := int(math.Floor(unitsSold))
			if sold > qty {
				sold = qty
			}
			if sold <= 0 {
				continue
			}

			earned := float64(sold) * price

			// Deduct inventory
			if err := s.companies.UpdateInventory(ctx, company.ID, resourceID, -sold); err != nil {
				slog.Warn("[retail] inventory deduction failed",
					"company", company.ID,
					"resource", resourceID,
					"error", err,
				)
				continue
			}

			// Credit money
			company.Money += earned
			if err := s.companies.UpdateCompany(ctx, company); err != nil {
				slog.Warn("[retail] money credit failed",
					"company", company.ID,
					"resource", resourceID,
					"error", err,
				)
				// Best-effort: rollback inventory
				_ = s.companies.UpdateInventory(ctx, company.ID, resourceID, sold)
				company.Money -= earned
				continue
			}

			// Ledger entry
			_ = s.finance.AppendLedgerEntry(ctx, &finance.LedgerEntry{
				CompanyID: company.ID,
				Kind:      "retail_sale",
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
	}

	return nil
}
