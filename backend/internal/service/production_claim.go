package service

import (
	"fmt"

	"go-sim-api/internal/anticheat"
	"go-sim-api/internal/model"
)

func (s *Service) ClaimProduction(companyID int, jobID string) (map[string]any, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return nil, fmt.Errorf("company not found")
	}
	pid := company.ID
	if ok, msg := s.AC.CheckRateLimit(pid); !ok {
		return nil, fmt.Errorf("cheat detected: %s", msg)
	}
	s.AC.RecordAction(pid, anticheat.ActClaimProd, fmt.Sprintf("job=%s", jobID))
	s.SD.RecordAction(pid)
	s.refreshProductionJobs(companyID)
	for i := range s.State.ProductionJobs {
		j := &s.State.ProductionJobs[i]
		if j.ID != jobID {
			continue
		}
		if j.Status == "claimed" {
			return nil, fmt.Errorf("job already claimed")
		}
		available := j.ClaimableAmount
		if available <= 0 {
			available = s.claimableAmountForJob(j, s.now().UTC())
		}
		if available <= 0 {
			return nil, fmt.Errorf("nothing to collect yet")
		}
		for k := range j.Output {
			s.inventoryAdd(company, k, j.Quality, available)
		}
		j.ClaimedAmount += available
		if j.ClaimedAmount >= j.Amount {
			j.ClaimedAmount = j.Amount
			j.ClaimableAmount = 0
			j.Status = "claimed"
		} else {
			s.refreshProductionJobs(companyID)
		}
		xpEarned := s.productionXPForClaim(j)
		j.XPAwarded += xpEarned
		s.addLedger("production_output", float64(available), "in", map[string]any{"resourceId": j.ResourceID, "jobId": j.ID, "quality": j.Quality, "partial": j.Status != "claimed"})
		if xpEarned > 0 {
			s.addXP(company, xpEarned)
		}
		marketUnlocked := false
		if company.Level < 2 {
			company.Level = 2
			if company.XpToNextLevel < company.Level*100 {
				company.XpToNextLevel = company.Level * 100
			}
			marketUnlocked = true
			s.addLedger("level_up", 0, "in", map[string]any{"newLevel": company.Level, "reason": "first_harvest"})
		}
		s.saveCompanyLocked(company)
		return map[string]any{
			"jobId":          j.ID,
			"status":         j.Status,
			"output":         map[int]int{j.ResourceID: available},
			"quality":        j.Quality,
			"claimedAmount":  j.ClaimedAmount,
			"remaining":      max(0, j.Amount-j.ClaimedAmount),
			"xp":             xpEarned,
			"level":          company.Level,
			"marketUnlocked": marketUnlocked,
		}, nil
	}
	return nil, fmt.Errorf("job not found")
}

func (s *Service) productionXPForClaim(j *model.ProductionJob) int {
	if j == nil || j.Amount <= 0 {
		return 0
	}
	totalXP := 10
	earnedSoFar := int(float64(j.ClaimedAmount) / float64(j.Amount) * float64(totalXP))
	if j.ClaimedAmount >= j.Amount {
		earnedSoFar = totalXP
	}
	xp := earnedSoFar - j.XPAwarded
	if xp < 0 {
		return 0
	}
	return xp
}
