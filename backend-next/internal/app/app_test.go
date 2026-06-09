package app

import (
	"testing"

	"github.com/newhaven/backend-next/internal/catalog"
	"github.com/newhaven/backend-next/internal/config"
	"github.com/newhaven/backend-next/internal/storage/memory"
)

func TestNewDoesNotPanic(t *testing.T) {
	cfg := &config.Config{
		Addr:          ":8080",
		JWTSigningKey: "test-secret",
		DatabaseURL:   "",
		DevMode:       false,
		Game: &config.GameConfig{
			BondFaceValue:        5000,
			ExchangeFeePct:       0.04,
			BotReplacementRate:   0.33,
			BondMinInterest:      0.5,
			BondMaxInterest:      2.0,
			ProductionMod:        1.0,
			AdminOverheadBase:    1.35,
			BaseBuildingCost:     50000,
			BaseProductionSlots:  3,
			WarehouseBaseCap:     1000,
			WarehouseUpgradeCost: 25000,
			BaseOutput:           100,
			MaxBuildings:         20,
		},
	}
	st := memory.New()
	resources := make(map[int]*catalog.ResourceEntry)
	buildings := make(map[int]*catalog.BuildingEntry)
	economy := make(map[int]*catalog.EconomyModelEntry)

	// New should not panic on empty catalogs.
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("New panicked: %v", r)
			}
		}()
		_ = New(cfg, st, resources, buildings, economy)
	}()
}
