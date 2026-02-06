package services

import (
	"strings"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type walletService struct {
	walletRepo      ports.WalletRepository
	transactionRepo ports.TransactionRepository
}

func NewWalletService(walletRepo ports.WalletRepository, transactionRepo ports.TransactionRepository) ports.WalletService {
	return &walletService{walletRepo: walletRepo, transactionRepo: transactionRepo}
}

func (s *walletService) CreateWallet(userID int, b domain.Wallet) error {
	b.UserID = userID

	if strings.TrimSpace(b.Name) == "" {
		return domain.ErrMissingWallet
	}

	return s.walletRepo.SaveWallet(b)
}

func (s *walletService) GetWallet(userID int, id int) (domain.Wallet, error) {
	wallet, err := s.walletRepo.GetWalletByID(id)
	if err != nil {
		return domain.Wallet{}, err
	}

	if wallet.UserID != userID {
		return domain.Wallet{}, domain.ErrUnauthorized
	}

	wallet.CanDelete = true
	if wallet.BalanceCents != 0 {
		wallet.CanDelete = false
	} else {
		count, err := s.transactionRepo.CountTransactionsByWalletID(wallet.ID)
		if err != nil {
			return domain.Wallet{}, err
		}
		if count > 0 {
			wallet.CanDelete = false
		}
	}

	return wallet, nil
}

func (s *walletService) GetWallets(userID int) ([]domain.Wallet, error) {
	wallets, err := s.walletRepo.FindWalletsByUser(userID)
	if err != nil {
		return nil, err
	}

	for i := range wallets {
		wallets[i].CanDelete = true
		if wallets[i].BalanceCents != 0 {
			wallets[i].CanDelete = false
		} else {
			count, err := s.transactionRepo.CountTransactionsByWalletID(wallets[i].ID)
			if err != nil {
				return nil, err
			}
			if count > 0 {
				wallets[i].CanDelete = false
			}
		}
	}

	return wallets, nil
}

func (s *walletService) GetTotalOfWallets(userID int) (int, error) {
	wallets, err := s.walletRepo.FindWalletsByUser(userID)
	if err != nil {
		return 0, err
	}

	var totalBalance int
	for _, w := range wallets {
		totalBalance += w.BalanceCents
	}

	return totalBalance, nil
}

func (s *walletService) UpdateWallet(userID int, wallet domain.Wallet) error {
	existingWallet, err := s.walletRepo.GetWalletByID(wallet.ID)
	if err != nil {
		return err
	}

	if existingWallet.UserID != userID {
		return domain.ErrUnauthorized
	}

	if strings.TrimSpace(wallet.Name) == "" {
		return domain.ErrMissingDescription
	}

	return s.walletRepo.UpdateWallet(wallet)
}

func (s *walletService) DeleteWallet(userID int, id int) error {
	existing, err := s.walletRepo.GetWalletByID(id)
	if err != nil {
		return err
	}
	if existing.UserID != userID {
		return domain.ErrUnauthorized
	}
	transactionCount, err := s.transactionRepo.CountTransactionsByWalletID(id)
	if err != nil {
		return err
	}
	if transactionCount > 0 {
		return domain.ErrNotEmpty
	}

	return s.walletRepo.DeleteWallet(id)
}
