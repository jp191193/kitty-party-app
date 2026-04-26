package dateproposal

import "time"

// Repository is the persistence contract for date proposals and votes.
type Repository interface {
	// Create inserts a new proposal and supersedes any existing OPEN proposal
	// for the same schedule_id in the same transaction.
	Create(scheduleID, proposedBy string, proposedDate time.Time) (*DateProposal, error)

	// GetByID returns a proposal.
	GetByID(id string) (*DateProposal, error)

	// ListBySchedule returns all proposals for a schedule entry, latest first.
	ListBySchedule(scheduleID string) ([]*DateProposal, error)

	// Accept sets the proposal status to ACCEPTED.
	Accept(id string) (*DateProposal, error)

	// Reject sets the proposal status to REJECTED.
	Reject(id string) (*DateProposal, error)

	// UpsertVote inserts or replaces a member's vote on a proposal.
	UpsertVote(proposalID, memberID, vote string) (*ProposalVote, error)

	// GetVotesByProposal returns all votes for a proposal.
	GetVotesByProposal(proposalID string) ([]*ProposalVote, error)

	// CountAcceptVotes returns the number of ACCEPT votes for a proposal.
	CountAcceptVotes(proposalID string) (int, error)
}
