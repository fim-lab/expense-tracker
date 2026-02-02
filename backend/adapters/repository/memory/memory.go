package memory

import (
	"fmt"
	"log"
	"sort"
	"strings"
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

func NewSeededRepository() *Repository {
	return NewRepository(true)
}

func NewCleanRepository() *Repository {
	return NewRepository(false)
}

func NewRepository(initWithSeedData bool) *Repository {
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

	if initWithSeedData {
		const AMOUNT_OF_SEEDED_TX = 21
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
		for i := 0; i < AMOUNT_OF_SEEDED_TX; i++ {
			repo.SaveTransaction(domain.Transaction{
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
	r.transactions[t.ID] = t

	if t.BudgetID != nil {
		budget, ok := r.budgets[*t.BudgetID]
		if ok {
			adjustment := t.AmountInCents
			if t.Type == domain.Expense {
				adjustment = -t.AmountInCents
			}
			budget.BalanceCents += adjustment
			r.budgets[*t.BudgetID] = budget
		}
	}

	wallet, ok := r.wallets[t.WalletID]
	if ok {
		adjustment := t.AmountInCents
		if t.Type == domain.Expense {
			adjustment = -t.AmountInCents
		}
		wallet.BalanceCents += adjustment
		r.wallets[t.WalletID] = wallet
	}

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

func (r *Repository) GetTransactionCount(userID int) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var res []domain.Transaction
	for _, t := range r.transactions {
		if t.UserID == userID {
			res = append(res, t)
		}
	}
	return len(res), nil
}

func (r *Repository) FindTransactionsByUser(userID int, limit int, offset int) ([]domain.TransactionDTO, error) {
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
		return []domain.TransactionDTO{}, nil
	}

	end := offset + limit
	if end > len(res) || limit <= 0 {
		end = len(res)
	}

	paginatedTxs := res[start:end]

	dtos := make([]domain.TransactionDTO, 0, len(paginatedTxs))
	for _, t := range paginatedTxs {
		var budgetName string
		if t.BudgetID != nil {
			if budget, ok := r.budgets[*t.BudgetID]; ok {
				budgetName = budget.Name
			}
		}
		wallet := r.wallets[t.WalletID]
		dtos = append(dtos, domain.TransactionDTO{
			ID:            t.ID,
			Date:          t.Date,
			Description:   t.Description,
			AmountInCents: t.AmountInCents,
			Type:          t.Type,
			BudgetName:    budgetName,
			WalletName:    wallet.Name,
			IsPending:     t.IsPending,
		})
	}

	return dtos, nil
}

func (r *Repository) SearchTransactions(userID int, criteria domain.TransactionSearchCriteria) ([]domain.TransactionDTO, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var filtered []domain.Transaction
	for _, t := range r.transactions {
		if t.UserID != userID {
			continue
		}

		if criteria.SearchTerm != nil && *criteria.SearchTerm != "" {
			term := strings.ToLower(*criteria.SearchTerm)
			if !strings.Contains(strings.ToLower(t.Description), term) {
				continue
			}
		}

		if criteria.FromDate != nil && t.Date.Before(*criteria.FromDate) {
			continue
		}
		if criteria.UntilDate != nil && t.Date.After(*criteria.UntilDate) {
			continue
		}

		if criteria.BudgetID != nil && *t.BudgetID != *criteria.BudgetID {
			continue
		}

		if criteria.WalletID != nil && t.WalletID != *criteria.WalletID {
			continue
		}

		if criteria.Type != nil && t.Type != *criteria.Type {
			continue
		}

		filtered = append(filtered, t)
	}

	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Date.Equal(filtered[j].Date) {
			return filtered[i].ID > filtered[j].ID
		}
		return filtered[i].Date.After(filtered[j].Date)
	})

	start := (criteria.Page - 1) * criteria.PageSize
	if start >= len(filtered) {
		return []domain.TransactionDTO{}, nil
	}

	end := start + criteria.PageSize
	if end > len(filtered) {
		end = len(filtered)
	}

	paginatedTxs := filtered[start:end]

	dtos := make([]domain.TransactionDTO, 0, len(paginatedTxs))
	for _, t := range paginatedTxs {
		var budgetName string
		if t.BudgetID != nil {
			if budget, ok := r.budgets[*t.BudgetID]; ok {
				budgetName = budget.Name
			}
		}
		wallet := r.wallets[t.WalletID]
		dtos = append(dtos, domain.TransactionDTO{
			ID:            t.ID,
			Date:          t.Date,
			Description:   t.Description,
			AmountInCents: t.AmountInCents,
			Type:          t.Type,
			BudgetName:    budgetName,
			WalletName:    wallet.Name,
			IsPending:     t.IsPending,
		})
	}

	return dtos, nil
}

func (r *Repository) CountSearchedTransactions(userID int, criteria domain.TransactionSearchCriteria) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var count int
	for _, t := range r.transactions {
		if t.UserID != userID {
			continue
		}

		if criteria.SearchTerm != nil && *criteria.SearchTerm != "" {
			term := strings.ToLower(*criteria.SearchTerm)
			if !strings.Contains(strings.ToLower(t.Description), term) {
				continue
			}
		}

		if criteria.FromDate != nil && t.Date.Before(*criteria.FromDate) {
			continue
		}
		if criteria.UntilDate != nil && t.Date.After(*criteria.UntilDate) {
			continue
		}

		if criteria.BudgetID != nil && *t.BudgetID != *criteria.BudgetID {
			continue
		}

		if criteria.WalletID != nil && t.WalletID != *criteria.WalletID {
			continue
		}

		if criteria.Type != nil && t.Type != *criteria.Type {
			continue
		}

		count++
	}

	return count, nil
}

func (r *Repository) DeleteTransaction(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	tx, exists := r.transactions[id]
	if !exists {
		return domain.ErrTransactionNotFound
	}

	adjustment := -tx.AmountInCents
	if tx.Type == domain.Expense {
		adjustment = tx.AmountInCents
	}

	if tx.BudgetID != nil {
		budget, ok := r.budgets[*tx.BudgetID]
		if ok {
			budget.BalanceCents += adjustment
			r.budgets[*tx.BudgetID] = budget
		}
	}

	wallet, ok := r.wallets[tx.WalletID]
	if ok {
		wallet.BalanceCents += adjustment
		r.wallets[tx.WalletID] = wallet
	}

	delete(r.transactions, id)
	return nil
}

func (r *Repository) UpdateTransaction(t domain.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	oldT, ok := r.transactions[t.ID]
	if !ok {
		return domain.ErrTransactionNotFound
	}

	oldAdjustment := oldT.AmountInCents
	if oldT.Type == domain.Income {
		oldAdjustment = -oldT.AmountInCents
	}
	if oldT.BudgetID != nil {
		if budget, ok := r.budgets[*oldT.BudgetID]; ok {
			budget.BalanceCents += oldAdjustment
			r.budgets[*oldT.BudgetID] = budget
		}
	}
	if wallet, ok := r.wallets[oldT.WalletID]; ok {
		wallet.BalanceCents += oldAdjustment
		r.wallets[oldT.WalletID] = wallet
	}

	r.transactions[t.ID] = t

	newAdjustment := t.AmountInCents
	if t.Type == domain.Expense {
		newAdjustment = -t.AmountInCents
	}
	if t.BudgetID != nil {
		if budget, ok := r.budgets[*t.BudgetID]; ok {
			budget.BalanceCents += newAdjustment
			r.budgets[*t.BudgetID] = budget
		}
	}
	if wallet, ok := r.wallets[t.WalletID]; ok {
		wallet.BalanceCents += newAdjustment
		r.wallets[t.WalletID] = wallet
	}

	return nil
}

func (r *Repository) CreateTransfer(from, to domain.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if from.ID == 0 {
		from.ID = r.nextID()
	}
	r.transactions[from.ID] = from
	fromWallet, ok := r.wallets[from.WalletID]
	if !ok {
		return domain.ErrWalletNotFound
	}
	fromWallet.BalanceCents -= from.AmountInCents
	r.wallets[from.WalletID] = fromWallet

	if to.ID == 0 {
		to.ID = r.nextID()
	}
	r.transactions[to.ID] = to
	toWallet, ok := r.wallets[to.WalletID]
	if !ok {
		return domain.ErrWalletNotFound
	}
	toWallet.BalanceCents += to.AmountInCents
	r.wallets[to.WalletID] = toWallet

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
	sort.Slice(res, func(i, j int) bool {
		return res[i].ID < res[j].ID
	})
	return res, nil
}

func (r *Repository) DeleteBudget(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.budgets, id)
	return nil
}

