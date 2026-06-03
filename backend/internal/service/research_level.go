package service

func (s *Service) CurrentLevelInfo(companyID int) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return map[string]any{"error": "company not found"}
	}
	return map[string]any{
		"level":         company.Level,
		"currentXp":     s.State.XP,
		"xpToNextLevel": s.State.XpToNextLevel,
	}
}

func (s *Service) AddXP(amount int) map[string]any {
	s.addXP(nil, amount)
	return map[string]any{
		"ok": true, "xpAdded": amount,
	}
}

func (s *Service) LevelRewards(level int) map[string]any {
	rewards := []map[string]any{}
	switch {
	case level >= 50:
		rewards = append(rewards, map[string]any{"type": "title_change", "value": "Tycoon"})
		fallthrough
	case level >= 40:
		rewards = append(rewards, map[string]any{"type": "title_change", "value": "Magnate"})
		fallthrough
	case level >= 30:
		rewards = append(rewards, map[string]any{"type": "max_buildings", "value": 14})
		fallthrough
	case level >= 20:
		rewards = append(rewards, map[string]any{"type": "title_change", "value": "Director"})
		fallthrough
	case level >= 15:
		rewards = append(rewards, map[string]any{"type": "time_limit", "value": "48h"})
		fallthrough
	case level >= 10:
		rewards = append(rewards, map[string]any{"type": "title_change", "value": "Manager"})
		fallthrough
	case level >= 5:
		rewards = append(rewards, map[string]any{"type": "production_speed_bonus", "value": 3})
		fallthrough
	default:
		rewards = append(rewards, map[string]any{"type": "title_change", "value": "Entrepreneur"})
	}

	if level%5 == 0 && level > 0 {
		rewards = append(rewards, map[string]any{"type": "speed_bonus_points", "value": 3})
	}

	return map[string]any{"level": level, "rewards": rewards}
}
