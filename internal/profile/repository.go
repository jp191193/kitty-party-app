// Package profile – repository layer.
package profile

// Repository defines the data-access contract for the MemberProfile domain.
type Repository interface {
	// FindByMemberID returns the extended profile for the given member.
	// Returns apperrors.ErrNotFound when no profile has been created yet.
	FindByMemberID(memberID string) (*MemberProfile, error)

	// Upsert creates the profile if it does not exist, or updates it if it does.
	Upsert(memberID string, req UpsertProfileRequest) (*MemberProfile, error)

	// GetSummary returns a fully assembled ProfileSummary for the given member,
	// including group memberships and contribution statistics.
	GetSummary(memberID string) (*ProfileSummary, error)
}
