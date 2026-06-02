package service

import (
	"fmt"
	"time"
)

func (s *Service) SimBoostTypes() []map[string]any {
	return []map[string]any{
		{"id": "boost-production", "name": "Production Boost", "desc": "+50% production speed for 30min", "duration": 1800},
		{"id": "boost-research", "name": "Research Boost", "desc": "+50% research speed for 30min", "duration": 1800},
	}
}

func (s *Service) SimBoostsUse() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now().UTC()
	active := []map[string]any{}
	if s.State.BoostEndsAt != "" {
		endsAt, _ := time.Parse(time.RFC3339, s.State.BoostEndsAt)
		if now.Before(endsAt) {
			active = append(active, map[string]any{
				"type":      s.State.BoostType,
				"endsAt":    s.State.BoostEndsAt,
				"remaining": endsAt.Sub(now).String(),
			})
		} else {
			s.State.BoostEndsAt = ""
			s.State.BoostMultiplier = 1.0
		}
	}
	return map[string]any{"remaining": 3, "active": active}
}

func (s *Service) UseSimBoost(boostID string) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.State.BoostEndsAt != "" {
		return nil, fmt.Errorf("boost already active")
	}
	duration := 30 * time.Minute
	mult := 1.5
	s.State.BoostType = boostID
	s.State.BoostEndsAt = s.now().UTC().Add(duration).Format(time.RFC3339)
	s.State.BoostMultiplier = mult
	s.addLedger("use_boost", 0, "out", map[string]any{"boostId": boostID})
	return map[string]any{"boostId": boostID, "endsAt": s.State.BoostEndsAt, "multiplier": mult}, nil
}
