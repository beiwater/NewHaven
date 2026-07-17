package research

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"github.com/beiwater/NewHaven/backend/internal/catalog"
	"github.com/beiwater/NewHaven/backend/internal/config"
	"github.com/beiwater/NewHaven/backend/internal/domain/finance"
	domainresearch "github.com/beiwater/NewHaven/backend/internal/domain/research"
	"github.com/beiwater/NewHaven/backend/internal/formula"
	"github.com/beiwater/NewHaven/backend/internal/platform"
	"github.com/beiwater/NewHaven/backend/internal/storage"
)

var (
	ErrMaxLevel          = errors.New("quality research already at Q12")
	ErrInsufficientFunds = errors.New("insufficient funds for quality research")
	ErrResearchSequence  = errors.New("quality research must advance one level at a time")
	ErrResourceNotFound  = errors.New("researchable resource not found")
)

// ResearchInfo describes one product's current quality licence and next step.
type ResearchInfo struct {
	ResourceID        int     `json:"resourceId"`
	Name              string  `json:"name"`
	Tier              int     `json:"tier"`
	MaxQuality        int     `json:"maxQuality"`
	NextQuality       int     `json:"nextQuality,omitempty"`
	NextCost          float64 `json:"nextCost,omitempty"`
	SalesSpeedBonus   int     `json:"salesSpeedBonus"`
	NextSalesSpeedPct int     `json:"nextSalesSpeedPct,omitempty"`
}

// UnlockQualityResponse is returned after a target quality has been unlocked
// or safely replayed.
type UnlockQualityResponse struct {
	ResourceID      int     `json:"resourceId"`
	MaxQuality      int     `json:"maxQuality"`
	Cost            float64 `json:"cost"`
	Charged         bool    `json:"charged"`
	SalesSpeedBonus int     `json:"salesSpeedBonus"`
	NextQuality     int     `json:"nextQuality,omitempty"`
	NextCost        float64 `json:"nextCost,omitempty"`
}

// Service is the quality research application use case.
type Service struct {
	research  storage.ResearchStorage
	companies storage.CompanyStorage
	finance   storage.FinanceStorage
	resources map[int]*catalog.ResourceEntry
	cfg       *config.GameConfig
	logger    *platform.Logger
	mu        sync.Mutex
}

// NewService creates a new research service.
func NewService(
	research storage.ResearchStorage,
	companies storage.CompanyStorage,
	financeStorage storage.FinanceStorage,
	resources map[int]*catalog.ResourceEntry,
	cfg *config.GameConfig,
	logger *platform.Logger,
) *Service {
	return &Service{
		research:  research,
		companies: companies,
		finance:   financeStorage,
		resources: resources,
		cfg:       cfg,
		logger:    logger,
	}
}

// ListResearch returns every researchable product with its unlocked quality.
// Missing records intentionally mean Q0. Legacy records above Q12 are clamped
// to Q12 so old snapshots remain playable after the research model migration.
func (s *Service) ListResearch(ctx context.Context, companyID int) ([]ResearchInfo, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	existing, err := s.research.GetCompanyResearch(ctx, companyID)
	if err != nil {
		return nil, err
	}
	levelByID := make(map[int]int, len(existing))
	for _, item := range existing {
		levelByID[item.ResourceID] = clampResearchQuality(item.Level)
	}

	result := make([]ResearchInfo, 0, len(s.resources))
	for _, resourceID := range sortResourceIDs(s.resources) {
		entry := s.resources[resourceID]
		if !isQualityResearchable(entry) {
			continue
		}
		maxQuality := levelByID[resourceID]
		info := ResearchInfo{
			ResourceID:      resourceID,
			Name:            entry.Name,
			Tier:            max(1, entry.Tier),
			MaxQuality:      maxQuality,
			SalesSpeedBonus: maxQuality * 2,
		}
		if maxQuality < domainresearch.MaxResearchLevel {
			info.NextQuality = maxQuality + 1
			info.NextCost = s.researchCost(entry, info.NextQuality)
			info.NextSalesSpeedPct = info.NextQuality * 2
		}
		result = append(result, info)
	}
	return result, nil
}

// UnlockQuality pays cash and advances exactly one product quality. Target
// quality makes retries idempotent even across two API service instances.
func (s *Service) UnlockQuality(ctx context.Context, companyID, resourceID, targetQuality int) (*UnlockQualityResponse, error) {
	entry, ok := s.resources[resourceID]
	if !ok || !isQualityResearchable(entry) {
		return nil, ErrResourceNotFound
	}
	if targetQuality < 1 || targetQuality > domainresearch.MaxResearchLevel {
		return nil, ErrMaxLevel
	}

	cost := s.researchCost(entry, targetQuality)
	unlocked, replayed, err := s.research.UnlockResourceQuality(ctx, companyID, resourceID, targetQuality, cost)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrInsufficientFunds):
			return nil, ErrInsufficientFunds
		case errors.Is(err, storage.ErrStateConflict):
			return nil, ErrResearchSequence
		default:
			return nil, err
		}
	}

	maxQuality := clampResearchQuality(unlocked.Level)
	response := &UnlockQualityResponse{
		ResourceID:      resourceID,
		MaxQuality:      maxQuality,
		Cost:            cost,
		Charged:         !replayed,
		SalesSpeedBonus: maxQuality * 2,
	}
	if maxQuality < domainresearch.MaxResearchLevel {
		response.NextQuality = maxQuality + 1
		response.NextCost = s.researchCost(entry, response.NextQuality)
	}

	if !replayed {
		s.recordResearchCost(ctx, companyID, resourceID, targetQuality, cost)
	}
	return response, nil
}

func (s *Service) researchCost(entry *catalog.ResourceEntry, targetQuality int) float64 {
	baseCost := 1000.0
	growth := 1.2
	if s.cfg != nil {
		if s.cfg.ResearchBaseCost > 0 {
			baseCost = s.cfg.ResearchBaseCost
		}
		if s.cfg.ResearchCostGrowth > 1 {
			growth = s.cfg.ResearchCostGrowth
		}
	}
	return formula.QualityResearchCost(entry.Tier, targetQuality, baseCost, growth)
}

func (s *Service) recordResearchCost(ctx context.Context, companyID, resourceID, targetQuality int, cost float64) {
	if s.finance == nil {
		return
	}
	company, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return
	}
	entry := &finance.LedgerEntry{
		CompanyID: companyID, Kind: "quality_research", Amount: cost, Direction: "out",
		BalanceAfter: company.Money, CreatedAt: time.Now().UTC().Format(time.RFC3339),
		Metadata: map[string]any{"resource_id": resourceID, "quality": targetQuality},
	}
	if err := s.finance.AppendLedgerEntry(ctx, entry); err != nil && s.logger != nil {
		s.logger.Warn("quality research ledger write failed", "company_id", companyID, "resource_id", resourceID, "quality", targetQuality, "error", err)
	}
}

func isQualityResearchable(entry *catalog.ResourceEntry) bool {
	return entry != nil && !entry.IsResearch && entry.ProducedPerHourRaw > 0
}

func clampResearchQuality(level int) int {
	if level < 0 {
		return 0
	}
	if level > domainresearch.MaxResearchLevel {
		return domainresearch.MaxResearchLevel
	}
	return level
}

func sortResourceIDs(resources map[int]*catalog.ResourceEntry) []int {
	ids := make([]int, 0, len(resources))
	for id := range resources {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	return ids
}
