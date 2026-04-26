// Package payout – repository layer.
package payout

// Repository defines the data-access contract for payout schedules and disbursements.
type Repository interface {
	// CreateSchedule inserts a new payout_schedule row.
	// Returns apperrors.ErrConflict when the (group_id, cycle_number) pair already exists.
	CreateSchedule(p *PayoutSchedule) (*PayoutSchedule, error)

	// GetSchedule fetches a single schedule by its ID.
	GetSchedule(id string) (*PayoutSchedule, error)

	// ListByGroup returns all schedule rows for a group, ordered by cycle_number ASC.
	ListByGroup(groupID string) ([]*PayoutSchedule, error)

	// ListByRecipient returns all schedules where the given member is recipient.
	ListByRecipient(recipientID string) ([]*PayoutSchedule, error)

	// RecordDisbursement inserts a disbursement and atomically sets the payout
	// status to DISBURSED. Returns an error if the payout is not in SCHEDULED state.
	RecordDisbursement(d *Disbursement) (*Disbursement, error)

	// GetDisbursementByPayout returns the disbursement for a payout, or nil if none exists.
	GetDisbursementByPayout(payoutID string) (*Disbursement, error)

	// CancelSchedule sets the payout status to CANCELLED.
	// Returns an error if the payout is not in SCHEDULED state.
	CancelSchedule(id string, notes *string) error
}
