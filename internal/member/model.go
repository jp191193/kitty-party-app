// Package member defines the domain model for a Kitty Party member.
package member

import "time"

// Member represents a participant in a kitty party group.
type Member struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// CreateMemberRequest carries the payload for creating a new member.
type CreateMemberRequest struct {
	Name  string `json:"name"  binding:"required,min=2,max=100"`
	Email string `json:"email" binding:"required,email"`
	Phone string `json:"phone" binding:"required,min=10,max=15"`
}

// UpdateMemberRequest carries the payload for updating an existing member.
type UpdateMemberRequest struct {
	Name  string `json:"name"  binding:"omitempty,min=2,max=100"`
	Email string `json:"email" binding:"omitempty,email"`
	Phone string `json:"phone" binding:"omitempty,min=10,max=15"`
}
