package services

import (
	"testing"
	"time"

	"github.com/fim-lab/expense-tracker/adapters/repository/memory"
	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

func TestImportTransactions(t *testing.T) {
	repos := memory.NewCleanRepositories()
	svc := NewTransactionService(repos.TransactionRepository(), repos.BudgetRepository(), repos.WalletRepository())
	importSvc := NewImportService(repos.UserRepository(), repos.BudgetRepository(), repos.WalletRepository(), repos.DepotRepository(), repos.TransactionRepository(), repos.TradeRepository(), repos.TransactionTemplateRepository())

	userID := 1
	repos.UserRepository().SaveUser(domain.User{ID: userID, Username: "test"})
	// Pre-create some data to test skipping existing
	repos.BudgetRepository().SaveBudget(domain.Budget{Name: "Food", UserID: userID})
	repos.WalletRepository().SaveWallet(domain.Wallet{Name: "Cash", UserID: userID})

	importData := domain.FullImportData{
		Settings: domain.ImportSettings{
			Gehalt: 300000,
			Budgets: []domain.ImportBudget{
				{Name: "Food", ValueInCents: 50000, Account: "Private"},   // Existing
				{Name: "Rent", ValueInCents: 100000, Account: "Private"},  // New
				{Name: "Shared", ValueInCents: 100000, Account: "Shared"}, // Should be skipped
			},
			Wallets: []domain.ImportWallet{
				{Name: "Cash", IsDepot: false}, // Existing
				{Name: "Bank", IsDepot: false}, // New
				{Name: "Lots", IsDepot: true},  // New Depot
			},
		},
		Transactions: []domain.ImportTransaction{
			{
				Date:          time.Now(),
				Budget:        "Food",
				Wallet:        "Cash",
				Description:   "Groceries",
				AmountInCents: 1500,
				Type:          "expense",
				IsPending:     false,
			},
			{
				Date:          time.Now(),
				Budget:        "Rent",
				Wallet:        "Bank",
				Description:   "Monthly Rent",
				AmountInCents: 100000,
				Type:          "expense",
				IsPending:     false,
			},
		},
	}

	err := importSvc.ImportData(userID, importData)
	if err != nil {
		t.Fatalf("ImportData failed: %v", err)
	}

	txs, err := svc.GetTransactions(userID, 10, 0)
	if err != nil {
		t.Fatalf("Failed to fetch transactions: %v", err)
	}

	if len(txs) != 2 {
		t.Errorf("Expected 2 transactions, got %d", len(txs))
	}

	budgets, _ := repos.BudgetRepository().FindBudgetsByUser(userID)
	if len(budgets) != 2 { // Food (existing), Rent (new)
		t.Errorf("Expected 2 budgets, got %d", len(budgets))
	}

	wallets, _ := repos.WalletRepository().FindWalletsByUser(userID)
	if len(wallets) != 2 { // Cash (existing), Bank (new)
		t.Errorf("Expected 2 wallets, got %d", len(wallets))
	}

	depots, _ := repos.DepotRepository().FindDepotsByUser(userID)
	if len(depots) != 1 {
		t.Errorf("Expected 1 depot, got %d", len(depots))
	} else if depots[0].Name != "Lots" {
		t.Errorf("Expected depot name 'Lots', got '%s'", depots[0].Name)
	}

	user, _ := repos.UserRepository().GetUserByID(userID)
	if user.SalaryCents != 300000 {
		t.Errorf("Expected salary 300000, got %d", user.SalaryCents)
	}
}

func TestDeleteAllUserDataRemovesTrades(t *testing.T) {
	f := newStockFixture(t)
	importSvc := NewImportService(
		f.repos.UserRepository(),
		f.repos.BudgetRepository(),
		f.repos.WalletRepository(),
		f.repos.DepotRepository(),
		f.repos.TransactionRepository(),
		f.repos.TradeRepository(),
		f.repos.TransactionTemplateRepository(),
	)
	if err := f.repos.UserRepository().SaveUser(domain.User{ID: f.userID, Username: "test"}); err != nil {
		t.Fatalf("could not seed the user: %v", err)
	}
	f.mustBuy(t, 1, 10, 10000)
	f.mustSell(t, 2, 4, 5000)

	if err := importSvc.DeleteAllUserData(f.userID); err != nil {
		t.Fatalf("deleting all user data failed: %v", err)
	}

	if count := f.tradeCount(t); count != 0 {
		t.Errorf("expected all trades to be deleted, got %d", count)
	}
	if count := f.transactionCount(t); count != 0 {
		t.Errorf("expected all transactions to be deleted, got %d", count)
	}
	if depots, err := f.repos.DepotRepository().FindDepotsByUser(f.userID); err != nil || len(depots) != 0 {
		t.Errorf("expected no depots left, got %v (err %v)", depots, err)
	}
}
