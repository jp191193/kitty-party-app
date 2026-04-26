package profile

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"kitty-party-app/internal/apperrors"
)

type postgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository returns a PostgreSQL-backed profile repository.
func NewPostgresRepository(pool *pgxpool.Pool) Repository {
	return &postgresRepository{pool: pool}
}

func (r *postgresRepository) FindByMemberID(memberID string) (*MemberProfile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	query := `
		SELECT id, member_id, date_of_birth, profile_picture_url,
		       address, city, state, pincode, occupation, about,
		       created_at, updated_at
		FROM member_profiles
		WHERE member_id = $1;
	`

	var p MemberProfile
	err := r.pool.QueryRow(ctx, query, memberID).Scan(
		&p.ID, &p.MemberID, &p.DateOfBirth, &p.ProfilePictureURL,
		&p.Address, &p.City, &p.State, &p.Pincode, &p.Occupation, &p.About,
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

func (r *postgresRepository) Upsert(memberID string, req UpsertProfileRequest) (*MemberProfile, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Parse optional date_of_birth string ("YYYY-MM-DD") into *time.Time.
	var dob *time.Time
	if req.DateOfBirth != nil && *req.DateOfBirth != "" {
		parsed, err := time.Parse("2006-01-02", *req.DateOfBirth)
		if err != nil {
			return nil, apperrors.New(400, fmt.Sprintf("invalid date_of_birth format, expected YYYY-MM-DD: %s", err.Error()))
		}
		dob = &parsed
	}

	query := `
		INSERT INTO member_profiles
			(member_id, date_of_birth, profile_picture_url, address, city, state, pincode, occupation, about)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (member_id) DO UPDATE SET
			date_of_birth       = COALESCE($2,  member_profiles.date_of_birth),
			profile_picture_url = COALESCE($3,  member_profiles.profile_picture_url),
			address             = COALESCE($4,  member_profiles.address),
			city                = COALESCE($5,  member_profiles.city),
			state               = COALESCE($6,  member_profiles.state),
			pincode             = COALESCE($7,  member_profiles.pincode),
			occupation          = COALESCE($8,  member_profiles.occupation),
			about               = COALESCE($9,  member_profiles.about),
			updated_at          = NOW()
		RETURNING id, member_id, date_of_birth, profile_picture_url,
		          address, city, state, pincode, occupation, about,
		          created_at, updated_at;
	`

	var p MemberProfile
	err := r.pool.QueryRow(ctx, query,
		memberID, dob, req.ProfilePictureURL, req.Address, req.City,
		req.State, req.Pincode, req.Occupation, req.About,
	).Scan(
		&p.ID, &p.MemberID, &p.DateOfBirth, &p.ProfilePictureURL,
		&p.Address, &p.City, &p.State, &p.Pincode, &p.Occupation, &p.About,
		&p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		// Foreign key violation – member does not exist.
		if isFK(err) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *postgresRepository) GetSummary(memberID string) (*ProfileSummary, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// ── 1. Member + profile ───────────────────────────────────────────────────
	memberQuery := `
		SELECT u.id, u.name, u.email, u.phone, u.created_at, u.updated_at,
		       mp.id, mp.member_id, mp.date_of_birth, mp.profile_picture_url,
		       mp.address, mp.city, mp.state, mp.pincode, mp.occupation, mp.about,
		       mp.created_at, mp.updated_at
		FROM users u
		LEFT JOIN member_profiles mp ON mp.member_id = u.id
		WHERE u.id = $1;
	`

	var s ProfileSummary
	var profileID, profileMemberID *string
	var profileCreatedAt, profileUpdatedAt *time.Time
	var profile MemberProfile

	err := r.pool.QueryRow(ctx, memberQuery, memberID).Scan(
		&s.ID, &s.Name, &s.Email, &s.Phone, &s.CreatedAt, &s.UpdatedAt,
		&profileID, &profileMemberID, &profile.DateOfBirth, &profile.ProfilePictureURL,
		&profile.Address, &profile.City, &profile.State, &profile.Pincode,
		&profile.Occupation, &profile.About,
		&profileCreatedAt, &profileUpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}

	if profileID != nil {
		profile.ID = *profileID
		profile.MemberID = *profileMemberID
		profile.CreatedAt = *profileCreatedAt
		profile.UpdatedAt = *profileUpdatedAt
		s.Profile = &profile
	}

	// ── 2. Group memberships ─────────────────────────────────────────────────
	groupQuery := `
		SELECT g.id, g.name, gm.role, gm.status, gm.joined_at
		FROM group_memberships gm
		JOIN groups g ON g.id = gm.group_id
		WHERE gm.member_id = $1
		ORDER BY gm.joined_at DESC;
	`

	rows, err := r.pool.Query(ctx, groupQuery, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	s.Groups = []GroupEntry{}
	for rows.Next() {
		var ge GroupEntry
		if err := rows.Scan(&ge.GroupID, &ge.GroupName, &ge.Role, &ge.Status, &ge.JoinedAt); err != nil {
			return nil, err
		}
		s.Groups = append(s.Groups, ge)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	s.GroupCount = len(s.Groups)

	// ── 3. Contribution statistics ───────────────────────────────────────────
	statsQuery := `
		SELECT
			COALESCE(SUM(CASE WHEN cp.status = 'SUCCESS' THEN cp.amount_paid ELSE 0 END), 0) AS total_paid,
			COUNT(CASE WHEN cd.status IN ('PENDING', 'OVERDUE') THEN 1 END)                  AS pending_dues
		FROM contribution_due cd
		LEFT JOIN contribution_payment cp ON cp.due_id = cd.id
		WHERE cd.member_id = $1;
	`

	err = r.pool.QueryRow(ctx, statsQuery, memberID).Scan(&s.TotalPaid, &s.PendingDues)
	if err != nil {
		return nil, err
	}

	return &s, nil
}

// isFK reports whether err is a PostgreSQL foreign-key violation (code 23503).
func isFK(err error) bool {
	type pgErr interface{ SQLState() string }
	var pe pgErr
	if errors.As(err, &pe) {
		return pe.SQLState() == "23503"
	}
	return false
}
