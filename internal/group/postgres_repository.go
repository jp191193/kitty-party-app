package group

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

// NewPostgresRepository returns a new PostgreSQL-backed group repository.
func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) FindAll() ([]*Group, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, name, monthly_amount, duration, start_date, created_by, created_at, updated_at FROM groups ORDER BY created_at DESC;`
	
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*Group
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.ID, &g.Name, &g.MonthlyAmount, &g.Duration, &g.StartDate, &g.CreatedBy, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, &g)
	}
	return groups, rows.Err()
}

func (r *postgresRepository) FindByID(id string) (*Group, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, name, monthly_amount, duration, start_date, created_by, created_at, updated_at FROM groups WHERE id = $1;`

	var g Group
	err := r.pool.QueryRow(ctx, query, id).
		Scan(&g.ID, &g.Name, &g.MonthlyAmount, &g.Duration, &g.StartDate, &g.CreatedBy, &g.CreatedAt, &g.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &g, nil
}

func (r *postgresRepository) FindByCreator(createdBy string) ([]*Group, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, name, monthly_amount, duration, start_date, created_by, created_at, updated_at FROM groups WHERE created_by = $1 ORDER BY created_at DESC;`
	
	rows, err := r.pool.Query(ctx, query, createdBy)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []*Group
	for rows.Next() {
		var g Group
		if err := rows.Scan(&g.ID, &g.Name, &g.MonthlyAmount, &g.Duration, &g.StartDate, &g.CreatedBy, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, err
		}
		groups = append(groups, &g)
	}
	return groups, rows.Err()
}

func (r *postgresRepository) Create(req CreateGroupRequest) (*Group, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO groups (name, monthly_amount, duration, start_date, created_by) 
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, monthly_amount, duration, start_date, created_by, created_at, updated_at;
	`
	
	var g Group
	err = tx.QueryRow(ctx, query, req.Name, req.MonthlyAmount, req.Duration, req.StartDate, req.CreatedBy).
		Scan(&g.ID, &g.Name, &g.MonthlyAmount, &g.Duration, &g.StartDate, &g.CreatedBy, &g.CreatedAt, &g.UpdatedAt)
	
	if err != nil {
		return nil, err
	}

	memberQuery := `
		INSERT INTO group_memberships (group_id, member_id, status, role)
		VALUES ($1, $2, 'ACTIVE', 'ADMIN');
	`
	_, err = tx.Exec(ctx, memberQuery, g.ID, g.CreatedBy)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &g, nil
}

func (r *postgresRepository) Update(id string, req UpdateGroupRequest) (*Group, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		UPDATE groups 
		SET 
			name = COALESCE(NULLIF($1, ''), name), 
			monthly_amount = CASE WHEN $2 > 0 THEN $2 ELSE monthly_amount END,
			duration = CASE WHEN $3 > 0 THEN $3 ELSE duration END,
			start_date = CASE WHEN $4 != '0001-01-01 00:00:00+00' THEN $4 ELSE start_date END,
			updated_at = NOW() 
		WHERE id = $5 
		RETURNING id, name, monthly_amount, duration, start_date, created_by, created_at, updated_at;
	`

	var g Group
	err := r.pool.QueryRow(ctx, query, req.Name, req.MonthlyAmount, req.Duration, req.StartDate, id).
		Scan(&g.ID, &g.Name, &g.MonthlyAmount, &g.Duration, &g.StartDate, &g.CreatedBy, &g.CreatedAt, &g.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &g, nil
}

func (r *postgresRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `DELETE FROM groups WHERE id = $1;`

	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
