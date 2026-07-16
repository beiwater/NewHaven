package production

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	"github.com/beiwater/NewHaven/backend/internal/apperr"
	"github.com/beiwater/NewHaven/backend/internal/domain/company"
	"github.com/beiwater/NewHaven/backend/internal/domain/finance"
	proddmn "github.com/beiwater/NewHaven/backend/internal/domain/production"
	"github.com/beiwater/NewHaven/backend/internal/formula"
	openapi "github.com/beiwater/NewHaven/backend/internal/generated/openapi"
	"github.com/beiwater/NewHaven/backend/internal/storage"
)

// ClaimProduction claims available produced resources from a production job.
// It validates ownership, checks claimable amount, adds output to inventory,
// awards incremental XP, and appends a finance ledger entry.
func (s *Service) ClaimProduction(ctx context.Context, companyID int, jobID string) (*openapi.ClaimProductionResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.claimProductionLocked(ctx, companyID, jobID)
}

// claimProductionLocked is the internal implementation of ClaimProduction.
// Caller MUST hold s.mu.
func (s *Service) claimProductionLocked(ctx context.Context, companyID int, jobID string) (*openapi.ClaimProductionResponse, error) {
	// Refresh job statuses so time-based computations are current.
	if err := s.refreshJobStatuses(ctx, companyID); err != nil {
		return nil, apperr.Internalf("refresh statuses: %v", err)
	}

	job, err := s.production.GetJob(ctx, jobID)
	if err != nil {
		return nil, apperr.NotFound("job not found")
	}
	if job.CompanyID != companyID {
		// Do not leak existence of other companies' jobs.
		return nil, apperr.NotFound("job not found")
	}
	if job.Status == proddmn.StatusClaimed {
		return nil, apperr.Conflict("job already claimed")
	}
	if job.Status == proddmn.StatusCancelled {
		return nil, apperr.Conflict("job was cancelled")
	}
	if job.ClaimableAmount <= 0 {
		return nil, apperr.Validation("nothing to claim yet")
	}

	claimAmount := job.ClaimableAmount
	payroll, workers, hourlyWage := s.productionPayrollSettlement(ctx, companyID, job)

	// Calculate the post-claim XP before committing, then atomically move the
	// output, XP, and job lifecycle together in storage. The expected amount is
	// an optimistic guard against another service instance claiming first.
	projectedJob := *job
	projectedJob.ClaimedAmount += claimAmount
	if projectedJob.ClaimedAmount > projectedJob.TargetQuantity {
		projectedJob.ClaimedAmount = projectedJob.TargetQuantity
	}
	xpEarned := s.productionXPForClaim(&projectedJob, claimAmount)
	job, err = s.production.ClaimProductionOutput(ctx, companyID, jobID, claimAmount, xpEarned, payroll)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrAlreadySettled):
			return nil, apperr.Conflict("job already claimed")
		case errors.Is(err, storage.ErrNothingToClaim):
			return nil, apperr.Validation("nothing to claim yet")
		case errors.Is(err, storage.ErrStateConflict):
			return nil, apperr.Conflict("production claim changed; refresh and try again")
		default:
			return nil, apperr.Internalf("commit production claim: %v", err)
		}
	}

	// Append finance ledger entry.
	isPartial := job.Status != proddmn.StatusClaimed
	metadata := map[string]any{
		"resourceId": job.ResourceID,
		"jobId":      job.ID,
		"partial":    isPartial,
	}
	if s.finance != nil {
		if payroll.Amount > 0 {
			if err := s.finance.AppendLedgerEntry(ctx, &finance.LedgerEntry{
				CompanyID: companyID,
				Kind:      "production_wages",
				Amount:    payroll.Amount,
				Direction: "out",
				Metadata: map[string]any{
					"jobId":         job.ID,
					"workers":       workers,
					"hourlyWage":    hourlyWage,
					"activeSeconds": payroll.SettledSeconds - payroll.ExpectedSeconds,
				},
				CreatedAt: s.clock.Now().UTC().Format(time.RFC3339),
			}); err != nil {
				return nil, apperr.Internalf("append production payroll ledger: %v", err)
			}
		}
		if err := s.finance.AppendLedgerEntry(ctx, &finance.LedgerEntry{
			CompanyID: companyID,
			Kind:      "production_output",
			Amount:    float64(claimAmount),
			Direction: "in",
			Metadata:  metadata,
		}); err != nil {
			return nil, apperr.Internalf("append production ledger: %v", err)
		}
	}
	output := map[string]int{fmt.Sprintf("%d", job.ResourceID): claimAmount}
	remaining := job.TargetQuantity - job.ClaimedAmount
	if remaining < 0 {
		remaining = 0
	}
	claimRespStatus := openapi.ClaimProductionResponseStatus(job.Status)

	resp := &openapi.ClaimProductionResponse{
		JobId:         &job.ID,
		Status:        &claimRespStatus,
		Output:        &output,
		ClaimedAmount: &job.ClaimedAmount,
		Remaining:     &remaining,
		Xp:            &xpEarned,
	}

	// Look up current level for response and auto-complete arrival story if in progress.
	comp, compErr := s.companies.GetCompany(ctx, companyID)
	if compErr == nil {
		lvl := comp.Level
		resp.Level = &lvl

		// Auto-complete the arrival story on the first production claim.
		prefs := comp.Preferences
		if prefs == nil {
			prefs = make(map[string]any)
		}
		stories, _ := prefs["storyProgress"].(map[string]any)
		if stories == nil {
			stories = make(map[string]any)
		}
		current, _ := stories[company.ArrivalStoryID].(map[string]any)
		if current != nil {
			if status, _ := current["status"].(string); status == "in_progress" {
				current["status"] = "completed"
				prefs["storyProgress"] = stories
				comp.Preferences = prefs

				// Apply newbie level-up.
				if s.cfg != nil && comp.Level < s.cfg.NewbieLevelUpTo {
					comp.Level = s.cfg.NewbieLevelUpTo
				}

				if saveErr := s.companies.UpdateCompany(ctx, comp); saveErr != nil {
					_ = saveErr
				}
			}
		}
	}

	return resp, nil
}

