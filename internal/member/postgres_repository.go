package member

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"kitty-party-app/internal/apperrors"
)

// postgresRepository implements Repository against a PostgreSQL database.
type postgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository returns a new PostgreSQL-backed member repository.
func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) FindAll() ([]*Member, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, name, email, phone, created_at, updated_at FROM users ORDER BY created_at DESC;`
	
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []*Member
	for rows.Next() {
		var m Member
		if err := rows.Scan(&m.ID, &m.Name, &m.Email, &m.Phone, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, err
		}
		members = append(members, &m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}
	return members, nil
}

func (r *postgresRepository) FindByID(id string) (*Member, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT id, name, email, phone, created_at, updated_at FROM users WHERE id = $1;`

	var m Member
	err := r.pool.QueryRow(ctx, query, id).Scan(&m.ID, &m.Name, &m.Email, &m.Phone, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *postgresRepository) Create(req CreateMemberRequest) (*Member, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO users (name, email, phone, password) 
		VALUES ($1, $2, $3, $4)
		RETURNING id, name, email, phone, created_at, updated_at;
	`
	
	// Password is required by schema, setting a dummy one for now.
	dummyPassword := uuid.NewString()

	var m Member
	err := r.pool.QueryRow(ctx, query, req.Name, req.Email, req.Phone, dummyPassword).
		Scan(&m.ID, &m.Name, &m.Email, &m.Phone, &m.CreatedAt, &m.UpdatedAt)
	
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique violation
			return nil, apperrors.ErrConflict
		}
		return nil, err
	}
	return &m, nil
}

func (r *postgresRepository) Update(id string, req UpdateMemberRequest) (*Member, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		UPDATE users 
		SET 
			name = COALESCE(NULLIF($1, ''), name), 
			email = COALESCE(NULLIF($2, ''), email), 
			phone = COALESCE(NULLIF($3, ''), phone), 
			updated_at = NOW() 
		WHERE id = $4 
		RETURNING id, name, email, phone, created_at, updated_at;
	`

	var m Member
	err := r.pool.QueryRow(ctx, query, req.Name, req.Email, req.Phone, id).
		Scan(&m.ID, &m.Name, &m.Email, &m.Phone, &m.CreatedAt, &m.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &m, nil
}

func (r *postgresRepository) Delete(id string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `DELETE FROM users WHERE id = $1;`

	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
