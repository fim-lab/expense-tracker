package services

import (
	"fmt"
	"time"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type transactionService struct {
	transactionRepo ports.TransactionRepository
	budgetRepo      ports.BudgetRepository
	walletRepo      ports.WalletRepository
}

func NewTransactionService(transactionRepo ports.TransactionRepository, budgetRepo ports.BudgetRepository, walletRepo ports.WalletRepository) ports.TransactionService {
	return &transactionService{transactionRepo: transactionRepo, budgetRepo: budgetRepo, walletRepo: walletRepo}
}

func (s *transactionService) CreateTransaction(userID int, t domain.Transaction) error {
	t.UserID = userID

	if t.BudgetID != nil {
		budget, err := s.budgetRepo.GetBudgetByID(*t.BudgetID)
		if err != nil || budget.UserID != userID {
			return domain.ErrBudgetNotFound
		}
	}

	if t.AmountInCents <= 0 {
		return domain.ErrInvalidAmount
	}

	return s.transactionRepo.SaveTransaction(t)
}

func (s *transactionService) CreateTransfer(userID, fromWalletID, toWalletID, amount int) error {
	if fromWalletID == toWalletID {
		return domain.ErrSameWalletTransfer
	}

	fromWallet, err := s.walletRepo.GetWalletByID(fromWalletID)
	if err != nil || fromWallet.UserID != userID {
		return domain.ErrWalletNotFound
	}

	toWallet, err := s.walletRepo.GetWalletByID(toWalletID)
	if err != nil || toWallet.UserID != userID {
		return domain.ErrWalletNotFound
	}

	if amount <= 0 {
		return domain.ErrInvalidAmount
	}

	fromTransaction := domain.Transaction{
		UserID:        userID,
		Date:          time.Now(),
		WalletID:      fromWalletID,
		Description:   fmt.Sprintf("Transfer to %s", toWallet.Name),
		AmountInCents: amount,
		Type:          domain.Expense,
	}

	toTransaction := domain.Transaction{
		UserID:        userID,
		Date:          time.Now(),
		WalletID:      toWalletID,
		Description:   fmt.Sprintf("Transfer from %s", fromWallet.Name),
		AmountInCents: amount,
		Type:          domain.Income,
	}

	return s.transactionRepo.CreateTransfer(fromTransaction, toTransaction)
}

func (s *transactionService) GetTransactions(userID int, limit int, offset int) ([]domain.TransactionDTO, error) {
	return s.transactionRepo.FindTransactionsByUser(userID, limit, offset)
}

func (s *transactionService) Search(userID int, criteria domain.TransactionSearchCriteria) (*domain.PaginatedTransactions, error) {
	if criteria.Page <= 0 {
		criteria.Page = 1
	}
	if criteria.PageSize <= 0 {
		criteria.PageSize = 10
	} else if criteria.PageSize > 100 {
		criteria.PageSize = 100
	}

	transactions, err := s.transactionRepo.SearchTransactions(userID, criteria)
	if err != nil {
		return nil, err
	}

	total, err := s.transactionRepo.CountSearchedTransactions(userID, criteria)
	if err != nil {
		return nil, err
	}

	return &domain.PaginatedTransactions{
		Transactions: transactions,
		Total:        total,
		Page:         criteria.Page,
		PageSize:     criteria.PageSize,
	}, nil
}

func (s *transactionService) GetTransactionCount(userID int) (int, error) {
	return s.transactionRepo.GetTransactionCount(userID)
}

func (s *transactionService) UpdateTransaction(userID int, t domain.Transaction) error {
	existing, err := s.transactionRepo.GetTransactionByID(t.ID)
	if err != nil || existing.UserID != userID {
		return domain.ErrUnauthorized
	}
	t.UserID = userID
	return s.transactionRepo.UpdateTransaction(t)
}

func (s *transactionService) DeleteTransaction(userID int, id int) error {
	existing, err := s.transactionRepo.GetTransactionByID(id)
	if err != nil || existing.UserID != userID {
		return domain.ErrUnauthorized
	}
	return s.transactionRepo.DeleteTransaction(id)
}

func (s *transactionService) GetTransactionByID(userID int, id int) (domain.Transaction, error) {
	transaction, err := s.transactionRepo.GetTransactionByID(id)
	if err != nil {
		return domain.Transaction{}, err
	}
	if transaction.UserID != userID {
		return domain.Transaction{}, domain.ErrUnauthorized
	}
	return transaction, nil
}

