package memory

import (
	"sync"
	"github.com/fim-lab/expense-tracker/backend/internal/core/domain"
)

type Repository struct {
	mu           sync.RWMutex
	transactions map[string]domain.Transaction
}

func NewRepository() *Repository {
	return &Repository{
		transactions: make(map[string]domain.Transaction),
	}
}

func (r *Repository) Save(t domain.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.transactions[t.ID] = t
	return nil
}

func (r *Repository) FindAll() ([]domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	var txs []domain.Transaction
	for _, t := range r.transactions {
		txs = append(txs, t)
	}
	return txs, nil
}

func (r *Repository) Delete(id string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.transactions, id)
	return nil
}