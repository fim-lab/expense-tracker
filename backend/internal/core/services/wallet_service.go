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

func (s *walletService) GetWallet(userID int, id int) (domain.Wallet, error) {
	wallet, err := s.repo.GetWalletByID(id)
	if err != nil {
		return domain.Wallet{}, err
	}

	if wallet.UserID != userID {
		return domain.Wallet{}, domain.ErrUnauthorized
	}

	return wallet, nil
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

func (s *walletService) UpdateWallet(userID int, wallet domain.Wallet) error {
	existingWallet, err := s.repo.GetWalletByID(wallet.ID)
	if err != nil {
		return err
	}

	if existingWallet.UserID != userID {
		return domain.ErrUnauthorized
	}

	if strings.TrimSpace(wallet.Name) == "" {
		return domain.ErrMissingDescription
	}

	return s.repo.UpdateWallet(wallet)
}

func (s *walletService) DeleteWallet(userID int, id int) error {
	existing, err := s.repo.GetWalletByID(id)
	if err != nil {
		return err
	}
	if existing.UserID != userID {
		return domain.ErrUnauthorized
	}
	transactionCount, err := s.repo.CountTransactionsByWalletID(id)
	if err != nil {
		return err
	}
	if transactionCount > 0 {
		return domain.ErrNotEmpty
	}

	return s.repo.DeleteWallet(id)
}
