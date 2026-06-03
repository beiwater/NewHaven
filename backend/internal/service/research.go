package service

import (
	"fmt"
	"time"

	"go-sim-api/internal/model"
)

func (s *Service) ResearchProjects() []map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()

	projects := make([]map[string]any, 0, len(s.State.ResearchProjects))
	for _, p := range s.State.ResearchProjects {
		projects = append(projects, map[string]any{
			"id":                p.ID,
			"name":              p.Name,
			"building":          p.Building,
			"resourceCost":      p.ResourceCost,
			"cashCost":          p.CashCost,
			"durationHours":     p.DurationHours,
			"unlockRecipeId":    p.UnlockRecipeID,
			"qualityResourceId": p.QualityResourceID,
			"unlockPct":         p.UnlockPct,
			"status":            p.Status,
			"progress":          p.Progress,
			"startedAt":         p.StartedAt,
			"completesAt":       p.CompletesAt,
		})
	}
	return projects
}

func (s *Service) StartResearch(companyID int, projectID string) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}
	for i := range s.State.ResearchProjects {
		p := &s.State.ResearchProjects[i]
		if p.ID != projectID {
			continue
		}
		if p.Status != "available" {
			return nil, fmt.Errorf("project not available")
		}
		for rid, qty := range p.ResourceCost {
			if company.Inventory[rid] < qty {
				return nil, fmt.Errorf("not enough resource %d: need %d, have %d", rid, qty, company.Inventory[rid])
			}
			company.Inventory[rid] -= qty
		}
		if company.Money < p.CashCost {
			return nil, fmt.Errorf("not enough money: need %.0f, have %.0f", p.CashCost, company.Money)
		}
		company.Money -= p.CashCost

		now := s.now().UTC()
		p.Status = "in_progress"
		p.StartedAt = now.Format(time.RFC3339)
		p.CompletesAt = now.Add(time.Duration(p.DurationHours) * time.Hour).Format(time.RFC3339)
		p.Progress = 0
		s.addLedger("research_start", -p.CashCost, "out", map[string]any{"project": p.ID})
		return map[string]any{"project": p, "status": "started"}, nil
	}
	return nil, fmt.Errorf("project not found")
}

func (s *Service) ResearchProgress() map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.now().UTC()
	for i := range s.State.ResearchProjects {
		p := &s.State.ResearchProjects[i]
		if p.Status != "in_progress" {
			continue
		}
		started, err := time.Parse(time.RFC3339, p.StartedAt)
		if err != nil {
			continue
		}
		completes, err := time.Parse(time.RFC3339, p.CompletesAt)
		if err != nil {
			continue
		}
		duration := completes.Sub(started)
		elapsed := now.Sub(started)
		if duration > 0 {
			p.Progress = int(elapsed * 100 / duration)
			if p.Progress > 100 {
				p.Progress = 100
			}
		}
		if p.Progress >= 100 {
			p.Status = "completed"
			if p.UnlockRecipeID > 0 {
				if s.State.UnlockedRecipes == nil {
					s.State.UnlockedRecipes = map[int]bool{}
				}
				s.State.UnlockedRecipes[p.UnlockRecipeID] = true
			}
			if p.QualityResourceID > 0 {
				if s.State.ResearchedQuality == nil {
					s.State.ResearchedQuality = map[int]int{}
				}
				s.State.ResearchedQuality[p.QualityResourceID]++
			}
			s.addXP(nil, 100)
			s.addLedger("research_complete", 0, "in", map[string]any{"project": p.ID, "unlockRecipe": p.UnlockRecipeID})
		}
	}
	return map[string]any{"projects": s.State.ResearchProjects}
}

func (s *Service) StartResearchProjects() []model.ResearchProject {
	return append([]model.ResearchProject{}, s.State.ResearchProjects...)
}

func (s *Service) CompleteResearch(projectID string) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.State.ResearchProjects {
		p := &s.State.ResearchProjects[i]
		if p.ID != projectID {
			continue
		}
		if p.Status != "in_progress" {
			return nil, fmt.Errorf("project %q is not in progress", projectID)
		}

		// Verify the project timer has actually elapsed
		completesAt, err := time.Parse(time.RFC3339, p.CompletesAt)
		if err == nil && s.now().UTC().Before(completesAt) {
			remaining := completesAt.Sub(s.now().UTC()).Round(time.Second)
			return nil, fmt.Errorf("research still in progress, %s remaining", remaining)
		}

		p.Status = "completed"
		p.Progress = 100
		if p.UnlockRecipeID > 0 {
			if s.State.UnlockedRecipes == nil {
				s.State.UnlockedRecipes = map[int]bool{}
			}
			s.State.UnlockedRecipes[p.UnlockRecipeID] = true
		}

		qualityImprove := 0
		if p.QualityResourceID > 0 {
			if s.State.ResearchedQuality == nil {
				s.State.ResearchedQuality = map[int]int{}
			}
			s.State.ResearchedQuality[p.QualityResourceID]++
			qualityImprove = 1
		}

		s.addXP(nil, 100)
		return map[string]any{
			"ok": true, "projectId": projectID,
			"patentsGained":   1,
			"qualityImproved": qualityImprove,
			"completedAt":     s.now().UTC().Format(time.RFC3339),
		}, nil
	}
	return nil, fmt.Errorf("project not found")
}

func (s *Service) SampleArticles() []map[string]any {
	return []map[string]any{
		{
			"id": "article-1", "title": "Market Report: Commodity Prices Surge",
			"author": "Financial Times", "body": "Commodity prices across all sectors have seen an unexpected surge this quarter, driven by increased demand and limited supply chain capacity.",
			"publishedAt": s.now().Add(-6 * time.Hour).UTC().Format(time.RFC3339),
			"readCount":   1452,
		},
		{
			"id": "article-2", "title": "New Research Breakthrough in Energy Sector",
			"author": "Science Daily", "body": "A breakthrough in energy research promises to revolutionize power generation efficiency across the industry.",
			"publishedAt": s.now().Add(-24 * time.Hour).UTC().Format(time.RFC3339),
			"readCount":   892,
		},
	}
}
