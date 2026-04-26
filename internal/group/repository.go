// Package group – repository layer.
//
// Repository is the only layer that knows about storage details.
// Swap inMemoryRepository for a real DB adapter (Postgres, Mongo, etc.)
// without touching the service or handler layers.
package group

import (
	"sync"
	"time"

	"github.com/google/uuid"
	"kitty-party-app/internal/apperrors"
)

// Repository defines the persistence contract for the Group domain.
type Repository interface {
	FindAll() ([]*Group, error)
	FindByID(id string) (*Group, error)
	FindByCreator(createdBy string) ([]*Group, error)
	Create(req CreateGroupRequest) (*Group, error)
	Update(id string, req UpdateGroupRequest) (*Group, error)
	Delete(id string) error
}

// inMemoryRepository is a thread-safe, in-memory implementation of Repository.
type inMemoryRepository struct {
	mu     sync.RWMutex
	groups map[string]*Group
}

// NewInMemoryRepository returns a ready-to-use in-memory Repository.
func NewInMemoryRepository() Repository {
	return &inMemoryRepository{
		groups: make(map[string]*Group),
	}
}

func (r *inMemoryRepository) FindAll() ([]*Group, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	list := make([]*Group, 0, len(r.groups))
	for _, g := range r.groups {
		list = append(list, g)
	}
	return list, nil
}

func (r *inMemoryRepository) FindByID(id string) (*Group, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	g, ok := r.groups[id]
	if !ok {
		return nil, apperrors.ErrNotFound
	}
	return g, nil
}

func (r *inMemoryRepository) FindByCreator(createdBy string) ([]*Group, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []*Group
	for _, g := range r.groups {
		if g.CreatedBy == createdBy {
			list = append(list, g)
		}
	}
	return list, nil
}

func (r *inMemoryRepository) Create(req CreateGroupRequest) (*Group, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().UTC()
	g := &Group{
		ID:            uuid.NewString(),
		Name:          req.Name,
		MonthlyAmount: req.MonthlyAmount,
		Duration:      req.Duration,
		StartDate:     req.StartDate.UTC(),
		CreatedBy:     req.CreatedBy,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	r.groups[g.ID] = g
	return g, nil
}

func (r *inMemoryRepository) Update(id string, req UpdateGroupRequest) (*Group, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	g, ok := r.groups[id]
	if !ok {
		return nil, apperrors.ErrNotFound
	}

	if req.Name != "" {
		g.Name = req.Name
	}
	if req.MonthlyAmount > 0 {
		g.MonthlyAmount = req.MonthlyAmount
	}
	if req.Duration > 0 {
		g.Duration = req.Duration
	}
	if !req.StartDate.IsZero() {
		g.StartDate = req.StartDate.UTC()
	}
	g.UpdatedAt = time.Now().UTC()

	return g, nil
}

func (r *inMemoryRepository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, ok := r.groups[id]; !ok {
		return apperrors.ErrNotFound
	}
	delete(r.groups, id)
	return nil
}
