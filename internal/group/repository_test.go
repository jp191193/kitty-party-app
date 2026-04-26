package group_test

import (
	"testing"
	"time"

	"kitty-party-app/internal/group"
)

func newRepo() group.Repository {
	return group.NewInMemoryRepository()
}

func validReq(email string) group.CreateGroupRequest {
	return group.CreateGroupRequest{
		Name:          "Diwali Kitty",
		MonthlyAmount: 5000,
		Duration:      12,
		StartDate:     time.Now().AddDate(0, 1, 0),
		CreatedBy:     email,
	}
}

func TestRepository_CreateAndFindByID(t *testing.T) {
	repo := newRepo()
	req := validReq("mem-001")

	g, err := repo.Create(req)
	if err != nil {
		t.Fatalf("create: unexpected error: %v", err)
	}
	if g.ID == "" {
		t.Fatal("expected non-empty ID")
	}

	found, err := repo.FindByID(g.ID)
	if err != nil {
		t.Fatalf("findByID: unexpected error: %v", err)
	}
	if found.Name != req.Name {
		t.Errorf("name: want %q, got %q", req.Name, found.Name)
	}
}

func TestRepository_FindAll(t *testing.T) {
	repo := newRepo()
	for i := 0; i < 3; i++ {
		if _, err := repo.Create(validReq("mem-001")); err != nil {
			t.Fatalf("create #%d: %v", i, err)
		}
	}

	all, err := repo.FindAll()
	if err != nil {
		t.Fatalf("findAll: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("want 3 groups, got %d", len(all))
	}
}

func TestRepository_FindByCreator(t *testing.T) {
	repo := newRepo()
	_, _ = repo.Create(validReq("alice"))
	_, _ = repo.Create(validReq("alice"))
	_, _ = repo.Create(validReq("bob"))

	aliceGroups, err := repo.FindByCreator("alice")
	if err != nil {
		t.Fatalf("findByCreator: %v", err)
	}
	if len(aliceGroups) != 2 {
		t.Errorf("want 2 groups for alice, got %d", len(aliceGroups))
	}
}

func TestRepository_FindByID_NotFound(t *testing.T) {
	_, err := newRepo().FindByID("does-not-exist")
	if err == nil {
		t.Fatal("expected not-found error, got nil")
	}
}

func TestRepository_Update(t *testing.T) {
	repo := newRepo()
	g, _ := repo.Create(validReq("mem-001"))

	updated, err := repo.Update(g.ID, group.UpdateGroupRequest{Name: "Holi Kitty", Duration: 6})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if updated.Name != "Holi Kitty" {
		t.Errorf("name: want 'Holi Kitty', got %q", updated.Name)
	}
	if updated.Duration != 6 {
		t.Errorf("duration: want 6, got %d", updated.Duration)
	}
}

func TestRepository_Update_NotFound(t *testing.T) {
	_, err := newRepo().Update("ghost-id", group.UpdateGroupRequest{Name: "Phantom"})
	if err == nil {
		t.Fatal("expected not-found error, got nil")
	}
}

func TestRepository_Delete(t *testing.T) {
	repo := newRepo()
	g, _ := repo.Create(validReq("mem-001"))

	if err := repo.Delete(g.ID); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if _, err := repo.FindByID(g.ID); err == nil {
		t.Fatal("expected not-found after delete, got nil")
	}
}

func TestRepository_Delete_NotFound(t *testing.T) {
	err := newRepo().Delete("ghost-id")
	if err == nil {
		t.Fatal("expected not-found error, got nil")
	}
}
