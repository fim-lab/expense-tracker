package memory

import (
	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

type TradeRepository struct {
	repo *inMemoryRepositories
}

func (r *TradeRepository) SaveTrade(t domain.Trade) (int, error) {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	if t.ID == 0 {
		t.ID = r.repo.nextID()
	}
	r.repo.trades[t.ID] = t
	return t.ID, nil
}

func (r *TradeRepository) GetTradeByID(id int) (domain.Trade, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()
	t, ok := r.repo.trades[id]
	if !ok {
		return domain.Trade{}, domain.ErrTradeNotFound
	}
	return t, nil
}

func (r *TradeRepository) FindTradesByDepot(depotID int) ([]domain.Trade, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()
	var results []domain.Trade
	for _, t := range r.repo.trades {
		if t.DepotID == depotID {
			results = append(results, t)
		}
	}
	return results, nil
}

func (r *TradeRepository) CountTradesByDepot(depotID int) (int, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()
	count := 0
	for _, t := range r.repo.trades {
		if t.DepotID == depotID {
			count++
		}
	}
	return count, nil
}

func (r *TradeRepository) UpdateTrade(t domain.Trade) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	if _, ok := r.repo.trades[t.ID]; !ok {
		return domain.ErrTradeNotFound
	}
	r.repo.trades[t.ID] = t
	return nil
}

func (r *TradeRepository) DeleteTrade(id int) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	delete(r.repo.trades, id)
	return nil
}

func (r *TradeRepository) DeleteAllTradesOfDepot(depotID int) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	for id, t := range r.repo.trades {
		if t.DepotID == depotID {
			delete(r.repo.trades, id)
		}
	}
	return nil
}

func (r *TradeRepository) DeleteAllByUser(userID int) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()

	// Find all depots of the user
	userDepots := make(map[int]bool)
	for id, d := range r.repo.depots {
		if d.UserID == userID {
			userDepots[id] = true
		}
	}

	// Delete all trades belonging to those depots
	for id, t := range r.repo.trades {
		if userDepots[t.DepotID] {
			delete(r.repo.trades, id)
		}
	}
	return nil
}
