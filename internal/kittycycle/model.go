// Package kittycycle defines models for multi-month kitty party cycles
// and the per-month schedule of hosts/recipients within them.
package kittycycle

import "time"

// KittyCycle is the parent record for a group's multi-month rotation.
type KittyCycle struct {
	ID             string     `json:"id"`
	GroupID        string     `json:"group_id"`
	Name           *string    `json:"name,omitempty"`
	TotalCycles    int        `json:"total_cycles"`
	MonthlyAmount  float64    `json:"monthly_amount"`
	PoolAmount     float64    `json:"pool_amount"`
	StartDate      time.Time  `json:"start_date"`
	EndDate        time.Time  `json:"end_date"`
	Status         string     `json:"status"` // PENDING | ACTIVE | COMPLETED | CANCELLED
	Notes          *string    `json:"notes,omitempty"`
	CreatedBy      string     `json:"created_by"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty"`
}

// KittySchedule is a single month's entry within a KittyCycle — who hosts
// and receives the pool that month. HostMemberID is nil until the host is
// chosen via the dice roll close to the scheduled month.
type KittySchedule struct {
	ID             string     `json:"id"`
	CycleID        string     `json:"cycle_id"`
	GroupID        string     `json:"group_id"`
	CycleNumber    int        `json:"cycle_number"`
	CycleMonth     int        `json:"cycle_month"`
	CycleYear      int        `json:"cycle_year"`
	HostMemberID   *string    `json:"host_member_id,omitempty"`
	ScheduledDate  time.Time  `json:"scheduled_date"`
	DueDate        time.Time  `json:"due_date"`
	PoolAmount     float64    `json:"pool_amount"`
	Status         string     `json:"status"` // SCHEDULED | IN_PROGRESS | COMPLETED | CANCELLED
	Venue          *string    `json:"venue,omitempty"`
	Notes          *string    `json:"notes,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      *time.Time `json:"updated_at,omitempty"`
}

// KittyCycleWithSchedule bundles a cycle with its monthly schedule entries.
type KittyCycleWithSchedule struct {
	KittyCycle
	Schedule []*KittySchedule `json:"schedule"`
}

// ScheduleEntryRequest describes one month within a ScheduleCycleRequest.
// HostMemberID is optional at scheduling time — the host is normally chosen
// later via the dice-roll endpoint.
type ScheduleEntryRequest struct {
	CycleNumber   int       `json:"cycle_number"   binding:"required,min=1"`
	CycleMonth    int       `json:"cycle_month"    binding:"required,min=1,max=12"`
	CycleYear     int       `json:"cycle_year"     binding:"required,min=2000"`
	HostMemberID  *string   `json:"host_member_id,omitempty"`
	ScheduledDate time.Time `json:"scheduled_date" binding:"required"`
	DueDate       time.Time `json:"due_date"       binding:"required"`
	Venue         *string   `json:"venue"`
	Notes         *string   `json:"notes"`
}

// RollHostResult is returned by the dice-roll endpoint with the picked host
// and the updated schedule entry.
type RollHostResult struct {
	ScheduleID     string         `json:"schedule_id"`
	HostMemberID   string         `json:"host_member_id"`
	HostMemberName string         `json:"host_member_name,omitempty"`
	Schedule       *KittySchedule `json:"schedule"`
}

// ScheduleCycleRequest creates a full kitty cycle and all its monthly schedule rows.
type ScheduleCycleRequest struct {
	GroupID       string                  `json:"group_id"        binding:"required"`
	Name          *string                 `json:"name"`
	MonthlyAmount float64                 `json:"monthly_amount"  binding:"required,gt=0"`
	StartDate     time.Time               `json:"start_date"      binding:"required"`
	EndDate       time.Time               `json:"end_date"        binding:"required"`
	Notes         *string                 `json:"notes"`
	Schedule      []ScheduleEntryRequest  `json:"schedule"        binding:"required,min=1,dive"`
}

// CancelCycleRequest carries an optional reason for cancellation.
type CancelCycleRequest struct {
	Notes *string `json:"notes"`
}

// StartCycleResult is returned when a PENDING cycle transitions to ACTIVE.
// It bundles the activated cycle + schedule with downstream records created automatically.
type StartCycleResult struct {
	KittyCycleWithSchedule
	DuesGenerated int    `json:"dues_generated"`
	PayoutID      string `json:"payout_id,omitempty"`
}
