package payout

import "kitty-party-app/internal/group"

// GroupInfoProvider abstracts the group and membership data the payout service needs,
// keeping the payout package free of direct dependencies on group internals.
type GroupInfoProvider interface {
	// GetGroup returns the group to validate existence and read Duration/MonthlyAmount.
	GetGroup(groupID string) (*group.Group, error)
	// IsMember reports whether memberID is an active member of groupID.
	IsMember(groupID, memberID string) (bool, error)
	// IsAdminOrCreator reports whether callerID is the group creator or has ADMIN role.
	IsAdminOrCreator(groupID, callerID string) (bool, error)
}

type groupInfoProviderAdapter struct {
	groupRepo      group.Repository
	membershipRepo group.MembershipRepository
}

// NewGroupInfoProviderAdapter wraps group and membership repositories as a GroupInfoProvider.
func NewGroupInfoProviderAdapter(groupRepo group.Repository, membershipRepo group.MembershipRepository) GroupInfoProvider {
	return &groupInfoProviderAdapter{
		groupRepo:      groupRepo,
		membershipRepo: membershipRepo,
	}
}

func (a *groupInfoProviderAdapter) GetGroup(groupID string) (*group.Group, error) {
	return a.groupRepo.FindByID(groupID)
}

func (a *groupInfoProviderAdapter) IsMember(groupID, memberID string) (bool, error) {
	return a.membershipRepo.Exists(groupID, memberID)
}

func (a *groupInfoProviderAdapter) IsAdminOrCreator(groupID, callerID string) (bool, error) {
	g, err := a.groupRepo.FindByID(groupID)
	if err != nil {
		return false, err
	}
	if g.CreatedBy == callerID {
		return true, nil
	}
	role, err := a.membershipRepo.GetRole(groupID, callerID)
	if err != nil {
		return false, err
	}
	return role == "ADMIN", nil
}
