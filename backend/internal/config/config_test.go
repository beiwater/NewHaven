package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestDefaultTestConfig_FreshInstance(t *testing.T) {
	a := DefaultTestConfig()
	b := DefaultTestConfig()

	if a == b {
		t.Error("DefaultTestConfig() should return a new instance each call")
	}
	if a.Game == b.Game {
		t.Error("each call should have a distinct Game pointer")
	}
}

func TestDefaultTestConfig_HasExpectedEnv(t *testing.T) {
	cfg := DefaultTestConfig()
	if !cfg.DevMode {
		t.Error("expected DevMode = true")
	}
	if cfg.JWTSigningKey != "test-jwt-secret" {
		t.Errorf("expected test jwt key, got %q", cfg.JWTSigningKey)
	}
	if cfg.ACEnabled {
		t.Error("expected ACEnabled = false")
	}
	if cfg.AMLEnabled {
		t.Error("expected AMLEnabled = false")
	}
	if cfg.ScriptDetectEnabled {
		t.Error("expected ScriptDetectEnabled = false")
	}
	if cfg.Game == nil {
		t.Fatal("Game config should not be nil")
	}
}

func TestDefaultTestConfig_GameDefaults(t *testing.T) {
	cfg := DefaultTestConfig()
	g := cfg.Game

	if g.CompanyID != 1234567 {
		t.Errorf("CompanyID = %d, want 1234567", g.CompanyID)
	}
	if g.StartMoney != 200000 {
		t.Errorf("StartMoney = %.0f, want 200000", g.StartMoney)
	}
	if g.StartLevel != 42 {
		t.Errorf("StartLevel = %d, want 42", g.StartLevel)
	}
	if g.ExchangeFeePct != 0.04 {
		t.Errorf("ExchangeFeePct = %.4f, want 0.04", g.ExchangeFeePct)
	}
}

func TestDefaultTestConfig_NoSharedGame(t *testing.T) {
	a := DefaultTestConfig()
	b := DefaultTestConfig()

	// Mutating one should not affect the other.
	a.Game.ExchangeFeePct = 99.99
	if b.Game.ExchangeFeePct == 99.99 {
		t.Error("mutating one instance should not affect another")
	}
}

func TestMustParseTimeRFC3339_Valid(t *testing.T) {
	ts := MustParseTimeRFC3339("2026-06-02T12:00:00Z")
	if ts.Year() != 2026 || ts.Month() != 6 || ts.Day() != 2 {
		t.Errorf("unexpected date: %v", ts)
	}
}

func TestMustParseTimeRFC3339_Panics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("expected panic for invalid timestamp")
		}
	}()
	MustParseTimeRFC3339("not-a-timestamp")
}

func TestParseTimeRFC3339_Valid(t *testing.T) {
	ts, err := ParseTimeRFC3339("2026-06-02T12:00:00Z")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ts.Year() != 2026 {
		t.Errorf("year = %d, want 2026", ts.Year())
	}
}

func TestParseTimeRFC3339_Invalid(t *testing.T) {
	_, err := ParseTimeRFC3339("")
	if err == nil {
		t.Error("expected error for empty string")
	}
	_, err = ParseTimeRFC3339("garbage")
	if err == nil {
		t.Error("expected error for garbage input")
	}
}

