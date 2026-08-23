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
	stockSvc := NewStockService(repos.StockRepository(), repos.TradeRepository())
	importSvc := NewImportService(repos.UserRepository(), repos.BudgetRepository(), repos.BudgetGroupRepository(), repos.WalletRepository(), repos.DepotRepository(), repos.TransactionRepository(), repos.TradeRepository(), repos.TransactionTemplateRepository(), stockSvc)

	userID := 1
	repos.UserRepository().SaveUser(domain.User{ID: userID, Username: "test"})
	// Pre-create some data to test skipping existing
	repos.BudgetRepository().SaveBudget(domain.Budget{Name: "Food", UserID: userID})
	repos.WalletRepository().SaveWallet(domain.Wallet{Name: "Cash", UserID: userID})

	importData := domain.FullImportData{
		Settings: domain.ImportSettings{
			Gehalt: 300000,
			Budgets: []domain.ImportBudget{
				{Name: "Food", ValueInCents: 50000, Account: "Private"},
				{Name: "Rent", ValueInCents: 100000, Account: "Private"},
				{Name: "Shared", ValueInCents: 100000, Account: "Shared"},
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
	if len(budgets) != 3 { // Food (existing), Rent (new), Shared (new)
		t.Errorf("Expected 3 budgets, got %d", len(budgets))
	}

	groups, _ := repos.BudgetGroupRepository().FindBudgetGroupsByUser(userID)
	if len(groups) != 1 {
		t.Errorf("Expected 1 budget group, got %d", len(groups))
	} else if groups[0].Name != "Shared" {
		t.Errorf("Expected budget group name 'Shared', got '%s'", groups[0].Name)
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

func TestImportTransactions_TradesAndSpecialCases(t *testing.T) {
	f := newStockFixture(t)
	importSvc := NewImportService(
		f.repos.UserRepository(),
		f.repos.BudgetRepository(),
		f.repos.BudgetGroupRepository(),
		f.repos.WalletRepository(),
		f.repos.DepotRepository(),
		f.repos.TransactionRepository(),
		f.repos.TradeRepository(),
		f.repos.TransactionTemplateRepository(),
		f.stockSvc,
	)
	if err := f.repos.UserRepository().SaveUser(domain.User{ID: f.userID, Username: "test"}); err != nil {
		t.Fatalf("could not seed the user: %v", err)
	}

	importData := domain.FullImportData{
		Transactions: []domain.ImportTransaction{
			{
				Date:          time.Now(),
				Budget:        "Übertrag",
				Wallet:        "Main Wallet",
				Description:   "Splitwise",
				AmountInCents: -1030,
				Type:          "expense",
				IsDebt:        true,
			},
			{
				Date:          time.Now(),
				Budget:        "Investments",
				Wallet:        "Main Wallet",
				Description:   "Buy 1.5Stk. TEST",
				AmountInCents: -25000,
				Type:          "expense",
			},
			{
				Date:          time.Now().Add(24 * time.Hour),
				Budget:        "Investments",
				Wallet:        "Main Wallet",
				Description:   "Sell 1 Stk. TEST",
				AmountInCents: 20000,
				Type:          "income",
			},
		},
	}

	if err := importSvc.ImportData(f.userID, importData); err != nil {
		t.Fatalf("ImportData failed: %v", err)
	}

	trades, err := f.repos.TradeRepository().FindTradesByDepot(f.depotID)
	if err != nil {
		t.Fatalf("could not read trades: %v", err)
	}
	if len(trades) != 2 {
		t.Fatalf("expected 2 trades to be created, got %d", len(trades))
	}
	for _, trade := range trades {
		stock, err := f.repos.StockRepository().GetStockByID(trade.StockID)
		if err != nil {
			t.Fatalf("could not read stock %d: %v", trade.StockID, err)
		}
		if stock.WKN != "TEST" {
			t.Errorf("expected WKN TEST, got %s", stock.WKN)
		}
		if trade.WalletTransactionID == nil {
			t.Errorf("expected trade %d to be linked to a wallet transaction", trade.ID)
		}
	}

	dtos, err := f.repos.TransactionRepository().FindTransactionsByUser(f.userID, 10, 0)
	if err != nil {
		t.Fatalf("could not read transactions: %v", err)
	}
	if len(dtos) != 3 {
		t.Fatalf("expected 3 transactions, got %d", len(dtos))
	}

	var splitwiseID int
	for _, dto := range dtos {
		if dto.Description == "Splitwise" {
			splitwiseID = dto.ID
			if dto.BudgetName != "" {
				t.Errorf("expected 'Übertrag' to import with no budget, got %q", dto.BudgetName)
			}
		}
		if dto.AmountInCents <= 0 {
			t.Errorf("expected imported amounts to always be positive, got %d for %q", dto.AmountInCents, dto.Description)
		}
	}
	if splitwiseID == 0 {
		t.Fatalf("could not find the Splitwise transaction")
	}

	splitwise, err := f.txSvc.GetTransactionByID(f.userID, splitwiseID)
	if err != nil {
		t.Fatalf("could not read the Splitwise transaction: %v", err)
	}
	if splitwise.IsDebt == nil || !*splitwise.IsDebt {
		t.Errorf("expected the Splitwise transaction to keep isDebt=true")
	}
}

func TestDeleteAllUserDataRemovesTrades(t *testing.T) {
	f := newStockFixture(t)
	importSvc := NewImportService(
		f.repos.UserRepository(),
		f.repos.BudgetRepository(),
		f.repos.BudgetGroupRepository(),
		f.repos.WalletRepository(),
		f.repos.DepotRepository(),
		f.repos.TransactionRepository(),
		f.repos.TradeRepository(),
		f.repos.TransactionTemplateRepository(),
		f.stockSvc,
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
