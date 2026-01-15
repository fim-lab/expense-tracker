package services

import (
	"log"
	"strings"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type walletService struct {
	repo ports.ExpenseRepository
}

func NewWalletService(repo ports.ExpenseRepository) ports.WalletService {
	return &walletService{repo: repo}
}

func (s *walletService) CreateWallet(userID int, b domain.Wallet) error {
	b.UserID = userID

	if strings.TrimSpace(b.Name) == "" {
		return domain.ErrMissingWallet
	}

	return s.repo.SaveWallet(b)
}

func (s *walletService) GetWallets(userID int) ([]domain.Wallet, error) {
	return s.repo.FindWalletsByUser(userID)
}

func (s *walletService) GetTotalOfWallets(userID int) (int, error) {
	wallets, err := s.repo.FindWalletsByUser(userID)
	if err != nil {
		return 0, err
	}
	log.Print("LOG")

	var totalBalance int
	for _, w := range wallets {
		totalBalance += w.BalanceCents
	}

	return totalBalance, nil
}

// TODO: func (s *walletService) UpdateWallets(...

func (s *walletService) DeleteWallet(userID int, id int) error {
	existing, err := s.repo.GetWalletByID(id)
	if err != nil || existing.UserID != userID {
		return domain.ErrUnauthorized
	}
	return s.repo.DeleteWallet(id)
}
