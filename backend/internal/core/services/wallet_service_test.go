package services

import (
	"testing"

	"github.com/fim-lab/expense-tracker/adapters/repository/memory"
	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

func TestCreateWallet(t *testing.T) {
	repo := memory.NewRepository()
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
