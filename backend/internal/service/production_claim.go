package service

import (
	"fmt"

	"go-sim-api/internal/anticheat"
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
		if j.Status != "ready" {
			return nil, fmt.Errorf("job not ready")
		}
		for k, q := range j.Output {
			s.inventoryAdd(company, k, j.Quality, q)
		}
		j.Status = "claimed"
		s.addLedger("production_output", float64(j.Amount), "in", map[string]any{"resourceId": j.ResourceID, "jobId": j.ID, "quality": j.Quality})
		s.addXP(company, 10)
		s.saveCompanyLocked(company)
		return map[string]any{"jobId": j.ID, "status": j.Status, "output": j.Output, "quality": j.Quality}, nil
	}
	return nil, fmt.Errorf("job not found")
}
