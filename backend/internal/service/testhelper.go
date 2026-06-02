package service

import (
	"go-sim-api/internal/config"
	"go-sim-api/internal/data"
)

// NewTestService creates a *Service with minimal static data suitable for
// testing handler endpoints and other packages that depend on Service.
func NewTestService() *Service {
	cfg := config.DefaultTestConfig()
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
