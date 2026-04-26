package contribution

import "time"

// ContributionDue represents the monthly financial obligation of a member.
type ContributionDue struct {
	ID             string    `json:"id"`
	GroupID        string    `json:"group_id"`
	MemberID       string    `json:"member_id"`
	HostMemberID   string    `json:"host_member_id"`
	CycleMonth     int       `json:"cycle_month"`
	CycleYear      int       `json:"cycle_year"`
	CycleNumber    int       `json:"cycle_number"`
	KittyPartyDate time.Time `json:"kitty_party_date"`
	DueDate        time.Time `json:"due_date"`
	Amount         float64   `json:"amount"`
	Status     string    `json:"status"` // PENDING, PAID, OVERDUE
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty"`
}

// ContributionPayment represents an actual money transaction towards a due.
type ContributionPayment struct {
	ID             string    `json:"id"`
	DueID          string    `json:"due_id"`
	AmountPaid     float64   `json:"amount_paid"`
	PaymentDate    time.Time `json:"payment_date"`
	PaymentMode    string    `json:"payment_mode"` // UPI, CASH, BANK
	TransactionRef *string   `json:"transaction_ref,omitempty"`
	Status         string    `json:"status"` // SUCCESS, FAILED
	CreatedAt      time.Time `json:"created_at"`
}

// RecordPaymentRequest is the payload for adding a new payment.
type RecordPaymentRequest struct {
	DueID          string  `json:"due_id" binding:"required"`
	AmountPaid     float64 `json:"amount_paid" binding:"required,gt=0"`
	PaymentMode    string  `json:"payment_mode" binding:"required,oneof=UPI CASH BANK"`
	TransactionRef *string `json:"transaction_ref"`
	Status         string  `json:"status" binding:"required,oneof=SUCCESS FAILED"`
}

// GenerateDuesRequest is the payload for generating monthly dues for a group.
type GenerateDuesRequest struct {
	GroupID        string    `json:"group_id" binding:"required"`
	HostMemberID   string    `json:"host_member_id" binding:"required"`
	CycleMonth     int       `json:"cycle_month" binding:"required,min=1"`
	CycleYear      int       `json:"cycle_year" binding:"required,min=2000"`
	CycleNumber    int       `json:"cycle_number" binding:"required,min=1"`
	KittyPartyDate time.Time `json:"kitty_party_date" binding:"required"`
	DueDate        time.Time `json:"due_date" binding:"required"`
	Amount         float64   `json:"amount" binding:"required,gt=0"`
}
