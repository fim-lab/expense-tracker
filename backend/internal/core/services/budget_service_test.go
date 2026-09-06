package services

import (
	"testing"
	"time"

	"github.com/fim-lab/expense-tracker/adapters/repository/memory"
	"github.com/fim-lab/expense-tracker/internal/core/domain"
	_ "github.com/fim-lab/expense-tracker/internal/core/ports"
)

func TestCreateBudget(t *testing.T) {
	repos := memory.NewSeededRepositories()
	svc := NewBudgetService(repos.BudgetRepository(), repos.TransactionRepository())

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

		saved, _ := repos.BudgetRepository().FindBudgetsByUser(23)
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
		repos := memory.NewCleanRepositories()
		svc := NewBudgetService(repos.BudgetRepository(), repos.TransactionRepository())
		budget := domain.Budget{UserID: userID, Name: "Non-Zero Balance", LimitCents: 10000, BalanceCents: 500}
		svc.CreateBudget(userID, budget)
		budgets, _ := repos.BudgetRepository().FindBudgetsByUser(userID)
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
		repos := memory.NewCleanRepositories()
		svc := NewBudgetService(repos.BudgetRepository(), repos.TransactionRepository())
		budget := domain.Budget{UserID: userID, Name: "Zero Balance, Has Transactions", LimitCents: 10000, BalanceCents: -100}
		svc.CreateBudget(userID, budget)
		budgets, _ := repos.BudgetRepository().FindBudgetsByUser(userID)
		createdBudget := budgets[0]

		_, err := repos.TransactionRepository().SaveTransaction(domain.Transaction{
			UserID:        userID,
			BudgetID:      &createdBudget.ID,
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
		repos := memory.NewCleanRepositories()
		svc := NewBudgetService(repos.BudgetRepository(), repos.TransactionRepository())
		budget := domain.Budget{UserID: userID, Name: "Zero Balance, No Transactions", LimitCents: 10000, BalanceCents: 0}
		svc.CreateBudget(userID, budget)
		budgets, _ := repos.BudgetRepository().FindBudgetsByUser(userID)
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
	repos := memory.NewCleanRepositories()
	svc := NewBudgetService(repos.BudgetRepository(), repos.TransactionRepository())

	budget1 := domain.Budget{UserID: userID, Name: "Budget 1", LimitCents: 10000, BalanceCents: 500}
	svc.CreateBudget(userID, budget1)

	budget2 := domain.Budget{UserID: userID, Name: "Budget 2", LimitCents: 10000, BalanceCents: -100}
	svc.CreateBudget(userID, budget2)
	budgets, _ := repos.BudgetRepository().FindBudgetsByUser(userID)
	createdBudget2 := budgets[1]

	_, err := repos.TransactionRepository().SaveTransaction(domain.Transaction{
		UserID:        userID,
		BudgetID:      &createdBudget2.ID,
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

func TestCreateBudgetTransfer(t *testing.T) {
	userID := 1

	t.Run("Successful transfer moves balance between budgets", func(t *testing.T) {
		repos := memory.NewCleanRepositories()
		svc := NewBudgetService(repos.BudgetRepository(), repos.TransactionRepository())
		svc.CreateBudget(userID, domain.Budget{Name: "From", LimitCents: 10000, BalanceCents: 500})
		svc.CreateBudget(userID, domain.Budget{Name: "To", LimitCents: 10000, BalanceCents: 100})
		budgets, _ := repos.BudgetRepository().FindBudgetsByUser(userID)
		fromBudget, toBudget := budgets[0], budgets[1]

		err := svc.CreateBudgetTransfer(userID, fromBudget.ID, toBudget.ID, 300)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		updated, _ := repos.BudgetRepository().FindBudgetsByUser(userID)
		for _, b := range updated {
			if b.ID == fromBudget.ID && b.BalanceCents != 200 {
				t.Errorf("expected from-budget balance 200, got %v", b.BalanceCents)
			}
			if b.ID == toBudget.ID && b.BalanceCents != 400 {
				t.Errorf("expected to-budget balance 400, got %v", b.BalanceCents)
			}
		}
	})

	t.Run("Same budget transfer is rejected", func(t *testing.T) {
		repos := memory.NewCleanRepositories()
		svc := NewBudgetService(repos.BudgetRepository(), repos.TransactionRepository())
		svc.CreateBudget(userID, domain.Budget{Name: "Only", LimitCents: 10000, BalanceCents: 500})
		budgets, _ := repos.BudgetRepository().FindBudgetsByUser(userID)

		err := svc.CreateBudgetTransfer(userID, budgets[0].ID, budgets[0].ID, 100)
		if err != domain.ErrSameBudgetTransfer {
			t.Errorf("expected ErrSameBudgetTransfer, got %v", err)
		}
	})

	t.Run("Non-positive amount is rejected", func(t *testing.T) {
		repos := memory.NewCleanRepositories()
		svc := NewBudgetService(repos.BudgetRepository(), repos.TransactionRepository())
		svc.CreateBudget(userID, domain.Budget{Name: "From", LimitCents: 10000, BalanceCents: 500})
		svc.CreateBudget(userID, domain.Budget{Name: "To", LimitCents: 10000, BalanceCents: 100})
		budgets, _ := repos.BudgetRepository().FindBudgetsByUser(userID)

		err := svc.CreateBudgetTransfer(userID, budgets[0].ID, budgets[1].ID, 0)
		if err != domain.ErrInvalidAmount {
			t.Errorf("expected ErrInvalidAmount, got %v", err)
		}
	})

	t.Run("Transfer involving another user's budget is unauthorized", func(t *testing.T) {
		repos := memory.NewCleanRepositories()
		svc := NewBudgetService(repos.BudgetRepository(), repos.TransactionRepository())
		svc.CreateBudget(userID, domain.Budget{Name: "Mine", LimitCents: 10000, BalanceCents: 500})
		svc.CreateBudget(999, domain.Budget{Name: "Theirs", LimitCents: 10000, BalanceCents: 100})
		mine, _ := repos.BudgetRepository().FindBudgetsByUser(userID)
		theirs, _ := repos.BudgetRepository().FindBudgetsByUser(999)

		err := svc.CreateBudgetTransfer(userID, mine[0].ID, theirs[0].ID, 100)
		if err != domain.ErrBudgetNotFound {
			t.Errorf("expected ErrBudgetNotFound, got %v", err)
		}
	})
}

func TestDeleteBudget(t *testing.T) {
	userID := 1

	t.Run("Successfully delete empty budget", func(t *testing.T) {
		repos := memory.NewCleanRepositories()
		svc := NewBudgetService(repos.BudgetRepository(), repos.TransactionRepository())
		budget := domain.Budget{UserID: userID, Name: "Test Budget", LimitCents: 10000}
		svc.CreateBudget(userID, budget)
		budgets, _ := repos.BudgetRepository().FindBudgetsByUser(userID)
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
		repos := memory.NewCleanRepositories()
		svc := NewBudgetService(repos.BudgetRepository(), repos.TransactionRepository())
		budget := domain.Budget{UserID: userID, Name: "Test Budget", LimitCents: 10000}
		svc.CreateBudget(userID, budget)
		budgets, _ := repos.BudgetRepository().FindBudgetsByUser(userID)
		testBudget := budgets[0]

		_, err := repos.TransactionRepository().SaveTransaction(domain.Transaction{
			UserID:        userID,
			BudgetID:      &testBudget.ID,
			AmountInCents: 100,
			Description:   "Test Transaction",
			Date:          time.Now(),
		})

		err = svc.DeleteBudget(userID, testBudget.ID)
		if err != domain.ErrNotEmpty {
			t.Errorf("Expected ErrNotEmpty, got %v", err)
		}

		_, err = svc.GetBudget(userID, testBudget.ID)
		if err != nil {
			t.Errorf("Expected budget to exist, got error %v", err)
		}
	})

	t.Run("Unauthorized deletion", func(t *testing.T) {
		repos := memory.NewCleanRepositories()
		svc := NewBudgetService(repos.BudgetRepository(), repos.TransactionRepository())
		budget := domain.Budget{UserID: userID, Name: "Test Budget", LimitCents: 10000}
		svc.CreateBudget(userID, budget)
		budgets, _ := repos.BudgetRepository().FindBudgetsByUser(userID)
		testBudget := budgets[0]

		err := svc.DeleteBudget(999, testBudget.ID)
		if err != domain.ErrUnauthorized {
			t.Errorf("Expected ErrUnauthorized, got %v", err)
		}
	})
}
