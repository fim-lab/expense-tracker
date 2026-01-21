package services

import (
	"testing"
	"time"

	"github.com/fim-lab/expense-tracker/adapters/repository/memory"
	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

func TestCreateBudget(t *testing.T) {
	repo := memory.NewSeededRepository()
	svc := NewBudgetService(repo)

	t.Run("Valid budget creation", func(t *testing.T) {
		budget := domain.Budget{
			Name:         "Groceries",
			LimitCents:   50000,
			BalanceCents: 100,
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

	t.Run("Sums totals correct", func(t *testing.T) {
		budget1 := domain.Budget{
			Name:         "More",
			LimitCents:   10,
			BalanceCents: 20,
		}
		err := svc.CreateBudget(23, budget1)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}
		budget2 := domain.Budget{
			Name:         "Others",
			LimitCents:   50000,
			BalanceCents: -49,
		}
		err = svc.CreateBudget(23, budget2)
		if err != nil {
			t.Errorf("expected no error, got %v", err)
		}

		total, _ := svc.GetTotalOfBudgets(23)

		if total != (100 + 20 - 49) {
			t.Errorf("expected 71 cents total for budgets for UserId 23, got %v", total)
		}

		total, _ = svc.GetTotalOfBudgets(27)

		if total != 0 {
			t.Errorf("expected 0 balance for unknown user, but got %v", total)
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

func TestGetBudgetCanDelete(t *testing.T) {
	userID := 1

	t.Run("CanDelete is false when BalanceCents is not zero", func(t *testing.T) {
		repo := memory.NewCleanRepository()
		svc := NewBudgetService(repo)
		budget := domain.Budget{UserID: userID, Name: "Non-Zero Balance", LimitCents: 10000, BalanceCents: 500}
		svc.CreateBudget(userID, budget)
		budgets, _ := repo.FindBudgetsByUser(userID)
		createdBudget := budgets[0]

		fetchedBudget, err := svc.GetBudget(userID, createdBudget.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if fetchedBudget.CanDelete {
			t.Errorf("Expected CanDelete to be false, got true")
		}
	})

	t.Run("CanDelete is false when BalanceCents is zero but transactions exist", func(t *testing.T) {
		repo := memory.NewCleanRepository()
		svc := NewBudgetService(repo)
		budget := domain.Budget{UserID: userID, Name: "Zero Balance, Has Transactions", LimitCents: 10000, BalanceCents: -100}
		svc.CreateBudget(userID, budget)
		budgets, _ := repo.FindBudgetsByUser(userID)
		createdBudget := budgets[0]

		_ = repo.SaveTransaction(domain.Transaction{
			UserID:        userID,
			BudgetID:      createdBudget.ID,
			AmountInCents: 100,
			Description:   "Test Transaction",
			Date:          time.Now(),
		})

		fetchedBudget, err := svc.GetBudget(userID, createdBudget.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if fetchedBudget.CanDelete {
			t.Errorf("Expected CanDelete to be false, got true")
		}
	})

	t.Run("CanDelete is true when BalanceCents is zero and no transactions exist", func(t *testing.T) {
		repo := memory.NewCleanRepository()
		svc := NewBudgetService(repo)
		budget := domain.Budget{UserID: userID, Name: "Zero Balance, No Transactions", LimitCents: 10000, BalanceCents: 0}
		svc.CreateBudget(userID, budget)
		budgets, _ := repo.FindBudgetsByUser(userID)
		createdBudget := budgets[0]

		fetchedBudget, err := svc.GetBudget(userID, createdBudget.ID)
		if err != nil {
			t.Fatalf("Expected no error, got %v", err)
		}
		if !fetchedBudget.CanDelete {
			t.Errorf("Expected CanDelete to be true, got false")
		}
	})
}

func TestGetBudgetsCanDelete(t *testing.T) {
	userID := 1
	repo := memory.NewCleanRepository()
	svc := NewBudgetService(repo)

	budget1 := domain.Budget{UserID: userID, Name: "Budget 1", LimitCents: 10000, BalanceCents: 500}
	svc.CreateBudget(userID, budget1)

	budget2 := domain.Budget{UserID: userID, Name: "Budget 2", LimitCents: 10000, BalanceCents: -100}
	svc.CreateBudget(userID, budget2)
	budgets, _ := repo.FindBudgetsByUser(userID)
	createdBudget2 := budgets[1]

	_ = repo.SaveTransaction(domain.Transaction{
		UserID:        userID,
		BudgetID:      createdBudget2.ID,
		AmountInCents: 100,
		Description:   "Test Transaction",
		Date:          time.Now(),
	})

	budget3 := domain.Budget{UserID: userID, Name: "Budget 3", LimitCents: 10000, BalanceCents: 0}
	svc.CreateBudget(userID, budget3)

	fetchedBudgets, err := svc.GetBudgets(userID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	for _, b := range fetchedBudgets {
		switch b.Name {
		case "Budget 1":
			if b.CanDelete {
				t.Errorf("Budget 1: Expected CanDelete to be false, got true")
			}
		case "Budget 2":
			if b.CanDelete {
				t.Errorf("Budget 2: Expected CanDelete to be false, got true")
			}
		case "Budget 3":
			if !b.CanDelete {
				t.Errorf("Budget 3: Expected CanDelete to be true, got false")
			}
		}
	}
}

func TestDeleteBudget(t *testing.T) {
	userID := 1

	t.Run("Successfully delete empty budget", func(t *testing.T) {
		repo := memory.NewCleanRepository()
		svc := NewBudgetService(repo)
		budget := domain.Budget{UserID: userID, Name: "Test Budget", LimitCents: 10000}
		svc.CreateBudget(userID, budget)
		budgets, _ := repo.FindBudgetsByUser(userID)
		testBudget := budgets[0]

		err := svc.DeleteBudget(userID, testBudget.ID)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		_, err = svc.GetBudget(userID, testBudget.ID)
		if err != domain.ErrBudgetNotFound {
			t.Errorf("Expected ErrBudgetNotFound, got %v", err)
		}
	})

	t.Run("Fail to delete budget with transactions", func(t *testing.T) {
		repo := memory.NewCleanRepository()
		svc := NewBudgetService(repo)
		budget := domain.Budget{UserID: userID, Name: "Test Budget", LimitCents: 10000}
		svc.CreateBudget(userID, budget)
		budgets, _ := repo.FindBudgetsByUser(userID)
		testBudget := budgets[0]

		_ = repo.SaveTransaction(domain.Transaction{
			UserID:        userID,
			BudgetID:      testBudget.ID,
			AmountInCents: 100,
			Description:   "Test Transaction",
			Date:          time.Now(),
		})

		err := svc.DeleteBudget(userID, testBudget.ID)
		if err != domain.ErrNotEmpty {
			t.Errorf("Expected ErrNotEmpty, got %v", err)
		}

		_, err = svc.GetBudget(userID, testBudget.ID)
		if err != nil {
			t.Errorf("Expected budget to exist, got error %v", err)
		}
	})

	t.Run("Unauthorized deletion", func(t *testing.T) {
		repo := memory.NewCleanRepository()
		svc := NewBudgetService(repo)
		budget := domain.Budget{UserID: userID, Name: "Test Budget", LimitCents: 10000}
		svc.CreateBudget(userID, budget)
		budgets, _ := repo.FindBudgetsByUser(userID)
		testBudget := budgets[0]

		err := svc.DeleteBudget(999, testBudget.ID)
		if err != domain.ErrUnauthorized {
			t.Errorf("Expected ErrUnauthorized, got %v", err)
		}
	})
}
