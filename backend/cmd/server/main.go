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

func main() {
	var repo ports.ExpenseRepository

	dbURL := os.Getenv("DATABASE_URL")
	useMemory := os.Getenv("USE_MEMORY_DB") == "true" || dbURL == ""

	if useMemory {
		log.Println("Initializing In-Memory Database")
		repo = memory.NewRepository()
	} else {
		db, err := sql.Open("postgres", dbURL)
		if err != nil {
			log.Fatalf("Failed to connect to database: %v", err)
		}
		defer db.Close()

		if err := db.Ping(); err != nil {
			log.Fatalf("Database unreachable: %v", err)
		}
		fmt.Println("Connected to Postgres DB")
		repo = postgres.NewRepository(db)
	}

	transService := services.NewTransactionService(repo)
	budgetService := services.NewBudgetService(repo)
	userService := services.NewUserService(repo)
	sessionService := services.NewSessionService(repo)

	staticFiles, err := fs.Sub(expensetracker.StaticAssets, "frontend")
	if err != nil {
		log.Fatal("Failed to sub into frontend folder:", err)
	}

	mainMux := http.NewServeMux()

	authHandler := httpadapter.NewAuthHandler(userService, sessionService)
	mainMux.HandleFunc("/auth/login", authHandler.Login)
	mainMux.HandleFunc("/auth/logout", authHandler.Logout)

	apiRouter := http.NewServeMux()
	transactionHandler := httpadapter.NewTransactionHandler(transService)
	budgetHandler := httpadapter.NewBudgetHandler(budgetService)
	apiRouter.HandleFunc("/api/transactions", transactionHandler.Handle)
	apiRouter.HandleFunc("/api/budgets", budgetHandler.Handle)

	authMiddleware := middleware.NewAuthMiddleware(sessionService)
	mainMux.Handle("/api/", authMiddleware.Handle(apiRouter))

	mainMux.Handle("/", http.FileServer(http.FS(staticFiles)))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting Hexagonal Server on port %s", port)
	if err := http.ListenAndServe(":"+port, mainMux); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
