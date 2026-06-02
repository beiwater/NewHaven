package service

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"go-sim-api/internal/model"
)

func (s *Service) StartBuildingProduction(companyID int, buildingID string, body map[string]any) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return map[string]any{"error": "company not found"}
	}
	resourceID := intFromAny(body["kind"])
	amount := intFromAny(body["amount"])
	if amount <= 0 {
		amount = 1
	}
	reqQuality := intFromAny(body["quality"])
	if reqQuality < 0 {
		reqQuality = 0
	}
	if reqQuality > s.Cfg.Game.MaxQuality {
		reqQuality = s.Cfg.Game.MaxQuality
	}

	if err := s.checkResearchUnlock(resourceID); err != nil {
		return map[string]any{"error": err.Error()}
	}
	if err := s.checkBuildingSlot(companyID, buildingID); err != nil {
		return map[string]any{"error": err.Error()}
	}
	if err := s.checkBuildingCanProduce(company, buildingID, resourceID); err != nil {
		return map[string]any{"error": err.Error()}
	}

	buildLevel := s.findBuilding(companyID, buildingID)
	input := s.findRecipe(resourceID, amount)
	if len(input) == 0 {
		return map[string]any{"error": "recipe not found"}
	}
	actualQuality := s.resolveQuality(company, reqQuality, input)
	inputQuality := 0
	if actualQuality > 0 {
		inputQuality = actualQuality - 1
	}

	if err := s.deductInputs(company, input, inputQuality); err != nil {
		return map[string]any{"error": err.Error()}
	}

	s.addLedger("production_input", float64(amount), "out", map[string]any{"resourceId": resourceID, "buildingId": buildingID, "quality": actualQuality})
	now := s.now().UTC()
	durSec := s.calcProductionDuration(resourceID, amount, intFromAny(body["estimatedSecondsToFinish"]))
	if s.State.BoostMultiplier > 1.0 {
		durSec = int(float64(durSec) / s.State.BoostMultiplier)
	}
	outputTotal := amount * buildLevel

	job := model.ProductionJob{
		ID:         fmt.Sprintf("job-%d", s.now().UnixNano()),
		BuildingID: buildingID, ResourceID: resourceID, Amount: amount, Quality: actualQuality,
		Input: input, Output: map[int]int{resourceID: outputTotal},
		StartedAt: now.Format(time.RFC3339), CompletesAt: now.Add(time.Duration(durSec) * time.Second).Format(time.RFC3339), Status: "running",
	}
	s.State.ProductionJobs = append([]model.ProductionJob{job}, s.State.ProductionJobs...)
	s.saveCompanyLocked(company)
	return map[string]any{
		"building":             map[string]any{"id": buildingID, "busy": true, "jobId": job.ID},
		"resourceTransactions": []map[string]any{{"kind": resourceID, "amount": outputTotal, "quality": actualQuality, "buildingLevel": buildLevel}},
		"followerErrors":       []any{},
	}
}

func (s *Service) checkResearchUnlock(resourceID int) error {
	if s.State.UnlockedRecipes != nil {
		if _, needsUnlock := s.State.UnlockedRecipes[resourceID]; needsUnlock {
			return nil
		} else if resourceID > 100 {
			return fmt.Errorf("recipe not unlocked")
		}
	}
	return nil
}

func (s *Service) checkBuildingSlot(companyID int, buildingID string) error {
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return fmt.Errorf("company not found")
	}
	maxSlots := 1 + company.Level/5
	if len(company.PlacedBuildings) >= maxSlots && buildingID == "b-new" {
		return fmt.Errorf("no available building slots")
	}
	return nil
}

func (s *Service) findBuilding(companyID int, buildingID string) int {
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return 1
	}
	for _, b := range company.PlacedBuildings {
		if b["id"] != buildingID {
			continue
		}
		if lv := intFromAny(b["level"]); lv > 1 {
			return lv
		}
		break
	}
	return 1
}

func (s *Service) findRecipe(resourceID, amount int) map[int]int {
	input := map[int]int{}
	for _, r := range s.Data.Resources {
		if intFromAny(r["dbLetter"]) != resourceID {
			continue
		}
		if pf, ok := r["producedFrom"].(map[string]any); ok {
			for k, q := range pf {
				kid, _ := strconv.Atoi(k)
				kq := int(math.Ceil(floatFromAny(q) * float64(amount)))
				if kid > 0 && kq > 0 {
					input[kid] = kq
				}
			}
		}
		break
	}
	return input
}

func (s *Service) checkBuildingCanProduce(company *model.Company, buildingID string, resourceID int) error {
	kind := 0
	for _, b := range company.PlacedBuildings {
		if b["id"] == buildingID {
			kind = intFromAny(b["kind"])
			break
		}
	}
	if kind == 0 {
		return fmt.Errorf("building not found")
	}
	if len(s.productionIDsForKind(kind)) == 0 {
		return fmt.Errorf("building cannot produce goods")
	}
	for _, id := range s.productionIDsForKind(kind) {
		if id == resourceID {
			return nil
		}
	}
	return fmt.Errorf("resource %d is not supported by this building", resourceID)
}

