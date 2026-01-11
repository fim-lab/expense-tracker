package services

import (
	"testing"
	"time"

	"github.com/fim-lab/expense-tracker/adapters/repository/memory"
	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

func TestTransactionOwnership(t *testing.T) {
	repo := memory.NewRepository()
	svc := NewTransactionService(repo)

	tx := domain.Transaction{
		UserID:        2,
		Description:   "Coffee",
		AmountInCents: 500,
		Date:          time.Now(),
	}
	repo.SaveTransaction(tx)

	t.Run("User 3 cannot delete User 2's transaction", func(t *testing.T) {
		tx, err := svc.GetTransactions(2)
		if err != nil {
			t.Errorf("expected saved transaction for testUser, got %v", err)
		}
		txId := tx[len(tx)-1].ID
		err = svc.DeleteTransaction(3, txId)
		if err != domain.ErrUnauthorized {
			t.Errorf("expected ErrUnauthorized, got %v", err)
		}
	})

	t.Run("User 2 can delete their own transaction", func(t *testing.T) {
		tx, err := svc.GetTransactions(2)
		if err != nil {
			t.Errorf("expected saved transaction for testUser, got %v", err)
		}
		txId := tx[len(tx)-1].ID
		err = svc.DeleteTransaction(2, txId)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
	})
}
