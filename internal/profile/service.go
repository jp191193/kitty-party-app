// Package profile – service (business-logic) layer.
package profile

import "kitty-party-app/internal/apperrors"

// MemberChecker is a minimal interface so the profile service can verify that
// a member exists without importing the full member package (avoids circular deps).
type MemberChecker interface {
	Exists(memberID string) error
}

// Service defines the use-case contract for the MemberProfile domain.
type Service interface {
	// GetProfile returns the member's basic info merged with their extended profile.
	// Returns apperrors.ErrNotFound when the member does not exist.
	GetProfile(memberID string) (*MemberProfileResponse, error)

	// UpsertProfile creates or updates the extended profile for the given member.
	// Only the authenticated member may update their own profile (ownership enforced
	// at the handler layer).
	UpsertProfile(memberID string, req UpsertProfileRequest) (*MemberProfile, error)

	// GetSummary returns a full dashboard view: member info, extended profile,
	// group memberships, and contribution statistics.
	GetSummary(memberID string) (*ProfileSummary, error)
}

type service struct {
	repo    Repository
	checker MemberChecker
}

// NewService constructs a profile Service.
func NewService(repo Repository, checker MemberChecker) Service {
	return &service{repo: repo, checker: checker}
}

func (s *service) GetProfile(memberID string) (*MemberProfileResponse, error) {
	// Delegate to GetSummary which already fetches member + profile in one query,
	// then project only the fields needed for MemberProfileResponse.
	summary, err := s.repo.GetSummary(memberID)
	if err != nil {
		return nil, err
	}
	return &MemberProfileResponse{
		ID:        summary.ID,
		Name:      summary.Name,
		Email:     summary.Email,
		Phone:     summary.Phone,
		Profile:   summary.Profile,
		CreatedAt: summary.CreatedAt,
		UpdatedAt: summary.UpdatedAt,
	}, nil
}

func (s *service) UpsertProfile(memberID string, req UpsertProfileRequest) (*MemberProfile, error) {
	// Verify the member exists before attempting the upsert.
	if err := s.checker.Exists(memberID); err != nil {
		return nil, apperrors.ErrNotFound
	}
	return s.repo.Upsert(memberID, req)
}

func (s *service) GetSummary(memberID string) (*ProfileSummary, error) {
	return s.repo.GetSummary(memberID)
}
