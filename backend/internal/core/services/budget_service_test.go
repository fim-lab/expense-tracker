package services

import (
	"testing"

	"github.com/fim-lab/expense-tracker/adapters/repository/memory"
	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

func TestCreateBudget(t *testing.T) {
	repo := memory.NewRepository()
	svc := NewBudgetService(repo)

	t.Run("Valid budget creation", func(t *testing.T) {
		testId := 2
		budget := domain.Budget{
			ID:         testId,
			Name:       "Groceries",
			LimitCents: 50000,
		}
		err := svc.CreateBudget(23, budget)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		saved, _ := repo.FindBudgetsByUser(23)
		found := false
		for _, b := range saved {
			if b.Name == "Groceries" {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("could not retreive correct budget for UserId 23, got %v", saved[0])
		}
	})

	t.Run("Invalid amount", func(t *testing.T) {
		budget := domain.Budget{ID: 3, Name: "Rent", LimitCents: -100}
		err := svc.CreateBudget(3, budget)
		if err != domain.ErrInvalidAmount {
			t.Errorf("expected ErrInvalidAmount, got %v", err)
		}
	})
}
