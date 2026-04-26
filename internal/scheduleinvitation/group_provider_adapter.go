package scheduleinvitation

import "kitty-party-app/internal/group"

// GroupInfoProvider abstracts the group/membership data this domain needs.
type GroupInfoProvider interface {
	// GetGroupIDForSchedule returns the group_id for a kitty_schedule entry.
	GetGroupIDForSchedule(scheduleID string) (string, error)
	// IsAdminOrCreator returns true if callerID created or is ADMIN of groupID.
	IsAdminOrCreator(groupID, callerID string) (bool, error)
	// ListMemberIDs returns all active member IDs in a group.
	ListMemberIDs(groupID string) ([]string, error)
}

type groupInfoProviderAdapter struct {
	scheduleQuerier ScheduleQuerier
	groupRepo       group.Repository
	membershipRepo  group.MembershipRepository
}

// ScheduleQuerier abstracts the kitty_schedule look-up needed by this domain.
type ScheduleQuerier interface {
	GetGroupIDForSchedule(scheduleID string) (string, error)
}

// NewGroupInfoProviderAdapter wires up the adapter with all needed repos.
func NewGroupInfoProviderAdapter(
	sq ScheduleQuerier,
	groupRepo group.Repository,
	membershipRepo group.MembershipRepository,
) GroupInfoProvider {
	return &groupInfoProviderAdapter{
		scheduleQuerier: sq,
		groupRepo:       groupRepo,
		membershipRepo:  membershipRepo,
	}
}

func (a *groupInfoProviderAdapter) GetGroupIDForSchedule(scheduleID string) (string, error) {
	return a.scheduleQuerier.GetGroupIDForSchedule(scheduleID)
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

func (a *groupInfoProviderAdapter) ListMemberIDs(groupID string) ([]string, error) {
	memberships, err := a.membershipRepo.ListByGroup(groupID)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(memberships))
	for _, m := range memberships {
		if m.Status != "" && m.Status != "ACTIVE" {
			continue
		}
		ids = append(ids, m.MemberID)
	}
	return ids, nil
}
