package memory

import (
	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

type StockRepository struct {
	repo *inMemoryRepositories
}

func (r *StockRepository) FindAllStocks() ([]domain.Stock, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()
	var res []domain.Stock
	for _, s := range r.repo.stocks {
		res = append(res, s)
	}
	return res, nil
}

func (r *StockRepository) GetStockByID(id int) (domain.Stock, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()
	s, ok := r.repo.stocks[id]
	if !ok {
		return domain.Stock{}, domain.ErrStockNotFound
	}
	return s, nil
}

func (r *StockRepository) FindStockByWKN(wkn string) (domain.Stock, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()
	for _, s := range r.repo.stocks {
		if s.WKN == wkn {
			return s, nil
		}
	}
	return domain.Stock{}, domain.ErrStockNotFound
}

func (r *StockRepository) SaveStock(s domain.Stock) (int, error) {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	if s.ID == 0 {
		s.ID = r.repo.nextID()
	}
	r.repo.stocks[s.ID] = s
	return s.ID, nil
}

func (r *StockRepository) UpdateStock(s domain.Stock) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	if _, ok := r.repo.stocks[s.ID]; !ok {
		return domain.ErrStockNotFound
	}
	r.repo.stocks[s.ID] = s
	return nil
}

func (r *StockRepository) DeleteStock(id int) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	delete(r.repo.stocks, id)
	return nil
}
