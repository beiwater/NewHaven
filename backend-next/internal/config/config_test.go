package config

import (
	"strings"
	"testing"
)

func TestValidateRejectsDefaultJWTSecretOutsideDevMode(t *testing.T) {
	cfg := &Config{JWTSigningKey: DevJWTSigningKey, DevMode: false}
	if err := cfg.Validate(); err == nil {
		t.Fatal("expected default JWT secret to be rejected outside dev mode")
	}
	cfg.JWTSigningKey = "production-secret"
	if err := cfg.Validate(); err != nil {
		t.Fatalf("expected custom JWT secret to pass: %v", err)
	}
}

func TestGameConfigValidation(t *testing.T) {
	t.Run("rejects invalid game config values", func(t *testing.T) {
		cfg := &Config{
			JWTSigningKey: "secret",
			Game: &GameConfig{
				ExchangeFeePct:   150,
				BondMinInterest: -1,
				BaseOutput:       0,
				MaxBuildings:     0,
				MaxMessageLength: -1,
			},
		}
		err := cfg.Validate()
		if err == nil {
			t.Fatal("expected validation errors, got nil")
		}
		errStr := err.Error()
		for _, want := range []string{
			"exchange_fee_pct must be between 0 and 100",
			"bond_min_interest must be >= 0",
			"base_output must be positive",
			"max_buildings must be > 0",
			"max_message_length must be > 0",
		} {
			if !strings.Contains(errStr, want) {
				t.Errorf("expected error to contain %q", want)
			}
		}
	})

	t.Run("passes valid game config", func(t *testing.T) {
		cfg := &Config{
			JWTSigningKey: "secret",
			Game: &GameConfig{
				ExchangeFeePct:  10,
				BondMinInterest: 0.5,
				BaseOutput:      100,
				MaxBuildings:    20,
				MaxMessageLength: 500,
			},
		}
		if err := cfg.Validate(); err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
	})

	t.Run("nil Game config skips game validation", func(t *testing.T) {
		cfg := &Config{
			JWTSigningKey: "secret",
			Game:          nil,
		}
		if err := cfg.Validate(); err != nil {
			t.Fatalf("expected no error with nil Game, got: %v", err)
		}
	})
}
