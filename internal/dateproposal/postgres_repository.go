package dateproposal

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"kitty-party-app/internal/apperrors"
)

type postgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository returns a PostgreSQL-backed date proposal repository.
func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) Create(scheduleID, proposedBy string, proposedDate time.Time) (*DateProposal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Supersede any currently OPEN proposal for this schedule entry.
	_, err = tx.Exec(ctx, `
		UPDATE date_proposals
		   SET status = 'SUPERSEDED', updated_at = NOW()
		 WHERE schedule_id = $1 AND status = 'OPEN'
	`, scheduleID)
	if err != nil {
		return nil, err
	}

	const insertQ = `
		INSERT INTO date_proposals (schedule_id, proposed_by, proposed_date)
		VALUES ($1, $2, $3)
		RETURNING id, schedule_id, proposed_by, proposed_date, status, created_at, updated_at;
	`

	var p DateProposal
	err = tx.QueryRow(ctx, insertQ, scheduleID, proposedBy, proposedDate).Scan(
		&p.ID, &p.ScheduleID, &p.ProposedBy, &p.ProposedDate,
		&p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *postgresRepository) GetByID(id string) (*DateProposal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const q = `
		SELECT dp.id, dp.schedule_id, dp.proposed_by,
		       COALESCE(u.name,'') AS proposer_name,
		       dp.proposed_date, dp.status, dp.created_at, dp.updated_at
		  FROM date_proposals dp
		  LEFT JOIN users u ON u.id = dp.proposed_by
		 WHERE dp.id = $1;
	`

	var p DateProposal
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&p.ID, &p.ScheduleID, &p.ProposedBy, &p.ProposerName,
		&p.ProposedDate, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *postgresRepository) ListBySchedule(scheduleID string) ([]*DateProposal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const q = `
		SELECT dp.id, dp.schedule_id, dp.proposed_by,
		       COALESCE(u.name,'') AS proposer_name,
		       dp.proposed_date, dp.status, dp.created_at, dp.updated_at
		  FROM date_proposals dp
		  LEFT JOIN users u ON u.id = dp.proposed_by
		 WHERE dp.schedule_id = $1
		 ORDER BY dp.created_at DESC;
	`

	rows, err := r.pool.Query(ctx, q, scheduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*DateProposal
	for rows.Next() {
		var p DateProposal
		if err := rows.Scan(
			&p.ID, &p.ScheduleID, &p.ProposedBy, &p.ProposerName,
			&p.ProposedDate, &p.Status, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	return list, rows.Err()
}

func (r *postgresRepository) Accept(id string) (*DateProposal, error) {
	return r.setStatus(id, "ACCEPTED")
}

func (r *postgresRepository) Reject(id string) (*DateProposal, error) {
	return r.setStatus(id, "REJECTED")
}

func (r *postgresRepository) setStatus(id, status string) (*DateProposal, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const q = `
		UPDATE date_proposals
		   SET status = $2, updated_at = NOW()
		 WHERE id = $1
		RETURNING id, schedule_id, proposed_by, proposed_date, status, created_at, updated_at;
	`

	var p DateProposal
	err := r.pool.QueryRow(ctx, q, id, status).Scan(
		&p.ID, &p.ScheduleID, &p.ProposedBy, &p.ProposedDate,
		&p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *postgresRepository) UpsertVote(proposalID, memberID, vote string) (*ProposalVote, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const q = `
		INSERT INTO date_proposal_votes (proposal_id, member_id, vote)
		VALUES ($1, $2, $3)
		ON CONFLICT (proposal_id, member_id)
		DO UPDATE SET vote = EXCLUDED.vote, voted_at = NOW()
		RETURNING id, proposal_id, member_id, vote, voted_at;
	`

	var v ProposalVote
	err := r.pool.QueryRow(ctx, q, proposalID, memberID, vote).Scan(
		&v.ID, &v.ProposalID, &v.MemberID, &v.Vote, &v.VotedAt,
	)
	if err != nil {
		return nil, err
	}
	return &v, nil
}

func (r *postgresRepository) GetVotesByProposal(proposalID string) ([]*ProposalVote, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const q = `
		SELECT v.id, v.proposal_id, v.member_id,
		       COALESCE(u.name,'') AS member_name,
		       v.vote, v.voted_at
		  FROM date_proposal_votes v
		  LEFT JOIN users u ON u.id = v.member_id
		 WHERE v.proposal_id = $1
		 ORDER BY v.voted_at;
	`

	rows, err := r.pool.Query(ctx, q, proposalID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*ProposalVote
	for rows.Next() {
		var v ProposalVote
		if err := rows.Scan(
			&v.ID, &v.ProposalID, &v.MemberID,
			&v.MemberName, &v.Vote, &v.VotedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &v)
	}
	return list, rows.Err()
}

func (r *postgresRepository) CountAcceptVotes(proposalID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var count int
	err := r.pool.QueryRow(ctx,
		`SELECT COUNT(*) FROM date_proposal_votes WHERE proposal_id = $1 AND vote = 'ACCEPT'`,
		proposalID,
	).Scan(&count)
	return count, err
}

// Compile-time interface check.
var _ Repository = (*postgresRepository)(nil)

