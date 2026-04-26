package kittycycle

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

// NewPostgresRepository returns a PostgreSQL-backed kitty cycle repository.
func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) CreateCycleWithSchedule(cycle *KittyCycle, entries []*KittySchedule) (*KittyCycleWithSchedule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	insertCycle := `
		INSERT INTO kitty_cycle
			(group_id, name, total_cycles, monthly_amount, pool_amount,
			 start_date, end_date, status, notes, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, 'PENDING', $8, $9)
		RETURNING id, group_id, name, total_cycles, monthly_amount, pool_amount,
		          start_date, end_date, status, notes, created_by, created_at, updated_at;
	`

	var out KittyCycle
	err = tx.QueryRow(ctx, insertCycle,
		cycle.GroupID, cycle.Name, cycle.TotalCycles, cycle.MonthlyAmount, cycle.PoolAmount,
		cycle.StartDate, cycle.EndDate, cycle.Notes, cycle.CreatedBy,
	).Scan(
		&out.ID, &out.GroupID, &out.Name, &out.TotalCycles, &out.MonthlyAmount, &out.PoolAmount,
		&out.StartDate, &out.EndDate, &out.Status, &out.Notes, &out.CreatedBy,
		&out.CreatedAt, &out.UpdatedAt,
	)
	if err != nil {
		if isPgCode(err, "23505") {
			return nil, apperrors.ErrConflict
		}
		return nil, err
	}

	insertEntry := `
		INSERT INTO kitty_schedule
			(cycle_id, group_id, cycle_number, cycle_month, cycle_year,
			 host_member_id, scheduled_date, due_date, pool_amount, status, venue, notes)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 'SCHEDULED', $10, $11)
		RETURNING id, cycle_id, group_id, cycle_number, cycle_month, cycle_year,
		          host_member_id, scheduled_date, due_date, pool_amount, status, venue, notes,
		          created_at, updated_at;
	`

	persistedEntries := make([]*KittySchedule, 0, len(entries))
	for _, e := range entries {
		ks, err := scanScheduleRow(tx.QueryRow(ctx, insertEntry,
			out.ID, out.GroupID, e.CycleNumber, e.CycleMonth, e.CycleYear,
			e.HostMemberID, e.ScheduledDate, e.DueDate, e.PoolAmount, e.Venue, e.Notes,
		))
		if err != nil {
			if isPgCode(err, "23505") {
				return nil, apperrors.New(http.StatusConflict,
					"duplicate cycle_number or host_member_id within this cycle")
			}
			return nil, err
		}
		persistedEntries = append(persistedEntries, ks)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &KittyCycleWithSchedule{KittyCycle: out, Schedule: persistedEntries}, nil
}

func (r *postgresRepository) GetCycle(id string) (*KittyCycleWithSchedule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, group_id, name, total_cycles, monthly_amount, pool_amount,
		       start_date, end_date, status, notes, created_by, created_at, updated_at
		FROM kitty_cycle
		WHERE id = $1;
	`

	var c KittyCycle
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.GroupID, &c.Name, &c.TotalCycles, &c.MonthlyAmount, &c.PoolAmount,
		&c.StartDate, &c.EndDate, &c.Status, &c.Notes, &c.CreatedBy,
		&c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}

	entries, err := r.ListSchedule(id)
	if err != nil {
		return nil, err
	}
	return &KittyCycleWithSchedule{KittyCycle: c, Schedule: entries}, nil
}

func (r *postgresRepository) ListByGroup(groupID string) ([]*KittyCycle, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, group_id, name, total_cycles, monthly_amount, pool_amount,
		       start_date, end_date, status, notes, created_by, created_at, updated_at
		FROM kitty_cycle
		WHERE group_id = $1
		ORDER BY start_date DESC;
	`

	rows, err := r.pool.Query(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*KittyCycle
	for rows.Next() {
		var c KittyCycle
		if err := rows.Scan(
			&c.ID, &c.GroupID, &c.Name, &c.TotalCycles, &c.MonthlyAmount, &c.PoolAmount,
			&c.StartDate, &c.EndDate, &c.Status, &c.Notes, &c.CreatedBy,
			&c.CreatedAt, &c.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &c)
	}
	return list, rows.Err()
}

func (r *postgresRepository) ListSchedule(cycleID string) ([]*KittySchedule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, cycle_id, group_id, cycle_number, cycle_month, cycle_year,
		       host_member_id, scheduled_date, due_date, pool_amount, status, venue, notes,
		       created_at, updated_at
		FROM kitty_schedule
		WHERE cycle_id = $1
		ORDER BY cycle_number ASC;
	`

	rows, err := r.pool.Query(ctx, query, cycleID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*KittySchedule
	for rows.Next() {
		var ks KittySchedule
		if err := rows.Scan(
			&ks.ID, &ks.CycleID, &ks.GroupID, &ks.CycleNumber, &ks.CycleMonth, &ks.CycleYear,
			&ks.HostMemberID, &ks.ScheduledDate, &ks.DueDate, &ks.PoolAmount, &ks.Status, &ks.Venue, &ks.Notes,
			&ks.CreatedAt, &ks.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, &ks)
	}
	return list, rows.Err()
}

func (r *postgresRepository) CancelCycle(id string, notes *string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var status string
	err = tx.QueryRow(ctx,
		`SELECT status FROM kitty_cycle WHERE id = $1 FOR UPDATE`, id,
	).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return apperrors.ErrNotFound
		}
		return err
	}

	if status == "COMPLETED" || status == "CANCELLED" {
		return apperrors.New(http.StatusConflict,
			"cannot cancel a cycle that is already "+status)
	}

	_, err = tx.Exec(ctx,
		`UPDATE kitty_cycle
		    SET status = 'CANCELLED',
		        notes = COALESCE($2, notes),
		        updated_at = NOW()
		  WHERE id = $1`,
		id, notes,
	)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx,
		`UPDATE kitty_schedule
		    SET status = 'CANCELLED', updated_at = NOW()
		  WHERE cycle_id = $1 AND status = 'SCHEDULED'`,
		id,
	)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *postgresRepository) StartCycle(id string) (*KittyCycleWithSchedule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var status string
	err = tx.QueryRow(ctx,
		`SELECT status FROM kitty_cycle WHERE id = $1 FOR UPDATE`, id,
	).Scan(&status)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}

	if status != "PENDING" {
		return nil, apperrors.New(http.StatusConflict,
			"cannot start a cycle that is already "+status)
	}

	var updated KittyCycle
	err = tx.QueryRow(ctx,
		`UPDATE kitty_cycle
		    SET status = 'ACTIVE', updated_at = NOW()
		  WHERE id = $1
		RETURNING id, group_id, name, total_cycles, monthly_amount, pool_amount,
		          start_date, end_date, status, notes, created_by, created_at, updated_at`,
		id,
	).Scan(
		&updated.ID, &updated.GroupID, &updated.Name, &updated.TotalCycles,
		&updated.MonthlyAmount, &updated.PoolAmount, &updated.StartDate, &updated.EndDate,
		&updated.Status, &updated.Notes, &updated.CreatedBy, &updated.CreatedAt, &updated.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx,
		`UPDATE kitty_schedule
		    SET status = 'IN_PROGRESS', updated_at = NOW()
		  WHERE cycle_id = $1
		    AND cycle_number = (SELECT MIN(cycle_number) FROM kitty_schedule WHERE cycle_id = $1)`,
		id,
	)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	entries, err := r.ListSchedule(id)
	if err != nil {
		return nil, err
	}
	return &KittyCycleWithSchedule{KittyCycle: updated, Schedule: entries}, nil
}

// isPgCode checks whether err is a PostgreSQL error with the given SQLSTATE code.
func isPgCode(err error, code string) bool {
	type pgErr interface{ SQLState() string }
	var pe pgErr
	return errors.As(err, &pe) && pe.SQLState() == code
}

// scanScheduleRow scans the standard kitty_schedule column tuple from any pgx.Row.
func scanScheduleRow(row pgx.Row) (*KittySchedule, error) {
	var ks KittySchedule
	if err := row.Scan(
		&ks.ID, &ks.CycleID, &ks.GroupID, &ks.CycleNumber, &ks.CycleMonth, &ks.CycleYear,
		&ks.HostMemberID, &ks.ScheduledDate, &ks.DueDate, &ks.PoolAmount, &ks.Status, &ks.Venue, &ks.Notes,
		&ks.CreatedAt, &ks.UpdatedAt,
	); err != nil {
		return nil, err
	}
	return &ks, nil
}

func (r *postgresRepository) GetSchedule(scheduleID string) (*KittySchedule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, cycle_id, group_id, cycle_number, cycle_month, cycle_year,
		       host_member_id, scheduled_date, due_date, pool_amount, status, venue, notes,
		       created_at, updated_at
		FROM kitty_schedule
		WHERE id = $1;
	`
	ks, err := scanScheduleRow(r.pool.QueryRow(ctx, query, scheduleID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return ks, nil
}

func (r *postgresRepository) AssignHost(scheduleID, hostMemberID string) (*KittySchedule, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	query := `
		UPDATE kitty_schedule
		   SET host_member_id = $2, updated_at = NOW()
		 WHERE id = $1
		   AND host_member_id IS NULL
		RETURNING id, cycle_id, group_id, cycle_number, cycle_month, cycle_year,
		          host_member_id, scheduled_date, due_date, pool_amount, status, venue, notes,
		          created_at, updated_at;
	`
	ks, err := scanScheduleRow(r.pool.QueryRow(ctx, query, scheduleID, hostMemberID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Either the schedule doesn't exist or it already has a host.
			return nil, apperrors.New(http.StatusConflict, "host already assigned for this schedule entry")
		}
		if isPgCode(err, "23505") {
			return nil, apperrors.New(http.StatusConflict, "this member is already a host in this cycle")
		}
		return nil, err
	}
	return ks, nil
}

func (r *postgresRepository) GetGroupIDForSchedule(scheduleID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var groupID string
	err := r.pool.QueryRow(ctx,
		`SELECT group_id FROM kitty_schedule WHERE id = $1`, scheduleID,
	).Scan(&groupID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", apperrors.ErrNotFound
		}
		return "", err
	}
	return groupID, nil
}

func (r *postgresRepository) UpdateScheduledDate(scheduleID string, newDate time.Time) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tag, err := r.pool.Exec(ctx,
		`UPDATE kitty_schedule SET scheduled_date = $2, updated_at = NOW() WHERE id = $1`,
		scheduleID, newDate,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}

func (r *postgresRepository) ListAssignedHosts(cycleID string) (map[string]bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := r.pool.Query(ctx,
		`SELECT host_member_id FROM kitty_schedule
		  WHERE cycle_id = $1 AND host_member_id IS NOT NULL`,
		cycleID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make(map[string]bool)
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out[id] = true
	}
	return out, rows.Err()
}
