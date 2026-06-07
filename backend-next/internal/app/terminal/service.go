package terminal

import (
	"context"
	"fmt"
	"strings"

	"github.com/newhaven/backend-next/internal/domain/company"
	"github.com/newhaven/backend-next/internal/platform"
	"github.com/newhaven/backend-next/internal/storage"
)

// Service handles terminal/console commands in chat.
type Service struct {
	companies storage.CompanyStorage
	logger    *platform.Logger
}

// NewService creates a new terminal service.
func NewService(companies storage.CompanyStorage, logger *platform.Logger) *Service {
	return &Service{companies: companies, logger: logger}
}

const TerminalCompanyName = "Terminal"

// EnsureTerminalCompany creates the Terminal system company if it doesn't exist.
func (s *Service) EnsureTerminalCompany(ctx context.Context) (int, error) {
	all, err := s.companies.GetAllCompanies(ctx)
	if err != nil {
		return 0, fmt.Errorf("get all companies: %w", err)
	}
	for _, c := range all {
		if c.Name == TerminalCompanyName {
			return c.ID, nil
		}
	}

	// Create Terminal company
	term := &company.Company{
		PlayerID: -1,
		Name:     TerminalCompanyName,
		Money:    0,
		Level:    0,
		XP:       0,
	}
	if err := s.companies.CreateCompany(ctx, term); err != nil {
		return 0, fmt.Errorf("create terminal company: %w", err)
	}
	s.logger.Info("terminal company created", "company_id", term.ID)
	return term.ID, nil
}

// CommandResult is the result of processing a terminal command.
type CommandResult struct {
	Reply string `json:"reply"`
}

// ProcessCommand handles a terminal command from a player.
func (s *Service) ProcessCommand(ctx context.Context, playerCompanyID int, command string) (*CommandResult, error) {
	cmd := strings.TrimSpace(strings.ToLower(command))

	switch {
	case cmd == "/up" || cmd == "/up ":
		return s.handleUp(ctx, playerCompanyID)
	default:
		return &CommandResult{Reply: fmt.Sprintf("Unknown command: %s\nAvailable: /up", command)}, nil
	}
}

func (s *Service) handleUp(ctx context.Context, companyID int) (*CommandResult, error) {
	c, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, fmt.Errorf("get company: %w", err)
	}

	if c.Level >= 8 {
		return &CommandResult{Reply: fmt.Sprintf("You are already level %d. No upgrade needed.", c.Level)}, nil
	}

	c.Level = 8
	if err := s.companies.UpdateCompany(ctx, c); err != nil {
		return nil, fmt.Errorf("update company level: %w", err)
	}

	s.logger.Info("terminal /up executed", "company_id", companyID, "new_level", 8)
	return &CommandResult{Reply: fmt.Sprintf("Level upgraded to 8! Enjoy your boost, Commander.")}, nil
}
