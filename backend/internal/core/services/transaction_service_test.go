package services

import (
	"log"
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
		tx, err := svc.GetTransactions(2, 2, 0)
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
		tx, err := svc.GetTransactions(2, 2, 0)
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

func TestGetTransactions_PaginationAndMapping(t *testing.T) {
	repo := memory.NewRepository()
	svc := NewTransactionService(repo)

	testUsername := "testuser"
	repo.SaveUser(domain.User{Username: testUsername, PasswordHash: "#"})
	testUser, err := repo.GetUserByUsername(testUsername)
	if err != nil {
		log.Fatal("Could not save test User.")
	}

	repo.SaveBudget(domain.Budget{Name: "Food", UserID: testUser.ID})
	repo.SaveWallet(domain.Wallet{Name: "Cash", UserID: testUser.ID})
	testBudget, err := repo.FindBudgetsByUser(testUser.ID)
	if err != nil {
		return
	}
	testWallet, err := repo.FindWalletsByUser(testUser.ID)
	if err != nil {
		return
	}

	now := time.Now()
	txs := []domain.Transaction{
		{Date: now.Add(-3 * time.Hour), BudgetID: &testBudget[0].ID, WalletID: testWallet[0].ID, Description: "Older", UserID: testUser.ID},
		{Date: now.Add(-1 * time.Hour), BudgetID: &testBudget[0].ID, WalletID: testWallet[0].ID, Description: "Middle", UserID: testUser.ID},
		{Date: now, BudgetID: &testBudget[0].ID, WalletID: testWallet[0].ID, Description: "Newest", UserID: testUser.ID},
		{Date: now, BudgetID: &testBudget[0].ID, WalletID: testWallet[0].ID, Description: "Newest", UserID: 99},
	}
	for _, tx := range txs {
		repo.SaveTransaction(tx)
	}

	t.Run("Verify Sorting and DTO Mapping", func(t *testing.T) {
		results, err := svc.GetTransactions(testUser.ID, 1, 0)
		if err != nil {
			t.Fatalf("Failed to get transactions: %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}

		if results[0].Description != "Newest" {
			t.Errorf("Expected 'Newest' transaction, got %s", results[0].Description)
		}

		if results[0].BudgetName != "Food" {
			t.Errorf("Expected BudgetName 'Food', got '%s'", results[0].BudgetName)
		}
		if results[0].WalletName != "Cash" {
			t.Errorf("Expected WalletName 'Cash', got '%s'", results[0].WalletName)
		}
	})

	t.Run("Verify Pagination Offset", func(t *testing.T) {
		results, err := svc.GetTransactions(testUser.ID, 2, 1)
		if err != nil {
			t.Fatalf("Failed to get transactions: %v", err)
		}

		if len(results) != 2 {
			t.Fatalf("Expected 1 result, got %d", len(results))
		}

		if results[0].Description != "Middle" {
			t.Errorf("Expected 'Middle' transaction at offset 1, got %s", results[0].Description)
		}
	})

	t.Run("Verify Out of Bounds Offset", func(t *testing.T) {
		results, err := svc.GetTransactions(testUser.ID, 10, 100)
		if err != nil {
			t.Fatalf("Failed: %v", err)
		}
		if len(results) != 0 {
			t.Errorf("Expected 0 results for huge offset, got %d", len(results))
		}
	})

	t.Run("Verify total count of transactions", func(t *testing.T) {
		results, err := svc.GetTransactionCount(testUser.ID)
		if err != nil {
			t.Fatalf("Failed: %v", err)
		}
		if results != 3 {
			t.Errorf("Expected 3 transactions for testUser, got %d", results)
		}
	})
}
