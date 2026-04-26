package dateproposal

import (
	"net/http"
	"time"

	"kitty-party-app/internal/apperrors"
)

// Service defines the use-case contract for date proposals and voting.
type Service interface {
	// CreateProposal creates a new date proposal for a schedule entry.
	// Any existing OPEN proposal for that entry is superseded.
	CreateProposal(callerID string, req CreateProposalRequest) (*DateProposal, error)

	// GetProposal returns a proposal with its vote details.
	GetProposal(id string) (*ProposalWithVotes, error)

	// ListProposalsBySchedule lists all proposals for a schedule entry.
	ListProposalsBySchedule(scheduleID string) ([]*DateProposal, error)

	// CastVote records or updates the caller's vote on a proposal.
	// When all group members have voted ACCEPT the proposal is auto-accepted
	// and the schedule entry's date is updated.
	CastVote(callerID, proposalID string, req CastVoteRequest) (*ProposalWithVotes, error)

	// AcceptProposal lets the group admin/creator manually accept a proposal,
	// updating the schedule entry's scheduled_date.
	AcceptProposal(callerID, proposalID string) (*ProposalWithVotes, error)
}

type service struct {
	repo             Repository
	scheduleProvider ScheduleInfoProvider
}

// NewService constructs the date-proposal Service.
func NewService(repo Repository, scheduleProvider ScheduleInfoProvider) Service {
	return &service{repo: repo, scheduleProvider: scheduleProvider}
}

func (s *service) CreateProposal(callerID string, req CreateProposalRequest) (*DateProposal, error) {
	if req.ProposedDate.IsZero() {
		return nil, apperrors.New(http.StatusBadRequest, "proposed_date is required")
	}
	return s.repo.Create(req.ScheduleID, callerID, req.ProposedDate)
}

func (s *service) GetProposal(id string) (*ProposalWithVotes, error) {
	proposal, err := s.repo.GetByID(id)
	if err != nil {
		return nil, err
	}
	votes, err := s.repo.GetVotesByProposal(id)
	if err != nil {
		return nil, err
	}
	return buildProposalWithVotes(proposal, votes), nil
}

func (s *service) ListProposalsBySchedule(scheduleID string) ([]*DateProposal, error) {
	return s.repo.ListBySchedule(scheduleID)
}

func (s *service) CastVote(callerID, proposalID string, req CastVoteRequest) (*ProposalWithVotes, error) {
	proposal, err := s.repo.GetByID(proposalID)
	if err != nil {
		return nil, err
	}
	if proposal.Status != "OPEN" {
		return nil, apperrors.New(http.StatusConflict,
			"voting is closed — proposal is "+proposal.Status)
	}

	if _, err := s.repo.UpsertVote(proposalID, callerID, req.Vote); err != nil {
		return nil, err
	}

	// After an ACCEPT vote, check for unanimous consensus.
	if req.Vote == "ACCEPT" {
		if err := s.checkAndAutoAccept(proposalID, proposal); err != nil {
			return nil, err
		}
	}

	return s.GetProposal(proposalID)
}

func (s *service) AcceptProposal(callerID, proposalID string) (*ProposalWithVotes, error) {
	proposal, err := s.repo.GetByID(proposalID)
	if err != nil {
		return nil, err
	}
	if proposal.Status != "OPEN" {
		return nil, apperrors.New(http.StatusConflict,
			"proposal is already "+proposal.Status)
	}

	groupID, err := s.scheduleProvider.GetGroupIDForSchedule(proposal.ScheduleID)
	if err != nil {
		return nil, err
	}
	isAdmin, err := s.scheduleProvider.IsAdminOrCreator(groupID, callerID)
	if err != nil {
		return nil, err
	}
	if !isAdmin {
		return nil, apperrors.New(http.StatusForbidden,
			"only group admin or creator can manually accept a proposal")
	}

	if err := s.acceptAndUpdateSchedule(proposalID, proposal.ScheduleID, proposal.ProposedDate); err != nil {
		return nil, err
	}

	return s.GetProposal(proposalID)
}

// checkAndAutoAccept accepts the proposal when every group member has voted ACCEPT.
func (s *service) checkAndAutoAccept(proposalID string, proposal *DateProposal) error {
	groupID, err := s.scheduleProvider.GetGroupIDForSchedule(proposal.ScheduleID)
	if err != nil {
		return err
	}
	memberCount, err := s.scheduleProvider.GetGroupMemberCount(groupID)
	if err != nil {
		return err
	}
	acceptCount, err := s.repo.CountAcceptVotes(proposalID)
	if err != nil {
		return err
	}
	if acceptCount >= memberCount {
		return s.acceptAndUpdateSchedule(proposalID, proposal.ScheduleID, proposal.ProposedDate)
	}
	return nil
}

func (s *service) acceptAndUpdateSchedule(proposalID, scheduleID string, newDate time.Time) error {
	if _, err := s.repo.Accept(proposalID); err != nil {
		return err
	}
	return s.scheduleProvider.UpdateScheduledDate(scheduleID, newDate)
}

func buildProposalWithVotes(p *DateProposal, votes []*ProposalVote) *ProposalWithVotes {
	accept, reject := 0, 0
	for _, v := range votes {
		if v.Vote == "ACCEPT" {
			accept++
		} else {
			reject++
		}
	}
	return &ProposalWithVotes{
		DateProposal: *p,
		Votes:        votes,
		AcceptCount:  accept,
		RejectCount:  reject,
	}
}
