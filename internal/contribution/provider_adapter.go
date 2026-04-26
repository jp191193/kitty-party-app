package contribution

import (
	"context"

	"kitty-party-app/internal/group"
)

// GroupMemberProviderAdapter adapts the group.MembershipRepository 
// to the contribution.GroupMemberProvider interface.
type GroupMemberProviderAdapter struct {
	repo group.MembershipRepository
}

// NewGroupMemberProviderAdapter creates a new adapter.
func NewGroupMemberProviderAdapter(repo group.MembershipRepository) *GroupMemberProviderAdapter {
	return &GroupMemberProviderAdapter{repo: repo}
}

// GetMemberIDs returns all member IDs for a given group.
func (a *GroupMemberProviderAdapter) GetMemberIDs(ctx context.Context, groupID string) ([]string, error) {
	// ListByGroup doesn't accept context currently based on its interface, so we just call it directly.
	memberships, err := a.repo.ListByGroup(groupID)
	if err != nil {
		return nil, err
	}

	var ids []string
	for _, m := range memberships {
		ids = append(ids, m.MemberID)
	}
	return ids, nil
}
