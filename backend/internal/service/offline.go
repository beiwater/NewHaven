package service

import (
	"log"
	"math"
	"time"
)

func (s *Service) CalculateOfflineIncome(companyID int) map[string]any {
	s.mu.Lock()
	defer s.mu.Unlock()
	company := s.getCompanyLocked(companyID)
	if company == nil {
		return map[string]any{"offlineHours": 0.0, "earned": 0.0, "produced": map[int]int{}}
	}
	if s.State.LastActiveAt == "" {
		s.State.LastActiveAt = s.now().UTC().Format(time.RFC3339)
		return map[string]any{"offlineHours": 0.0, "earned": 0.0, "produced": map[int]int{}}
	}
	lastActive, err := time.Parse(time.RFC3339, s.State.LastActiveAt)
	if err != nil {
		return map[string]any{"offlineHours": 0.0, "earned": 0.0, "produced": map[int]int{}}
	}
	now := s.now().UTC()
	offlineHours := now.Sub(lastActive).Hours()
	if offlineHours < 0.1 {
		// Less than 6 minutes offline — skip
		return map[string]any{"offlineHours": offlineHours, "earned": 0.0, "produced": map[int]int{}}
	}
	// Cap at 8 hours max offline income
	if offlineHours > 8 {
		offlineHours = 8
	}

	// Calculate production that would have completed
	totalProduced := map[int]int{}
	maxCapacity := 10000.0
	for _, j := range s.State.ProductionJobs {
		if j.Status != "running" {
			continue
		}
		started, err := time.Parse(time.RFC3339, j.StartedAt)
		if err != nil {
			continue
		}
		completes, err := time.Parse(time.RFC3339, j.CompletesAt)
		if err != nil {
			continue
		}
		duration := completes.Sub(started).Hours()
		if duration <= 0 {
			continue
		}
		elapsed := now.Sub(started).Hours()
		completeCycles := int(math.Floor(elapsed / duration))
		if completeCycles > 0 {
			baseQty := j.Amount
			for k := range j.Output {
				produced := baseQty * completeCycles
				if float64(totalProduced[k])+float64(produced) > maxCapacity {
					produced = int(maxCapacity) - totalProduced[k]
					if produced <= 0 {
						continue
					}
				}
				// Add to inventory at the job's quality
				s.inventoryAdd(company, k, j.Quality, produced)
				totalProduced[k] += produced
				// Reset job timer (advance one cycle)
				newStart := started.Add(time.Duration(completeCycles) * time.Duration(duration*3600) * time.Second)
				newComplete := newStart.Add(time.Duration(duration*3600) * time.Second)
				j.StartedAt = newStart.Format(time.RFC3339)
				j.CompletesAt = newComplete.Format(time.RFC3339)
				j.Status = "running"
			}
		}
	}

	// Calculate passive income from bonds
	bondIncome := 0.0
	for _, b := range s.State.Bonds {
		if b.OwnerCompanyID == company.ID && b.IssuerCompanyID != company.ID {
			daily := math.Floor(float64(b.Amount) * 50.0 * b.Interest * 100)
			bondIncome += daily * (offlineHours / 24.0)
		}
	}
	if bondIncome > 0 {
		company.Money += bondIncome
	}

	s.State.LastActiveAt = now.Format(time.RFC3339)
	s.saveCompanyLocked(company)
	s.saveOrdersLocked()

	log.Printf("[offline] user was offline %.1fh, produced %d items, bond income %.2f",
		offlineHours, len(totalProduced), bondIncome)
	return map[string]any{
		"offlineHours": math.Round(offlineHours*10) / 10,
		"earned":       math.Round(bondIncome*100) / 100,
		"produced":     totalProduced,
	}
}