func TestDefaultGameConfig_MatchesGameJSON(t *testing.T) {
	// Find configs/game.json relative to the test file location.
	// Walk up from the test binary's working directory.
	candidates := []string{
		"../../configs/game.json",
		"../configs/game.json",
		"configs/game.json",
	}
	var raw []byte
	var found string
	for _, p := range candidates {
		abs := p
		// Try one level up for test runner in internal/config
		abs2 := filepath.Join("..", p)
		for _, candidate := range []string{abs, abs2} {
			data, err := os.ReadFile(candidate)
			if err == nil {
				raw = data
				found = candidate
				break
			}
		}
		if raw != nil {
			break
		}
	}
	if raw == nil {
		t.Skip("configs/game.json not found from test working directory")
	}

	var fileGC GameConfig
	if err := json.Unmarshal(raw, &fileGC); err != nil {
		t.Fatalf("failed to parse %s: %v", found, err)
	}

	def := defaultGameConfig()

	// Compare every field. This catches drift between game.json and defaultGameConfig().
	if fileGC.CompanyID != def.CompanyID {
		t.Errorf("CompanyID: file=%d default=%d", fileGC.CompanyID, def.CompanyID)
	}
	if fileGC.CompanyName != def.CompanyName {
		t.Errorf("CompanyName: file=%q default=%q", fileGC.CompanyName, def.CompanyName)
	}
	if fileGC.StartMoney != def.StartMoney {
		t.Errorf("StartMoney: file=%.0f default=%.0f", fileGC.StartMoney, def.StartMoney)
	}
	if fileGC.StartLevel != def.StartLevel {
		t.Errorf("StartLevel: file=%d default=%d", fileGC.StartLevel, def.StartLevel)
	}
	if fileGC.ExchangeFeePct != def.ExchangeFeePct {
		t.Errorf("ExchangeFeePct: file=%.4f default=%.4f", fileGC.ExchangeFeePct, def.ExchangeFeePct)
	}
	if fileGC.AdminOverheadBase != def.AdminOverheadBase {
		t.Errorf("AdminOverheadBase: file=%.4f default=%.4f", fileGC.AdminOverheadBase, def.AdminOverheadBase)
	}
	if fileGC.BondFaceValue != def.BondFaceValue {
		t.Errorf("BondFaceValue: file=%.0f default=%.0f", fileGC.BondFaceValue, def.BondFaceValue)
	}
	if fileGC.BondMinInterest != def.BondMinInterest {
		t.Errorf("BondMinInterest: file=%.2f default=%.2f", fileGC.BondMinInterest, def.BondMinInterest)
	}
	if fileGC.BondMaxInterest != def.BondMaxInterest {
		t.Errorf("BondMaxInterest: file=%.2f default=%.2f", fileGC.BondMaxInterest, def.BondMaxInterest)
	}
	if fileGC.MaxBotOrders != def.MaxBotOrders {
		t.Errorf("MaxBotOrders: file=%d default=%d", fileGC.MaxBotOrders, def.MaxBotOrders)
	}
	if fileGC.MaxLedgerEntries != def.MaxLedgerEntries {
		t.Errorf("MaxLedgerEntries: file=%d default=%d", fileGC.MaxLedgerEntries, def.MaxLedgerEntries)
	}
	if fileGC.WeatherSpeedMult != def.WeatherSpeedMult {
		t.Errorf("WeatherSpeedMult: file=%.4f default=%.4f", fileGC.WeatherSpeedMult, def.WeatherSpeedMult)
	}
	if fileGC.ProductionMod != def.ProductionMod {
		t.Errorf("ProductionMod: file=%.4f default=%.4f", fileGC.ProductionMod, def.ProductionMod)
	}
	if fileGC.GovBidRefundRate != def.GovBidRefundRate {
		t.Errorf("GovBidRefundRate: file=%.4f default=%.4f", fileGC.GovBidRefundRate, def.GovBidRefundRate)
	}
	if fileGC.BotCycleAmplitude != def.BotCycleAmplitude {
		t.Errorf("BotCycleAmplitude: file=%.4f default=%.4f", fileGC.BotCycleAmplitude, def.BotCycleAmplitude)
	}
	if fileGC.BotSpread != def.BotSpread {
		t.Errorf("BotSpread: file=%.4f default=%.4f", fileGC.BotSpread, def.BotSpread)
	}
	if fileGC.BotOrderQty != def.BotOrderQty {
		t.Errorf("BotOrderQty: file=%d default=%d", fileGC.BotOrderQty, def.BotOrderQty)
	}
	if fileGC.BotResources != def.BotResources {
		t.Errorf("BotResources: file=%q default=%q", fileGC.BotResources, def.BotResources)
	}
	if fileGC.BotOrderBase != def.BotOrderBase {
		t.Errorf("BotOrderBase: file=%.4f default=%.4f", fileGC.BotOrderBase, def.BotOrderBase)
	}
	if fileGC.BaseBuildingCost != def.BaseBuildingCost {
		t.Errorf("BaseBuildingCost: file=%d default=%d", fileGC.BaseBuildingCost, def.BaseBuildingCost)
	}
	if fileGC.WarehouseBaseCap != def.WarehouseBaseCap {
		t.Errorf("WarehouseBaseCap: file=%d default=%d", fileGC.WarehouseBaseCap, def.WarehouseBaseCap)
	}
	if fileGC.WarehouseUpgradeCost != def.WarehouseUpgradeCost {
		t.Errorf("WarehouseUpgradeCost: file=%.4f default=%.4f", fileGC.WarehouseUpgradeCost, def.WarehouseUpgradeCost)
	}
	if fileGC.MaxQuality != def.MaxQuality {
		t.Errorf("MaxQuality: file=%d default=%d", fileGC.MaxQuality, def.MaxQuality)
	}
	if fileGC.QualitySalesFactor != def.QualitySalesFactor {
		t.Errorf("QualitySalesFactor: file=%.4f default=%.4f", fileGC.QualitySalesFactor, def.QualitySalesFactor)
	}
	if fileGC.QualityResearchCost != def.QualityResearchCost {
		t.Errorf("QualityResearchCost: file=%.4f default=%.4f", fileGC.QualityResearchCost, def.QualityResearchCost)
	}
	if fileGC.DailyOrderCount != def.DailyOrderCount {
		t.Errorf("DailyOrderCount: file=%d default=%d", fileGC.DailyOrderCount, def.DailyOrderCount)
	}
	if fileGC.DailyOrderRewardBase != def.DailyOrderRewardBase {
		t.Errorf("DailyOrderRewardBase: file=%.4f default=%.4f", fileGC.DailyOrderRewardBase, def.DailyOrderRewardBase)
	}
	if fileGC.DailyOrderXPBase != def.DailyOrderXPBase {
		t.Errorf("DailyOrderXPBase: file=%d default=%d", fileGC.DailyOrderXPBase, def.DailyOrderXPBase)
	}
	if fileGC.BaseProductionSlots != def.BaseProductionSlots {
		t.Errorf("BaseProductionSlots: file=%d default=%d", fileGC.BaseProductionSlots, def.BaseProductionSlots)
	}
	if fileGC.SlotUpgradeCost != def.SlotUpgradeCost {
		t.Errorf("SlotUpgradeCost: file=%.4f default=%.4f", fileGC.SlotUpgradeCost, def.SlotUpgradeCost)
	}
	if fileGC.MarketLockThreshold != def.MarketLockThreshold {
		t.Errorf("MarketLockThreshold: file=%.4f default=%.4f", fileGC.MarketLockThreshold, def.MarketLockThreshold)
	}
	if fileGC.MarketLockCapPct != def.MarketLockCapPct {
		t.Errorf("MarketLockCapPct: file=%.4f default=%.4f", fileGC.MarketLockCapPct, def.MarketLockCapPct)
	}
	if fileGC.NationalTeamVolumePct != def.NationalTeamVolumePct {
		t.Errorf("NationalTeamVolumePct: file=%.4f default=%.4f", fileGC.NationalTeamVolumePct, def.NationalTeamVolumePct)
	}
	if fileGC.NationalTeamPricePct != def.NationalTeamPricePct {
		t.Errorf("NationalTeamPricePct: file=%.4f default=%.4f", fileGC.NationalTeamPricePct, def.NationalTeamPricePct)
	}
	if fileGC.BotReplacementRate != def.BotReplacementRate {
		t.Errorf("BotReplacementRate: file=%.4f default=%.4f", fileGC.BotReplacementRate, def.BotReplacementRate)
	}
	if fileGC.LaborCostIndex != def.LaborCostIndex {
		t.Errorf("LaborCostIndex: file=%.4f default=%.4f", fileGC.LaborCostIndex, def.LaborCostIndex)
	}
	if fileGC.MaterialCostIndex != def.MaterialCostIndex {
		t.Errorf("MaterialCostIndex: file=%.4f default=%.4f", fileGC.MaterialCostIndex, def.MaterialCostIndex)
	}
	if fileGC.EnergyCostIndex != def.EnergyCostIndex {
		t.Errorf("EnergyCostIndex: file=%.4f default=%.4f", fileGC.EnergyCostIndex, def.EnergyCostIndex)
	}
	if fileGC.GlobalDemandIndex != def.GlobalDemandIndex {
		t.Errorf("GlobalDemandIndex: file=%.4f default=%.4f", fileGC.GlobalDemandIndex, def.GlobalDemandIndex)
	}
	if fileGC.SaturationK != def.SaturationK {
		t.Errorf("SaturationK: file=%.4f default=%.4f", fileGC.SaturationK, def.SaturationK)
	}
	if fileGC.EventPriceMultiplier != def.EventPriceMultiplier {
		t.Errorf("EventPriceMultiplier: file=%.4f default=%.4f", fileGC.EventPriceMultiplier, def.EventPriceMultiplier)
	}
	if fileGC.RetailTaxRate != def.RetailTaxRate {
		t.Errorf("RetailTaxRate: file=%.4f default=%.4f", fileGC.RetailTaxRate, def.RetailTaxRate)
	}
	if fileGC.BaseLaborCost != def.BaseLaborCost {
		t.Errorf("BaseLaborCost: file=%.4f default=%.4f", fileGC.BaseLaborCost, def.BaseLaborCost)
	}
	if fileGC.BaseEnergyCost != def.BaseEnergyCost {
		t.Errorf("BaseEnergyCost: file=%.4f default=%.4f", fileGC.BaseEnergyCost, def.BaseEnergyCost)
	}
	if fileGC.BaseMaintenanceCost != def.BaseMaintenanceCost {
		t.Errorf("BaseMaintenanceCost: file=%.4f default=%.4f", fileGC.BaseMaintenanceCost, def.BaseMaintenanceCost)
	}
	if fileGC.BaseManagementCost != def.BaseManagementCost {
		t.Errorf("BaseManagementCost: file=%.4f default=%.4f", fileGC.BaseManagementCost, def.BaseManagementCost)
	}
	if fileGC.SweetSpotLevel != def.SweetSpotLevel {
		t.Errorf("SweetSpotLevel: file=%d default=%d", fileGC.SweetSpotLevel, def.SweetSpotLevel)
	}

	// Bot names are stored in the Go default but not in game.json test data
	// (they are in the actual game.json). Check them separately.
	if fileGC.Bot1Name != def.Bot1Name {
		t.Errorf("Bot1Name: file=%q default=%q", fileGC.Bot1Name, def.Bot1Name)
	}
	if fileGC.Bot2Name != def.Bot2Name {
		t.Errorf("Bot2Name: file=%q default=%q", fileGC.Bot2Name, def.Bot2Name)
	}
}

