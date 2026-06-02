package service

import (
	"fmt"
	"go-sim-api/internal/middleware"
	"go-sim-api/internal/model"
	"time"

	"golang.org/x/crypto/bcrypt"
)

func (s *Service) RegisterPlayer(username, password string) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check duplicate username
	for _, p := range s.State.Players {
		if p.Username == username {
			return nil, fmt.Errorf("username taken")
		}
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("password hash failed: %w", err)
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
	companyID := 1000000 + playerID
	now := s.now().UTC().Format(time.RFC3339)

	player := model.Player{
		ID:           playerID,
		Username:     username,
		PasswordHash: string(hash),
		CompanyID:    companyID,
		RegisteredAt: now,
	}

	// Sign JWT
	token, err := middleware.SignJWT(playerID, companyID, s.Cfg.JWTSigningKey)
	if err != nil {
		return nil, fmt.Errorf("sign token: %w", err)
	}
	player.Token = token
	s.State.Players = append(s.State.Players, player)

	// Create a new company for this player
	company := model.Company{
		ID: companyID, Name: username + "'s Company",
		Money: s.Cfg.Game.StartMoney, Level: 1,
		Inventory:         map[int]int{1: 500},
		UnplacedBuildings: []map[string]any{},
	}
	s.State.Companies = append(s.State.Companies, company)
	s.saveStateLocked()
	return map[string]any{
		"player":    player,
		"company":   company,
		"token":     token,
		"companyId": companyID,
		"companyID": companyID,
	}, nil
}

func (s *Service) LoginPlayer(username, password string) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.State.Players {
		p := &s.State.Players[i]
		if p.Username != username {
			continue
		}
		// Verify password
		if err := bcrypt.CompareHashAndPassword([]byte(p.PasswordHash), []byte(password)); err != nil {
			return nil, fmt.Errorf("invalid username or password")
		}
		// Sign a fresh JWT
		token, err := middleware.SignJWT(p.ID, p.CompanyID, s.Cfg.JWTSigningKey)
		if err != nil {
			return nil, fmt.Errorf("sign token: %w", err)
		}
		p.Token = token
		s.saveStateLocked()
		for _, c := range s.State.Companies {
			if c.ID == p.CompanyID {
				return map[string]any{
					"player":    p,
					"company":   c,
					"token":     token,
					"companyId": p.CompanyID,
				}, nil
			}
		}
	}
	return nil, fmt.Errorf("invalid username or password")
}

// ValidateToken is no longer used — JWT parsing is done in the handler layer.
// Kept for data migration: scans old tokens stored in Player records.
// New registrations and logins use JWT only.
func (s *Service) ValidateToken(token string) (int, int, bool) {
	// First try JWT parsing
	if pid, cid, err := middleware.ParseJWT(token, s.Cfg.JWTSigningKey); err == nil {
		return pid, cid, true
	}
	// Fallback: scan legacy tokens from in-memory player records
	// (for backward compatibility during migration)
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, p := range s.State.Players {
		if p.Token == token {
			return p.ID, p.CompanyID, true
		}
	}
	return 0, 0, false
}
