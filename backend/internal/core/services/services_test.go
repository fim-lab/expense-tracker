package services

import (
	"testing"
	"github.com/fim-lab/expense-tracker/backend/adapters/repository/memory"
	"github.com/fim-lab/expense-tracker/backend/internal/core/domain"
)

func TestCreateAndGetTransactions(t *testing.T) {
	repo := memory.NewRepository()
	service := NewExpenseService(repo)

	testTx := domain.Transaction{
		ID:            "123",
		Description:   "Test Transaction",
		AmountInCents: 1000,
	}

	err := service.CreateTransaction(testTx)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	txs, err := service.GetTransactions()
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(txs) != 1 {
		t.Errorf("Expected 1 transaction, got %d", len(txs))
	}

	if txs[0].ID != "123" {
		t.Errorf("Expected ID 123, got %s", txs[0].ID)
	}
}