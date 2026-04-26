package contribution

import (
	"context"
	"errors"
	"time"

	"kitty-party-app/internal/apperrors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Repository defines the persistence contract for contributions (dues and payments).
type Repository interface {
	// GenerateDue generates a single monthly due.
	GenerateDue(ctx context.Context, due *ContributionDue) error
	
	// RecordPayment records a payment and updates the due status transactionally.
	RecordPayment(ctx context.Context, payment *ContributionPayment) error

	// GetDue retrieves a specific due by ID.
	GetDue(ctx context.Context, dueID string) (*ContributionDue, error)

	// ListGroupDues retrieves all dues for a group, optionally filtered by month.
	ListGroupDues(ctx context.Context, groupID string, cycleMonth *int) ([]*ContributionDue, error)
}

type postgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository returns a Postgres-backed contribution repository.
func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) GenerateDue(ctx context.Context, due *ContributionDue) error {
	query := `
		INSERT INTO contribution_due (id, group_id, member_id, host_member_id, cycle_month, cycle_year, cycle_number, kitty_party_date, due_date, amount, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11);
	`
	if due.ID == "" {
		due.ID = uuid.NewString()
	}

	_, err := r.pool.Exec(ctx, query, due.ID, due.GroupID, due.MemberID, due.HostMemberID, due.CycleMonth, due.CycleYear, due.CycleNumber, due.KittyPartyDate, due.DueDate, due.Amount, due.Status)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique violation
			return apperrors.ErrConflict
		}
		return err
	}
	return nil
}

func (r *postgresRepository) GetDue(ctx context.Context, dueID string) (*ContributionDue, error) {
	query := `
		SELECT id, group_id, member_id, host_member_id, cycle_month, cycle_year, cycle_number, kitty_party_date, due_date, amount, status, created_at, updated_at
		FROM contribution_due
		WHERE id = $1;
	`
	var d ContributionDue
	err := r.pool.QueryRow(ctx, query, dueID).Scan(
		&d.ID, &d.GroupID, &d.MemberID, &d.HostMemberID, &d.CycleMonth, &d.CycleYear, &d.CycleNumber, &d.KittyPartyDate, &d.DueDate, &d.Amount, &d.Status, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &d, nil
}

func (r *postgresRepository) ListGroupDues(ctx context.Context, groupID string, cycleMonth *int) ([]*ContributionDue, error) {
	var rows pgx.Rows
	var err error

	if cycleMonth != nil {
		query := `SELECT id, group_id, member_id, host_member_id, cycle_month, cycle_year, cycle_number, kitty_party_date, due_date, amount, status, created_at, updated_at
				  FROM contribution_due WHERE group_id = $1 AND cycle_month = $2 ORDER BY created_at ASC;`
		rows, err = r.pool.Query(ctx, query, groupID, *cycleMonth)
	} else {
		query := `SELECT id, group_id, member_id, host_member_id, cycle_month, cycle_year, cycle_number, kitty_party_date, due_date, amount, status, created_at, updated_at
				  FROM contribution_due WHERE group_id = $1 ORDER BY cycle_month ASC, created_at ASC;`
		rows, err = r.pool.Query(ctx, query, groupID)
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*ContributionDue
	for rows.Next() {
		var d ContributionDue
		if err := rows.Scan(&d.ID, &d.GroupID, &d.MemberID, &d.HostMemberID, &d.CycleMonth, &d.CycleYear, &d.CycleNumber, &d.KittyPartyDate, &d.DueDate, &d.Amount, &d.Status, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, err
		}
		list = append(list, &d)
	}
	return list, rows.Err()
}

func (r *postgresRepository) RecordPayment(ctx context.Context, payment *ContributionPayment) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if payment.ID == "" {
		payment.ID = uuid.NewString()
	}

	// 1. Insert Payment
	insertCmd := `
		INSERT INTO contribution_payment (id, due_id, amount_paid, payment_date, payment_mode, transaction_ref, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`
	_, err = tx.Exec(ctx, insertCmd, payment.ID, payment.DueID, payment.AmountPaid, payment.PaymentDate, payment.PaymentMode, payment.TransactionRef, payment.Status)
	if err != nil {
		return err
	}

	// 2. Recalculate total paid & update due
	// "Only SUCCESS counts financially"
	var totalPaid float64
	sumQ := `SELECT COALESCE(SUM(amount_paid), 0) FROM contribution_payment WHERE due_id = $1 AND status = 'SUCCESS'`
	err = tx.QueryRow(ctx, sumQ, payment.DueID).Scan(&totalPaid)
	if err != nil {
		return err
	}

	var d ContributionDue
	dueQ := `SELECT amount, due_date FROM contribution_due WHERE id = $1 FOR UPDATE`
	err = tx.QueryRow(ctx, dueQ, payment.DueID).Scan(&d.Amount, &d.DueDate)
	if err != nil {
		return err
	}

	newStatus := "PENDING"
	if totalPaid >= d.Amount {
		newStatus = "PAID"
	} else if time.Now().After(d.DueDate) {
		newStatus = "OVERDUE"
	}

	updateQ := `UPDATE contribution_due SET status = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err = tx.Exec(ctx, updateQ, newStatus, payment.DueID)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}
