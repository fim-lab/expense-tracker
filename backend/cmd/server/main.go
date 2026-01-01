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

	coreService := services.NewExpenseService(repo)

	httpHandler := httpadapter.NewAdapter(coreService)
	
	staticFiles, err := fs.Sub(expensetracker.StaticAssets, "frontend")
	if err != nil {
		log.Fatal("Failed to sub into frontend folder:", err)
	}

	router := httpadapter.NewRouter(httpHandler, staticFiles)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting Hexagonal Server on port %s", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
