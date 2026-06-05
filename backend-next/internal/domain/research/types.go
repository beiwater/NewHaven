package research

// Project represents a research project that a company can undertake.
type Project struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Category        string  `json:"category"`
	Cost            float64 `json:"cost"`
	DurationSeconds float64 `json:"duration_seconds"`
	Prerequisites   []string `json:"prerequisites,omitempty"`
	Effects         map[string]any `json:"effects,omitempty"`
}

// CompanyProgress tracks a company's progress on a research project.
type CompanyProgress struct {
	CompanyID     int    `json:"company_id"`
	ProjectID     string `json:"project_id"`
	StartedAt     string `json:"started_at"`
	CompletedAt   string `json:"completed_at,omitempty"`
}
