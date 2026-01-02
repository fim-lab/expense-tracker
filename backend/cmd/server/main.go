package main

import (
	"database/sql"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	expensetracker "github.com/fim-lab/expense-tracker"
	httpadapter "github.com/fim-lab/expense-tracker/backend/adapters/handler/http"
	"github.com/fim-lab/expense-tracker/backend/adapters/handler/middleware"
	"github.com/fim-lab/expense-tracker/backend/adapters/repository/memory"
	"github.com/fim-lab/expense-tracker/backend/adapters/repository/postgres"
	"github.com/fim-lab/expense-tracker/backend/internal/core/ports"
	"github.com/fim-lab/expense-tracker/backend/internal/core/services"
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
	defer db.Close()

	transService, budgetService, userService, sessionService := initializeServices(repo)

	staticFiles, err := fs.Sub(expensetracker.StaticAssets, "frontend")
	if err != nil {
		log.Fatal("Failed to sub into frontend folder:", err)
	}

	mainMux := http.NewServeMux()
	setupAuthRoutes(mainMux, userService, sessionService)
	setupApiRoutes(env, mainMux, transService, budgetService, sessionService)
	mainMux.Handle("/", http.FileServer(http.FS(staticFiles)))

	port := getPort()

	log.Printf("Start Server on port %s", port)
	if err := http.ListenAndServe(":"+port, mainMux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func determineEnvironment() string {
	env := os.Getenv("APP_ENV")
	if env != "" {
		return env
	}

	if os.Getenv("DATABASE_URL") != "" {
		return EnvProduction
	}
	return EnvDemo
}

func getRepo(env string) (ports.ExpenseRepository, *sql.DB) {
	switch env {
	case EnvDemo:
		fmt.Println("Running in DEMO mode")
		return memory.NewRepository(), nil
	case EnvProduction:
		fmt.Println("Running in PRODUCTION mode")
		return setupPostgresDB()
	default:
		log.Fatalf("Unknown environment: %s", env)
		return nil, nil
	}
}

func setupPostgresDB() (ports.ExpenseRepository, *sql.DB) {
	dbUrl := os.Getenv("DATABASE_URL")
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

func initializeServices(repo ports.ExpenseRepository) (
	*ports.TransactionService, *ports.BudgetService, *ports.UserService, *ports.SessionService) {

	transService := services.NewTransactionService(repo)
	budgetService := services.NewBudgetService(repo)
	userService := services.NewUserService(repo)
	sessionService := services.NewSessionService(repo)

	return &transService, &budgetService, &userService, &sessionService
}

func setupAuthRoutes(mainMux *http.ServeMux, userService *ports.UserService, sessionService *ports.SessionService) {
	authHandler := httpadapter.NewAuthHandler(userService, sessionService)
	mainMux.HandleFunc("/auth/login", authHandler.Login)
	mainMux.HandleFunc("/auth/logout", authHandler.Logout)
}

func setupApiRoutes(env string, mainMux *http.ServeMux, transService *ports.TransactionService, budgetService *ports.BudgetService, sessionService *ports.SessionService) {
	apiRouter := http.NewServeMux()
	transactionHandler := httpadapter.NewTransactionHandler(transService)
	budgetHandler := httpadapter.NewBudgetHandler(budgetService)
	apiRouter.HandleFunc("/api/transactions", transactionHandler.Handle)
	apiRouter.HandleFunc("/api/budgets", budgetHandler.Handle)

	addMiddleware(env, mainMux, apiRouter, sessionService)
}

func addMiddleware(env string, mainMux *http.ServeMux, apiRouter *http.ServeMux, sessionService *ports.SessionService) {
	if env == EnvDemo {
		demoMidlleware := middleware.NewDemoMiddleware()
		mainMux.Handle("/api/", demoMidlleware.Handle(apiRouter))
	} else {
		authMiddleware := middleware.NewAuthMiddleware(sessionService)
		mainMux.Handle("/api/", authMiddleware.Handle(apiRouter))
	}
}

func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = DefaultPort
	}
	return port
}
