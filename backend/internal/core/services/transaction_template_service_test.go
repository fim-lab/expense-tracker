package services

import (
	"testing"

	"github.com/fim-lab/expense-tracker/adapters/repository/memory"
	"github.com/fim-lab/expense-tracker/internal/core/domain"
	_ "github.com/fim-lab/expense-tracker/internal/core/ports"
)

func TestTransactionTemplateService(t *testing.T) {
	repos := memory.NewCleanRepositories()
	svc := NewTransactionTemplateService(
		repos.TransactionTemplateRepository(),
		repos.WalletRepository(),
		repos.BudgetRepository(),
	)

	testUser := domain.User{Username: "test", PasswordHash: "hash"}
	repos.UserRepository().SaveUser(testUser)
	user, _ := repos.UserRepository().GetUserByUsername("test")

	testWallet := domain.Wallet{UserID: user.ID, Name: "Test Wallet", BalanceCents: 0}
	repos.WalletRepository().SaveWallet(testWallet)
	wallet, _ := repos.WalletRepository().FindWalletsByUser(user.ID)
	walletId := wallet[0].ID

	testBudget := domain.Budget{UserID: user.ID, Name: "Test Budget", LimitCents: 10000}
	repos.BudgetRepository().SaveBudget(testBudget)
	budget, _ := repos.BudgetRepository().FindBudgetsByUser(user.ID)
	budgetId := budget[0].ID

	t.Run("CreateTransactionTemplate - Valid", func(t *testing.T) {
		template := domain.TransactionTemplate{
			UserID:        user.ID,
			Day:           15,
			BudgetID:      &budgetId,
			WalletID:      walletId,
			Description:   "Monthly Subscription",
			AmountInCents: 1500,
			Type:          domain.Expense,
		}

		err := svc.CreateTransactionTemplate(user.ID, template)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		fetched, err := svc.GetTransactionTemplates(user.ID)
		if err != nil {
			t.Errorf("Expected no error fetching template, got %v", err)
		}
		if len(fetched) != 1 {
			t.Errorf("Expected 1 Template, but got %v", len(fetched))
		}
		if fetched[0].Description != template.Description {
			t.Errorf("Expected description %s, got %s", template.Description, fetched[0].Description)
		}
	})

	t.Run("CreateTransactionTemplate - Invalid Template Data", func(t *testing.T) {
		template := domain.TransactionTemplate{
			UserID:        user.ID,
			Day:           0,
			WalletID:      walletId,
			Description:   "Invalid Day",
			AmountInCents: 100,
			Type:          domain.Expense,
		}
		err := svc.CreateTransactionTemplate(user.ID, template)
		if err == nil {
			t.Errorf("Expected error for invalid template data, got none")
		}
	})

	t.Run("CreateTransactionTemplate - Invalid Wallet", func(t *testing.T) {
		template := domain.TransactionTemplate{
			UserID:        user.ID,
			Day:           1,
			WalletID:      999,
			Description:   "Invalid Wallet",
			AmountInCents: 100,
			Type:          domain.Expense,
		}
		err := svc.CreateTransactionTemplate(user.ID, template)
		if err == nil {
			t.Errorf("Expected error for invalid wallet, got none")
		}
		if err != nil && err.Error() != "wallet validation failed: wallet not found or unauthorized" {
			t.Errorf("Expected 'wallet validation failed', got %v", err)
		}
	})

	t.Run("CreateTransactionTemplate - Invalid Budget", func(t *testing.T) {
		invalidBudgetID := 999
		template := domain.TransactionTemplate{
			UserID:        user.ID,
			Day:           1,
			WalletID:      walletId,
			BudgetID:      &invalidBudgetID,
			Description:   "Invalid Budget",
			AmountInCents: 100,
			Type:          domain.Expense,
		}
		err := svc.CreateTransactionTemplate(user.ID, template)
		if err == nil {
			t.Errorf("Expected error for invalid budget, got none")
		}
		if err != nil && err.Error() != "budget validation failed: budget not found or unauthorized" {
			t.Errorf("Expected 'budget validation failed', got %v", err)
		}
	})

	t.Run("GetTransactionTemplate - NotFound", func(t *testing.T) {
		_, err := svc.GetTransactionTemplate(user.ID, 999)
		if err != domain.ErrTransactionTemplateNotFound {
			t.Errorf("Expected ErrTransactionTemplateNotFound, got %v", err)
		}
	})

	t.Run("GetTransactionTemplates - Valid", func(t *testing.T) {
		templates, err := svc.GetTransactionTemplates(user.ID)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}
		if len(templates) == 0 {
			t.Errorf("Expected to get templates, got 0")
		}
	})

	t.Run("UpdateTransactionTemplate - Valid", func(t *testing.T) {
		templatesToUpdate, err := repos.TransactionTemplateRepository().FindTransactionTemplatesByUser(user.ID)
		if (err != nil) || (len(templatesToUpdate) < 1) {
			t.Error("Expected at least one Template.")
		}

		templateToUpdate := templatesToUpdate[0]
		templateToUpdate.Description = "Updated Description"
		templateToUpdate.AmountInCents = 2345
		err = svc.UpdateTransactionTemplate(user.ID, templateToUpdate)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		fetched, err := svc.GetTransactionTemplate(user.ID, templateToUpdate.ID)
		if err != nil {
			t.Errorf("Expected no error fetching updated template, got %v", err)
		}
		if fetched.Description != templateToUpdate.Description {
			t.Errorf("Expected description %s, got %s", templateToUpdate.Description, fetched.Description)
		}
	})

	t.Run("DeleteTransactionTemplate - Valid", func(t *testing.T) {

		templates, err := svc.GetTransactionTemplates(user.ID)
		if err != nil || len(templates) < 1 {
			t.Errorf("Expected at least one template.")
		}
		templateToDelete := templates[0]

		err = svc.DeleteTransactionTemplate(user.ID, templateToDelete.ID)
		if err != nil {
			t.Errorf("Expected no error, got %v", err)
		}

		_, err = svc.GetTransactionTemplate(user.ID, templateToDelete.ID)
		if err != domain.ErrTransactionTemplateNotFound {
			t.Errorf("Expected ErrTransactionTemplateNotFound after deletion, got %v", err)
		}
	})

	t.Run("DeleteTransactionTemplate - NotFound", func(t *testing.T) {
		err := svc.DeleteTransactionTemplate(user.ID, 999)
		if err != domain.ErrTransactionTemplateNotFound {
			t.Errorf("Expected ErrTransactionTemplateNotFound, got %v", err)
		}
	})
}
