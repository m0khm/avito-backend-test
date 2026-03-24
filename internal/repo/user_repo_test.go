package repo

import (
	"context"
	"testing"

	"room-booking-service/internal/domain"
)

func TestUserRepository_UpsertAndGetByEmail(t *testing.T) {
	db := testDB(t)
	r := NewUserRepository(db)

	user := domain.User{
		ID:    "00000000-0000-0000-0000-000000000201",
		Email: "repo-user@example.com",
		Role:  domain.RoleUser,
	}

	if err := r.Upsert(context.Background(), user); err != nil {
		t.Fatalf("Upsert() error = %v", err)
	}

	got, err := r.GetByEmail(context.Background(), user.Email)
	if err != nil {
		t.Fatalf("GetByEmail() error = %v", err)
	}
	if got.Email != user.Email {
		t.Fatalf("expected %s, got %s", user.Email, got.Email)
	}
}

func TestUserRepository_Create(t *testing.T) {
	db := testDB(t)
	r := NewUserRepository(db)

	passwordHash := "hashed-password"
	user := domain.User{
		ID:           "00000000-0000-0000-0000-000000000202",
		Email:        "repo-create-user@example.com",
		Role:         domain.RoleAdmin,
		PasswordHash: &passwordHash,
	}

	created, err := r.Create(context.Background(), user)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if created.CreatedAt == nil {
		t.Fatal("expected CreatedAt to be set")
	}

	got, err := r.GetByEmail(context.Background(), user.Email)
	if err != nil {
		t.Fatalf("GetByEmail() error = %v", err)
	}
	if got.ID != user.ID {
		t.Fatalf("expected id %s, got %s", user.ID, got.ID)
	}
	if got.Role != user.Role {
		t.Fatalf("expected role %s, got %s", user.Role, got.Role)
	}
	if got.PasswordHash == nil || *got.PasswordHash != passwordHash {
		t.Fatalf("expected password hash %q, got %+v", passwordHash, got.PasswordHash)
	}
}