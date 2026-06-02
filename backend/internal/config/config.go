package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	// Deployment (env only)
	Addr        string
	DataDir     string
	Debug       bool
	DBPath      string
	DatabaseURL string
	CSRFToken   string

	// Feature flags (env only)
	ACEnabled           bool
	AMLEnabled          bool
	ScriptDetectEnabled bool
	BotEnabled          bool

	// Game tuning (from game.json)
	Game *GameConfig
}

type GameConfig struct {
	CompanyID             int     `json:"company_id"`
	CompanyName           string  `json:"company_name"`
	StartMoney            float64 `json:"start_money"`
	StartLevel            int     `json:"start_level"`
	Bot1ID                int     `json:"bot1_id"`
	Bot2ID                int     `json:"bot2_id"`
	Bot1Name              string  `json:"bot1_name"`
	Bot2Name              string  `json:"bot2_name"`
	BotMoney              float64 `json:"bot_money"`
	BotLevel              int     `json:"bot_level"`
	ExchangeFeePct        float64 `json:"exchange_fee_pct"`
	AdminOverheadBase     float64 `json:"admin_overhead_base"`
	BondFaceValue         float64 `json:"bond_face_value"`
	BondMinInterest       float64 `json:"bond_min_interest"`
	BondMaxInterest       float64 `json:"bond_max_interest"`
	MaxBotOrders          int     `json:"max_bot_orders"`
	MaxLedgerEntries      int     `json:"max_ledger_entries"`
	WeatherSpeedMult      float64 `json:"weather_speed_mult"`
	ProductionMod         float64 `json:"production_mod"`
	GovBidRefundRate      float64 `json:"gov_bid_refund_rate"`
	BotCycleAmplitude     float64 `json:"bot_cycle_amplitude"`
	BotSpread             float64 `json:"bot_spread"`
	BotOrderQty           int     `json:"bot_order_qty"`
	BotResources          string  `json:"bot_resources"`
	BotOrderBase          float64 `json:"bot_order_base"`
	BaseBuildingCost      int     `json:"base_building_cost"`
	WarehouseBaseCap      int     `json:"warehouse_base_cap"`
	WarehouseUpgradeCost  float64 `json:"warehouse_upgrade_cost"`
	MaxQuality            int     `json:"max_quality"`
	QualitySalesFactor    float64 `json:"quality_sales_factor"`
	QualityResearchCost   float64 `json:"quality_research_cost"`
	DailyOrderCount       int     `json:"daily_order_count"`
	DailyOrderRewardBase  float64 `json:"daily_order_reward_base"`
	DailyOrderXPBase      int     `json:"daily_order_xp_base"`
	BaseProductionSlots   int     `json:"base_production_slots"`
	SlotUpgradeCost       float64 `json:"slot_upgrade_cost"`
	MarketLockThreshold   float64 `json:"market_lock_threshold"`
	MarketLockCapPct      float64 `json:"market_lock_cap_pct"`
	NationalTeamVolumePct float64 `json:"national_team_volume_pct"`
	NationalTeamPricePct  float64 `json:"national_team_price_pct"`
	BotReplacementRate    float64 `json:"bot_replacement_rate"`
}

func Load() *Config {
	cfg := &Config{
		Addr:                envStr("SIM_API_ADDR", "127.0.0.1:8088"),
		DataDir:             envStr("SIM_API_DATA_DIR", "decompiled/data"),
		Debug:               os.Getenv("SIM_API_DEBUG") == "1",
		DBPath:              envStr("SIM_API_DB_PATH", ""),
		DatabaseURL:         os.Getenv("SIM_API_DATABASE_URL"),
		CSRFToken:           envStr("SIM_API_CSRF_TOKEN", "dev-csrf-token"),
		ACEnabled:           os.Getenv("SIM_API_AC_ENABLED") != "0",
		AMLEnabled:          os.Getenv("SIM_API_AML_ENABLED") != "0",
		ScriptDetectEnabled: os.Getenv("SIM_API_SCRIPT_DETECT_ENABLED") != "0",
		BotEnabled:          os.Getenv("SIM_API_BOT_ENABLED") != "0",
	}
	// Load game config from JSON
	cfg.Game = loadGameConfig(cfg.DataDir)
	return cfg
}

func loadGameConfig(dataDir string) *GameConfig {
	g := &GameConfig{}
	// Try configs/ dir from project root (up 1 level from decompiled/data -> decompiled, then up another -> project root)
	candidates := []string{
		filepath.Join(dataDir, "..", "..", "configs", "game.json"),
		filepath.Join(dataDir, "configs", "game.json"),
		filepath.Join(dataDir, "game.json"),
	}
	for _, p := range candidates {
		raw, err := os.ReadFile(p)
		if err == nil {
			if err := json.Unmarshal(raw, g); err == nil {
				return g
			}
		}
	}
	return defaultGameConfig()
}

func defaultGameConfig() *GameConfig {
	return &GameConfig{
		CompanyID: 1234567, CompanyName: "Example Company Inc", StartMoney: 200000, StartLevel: 42,
		Bot1ID: 900001, Bot2ID: 900002, Bot1Name: "Atlas Trading Bot", Bot2Name: "Nova Market Bot",
		BotMoney: 5000000, BotLevel: 99,
		ExchangeFeePct: 0.04, AdminOverheadBase: 1.35, BondFaceValue: 5000,
		BondMinInterest: 0.5, BondMaxInterest: 2.0,
		MaxBotOrders: 600, MaxLedgerEntries: 5000,
		WeatherSpeedMult: 1.06, ProductionMod: 1.02, GovBidRefundRate: 0.8,
		BotCycleAmplitude: 0.06, BotSpread: 0.05, BotOrderQty: 200,
		BotResources: "1,2,3,4,5,6,7,8,9,66,72,115,116,117,120,121,122,127,133,134,135,137,139,141", BotOrderBase: 8.0,
		BaseBuildingCost: 50000, WarehouseBaseCap: 1000, WarehouseUpgradeCost: 25000,
		MaxQuality: 100, QualitySalesFactor: 0.0833, QualityResearchCost: 5000,
		DailyOrderCount: 5, DailyOrderRewardBase: 1000, DailyOrderXPBase: 50,
		BaseProductionSlots: 3, SlotUpgradeCost: 50000,
		MarketLockThreshold:   0.05,
		MarketLockCapPct:      1.2,
		NationalTeamVolumePct: 0.3,
		NationalTeamPricePct:  1.5,
		BotReplacementRate:    0.3,
	}
}

func envStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
