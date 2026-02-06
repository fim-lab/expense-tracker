package memory

import (
	"sort"
	"strings"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

type TransactionRepository struct {
	repo *inMemoryRepositories
}

func (r *TransactionRepository) SaveTransaction(t domain.Transaction) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	if t.ID == 0 {
		t.ID = r.repo.nextID()
	}
	r.repo.transactions[t.ID] = t

	if t.BudgetID != nil {
		budget, ok := r.repo.budgets[*t.BudgetID]
		if ok {
			adjustment := t.AmountInCents
			if t.Type == domain.Expense {
				adjustment = -t.AmountInCents
			}
			budget.BalanceCents += adjustment
			r.repo.budgets[*t.BudgetID] = budget
		}
	}

	wallet, ok := r.repo.wallets[t.WalletID]
	if ok {
		adjustment := t.AmountInCents
		if t.Type == domain.Expense {
			adjustment = -t.AmountInCents
		}
		wallet.BalanceCents += adjustment
		r.repo.wallets[t.WalletID] = wallet
	}

	return nil
}

func (r *TransactionRepository) GetTransactionByID(id int) (domain.Transaction, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()
	t, ok := r.repo.transactions[id]
	if !ok {
		return domain.Transaction{}, domain.ErrTransactionNotFound
	}
	return t, nil
}

func (r *TransactionRepository) GetTransactionCount(userID int) (int, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()
	var res []domain.Transaction
	for _, t := range r.repo.transactions {
		if t.UserID == userID {
			res = append(res, t)
		}
	}
	return len(res), nil
}

func (r *TransactionRepository) FindTransactionsByUser(userID int, limit int, offset int) ([]domain.TransactionDTO, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()
	var res []domain.Transaction
	for _, t := range r.repo.transactions {
		if t.UserID == userID {
			res = append(res, t)
		}
	}

	sort.Slice(res, func(i, j int) bool {
		if res[i].Date.Equal(res[j].Date) {
			return res[i].ID > res[j].ID
		}
		return res[i].Date.After(res[j].Date)
	})

	start := offset
	if start >= len(res) {
		return []domain.TransactionDTO{}, nil
	}

	end := offset + limit
	if end > len(res) || limit <= 0 {
		end = len(res)
	}

	paginatedTxs := res[start:end]

	dtos := make([]domain.TransactionDTO, 0, len(paginatedTxs))
	for _, t := range paginatedTxs {
		var budgetName string
		if t.BudgetID != nil {
			if budget, ok := r.repo.budgets[*t.BudgetID]; ok {
				budgetName = budget.Name
			}
		}
		wallet := r.repo.wallets[t.WalletID]
		dtos = append(dtos, domain.TransactionDTO{
			ID:            t.ID,
			Date:          t.Date,
			Description:   t.Description,
			AmountInCents: t.AmountInCents,
			Type:          t.Type,
			BudgetName:    budgetName,
			WalletName:    wallet.Name,
			IsPending:     t.IsPending,
		})
	}

	return dtos, nil
}

func (r *TransactionRepository) SearchTransactions(userID int, criteria domain.TransactionSearchCriteria) ([]domain.TransactionDTO, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()

	var filtered []domain.Transaction
	for _, t := range r.repo.transactions {
		if t.UserID != userID {
			continue
		}

		if criteria.SearchTerm != nil && *criteria.SearchTerm != "" {
			term := strings.ToLower(*criteria.SearchTerm)
			if !strings.Contains(strings.ToLower(t.Description), term) {
				continue
			}
		}

		if criteria.FromDate != nil && t.Date.Before(*criteria.FromDate) {
			continue
		}
		if criteria.UntilDate != nil && t.Date.After(*criteria.UntilDate) {
			continue
		}

		if criteria.BudgetID != nil {
			if t.BudgetID == nil {
				continue
			}
			if *t.BudgetID != *criteria.BudgetID {
				continue
			}
		}

		if criteria.WalletID != nil && t.WalletID != *criteria.WalletID {
			continue
		}

		if criteria.Type != nil && t.Type != *criteria.Type {
			continue
		}

		filtered = append(filtered, t)
	}

	sort.Slice(filtered, func(i, j int) bool {
		if filtered[i].Date.Equal(filtered[j].Date) {
			return filtered[i].ID > filtered[j].ID
		}
		return filtered[i].Date.After(filtered[j].Date)
	})

	start := (criteria.Page - 1) * criteria.PageSize
	if start >= len(filtered) {
		return []domain.TransactionDTO{}, nil
	}

	end := start + criteria.PageSize
	if end > len(filtered) {
		end = len(filtered)
	}

	paginatedTxs := filtered[start:end]

	dtos := make([]domain.TransactionDTO, 0, len(paginatedTxs))
	for _, t := range paginatedTxs {
		var budgetName string
		if t.BudgetID != nil {
			if budget, ok := r.repo.budgets[*t.BudgetID]; ok {
				budgetName = budget.Name
			}
		}
		wallet := r.repo.wallets[t.WalletID]
		dtos = append(dtos, domain.TransactionDTO{
			ID:            t.ID,
			Date:          t.Date,
			Description:   t.Description,
			AmountInCents: t.AmountInCents,
			Type:          t.Type,
			BudgetName:    budgetName,
			WalletName:    wallet.Name,
			IsPending:     t.IsPending,
		})
	}

	return dtos, nil
}

