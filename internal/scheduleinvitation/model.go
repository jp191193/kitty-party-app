package scheduleinvitation

import "time"

// ScheduleInvitation tracks a single member's RSVP for a kitty schedule entry.
type ScheduleInvitation struct {
	ID          string     `json:"id"`
	ScheduleID  string     `json:"schedule_id"`
	MemberID    string     `json:"member_id"`
	MemberName  string     `json:"member_name,omitempty"`
	Status      string     `json:"status"` // PENDING | ACCEPTED | DATE_PROPOSED
	RespondedAt *time.Time `json:"responded_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
}

// InvitationWithSchedule enriches an invitation with schedule context so
// members can see what event they are responding to.
type InvitationWithSchedule struct {
	ScheduleInvitation
	ScheduledDate time.Time `json:"scheduled_date"`
	GroupID       string    `json:"group_id"`
	GroupName     string    `json:"group_name,omitempty"`
	CycleNumber   int       `json:"cycle_number"`
	CycleMonth    int       `json:"cycle_month"`
	CycleYear     int       `json:"cycle_year"`
}

// SendInvitationsResult summarises the outcome of a bulk invitation send.
type SendInvitationsResult struct {
	ScheduleID  string                `json:"schedule_id"`
	Count       int                   `json:"count"`
	Invitations []*ScheduleInvitation `json:"invitations"`
}

// ProposeDateRequest carries the new proposed date from a member.
type ProposeDateRequest struct {
	ProposedDate time.Time `json:"proposed_date" binding:"required"`
}
