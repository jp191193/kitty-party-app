package scheduleinvitation

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

// NewPostgresRepository returns a PostgreSQL-backed invitation repository.
func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) BulkCreate(scheduleID string, memberIDs []string) ([]*ScheduleInvitation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const q = `
		INSERT INTO schedule_invitations (schedule_id, member_id)
		VALUES ($1, $2)
		ON CONFLICT (schedule_id, member_id) DO NOTHING
		RETURNING id, schedule_id, member_id, status, responded_at, created_at, updated_at;
	`

	var out []*ScheduleInvitation
	for _, mid := range memberIDs {
		var inv ScheduleInvitation
		err := tx.QueryRow(ctx, q, scheduleID, mid).Scan(
			&inv.ID, &inv.ScheduleID, &inv.MemberID,
			&inv.Status, &inv.RespondedAt, &inv.CreatedAt, &inv.UpdatedAt,
		)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				// ON CONFLICT DO NOTHING — already exists, skip
				continue
			}
			return nil, err
		}
		out = append(out, &inv)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return out, nil
}

func (r *postgresRepository) ListBySchedule(scheduleID string) ([]*ScheduleInvitation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const q = `
		SELECT si.id, si.schedule_id, si.member_id,
		       COALESCE(u.name,'') AS member_name,
		       si.status, si.responded_at, si.created_at, si.updated_at
		  FROM schedule_invitations si
		  LEFT JOIN users u ON u.id = si.member_id
		 WHERE si.schedule_id = $1
		 ORDER BY si.created_at;
	`

	rows, err := r.pool.Query(ctx, q, scheduleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*ScheduleInvitation
	for rows.Next() {
		var inv ScheduleInvitation
		if err := rows.Scan(
			&inv.ID, &inv.ScheduleID, &inv.MemberID,
			&inv.MemberName, &inv.Status, &inv.RespondedAt,
			&inv.CreatedAt, &inv.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &inv)
	}
	return list, rows.Err()
}

func (r *postgresRepository) ListByMember(memberID string) ([]*InvitationWithSchedule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const q = `
		SELECT si.id, si.schedule_id, si.member_id,
		       COALESCE(u.name,'') AS member_name,
		       si.status, si.responded_at, si.created_at, si.updated_at,
		       ks.scheduled_date, ks.group_id, ks.cycle_number, ks.cycle_month, ks.cycle_year,
		       COALESCE(g.name,'') AS group_name
		  FROM schedule_invitations si
		  LEFT JOIN users u ON u.id = si.member_id
		  JOIN kitty_schedule ks ON ks.id = si.schedule_id
		  LEFT JOIN groups g ON g.id = ks.group_id
		 WHERE si.member_id = $1
		 ORDER BY ks.scheduled_date ASC;
	`

	rows, err := r.pool.Query(ctx, q, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*InvitationWithSchedule
	for rows.Next() {
		var inv InvitationWithSchedule
		if err := rows.Scan(
			&inv.ID, &inv.ScheduleID, &inv.MemberID,
			&inv.MemberName, &inv.Status, &inv.RespondedAt,
			&inv.CreatedAt, &inv.UpdatedAt,
			&inv.ScheduledDate, &inv.GroupID, &inv.CycleNumber,
			&inv.CycleMonth, &inv.CycleYear, &inv.GroupName,
		); err != nil {
			return nil, err
		}
		list = append(list, &inv)
	}
	return list, rows.Err()
}

func (r *postgresRepository) GetByID(id string) (*ScheduleInvitation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const q = `
		SELECT id, schedule_id, member_id, status, responded_at, created_at, updated_at
		  FROM schedule_invitations
		 WHERE id = $1;
	`

	var inv ScheduleInvitation
	err := r.pool.QueryRow(ctx, q, id).Scan(
		&inv.ID, &inv.ScheduleID, &inv.MemberID,
		&inv.Status, &inv.RespondedAt, &inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &inv, nil
}

func (r *postgresRepository) Accept(id string) (*ScheduleInvitation, error) {
	return r.updateStatus(id, "ACCEPTED")
}

func (r *postgresRepository) SetDateProposed(id string) (*ScheduleInvitation, error) {
	return r.updateStatus(id, "DATE_PROPOSED")
}

func (r *postgresRepository) updateStatus(id, status string) (*ScheduleInvitation, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const q = `
		UPDATE schedule_invitations
		   SET status = $2, responded_at = NOW(), updated_at = NOW()
		 WHERE id = $1
		RETURNING id, schedule_id, member_id, status, responded_at, created_at, updated_at;
	`

	var inv ScheduleInvitation
	err := r.pool.QueryRow(ctx, q, id, status).Scan(
		&inv.ID, &inv.ScheduleID, &inv.MemberID,
		&inv.Status, &inv.RespondedAt, &inv.CreatedAt, &inv.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &inv, nil
}


// Ensure interface is satisfied at compile time.
var _ Repository = (*postgresRepository)(nil)
