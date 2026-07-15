package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	domain "github.com/beiwater/NewHaven/backend/internal/domain/auth"
	"github.com/beiwater/NewHaven/backend/internal/domain/company"
	warehousedomain "github.com/beiwater/NewHaven/backend/internal/domain/warehouse"
	"github.com/beiwater/NewHaven/backend/internal/platform"
	"github.com/beiwater/NewHaven/backend/internal/storage"
)

var (
	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrUsernameTaken      = errors.New("username already exists")
	ErrCompanyNotFound    = errors.New("company not found")
)

// Service is the auth application use case.
type Service struct {
	players     storage.PlayerStorage
	companies   storage.CompanyStorage
	warehouses  storage.WarehouseStorage
	clock       platform.Clock
	idgen       *platform.IDGen
	logger      *platform.Logger
	jwtKey      string
	devPassword string
}

func NewService(
	players storage.PlayerStorage,
	companies storage.CompanyStorage,
	warehouses storage.WarehouseStorage,
	clock platform.Clock,
	idgen *platform.IDGen,
	logger *platform.Logger,
	jwtKey string,
	devPassword string,
) *Service {
	return &Service{
		players:     players,
		companies:   companies,
		warehouses:  warehouses,
		clock:       clock,
		idgen:       idgen,
		logger:      logger,
		jwtKey:      jwtKey,
		devPassword: devPassword,
	}
}

// Register creates a new player account and company, returning a JWT.
func (s *Service) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.LoginResponse, error) {
	return s.register(ctx, req, true)
}

func (s *Service) register(ctx context.Context, req *domain.RegisterRequest, validate bool) (*domain.LoginResponse, error) {
	reqCopy := *req
	req = &reqCopy
	req.Username = normalizeUsername(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	if validate {
		if err := validateRegistration(req); err != nil {
			return nil, err
		}
	}
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
		if errors.Is(err, storage.ErrAlreadyExists) {
			return nil, ErrUsernameTaken
		}
		return nil, fmt.Errorf("create player: %w", err)
	}

	// Create company
	companyName := req.Username
	if req.Name != "" {
		companyName = req.Name
	}

	c := &company.Company{
		PlayerID:     player.ID,
		Name:         companyName,
		Money:        100000, // starting capital
		Level:        1,
		XP:           0,
		Preferences:  company.NewPlayerPreferences(),
		Inventory:    make(map[int]int),
		CreatedAt:    now,
		LastRetailAt: now,
	}

	if err := s.companies.CreateCompany(ctx, c); err != nil {
		if rollbackErr := s.players.DeletePlayer(ctx, player.ID); rollbackErr != nil {
			return nil, fmt.Errorf("create company: %w", errors.Join(err, fmt.Errorf("rollback player: %w", rollbackErr)))
		}
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
		Username:  player.Username,
	}, nil
}

// Login authenticates a user and returns a JWT.
func (s *Service) Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error) {
	username := normalizeUsername(req.Username)
	if username == "" || req.Password == "" {
		return nil, ErrInvalidCredentials
	}
	player, err := s.players.GetPlayerByUsername(ctx, username)
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
		"username", player.Username,
	)

	return &domain.LoginResponse{
		Token:     token,
		PlayerID:  player.ID,
		CompanyID: c.ID,
		Username:  player.Username,
	}, nil
}

// DevBootstrap creates a dev user if no users exist and DevMode is enabled.
func (s *Service) DevBootstrap(ctx context.Context) error {
	player, err := s.players.GetPlayerByUsername(ctx, "dev")
	if err != nil {
		req := &domain.RegisterRequest{
			Username: "dev",
			Password: s.devPassword,
			Name:     "Developer Merchant",
			Gender:   "other",
			Email:    "dev@newhaven.game",
		}
		if _, err := s.register(ctx, req, false); err != nil {
			return fmt.Errorf("dev bootstrap: %w", err)
		}
		player, err = s.players.GetPlayerByUsername(ctx, "dev")
		if err != nil {
			return fmt.Errorf("dev bootstrap player: %w", err)
		}
	}

	devCompany, err := s.companies.GetCompanyByPlayerID(ctx, player.ID)
	if err != nil {
		return fmt.Errorf("dev bootstrap company: %w", err)
	}
	provisionDeveloperMerchant(devCompany, s.idgen)
	if err := s.companies.UpdateCompany(ctx, devCompany); err != nil {
		return fmt.Errorf("dev bootstrap update: %w", err)
	}
	if err := s.syncDeveloperWarehouse(ctx, devCompany); err != nil {
		return fmt.Errorf("dev bootstrap warehouse: %w", err)
	}

	s.logger.Info("dev bootstrap complete",
		"player_id", player.ID,
		"company_id", devCompany.ID,
	)
	return nil
}

func provisionDeveloperMerchant(c *company.Company, idgen *platform.IDGen) {
	c.Name = "Developer Merchant"
	c.Level = 100
	c.XP = 1000000000
	c.Money = 1000000000
	c.WarehouseLevel = 100
	c.Preferences = completedArrivalPreferences()
	c.Inventory = make(map[int]int, 12)
	for resourceID := 1; resourceID <= 12; resourceID++ {
		c.Inventory[resourceID] = 5000
	}

	if len(c.Buildings) == 0 {
		names := []string{"Farm", "Barn", "Mill", "Kitchen", "Bakery", "Market Stall", "Cafe", "Food Truck"}
		c.Buildings = make([]company.Building, 0, len(names))
		for i, name := range names {
			kind := i + 1
			c.Buildings = append(c.Buildings, company.Building{
				ID:         idgen.Next("dev-building"),
				BuildingID: kind,
				Kind:       kind,
				Name:       name,
				Level:      5,
				MapID:      "harbor",
				SlotID:     fmt.Sprintf("harbor-plot-%02d", kind),
				X:          i%3 + 1,
				Y:          i/3 + 1,
			})
		}
	}
}

func (s *Service) syncDeveloperWarehouse(ctx context.Context, c *company.Company) error {
	if s.warehouses == nil {
		return errors.New("warehouse storage is required for dev bootstrap")
	}
	w, err := s.warehouses.GetWarehouse(ctx, c.ID)
	if err != nil {
		return err
	}
	w.Capacity = 102000 // level 100 with the standard 1,000 base capacity
	w.UsedCapacity = 0
	w.Items = make([]warehousedomain.Item, 0, len(c.Inventory))
	for resourceID := 1; resourceID <= 12; resourceID++ {
		amount := c.Inventory[resourceID]
		w.Items = append(w.Items, warehousedomain.Item{ResourceID: resourceID, Amount: amount})
		w.UsedCapacity += amount
	}
	return s.warehouses.UpdateWarehouse(ctx, w)
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
