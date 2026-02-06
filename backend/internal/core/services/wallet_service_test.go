package services

import (
	"testing"

	"github.com/fim-lab/expense-tracker/adapters/repository/memory"
	"github.com/fim-lab/expense-tracker/internal/core/domain"
	_ "github.com/fim-lab/expense-tracker/internal/core/ports"
)

func TestCreateWallet(t *testing.T) {
	repos := memory.NewSeededRepositories()
	svc := NewWalletService(repos.WalletRepository(), repos.TransactionRepository())

	t.Run("Valid wallet creation", func(t *testing.T) {
		wallet := domain.Wallet{
			Name:         "Cash",
			BalanceCents: 100,
		}
		err := svc.CreateWallet(23, wallet)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		saved, _ := repos.WalletRepository().FindWalletsByUser(23)
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

func TestGetWalletCanDelete(t *testing.T) {
	userID := 1

	t.Run("CanDelete is false when BalanceCents is not zero", func(t *testing.T) {
		repos := memory.NewCleanRepositories()
		svc := NewWalletService(repos.WalletRepository(), repos.TransactionRepository())
		wallet := domain.Wallet{UserID: userID, Name: "Non-Zero Balance", BalanceCents: 500}
		svc.CreateWallet(userID, wallet)
		wallets, _ := repos.WalletRepository().FindWalletsByUser(userID)
		createdWallet := wallets[0]

		fetchedWallet, err := svc.GetWallet(userID, createdWallet.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if fetchedWallet.CanDelete {
			t.Errorf("Expected CanDelete to be false, got true")
		}
	})

	t.Run("CanDelete is false when BalanceCents is zero but transactions exist", func(t *testing.T) {
		repos := memory.NewCleanRepositories()
		svc := NewWalletService(repos.WalletRepository(), repos.TransactionRepository())
		wallet := domain.Wallet{UserID: userID, Name: "Zero Balance, Has Transactions", BalanceCents: -100}
		svc.CreateWallet(userID, wallet)
		wallets, _ := repos.WalletRepository().FindWalletsByUser(userID)
		createdWallet := wallets[0]

		_ = repos.TransactionRepository().SaveTransaction(domain.Transaction{
			UserID:        userID,
			WalletID:      createdWallet.ID,
			AmountInCents: 100,
			Description:   "Test Transaction",
		})

		fetchedWallet, err := svc.GetWallet(userID, createdWallet.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if fetchedWallet.CanDelete {
			t.Errorf("Expected CanDelete to be false, got true")
		}
	})

	t.Run("CanDelete is true when BalanceCents is zero and no transactions exist", func(t *testing.T) {
		repos := memory.NewCleanRepositories()
		svc := NewWalletService(repos.WalletRepository(), repos.TransactionRepository())
		wallet := domain.Wallet{UserID: userID, Name: "Zero Balance, No Transactions", BalanceCents: 0}
		svc.CreateWallet(userID, wallet)
		wallets, _ := repos.WalletRepository().FindWalletsByUser(userID)
		createdWallet := wallets[0]

		fetchedWallet, err := svc.GetWallet(userID, createdWallet.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !fetchedWallet.CanDelete {
			t.Errorf("Expected CanDelete to be true, got false")
		}
	})
}

func TestGetWalletsCanDelete(t *testing.T) {
	userID := 1
	repos := memory.NewCleanRepositories()
	svc := NewWalletService(repos.WalletRepository(), repos.TransactionRepository())

	wallet1 := domain.Wallet{UserID: userID, Name: "Wallet 1", BalanceCents: 500}
	svc.CreateWallet(userID, wallet1)

	wallet2 := domain.Wallet{UserID: userID, Name: "Wallet 2", BalanceCents: -100}
	svc.CreateWallet(userID, wallet2)
	wallets, _ := repos.WalletRepository().FindWalletsByUser(userID)
	createdWallet2 := wallets[1]

	_ = repos.TransactionRepository().SaveTransaction(domain.Transaction{
		UserID:        userID,
		WalletID:      createdWallet2.ID,
		AmountInCents: 100,
		Description:   "Test Transaction",
	})

	wallet3 := domain.Wallet{UserID: userID, Name: "Wallet 3", BalanceCents: 0}
	svc.CreateWallet(userID, wallet3)

	fetchedWallets, err := svc.GetWallets(userID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	for _, w := range fetchedWallets {
		switch w.Name {
		case "Wallet 1":
			if w.CanDelete {
				t.Errorf("Wallet 1: Expected CanDelete to be false, got true")
			}
		case "Wallet 2":
			if w.CanDelete {
				t.Errorf("Wallet 2: Expected CanDelete to be false, got true")
			}
		case "Wallet 3":
			if !w.CanDelete {
				t.Errorf("Wallet 3: Expected CanDelete to be true, got false")
			}
		}
	}
}

func TestDeleteWallet(t *testing.T) {
	userID := 1

	t.Run("Successfully delete empty wallet", func(t *testing.T) {
		repos := memory.NewCleanRepositories()
		svc := NewWalletService(repos.WalletRepository(), repos.TransactionRepository())
		wallet := domain.Wallet{UserID: userID, Name: "Test Wallet"}
		svc.CreateWallet(userID, wallet)
		wallets, _ := repos.WalletRepository().FindWalletsByUser(userID)
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
		repos := memory.NewCleanRepositories()
		svc := NewWalletService(repos.WalletRepository(), repos.TransactionRepository())
		wallet := domain.Wallet{UserID: userID, Name: "Test Wallet"}
		svc.CreateWallet(userID, wallet)
		wallets, _ := repos.WalletRepository().FindWalletsByUser(userID)
		testWallet := wallets[0]

		_ = repos.TransactionRepository().SaveTransaction(domain.Transaction{
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
		repos := memory.NewCleanRepositories()
		svc := NewWalletService(repos.WalletRepository(), repos.TransactionRepository())
		wallet := domain.Wallet{UserID: userID, Name: "Test Wallet"}
		svc.CreateWallet(userID, wallet)
		wallets, _ := repos.WalletRepository().FindWalletsByUser(userID)
		testWallet := wallets[0]

		err := svc.DeleteWallet(999, testWallet.ID)
		if err != domain.ErrUnauthorized {
			t.Errorf("Expected ErrUnauthorized, got %v", err)
		}
	})
}