func (r *Repository) UpdateBudget(b domain.Budget) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	existingBudget, ok := r.budgets[b.ID]
	if !ok {
		return domain.ErrBudgetNotFound
	}
	existingBudget.Name = b.Name
	existingBudget.LimitCents = b.LimitCents
	r.budgets[b.ID] = existingBudget
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
			userWallets = append(userWallets, w)
		}
	}
	sort.Slice(userWallets, func(i, j int) bool {
		return userWallets[i].ID < userWallets[j].ID
	})
	return userWallets, nil
}

func (r *Repository) DeleteWallet(id int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.wallets, id)
	return nil
}

func (r *Repository) UpdateWallet(w domain.Wallet) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	existingWallet, ok := r.wallets[w.ID]
	if !ok {
		return domain.ErrWalletNotFound
	}
	existingWallet.Name = w.Name
	r.wallets[w.ID] = existingWallet
	return nil
}

// --- Depot Methods ---

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

func (r *Repository) CountTransactionsByBudgetID(budgetID int) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, t := range r.transactions {
		if *t.BudgetID == budgetID {
			count++
		}
	}
	return count, nil
}

func (r *Repository) CountTransactionsByWalletID(walletID int) (int, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	count := 0
	for _, t := range r.transactions {
		if t.WalletID == walletID {
			count++
		}
	}
	return count, nil
}
