package services

import (
	"fmt"
	"strings"

	"github.com/fim-lab/expense-tracker/backend/internal/core/domain"
	"github.com/fim-lab/expense-tracker/backend/internal/core/ports"
)

type service struct {
	repo ports.ExpenseRepository
}

func NewExpenseService(repo ports.ExpenseRepository) ports.ExpenseService {
	return &service{
		repo: repo,
	}
}

func (s *service) validateTransaction(t domain.Transaction) error {
	if t.AmountInCents <= 0 {
		return domain.ErrInvalidAmount
	}
	if strings.TrimSpace(t.Description) == "" {
		return domain.ErrMissingDescription
	}
	if strings.TrimSpace(t.Budget) == "" {
		return domain.ErrMissingBudget
	}
	return nil
}

func (s *service) CreateTransaction(t domain.Transaction) error {
	if err := s.validateTransaction(t); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	return s.repo.Save(t)
}

func (s *service) GetTransactions() ([]domain.Transaction, error) {
	return s.repo.FindAll()
}

func (s *service) DeleteTransaction(id string) error {
	if strings.TrimSpace(id) == "" {
		return fmt.Errorf("id is required for deletion")
	}
	return s.repo.Delete(id)
}
