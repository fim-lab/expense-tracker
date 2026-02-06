package memory

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
	"golang.org/x/crypto/bcrypt"
)

type inMemoryRepositories struct {
	mu           sync.RWMutex
	transactions map[int]domain.Transaction
	budgets      map[int]domain.Budget
	wallets      map[int]domain.Wallet
	users        map[string]domain.User
	sessions     map[string]domain.Session
	depots       map[int]domain.Depot
	stocks       map[int]domain.Stock
	lastID       int
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
		transactions: make(map[int]domain.Transaction),
		budgets:      make(map[int]domain.Budget),
		wallets:      make(map[int]domain.Wallet),
		users:        make(map[string]domain.User),
		sessions:     make(map[string]domain.Session),
		depots:       make(map[int]domain.Depot),
		stocks:       make(map[int]domain.Stock),
		lastID:       0,
	}
}

func (r *inMemoryRepositories) nextID() int {
	r.lastID++
	return r.lastID
}

func (r *inMemoryRepositories) seed() {
	const AMOUNT_OF_SEEDED_TX = 21
	// SEED DATA
	// Username: demo | Password: demo | Salary: 100€
	demoUsername := "demo"
	// "Demo Budget" | 5€ Limit
	// "Demo Cash Wallet"
	hash, _ := bcrypt.GenerateFromPassword([]byte(demoUsername), bcrypt.DefaultCost)

	userRepo := r.UserRepository()
	budgetRepo := r.BudgetRepository()
	walletRepo := r.WalletRepository()
	transactionRepo := r.TransactionRepository()

	userRepo.SaveUser(domain.User{
		Username:     demoUsername,
		PasswordHash: string(hash),
		SalaryCents:  10000,
	})
	demoUser, err := userRepo.GetUserByUsername(demoUsername)
	if err != nil {
		log.Fatal("Could not initiate demo User", err)
	}

	budgetRepo.SaveBudget(domain.Budget{
		UserID:     demoUser.ID,
		Name:       "Demo Budget",
		LimitCents: 500,
	})
	budgets, err := budgetRepo.FindBudgetsByUser(demoUser.ID)
	if err != nil {
		log.Fatal("Could not initiate demo Budget", err)
	}
	demoBudget := budgets[0]
	walletRepo.SaveWallet(domain.Wallet{
		UserID: demoUser.ID,
		Name:   "Demo Cash Wallet",
	})
	wallets, err := walletRepo.FindWalletsByUser(demoUser.ID)
	if err != nil {
		log.Fatal("Could not initiate demo Wallet", err)
	}
	demoWallet := wallets[0]
	for i := 0; i < AMOUNT_OF_SEEDED_TX; i++ {
		transactionRepo.SaveTransaction(domain.Transaction{
			UserID:        demoUser.ID,
			Date:          time.Now().AddDate(0, 0, -i),
			BudgetID:      &demoBudget.ID,
			WalletID:      demoWallet.ID,
			Description:   fmt.Sprintf("Transaction%v%v", i%2, i%3),
			AmountInCents: 104 * i,
			Type: func() domain.TransactionType {
				if i%4 == 0 {
					return domain.Expense
				}
				return domain.Income
			}(),
		})
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
