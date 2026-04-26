// Package profile defines the domain model for a kitty party member's extended profile.
package profile

import "time"

// MemberProfile holds optional extended information about a member.
type MemberProfile struct {
	ID                string     `json:"id"`
	MemberID          string     `json:"member_id"`
	DateOfBirth       *time.Time `json:"date_of_birth,omitempty"`
	ProfilePictureURL *string    `json:"profile_picture_url,omitempty"`
	Address           *string    `json:"address,omitempty"`
	City              *string    `json:"city,omitempty"`
	State             *string    `json:"state,omitempty"`
	Pincode           *string    `json:"pincode,omitempty"`
	Occupation        *string    `json:"occupation,omitempty"`
	About             *string    `json:"about,omitempty"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
}

// UpsertProfileRequest carries the payload for creating or updating a member profile.
// All fields are optional; only supplied fields are persisted.
type UpsertProfileRequest struct {
	// DateOfBirth in "YYYY-MM-DD" format.
	DateOfBirth       *string `json:"date_of_birth"        binding:"omitempty"`
	ProfilePictureURL *string `json:"profile_picture_url"  binding:"omitempty,max=2048"`
	Address           *string `json:"address"              binding:"omitempty,max=255"`
	City              *string `json:"city"                 binding:"omitempty,max=100"`
	State             *string `json:"state"                binding:"omitempty,max=100"`
	Pincode           *string `json:"pincode"              binding:"omitempty,max=10"`
	Occupation        *string `json:"occupation"           binding:"omitempty,max=100"`
	About             *string `json:"about"                binding:"omitempty,max=500"`
}

// MemberProfileResponse combines basic member information with extended profile data.
type MemberProfileResponse struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Email     string         `json:"email"`
	Phone     string         `json:"phone"`
	Profile   *MemberProfile `json:"profile"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// GroupEntry is a lightweight summary of one group membership used in ProfileSummary.
type GroupEntry struct {
	GroupID   string    `json:"group_id"`
	GroupName string    `json:"group_name"`
	Role      string    `json:"role"`
	Status    string    `json:"status"`
	JoinedAt  time.Time `json:"joined_at"`
}

// ProfileSummary is a rich read-only view combining member info, extended profile,
// group memberships, and financial contribution statistics.
type ProfileSummary struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	Email       string         `json:"email"`
	Phone       string         `json:"phone"`
	Profile     *MemberProfile `json:"profile"`
	Groups      []GroupEntry   `json:"groups"`
	GroupCount  int            `json:"group_count"`
	TotalPaid   float64        `json:"total_paid"`
	PendingDues int            `json:"pending_dues"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}
