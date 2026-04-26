package group

import (
	"context"
	"errors"
	"time"

	"kitty-party-app/internal/apperrors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type postgresMembershipRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresMembershipRepository returns a Postgres-backed membership repository.
func NewPostgresMembershipRepository(pool *pgxpool.Pool) MembershipRepository {
	return &postgresMembershipRepository{pool: pool}
}

func (r *postgresMembershipRepository) ListByGroup(groupID string) ([]*GroupMembership, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT gm.id, gm.group_id, gm.member_id, u.name, gm.status, gm.role, gm.joined_at 
		FROM group_memberships gm
		JOIN users u ON gm.member_id = u.id
		WHERE gm.group_id = $1 
		ORDER BY gm.joined_at ASC;`

	rows, err := r.pool.Query(ctx, query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*GroupMembership
	for rows.Next() {
		var m GroupMembership
		if err := rows.Scan(&m.ID, &m.GroupID, &m.MemberID, &m.MemberName, &m.Status, &m.Role, &m.JoinedAt); err != nil {
			return nil, err
		}
		list = append(list, &m)
	}
	return list, rows.Err()
}

func (r *postgresMembershipRepository) Add(groupID, memberID string) (*GroupMembership, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		INSERT INTO group_memberships (group_id, member_id, status, role) 
		VALUES ($1, $2, 'ACTIVE', 'MEMBER')
		RETURNING id, group_id, member_id, status, role, joined_at;
	`

	var m GroupMembership
	err := r.pool.QueryRow(ctx, query, groupID, memberID).
		Scan(&m.ID, &m.GroupID, &m.MemberID, &m.Status, &m.Role, &m.JoinedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique violation
			return nil, apperrors.ErrConflict
		}
		return nil, err
	}
	return &m, nil
}

func (r *postgresMembershipRepository) Exists(groupID, memberID string) (bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT 1 FROM group_memberships WHERE group_id = $1 AND member_id = $2;`

	var exists int
	err := r.pool.QueryRow(ctx, query, groupID, memberID).Scan(&exists)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *postgresMembershipRepository) Count(groupID string) (int, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT COUNT(*) FROM group_memberships WHERE group_id = $1;`
	var count int
	err := r.pool.QueryRow(ctx, query, groupID).Scan(&count)
	return count, err
}

func (r *postgresMembershipRepository) CountAll() ([]GroupMemberCount, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `select gm.group_id
                    ,g.name
					,count(*) as count
                from group_memberships gm
               inner join groups g
                       on gm.group_id = g.id
               group by gm.group_id, g.name;`

	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var counts []GroupMemberCount
	for rows.Next() {
		var c GroupMemberCount
		if err := rows.Scan(&c.IDName.ID, &c.IDName.Name, &c.Count); err != nil {
			return nil, err
		}
		counts = append(counts, c)
	}
	return counts, rows.Err()
}

func (r *postgresMembershipRepository) GetRole(groupID, memberID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `SELECT role FROM group_memberships WHERE group_id = $1 AND member_id = $2;`
	
	var role string
	err := r.pool.QueryRow(ctx, query, groupID, memberID).Scan(&role)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil 
		}
		return "", err
	}
	return role, nil
}
