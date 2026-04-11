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
	GetBudget(userID int, id int) (domain.Budget, error)
	UpdateBudget(userID int, budget domain.Budget) error
	GetBudgets(userID int) ([]domain.Budget, error)
	GetTotalOfBudgets(userID int) (int, error)
	DeleteBudget(userID int, id int) error
}

type WalletService interface {
	CreateWallet(userID int, w domain.Wallet) error
	GetWallet(userID int, id int) (domain.Wallet, error)
	UpdateWallet(userID int, wallet domain.Wallet) error
	GetWallets(userID int) ([]domain.Wallet, error)
	GetTotalOfWallets(userID int) (int, error)
	DeleteWallet(userID int, id int) error
}

type UserService interface {
	Authenticate(username, password string) (domain.User, error)
	GetUserByID(userID int) (domain.User, error)
	UpdateSalary(userID int, salary int) error
}

type SessionService interface {
	CreateSession(session domain.Session) error
	ValidateSession(token string) (bool, int)
	DeleteSession(sessionID string) error
}

type DepotService interface {
	CreateDepot(userID int, d domain.Depot) error
	GetDepots(userID int) ([]domain.Depot, error)
	UpdateDepot(userID int, d domain.Depot) error
	DeleteDepot(userID int, id int) error
}

type LotService interface {
	CreateLot(userID int, l domain.Lot) error
	GetLots(userID int) ([]domain.Lot, error)
	DeleteLot(userID int, id int) error
}

type TradeService interface {
	CreateTrade(userID int, t domain.Trade) error
	GetTradesByDepot(userID int, depotID int) ([]domain.Trade, error)
	UpdateTrade(userID int, t domain.Trade) error
	DeleteTrade(userID int, id int) error
}

type TransactionTemplateService interface {
	CreateTransactionTemplate(userID int, tt domain.TransactionTemplate) error
	GetTransactionTemplate(userID int, id int) (domain.TransactionTemplate, error)
	GetTransactionTemplates(userID int) ([]domain.TransactionTemplate, error)
	UpdateTransactionTemplate(userID int, tt domain.TransactionTemplate) error
	DeleteTransactionTemplate(userID int, id int) error
}

type ImportService interface {
	ImportData(userID int, data domain.FullImportData) error
	ImportTestData(userID int) error
	DeleteAllUserData(userID int) error
}

// --- Driven Ports  ---

type UserRepository interface {
	GetUserByUsername(username string) (domain.User, error)
	GetUserByID(userID int) (domain.User, error)
	SaveUser(u domain.User) error
	UpdateUserSalary(userID int, salary int) error
}

type SessionRepository interface {
	SaveSession(s domain.Session) error
	GetSessionByToken(token string) (domain.Session, error)
	DeleteSession(sessionId string) error
}

type BudgetRepository interface {
	SaveBudget(b domain.Budget) error
	GetBudgetByID(id int) (domain.Budget, error)
	UpdateBudget(budget domain.Budget) error
	FindBudgetsByUser(userID int) ([]domain.Budget, error)
	DeleteBudget(id int) error
	DeleteAllByUser(userID int) error
}

type WalletRepository interface {
	SaveWallet(w domain.Wallet) error
	GetWalletByID(id int) (domain.Wallet, error)
	UpdateWallet(wallet domain.Wallet) error
	FindWalletsByUser(userID int) ([]domain.Wallet, error)
	DeleteWallet(id int) error
	DeleteAllByUser(userID int) error
}

type DepotRepository interface {
	SaveDepot(d domain.Depot) error
	GetDepotByID(id int) (domain.Depot, error)
	UpdateDepot(d domain.Depot) error
	FindDepotsByUser(userID int) ([]domain.Depot, error)
	DeleteDepot(id int) error
	DeleteAllByUser(userID int) error
}

type TransactionRepository interface {
	SaveTransaction(t domain.Transaction) error
	GetTransactionByID(id int) (domain.Transaction, error)
	GetTransactionCount(userId int) (int, error)
	FindTransactionsByUser(userID int, limit int, offset int) ([]domain.TransactionDTO, error)
	SearchTransactions(userID int, criteria domain.TransactionSearchCriteria) ([]domain.TransactionDTO, error)
	CountSearchedTransactions(userID int, criteria domain.TransactionSearchCriteria) (int, error)
	UpdateTransaction(t domain.Transaction) error
	DeleteTransaction(id int) error
	DeleteAllByUser(userID int) error
	CreateTransfer(from, to domain.Transaction) error
	CountTransactionsByBudgetID(budgetID int) (int, error)
	CountTransactionsByWalletID(walletID int) (int, error)
}

type LotRepository interface {
	SaveLot(l domain.Lot) error
	GetLotByID(id int) (domain.Lot, error)
	FindLotsByDepot(depotID int) ([]domain.Lot, error)
	DeleteLot(id int) error
	DeleteAllLotsOfDepot(depotID int) error
	DeleteAllByUser(userID int) error
}

type TradeRepository interface {
	SaveTrade(t domain.Trade) error
	GetTradeByID(id int) (domain.Trade, error)
	FindTradesByDepot(depotID int) ([]domain.Trade, error)
	UpdateTrade(t domain.Trade) error
	DeleteTrade(id int) error
	DeleteAllByUser(userID int) error
}

type TransactionTemplateRepository interface {
	SaveTransactionTemplate(tt domain.TransactionTemplate) error
	GetTransactionTemplateByID(id int) (domain.TransactionTemplate, error)
	FindTransactionTemplatesByUser(userID int) ([]domain.TransactionTemplate, error)
	UpdateTransactionTemplate(tt domain.TransactionTemplate) error
	DeleteTransactionTemplate(id int) error
	DeleteAllByUser(userID int) error
}

type Repositories interface {
	UserRepository() UserRepository
	SessionRepository() SessionRepository
	BudgetRepository() BudgetRepository
	WalletRepository() WalletRepository
	DepotRepository() DepotRepository
	TransactionRepository() TransactionRepository
	LotRepository() LotRepository
	TradeRepository() TradeRepository
	TransactionTemplateRepository() TransactionTemplateRepository
}
