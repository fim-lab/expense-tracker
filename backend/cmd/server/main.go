package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"

	"github.com/fim-lab/expense-tracker/adapters/handler/httpadapter"
	"github.com/fim-lab/expense-tracker/adapters/handler/middleware"
	"github.com/fim-lab/expense-tracker/adapters/repository/memory"
	"github.com/fim-lab/expense-tracker/adapters/repository/postgres"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
	"github.com/fim-lab/expense-tracker/internal/core/services"
	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
)

const (
	EnvTest       = "test"
	EnvDemo       = "demo"
	EnvProduction = "production"
	DefaultPort   = "8080"
)

func main() {
	env := os.Getenv("APP_ENV")
	var repos ports.Repositories

	switch env {
	case EnvProduction:
		var db *sql.DB
		db, repos = postgres.NewPostgresRepositoryCollection()
		defer db.Close()
	case EnvDemo:
		repos = memory.NewCleanRepositories()
	default:
		repos = memory.NewSeededRepositories()
	}

	// Setup services
	userService := services.NewUserService(repos.UserRepository())
	sessionService := services.NewSessionService(repos.SessionRepository())
	budgetService := services.NewBudgetService(repos.BudgetRepository(), repos.TransactionRepository())
	budgetGroupService := services.NewBudgetGroupService(repos.BudgetGroupRepository())
	walletService := services.NewWalletService(repos.WalletRepository(), repos.TransactionRepository())
	stockService := services.NewStockService(repos.StockRepository(), repos.TradeRepository())
	depotService := services.NewDepotService(repos.DepotRepository(), repos.WalletRepository(), repos.BudgetRepository(), repos.TradeRepository(), stockService)
	transactionService := services.NewTransactionService(repos.TransactionRepository(), repos.BudgetRepository(), repos.WalletRepository())
	tradeService := services.NewTradeService(repos.TradeRepository(), depotService, transactionService, stockService)
	portfolioService := services.NewPortfolioService(repos.TradeRepository(), depotService, stockService)
	templateGroupService := services.NewTemplateGroupService(repos.TemplateGroupRepository())
	transactionTemplateService := services.NewTransactionTemplateService(repos.TransactionTemplateRepository(), repos.WalletRepository(), repos.BudgetRepository(), repos.TemplateGroupRepository())
	importService := services.NewImportService(
		repos.UserRepository(),
		repos.BudgetRepository(),
		repos.BudgetGroupRepository(),
		repos.WalletRepository(),
		repos.DepotRepository(),
		repos.TransactionRepository(),
		repos.TradeRepository(),
		repos.TransactionTemplateRepository(),
		repos.TemplateGroupRepository(),
		stockService,
	)

	// Setup router
	router := chi.NewRouter()
	router.Use(middleware.RequestLogger)

	// Mount routers
	router.Mount("/auth", authRouter(&userService, &sessionService))
	router.Mount("/api", apiRouter(env, &sessionService, &budgetService, &budgetGroupService, &walletService, &depotService, &transactionService, &portfolioService, &tradeService, &userService, &transactionTemplateService, &importService, &stockService, &templateGroupService))

	log.Printf("Start Server on port %s in %s mode", DefaultPort, env)
	if err := http.ListenAndServe(":"+DefaultPort, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func authRouter(userService *ports.UserService, sessionService *ports.SessionService) http.Handler {
	r := chi.NewRouter()
	authHandler := httpadapter.NewAuthHandler(userService, sessionService)
	loginLimiter := middleware.NewLoginLimiter()
	r.With(loginLimiter.Handle).Post("/login", authHandler.Login)
	r.Post("/logout", authHandler.Logout)
	return r
}

func apiRouter(env string, sessionService *ports.SessionService, budgetService *ports.BudgetService, budgetGroupService *ports.BudgetGroupService, walletService *ports.WalletService, depotService *ports.DepotService, transactionService *ports.TransactionService, portfolioService *ports.PortfolioService, tradeService *ports.TradeService, userService *ports.UserService, transactionTemplateService *ports.TransactionTemplateService, importService *ports.ImportService, stockService *ports.StockService, templateGroupService *ports.TemplateGroupService) http.Handler {
	r := chi.NewRouter()

	// Middleware
	if env == EnvDemo {
		demoMiddleware := middleware.NewDemoMiddleware()
		r.Use(demoMiddleware.Handle)
	} else {
		authMiddleware := middleware.NewAuthMiddleware(sessionService)
		r.Use(authMiddleware.Handle)
	}

	// Handlers
	budgetHandler := httpadapter.NewBudgetHandler(budgetService)
	budgetGroupHandler := httpadapter.NewBudgetGroupHandler(budgetGroupService)
	walletHandler := httpadapter.NewWalletHandler(walletService)
	depotHandler := httpadapter.NewDepotHandler(depotService)
	transactionHandler := httpadapter.NewTransactionHandler(*transactionService, *importService)
	portfolioHandler := httpadapter.NewPortfolioHandler(*portfolioService)
	tradeHandler := httpadapter.NewTradeHandler(*tradeService)
	userHandler := httpadapter.NewUserHandler(userService)
	transactionTemplateHandler := httpadapter.NewTransactionTemplateHandler(transactionTemplateService)
	stockHandler := httpadapter.NewStockHandler(*stockService)
	templateGroupHandler := httpadapter.NewTemplateGroupHandler(templateGroupService)

	// Routes
	r.Get("/users/me", userHandler.GetUser)

	r.Get("/budgets", budgetHandler.GetBudgets)
	r.Get("/budgets/{id}", budgetHandler.GetBudget)
	r.Post("/budgets", budgetHandler.CreateBudget)
	r.Put("/budgets/{id}", budgetHandler.UpdateBudget)
	r.Delete("/budgets/{id}", budgetHandler.DeleteBudget)

	r.Get("/budget-groups", budgetGroupHandler.GetBudgetGroups)
	r.Post("/budget-groups", budgetGroupHandler.CreateBudgetGroup)
	r.Put("/budget-groups/{id}", budgetGroupHandler.UpdateBudgetGroup)
	r.Delete("/budget-groups/{id}", budgetGroupHandler.DeleteBudgetGroup)

	r.Get("/wallets", walletHandler.GetWallets)
	r.Get("/wallets/{id}", walletHandler.GetWallet)
	r.Post("/wallets", walletHandler.CreateWallet)
	r.Put("/wallets/{id}", walletHandler.UpdateWallet)
	r.Delete("/wallets/{id}", walletHandler.DeleteWallet)

	r.Get("/depots", depotHandler.GetDepots)
	r.Get("/depots/{id}", depotHandler.GetDepot)
	r.Post("/depots", depotHandler.CreateDepot)
	r.Put("/depots/{id}", depotHandler.UpdateDepot)
	r.Delete("/depots/{id}", depotHandler.DeleteDepot)

	r.Get("/transactions", transactionHandler.GetTransactions)
	r.Get("/transactions/search", transactionHandler.SearchTransactions)
	r.Get("/transactions/{id}", transactionHandler.GetTransaction)
	r.Post("/transactions", transactionHandler.CreateTransaction)
	r.Post("/transactions/transfer", transactionHandler.Transfer)
	r.Put("/transactions/{id}", transactionHandler.UpdateTransaction)
	r.Delete("/transactions/{id}", transactionHandler.DeleteTransaction)
	r.Post("/transactions/import", transactionHandler.ImportTransactions)
	r.Post("/transactions/import/testdata", transactionHandler.ImportTestData)
	r.Delete("/users/me/data", transactionHandler.DeleteAllUserData)

	r.Get("/depots/{id}/portfolio", portfolioHandler.GetPortfolio)
	r.Get("/depots/{id}/trades", portfolioHandler.GetTrades)
	r.Post("/depots/{id}/trades", tradeHandler.CreateTrade)
	r.Put("/trades/{id}", tradeHandler.UpdateTrade)
	r.Delete("/trades/{id}", tradeHandler.DeleteTrade)

	r.Get("/transaction-templates", transactionTemplateHandler.GetTransactionTemplates)
	r.Get("/transaction-templates/{id}", transactionTemplateHandler.GetTransactionTemplateByID)
	r.Post("/transaction-templates", transactionTemplateHandler.CreateTransactionTemplate)
	r.Put("/transaction-templates/{id}", transactionTemplateHandler.UpdateTransactionTemplate)
	r.Delete("/transaction-templates/{id}", transactionTemplateHandler.DeleteTransactionTemplate)

	r.Get("/template-groups", templateGroupHandler.GetTemplateGroups)
	r.Post("/template-groups", templateGroupHandler.CreateTemplateGroup)
	r.Put("/template-groups/{id}", templateGroupHandler.UpdateTemplateGroup)
	r.Delete("/template-groups/{id}", templateGroupHandler.DeleteTemplateGroup)

	r.Get("/stocks", stockHandler.GetStocks)
	r.Post("/stocks", stockHandler.CreateStock)
	r.Put("/stocks/{id}", stockHandler.UpdateStock)
	r.Delete("/stocks/{id}", stockHandler.DeleteStock)

	return r
}
