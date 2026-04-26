package member_test

import (
	"testing"

	"kitty-party-app/internal/member"
)

func TestInMemoryRepository_CreateAndFindByID(t *testing.T) {
	repo := member.NewInMemoryRepository()

	req := member.CreateMemberRequest{
		Name:  "Test User",
		Email: "test@example.com",
		Phone: "9000000001",
	}

	created, err := repo.Create(req)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if created.ID == "" {
		t.Fatal("expected non-empty ID")
	}

	found, err := repo.FindByID(created.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if found.Email != req.Email {
		t.Errorf("expected email %s, got %s", req.Email, found.Email)
	}
}

func TestInMemoryRepository_DuplicateEmail(t *testing.T) {
	repo := member.NewInMemoryRepository()

	req := member.CreateMemberRequest{
		Name:  "Alice",
		Email: "alice@example.com",
		Phone: "9000000002",
	}

	if _, err := repo.Create(req); err != nil {
		t.Fatalf("first create failed: %v", err)
	}

	_, err := repo.Create(req)
	if err == nil {
		t.Fatal("expected conflict error for duplicate email, got nil")
	}
}

func TestInMemoryRepository_FindByID_NotFound(t *testing.T) {
	repo := member.NewInMemoryRepository()
	_, err := repo.FindByID("non-existent-id")
	if err == nil {
		t.Fatal("expected not-found error, got nil")
	}
}

func TestInMemoryRepository_Update(t *testing.T) {
	repo := member.NewInMemoryRepository()

	created, _ := repo.Create(member.CreateMemberRequest{
		Name:  "Bob",
		Email: "bob@example.com",
		Phone: "9000000003",
	})

	updated, err := repo.Update(created.ID, member.UpdateMemberRequest{Name: "Robert"})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Name != "Robert" {
		t.Errorf("expected name 'Robert', got %s", updated.Name)
	}
}

func TestInMemoryRepository_Delete(t *testing.T) {
	repo := member.NewInMemoryRepository()

	created, _ := repo.Create(member.CreateMemberRequest{
		Name:  "Charlie",
		Email: "charlie@example.com",
		Phone: "9000000004",
	})

	if err := repo.Delete(created.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	_, err := repo.FindByID(created.ID)
	if err == nil {
		t.Fatal("expected not-found after delete, got nil")
	}
}
