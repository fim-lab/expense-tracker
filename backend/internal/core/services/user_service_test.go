package services

import (
	"testing"

	"github.com/fim-lab/expense-tracker/adapters/repository/memory"
	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"golang.org/x/crypto/bcrypt"
)

func TestAuthentication(t *testing.T) {
	repo := memory.NewSeededRepository()
	svc := NewUserService(repo)

	pass := "secret"
	hash, _ := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	testUser := domain.User{
		ID:           2,
		Username:     "alice",
		PasswordHash: string(hash),
	}
	repo.SaveUser(testUser)

	t.Run("User gets authenticated if right credentials are provided", func(t *testing.T) {
		authed, err := svc.Authenticate("alice", "secret")
		if err != nil {
			t.Fatalf("auth failed: %v", err)
		}
		if authed.ID != 2 {
			t.Errorf("expected ID u1, got %v", authed.ID)
		}
	})

	t.Run("User gets Unauthorized status if password is wrong", func(t *testing.T) {
		_, err := svc.Authenticate("alice", "secret1")
		if err != domain.ErrUnauthorized {
			t.Errorf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("User gets Unauthorized status if username is unknown", func(t *testing.T) {
		_, err := svc.Authenticate("bob", "secret")
		if err != domain.ErrUserNotFound {
			t.Errorf("expected ErrUserNotFound, got %v", err)
		}
	})
}
