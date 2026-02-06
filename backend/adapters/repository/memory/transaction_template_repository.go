package memory

import (
	"sort"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

type TransactionTemplateRepository struct {
	repo *inMemoryRepositories
}

func (r *TransactionTemplateRepository) SaveTransactionTemplate(tt domain.TransactionTemplate) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	if tt.ID == 0 {
		tt.ID = r.repo.nextID()
	}
	r.repo.transactionTemplates[tt.ID] = tt
	return nil
}

func (r *TransactionTemplateRepository) GetTransactionTemplateByID(id int) (domain.TransactionTemplate, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()
	tt, ok := r.repo.transactionTemplates[id]
	if !ok {
		return domain.TransactionTemplate{}, domain.ErrTransactionTemplateNotFound
	}
	return tt, nil
}

func (r *TransactionTemplateRepository) FindTransactionTemplatesByUser(userID int) ([]domain.TransactionTemplate, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()
	var res []domain.TransactionTemplate
	for _, tt := range r.repo.transactionTemplates {
		if tt.UserID == userID {
			res = append(res, tt)
		}
	}
	sort.Slice(res, func(i, j int) bool {
		return res[i].ID < res[j].ID
	})
	return res, nil
}

func (r *TransactionTemplateRepository) UpdateTransactionTemplate(tt domain.TransactionTemplate) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	existingTemplate, ok := r.repo.transactionTemplates[tt.ID]
	if !ok {
		return domain.ErrTransactionTemplateNotFound
	}
	existingTemplate.Day = tt.Day
	existingTemplate.BudgetID = tt.BudgetID
	existingTemplate.WalletID = tt.WalletID
	existingTemplate.Description = tt.Description
	existingTemplate.AmountInCents = tt.AmountInCents
	existingTemplate.Type = tt.Type
	existingTemplate.Tags = tt.Tags
	r.repo.transactionTemplates[tt.ID] = existingTemplate
	return nil
}

func (r *TransactionTemplateRepository) DeleteTransactionTemplate(id int) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	if _, ok := r.repo.transactionTemplates[id]; !ok {
		return domain.ErrTransactionTemplateNotFound
	}
	delete(r.repo.transactionTemplates, id)
	return nil
}
