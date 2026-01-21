package services

import (
	"testing"

	"github.com/fim-lab/expense-tracker/adapters/repository/memory"
	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

func TestCreateWallet(t *testing.T) {
	repo := memory.NewSeededRepository()
	svc := NewWalletService(repo)

	t.Run("Valid wallet creation", func(t *testing.T) {
		wallet := domain.Wallet{
			Name:         "Cash",
			BalanceCents: 100,
		}
		err := svc.CreateWallet(23, wallet)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		saved, _ := repo.FindWalletsByUser(23)
		found := false
		for _, w := range saved {
			if w.Name == "Cash" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("could not retreive correct wallet for UserId 23, got %v", saved[0])
		}
	})

	t.Run("Sums totals correct", func(t *testing.T) {
		wallet1 := domain.Wallet{
			Name:         "Giro",
			BalanceCents: 20,
		}
		err := svc.CreateWallet(23, wallet1)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		wallet2 := domain.Wallet{
			Name:         "Others",
			BalanceCents: -49,
		}
		err = svc.CreateWallet(23, wallet2)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		total, _ := svc.GetTotalOfWallets(23)

		if total != (100 + 20 - 49) {
			t.Errorf("expected 71 cents total for wallets for UserId 23, got %v", total)
		}

		total, _ = svc.GetTotalOfWallets(27)

		if total != 0 {
			t.Errorf("expected 0 balance for unknown user, but got %v", total)
		}
	})
}

func TestDeleteWallet(t *testing.T) {
	userID := 1

	t.Run("Successfully delete empty wallet", func(t *testing.T) {
		repo := memory.NewCleanRepository()
		svc := NewWalletService(repo)
		wallet := domain.Wallet{UserID: userID, Name: "Test Wallet"}
		svc.CreateWallet(userID, wallet)
		wallets, _ := repo.FindWalletsByUser(userID)
		testWallet := wallets[0]

		err := svc.DeleteWallet(userID, testWallet.ID)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		_, err = svc.GetWallet(userID, testWallet.ID)
		if err != domain.ErrWalletNotFound {
			t.Errorf("Expected ErrWalletNotFound, got %v", err)
		}
	})

	t.Run("Fail to delete wallet with transactions", func(t *testing.T) {
		repo := memory.NewCleanRepository()
		svc := NewWalletService(repo)
		wallet := domain.Wallet{UserID: userID, Name: "Test Wallet"}
		svc.CreateWallet(userID, wallet)
		wallets, _ := repo.FindWalletsByUser(userID)
		testWallet := wallets[0]

		_ = repo.SaveTransaction(domain.Transaction{
			UserID:        userID,
			WalletID:      testWallet.ID,
			AmountInCents: 100,
			Description:   "Test Transaction",
		})

		err := svc.DeleteWallet(userID, testWallet.ID)
		if err != domain.ErrNotEmpty {
			t.Errorf("Expected ErrNotEmpty, got %v", err)
		}

		_, err = svc.GetWallet(userID, testWallet.ID)
		if err != nil {
			t.Errorf("Expected wallet to exist, got error %v", err)
		}
	})

	t.Run("Unauthorized deletion", func(t *testing.T) {
		repo := memory.NewCleanRepository()
		svc := NewWalletService(repo)
		wallet := domain.Wallet{UserID: userID, Name: "Test Wallet"}
		svc.CreateWallet(userID, wallet)
		wallets, _ := repo.FindWalletsByUser(userID)
		testWallet := wallets[0]

		err := svc.DeleteWallet(999, testWallet.ID)
		if err != domain.ErrUnauthorized {
			t.Errorf("Expected ErrUnauthorized, got %v", err)
		}
	})
}
