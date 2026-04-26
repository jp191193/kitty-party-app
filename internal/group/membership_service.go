// Package group – membership service (business-logic) layer.
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
	// executingUserGlobalRole is "SuperAdmin" to bypass the role check.
	AddMember(groupID string, executingUserID string, executingUserGlobalRole string, req AddMemberRequest) (*GroupMembership, error)

	// CountMembers returns the total number of members in a group.
	CountMembers(groupID string) (int, error)

	// CountAllMembers returns the members count for every individual group.
	CountAllMembers() ([]GroupMemberCount, error)

	// GetMemberRole returns the calling user's role in a group.
	GetMemberRole(groupID, memberID string) (string, error)

	// UpdateMemberRole changes a member's role within a group.
	// Only a group ADMIN or SuperAdmin may perform this action.
	UpdateMemberRole(groupID, executingUserID, executingUserGlobalRole, targetMemberID, newRole string) error
}

type membershipService struct {
	groupRepo      Repository
	membershipRepo MembershipRepository
}

// NewMembershipService constructs a MembershipService.
func NewMembershipService(groupRepo Repository, membershipRepo MembershipRepository) MembershipService {
	return &membershipService{
		groupRepo:      groupRepo,
		membershipRepo: membershipRepo,
	}
}

func (s *membershipService) ListMembers(groupID string) ([]*GroupMembership, error) {
	if _, err := s.groupRepo.FindByID(groupID); err != nil {
		return nil, err
	}
	return s.membershipRepo.ListByGroup(groupID)
}

func (s *membershipService) CountMembers(groupID string) (int, error) {
	if _, err := s.groupRepo.FindByID(groupID); err != nil {
		return 0, err
	}
	return s.membershipRepo.Count(groupID)
}

func (s *membershipService) CountAllMembers() ([]GroupMemberCount, error) {
	return s.membershipRepo.CountAll()
}

func (s *membershipService) GetMemberRole(groupID, memberID string) (string, error) {
	return s.membershipRepo.GetRole(groupID, memberID)
}

func (s *membershipService) AddMember(groupID string, executingUserID string, executingUserGlobalRole string, req AddMemberRequest) (*GroupMembership, error) {
	// Rule 1 – group must exist.
	group, err := s.groupRepo.FindByID(groupID)
	if err != nil {
		return nil, err
	}

	// Authorization: SuperAdmin bypasses group-role checks.
	if executingUserGlobalRole != "SuperAdmin" {
		isCreator := group.CreatedBy == executingUserID
		isAdmin := false
		if !isCreator {
			role, _ := s.membershipRepo.GetRole(groupID, executingUserID)
			isAdmin = role == "ADMIN"
		}
		if !isCreator && !isAdmin {
			return nil, apperrors.New(http.StatusForbidden, "only group admins can add members")
		}
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

var validGroupRoles = map[string]bool{
	"ADMIN": true, "TREASURER": true, "MEMBER": true, "VIEWER": true,
}

func (s *membershipService) UpdateMemberRole(groupID, executingUserID, executingUserGlobalRole, targetMemberID, newRole string) error {
	if !validGroupRoles[newRole] {
		return apperrors.New(http.StatusBadRequest, "invalid role: must be ADMIN, TREASURER, MEMBER, or VIEWER")
	}

	if _, err := s.groupRepo.FindByID(groupID); err != nil {
		return err
	}

	// Authorization: SuperAdmin or group ADMIN only.
	if executingUserGlobalRole != "SuperAdmin" {
		role, _ := s.membershipRepo.GetRole(groupID, executingUserID)
		if role != "ADMIN" {
			return apperrors.New(http.StatusForbidden, "only group admins can change member roles")
		}
	}

	exists, err := s.membershipRepo.Exists(groupID, targetMemberID)
	if err != nil {
		return err
	}
	if !exists {
		return apperrors.ErrNotFound
	}

	return s.membershipRepo.UpdateRole(groupID, targetMemberID, newRole)
}
