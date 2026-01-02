package services

import (
	"testing"
	"time"

	"github.com/fim-lab/expense-tracker/backend/adapters/repository/memory"
	"github.com/fim-lab/expense-tracker/backend/internal/core/domain"
	"github.com/google/uuid"
)

func TestTransactionOwnership(t *testing.T) {
	repo := memory.NewRepository()
	svc := NewTransactionService(repo)

	testId := uuid.New()
	tx := domain.Transaction{
		ID:            testId,
		UserID:        2,
		Description:   "Coffee",
		AmountInCents: 500,
		Date:          time.Now(),
	}
	repo.SaveTransaction(tx)

	t.Run("User 3 cannot delete User 2's transaction", func(t *testing.T) {
		err := svc.DeleteTransaction(3, testId)
		if err != domain.ErrUnauthorized {
			t.Errorf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("User 2 can delete their own transaction", func(t *testing.T) {
		err := svc.DeleteTransaction(2, testId)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}