func (r *TransactionRepository) CountSearchedTransactions(userID int, criteria domain.TransactionSearchCriteria) (int, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()

	var count int
	for _, t := range r.repo.transactions {
		if t.UserID != userID {
			continue
		}

		if criteria.SearchTerm != nil && *criteria.SearchTerm != "" {
			term := strings.ToLower(*criteria.SearchTerm)
			if !strings.Contains(strings.ToLower(t.Description), term) {
				continue
			}
		}

		if criteria.FromDate != nil && t.Date.Before(*criteria.FromDate) {
			continue
		}
		if criteria.UntilDate != nil && t.Date.After(*criteria.UntilDate) {
			continue
		}

		if criteria.BudgetID != nil {
			if t.BudgetID == nil {
				continue
			}
			if *t.BudgetID != *criteria.BudgetID {
				continue
			}
		}

		if criteria.WalletID != nil && t.WalletID != *criteria.WalletID {
			continue
		}

		if criteria.Type != nil && t.Type != *criteria.Type {
			continue
		}

		count++
	}

	return count, nil
}

func (r *TransactionRepository) DeleteTransaction(id int) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()

	tx, exists := r.repo.transactions[id]
	if !exists {
		return domain.ErrTransactionNotFound
	}

	adjustment := -tx.AmountInCents
	if tx.Type == domain.Expense {
		adjustment = tx.AmountInCents
	}

	if tx.BudgetID != nil {
		budget, ok := r.repo.budgets[*tx.BudgetID]
		if ok {
			budget.BalanceCents += adjustment
			r.repo.budgets[*tx.BudgetID] = budget
		}
	}

	wallet, ok := r.repo.wallets[tx.WalletID]
	if ok {
		wallet.BalanceCents += adjustment
		r.repo.wallets[tx.WalletID] = wallet
	}

	delete(r.repo.transactions, id)
	return nil
}

func (r *TransactionRepository) UpdateTransaction(t domain.Transaction) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()

	oldT, ok := r.repo.transactions[t.ID]
	if !ok {
		return domain.ErrTransactionNotFound
	}

	oldAdjustment := oldT.AmountInCents
	if oldT.Type == domain.Income {
		oldAdjustment = -oldT.AmountInCents
	}
	if oldT.BudgetID != nil {
		if budget, ok := r.repo.budgets[*oldT.BudgetID]; ok {
			budget.BalanceCents += oldAdjustment
			r.repo.budgets[*oldT.BudgetID] = budget
		}
	}
	if wallet, ok := r.repo.wallets[oldT.WalletID]; ok {
		wallet.BalanceCents += oldAdjustment
		r.repo.wallets[oldT.WalletID] = wallet
	}

	r.repo.transactions[t.ID] = t

	newAdjustment := t.AmountInCents
	if t.Type == domain.Expense {
		newAdjustment = -t.AmountInCents
	}
	if t.BudgetID != nil {
		if budget, ok := r.repo.budgets[*t.BudgetID]; ok {
			budget.BalanceCents += newAdjustment
			r.repo.budgets[*t.BudgetID] = budget
		}
	}
	if wallet, ok := r.repo.wallets[t.WalletID]; ok {
		wallet.BalanceCents += newAdjustment
		r.repo.wallets[t.WalletID] = wallet
	}

	return nil
}

func (r *TransactionRepository) CreateTransfer(from, to domain.Transaction) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()

	if from.ID == 0 {
		from.ID = r.repo.nextID()
	}
	r.repo.transactions[from.ID] = from
	fromWallet, ok := r.repo.wallets[from.WalletID]
	if !ok {
		return domain.ErrWalletNotFound
	}
	fromWallet.BalanceCents -= from.AmountInCents
	r.repo.wallets[from.WalletID] = fromWallet

	if to.ID == 0 {
		to.ID = r.repo.nextID()
	}
	r.repo.transactions[to.ID] = to
	toWallet, ok := r.repo.wallets[to.WalletID]
	if !ok {
		return domain.ErrWalletNotFound
	}
	toWallet.BalanceCents += to.AmountInCents
	r.repo.wallets[to.WalletID] = toWallet

	return nil
}

func (r *TransactionRepository) CountTransactionsByBudgetID(budgetID int) (int, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()
	count := 0
	for _, t := range r.repo.transactions {
		if t.BudgetID != nil && *t.BudgetID == budgetID {
			count++
		}
	}
	return count, nil
}

func (r *TransactionRepository) CountTransactionsByWalletID(walletID int) (int, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()
	count := 0
	for _, t := range r.repo.transactions {
		if t.WalletID == walletID {
			count++
		}
	}
	return count, nil
}
