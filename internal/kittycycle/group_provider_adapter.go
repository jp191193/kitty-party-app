package kittycycle

import "kitty-party-app/internal/group"

// MemberInfo is a lightweight projection of a group member used by the
// kittycycle service when picking a random host.
type MemberInfo struct {
	MemberID   string
	MemberName string
}

// GroupInfoProvider abstracts the group / membership data the kittycycle
// service needs, keeping this package free of direct dependencies on group internals.
type GroupInfoProvider interface {
	GetGroup(groupID string) (*group.Group, error)
	IsMember(groupID, memberID string) (bool, error)
	IsAdminOrCreator(groupID, callerID string) (bool, error)
	ListMembers(groupID string) ([]MemberInfo, error)
}

type groupInfoProviderAdapter struct {
	groupRepo      group.Repository
	membershipRepo group.MembershipRepository
}

// NewGroupInfoProviderAdapter wraps group and membership repositories.
func NewGroupInfoProviderAdapter(groupRepo group.Repository, membershipRepo group.MembershipRepository) GroupInfoProvider {
	return &groupInfoProviderAdapter{groupRepo: groupRepo, membershipRepo: membershipRepo}
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

func (a *groupInfoProviderAdapter) ListMembers(groupID string) ([]MemberInfo, error) {
	memberships, err := a.membershipRepo.ListByGroup(groupID)
	if err != nil {
		return nil, err
	}
	out := make([]MemberInfo, 0, len(memberships))
	for _, m := range memberships {
		if m.Status != "" && m.Status != "ACTIVE" {
			continue
		}
		out = append(out, MemberInfo{MemberID: m.MemberID, MemberName: m.MemberName})
	}
	return out, nil
}