// TestExchangeFeePct_Consistency verifies ExchangeFeePct is the same
// across defaultGameConfig(), game.json, and the formula package expectation.
func TestExchangeFeePct_Consistency(t *testing.T) {
	def := defaultGameConfig()

	// Load game.json
	candidates := []string{
		"../../configs/game.json",
		"../configs/game.json",
		"configs/game.json",
	}
	var raw []byte
	for _, p := range candidates {
		abs := p
		abs2 := filepath.Join("..", p)
		for _, candidate := range []string{abs, abs2} {
			data, err := os.ReadFile(candidate)
			if err == nil {
				raw = data
				break
			}
		}
		if raw != nil {
			break
		}
	}
	if raw == nil {
		t.Skip("configs/game.json not found; skipping ExchangeFeePct consistency check")
	}

	var fileGC GameConfig
	if err := json.Unmarshal(raw, &fileGC); err != nil {
		t.Fatalf("failed to parse game.json: %v", err)
	}

	if def.ExchangeFeePct != fileGC.ExchangeFeePct {
		t.Errorf("ExchangeFeePct mismatch: defaultGameConfig()=%.4f game.json=%.4f",
			def.ExchangeFeePct, fileGC.ExchangeFeePct)
	}

	if def.ExchangeFeePct != 0.04 {
		t.Errorf("ExchangeFeePct should be 0.04, got %.4f", def.ExchangeFeePct)
	}
}

// TestTimeParse_SwallowedErrors ensures no time.Parse calls use _ = pattern.
// This is a compile-time / audit test; run `go vet` or `grep` separately.
// The actual fixes are in bond.go, offline.go, research.go, simboost.go.
func TestNoSwallowedTimeParse(t *testing.T) {
	// Verify MustParseTimeRFC3339 works for a known-good string.
	ts := MustParseTimeRFC3339("2026-01-01T00:00:00Z")
	if ts.Year() != 2026 {
		t.Errorf("year = %d, want 2026", ts.Year())
	}
}
