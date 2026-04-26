// Package member – repository layer (in-memory implementation).
package member

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"kitty-party-app/internal/apperrors"
)

// Repository defines the data-access contract for Member persistence.
type Repository interface {
	FindAll() ([]*Member, error)
	FindByID(id string) (*Member, error)
	FindByEmail(email string) (*Credentials, error)
	Create(req CreateMemberRequest) (*Member, error)
	Update(id string, req UpdateMemberRequest) (*Member, error)
	Delete(id string) error
}

// inMemoryRepository is a thread-safe, in-memory implementation of Repository.
// Replace this with a real database adapter (PostgreSQL, MongoDB, etc.) without
// changing any other layer.
type inMemoryRepository struct {
	mu      sync.RWMutex
	members map[string]*Member
}

// NewInMemoryRepository constructs an inMemoryRepository with seed data.
func NewInMemoryRepository() Repository {
	r := &inMemoryRepository{
		members: make(map[string]*Member),
	}
	// Seed a couple of members for demonstration purposes.
	seed := []CreateMemberRequest{
		{Name: "Anjali Sharma", Email: "anjali@example.com", Phone: "9876543210"},
		{Name: "Priya Mehta", Email: "priya@example.com", Phone: "9123456789"},
	}
	for _, s := range seed {
		_, _ = r.Create(s) //nolint:errcheck – seed is safe
	}
	return r
}

func (r *inMemoryRepository) FindAll() ([]*Member, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*Member, 0, len(r.members))
	for _, m := range r.members {
		list = append(list, m)
	}
	return list, nil
}

func (r *inMemoryRepository) FindByID(id string) (*Member, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	m, ok := r.members[id]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	return m, nil
}

func (r *inMemoryRepository) FindByEmail(email string) (*Credentials, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, m := range r.members {
		if m.Email == email {
			return &Credentials{Member: m, PasswordHash: ""}, nil
		}
	}
	return nil, apperrors.ErrNotFound
}

func (r *inMemoryRepository) Create(req CreateMemberRequest) (*Member, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check for duplicate email.
	for _, m := range r.members {
		if m.Email == req.Email {
			return nil, apperrors.ErrConflict
		}
	}

	now := time.Now().UTC()
	m := &Member{
		ID:        uuid.NewString(),
		Name:      req.Name,
		Email:     req.Email,
		Phone:     req.Phone,
		CreatedAt: now,
		UpdatedAt: now,
	}
	r.members[m.ID] = m
	return m, nil
}

func (r *inMemoryRepository) Update(id string, req UpdateMemberRequest) (*Member, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, ok := r.members[id]
	if !ok {
		return nil, apperrors.ErrNotFound
	}

	if req.Name != "" {
		m.Name = req.Name
	}
	if req.Email != "" {
		m.Email = req.Email
	}
	if req.Phone != "" {
		m.Phone = req.Phone
	}
	m.UpdatedAt = time.Now().UTC()

	return m, nil
}

func (r *inMemoryRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.members[id]; !ok {
		return apperrors.ErrNotFound
	}
	delete(r.members, id)
	return nil
}
