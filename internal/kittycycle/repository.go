// Package kittycycle – repository layer.
package kittycycle

// Repository defines the persistence contract for KittyCycle and KittySchedule.
type Repository interface {
	// CreateCycleWithSchedule atomically inserts a kitty_cycle row and all
	// its kitty_schedule rows. Returns the persisted cycle with schedule entries.
	CreateCycleWithSchedule(cycle *KittyCycle, entries []*KittySchedule) (*KittyCycleWithSchedule, error)

	// GetCycle returns a single cycle by ID, enriched with its schedule rows.
	GetCycle(id string) (*KittyCycleWithSchedule, error)

	// ListByGroup returns all cycles for a group (without schedule details).
	ListByGroup(groupID string) ([]*KittyCycle, error)

	// ListSchedule returns schedule rows for a cycle ordered by cycle_number.
	ListSchedule(cycleID string) ([]*KittySchedule, error)

	// CancelCycle sets a cycle (and its SCHEDULED entries) to CANCELLED.
	// Fails if the cycle is already COMPLETED or CANCELLED.
	CancelCycle(id string, notes *string) error

	// StartCycle transitions a PENDING cycle to ACTIVE and marks its first
	// schedule entry (lowest cycle_number) as IN_PROGRESS.
	// Fails if the cycle is not in PENDING status.
	StartCycle(id string) (*KittyCycleWithSchedule, error)

	// GetSchedule returns a single kitty_schedule entry by ID.
	GetSchedule(scheduleID string) (*KittySchedule, error)

	// AssignHost atomically sets host_member_id on a schedule entry.
	// Fails if the entry already has a host assigned.
	AssignHost(scheduleID, hostMemberID string) (*KittySchedule, error)

	// ListAssignedHosts returns the set of host_member_ids already chosen
	// for a cycle (used to exclude them from the dice roll).
	ListAssignedHosts(cycleID string) (map[string]bool, error)
}
