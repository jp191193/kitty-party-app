package scheduleinvitation

// Repository is the persistence contract for schedule invitations.
type Repository interface {
	// BulkCreate inserts one invitation per memberID. Existing rows (same
	// schedule_id + member_id) are left untouched.
	BulkCreate(scheduleID string, memberIDs []string) ([]*ScheduleInvitation, error)

	// ListBySchedule returns all invitations for a schedule entry.
	ListBySchedule(scheduleID string) ([]*ScheduleInvitation, error)

	// ListByMember returns all invitations for a member, enriched with
	// schedule context (date, group, cycle info).
	ListByMember(memberID string) ([]*InvitationWithSchedule, error)

	// GetByID returns a single invitation.
	GetByID(id string) (*ScheduleInvitation, error)

	// Accept sets status = ACCEPTED and records responded_at.
	Accept(id string) (*ScheduleInvitation, error)

	// SetDateProposed sets status = DATE_PROPOSED and records responded_at.
	SetDateProposed(id string) (*ScheduleInvitation, error)
}
