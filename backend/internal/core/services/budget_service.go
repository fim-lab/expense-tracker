package services

import (
	"strings"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type budgetService struct {
	repo ports.ExpenseRepository
}

func NewBudgetService(repo ports.ExpenseRepository) ports.BudgetService {
	return &budgetService{repo: repo}
}

func (s *budgetService) CreateBudget(userID int, b domain.Budget) error {
	b.UserID = userID

	if strings.TrimSpace(b.Name) == "" {
		return domain.ErrMissingBudget
	}

	if b.LimitCents <= 0 {
		return domain.ErrInvalidAmount
	}

	return s.repo.SaveBudget(b)
}

func (s *budgetService) GetBudgets(userID int) ([]domain.Budget, error) {
	return s.repo.FindBudgetsByUser(userID)
}

// TODO: func (s *budgetService) UpdateBudgets(...

func (s *budgetService) DeleteBudget(userID int, id int) error {
	existing, err := s.repo.GetBudgetByID(id)
	if err != nil || existing.UserID != userID {
		return domain.ErrUnauthorized
	}
	return s.repo.DeleteBudget(id)
}
