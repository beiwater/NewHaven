// Package executive owns the player-facing executive market and the small,
// explicit leadership system. It intentionally does not implement poaching or
// a global NPC auction: those need durable cross-company transactions and are
// outside the first safe version of the feature.
package executive

import (
	"context"
	"errors"
	"math"

	"github.com/beiwater/NewHaven/backend/internal/apperr"
	"github.com/beiwater/NewHaven/backend/internal/domain/company"
	"github.com/beiwater/NewHaven/backend/internal/platform"
	"github.com/beiwater/NewHaven/backend/internal/storage"
)

// Service handles executive candidates, recruitment, development and role
// assignment. All money-changing writes go through ExecutiveStorage.
type Service struct {
	companies  storage.CompanyStorage
	executives storage.ExecutiveStorage
	clock      platform.Clock
}

func NewService(companies storage.CompanyStorage, executives storage.ExecutiveStorage, clock platform.Clock) *Service {
	return &Service{companies: companies, executives: executives, clock: clock}
}

func (s *Service) MyExecutives(ctx context.Context, companyID int) ([]company.Executive, error) {
	c, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, apperr.NotFound("company not found")
	}
	if c.Executives == nil {
		return []company.Executive{}, nil
	}
	return append([]company.Executive(nil), c.Executives...), nil
}

func (s *Service) MarketCandidates() []Candidate {
	return candidatesForHour(s.clock.Now().UTC())
}

func (s *Service) Recruit(ctx context.Context, companyID int, candidateID string) (*company.Executive, int, error) {
	candidate, ok := candidateByID(s.clock.Now().UTC(), candidateID)
	if !ok {
		return nil, 0, apperr.NotFound("executive candidate expired; refresh the market")
	}
	cost := candidate.RecruitCost
	executive, err := s.executives.RecruitExecutive(ctx, companyID, candidate.Executive, cost)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrAlreadyExists):
			return nil, 0, apperr.Conflict("this executive is already in your company")
		case errors.Is(err, storage.ErrInsufficientFunds):
			return nil, 0, apperr.InsufficientFunds("not enough cash to recruit this executive")
		default:
			return nil, 0, apperr.Internalf("recruit executive: %v", err)
		}
	}
	return executive, cost, nil
}

func (s *Service) Train(ctx context.Context, companyID int, executiveID string) (*company.Executive, int, error) {
	executive, err := s.findExecutive(ctx, companyID, executiveID)
	if err != nil {
		return nil, 0, err
	}
	cost := TrainingCost(executive.Level)
	nextSkills := developSkills(executive.Skills, executive.Specialty)
	updated, err := s.executives.TrainExecutive(ctx, companyID, executiveID, executive.Level, cost, nextSkills)
	if err != nil {
		switch {
		case errors.Is(err, storage.ErrInsufficientFunds):
			return nil, 0, apperr.InsufficientFunds("not enough cash for executive development")
		case errors.Is(err, storage.ErrStateConflict):
			return nil, 0, apperr.Conflict("executive changed; refresh and try again")
		default:
			return nil, 0, apperr.Internalf("train executive: %v", err)
		}
	}
	updated.Stage = stageAtLevel(updated.Level)
	updated.Salary = salaryAtLevel(updated.Level)
	updated.ProductionBonus = productionBonusFor(updated)
	updated.SalesBonus = salesBonusFor(updated)
	updated.MgmtDiscount = managementBonusFor(updated)
	// The atomic storage mutation already persisted the gameplay state (money,
	// level and skills). The remaining fields are display projections; clients
	// derive them from the authoritative level and skill values on future reads.
	return updated, cost, nil
}

func (s *Service) AssignPosition(ctx context.Context, companyID int, executiveID string, position company.ExecutivePosition) (*company.Executive, error) {
	if !validPosition(position) {
		return nil, apperr.Validation("position must be coo, cfo, cmo, cto, or empty")
	}
	updated, err := s.executives.AssignExecutivePosition(ctx, companyID, executiveID, position)
	if err != nil {
		return nil, apperr.NotFound("executive not found")
	}
	return updated, nil
}

func (s *Service) Detail(ctx context.Context, companyID int, executiveID string) (*company.Executive, error) {
	return s.findExecutive(ctx, companyID, executiveID)
}

func (s *Service) findExecutive(ctx context.Context, companyID int, executiveID string) (*company.Executive, error) {
	c, err := s.companies.GetCompany(ctx, companyID)
	if err != nil {
		return nil, apperr.NotFound("company not found")
	}
	for i := range c.Executives {
		if c.Executives[i].ID == executiveID {
			copy := c.Executives[i]
			return &copy, nil
		}
	}
	return nil, apperr.NotFound("executive not found")
}

func validPosition(position company.ExecutivePosition) bool {
	switch position {
	case company.ExecutivePositionUnassigned, company.ExecutivePositionCOO, company.ExecutivePositionCFO, company.ExecutivePositionCMO, company.ExecutivePositionCTO:
		return true
	default:
		return false
	}
}

// TrainingCost is a visible, server-enforced cash cost. Development is
// immediate in the first version; the former UI presented a timer even though
// the backend had already applied a level, which was misleading.
func TrainingCost(level int) int {
	if level < 1 {
		level = 1
	}
	return int(math.Round(5000 * math.Pow(float64(level), 1.6)))
}

func developSkills(skills company.ExecutiveSkills, specialty company.ExecutivePosition) company.ExecutiveSkills {
	skills.Management = math.Min(100, skills.Management+1)
	skills.Accounting = math.Min(100, skills.Accounting+1)
	skills.Communication = math.Min(100, skills.Communication+1)
	skills.Science = math.Min(100, skills.Science+1)
	switch specialty {
	case company.ExecutivePositionCOO:
		skills.Management = math.Min(100, skills.Management+3)
	case company.ExecutivePositionCFO:
		skills.Accounting = math.Min(100, skills.Accounting+3)
	case company.ExecutivePositionCMO:
		skills.Communication = math.Min(100, skills.Communication+3)
	case company.ExecutivePositionCTO:
		skills.Science = math.Min(100, skills.Science+3)
	}
	return skills
}

func stageAtLevel(level int) string {
	switch {
	case level >= 10:
		return "Executive VP"
	case level >= 8:
		return "Director"
	case level >= 6:
		return "Senior Manager"
	case level >= 4:
		return "Manager"
	default:
		return "Associate"
	}
}

func salaryAtLevel(level int) float64 {
	return 600 + 80*math.Pow(float64(level), 1.3)
}

func productionBonusFor(executive *company.Executive) float64 {
	if executive.Specialty != company.ExecutivePositionCTO {
		return 0
	}
	return math.Min(100, executive.Skills.Science*2)
}

func salesBonusFor(executive *company.Executive) float64 {
	if executive.Specialty != company.ExecutivePositionCMO {
		return 0
	}
	return math.Min(50, executive.Skills.Communication/2)
}

func managementBonusFor(executive *company.Executive) float64 {
	if executive.Specialty != company.ExecutivePositionCOO {
		return 0
	}
	return math.Min(100, executive.Skills.Management)
}

func executiveTitle(position company.ExecutivePosition) string {
	switch position {
	case company.ExecutivePositionCOO:
		return "Chief Operating Officer"
	case company.ExecutivePositionCFO:
		return "Chief Financial Officer"
	case company.ExecutivePositionCMO:
		return "Chief Marketing Officer"
	case company.ExecutivePositionCTO:
		return "Chief Technology Officer"
	default:
		return "Executive"
	}
}
