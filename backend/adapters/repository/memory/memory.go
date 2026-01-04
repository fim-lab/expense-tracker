package memory

import (
	"sync"

	"github.com/fim-lab/expense-tracker/backend/internal/core/domain"
	"golang.org/x/crypto/bcrypt"
)

type Repository struct {
	mu           sync.RWMutex
	transactions map[int]domain.Transaction
	budgets      map[int]domain.Budget
	wallets      map[int]domain.Wallet
	users        map[string]domain.User
	sessions     map[string]domain.Session
}

func NewRepository() *Repository {
	repo := &Repository{
		transactions: make(map[int]domain.Transaction),
		budgets:      make(map[int]domain.Budget),
		wallets:      make(map[int]domain.Wallet),
		users:        make(map[string]domain.User),
		sessions:     make(map[string]domain.Session),
	}

	// SEED DATA
	// Username: demo | Password: demo
	// "Demo Budget" | 5€ Limit
	// "Demo Cash Wallet"
	hash, _ := bcrypt.GenerateFromPassword([]byte("demo"), bcrypt.DefaultCost)
	demoUserId := 0
	demoBudgetId := 2
	demoWalletId := 3
	repo.users["demo"] = domain.User{
		ID:           demoUserId,
		Username:     "demo",
		PasswordHash: string(hash),
	}
	repo.budgets[demoBudgetId] = domain.Budget{
		ID:         demoBudgetId,
		UserID:     demoUserId,
		Name:       "Demo Budget",
		LimitCents: 500,
	}
	repo.wallets[demoWalletId] = domain.Wallet{
		ID:     demoWalletId,
		UserID: demoUserId,
		Name:   "Demo Cash Wallet",
	}

	return repo
}

// --- User Methods ---

func (r *Repository) GetUserByUsername(username string) (domain.User, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	user, ok := r.users[username]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return user, nil
}

func (r *Repository) SaveUser(u domain.User) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.users[u.Username] = u
	return nil
}

// --- Transaction Methods ---

func (r *Repository) SaveTransaction(t domain.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	t.ID = len(r.transactions)
	r.transactions[t.ID] = t
	return nil
}

func (r *Repository) GetTransactionByID(id int) (domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	t, ok := r.transactions[id]
	if !ok {
		return domain.Transaction{}, domain.ErrTransactionNotFound
	}
	return t, nil
}

func (r *Repository) FindTransactionsByUser(userID int) ([]domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []domain.Transaction
	for _, t := range r.transactions {
		if t.UserID == userID {
			res = append(res, t)
		}
	}
	return res, nil
}

func (r *Repository) DeleteTransaction(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.transactions, id)
	return nil
}

// --- Budget Methods ---

func (r *Repository) SaveBudget(b domain.Budget) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.budgets[b.ID] = b
	return nil
}

func (r *Repository) GetBudgetByID(id int) (domain.Budget, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.budgets[id]
	if !ok {
		return domain.Budget{}, domain.ErrBudgetNotFound
	}
	return b, nil
}

func (r *Repository) FindBudgetsByUser(userID int) ([]domain.Budget, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []domain.Budget
	for _, b := range r.budgets {
		if b.UserID == userID {
			res = append(res, b)
		}
	}
	return res, nil
}

func (r *Repository) DeleteBudget(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.budgets, id)
	return nil
}

func (r *Repository) SaveSession(session domain.Session) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.sessions[session.SessionToken] = session
	return nil
}

func (r *Repository) GetSessionByToken(token string) (domain.Session, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.sessions[token]
	if !ok {
		return domain.Session{}, domain.ErrSessionNotFound
	}
	return s, nil
}

func (r *Repository) DeleteSession(sessionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.sessions, sessionID)
	return nil
}

// --- Wallet Methods ---

func (r *Repository) SaveWallet(w domain.Wallet) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.wallets[w.ID] = w
	return nil
}

func (r *Repository) GetWalletByID(id int) (domain.Wallet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	w, ok := r.wallets[id]
	if !ok {
		return domain.Wallet{}, domain.ErrWalletNotFound
	}
	return w, nil
}

func (r *Repository) FindWalletsByUser(userID int) ([]domain.Wallet, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var userWallets []domain.Wallet
	for _, w := range r.wallets {
		if w.UserID == userID {
			var balance int
			for _, t := range r.transactions {
				if t.WalletID == w.ID {
					if t.Type == domain.Income {
						balance += t.AmountInCents
					} else {
						balance -= t.AmountInCents
					}
				}
			}
			w.Balance = balance
			userWallets = append(userWallets, w)
		}
	}
	return userWallets, nil
}

func (r *Repository) DeleteWallet(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.wallets, id)
	return nil
}
