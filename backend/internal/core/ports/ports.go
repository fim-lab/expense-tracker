package ports

import "github.com/fim-lab/expense-tracker/backend/internal/core/domain"

// --- Driving Ports ---
type TransactionService interface {
	CreateTransaction(userID int, t domain.Transaction) error
	GetTransactions(userID int) ([]domain.Transaction, error)
	DeleteTransaction(userID int, id int) error
}

type BudgetService interface {
	CreateBudget(userID int, b domain.Budget) error
	GetBudgets(userID int) ([]domain.Budget, error)
	DeleteBudget(userID int, id int) error
}

type WalletService interface {
	CreateWallet(userID int, w domain.Wallet) error
	GetWallets(userID int) ([]domain.Wallet, error)
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

// --- Driven Ports  ---
type ExpenseRepository interface {
	SaveTransaction(t domain.Transaction) error
	GetTransactionByID(id int) (domain.Transaction, error)
	FindTransactionsByUser(userID int) ([]domain.Transaction, error)
	DeleteTransaction(id int) error

	SaveBudget(b domain.Budget) error
	GetBudgetByID(id int) (domain.Budget, error)
	FindBudgetsByUser(userID int) ([]domain.Budget, error)
	DeleteBudget(id int) error

	SaveWallet(w domain.Wallet) error
	GetWalletByID(id int) (domain.Wallet, error)
	FindWalletsByUser(userID int) ([]domain.Wallet, error)
	DeleteWallet(id int) error

	GetUserByUsername(username string) (domain.User, error)
	SaveUser(u domain.User) error

	SaveSession(s domain.Session) error
	GetSessionByToken(token string) (domain.Session, error)
	DeleteSession(sessionId string) error
}
