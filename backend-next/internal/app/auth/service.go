package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"

	domain "github.com/newhaven/backend-next/internal/domain/auth"
	"github.com/newhaven/backend-next/internal/domain/company"
	"github.com/newhaven/backend-next/internal/platform"
	"github.com/newhaven/backend-next/internal/storage"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUsernameTaken      = errors.New("username already exists")
	ErrCompanyNotFound    = errors.New("company not found")
)

// Service is the auth application use case.
type Service struct {
	players   storage.PlayerStorage
	companies storage.CompanyStorage
	clock     platform.Clock
	idgen     *platform.IDGen
	logger    *platform.Logger
	jwtKey    string
}

func NewService(
	players storage.PlayerStorage,
	companies storage.CompanyStorage,
	clock platform.Clock,
	idgen *platform.IDGen,
	logger *platform.Logger,
	jwtKey string,
) *Service {
	return &Service{
		players:   players,
		companies: companies,
		clock:     clock,
		idgen:     idgen,
		logger:    logger,
		jwtKey:    jwtKey,
	}
}

// Register creates a new player account and company, returning a JWT.
func (s *Service) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.LoginResponse, error) {
	now := s.clock.Now().UTC().Format(time.RFC3339)

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("password hash: %w", err)
	}

	player := &domain.Player{
		Username:     req.Username,
		PasswordHash: string(hash),
		DisplayName:  req.Name,
		Gender:       req.Gender,
		Email:        req.Email,
	}

	if err := s.players.CreatePlayer(ctx, player); err != nil {
		return nil, ErrUsernameTaken
	}

	// Create company
	companyName := req.Username
	if req.Name != "" {
		companyName = req.Name
	}

	c := &company.Company{
		PlayerID:    player.ID,
		Name:        companyName,
		Money:       100000, // starting capital
		Level:       1,
		XP:          0,
		Preferences: company.NewPlayerPreferences(),
		Inventory:   make(map[int]int),
		CreatedAt:   now,
	}

	if err := s.companies.CreateCompany(ctx, c); err != nil {
		return nil, fmt.Errorf("create company: %w", err)
	}

	// Sign JWT
	token, err := SignJWT(player.ID, c.ID, s.jwtKey)
	if err != nil {
		return nil, fmt.Errorf("sign token: %w", err)
	}

	s.logger.Info("player registered",
		"player_id", player.ID,
		"company_id", c.ID,
		"username", req.Username,
	)

	return &domain.LoginResponse{
		Token:     token,
		PlayerID:  player.ID,
		CompanyID: c.ID,
		Username:  req.Username,
	}, nil
}

// Login authenticates a user and returns a JWT.
func (s *Service) Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error) {
	player, err := s.players.GetPlayerByUsername(ctx, req.Username)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(player.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	c, err := s.companies.GetCompanyByPlayerID(ctx, player.ID)
	if err != nil {
		return nil, ErrCompanyNotFound
	}

	token, err := SignJWT(player.ID, c.ID, s.jwtKey)
	if err != nil {
		return nil, fmt.Errorf("sign token: %w", err)
	}

	s.logger.Info("player logged in",
		"player_id", player.ID,
		"company_id", c.ID,
		"username", req.Username,
	)

	return &domain.LoginResponse{
		Token:     token,
		PlayerID:  player.ID,
		CompanyID: c.ID,
		Username:  req.Username,
	}, nil
}

// DevBootstrap creates a dev user if no users exist and DevMode is enabled.
func (s *Service) DevBootstrap(ctx context.Context) error {
	_, err := s.players.GetPlayerByUsername(ctx, "dev")
	if err == nil {
		// dev user already exists
		return nil
	}

	req := &domain.RegisterRequest{
		Username: "dev",
		Password: "dev",
		Name:     "Dev Player",
		Gender:   "other",
		Email:    "dev@newhaven.game",
	}

	resp, err := s.Register(ctx, req)
	if err != nil {
		return fmt.Errorf("dev bootstrap: %w", err)
	}
	devCompany, err := s.companies.GetCompany(ctx, resp.CompanyID)
	if err != nil {
		return fmt.Errorf("dev bootstrap company: %w", err)
	}
	devCompany.Level = 100
	devCompany.XP = 1000000000
	devCompany.Money = 1000000000
	devCompany.Preferences = completedArrivalPreferences()
	if err := s.companies.UpdateCompany(ctx, devCompany); err != nil {
		return fmt.Errorf("dev bootstrap update: %w", err)
	}

	s.logger.Info("dev bootstrap complete",
		"player_id", resp.PlayerID,
		"company_id", resp.CompanyID,
	)
	return nil
}

func completedArrivalPreferences() map[string]any {
	return map[string]any{
		"storyProgress": map[string]any{
			company.ArrivalStoryID: map[string]any{
				"status": "completed",
				"stepId": company.ArrivalStoryFirstStep,
			},
		},
	}
}
