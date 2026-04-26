package contributiontracking

import (
	"context"
	"errors"

	"kitty-party-app/internal/apperrors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the persistence contract for contribution tracking rows.
type Repository interface {
	Create(ctx context.Context, t *ContributionTracking) error
	GetByID(ctx context.Context, id string) (*ContributionTracking, error)
	GetByDueID(ctx context.Context, dueID string) (*ContributionTracking, error)
	ListByGroup(ctx context.Context, groupID string, cycleMonth, cycleYear *int) ([]*ContributionTracking, error)
	ListByMember(ctx context.Context, memberID string) ([]*ContributionTracking, error)
	Update(ctx context.Context, t *ContributionTracking) error
}

type postgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository returns a Postgres-backed tracking repository.
func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) Create(ctx context.Context, t *ContributionTracking) error {
	if t.ID == "" {
		t.ID = uuid.NewString()
	}
	query := `
		INSERT INTO contribution_tracking
			(id, group_id, member_id, due_id, cycle_month, cycle_year, cycle_number,
			 total_due, total_paid, balance, status, last_payment_date)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING created_at, updated_at;
	`
	err := r.pool.QueryRow(ctx, query,
		t.ID, t.GroupID, t.MemberID, t.DueID,
		t.CycleMonth, t.CycleYear, t.CycleNumber,
		t.TotalDue, t.TotalPaid, t.Balance, t.Status, t.LastPaymentDate,
	).Scan(&t.CreatedAt, &t.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperrors.ErrConflict
		}
		return err
	}
	return nil
}

func (r *postgresRepository) GetByID(ctx context.Context, id string) (*ContributionTracking, error) {
	return r.queryOne(ctx,
		`SELECT id, group_id, member_id, due_id, cycle_month, cycle_year, cycle_number,
		        total_due, total_paid, balance, status, last_payment_date, created_at, updated_at
		 FROM contribution_tracking WHERE id = $1`,
		id,
	)
}

func (r *postgresRepository) GetByDueID(ctx context.Context, dueID string) (*ContributionTracking, error) {
	return r.queryOne(ctx,
		`SELECT id, group_id, member_id, due_id, cycle_month, cycle_year, cycle_number,
		        total_due, total_paid, balance, status, last_payment_date, created_at, updated_at
		 FROM contribution_tracking WHERE due_id = $1`,
		dueID,
	)
}

func (r *postgresRepository) queryOne(ctx context.Context, q string, args ...any) (*ContributionTracking, error) {
	var t ContributionTracking
	err := r.pool.QueryRow(ctx, q, args...).Scan(
		&t.ID, &t.GroupID, &t.MemberID, &t.DueID,
		&t.CycleMonth, &t.CycleYear, &t.CycleNumber,
		&t.TotalDue, &t.TotalPaid, &t.Balance, &t.Status,
		&t.LastPaymentDate, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &t, nil
}

func (r *postgresRepository) ListByGroup(ctx context.Context, groupID string, cycleMonth, cycleYear *int) ([]*ContributionTracking, error) {
	query := `SELECT id, group_id, member_id, due_id, cycle_month, cycle_year, cycle_number,
	                 total_due, total_paid, balance, status, last_payment_date, created_at, updated_at
	          FROM contribution_tracking WHERE group_id = $1`
	args := []any{groupID}
	if cycleMonth != nil {
		query += " AND cycle_month = $2"
		args = append(args, *cycleMonth)
	}
	if cycleYear != nil {
		query += " AND cycle_year = $" + itoa(len(args)+1)
		args = append(args, *cycleYear)
	}
	query += " ORDER BY cycle_year DESC, cycle_month DESC, created_at ASC;"

	return r.queryMany(ctx, query, args...)
}

func (r *postgresRepository) ListByMember(ctx context.Context, memberID string) ([]*ContributionTracking, error) {
	return r.queryMany(ctx,
		`SELECT id, group_id, member_id, due_id, cycle_month, cycle_year, cycle_number,
		        total_due, total_paid, balance, status, last_payment_date, created_at, updated_at
		 FROM contribution_tracking WHERE member_id = $1
		 ORDER BY cycle_year DESC, cycle_month DESC;`,
		memberID,
	)
}

func (r *postgresRepository) queryMany(ctx context.Context, q string, args ...any) ([]*ContributionTracking, error) {
	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*ContributionTracking
	for rows.Next() {
		var t ContributionTracking
		if err := rows.Scan(
			&t.ID, &t.GroupID, &t.MemberID, &t.DueID,
			&t.CycleMonth, &t.CycleYear, &t.CycleNumber,
			&t.TotalDue, &t.TotalPaid, &t.Balance, &t.Status,
			&t.LastPaymentDate, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &t)
	}
	return list, rows.Err()
}

func (r *postgresRepository) Update(ctx context.Context, t *ContributionTracking) error {
	query := `
		UPDATE contribution_tracking
		SET total_paid        = $2,
		    balance           = $3,
		    status            = $4,
		    last_payment_date = $5,
		    updated_at        = CURRENT_TIMESTAMP
		WHERE id = $1
		RETURNING updated_at;
	`
	err := r.pool.QueryRow(ctx, query,
		t.ID, t.TotalPaid, t.Balance, t.Status, t.LastPaymentDate,
	).Scan(&t.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return err
	}
	return nil
}

// itoa avoids importing strconv solely for a one-off integer to string.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	pos := len(buf)
	for n > 0 {
		pos--
		buf[pos] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[pos:])
}
