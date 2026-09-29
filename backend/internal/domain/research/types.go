package research

// MaxResearchLevel is the highest product quality a company may unlock.
// Q0 is available without research, so paid research advances Q1 through Q12.
const MaxResearchLevel = 12

// ResourceResearch tracks a company's research level for a specific resource.
type ResourceResearch struct {
	CompanyID  int `json:"company_id"`
	ResourceID int `json:"resource_id"`
	Level      int `json:"level"` // highest unlocked product quality, Q0-Q12
}
