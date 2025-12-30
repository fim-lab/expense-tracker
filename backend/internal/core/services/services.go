package services

import (
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

func (s *service) CreateTransaction(t domain.Transaction) error {
	return s.repo.Save(t)
}

func (s *service) GetTransactions() ([]domain.Transaction, error) {
	return s.repo.FindAll()
}

func (s *service) DeleteTransaction(id string) error {
	return s.repo.Delete(id)
}