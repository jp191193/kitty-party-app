package group_test

import (
	"fmt"
	"testing"
	"time"

	"kitty-party-app/internal/group"
)

// helpers ────────────────────────────────────────────────────────────────────

func newGroupRepo() group.Repository {
	return group.NewInMemoryRepository()
}

func newMembershipRepo() group.MembershipRepository {
	return group.NewInMemoryMembershipRepository()
}

func seedGroup(t *testing.T, repo group.Repository) *group.Group {
	t.Helper()
	g, err := repo.Create(group.CreateGroupRequest{
		Name:          "Test Kitty",
		MonthlyAmount: 1000,
		Duration:      12,
		StartDate:     time.Now().AddDate(0, 1, 0),
		CreatedBy:     "organiser-001",
	})
	if err != nil {
		t.Fatalf("seed group: %v", err)
	}
	return g
}

// ── Repository tests ────────────────────────────────────────────────────────

func TestMembershipRepository_AddAndList(t *testing.T) {
	repo := newMembershipRepo()

	m, err := repo.Add("group-1", "member-1")
	if err != nil {
		t.Fatalf("Add: %v", err)
	}
	if m.ID == "" || m.GroupID != "group-1" || m.MemberID != "member-1" {
		t.Errorf("unexpected membership: %+v", m)
	}

	list, err := repo.ListByGroup("group-1")
	if err != nil {
		t.Fatalf("ListByGroup: %v", err)
	}
	if len(list) != 1 {
		t.Errorf("want 1, got %d", len(list))
	}
}

func TestMembershipRepository_Exists(t *testing.T) {
	repo := newMembershipRepo()
	_, _ = repo.Add("group-1", "member-1")

	exists, err := repo.Exists("group-1", "member-1")
	if err != nil || !exists {
		t.Errorf("Exists: want true, got false (err=%v)", err)
	}

	exists, err = repo.Exists("group-1", "member-999")
	if err != nil || exists {
		t.Errorf("Exists: want false for unknown member, got true (err=%v)", err)
	}
}

func TestMembershipRepository_Count(t *testing.T) {
	repo := newMembershipRepo()
	for i := 0; i < 5; i++ {
		_, _ = repo.Add("group-1", fmt.Sprintf("member-%d", i))
	}

	count, err := repo.Count("group-1")
	if err != nil || count != 5 {
		t.Errorf("Count: want 5, got %d (err=%v)", count, err)
	}
}

// ── Service (business-rule) tests ───────────────────────────────────────────

func newMembershipService() (group.MembershipService, group.Repository) {
	gRepo := newGroupRepo()
	mRepo := newMembershipRepo()
	svc := group.NewMembershipService(gRepo, mRepo)
	return svc, gRepo
}

func TestMembershipService_AddMember_HappyPath(t *testing.T) {
	svc, gRepo := newMembershipService()
	g := seedGroup(t, gRepo)

	m, err := svc.AddMember(g.ID, "organiser-001", "User", group.AddMemberRequest{MemberID: "mem-001"})
	if err != nil {
		t.Fatalf("AddMember: %v", err)
	}
	if m.MemberID != "mem-001" {
		t.Errorf("want member_id mem-001, got %s", m.MemberID)
	}
}

func TestMembershipService_AddMember_GroupNotFound(t *testing.T) {
	svc, _ := newMembershipService()

	_, err := svc.AddMember("non-existent-group", "user1", "User", group.AddMemberRequest{MemberID: "mem-001"})
	if err == nil {
		t.Fatal("expected error for unknown group, got nil")
	}
}

func TestMembershipService_AddMember_Duplicate(t *testing.T) {
	svc, gRepo := newMembershipService()
	g := seedGroup(t, gRepo)

	_, _ = svc.AddMember(g.ID, "organiser-001", "User", group.AddMemberRequest{MemberID: "mem-dup"})
	_, err := svc.AddMember(g.ID, "organiser-001", "User", group.AddMemberRequest{MemberID: "mem-dup"})
	if err == nil {
		t.Fatal("expected conflict error for duplicate member, got nil")
	}
}

func TestMembershipService_AddMember_MaxCapacity(t *testing.T) {
	svc, gRepo := newMembershipService()
	g := seedGroup(t, gRepo)

	for i := 0; i < group.MaxGroupMembers; i++ {
		if _, err := svc.AddMember(g.ID, "organiser-001", "User", group.AddMemberRequest{MemberID: fmt.Sprintf("mem-%d", i)}); err != nil {
			t.Fatalf("AddMember #%d: %v", i, err)
		}
	}

	// The (MaxGroupMembers+1)th addition must fail.
	_, err := svc.AddMember(g.ID, "organiser-001", "User", group.AddMemberRequest{MemberID: "mem-overflow"})
	if err == nil {
		t.Fatalf("expected full-group error, got nil")
	}
}

func TestMembershipService_ListMembers_GroupNotFound(t *testing.T) {
	svc, _ := newMembershipService()
	_, err := svc.ListMembers("ghost-group")
	if err == nil {
		t.Fatal("expected error for unknown group, got nil")
	}
}
