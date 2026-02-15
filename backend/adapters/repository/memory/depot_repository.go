package memory

import (
	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

type DepotRepository struct {
	repo *inMemoryRepositories
}

func (r *DepotRepository) SaveDepot(d domain.Depot) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	if d.ID == 0 {
		d.ID = r.repo.nextID()
	}
	r.repo.depots[d.ID] = d
	return nil
}

func (r *DepotRepository) GetDepotByID(id int) (domain.Depot, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()
	d, ok := r.repo.depots[id]
	if !ok {
		return domain.Depot{}, domain.ErrMissingDepot
	}
	return d, nil
}

func (r *DepotRepository) FindDepotsByUser(userID int) ([]domain.Depot, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()
	var results []domain.Depot
	for _, d := range r.repo.depots {
		if d.UserID == userID {
			results = append(results, d)
		}
	}
	return results, nil
}

func (r *DepotRepository) DeleteDepot(id int) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	delete(r.repo.depots, id)
	return nil
}
