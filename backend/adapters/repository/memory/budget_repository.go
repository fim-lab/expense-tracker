package memory

import (
	"sort"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

type BudgetRepository struct {
	repo *inMemoryRepositories
}

func (r *BudgetRepository) SaveBudget(b domain.Budget) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	if b.ID == 0 {
		b.ID = r.repo.nextID()
	}
	r.repo.budgets[b.ID] = b
	return nil
}

func (r *BudgetRepository) GetBudgetByID(id int) (domain.Budget, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()
	b, ok := r.repo.budgets[id]
	if !ok {
		return domain.Budget{}, domain.ErrBudgetNotFound
	}
	return b, nil
}

func (r *BudgetRepository) FindBudgetsByUser(userID int) ([]domain.Budget, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()
	var res []domain.Budget
	for _, b := range r.repo.budgets {
		if b.UserID == userID {
			res = append(res, b)
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].ID < res[j].ID
	})
	return res, nil
}

func (r *BudgetRepository) DeleteBudget(id int) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	delete(r.repo.budgets, id)
	return nil
}

func (r *BudgetRepository) UpdateBudget(b domain.Budget) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	existingBudget, ok := r.repo.budgets[b.ID]
	if !ok {
		return domain.ErrBudgetNotFound
	}
	existingBudget.Name = b.Name
	existingBudget.LimitCents = b.LimitCents
	r.repo.budgets[b.ID] = existingBudget
	return nil
}
