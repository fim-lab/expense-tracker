package memory

import (
	"log"
	"sync"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
	"golang.org/x/crypto/bcrypt"
)

type inMemoryRepositories struct {
	mu                   sync.RWMutex
	transactions         map[int]domain.Transaction
	budgets              map[int]domain.Budget
	wallets              map[int]domain.Wallet
	users                map[string]domain.User
	sessions             map[string]domain.Session
	depots               map[int]domain.Depot
	stocks               map[int]domain.Stock
	transactionTemplates map[int]domain.TransactionTemplate
	lastID               int
}

func NewSeededRepositories() ports.Repositories {
	repo := NewInMemoryRepositories()
	repo.seed()
	return repo
}

func NewCleanRepositories() ports.Repositories {
	return NewInMemoryRepositories()
}

func NewInMemoryRepositories() *inMemoryRepositories {
	return &inMemoryRepositories{
		transactions:         make(map[int]domain.Transaction),
		budgets:              make(map[int]domain.Budget),
		wallets:              make(map[int]domain.Wallet),
		users:                make(map[string]domain.User),
		sessions:             make(map[string]domain.Session),
		depots:               make(map[int]domain.Depot),
		stocks:               make(map[int]domain.Stock),
		transactionTemplates: make(map[int]domain.TransactionTemplate),
		lastID:               0,
	}
}

func (r *inMemoryRepositories) nextID() int {
	r.lastID++
	return r.lastID
}

func (r *inMemoryRepositories) seed() {
	// Username: demo | Password: demo | Salary: 100€
	demoUsername := "demo"
	hash, _ := bcrypt.GenerateFromPassword([]byte(demoUsername), bcrypt.DefaultCost)

	userRepo := r.UserRepository()

	userRepo.SaveUser(domain.User{
		Username:     demoUsername,
		PasswordHash: string(hash),
		SalaryCents:  10000,
	})
	_, err := userRepo.GetUserByUsername(demoUsername)
	if err != nil {
		log.Fatal("Could not initiate demo User", err)
	}
}

func (r *inMemoryRepositories) UserRepository() ports.UserRepository {
	return &UserRepository{repo: r}
}

func (r *inMemoryRepositories) SessionRepository() ports.SessionRepository {
	return &SessionRepository{repo: r}
}

func (r *inMemoryRepositories) BudgetRepository() ports.BudgetRepository {
	return &BudgetRepository{repo: r}
}

func (r *inMemoryRepositories) WalletRepository() ports.WalletRepository {
	return &WalletRepository{repo: r}
}

func (r *inMemoryRepositories) DepotRepository() ports.DepotRepository {
	return &DepotRepository{repo: r}
}

func (r *inMemoryRepositories) StockRepository() ports.StockRepository {
	return &StockRepository{repo: r}
}

func (r *inMemoryRepositories) TransactionRepository() ports.TransactionRepository {
	return &TransactionRepository{repo: r}
}

func (r *inMemoryRepositories) TransactionTemplateRepository() ports.TransactionTemplateRepository {
	return &TransactionTemplateRepository{repo: r}
}
