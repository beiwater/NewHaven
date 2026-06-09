package research

import (
	"context"
	"errors"
	"sort"
	"sync"

	"github.com/newhaven/backend-next/internal/catalog"
	"github.com/newhaven/backend-next/internal/config"
	"github.com/newhaven/backend-next/internal/domain/research"
	"github.com/newhaven/backend-next/internal/formula"
	"github.com/newhaven/backend-next/internal/platform"
	"github.com/newhaven/backend-next/internal/storage"
)

var (
	ErrMaxLevel          = errors.New("research already at max level")
	ErrInsufficientFunds = errors.New("insufficient funds for research")
)

// ResearchInfo is the DTO for a single resource's research status.
type ResearchInfo struct {
	ResourceID int     `json:"resourceId"`
	Name       string  `json:"name"`
	Tier       int     `json:"tier"`
	Level      int     `json:"level"`
	NextCost   float64 `json:"nextCost,omitempty"`
	SpeedBonus float64 `json:"speedBonus"`
}

// LevelUpResponse is the DTO for a successful research level-up.
type LevelUpResponse struct {
	ResourceID int     `json:"resourceId"`
	NewLevel   int     `json:"newLevel"`
	Cost       float64 `json:"cost"`
	SpeedBonus float64 `json:"speedBonus"`
	NextCost   float64 `json:"nextCost,omitempty"`
}

// Service is the research application use case.
type Service struct {
	research  storage.ResearchStorage
	companies storage.CompanyStorage
	resources map[int]*catalog.ResourceEntry
	cfg       *config.GameConfig
	logger    *platform.Logger
	mu        sync.Mutex
}

// NewService creates a new research service.
func NewService(
	research storage.ResearchStorage,
	companies storage.CompanyStorage,
	resources map[int]*catalog.ResourceEntry,
	cfg *config.GameConfig,
	logger *platform.Logger,
) *Service {
	return &Service{
		research:  research,
		companies: companies,
		resources: resources,
		cfg:       cfg,
		logger:    logger,
	}
}

// ListResearch returns all researchable resources with their current level and next cost.
// Resources with no research record are shown at level 0.
func (s *Service) ListResearch(ctx context.Context, companyID int) ([]ResearchInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, err := s.research.GetCompanyResearch(ctx, companyID)
	if err != nil {
		return nil, err
	}

	// Build a lookup: resourceID -> level
	levelByID := make(map[int]int, len(existing))
	for _, rr := range existing {
		levelByID[rr.ResourceID] = rr.Level
	}

	sortedIDs := sortResourceIDs(s.resources)
	result := make([]ResearchInfo, 0, len(s.resources))

	for _, rid := range sortedIDs {
		entry, ok := s.resources[rid]
		if !ok {
			continue
		}

		lvl := levelByID[rid]
		baseCost := formula.ResearchBaseCost(entry.Tier, s.cfg.ResearchBaseCost)
		speedBonus := formula.ResearchSpeedBonus(lvl)

		info := ResearchInfo{
			ResourceID: rid,
			Name:       entry.Name,
			Tier:       entry.Tier,
			Level:      lvl,
			SpeedBonus: speedBonus,
		}
		if lvl < research.MaxResearchLevel {
			info.NextCost = formula.ResearchLevelCost(baseCost, lvl+1)
		}
		result = append(result, info)
	}

	if result == nil {
		result = []ResearchInfo{}
	}
	return result, nil
}

// LevelUp pays money and increases a resource's research level by 1.
func (s *Service) LevelUp(ctx context.Context, companyID int, resourceID int) (*LevelUpResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Validate resource exists.
	entry, ok := s.resources[resourceID]
	if !ok {
		return nil, errors.New("resource not found")
	}

	// Load current research level.
	rr, err := s.research.GetResourceResearch(ctx, companyID, resourceID)
	if err != nil {
		return nil, err
	}

	currentLevel := 0
	if rr != nil {
		currentLevel = rr.Level
	}

	if currentLevel >= research.MaxResearchLevel {
		return nil, ErrMaxLevel
	}

	// Calculate cost for next level.
	baseCost := formula.ResearchBaseCost(entry.Tier, s.cfg.ResearchBaseCost)
	cost := formula.ResearchLevelCost(baseCost, currentLevel+1)

	// Load company and check money.
	company, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, err
	}

	if company.Money < cost {
		return nil, ErrInsufficientFunds
	}

	// Deduct money.
	company.Money -= cost
	if err := s.companies.UpdateCompany(ctx, company); err != nil {
		return nil, err
	}

	// Save new research level.
	newLevel := currentLevel + 1
	newRR := &research.ResourceResearch{
		CompanyID:  companyID,
		ResourceID: resourceID,
		Level:      newLevel,
	}
	if err := s.research.SaveResourceResearch(ctx, newRR); err != nil {
		return nil, err
	}

	speedBonus := formula.ResearchSpeedBonus(newLevel)
	resp := &LevelUpResponse{
		ResourceID: resourceID,
		NewLevel:   newLevel,
		Cost:       cost,
		SpeedBonus: speedBonus,
	}
	if newLevel < research.MaxResearchLevel {
		resp.NextCost = formula.ResearchLevelCost(baseCost, newLevel+1)
	}

	return resp, nil
}

// sortResourceIDs returns sorted resource IDs from the map for deterministic output.
func sortResourceIDs(m map[int]*catalog.ResourceEntry) []int {
	ids := make([]int, 0, len(m))
	for id := range m {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids
}
