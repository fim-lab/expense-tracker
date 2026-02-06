package postgres

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type postgresRepositoryCollection struct {
	userRepo        *UserRepository
	sessionRepo     *SessionRepository
	budgetRepo      *BudgetRepository
	walletRepo      *WalletRepository
	depotRepo       *DepotRepository
	transactionRepo *TransactionRepository
	stockRepo       *StockRepository
}

func NewPostgresRepositoryCollection() (*sql.DB, ports.Repositories) {
	db := setupPostgresDB()
	return db, &postgresRepositoryCollection{
		userRepo:        NewUserRepository(db),
		sessionRepo:     NewSessionRepository(db),
		budgetRepo:      NewBudgetRepository(db),
		walletRepo:      NewWalletRepository(db),
		depotRepo:       NewDepotRepository(db),
		transactionRepo: NewTransactionRepository(db),
		stockRepo:       NewStockRepository(db),
	}
}

func setupPostgresDB() *sql.DB {
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
	return db
}

func (prc *postgresRepositoryCollection) UserRepository() ports.UserRepository {
	return prc.userRepo
}

func (prc *postgresRepositoryCollection) SessionRepository() ports.SessionRepository {
	return prc.sessionRepo
}

func (prc *postgresRepositoryCollection) BudgetRepository() ports.BudgetRepository {
	return prc.budgetRepo
}

func (prc *postgresRepositoryCollection) WalletRepository() ports.WalletRepository {
	return prc.walletRepo
}

func (prc *postgresRepositoryCollection) DepotRepository() ports.DepotRepository {
	return prc.depotRepo
}

func (prc *postgresRepositoryCollection) TransactionRepository() ports.TransactionRepository {
	return prc.transactionRepo
}

func (prc *postgresRepositoryCollection) StockRepository() ports.StockRepository {
	return prc.stockRepo
}
