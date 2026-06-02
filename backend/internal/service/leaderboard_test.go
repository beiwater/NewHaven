package service

import (
	"testing"

	"go-sim-api/internal/config"
	"go-sim-api/internal/data"
	"go-sim-api/internal/model"
)

func TestLeaderboard_SortsByNetWorth(t *testing.T) {
	svc := leaderboardTestService(t)

	result := svc.Leaderboard("net_worth", 1, 10)
	if len(result.Entries) == 0 {
		t.Fatal("expected at least one leaderboard entry")
	}
	for i := 1; i < len(result.Entries); i++ {
		if result.Entries[i-1].MainStat < result.Entries[i].MainStat {
			t.Fatalf("not sorted descending: rank %d (%.0f) < rank %d (%.0f)",
				result.Entries[i-1].Rank, result.Entries[i-1].MainStat,
				result.Entries[i].Rank, result.Entries[i].MainStat)
		}
	}
	e := result.Entries[0]
	if e.CompanyName == "" {
		t.Error("company name should not be empty")
	}
	if e.Level < 1 {
		t.Error("level should be >= 1")
	}
	if e.MainStat <= 0 {
		t.Error("mainStat should be positive")
	}
}

func TestLeaderboard_SortsByLevel(t *testing.T) {
	svc := leaderboardTestService(t)
	result := svc.Leaderboard("level", 1, 10)
	if len(result.Entries) == 0 {
		t.Fatal("expected entries")
	}
	for i := 1; i < len(result.Entries); i++ {
		if result.Entries[i-1].MainStat < result.Entries[i].MainStat {
			t.Fatalf("not sorted descending by level")
		}
	}
}

func TestLeaderboard_Pagination(t *testing.T) {
	svc := leaderboardTestService(t)
	p1 := svc.Leaderboard("net_worth", 1, 2)
	if len(p1.Entries) > 2 {
		t.Fatalf("page 1 should have at most 2 entries, got %d", len(p1.Entries))
	}
	if p1.Page != 1 {
		t.Errorf("expected page 1, got %d", p1.Page)
	}
	if p1.Limit != 2 {
		t.Errorf("expected limit 2, got %d", p1.Limit)
	}
	p2 := svc.Leaderboard("net_worth", 2, 2)
	if p2.Page != 2 {
		t.Errorf("expected page 2, got %d", p2.Page)
	}
	if len(p1.Entries) > 0 && len(p2.Entries) > 0 && p1.Total > 2 {
		if p1.Entries[0].CompanyID == p2.Entries[0].CompanyID {
			t.Error("page 1 and page 2 returned same top entry")
		}
	}
}

func TestLeaderboard_ExcludesBots(t *testing.T) {
	svc := leaderboardTestService(t)
	result := svc.Leaderboard("net_worth", 1, 100)
	for _, e := range result.Entries {
		if e.CompanyName == "Atlas Trading Bot" || e.CompanyName == "Nova Market Bot" {
			t.Errorf("bot company %q should not appear in leaderboard", e.CompanyName)
		}
	}
}

func leaderboardTestService(t *testing.T) *Service {
	t.Helper()
	cfg := config.DefaultTestConfig()
	d := &data.StaticData{
		Resources: []map[string]any{
			{"id": 1, "name": "Power", "dbLetter": 1, "producedPerHourRaw": 100.0},
		},
		EconomyModel: map[string]any{"models": map[string]any{}},
	}
	st := &fakeStorage{
		state: &model.GameState{
			Companies: []model.Company{
				{ID: 1, Name: "Alpha Corp", Money: 500000, Level: 30,
					PlacedBuildings: []map[string]any{{"id": "b1"}, {"id": "b2"}},
					Inventory:       map[int]int{1: 100, 2: 200}},
				{ID: 2, Name: "Bravo Industries", Money: 1200000, Level: 45,
					PlacedBuildings: []map[string]any{{"id": "b1"}, {"id": "b2"}, {"id": "b3"}},
					Inventory:       map[int]int{1: 500, 3: 150}},
				{ID: 3, Name: "Charlie LLC", Money: 250000, Level: 20,
					PlacedBuildings: []map[string]any{},
					Inventory:       map[int]int{}},
				{ID: 900001, Name: "Atlas Trading Bot", Money: 5000000, Level: 99},
				{ID: 900002, Name: "Nova Market Bot", Money: 5000000, Level: 99},
			},
		},
	}
	svc := New(d, cfg, st)
	if svc == nil {
		t.Fatal("service creation returned nil")
	}
	return svc
}
