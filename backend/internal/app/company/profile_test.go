package company_test

import (
	"context"
	"log/slog"
	"testing"

	appcompany "github.com/beiwater/NewHaven/backend/internal/app/company"
	domain "github.com/beiwater/NewHaven/backend/internal/domain/company"
	"github.com/beiwater/NewHaven/backend/internal/platform"
	"github.com/beiwater/NewHaven/backend/internal/storage/memory"
)

func TestStoryProgressLifecycle(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	c := &domain.Company{
		PlayerID: 7, Name: "Story Co", Level: 1,
		Preferences: domain.NewPlayerPreferences(),
	}
	if err := store.CreateCompany(ctx, c); err != nil {
		t.Fatal(err)
	}
	svc := appcompany.NewService(store, platform.NewLogger(slog.Default()), 0)

	profile, err := svc.GetProfile(ctx, 7, c.ID, c.ID)
	if err != nil {
		t.Fatal(err)
	}
	if !profile.LevelInfo.InTutorial || profile.LevelInfo.TutorialCompleted {
		t.Fatalf("new account has wrong tutorial state: %+v", profile.LevelInfo)
	}

	_, err = svc.UpdateStoryProgress(ctx, c.ID, appcompany.StoryProgressRequest{
		StoryID: domain.ArrivalStoryID, StepID: "intro-bells", Status: "in_progress",
	})
	if err != nil {
		t.Fatal(err)
	}
	done, err := svc.UpdateStoryProgress(ctx, c.ID, appcompany.StoryProgressRequest{
		StoryID: domain.ArrivalStoryID, StepID: "system-objective", Status: "completed",
	})
	if err != nil {
		t.Fatal(err)
	}
	if done.Status != "completed" {
		t.Fatalf("expected completed, got %+v", done)
	}

	late, err := svc.UpdateStoryProgress(ctx, c.ID, appcompany.StoryProgressRequest{
		StoryID: domain.ArrivalStoryID, StepID: "intro-crest", Status: "in_progress",
	})
	if err != nil {
		t.Fatal(err)
	}
	if late.Status != "completed" || late.StepID != "system-objective" {
		t.Fatalf("late progress request regressed terminal state: %+v", late)
	}
}

func TestProfileWithoutStoryProgressDoesNotStartTutorial(t *testing.T) {
	ctx := context.Background()
	store := memory.New()
	c := &domain.Company{PlayerID: 8, Name: "Legacy Co", Level: 2}
	if err := store.CreateCompany(ctx, c); err != nil {
		t.Fatal(err)
	}
	svc := appcompany.NewService(store, platform.NewLogger(slog.Default()), 0)
	profile, err := svc.GetProfile(ctx, 8, c.ID, c.ID)
	if err != nil {
		t.Fatal(err)
	}
	if profile.LevelInfo.InTutorial || profile.LevelInfo.TutorialCompleted {
		t.Fatalf("legacy account should have no forced tutorial state: %+v", profile.LevelInfo)
	}
}
