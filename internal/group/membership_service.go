// Package group – membership service (business-logic) layer.
//
// MembershipService enforces all domain rules:
//   - The target group must exist.
//   - A member cannot be added twice to the same group.
//   - A group cannot exceed MaxGroupMembers participants.
//
// It depends on both Repository (to validate group existence) and
// MembershipRepository (to manage membership records). This avoids
// cross-package coupling while keeping the rules in one place.
package group

import (
	"fmt"
	"net/http"

	"kitty-party-app/internal/apperrors"
)

// MembershipService defines the use-case contract for group membership.
type MembershipService interface {
	// ListMembers returns all memberships for the given group.
	ListMembers(groupID string) ([]*GroupMembership, error)

	// AddMember validates and records a new member in the group.
	AddMember(groupID string, executingUserID string, req AddMemberRequest) (*GroupMembership, error)

	// CountMembers returns the total number of members in a group.
	CountMembers(groupID string) (int, error)

	// CountAllMembers returns the members count for every individual group.
	CountAllMembers() ([]GroupMemberCount, error)
}

type membershipService struct {
	groupRepo      Repository           // used to verify the group exists
	membershipRepo MembershipRepository // used to manage membership records
}

// NewMembershipService constructs a MembershipService.
// groupRepo is used to confirm the group exists before mutating memberships.
func NewMembershipService(groupRepo Repository, membershipRepo MembershipRepository) MembershipService {
	return &membershipService{
		groupRepo:      groupRepo,
		membershipRepo: membershipRepo,
	}
}

func (s *membershipService) ListMembers(groupID string) ([]*GroupMembership, error) {
	// Confirm the group exists first so callers receive a proper 404.
	if _, err := s.groupRepo.FindByID(groupID); err != nil {
		return nil, err
	}
	return s.membershipRepo.ListByGroup(groupID)
}

func (s *membershipService) CountMembers(groupID string) (int, error) {
	// Confirm the group exists first so callers receive a proper 404.
	if _, err := s.groupRepo.FindByID(groupID); err != nil {
		return 0, err
	}
	return s.membershipRepo.Count(groupID)
}

func (s *membershipService) CountAllMembers() ([]GroupMemberCount, error) {
	return s.membershipRepo.CountAll()
}

func (s *membershipService) AddMember(groupID string, executingUserID string, req AddMemberRequest) (*GroupMembership, error) {
	// Rule 1 – group must exist.
	group, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return nil, err
	}

	// Authorization Check:
	// 1. executingUserID is the creator of the group
	// 2. executingUserID is an ADMIN of the group
	isCreator := group.CreatedBy == executingUserID
	isAdmin := false
	
	if !isCreator {
		role, _ := s.membershipRepo.GetRole(groupID, executingUserID)
		if role == "ADMIN" {
			isAdmin = true
		}
	}

	if !isCreator && !isAdmin {
		return nil, apperrors.New(http.StatusForbidden, "Only admin can add members")
	}

	// Rule 2 – member must not already be in this group.
	exists, err := s.membershipRepo.Exists(groupID, req.MemberID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperrors.New(http.StatusConflict, "member is already in this group")
	}

	// Rule 3 – group must not exceed the maximum allowed members.
	count, err := s.membershipRepo.Count(groupID)
	if err != nil {
		return nil, err
	}
	if count >= MaxGroupMembers {
		return nil, apperrors.New(
			http.StatusUnprocessableEntity,
			fmt.Sprintf("group is full: maximum %d members allowed", MaxGroupMembers),
		)
	}

	return s.membershipRepo.Add(groupID, req.MemberID)
}
