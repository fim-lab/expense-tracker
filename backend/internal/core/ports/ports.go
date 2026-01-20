package ports

import "github.com/fim-lab/expense-tracker/internal/core/domain"

// --- Driving Ports ---
type TransactionService interface {
	CreateTransaction(userID int, t domain.Transaction) error
	CreateTransfer(userID, fromWalletID, toWalletID, amount int) error
	GetTransactions(userID int, limit int, offset int) ([]domain.TransactionDTO, error)
	Search(userID int, criteria domain.TransactionSearchCriteria) (*domain.PaginatedTransactions, error)
	GetTransactionCount(userID int) (int, error)
	UpdateTransaction(userID int, t domain.Transaction) error
	DeleteTransaction(userID int, id int) error
	GetTransactionByID(userID int, id int) (domain.Transaction, error)
}

type BudgetService interface {
	CreateBudget(userID int, b domain.Budget) error
	GetBudgets(userID int) ([]domain.Budget, error)
	GetTotalOfBudgets(userID int) (int, error)
	DeleteBudget(userID int, id int) error
}

type WalletService interface {
	CreateWallet(userID int, w domain.Wallet) error
	GetWallets(userID int) ([]domain.Wallet, error)
	GetTotalOfWallets(userID int) (int, error)
	DeleteWallet(userID int, id int) error
}

type UserService interface {
	Authenticate(username, password string) (domain.User, error)
}

type SessionService interface {
	CreateSession(session domain.Session) error
	ValidateSession(token string) (bool, int)
	DeleteSession(sessionID string) error
}

type DepotService interface {
	CreateDepot(userID int, d domain.Depot) error
	GetDepots(userID int) ([]domain.Depot, error)
	DeleteDepot(userID int, id int) error
}

type StockService interface {
	CreateStock(userID int, s domain.Stock) error
	GetStocks(userID int) ([]domain.Stock, error)
	DeleteStock(userID int, id int) error
}

// --- Driven Ports  ---
type ExpenseRepository interface {
	GetUserByUsername(username string) (domain.User, error)
	SaveUser(u domain.User) error

	SaveSession(s domain.Session) error
	GetSessionByToken(token string) (domain.Session, error)
	DeleteSession(sessionId string) error

	SaveBudget(b domain.Budget) error
	GetBudgetByID(id int) (domain.Budget, error)
	FindBudgetsByUser(userID int) ([]domain.Budget, error)
	DeleteBudget(id int) error

	SaveWallet(w domain.Wallet) error
	GetWalletByID(id int) (domain.Wallet, error)
	FindWalletsByUser(userID int) ([]domain.Wallet, error)
	DeleteWallet(id int) error

	SaveDepot(d domain.Depot) error
	GetDepotByID(id int) (domain.Depot, error)
	FindDepotsByUser(userID int) ([]domain.Depot, error)
	DeleteDepot(id int) error

	SaveTransaction(t domain.Transaction) error
	GetTransactionByID(id int) (domain.Transaction, error)
	GetTransactionCount(userId int) (int, error)
	FindTransactionsByUser(userID int, limit int, offset int) ([]domain.TransactionDTO, error)
	SearchTransactions(userID int, criteria domain.TransactionSearchCriteria) ([]domain.TransactionDTO, error)
	CountSearchedTransactions(userID int, criteria domain.TransactionSearchCriteria) (int, error)
	UpdateTransaction(t domain.Transaction) error
	DeleteTransaction(id int) error
	CreateTransfer(from, to domain.Transaction) error

	SaveStock(s domain.Stock) error
	GetStockByID(id int) (domain.Stock, error)
	FindStocksByUser(userID int) ([]domain.Stock, error)
	DeleteStock(id int) error
}
