package market

import (
	"context"
	"math"
	"time"

	domainmarket "github.com/beiwater/NewHaven/backend/internal/domain/market"
	"github.com/beiwater/NewHaven/backend/internal/formula"
)

const legacyRetailDemandPerHour = 30.0

// marketReferencePrice is the bot's moving economic centre. Processed goods
// inherit their live input costs, while a stable wage-and-profit contribution
// keeps the baseline producer profitable. A shared, deterministic market pulse
// and the public order-book imbalance then move the centre for everyone.
func (s *Service) marketReferencePrice(ctx context.Context, resourceID int, orders []domainmarket.MarketOrder) float64 {
	anchor := s.economicAnchorPrice(ctx, resourceID, make(map[int]bool))
	if anchor <= 0 {
		anchor = s.basePriceForResource(resourceID)
	}
	if anchor <= 0 {
		anchor = 20 + float64(resourceID%11)*3
	}
	pressure := s.marketPricePressure(resourceID, orders, s.clock.Now().UTC())
	return snapToMarketTick(anchor * (1 + pressure))
}

// economicAnchorPrice carries the cost of a recipe forward through the whole
// chain. It deliberately uses the current ticker of every input when one is
// available: a milk shortage therefore raises butter, cheese, pizza and cake
// reference prices without a hand-authored price table for every scenario.
func (s *Service) economicAnchorPrice(ctx context.Context, resourceID int, visiting map[int]bool) float64 {
	resource, ok := s.resources[resourceID]
	if !ok || resource == nil {
		return 0
	}
	if visiting[resourceID] {
		return resource.BasePrice
	}
	visiting[resourceID] = true
	defer delete(visiting, resourceID)

	inputCost := 0.0
	for inputID, quantity := range resource.ProducedFrom {
		inputPrice := 0.0
		if ticker, err := s.market.GetTicker(ctx, inputID); err == nil && ticker != nil && ticker.LastPrice > 0 {
			inputPrice = ticker.LastPrice
		}
		if inputPrice <= 0 {
			inputPrice = s.economicAnchorPrice(ctx, inputID, visiting)
		}
		inputCost += inputPrice * float64(quantity)
	}

	rate := float64(resource.ProducedPerHourRaw)
	kind := resource.PrimaryBuildingKind
	if rate <= 0 || kind <= 0 {
		return resource.BasePrice
	}
	producerMargin := (formula.BuildingHourlyWage(kind, 1) + formula.TargetBuildingProfitPerHour) / rate
	feeMultiplier := 1 - s.exchangeFeePct()
	if feeMultiplier <= 0 {
		feeMultiplier = 0.96
	}
	return (inputCost + producerMargin) / feeMultiplier
}

func (s *Service) marketPricePressure(resourceID int, orders []domainmarket.MarketOrder, now time.Time) float64 {
	cycle := 0.08*math.Sin(2*math.Pi*float64(now.Unix())/1800+float64(resourceID)*1.37) +
		0.04*math.Sin(2*math.Pi*float64(now.Unix())/7200+float64(resourceID)*2.11)
	return clampMarketValue(cycle+0.12*s.liveOrderImbalance(orders), -0.18, 0.18)
}

func (s *Service) retailDemandMultiplier(resourceID int, orders []domainmarket.MarketOrder, now time.Time) float64 {
	cycle := 0.13*math.Sin(2*math.Pi*float64(now.Unix())/1500+float64(resourceID)*0.91) +
		0.07*math.Sin(2*math.Pi*float64(now.Unix())/5400+float64(resourceID)*2.73)
	return clampMarketValue(1+cycle+0.35*s.liveOrderImbalance(orders), 0.55, 1.45)
}

// liveOrderImbalance intentionally ignores the liquidity bot. Bot orders keep
// the book usable; player orders are the signal that should alter consumer
// appetite and the moving economic centre.
func (s *Service) liveOrderImbalance(orders []domainmarket.MarketOrder) float64 {
	var bids, asks float64
	for _, order := range orders {
		if order.CompanyID == s.botCompanyID || (order.Status != domainmarket.StatusOpen && order.Status != domainmarket.StatusPartial) {
			continue
		}
		remaining := order.Remaining()
		if remaining <= 0 {
			continue
		}
		if order.IsBuy {
			bids += float64(remaining)
		} else {
			asks += float64(remaining)
		}
	}
	return clampMarketValue((bids-asks)/(bids+asks+200), -0.5, 0.5)
}

func clampMarketValue(value, lower, upper float64) float64 {
	return math.Max(lower, math.Min(upper, value))
}

func (s *Service) retailBaseUnitsPerHour(ctx context.Context, resourceID, level int, salesModifierPct float64) (float64, float64) {
	resource := s.resources[resourceID]
	if resource == nil {
		return 0, 0
	}
	baseDemand := resource.RetailDemandPerHour
	if baseDemand <= 0 {
		baseDemand = legacyRetailDemandPerHour
	}
	if level < 1 {
		level = 1
	}
	orders, _ := s.market.GetOrdersByResource(ctx, resourceID)
	demandMultiplier := s.retailDemandMultiplier(resourceID, orders, s.clock.Now().UTC())
	bonusMultiplier := math.Max(0, 1+salesModifierPct/100)
	return baseDemand * demandMultiplier * float64(level) * bonusMultiplier, demandMultiplier
}

// retailRecommendedPrice is deliberately separate from the wholesale market
// quote. It turns the current input value, worker payroll, live demand and the
// $300/hour target into a retail price for this specific building.
func (s *Service) retailRecommendedPrice(ctx context.Context, resourceID, buildingKind, level int, salesModifierPct float64) (price, demandMultiplier float64) {
	wholesale := s.salePriceForResource(ctx, resourceID)
	baseUnits, multiplier := s.retailBaseUnitsPerHour(ctx, resourceID, level, salesModifierPct)
	if baseUnits <= 0 || wholesale <= 0 {
		return wholesale, multiplier
	}
	if level < 1 {
		level = 1
	}
	targetProfit := formula.TargetBuildingProfitPerHour * float64(level)
	markup := (formula.BuildingHourlyWage(buildingKind, level) + targetProfit) / baseUnits
	return snapToMarketTick(wholesale + markup), multiplier
}
