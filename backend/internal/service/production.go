package service

import (
	"fmt"
	"math"
	"strconv"
	"time"

	"go-sim-api/internal/model"
)

const maxProductionDurationSeconds = 48 * 60 * 60

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

	if err := s.checkResearchUnlock(company, resourceID); err != nil {
		return map[string]any{"error": err.Error()}
	}
	if err := s.checkBuildingSlot(companyID, buildingID); err != nil {
		return map[string]any{"error": err.Error()}
	}
	if err := s.checkBuildingCanProduce(company, buildingID, resourceID); err != nil {
		return map[string]any{"error": err.Error()}
	}

	buildLevel := s.findBuilding(companyID, buildingID)
	durSec := s.calcProductionDuration(resourceID, amount, buildLevel)
	if s.State.BoostMultiplier > 1.0 {
		durSec = int(math.Ceil(float64(durSec) / s.State.BoostMultiplier))
	}
	if durSec > maxProductionDurationSeconds {
		maxAmount := s.maxProductionAmount(resourceID, buildLevel)
		return map[string]any{
			"error":     fmt.Sprintf("production duration exceeds 48 hours; max amount is %d", maxAmount),
			"maxAmount": maxAmount,
		}
	}

	input, recipeKnown := s.findRecipe(resourceID, amount)
	if !recipeKnown {
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

	job := model.ProductionJob{
		ID:         fmt.Sprintf("job-%d", s.now().UnixNano()),
		BuildingID: buildingID, ResourceID: resourceID, Amount: amount, Quality: actualQuality,
		Input: input, Output: map[int]int{resourceID: amount},
		StartedAt: now.Format(time.RFC3339), CompletesAt: now.Add(time.Duration(durSec) * time.Second).Format(time.RFC3339), Status: "running",
	}
	s.State.ProductionJobs = append([]model.ProductionJob{job}, s.State.ProductionJobs...)
	s.saveCompanyLocked(company)
	return map[string]any{
		"building":             map[string]any{"id": buildingID, "busy": true, "jobId": job.ID},
		"resourceTransactions": []map[string]any{{"kind": resourceID, "amount": amount, "quality": actualQuality, "buildingLevel": buildLevel}},
		"followerErrors":       []any{},
	}
}

func (s *Service) checkResearchUnlock(company *model.Company, resourceID int) error {
	if company == nil {
		return fmt.Errorf("company not found")
	}
	if company.UnlockedRecipes != nil {
		if _, needsUnlock := company.UnlockedRecipes[resourceID]; needsUnlock {
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
		if b["id"] == buildingID {
			if lv := intFromAny(b["level"]); lv >= 1 {
				return lv
			}
			return 1
		}
	}
	return 1
}

func (s *Service) findRecipe(resourceID, amount int) (map[int]int, bool) {
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
		return input, true
	}
	return input, false
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
	if s.Data != nil && s.Data.Buildings != nil {
		for _, raw := range s.Data.Buildings {
			building, ok := raw.(map[string]any)
			if !ok {
				continue
			}
			buildingKind := intFromAny(building["kind"])
			if buildingKind <= 0 {
				buildingKind = intFromAny(building["id"])
			}
			if buildingKind != kind {
				continue
			}
			produces := intSliceFromAny(building["produces"])
			if len(produces) > 0 {
				return produces
			}
			if output := intFromAny(building["outputResourceId"]); output > 0 {
				return []int{output}
			}
			return []int{}
		}
	}
	if kind > 0 {
		for _, r := range s.Data.Resources {
			id := intFromAny(r["dbLetter"])
			if id <= 0 {
				id = intFromAny(r["id"])
			}
			if id == kind {
				return []int{kind}
			}
		}
	}
	return []int{}
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

func (s *Service) calcProductionDuration(resourceID, amount int, level int) int {
	if amount <= 0 {
		amount = 1
	}
	rate := s.outputPerHour(resourceID, level)
	if rate <= 0 {
		return maxProductionDurationSeconds
	}
	return max(30, int(math.Ceil(float64(amount)/rate*3600)))
}

func (s *Service) maxProductionAmount(resourceID, level int) int {
	rate := s.outputPerHour(resourceID, level)
	if rate <= 0 {
		return 1
	}
	return max(1, int(math.Floor(rate*48)))
}

func (s *Service) outputPerHour(resourceID, level int) float64 {
	if level < 1 {
		level = 1
	}
	for _, r := range s.Data.Resources {
		if intFromAny(r["dbLetter"]) == resourceID {
			base := floatFromAny(r["producedPerHourRaw"])
			if base <= 0 {
				return 0
			}
			return base * float64(level)
		}
	}
	return 0
}

func (s *Service) refreshProductionJobs(companyID int) {
	now := s.now().UTC()
	for i := range s.State.ProductionJobs {
		j := &s.State.ProductionJobs[i]
		if j.Status == "claimed" {
			j.ClaimableAmount = 0
			continue
		}
		j.ClaimableAmount = s.claimableAmountForJob(j, now)
		if j.ClaimedAmount >= j.Amount {
			j.Status = "claimed"
			j.ClaimableAmount = 0
			continue
		}
		completeAt, err := time.Parse(time.RFC3339, j.CompletesAt)
		if err == nil && !now.Before(completeAt) {
			j.Status = "ready"
		} else {
			j.Status = "running"
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

func (s *Service) claimableAmountForJob(j *model.ProductionJob, now time.Time) int {
	if j == nil || j.Amount <= 0 || j.Status == "claimed" {
		return 0
	}
	start, errStart := time.Parse(time.RFC3339, j.StartedAt)
	complete, errComplete := time.Parse(time.RFC3339, j.CompletesAt)
	if errStart != nil || errComplete != nil {
		return 0
	}
	totalSeconds := complete.Sub(start).Seconds()
	if totalSeconds <= 0 || !now.Before(complete) {
		return max(0, j.Amount-j.ClaimedAmount)
	}
	if now.Before(start) {
		return 0
	}
	produced := int(math.Floor(now.Sub(start).Seconds() / totalSeconds * float64(j.Amount)))
	if produced > j.Amount {
		produced = j.Amount
	}
	return max(0, produced-j.ClaimedAmount)
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
	inUse := 0
	for _, j := range s.State.ProductionJobs {
		if j.Status == "claimed" {
			continue
		}
		bid := j.BuildingID
		byBuilding[bid] = append(byBuilding[bid], j)
		inUse++
	}
	slots := company.ProductionSlots
	if slots <= 0 {
		slots = s.Cfg.Game.BaseProductionSlots
	}
	return map[string]any{
		"byBuilding": byBuilding,
		"maxSlots":   slots * len(company.PlacedBuildings),
		"inUse":      inUse,
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
	s.addLedger("slot_upgrade", cost, "out", map[string]any{"newSlots": company.ProductionSlots})
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