func (s *Service) productionIDsForKind(kind int) []int {
	switch kind {
	case 1:
		return []int{3, 4, 5, 6, 66, 72, 120}
	case 2:
		return []int{7, 8, 9, 121, 122, 127, 133, 134, 135, 137, 139, 141}
	default:
		return []int{}
	}
}

func (s *Service) resolveQuality(company *model.Company, reqQuality int, input map[int]int) int {
	if reqQuality == 0 || len(input) == 0 {
		return reqQuality
	}
	inputQuality := reqQuality - 1
	minAvail := -1
	for k := range input {
		if s.inventoryGet(company, k, inputQuality) > 0 {
			if minAvail < 0 || inputQuality < minAvail {
				minAvail = inputQuality
			}
			continue
		}
		found := false
		for q := inputQuality - 1; q >= 0; q-- {
			if s.inventoryGet(company, k, q) > 0 {
				if minAvail < 0 || q < minAvail {
					minAvail = q
				}
				found = true
				break
			}
		}
		if !found {
			return 0
		}
	}
	if minAvail >= 0 {
		return minAvail + 1
	}
	return reqQuality
}

func (s *Service) deductInputs(company *model.Company, input map[int]int, quality int) error {
	for k, q := range input {
		if have := s.inventoryGet(company, k, quality); have < q {
			return fmt.Errorf("not enough input resources")
		}
	}
	for k, q := range input {
		s.inventorySub(company, k, quality, q)
	}
	return nil
}

func (s *Service) calcProductionDuration(resourceID, amount int, durSec int) int {
	if durSec <= 0 {
		durSec = max(30, amount*6)
	}
	if models, ok := s.Data.EconomyModel["models"].(map[string]any); ok {
		if resM, ok := models[fmt.Sprintf("%d", resourceID)].(map[string]any); ok {
			if st1, ok := resM["state_1"].(map[string]any); ok {
				if bl, ok := st1["buildingLevelsNeededPerUnitPerHour"].(float64); ok && bl > 0 {
					durSec = max(durSec, int(math.Round(float64(amount)/bl/20)))
				}
			}
		}
	}
	return durSec
}

func (s *Service) refreshProductionJobs(companyID int) {
	now := s.now().UTC()
	for i := range s.State.ProductionJobs {
		j := &s.State.ProductionJobs[i]
		if j.Status != "running" {
			continue
		}
		t, err := time.Parse(time.RFC3339, j.CompletesAt)
		if err == nil && !now.Before(t) {
			j.Status = "ready"
		}
	}
	// Count running jobs per building, enforce slot limits
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return
	}
	slots := company.ProductionSlots
	if slots <= 0 {
		slots = s.Cfg.Game.BaseProductionSlots
	}
	byBuilding := map[string]int{}
	for _, j := range s.State.ProductionJobs {
		if j.Status == "running" || j.Status == "ready" {
			byBuilding[j.BuildingID]++
		}
	}
	for i := range company.PlacedBuildings {
		b := company.PlacedBuildings[i]
		bid := fmt.Sprint(b["id"])
		b["busy"] = byBuilding[bid] >= slots
	}
}
func (s *Service) RefreshProductionJobs(companyID int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.refreshProductionJobs(companyID)
}
func (s *Service) ProductionQueue(companyID int) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return map[string]any{"error": "company not found"}
	}
	// Group jobs by building
	byBuilding := map[string][]model.ProductionJob{}
	for _, j := range s.State.ProductionJobs {
		bid := j.BuildingID
		byBuilding[bid] = append(byBuilding[bid], j)
	}
	slots := company.ProductionSlots
	if slots <= 0 {
		slots = s.Cfg.Game.BaseProductionSlots
	}
	return map[string]any{
		"byBuilding": byBuilding,
		"maxSlots":   slots * len(company.PlacedBuildings),
		"inUse":      len(s.State.ProductionJobs),
	}
}

func (s *Service) AddProductionSlot(companyID int) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}
	current := company.ProductionSlots
	cost := float64(current+1) * s.Cfg.Game.SlotUpgradeCost
	if company.Money < cost {
		return nil, fmt.Errorf("need %.0f, have %.0f", cost, company.Money)
	}
	company.Money -= cost
	company.ProductionSlots++
	s.addLedger("slot_upgrade", -cost, "out", map[string]any{"newSlots": company.ProductionSlots})
	s.saveCompanyLocked(company)
	return map[string]any{"slots": company.ProductionSlots, "cost": cost}, nil
}

func (s *Service) CancelProductionJob(companyID int, jobID string) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}
	for i, j := range s.State.ProductionJobs {
		if j.ID != jobID {
			continue
		}
		if j.Status == "claimed" {
			return nil, fmt.Errorf("job already claimed")
		}
		// Refund 50% of inputs
		for k, q := range j.Input {
			company.Inventory[k] += q / 2
		}
		s.State.ProductionJobs = append(s.State.ProductionJobs[:i], s.State.ProductionJobs[i+1:]...)
		s.addLedger("cancel_production", 0, "out", map[string]any{"jobId": jobID})
		s.saveCompanyLocked(company)
		return map[string]any{"status": "cancelled", "jobId": jobID}, nil
	}
	return nil, fmt.Errorf("job not found")
}
