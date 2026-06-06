package config

import "testing"

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
