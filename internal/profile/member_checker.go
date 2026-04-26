package profile

import "kitty-party-app/internal/member"

// memberCheckerAdapter wraps member.Repository to satisfy the MemberChecker interface.
type memberCheckerAdapter struct {
	repo member.Repository
}

// NewMemberChecker wraps a member.Repository as a MemberChecker.
func NewMemberChecker(repo member.Repository) MemberChecker {
	return &memberCheckerAdapter{repo: repo}
}

func (a *memberCheckerAdapter) Exists(memberID string) error {
	_, err := a.repo.FindByID(memberID)
	return err
}
