package payout

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"kitty-party-app/internal/apperrors"
)

type postgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository returns a PostgreSQL-backed payout repository.
func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) CreateSchedule(p *PayoutSchedule) (*PayoutSchedule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO payout_schedule
			(group_id, recipient_id, cycle_number, cycle_month, cycle_year,
			 scheduled_date, pool_amount, status, notes, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'SCHEDULED', $8, $9)
		RETURNING id, group_id, recipient_id, cycle_number, cycle_month, cycle_year,
		          scheduled_date, pool_amount, status, notes, created_by, created_at, updated_at;
	`

	var out PayoutSchedule
	err := r.pool.QueryRow(ctx, query,
		p.GroupID, p.RecipientID, p.CycleNumber, p.CycleMonth, p.CycleYear,
		p.ScheduledDate, p.PoolAmount, p.Notes, p.CreatedBy,
	).Scan(
		&out.ID, &out.GroupID, &out.RecipientID, &out.CycleNumber, &out.CycleMonth, &out.CycleYear,
		&out.ScheduledDate, &out.PoolAmount, &out.Status, &out.Notes, &out.CreatedBy,
		&out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		if isPgCode(err, "23505") { // unique_violation
			return nil, apperrors.ErrConflict
		}
		return nil, err
	}
	return &out, nil
}

func (r *postgresRepository) GetSchedule(id string) (*PayoutSchedule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, group_id, recipient_id, cycle_number, cycle_month, cycle_year,
		       scheduled_date, pool_amount, status, notes, created_by, created_at, updated_at
		FROM payout_schedule
		WHERE id = $1;
	`

	var p PayoutSchedule
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.GroupID, &p.RecipientID, &p.CycleNumber, &p.CycleMonth, &p.CycleYear,
		&p.ScheduledDate, &p.PoolAmount, &p.Status, &p.Notes, &p.CreatedBy,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *postgresRepository) ListByGroup(groupID string) ([]*PayoutSchedule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, group_id, recipient_id, cycle_number, cycle_month, cycle_year,
		       scheduled_date, pool_amount, status, notes, created_by, created_at, updated_at
		FROM payout_schedule
		WHERE group_id = $1
		ORDER BY cycle_number ASC;
	`

	rows, err := r.pool.Query(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*PayoutSchedule
	for rows.Next() {
		var p PayoutSchedule
		if err := rows.Scan(
			&p.ID, &p.GroupID, &p.RecipientID, &p.CycleNumber, &p.CycleMonth, &p.CycleYear,
			&p.ScheduledDate, &p.PoolAmount, &p.Status, &p.Notes, &p.CreatedBy,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	return list, rows.Err()
}

func (r *postgresRepository) ListByRecipient(recipientID string) ([]*PayoutSchedule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, group_id, recipient_id, cycle_number, cycle_month, cycle_year,
		       scheduled_date, pool_amount, status, notes, created_by, created_at, updated_at
		FROM payout_schedule
		WHERE recipient_id = $1
		ORDER BY cycle_year DESC, cycle_month DESC;
	`

	rows, err := r.pool.Query(ctx, query, recipientID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*PayoutSchedule
	for rows.Next() {
		var p PayoutSchedule
		if err := rows.Scan(
			&p.ID, &p.GroupID, &p.RecipientID, &p.CycleNumber, &p.CycleMonth, &p.CycleYear,
			&p.ScheduledDate, &p.PoolAmount, &p.Status, &p.Notes, &p.CreatedBy,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &p)
	}
	return list, rows.Err()
}

// RecordDisbursement inserts the disbursement row and flips the payout status to DISBURSED
// inside a single database transaction. The SELECT … FOR UPDATE prevents concurrent
// double-disbursements on the same payout.
func (r *postgresRepository) RecordDisbursement(d *Disbursement) (*Disbursement, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	// Lock the payout row and read its current status.
	var status string
	err = tx.QueryRow(ctx,
		`SELECT status FROM payout_schedule WHERE id = $1 FOR UPDATE`,
		d.PayoutID,
	).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}

	if status != "SCHEDULED" {
		return nil, apperrors.New(http.StatusConflict,
			"payout is not in SCHEDULED state; current status: "+status)
	}

	// Insert disbursement record.
	insertQuery := `
		INSERT INTO disbursement
			(payout_id, amount_paid, payment_mode, transaction_ref, disbursed_by, notes)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, payout_id, amount_paid, payment_date, payment_mode,
		          transaction_ref, disbursed_by, notes, created_at;
	`
	var out Disbursement
	err = tx.QueryRow(ctx, insertQuery,
		d.PayoutID, d.AmountPaid, d.PaymentMode, d.TransactionRef, d.DisbursedBy, d.Notes,
	).Scan(
		&out.ID, &out.PayoutID, &out.AmountPaid, &out.PaymentDate, &out.PaymentMode,
		&out.TransactionRef, &out.DisbursedBy, &out.Notes, &out.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	// Mark the payout as DISBURSED.
	_, err = tx.Exec(ctx,
		`UPDATE payout_schedule SET status = 'DISBURSED', updated_at = NOW() WHERE id = $1`,
		d.PayoutID,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &out, nil
}

func (r *postgresRepository) GetDisbursementByPayout(payoutID string) (*Disbursement, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, payout_id, amount_paid, payment_date, payment_mode,
		       transaction_ref, disbursed_by, notes, created_at
		FROM disbursement
		WHERE payout_id = $1;
	`

	var d Disbursement
	err := r.pool.QueryRow(ctx, query, payoutID).Scan(
		&d.ID, &d.PayoutID, &d.AmountPaid, &d.PaymentDate, &d.PaymentMode,
		&d.TransactionRef, &d.DisbursedBy, &d.Notes, &d.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil // no disbursement yet — not an error
		}
		return nil, err
	}
	return &d, nil
}

func (r *postgresRepository) CancelSchedule(id string, notes *string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var status string
	err = tx.QueryRow(ctx,
		`SELECT status FROM payout_schedule WHERE id = $1 FOR UPDATE`, id,
	).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return err
	}

	if status != "SCHEDULED" {
		return apperrors.New(http.StatusConflict,
			"cannot cancel a payout that is already "+status)
	}

	_, err = tx.Exec(ctx,
		`UPDATE payout_schedule SET status = 'CANCELLED', notes = COALESCE($2, notes), updated_at = NOW() WHERE id = $1`,
		id, notes,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// isPgCode checks whether err is a PostgreSQL error with the given SQLSTATE code.
func isPgCode(err error, code string) bool {
	type pgErr interface{ SQLState() string }
	var pe pgErr
	return errors.As(err, &pe) && pe.SQLState() == code
}
