package services

import (
	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type transactionService struct {
	repo ports.ExpenseRepository
}

func NewTransactionService(repo ports.ExpenseRepository) ports.TransactionService {
	return &transactionService{repo: repo}
}

func (s *transactionService) CreateTransaction(userID int, t domain.Transaction) error {
	t.UserID = userID

	budget, err := s.repo.GetBudgetByID(t.BudgetID)
	if err != nil || budget.UserID != userID {
		return domain.ErrBudgetNotFound
	}

	if t.AmountInCents <= 0 {
		return domain.ErrInvalidAmount
	}

	return s.repo.SaveTransaction(t)
}

func (s *transactionService) GetTransactions(userID int, limit int, offset int) ([]domain.TransactionDTO, error) {
	return s.repo.FindTransactionsByUser(userID, limit, offset)
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

	transactions, err := s.repo.SearchTransactions(userID, criteria)
	if err != nil {
		return nil, err
	}

	total, err := s.repo.CountSearchedTransactions(userID, criteria)
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
	return s.repo.GetTransactionCount(userID)
}

func (s *transactionService) UpdateTransaction(userID int, t domain.Transaction) error {
	existing, err := s.repo.GetTransactionByID(t.ID)
	if err != nil || existing.UserID != userID {
		return domain.ErrUnauthorized
	}
	t.UserID = userID
	return s.repo.UpdateTransaction(t)
}

func (s *transactionService) DeleteTransaction(userID int, id int) error {
	existing, err := s.repo.GetTransactionByID(id)
	if err != nil || existing.UserID != userID {
		return domain.ErrUnauthorized
	}
	return s.repo.DeleteTransaction(id)
}