// productionPayrollSettlement calculates the unbilled portion of a job's
// active time. Settlement is deferred until a player returns to claim or
// cancel, but it is always capped at the job duration and then committed with
// the same compare-and-set mutation as the job lifecycle.
func (s *Service) productionPayrollSettlement(ctx context.Context, companyID int, job *proddmn.ProductionJob) (proddmn.PayrollSettlement, int, float64) {
	settlement := proddmn.PayrollSettlement{}
	if job == nil {
		return settlement, 0, 0
	}
	settlement.ExpectedSeconds = job.PayrollSettledSeconds
	settlement.SettledSeconds = job.PayrollSettledSeconds

	comp, err := s.companies.GetCompany(ctx, companyID)
	if err != nil || comp == nil {
		return settlement, 0, 0
	}
	var building *company.Building
	for i := range comp.Buildings {
		if comp.Buildings[i].ID == job.BuildingID {
			building = &comp.Buildings[i]
			break
		}
	}
	if building == nil || job.DurationSeconds <= 0 {
		return settlement, 0, 0
	}

	activeSeconds := s.clock.Now().Sub(job.StartedAt).Seconds()
	activeSeconds = math.Max(0, math.Min(activeSeconds, job.DurationSeconds))
	if activeSeconds <= job.PayrollSettledSeconds {
		return settlement, 0, 0
	}
	kind := building.Kind
	if kind <= 0 {
		kind = building.BuildingID
	}
	workers := formula.BuildingWorkerCount(kind, building.Level)
	hourlyWage := formula.BuildingHourlyWage(kind, building.Level)
	settlement.SettledSeconds = activeSeconds
	settlement.Amount = hourlyWage * (activeSeconds - job.PayrollSettledSeconds) / 3600
	return settlement, workers, hourlyWage
}

// productionXPForClaim computes the incremental XP earned for this claim.
// Total XP per job is 10, awarded proportionally to the fraction claimed.
func (s *Service) productionXPForClaim(job *proddmn.ProductionJob, claimAmount int) int {
	if job == nil || job.TargetQuantity <= 0 || claimAmount <= 0 {
		return 0
	}
	totalXP := 10
	// Fraction of total job completed including this claim.
	completedFraction := float64(job.ClaimedAmount) / float64(job.TargetQuantity)
	if completedFraction > 1.0 {
		completedFraction = 1.0
	}
	totalEarned := int(math.Floor(completedFraction * float64(totalXP)))
	if totalEarned > totalXP {
		totalEarned = totalXP
	}
	xp := totalEarned - job.XPAwarded
	if xp < 0 {
		return 0
	}
	return xp
}
