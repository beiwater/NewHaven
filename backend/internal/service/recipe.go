package service

import (
	"fmt"
	"strconv"
)

func (s *Service) RecipeList() []map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]map[string]any, 0)
	for _, r := range s.Data.Resources {
		rid := intFromAny(r["dbLetter"])
		if rid <= 0 {
			continue
		}
		pf, ok := r["producedFrom"].(map[string]any)
		if !ok || len(pf) == 0 {
			continue
		}
		recipe := make([]map[string]any, 0)
		for k, q := range pf {
			kid, _ := strconv.Atoi(k)
			recipe = append(recipe, map[string]any{"resourceId": kid, "quantity": q})
		}
		unlocked := true
		if s.State.UnlockedRecipes != nil && rid > 100 {
			_, unlocked = s.State.UnlockedRecipes[rid]
		}
		out = append(out, map[string]any{
			"resourceId": rid,
			"name":       r["name"],
			"recipe":     recipe,
			"unlocked":   unlocked,
			"buildingId": "b-1",
		})
	}
	return out
}

func (s *Service) ProductionOptions(companyID int, buildingID string) []map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return []map[string]any{}
	}
	kind := 0
	for _, b := range company.PlacedBuildings {
		if b["id"] == buildingID {
			kind = intFromAny(b["kind"])
			break
		}
	}
	ids := s.productionIDsForKind(kind)
	out := make([]map[string]any, 0, len(ids))
	for _, id := range ids {
		detail := s.recipeDetailLocked(id)
		if _, ok := detail["error"]; ok {
			continue
		}
		out = append(out, detail)
	}
	return out
}

func (s *Service) RecipeDetail(resourceID int) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.recipeDetailLocked(resourceID)
}

func (s *Service) recipeDetailLocked(resourceID int) map[string]any {
	for _, r := range s.Data.Resources {
		if intFromAny(r["dbLetter"]) != resourceID {
			continue
		}
		info := map[string]any{"resourceId": resourceID, "name": r["name"]}
		if pf, ok := r["producedFrom"].(map[string]any); ok {
			recipe := make([]map[string]any, 0, 10)
			for k, q := range pf {
				kid, _ := strconv.Atoi(k)
				rname := ""
				for _, rr := range s.Data.Resources {
					if intFromAny(rr["dbLetter"]) == kid {
						rname = fmt.Sprint(rr["name"])
						break
					}
				}
				recipe = append(recipe, map[string]any{"resourceId": kid, "resourceName": rname, "quantity": q})
			}
			info["recipe"] = recipe
			info["producedPerHourRaw"] = r["producedPerHourRaw"]
			info["unitsSoldAnHour"] = r["unitsSoldAnHour"]
		}
		return info
	}
	return map[string]any{"error": "not found"}
}
