// Package group – membership repository layer.
//
// MembershipRepository owns only the persistence of group-member links.
// All business rules (max members, duplicate checks) live in MembershipService.
package group

import (
	"sync"
	"time"

	"kitty-party-app/internal/apperrors"

	"github.com/google/uuid"
)

// MembershipRepository defines the persistence contract for group memberships.
type MembershipRepository interface {
	// ListByGroup returns all memberships for a given group.
	ListByGroup(groupID string) ([]*GroupMembership, error)

	// Add persists a new membership record and returns it.
	Add(groupID, memberID string) (*GroupMembership, error)

	// Exists reports whether the member is already in the group.
	Exists(groupID, memberID string) (bool, error)

	// Count returns the current number of members in a group.
	Count(groupID string) (int, error)

	// CountAll returns the number of members for every group.
	CountAll() ([]GroupMemberCount, error)

	// GetRole returns the role of a member in a group.
	GetRole(groupID, memberID string) (string, error)

	// UpdateRole changes a member's role within a group.
	UpdateRole(groupID, memberID, newRole string) error
}

// inMemoryMembershipRepository is a thread-safe in-memory implementation.
type inMemoryMembershipRepository struct {
	mu          sync.RWMutex
	memberships map[string]*GroupMembership // keyed by membership ID
}

// NewInMemoryMembershipRepository returns a ready-to-use MembershipRepository.
func NewInMemoryMembershipRepository() MembershipRepository {
	return &inMemoryMembershipRepository{
		memberships: make(map[string]*GroupMembership),
	}
}

func (r *inMemoryMembershipRepository) ListByGroup(groupID string) ([]*GroupMembership, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var list []*GroupMembership
	for _, m := range r.memberships {
		if m.GroupID == groupID {
			list = append(list, m)
		}
	}
	return list, nil
}

func (r *inMemoryMembershipRepository) Add(groupID, memberID string) (*GroupMembership, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	m := &GroupMembership{
		ID:       uuid.NewString(),
		GroupID:  groupID,
		MemberID: memberID,
		Status:   "ACTIVE",
		JoinedAt: time.Now().UTC(),
	}
	r.memberships[m.ID] = m
	return m, nil
}

func (r *inMemoryMembershipRepository) Exists(groupID, memberID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, m := range r.memberships {
		if m.GroupID == groupID && m.MemberID == memberID {
			return true, nil
		}
	}
	return false, nil
}

func (r *inMemoryMembershipRepository) Count(groupID string) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	count := 0
	for _, m := range r.memberships {
		if m.GroupID == groupID {
			count++
		}
	}
	return count, nil
}

func (r *inMemoryMembershipRepository) CountAll() ([]GroupMemberCount, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	countsMap := make(map[IDName]int)
	for _, m := range r.memberships {
		countsMap[IDName{ID: m.GroupID, Name: m.MemberName}]++
	}

	var counts []GroupMemberCount
	for k, v := range countsMap {
		counts = append(counts, GroupMemberCount{IDName: IDName{ID: k.ID, Name: k.Name}, Count: v})
	}
	return counts, nil
}

func (r *inMemoryMembershipRepository) GetRole(groupID, memberID string) (string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, m := range r.memberships {
		if m.GroupID == groupID && m.MemberID == memberID {
			return m.Role, nil
		}
	}
	return "", nil
}

func (r *inMemoryMembershipRepository) UpdateRole(groupID, memberID, newRole string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, m := range r.memberships {
		if m.GroupID == groupID && m.MemberID == memberID {
			m.Role = newRole
			return nil
		}
	}
	return apperrors.ErrNotFound
}

// ErrGroupNotFound is returned when the target group does not exist.
// (re-exported from apperrors for convenience inside this package)
var errGroupNotFound = apperrors.ErrNotFound
