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
			{"id": 1, "name": "Wheat", "dbLetter": 1, "producedPerHourRaw": 100.0},
			{"id": 2, "name": "Flour", "dbLetter": 2, "producedPerHourRaw": 80.0, "producedFrom": map[string]any{"1": 2}},
			{"id": 3, "name": "Bread", "dbLetter": 3, "producedPerHourRaw": 40.0, "producedFrom": map[string]any{"2": 2}},
			{"id": 4, "name": "Meals", "dbLetter": 4, "producedPerHourRaw": 30.0, "producedFrom": map[string]any{"3": 2}},
		},
		EconomyModel: map[string]any{"models": map[string]any{}},
	}
	return New(d, cfg, nil)
}
