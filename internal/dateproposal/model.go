package dateproposal

import "time"

// DateProposal is an alternative date suggested by a member for a schedule entry.
type DateProposal struct {
	ID           string     `json:"id"`
	ScheduleID   string     `json:"schedule_id"`
	ProposedBy   string     `json:"proposed_by"`
	ProposerName string     `json:"proposer_name,omitempty"`
	ProposedDate time.Time  `json:"proposed_date"`
	Status       string     `json:"status"` // OPEN | ACCEPTED | REJECTED | SUPERSEDED
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    *time.Time `json:"updated_at,omitempty"`
}

// ProposalVote records a single member's ACCEPT or REJECT on a proposal.
type ProposalVote struct {
	ID         string    `json:"id"`
	ProposalID string    `json:"proposal_id"`
	MemberID   string    `json:"member_id"`
	MemberName string    `json:"member_name,omitempty"`
	Vote       string    `json:"vote"` // ACCEPT | REJECT
	VotedAt    time.Time `json:"voted_at"`
}

// ProposalWithVotes bundles a proposal with its vote list and summary.
type ProposalWithVotes struct {
	DateProposal
	Votes        []*ProposalVote `json:"votes"`
	AcceptCount  int             `json:"accept_count"`
	RejectCount  int             `json:"reject_count"`
}

// CreateProposalRequest carries the data for a new date proposal.
type CreateProposalRequest struct {
	ScheduleID   string    `json:"schedule_id"   binding:"required"`
	ProposedDate time.Time `json:"proposed_date" binding:"required"`
}

// CastVoteRequest carries a member's vote.
type CastVoteRequest struct {
	Vote string `json:"vote" binding:"required,oneof=ACCEPT REJECT"`
}
