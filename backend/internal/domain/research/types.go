package research

// MaxResearchLevel is the maximum research level for any resource.
const MaxResearchLevel = 100

// ResourceResearch tracks a company's research level for a specific resource.
type ResourceResearch struct {
	CompanyID  int `json:"company_id"`
	ResourceID int `json:"resource_id"`
	Level      int `json:"level"` // 0 = not researched, 1-100
}
