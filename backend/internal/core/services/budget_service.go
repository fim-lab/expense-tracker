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

func (s *budgetService) GetBudget(userID int, id int) (domain.Budget, error) {
	budget, err := s.repo.GetBudgetByID(id)
	if err != nil {
		return domain.Budget{}, err
	}

	if budget.UserID != userID {
		return domain.Budget{}, domain.ErrUnauthorized
	}

	return budget, nil
}

func (s *budgetService) GetBudgets(userID int) ([]domain.Budget, error) {
	return s.repo.FindBudgetsByUser(userID)
}

func (s *budgetService) GetTotalOfBudgets(userID int) (int, error) {
	budgets, err := s.repo.FindBudgetsByUser(userID)
	if err != nil {
		return 0, err
	}

	var totalBalance int
	for _, b := range budgets {
		totalBalance += b.BalanceCents
	}

	return totalBalance, nil
}

func (s *budgetService) UpdateBudget(userID int, budget domain.Budget) error {
	existingBudget, err := s.repo.GetBudgetByID(budget.ID)
	if err != nil {
		return err
	}

	if existingBudget.UserID != userID {
		return domain.ErrUnauthorized
	}

	if strings.TrimSpace(budget.Name) == "" {
		return domain.ErrMissingDescription
	}
	if budget.LimitCents <= 0 {
		return domain.ErrInvalidAmount
	}

	return s.repo.UpdateBudget(budget)
}

func (s *budgetService) DeleteBudget(userID int, id int) error {
	existing, err := s.repo.GetBudgetByID(id)
	if err != nil {
		return err
	}
	if existing.UserID != userID {
		return domain.ErrUnauthorized
	}
	transactionCount, err := s.repo.CountTransactionsByBudgetID(id)
	if err != nil {
		return err
	}
	if transactionCount > 0 {
		return domain.ErrNotEmpty
	}

	return s.repo.DeleteBudget(id)
}
