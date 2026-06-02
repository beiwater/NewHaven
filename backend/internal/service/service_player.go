package service

import (
	"fmt"
	"go-sim-api/internal/middleware"
	"go-sim-api/internal/model"
	"time"
)

func (s *Service) RegisterPlayer(username string) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Check duplicate username
	for _, p := range s.State.Players {
		if p.Username == username {
			return nil, fmt.Errorf("username taken")
		}
	}
	if s.State.NextPlayerID <= 0 {
		s.State.NextPlayerID = 1
		for _, p := range s.State.Players {
			if p.ID >= s.State.NextPlayerID {
				s.State.NextPlayerID = p.ID + 1
			}
		}
	}
	playerID := s.State.NextPlayerID
	s.State.NextPlayerID++
	token := middleware.GenerateToken()
	companyID := 1000000 + playerID
	now := s.now().UTC().Format(time.RFC3339)
	player := model.Player{
		ID: playerID, Username: username, Token: token,
		CompanyID: companyID, RegisteredAt: now,
	}
	s.State.Players = append(s.State.Players, player)
	// Create a new company for this player
	company := model.Company{
		ID: companyID, Name: username + "'s Company",
		Money: s.Cfg.Game.StartMoney, Level: 1,
		Inventory:         map[int]int{1: 500, 2: 500},
		UnplacedBuildings: []map[string]any{},
	}
	s.State.Companies = append(s.State.Companies, company)
	s.saveStateLocked()
	return map[string]any{
		"player":    player,
		"company":   company,
		"companyID": companyID,
	}, nil
}

func (s *Service) LoginPlayer(username string) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range s.State.Players {
		if p.Username == username {
			for _, c := range s.State.Companies {
				if c.ID == p.CompanyID {
					return map[string]any{"player": p, "company": c}, nil
				}
			}
		}
	}
	return nil, fmt.Errorf("player not found")
}

// --- Write-through persistence helpers ---

func (s *Service) ValidateToken(token string) (int, int, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range s.State.Players {
		if p.Token == token {
			return p.ID, p.CompanyID, true
		}
	}
	return 0, 0, false
}
