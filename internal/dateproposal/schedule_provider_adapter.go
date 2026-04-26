package dateproposal

import "time"

// ScheduleInfoProvider abstracts the kitty_schedule data this domain needs.
type ScheduleInfoProvider interface {
	// GetGroupIDForSchedule returns the group_id for a kitty_schedule entry.
	GetGroupIDForSchedule(scheduleID string) (string, error)
	// GetGroupMemberCount returns the number of active members in a group.
	GetGroupMemberCount(groupID string) (int, error)
	// UpdateScheduledDate sets a new scheduled_date on a kitty_schedule row.
	UpdateScheduledDate(scheduleID string, newDate time.Time) error
	// IsAdminOrCreator returns true if callerID created or is ADMIN of groupID.
	IsAdminOrCreator(groupID, callerID string) (bool, error)
}
