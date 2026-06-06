package production

import (
	"context"
	"fmt"
	"math"

	"github.com/newhaven/backend-next/internal/apperr"
	"github.com/newhaven/backend-next/internal/domain/finance"
	proddmn "github.com/newhaven/backend-next/internal/domain/production"
	openapi "github.com/newhaven/backend-next/internal/generated/openapi"
)

// ClaimProduction claims available produced resources from a production job.
// It validates ownership, checks claimable amount, adds output to inventory,
// awards incremental XP, and appends a finance ledger entry.
func (s *Service) ClaimProduction(ctx context.Context, companyID int, jobID string) (*openapi.ClaimProductionResponse, error) {
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
	if job.ClaimableAmount <= 0 {
		return nil, apperr.Validation("nothing to claim yet")
	}

	claimAmount := job.ClaimableAmount

	// Add produced resource to inventory.
	if err := s.companies.UpdateInventory(ctx, companyID, job.ResourceID, claimAmount); err != nil {
		return nil, apperr.Internalf("add output to inventory: %v", err)
	}

	// Update job fields.
	job.ClaimedAmount += claimAmount
	if job.ClaimedAmount >= job.TargetQuantity {
		job.ClaimedAmount = job.TargetQuantity
		job.ClaimableAmount = 0
		job.Status = proddmn.StatusClaimed
	} else {
		job.ClaimableAmount = 0
	}

	if err := s.production.UpdateJob(ctx, job); err != nil {
		return nil, apperr.Internalf("update job: %v", err)
	}

	// Award incremental production XP.
	xpEarned := s.productionXPForClaim(job, claimAmount)
	if xpEarned > 0 {
		company, err := s.companies.GetCompany(ctx, companyID)
		if err != nil {
			return nil, apperr.Internalf("get company for XP: %v", err)
		}
		company.XP += int64(xpEarned)
		if err := s.companies.UpdateCompany(ctx, company); err != nil {
			return nil, apperr.Internalf("save company XP: %v", err)
		}
		// Persist XP awarded on job for future incremental calculation.
		job.XPAwarded += xpEarned
		if err := s.production.UpdateJob(ctx, job); err != nil {
			return nil, apperr.Internalf("update job xp: %v", err)
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

	// Look up current level for response.
	company, err := s.companies.GetCompany(ctx, companyID)
	if err == nil {
		lvl := company.Level
		resp.Level = &lvl
	}

	return resp, nil
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
