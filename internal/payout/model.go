// Package payout defines models for kitty party pool payouts and disbursements.
package payout

import "time"

// PayoutSchedule is the authoritative record of who receives the pooled amount
// for a given cycle within a kitty party group.
type PayoutSchedule struct {
	ID            string     `json:"id"`
	GroupID       string     `json:"group_id"`
	RecipientID   string     `json:"recipient_id"`
	CycleNumber   int        `json:"cycle_number"`
	CycleMonth    int        `json:"cycle_month"`
	CycleYear     int        `json:"cycle_year"`
	ScheduledDate time.Time  `json:"scheduled_date"`
	PoolAmount    float64    `json:"pool_amount"`
	Status        string     `json:"status"` // SCHEDULED | DISBURSED | CANCELLED
	Notes         *string    `json:"notes,omitempty"`
	CreatedBy     string     `json:"created_by"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     *time.Time `json:"updated_at,omitempty"`
}

// Disbursement records the actual money transfer that settles a PayoutSchedule.
type Disbursement struct {
	ID             string    `json:"id"`
	PayoutID       string    `json:"payout_id"`
	AmountPaid     float64   `json:"amount_paid"`
	PaymentDate    time.Time `json:"payment_date"`
	PaymentMode    string    `json:"payment_mode"` // UPI | CASH | BANK
	TransactionRef *string   `json:"transaction_ref,omitempty"`
	DisbursedBy    string    `json:"disbursed_by"`
	Notes          *string   `json:"notes,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
}

// PayoutSummary combines a schedule entry with its disbursement (if any).
type PayoutSummary struct {
	PayoutSchedule
	Disbursement *Disbursement `json:"disbursement,omitempty"`
}

// CreateScheduleRequest is the payload for scheduling a future payout.
type CreateScheduleRequest struct {
	GroupID       string    `json:"group_id"        binding:"required"`
	RecipientID   string    `json:"recipient_id"    binding:"required"`
	CycleNumber   int       `json:"cycle_number"    binding:"required,min=1"`
	CycleMonth    int       `json:"cycle_month"     binding:"required,min=1,max=12"`
	CycleYear     int       `json:"cycle_year"      binding:"required,min=2000"`
	ScheduledDate time.Time `json:"scheduled_date"  binding:"required"`
	PoolAmount    float64   `json:"pool_amount"     binding:"required,gt=0"`
	Notes         *string   `json:"notes"`
}

// RecordDisbursementRequest is the payload for recording an actual money transfer.
type RecordDisbursementRequest struct {
	PayoutID       string  `json:"payout_id"        binding:"required"`
	AmountPaid     float64 `json:"amount_paid"      binding:"required,gt=0"`
	PaymentMode    string  `json:"payment_mode"     binding:"required,oneof=UPI CASH BANK"`
	TransactionRef *string `json:"transaction_ref"`
	Notes          *string `json:"notes"`
}

// CancelScheduleRequest carries an optional reason for cancellation.
type CancelScheduleRequest struct {
	Notes *string `json:"notes"`
}
