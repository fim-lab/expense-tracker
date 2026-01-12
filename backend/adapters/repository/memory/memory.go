package memory

import (
	"log"
	"sort"
	"sync"
	"time"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"golang.org/x/crypto/bcrypt"
)

type Repository struct {
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

func NewRepository() *Repository {
	repo := &Repository{
		transactions: make(map[int]domain.Transaction),
		budgets:      make(map[int]domain.Budget),
		wallets:      make(map[int]domain.Wallet),
		users:        make(map[string]domain.User),
		sessions:     make(map[string]domain.Session),
		depots:       make(map[int]domain.Depot),
		stocks:       make(map[int]domain.Stock),
		lastID:       0,
	}

	// SEED DATA
	// Username: demo | Password: demo
	demoUsername := "demo"
	// "Demo Budget" | 5€ Limit
	// "Demo Cash Wallet"
	hash, _ := bcrypt.GenerateFromPassword([]byte(demoUsername), bcrypt.DefaultCost)
	repo.SaveUser(domain.User{
		Username:     demoUsername,
		PasswordHash: string(hash),
	})
	demoUser, err := repo.GetUserByUsername(demoUsername)
	if err != nil {
		log.Fatal("Could not initiate demo User", err)
	}

	repo.SaveBudget(domain.Budget{
		UserID:     demoUser.ID,
		Name:       "Demo Budget",
		LimitCents: 500,
	})
	budgets, err := repo.FindBudgetsByUser(demoUser.ID)
	if err != nil {
		log.Fatal("Could not initiate demo Budget", err)
	}
	demoBudget := budgets[0]
	repo.SaveWallet(domain.Wallet{
		UserID: demoUser.ID,
		Name:   "Demo Cash Wallet",
	})
	wallets, err := repo.FindWalletsByUser(demoUser.ID)
	if err != nil {
		log.Fatal("Could not initiate demo Wallet", err)
	}
	demoWallet := wallets[0]
	repo.SaveTransaction(domain.Transaction{
		UserID:        demoUser.ID,
		Date:          time.Now(),
		BudgetID:      demoBudget.ID,
		WalletID:      demoWallet.ID,
		Description:   "Test Transaction",
		AmountInCents: 500,
		Type:          domain.Expense,
	})

	return repo
}

func (r *Repository) nextID() int {
	r.lastID++
	return r.lastID
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
	if u.ID == 0 {
		u.ID = r.nextID()
	}
	r.users[u.Username] = u
	return nil
}

// --- Transaction Methods ---

func (r *Repository) SaveTransaction(t domain.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if t.ID == 0 {
		t.ID = r.nextID()
	}
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

func (r *Repository) FindTransactionsByUser(userID int, limit int, offset int) ([]domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []domain.Transaction
	for _, t := range r.transactions {
		if t.UserID == userID {
			res = append(res, t)
		}
	}

	sort.Slice(res, func(i, j int) bool {
		if res[i].Date.Equal(res[j].Date) {
			return res[i].ID > res[j].ID
		}
		return res[i].Date.After(res[j].Date)
	})

	start := offset
	if start >= len(res) {
		return []domain.Transaction{}, nil
	}

	end := offset + limit
	if end > len(res) || limit <= 0 {
		end = len(res)
	}

	return res[start:end], nil
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
	if b.ID == 0 {
		b.ID = r.nextID()
	}
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
	if w.ID == 0 {
		w.ID = r.nextID()
	}
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

// --- Wallet Methods ---

func (r *Repository) SaveDepot(d domain.Depot) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if d.ID == 0 {
		d.ID = r.nextID()
	}
	r.depots[d.ID] = d
	return nil
}

func (r *Repository) GetDepotByID(id int) (domain.Depot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	d, ok := r.depots[id]
	if !ok {
		return domain.Depot{}, domain.ErrMissingDepot
	}
	return d, nil
}

func (r *Repository) FindDepotsByUser(userID int) ([]domain.Depot, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var results []domain.Depot
	for _, d := range r.depots {
		if d.UserID == userID {
			results = append(results, d)
		}
	}
	return results, nil
}

func (r *Repository) DeleteDepot(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.depots, id)
	return nil
}

// --- Stock Methods ---

func (r *Repository) SaveStock(s domain.Stock) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if s.ID == 0 {
		s.ID = r.nextID()
	}
	r.stocks[s.ID] = s
	return nil
}

func (r *Repository) GetStockByID(id int) (domain.Stock, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.stocks[id]
	if !ok {
		return domain.Stock{}, domain.ErrStockNotFound
	}
	return s, nil
}

func (r *Repository) FindStocksByUser(userID int) ([]domain.Stock, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var results []domain.Stock
	for _, s := range r.stocks {
		if s.UserID == userID {
			results = append(results, s)
		}
	}
	return results, nil
}

func (r *Repository) DeleteStock(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.stocks, id)
	return nil
}
