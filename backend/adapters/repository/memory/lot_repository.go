package memory

import (
	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

type LotRepository struct {
	repo *inMemoryRepositories
}

func (r *LotRepository) SaveLot(s domain.Lot) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	if s.ID == 0 {
		s.ID = r.repo.nextID()
	}
	r.repo.lots[s.ID] = s
	return nil
}

func (r *LotRepository) GetLotByID(id int) (domain.Lot, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()
	s, ok := r.repo.lots[id]
	if !ok {
		return domain.Lot{}, domain.ErrLotNotFound
	}
	return s, nil
}

func (r *LotRepository) FindLotsByDepot(depotID int) ([]domain.Lot, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()
	var results []domain.Lot
	for _, s := range r.repo.lots {
		if s.DepotID == depotID {
			results = append(results, s)
		}
	}
	return results, nil
}

func (r *LotRepository) DeleteLot(id int) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	delete(r.repo.lots, id)
	return nil
}

func (r *LotRepository) DeleteAllLotsOfDepot(depotID int) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	for id, s := range r.repo.lots {
		if s.DepotID == depotID {
			delete(r.repo.lots, id)
		}
	}
	return nil
}

func (r *LotRepository) DeleteAllByUser(userID int) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()

	userDepots := make(map[int]bool)
	for id, d := range r.repo.depots {
		if d.UserID == userID {
			userDepots[id] = true
		}
	}

	for id, l := range r.repo.lots {
		if userDepots[l.DepotID] {
			delete(r.repo.lots, id)
		}
	}
	return nil
}
