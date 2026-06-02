package service

import (
	"go-sim-api/internal/config"
	"go-sim-api/internal/data"
)

// NewTestService creates a *Service with minimal static data suitable for
// testing handler endpoints and other packages that depend on Service.
func NewTestService() *Service {
	cfg := &config.Config{
		ACEnabled: false, AMLEnabled: false, ScriptDetectEnabled: false,
		Game: &config.GameConfig{
			CompanyID: 1, CompanyName: "Test Inc",
			StartMoney: 200000, StartLevel: 42,
			MaxQuality: 100, ExchangeFeePct: 0.04,
			MaxBotOrders: 600, AdminOverheadBase: 1.35,
			MaxLedgerEntries: 5000,
			Bot1ID:           900001, Bot2ID: 900002,
			BotMoney: 5000000, BotLevel: 99,
			BotOrderBase:     8.0,
			GovBidRefundRate: 0.8, BondFaceValue: 5000,
			QualitySalesFactor:   0.0833,
			DailyOrderCount:      5,
			DailyOrderRewardBase: 5000,
			DailyOrderXPBase:     50,
		},
	}
	d := &data.StaticData{
		Resources: []map[string]any{
			{"id": 1, "name": "Power", "dbLetter": 1, "producedPerHourRaw": 100.0},
			{"id": 2, "name": "Water", "dbLetter": 2, "producedPerHourRaw": 80.0},
			{"id": 8, "name": "Beef", "dbLetter": 8, "producedPerHourRaw": 12.0, "producedFrom": map[string]any{"2": 3}},
		},
		EconomyModel: map[string]any{"models": map[string]any{}},
	}
	return New(d, cfg, nil)
}
