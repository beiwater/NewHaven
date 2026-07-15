package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

const DevJWTSigningKey = "dev-secret-change-in-production"

type Config struct {
	// Server
	Addr string

	// Auth
	JWTSigningKey string

	// Database
	DatabaseURL string

	// Rate limiting
	RateLimitEnabled bool

	// Dev mode (creates dev user/company)
	DevMode     bool
	DevPassword string

	// Game tuning (from game.json)
	Game *GameConfig
}

type GameConfig struct {
	BondFaceValue        float64  `json:"bond_face_value"`
	ExchangeFeePct       float64  `json:"exchange_fee_pct"`
	BotReplacementRate   float64  `json:"bot_replacement_rate"`
	BondMinInterest      float64  `json:"bond_min_interest"`
	BondMaxInterest      float64  `json:"bond_max_interest"`
	ProductionMod        float64  `json:"production_mod"`
	AdminOverheadBase    float64  `json:"admin_overhead_base"`
	BaseBuildingCost     float64  `json:"base_building_cost"`
	WarehouseBaseCap     int      `json:"warehouse_base_cap"`
	BaseProductionSlots  int      `json:"base_production_slots"`
	WarehouseUpgradeCost float64  `json:"warehouse_upgrade_cost"`
	BaseOutput           float64  `json:"base_output"`
	MaxBuildings         int      `json:"max_buildings"`
	ResearchBaseCost     float64  `json:"research_base_cost"`
	ResearchCostGrowth   float64  `json:"research_cost_growth"`
	ResearchSpeedBonus   float64  `json:"research_speed_bonus"`
	NewbieLevelUpTo      int      `json:"newbie_level_up_to"`
	MaxMessageLength     int      `json:"max_message_length"`
	ImageHostAllowlist   []string `json:"image_host_allowlist"`
}

func Load() *Config {
	return &Config{
		Addr:             envStr("SIM_API_ADDR", ":8088"),
		JWTSigningKey:    envStr("SIM_API_JWT_SECRET", DevJWTSigningKey),
		DatabaseURL:      os.Getenv("SIM_API_DATABASE_URL"),
		RateLimitEnabled: envStr("SIM_API_RATE_LIMIT", "false") == "true",
		DevMode:          envStr("SIM_API_DEV_MODE", "true") == "true",
		DevPassword:      envStr("SIM_API_DEV_PASSWORD", "123"),
		Game:             loadGameConfig(FindProjectRoot()),
	}
}

// Validate validates the configuration, collecting all errors via errors.Join.
func (c *Config) Validate() error {
	var errs []error
	if c.JWTSigningKey == "" {
		errs = append(errs, errors.New("SIM_API_JWT_SECRET is required"))
	}
	if !c.DevMode && c.JWTSigningKey == DevJWTSigningKey {
		errs = append(errs, errors.New("SIM_API_JWT_SECRET must be changed when dev mode is disabled"))
	}
	// Game config validation
	if c.Game != nil {
		if c.Game.ExchangeFeePct < 0 || c.Game.ExchangeFeePct > 100 {
			errs = append(errs, errors.New("exchange_fee_pct must be between 0 and 100"))
		}
		if c.Game.BondMinInterest < 0 {
			errs = append(errs, errors.New("bond_min_interest must be >= 0"))
		}
		if c.Game.BaseOutput <= 0 {
			errs = append(errs, errors.New("base_output must be positive"))
		}
		if c.Game.MaxBuildings <= 0 {
			errs = append(errs, errors.New("max_buildings must be > 0"))
		}
		if c.Game.MaxMessageLength <= 0 {
			errs = append(errs, errors.New("max_message_length must be > 0"))
		}
	}
	return errors.Join(errs...)
}

func FindProjectRoot() string {
	wd, _ := os.Getwd()
	root := wd
	for {
		if _, err := os.Stat(filepath.Join(root, "decompiled", "data")); err == nil {
			return root
		}
		parent := filepath.Dir(root)
		if parent == root {
			return wd
		}
		root = parent
	}
}

func loadGameConfig(root string) *GameConfig {
	path := filepath.Join(root, "backend", "configs", "game.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		return defaultGameConfig()
	}
	var cfg GameConfig
	if err := json.Unmarshal(raw, &cfg); err != nil {
		return defaultGameConfig()
	}
	return &cfg
}

func defaultGameConfig() *GameConfig {
	return &GameConfig{
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
		ResearchBaseCost:     1000,
		NewbieLevelUpTo:      7,
		MaxMessageLength:     500,
		ResearchCostGrowth:   1.2,
		ResearchSpeedBonus:   0.002,
		ImageHostAllowlist: []string{
			// 国内图床
			"imgse.com", "www.imgse.com",
			"imgchr.com", "www.imgchr.com",
			"superbed.cn", "www.superbed.cn",
			"superbed.cc", "www.superbed.cc",
			"imgurl.org", "www.imgurl.org",
			"imgbed.cn", "www.imgbed.cn",
			"img.st", "www.img.st",
			// 海外图床
			"postimages.org", "postimg.cc",
			"i.postimg.cc",
			"imgbb.com", "ibb.co", "i.ibb.co",
			"freeimage.host", "iili.io",
			"imgbox.com", "images2.imgbox.com",
			"thumbs2.imgbox.com",
			"catbox.moe", "files.catbox.moe",
		},
	}
}

func envStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
