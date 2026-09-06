package memory

import (
	"sort"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

type TemplateGroupRepository struct {
	repo *inMemoryRepositories
}

func (r *TemplateGroupRepository) SaveTemplateGroup(g domain.TemplateGroup) (int, error) {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	if g.ID == 0 {
		g.ID = r.repo.nextID()
	}
	r.repo.templateGroups[g.ID] = g
	return g.ID, nil
}

func (r *TemplateGroupRepository) FindTemplateGroupsByUser(userID int) ([]domain.TemplateGroup, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()
	var res []domain.TemplateGroup
	for _, g := range r.repo.templateGroups {
		if g.UserID == userID {
			res = append(res, g)
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].ID < res[j].ID
	})
	return res, nil
}

func (r *TemplateGroupRepository) UpdateTemplateGroup(g domain.TemplateGroup) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	existing, ok := r.repo.templateGroups[g.ID]
	if !ok {
		return domain.ErrTemplateGroupNotFound
	}
	existing.Name = g.Name
	r.repo.templateGroups[g.ID] = existing
	return nil
}

func (r *TemplateGroupRepository) DeleteTemplateGroup(id int) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	for ttID, tt := range r.repo.transactionTemplates {
		if tt.GroupID != nil && *tt.GroupID == id {
			tt.GroupID = nil
			r.repo.transactionTemplates[ttID] = tt
		}
	}
	delete(r.repo.templateGroups, id)
	return nil
}

func (r *TemplateGroupRepository) DeleteAllByUser(userID int) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	for id, g := range r.repo.templateGroups {
		if g.UserID == userID {
			delete(r.repo.templateGroups, id)
		}
	}
	return nil
}
