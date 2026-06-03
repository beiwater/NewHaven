package service

import (
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"time"

	"go-sim-api/internal/formula"
)

func (s *Service) BuildTicker(resourceID int) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now().UTC().Truncate(time.Hour)
	// Gather recent trade prices for this resource
	var prices []float64
	for _, t := range s.State.Trades {
		if t.ResourceID == resourceID {
			prices = append(prices, t.Price)
			if len(prices) >= 48 {
				break
			}
		}
	}
	// Generate 48 candles: use real prices where available, interpolate gaps
	series := make([]map[string]any, 0, 48)
	base := s.Cfg.Game.BotOrderBase + float64(resourceID%11)
	if price := s.State.LastTradePrice[resourceID]; price > 0 {
		base = price
	}
	lastPrice := base
	priceIdx := 0
	for i := 47; i >= 0; i-- {
		if priceIdx < len(prices) {
			lastPrice = prices[priceIdx]
			priceIdx++
		} else {
			// Deterministic reference curve. It must not change on every refresh.
			hour := now.Add(-time.Duration(i)*time.Hour).Unix() / 3600
			wave := math.Sin(float64(hour+int64(resourceID*17))*0.37) * 0.025
			wave += math.Cos(float64(hour+int64(resourceID*31))*0.11) * 0.015
			lastPrice = base * (1 + wave)
		}
		series = append(series, map[string]any{
			"time":  now.Add(-time.Duration(i) * time.Hour).UTC().Format(time.RFC3339),
			"price": math.Round(lastPrice*100) / 100,
		})
	}
	return map[string]any{"resource": resourceID, "series": series}
}

func (s *Service) ResourceInfo(resourceID int) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	info := map[string]any{"resourceId": resourceID}

	// Gather resource data from resources.json
	for _, r := range s.Data.Resources {
		if intFromAny(r["dbLetter"]) != resourceID {
			continue
		}
		info["name"] = r["name"]
		info["producedPerHourRaw"] = r["producedPerHourRaw"]
		info["unitsSoldAnHour"] = r["unitsSoldAnHour"]
		info["hasEconomyModel"] = r["hasEconomyModel"]

		// Production recipe
		if pf, ok := r["producedFrom"].(map[string]any); ok {
			recipe := make([]map[string]any, 0, 10)
			for k, q := range pf {
				rid, _ := strconv.Atoi(k)
				rname := ""
				for _, rr := range s.Data.Resources {
					if intFromAny(rr["dbLetter"]) == rid {
						rname = fmt.Sprint(rr["name"])
						break
					}
				}
				recipe = append(recipe, map[string]any{"resourceId": rid, "resourceName": rname, "quantity": q})
			}
			info["recipe"] = recipe
		}
		break
	}

	// Gather economy model and simulate market dynamics
	_, mPC, _, _ := s.getEconomyModelCost(resourceID)
	if models, ok := s.Data.EconomyModel["models"].(map[string]any); ok {
		if resM, ok := models[fmt.Sprintf("%d", resourceID)].(map[string]any); ok {
			if st1, ok := resM["state_1"].(map[string]any); ok {
				for k, v := range st1 {
					info[k] = v
				}
				// Simulate retail sales speed with representative values
				_ = floatFromAny(st1["buildingLevelsNeededPerUnitPerHour"])
				bkm := floatFromAny(st1["buildingKindModifier"])
				if bkm <= 0 {
					bkm = 1.0
				}
				// CTO skill bonus
				ctoSkill := 7.0
				if len(s.State.Executives) > 0 {
					for _, ex := range s.State.Executives {
						if ex["role"] == "CTO" {
							ctoSkill = floatFromAny(ex["skill"])
							break
						}
					}
				}
				prodMult := formula.CTOProductionMultiplier(ctoSkill)

				// Simulate output with new formula
				producedPerHourRaw := floatFromAny(info["producedPerHourRaw"])
				if producedPerHourRaw <= 0 {
					producedPerHourRaw = 500.0
				}
				outputPH := formula.OutputPerHour(producedPerHourRaw, 10.0, 1) // Lv1 +10% speed
				secPerUnit := 3600.0 / outputPH

				_ = s.State.ResearchedQuality
				// Saturation-based price simulation (balanced market)
				satMul := formula.SaturationPriceMultiplier(1.0, s.Cfg.Game.SaturationK)
				effectivePrice := formula.EffectivePrice(mPC*1.3, satMul, s.Cfg.Game.EventPriceMultiplier)

				info["simulatedMarketPrice"] = math.Round(effectivePrice*100) / 100
				info["outputPerHour"] = math.Round(outputPH*100) / 100
				info["secondsPerUnit"] = math.Round(secPerUnit*100) / 100
				info["saturationPriceMultiplier"] = math.Round(satMul*100) / 100
				info["ctoMultiplier"] = prodMult

				// Recommended price range (cost+margin)
				info["recommendedPrice"] = map[string]float64{
					"min": math.Round(mPC*1.1*100) / 100,
					"max": math.Round(mPC*1.5*100) / 100,
				}
			}
		}
	}

	// Add chain pricing
	for k, v := range s.ComputeChainPrice(resourceID) {
		info[k] = v
	}
	return info
}

