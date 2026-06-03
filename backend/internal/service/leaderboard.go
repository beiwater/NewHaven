package service

import (
	"go-sim-api/internal/model"
	"sort"
)

// LeaderboardEntry is a public-safe projection of a company for leaderboard display.
type LeaderboardEntry struct {
	Rank        int     `json:"rank"`
	CompanyID   int     `json:"companyId"`
	CompanyName string  `json:"companyName"`
	Level       int     `json:"level"`
	MainStat    float64 `json:"mainStat"` // the value of the current sort dimension
}

// LeaderboardResult is the full paginated response.
type LeaderboardResult struct {
	Entries    []LeaderboardEntry `json:"entries"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	Limit      int                `json:"limit"`
	TotalPages int                `json:"totalPages"`
	Sort       string             `json:"sort"`
}

// Leaderboard returns a paginated, sorted leaderboard of all non-bot companies.
// Supported sortBy: "net_worth", "level", "production", "sales", "contracts".
func (s *Service) Leaderboard(sortBy string, page, limit int) LeaderboardResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Collect candidates: exclude bot companies by known IDs
	botIDs := map[int]bool{
		s.Cfg.Game.Bot1ID: true,
		s.Cfg.Game.Bot2ID: true,
	}
	candidates := make([]companyRank, 0, len(s.State.Companies))
	for _, c := range s.State.Companies {
		if botIDs[c.ID] {
			continue
		}
		candidates = append(candidates, companyRank{
			Company:   c,
			sortValue: sortValueFor(c, sortBy),
		})
	}

	// Sort descending by sortValue (ties broken by ID)
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].sortValue != candidates[j].sortValue {
			return candidates[i].sortValue > candidates[j].sortValue
		}
		return candidates[i].ID < candidates[j].ID
	})

	total := len(candidates)
	totalPages := (total + limit - 1) / limit
	if totalPages < 1 {
		totalPages = 1
	}
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	// Slice the page
	start := (page - 1) * limit
	if start >= total {
		start = 0
		page = 1
	}
	end := start + limit
	if end > total {
		end = total
	}
	paged := candidates[start:end]

	entries := make([]LeaderboardEntry, 0, len(paged))
	for i, cr := range paged {
		entries = append(entries, LeaderboardEntry{
			Rank:        start + i + 1,
			CompanyID:   cr.ID,
			CompanyName: cr.Name,
			Level:       cr.Level,
			MainStat:    cr.sortValue,
		})
	}

	return LeaderboardResult{
		Entries:    entries,
		Total:      total,
		Page:       page,
		Limit:      limit,
		TotalPages: totalPages,
		Sort:       sortBy,
	}
}

type companyRank struct {
	model.Company
	sortValue float64
}

// sortValueFor computes the main stat for a given sort dimension.
func sortValueFor(c model.Company, sortBy string) float64 {
	switch sortBy {
	case "net_worth":
		invValue := 0.0
		for _, qty := range c.Inventory {
			invValue += float64(qty) * 5.0
		}
		buildingValue := float64(len(c.PlacedBuildings)) * 25000.0
		return c.Money + buildingValue + invValue

	case "level":
		return float64(c.Level)

	case "production":
		return float64(len(c.PlacedBuildings))

	case "sales":
		return float64(len(c.Inventory))

	case "contracts":
		return c.Money / 1000.0

	default:
		return c.Money
	}
}
