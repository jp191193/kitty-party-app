// Package group defines the domain model for a Kitty Party group.
package group

import "time"

// MaxGroupMembers is the hard cap on participants per kitty party group.
const MaxGroupMembers = 10

// Group represents a single kitty party pool.
type Group struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	MonthlyAmount float64   `json:"monthly_amount"`
	Duration      int       `json:"duration"` // number of months
	StartDate     time.Time `json:"start_date"`
	CreatedBy     string    `json:"created_by"` // member ID of the organiser
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// CreateGroupRequest carries the validated payload for creating a new group.
type CreateGroupRequest struct {
	Name          string    `json:"name"           binding:"required,min=2,max=100"`
	MonthlyAmount float64   `json:"monthly_amount"  binding:"required,gt=0"`
	Duration      int       `json:"duration"        binding:"required,min=1"`
	StartDate     time.Time `json:"start_date"      binding:"required"`
	CreatedBy     string    `json:"created_by"      binding:"required"`
}

// UpdateGroupRequest carries optional fields for patching an existing group.
type UpdateGroupRequest struct {
	Name          string    `json:"name"           binding:"omitempty,min=2,max=100"`
	MonthlyAmount float64   `json:"monthly_amount"  binding:"omitempty,gt=0"`
	Duration      int       `json:"duration"        binding:"omitempty,min=1"`
	StartDate     time.Time `json:"start_date"      binding:"omitempty"`
}

// GroupMembership records a single member's participation in a group.
type GroupMembership struct {
	ID         string    `json:"id"`
	GroupID    string    `json:"group_id"`
	MemberID   string    `json:"member_id"`
	MemberName string    `json:"member_name,omitempty"`
	Status     string    `json:"status"`
	Role       string    `json:"role"`
	JoinedAt   time.Time `json:"joined_at"`
}

// AddMemberRequest is the payload for adding a member to a group.
type AddMemberRequest struct {
	MemberID string `json:"member_id" binding:"required"`
}

// GroupMemberCount contains the number of members in a specific group.
type GroupMemberCount struct {
	IDName
	Count int `json:"count"`
}

type IDName struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
