package scheduleinvitation

import (
	"net/http"

	"kitty-party-app/internal/apperrors"
)

// Service defines the use-case contract for schedule invitations.
type Service interface {
	// SendInvitations creates PENDING invitation rows for every group member
	// of the given schedule entry. Only group admin/creator may call this.
	SendInvitations(callerID, scheduleID string) (*SendInvitationsResult, error)

	// GetInvitationsBySchedule lists all invitations for a schedule entry.
	GetInvitationsBySchedule(scheduleID string) ([]*ScheduleInvitation, error)

	// GetMyInvitations lists all invitations for the calling member.
	GetMyInvitations(memberID string) ([]*InvitationWithSchedule, error)

	// AcceptDate marks the invitation as ACCEPTED.
	// Only the invitation's owner may call this.
	AcceptDate(callerID, invitationID string) (*ScheduleInvitation, error)

	// MarkDateProposed marks the invitation as DATE_PROPOSED.
	// Only the invitation's owner may call this.
	MarkDateProposed(callerID, invitationID string) (*ScheduleInvitation, error)
}

type service struct {
	repo          Repository
	groupProvider GroupInfoProvider
}

// NewService constructs the invitation Service.
func NewService(repo Repository, groupProvider GroupInfoProvider) Service {
	return &service{repo: repo, groupProvider: groupProvider}
}

func (s *service) SendInvitations(callerID, scheduleID string) (*SendInvitationsResult, error) {
	groupID, err := s.groupProvider.GetGroupIDForSchedule(scheduleID)
	if err != nil {
		return nil, err
	}

	isAdmin, err := s.groupProvider.IsAdminOrCreator(groupID, callerID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, apperrors.New(http.StatusForbidden, "only group admin or creator can send invitations")
	}

	memberIDs, err := s.groupProvider.ListMemberIDs(groupID)
	if err != nil {
		return nil, err
	}

	if len(memberIDs) == 0 {
		return nil, apperrors.New(http.StatusBadRequest, "no active members in this group")
	}

	invitations, err := s.repo.BulkCreate(scheduleID, memberIDs)
	if err != nil {
		return nil, err
	}

	return &SendInvitationsResult{
		ScheduleID:  scheduleID,
		Count:       len(invitations),
		Invitations: invitations,
	}, nil
}

func (s *service) GetInvitationsBySchedule(scheduleID string) ([]*ScheduleInvitation, error) {
	return s.repo.ListBySchedule(scheduleID)
}

func (s *service) GetMyInvitations(memberID string) ([]*InvitationWithSchedule, error) {
	return s.repo.ListByMember(memberID)
}

func (s *service) AcceptDate(callerID, invitationID string) (*ScheduleInvitation, error) {
	inv, err := s.repo.GetByID(invitationID)
	if err != nil {
		return nil, err
	}
	if inv.MemberID != callerID {
		return nil, apperrors.New(http.StatusForbidden, "you can only respond to your own invitation")
	}
	return s.repo.Accept(invitationID)
}

func (s *service) MarkDateProposed(callerID, invitationID string) (*ScheduleInvitation, error) {
	inv, err := s.repo.GetByID(invitationID)
	if err != nil {
		return nil, err
	}
	if inv.MemberID != callerID {
		return nil, apperrors.New(http.StatusForbidden, "you can only respond to your own invitation")
	}
	return s.repo.SetDateProposed(invitationID)
}