// getEconomyModelCost extracts resource name and the economy model cost/wages/sales
// values for a given resource ID. This is shared between ResourceInfo and ComputeChainPrice.
func (s *Service) getEconomyModelCost(resourceID int) (string, float64, float64, float64) {
	resourceName := ""
	cost := s.Cfg.Game.BotOrderBase + float64(resourceID%7)
	wages := 0.0
	sales := 100.0

	for _, r := range s.Data.Resources {
		if intFromAny(r["dbLetter"]) != resourceID {
			continue
		}
		resourceName = fmt.Sprint(r["name"])
		if u, ok := r["unitsSoldAnHour"].(float64); ok && u > 0 {
			sales = u
		}
		break
	}

	if models, ok := s.Data.EconomyModel["models"].(map[string]any); ok {
		if resM, ok := models[fmt.Sprintf("%d", resourceID)].(map[string]any); ok {
			if st1, ok := resM["state_1"].(map[string]any); ok {
				cost = floatFromAny(st1["modeledProductionCostPerUnit"])
				wages = floatFromAny(st1["modeledStoreWages"])
				sales = floatFromAny(st1["modeledUnitsSoldAnHour"])
			}
		}
	}

	return resourceName, cost, wages, sales
}

// ComputeChainPrice calculates terminal price and splits margin across 3 tiers.
// Tiers get roughly equal net profit (10-15% random variance to create market dynamics).
func (s *Service) ComputeChainPrice(resourceID int) map[string]float64 {
	cost := s.Cfg.Game.BotOrderBase + float64(resourceID%7)
	wages := 0.0
	sales := 100.0
	if models, ok := s.Data.EconomyModel["models"].(map[string]any); ok {
		if resM, ok := models[fmt.Sprintf("%d", resourceID)].(map[string]any); ok {
			if st1, ok := resM["state_1"].(map[string]any); ok {
				cost = floatFromAny(st1["modeledProductionCostPerUnit"])
				wages = floatFromAny(st1["modeledStoreWages"])
				sales = floatFromAny(st1["modeledUnitsSoldAnHour"])
			}
		}
	}
	for _, r := range s.Data.Resources {
		if intFromAny(r["dbLetter"]) != resourceID {
			continue
		}
		if u, ok := r["unitsSoldAnHour"].(float64); ok && u > 0 {
			sales = u
		}
		break
	}
	if sales < 1 {
		sales = 100
	}
	// Apply supply/demand pressure to cost (±20% max)
	pressure := s.State.MarketPressure[resourceID]
	cost = cost * (1 + pressure*0.2)
	wpu := wages / sales
	if cost <= 0 {
		cost = s.Cfg.Game.BotOrderBase + float64(resourceID%7)
	}
	// Terminal price: production cost + 30% margin + per-unit wages
	terminal := cost*1.30 + wpu
	// Gross margin to split across 3 tiers
	gross := terminal - cost - wpu
	if gross < 0 {
		gross = 0
	}
	// Equal split with ±10% random variance so no two resources are identical
	variance := 0.85 + rand.Float64()*0.30 // 0.85-1.15
	baseShare := gross / 3.0 * variance
	producer := cost + baseShare
	processor := cost + baseShare*2 + wpu*0.5
	retailer := cost + baseShare*3 + wpu
	return map[string]float64{
		"terminalPrice":  math.Round(retailer*100) / 100,
		"producerPrice":  math.Round(producer*100) / 100,
		"processorPrice": math.Round(processor*100) / 100,
		"retailerPrice":  math.Round(retailer*100) / 100,
		"productionCost": math.Round(cost*100) / 100,
		"wagesPerUnit":   math.Round(wpu*100) / 100,
		"tierProfit":     math.Round(baseShare*100) / 100,
	}
}
