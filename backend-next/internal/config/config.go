package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Config struct {
	// Server
	Addr string

	// Auth
	JWTSigningKey string

	// Database
	DatabaseURL string

	// Dev mode (creates dev user/company)
	DevMode bool

	// Game tuning (from game.json)
	Game *GameConfig
}

type GameConfig struct {
	BondFaceValue         float64 `json:"bond_face_value"`
	ExchangeFeePct        float64 `json:"exchange_fee_pct"`
	BotReplacementRate    float64 `json:"bot_replacement_rate"`
	BondMinInterest       float64 `json:"bond_min_interest"`
	BondMaxInterest       float64 `json:"bond_max_interest"`
	ProductionMod         float64 `json:"production_mod"`
	AdminOverheadBase     float64 `json:"admin_overhead_base"`
}

func Load() *Config {
	return &Config{
		Addr:          envStr("SIM_API_ADDR", ":8088"),
		JWTSigningKey: envStr("SIM_API_JWT_SECRET", "dev-secret-change-in-production"),
		DatabaseURL:   os.Getenv("SIM_API_DATABASE_URL"),
		DevMode:       envStr("SIM_API_DEV_MODE", "true") == "true",
		Game:          loadGameConfig(findProjectRoot()),
	}
}

func findProjectRoot() string {
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
		BondFaceValue:      5000,
		ExchangeFeePct:     0.04,
		BotReplacementRate: 0.33,
		BondMinInterest:    0.5,
		BondMaxInterest:    2.0,
		ProductionMod:      1.0,
		AdminOverheadBase:  1.35,
	}
}

func envStr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
