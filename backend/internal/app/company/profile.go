package company

import (
	"context"
	"errors"
	"fmt"

	"github.com/beiwater/NewHaven/backend/internal/apperr"
	domain "github.com/beiwater/NewHaven/backend/internal/domain/company"
	"github.com/beiwater/NewHaven/backend/internal/formula"
)

// ProfileResponse matches the company profile shape consumed by the frontend.
type ProfileResponse struct {
	AuthCompany CompanyIdentity  `json:"authCompany"`
	AuthUser    PlayerIdentity   `json:"authUser"`
	LevelInfo   CompanyLevelInfo `json:"levelInfo"`
	Unlocks     map[string]any   `json:"unlocks,omitempty"`
	Preferences map[string]any   `json:"preferences"`
}

type CompanyIdentity struct {
	Company   string  `json:"company"`
	CompanyID int     `json:"companyId"`
	Money     float64 `json:"money"`
	Level     int     `json:"level"`
	SimBoosts int     `json:"simBoosts"`
}

type PlayerIdentity struct {
	IsModerator bool   `json:"isModerator"`
	PlayerID    string `json:"playerId"`
}

type CompanyLevelInfo struct {
	Level             int   `json:"level"`
	XP                int64 `json:"xp"`
	InTutorial        bool  `json:"inTutorial"`
	TutorialCompleted bool  `json:"tutorialCompleted"`
}

type StoryProgressRequest struct {
	StoryID string `json:"storyId"`
	StepID  string `json:"stepId"`
	Status  string `json:"status"`
}

// GetProfile returns the authenticated owner's company profile.
func (s *Service) GetProfile(ctx context.Context, playerID, authenticatedCompanyID, requestedCompanyID int) (*ProfileResponse, error) {
	if authenticatedCompanyID != requestedCompanyID {
		return nil, apperr.Forbidden("company profile is private")
	}
	c, err := s.companies.GetCompany(ctx, requestedCompanyID)
	if err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "company not found", errors.Join(ErrNotFound, err))
	}
	status := arrivalStoryStatus(c.Preferences)
	unlocks := formula.FeatureUnlocksAtLevel(c.Level)

	return &ProfileResponse{
		AuthCompany: CompanyIdentity{
			Company: c.Name, CompanyID: c.ID, Money: c.Money, Level: c.Level,
		},
		AuthUser: PlayerIdentity{PlayerID: formatPlayerID(playerID)},
		LevelInfo: CompanyLevelInfo{
			Level: c.Level, XP: c.XP,
			InTutorial:        status == "not_started" || status == "in_progress",
			TutorialCompleted: status == "completed" || status == "skipped",
		},
		Preferences: cloneMap(c.Preferences),
		Unlocks:     unlocks,
	}, nil
}

// UpdateStoryProgress stores the authenticated company's current story position.
func (s *Service) UpdateStoryProgress(ctx context.Context, companyID int, req StoryProgressRequest) (*domain.StoryProgress, error) {
	if req.StoryID == "" || req.StepID == "" || !validStoryStatus(req.Status) {
		return nil, apperr.Validation("storyId, stepId, and a valid status are required")
	}
	c, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, apperr.WrapMsg(apperr.KindNotFound, "company not found", errors.Join(ErrNotFound, err))
	}
	preferences := cloneMap(c.Preferences)
	stories, _ := preferences["storyProgress"].(map[string]any)
	if stories == nil {
		stories = make(map[string]any)
	}
	current, _ := stories[req.StoryID].(map[string]any)
	currentStatus, _ := current["status"].(string)
	if (currentStatus == "completed" || currentStatus == "skipped") && req.Status == "in_progress" {
		return &domain.StoryProgress{
			Status: currentStatus,
			StepID: stringValue(current["stepId"], req.StepID),
		}, nil
	}
	progress := domain.StoryProgress{Status: req.Status, StepID: req.StepID}
	stories[req.StoryID] = map[string]any{"status": progress.Status, "stepId": progress.StepID}
	preferences["storyProgress"] = stories
	c.Preferences = preferences

	// Auto level-up when the arrival story is completed for the first time.
	if req.StoryID == domain.ArrivalStoryID && req.Status == "completed" && currentStatus != "completed" && currentStatus != "skipped" {
		if c.Level < s.newbieLevelUpTo {
			c.Level = s.newbieLevelUpTo
		}
	}

	if err := s.companies.UpdateCompany(ctx, c); err != nil {
		return nil, apperr.WrapMsg(apperr.KindInternal, "save story progress", err)
	}
	return &progress, nil
}

func validStoryStatus(status string) bool {
	return status == "not_started" || status == "in_progress" || status == "completed" || status == "skipped"
}

func arrivalStoryStatus(preferences map[string]any) string {
	stories, _ := preferences["storyProgress"].(map[string]any)
	if stories == nil {
		return ""
	}
	arrival, _ := stories[domain.ArrivalStoryID].(map[string]any)
	if arrival == nil {
		return ""
	}
	status, _ := arrival["status"].(string)
	return status
}

func cloneMap(source map[string]any) map[string]any {
	result := make(map[string]any, len(source))
	for key, value := range source {
		if nested, ok := value.(map[string]any); ok {
			result[key] = cloneMap(nested)
			continue
		}
		result[key] = value
	}
	return result
}

func formatPlayerID(playerID int) string {
	return fmt.Sprintf("%d", playerID)
}

func stringValue(value any, fallback string) string {
	if result, ok := value.(string); ok {
		return result
	}
	return fallback
}
