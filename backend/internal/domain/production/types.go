package production

import "time"

// JobStatus represents the lifecycle of a production job.
type JobStatus string

const (
	StatusRunning   JobStatus = "running"
	StatusReady     JobStatus = "ready"
	StatusClaimed   JobStatus = "claimed"
	StatusCancelled JobStatus = "cancelled"
)

// ProductionJob represents a single production run.
type ProductionJob struct {
	ID              string    `json:"id"`
	ClientRequestID string    `json:"client_request_id,omitempty"`
	CompanyID       int       `json:"company_id"`
	BuildingID      string    `json:"building_id"`
	ResourceID      int       `json:"resource_id"`
	Quantity        int       `json:"quantity"`
	TargetQuantity  int       `json:"target_quantity"`
	StartedAt       time.Time `json:"started_at"`
	DurationSeconds float64   `json:"duration_seconds"`
	ClaimedAmount   int       `json:"claimed_amount"`
	ClaimableAmount int       `json:"claimable_amount"`
	XPAwarded       int       `json:"xp_awarded"`
	Status          JobStatus `json:"status"`
}

// Slot represents a building's production slot.
type Slot struct {
	Index int            `json:"index"`
	Job   *ProductionJob `json:"job,omitempty"`
}

// ProductionOption represents a producible resource in a building.
type ProductionOption struct {
	ResourceID      int     `json:"resource_id"`
	Name            string  `json:"name"`
	PerHour         float64 `json:"per_hour"`
	DurationSeconds float64 `json:"duration_seconds"`
}

// StartProductionRequest is the DTO for starting production.
type StartProductionRequest struct {
	RequestID  string `json:"requestId,omitempty" validate:"max=128"`
	BuildingID string `json:"building_id" validate:"required"`
	ResourceID int    `json:"resource_id" validate:"required"`
	Quantity   int    `json:"quantity" validate:"required,min=1"`
}
