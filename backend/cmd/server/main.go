package main

import (
	"database/sql"
	"fmt"
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
	EnvDemo       = "demo"
	EnvProduction = "production"
	DefaultPort   = "8080"
)

func main() {
	env := determineEnvironment()
	repo, db := getRepo(env)
	if db != nil {
		defer db.Close()
	}

	// Setup services
	userService := services.NewUserService(repo)
	sessionService := services.NewSessionService(repo)
	budgetService := services.NewBudgetService(repo)
	walletService := services.NewWalletService(repo)
	depotService := services.NewDepotService(repo)
	transactionService := services.NewTransactionService(repo)
	stockService := services.NewStockService(repo)

	// Setup router
	router := chi.NewRouter()
	router.Use(chimiddleware.Logger)

	// Mount routers
	router.Mount("/auth", authRouter(&userService, &sessionService))
	router.Mount("/api", apiRouter(env, &sessionService, &budgetService, &walletService, &depotService, &transactionService, &stockService))

	port := getPort()
	log.Printf("Start Server on port %s in %s mode", port, env)
	if err := http.ListenAndServe(":"+port, router); err != nil {
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

func apiRouter(env string, sessionService *ports.SessionService, budgetService *ports.BudgetService, walletService *ports.WalletService, depotService *ports.DepotService, transactionService *ports.TransactionService, stockService *ports.StockService) http.Handler {
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

	// Routes
	r.Get("/budgets", budgetHandler.GetBudgets)
	r.Post("/budgets", budgetHandler.CreateBudget)
	r.Delete("/budgets/{id}", budgetHandler.DeleteBudget)

	r.Get("/wallets", walletHandler.GetWallets)
	r.Post("/wallets", walletHandler.CreateWallet)
	r.Delete("/wallets/{id}", walletHandler.DeleteWallet)

	r.Get("/depots", depotHandler.GetDepots)
	r.Post("/depots", depotHandler.CreateDepot)
	r.Delete("/depots/{id}", depotHandler.DeleteDepot)

	r.Get("/transactions", transactionHandler.GetTransactions)
	r.Post("/transactions", transactionHandler.CreateTransaction)
	r.Delete("/transactions/{id}", transactionHandler.DeleteTransaction)

	r.Get("/stocks", stockHandler.GetStocks)
	r.Post("/stocks", stockHandler.CreateStock)
	r.Delete("/stocks/{id}", stockHandler.DeleteStock)

	return r
}

func determineEnvironment() string {
	env := os.Getenv("APP_ENV")
	if env == EnvProduction {
		return EnvProduction
	}
	return EnvDemo
}

func getRepo(env string) (ports.ExpenseRepository, *sql.DB) {
	if env == EnvProduction {
		return setupPostgresDB()
	}
	return memory.NewRepository(), nil
}

func setupPostgresDB() (ports.ExpenseRepository, *sql.DB) {
	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		log.Fatal("DATABASE_URL environment variable is not set for production mode")
	}

	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("Database unreachable: %v", err)
	}
	fmt.Println("Connected to Postgres DB")
	return postgres.NewRepository(db), db
}

func getPort() string {
	return DefaultPort
}
