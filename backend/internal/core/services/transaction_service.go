package services

import (
	"github.com/fim-lab/expense-tracker/backend/internal/core/domain"
	"github.com/fim-lab/expense-tracker/backend/internal/core/ports"
	"github.com/google/uuid"
)

type transactionService struct {
	repo ports.ExpenseRepository
}

func NewTransactionService(repo ports.ExpenseRepository) ports.TransactionService {
	return &transactionService{repo: repo}
}

func (s *transactionService) CreateTransaction(userID int, t domain.Transaction) error {
	t.ID = uuid.New()
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

func (s *transactionService) GetTransactions(userID int) ([]domain.Transaction, error) {
	return s.repo.FindTransactionsByUser(userID)
}

// TODO: UpdateTransactions

func (s *transactionService) DeleteTransaction(userID int, id uuid.UUID) error {
	existing, err := s.repo.GetTransactionByID(id)
	if err != nil || existing.UserID != userID {
		return domain.ErrUnauthorized
	}
	return s.repo.DeleteTransaction(id)
}
