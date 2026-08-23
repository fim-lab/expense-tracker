package memory

import (
	"sort"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

type BudgetGroupRepository struct {
	repo *inMemoryRepositories
}

func (r *BudgetGroupRepository) SaveBudgetGroup(g domain.BudgetGroup) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	if g.ID == 0 {
		g.ID = r.repo.nextID()
	}
	r.repo.budgetGroups[g.ID] = g
	return nil
}

func (r *BudgetGroupRepository) FindBudgetGroupsByUser(userID int) ([]domain.BudgetGroup, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()
	var res []domain.BudgetGroup
	for _, g := range r.repo.budgetGroups {
		if g.UserID == userID {
			res = append(res, g)
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].ID < res[j].ID
	})
	return res, nil
}

func (r *BudgetGroupRepository) UpdateBudgetGroup(g domain.BudgetGroup) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	existing, ok := r.repo.budgetGroups[g.ID]
	if !ok {
		return domain.ErrBudgetGroupNotFound
	}
	existing.Name = g.Name
	r.repo.budgetGroups[g.ID] = existing
	return nil
}

func (r *BudgetGroupRepository) DeleteBudgetGroup(id int) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	for budgetID, b := range r.repo.budgets {
		if b.GroupID != nil && *b.GroupID == id {
			b.GroupID = nil
			r.repo.budgets[budgetID] = b
		}
	}
	delete(r.repo.budgetGroups, id)
	return nil
}

func (r *BudgetGroupRepository) DeleteAllByUser(userID int) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	for id, g := range r.repo.budgetGroups {
		if g.UserID == userID {
			delete(r.repo.budgetGroups, id)
		}
	}
	return nil
}
