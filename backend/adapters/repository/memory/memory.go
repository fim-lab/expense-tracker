package memory

import (
	"sync"

	"github.com/fim-lab/expense-tracker/backend/internal/core/domain"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type Repository struct {
	mu           sync.RWMutex
	transactions map[uuid.UUID]domain.Transaction
	budgets      map[uuid.UUID]domain.Budget
	users        map[string]domain.User
	sessions     map[string]domain.Session
}

func NewRepository() *Repository {
	repo := &Repository{
		transactions: make(map[uuid.UUID]domain.Transaction),
		budgets:      make(map[uuid.UUID]domain.Budget),
		users:        make(map[string]domain.User),
		sessions:     make(map[string]domain.Session),
	}

	// SEED DATA
	// Username: demo | Password: demo
	// "Demo Budget" | 5€ Limit
	hash, _ := bcrypt.GenerateFromPassword([]byte("demo"), bcrypt.DefaultCost)
	demoUserId := 23
	demoBudgetId := uuid.New()
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
	r.transactions[t.ID] = t
	return nil
}

func (r *Repository) GetTransactionByID(id uuid.UUID) (domain.Transaction, error) {
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

func (r *Repository) DeleteTransaction(id uuid.UUID) error {
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

func (r *Repository) GetBudgetByID(id uuid.UUID) (domain.Budget, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	b, ok := r.budgets[id]
	if !ok {
		return domain.Budget{}, domain.ErrMissingBudget
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

func (r *Repository) DeleteBudget(id uuid.UUID) error {
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
