// Package contributiontracking provides a per-member, per-cycle summary
// of contribution dues and payments.
package contributiontracking

import "time"

// Status values mirror the DB CHECK constraint on contribution_tracking.status.
const (
	StatusPending = "PENDING"
	StatusPartial = "PARTIAL"
	StatusPaid    = "PAID"
	StatusOverdue = "OVERDUE"
)

// ContributionTracking is a per-member rollup for a single kitty cycle.
type ContributionTracking struct {
	ID              string     `json:"id"`
	GroupID         string     `json:"group_id"`
	MemberID        string     `json:"member_id"`
	DueID           string     `json:"due_id"`
	CycleMonth      int        `json:"cycle_month"`
	CycleYear       int        `json:"cycle_year"`
	CycleNumber     int        `json:"cycle_number"`
	TotalDue        float64    `json:"total_due"`
	TotalPaid       float64    `json:"total_paid"`
	Balance         float64    `json:"balance"`
	Status          string     `json:"status"`
	LastPaymentDate *time.Time `json:"last_payment_date,omitempty"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// CreateTrackingRequest is the payload for initializing a tracking row.
type CreateTrackingRequest struct {
	GroupID     string  `json:"group_id"     binding:"required"`
	MemberID    string  `json:"member_id"    binding:"required"`
	DueID       string  `json:"due_id"       binding:"required"`
	CycleMonth  int     `json:"cycle_month"  binding:"required,min=1,max=12"`
	CycleYear   int     `json:"cycle_year"   binding:"required,min=2000"`
	CycleNumber int     `json:"cycle_number" binding:"required,min=1"`
	TotalDue    float64 `json:"total_due"    binding:"required,gt=0"`
}

// UpdateTrackingRequest applies a payment delta to the tracking row.
type UpdateTrackingRequest struct {
	AmountPaid      float64    `json:"amount_paid"       binding:"required,gt=0"`
	LastPaymentDate *time.Time `json:"last_payment_date"`
}
