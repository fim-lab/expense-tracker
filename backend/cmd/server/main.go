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
	chimiddleware "github.com/go-chi/chi/v5/middleware"
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
	walletService := services.NewWalletService(repos.WalletRepository(), repos.TransactionRepository())
	depotService := services.NewDepotService(repos.DepotRepository(), repos.WalletRepository())
	transactionService := services.NewTransactionService(repos.TransactionRepository(), repos.BudgetRepository(), repos.WalletRepository())
	stockService := services.NewStockService(repos.StockRepository(), repos.DepotRepository())

	// Setup router
	router := chi.NewRouter()
	router.Use(chimiddleware.Logger)

	// Mount routers
	router.Mount("/auth", authRouter(&userService, &sessionService))
	router.Mount("/api", apiRouter(env, &sessionService, &budgetService, &walletService, &depotService, &transactionService, &stockService, &userService))

	log.Printf("Start Server on port %s in %s mode", DefaultPort, env)
	if err := http.ListenAndServe(":"+DefaultPort, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func authRouter(userService *ports.UserService, sessionService *ports.SessionService) http.Handler {
	r := chi.NewRouter()
	authHandler := httpadapter.NewAuthHandler(userService, sessionService)
	r.Post("/login", authHandler.Login)
	r.Post("/logout", authHandler.Logout)
	return r
}

func apiRouter(env string, sessionService *ports.SessionService, budgetService *ports.BudgetService, walletService *ports.WalletService, depotService *ports.DepotService, transactionService *ports.TransactionService, stockService *ports.StockService, userService *ports.UserService) http.Handler {
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
	walletHandler := httpadapter.NewWalletHandler(walletService)
	depotHandler := httpadapter.NewDepotHandler(depotService)
	transactionHandler := httpadapter.NewTransactionHandler(transactionService)
	stockHandler := httpadapter.NewStockHandler(stockService)
	userHandler := httpadapter.NewUserHandler(userService)

	// Routes
	r.Get("/users/me", userHandler.GetUser)
	r.Put("/users/me/salary", userHandler.UpdateSalary)

	r.Get("/budgets", budgetHandler.GetBudgets)
	r.Get("/budgets/{id}", budgetHandler.GetBudget)
	r.Post("/budgets", budgetHandler.CreateBudget)
	r.Put("/budgets/{id}", budgetHandler.UpdateBudget)
	r.Delete("/budgets/{id}", budgetHandler.DeleteBudget)

	r.Get("/wallets", walletHandler.GetWallets)
	r.Get("/wallets/{id}", walletHandler.GetWallet)
	r.Post("/wallets", walletHandler.CreateWallet)
	r.Put("/wallets/{id}", walletHandler.UpdateWallet)
	r.Delete("/wallets/{id}", walletHandler.DeleteWallet)

	r.Get("/depots", depotHandler.GetDepots)
	r.Post("/depots", depotHandler.CreateDepot)
	r.Delete("/depots/{id}", depotHandler.DeleteDepot)

	r.Get("/transactions", transactionHandler.GetTransactions)
	r.Get("/transactions/search", transactionHandler.SearchTransactions)
	r.Get("/transactions/{id}", transactionHandler.GetTransaction)
	r.Post("/transactions", transactionHandler.CreateTransaction)
	r.Post("/transactions/transfer", transactionHandler.Transfer)
	r.Put("/transactions/{id}", transactionHandler.UpdateTransaction)
	r.Delete("/transactions/{id}", transactionHandler.DeleteTransaction)

	r.Get("/stocks", stockHandler.GetStocks)
	r.Post("/stocks", stockHandler.CreateStock)
	r.Delete("/stocks/{id}", stockHandler.DeleteStock)

	return r
}
