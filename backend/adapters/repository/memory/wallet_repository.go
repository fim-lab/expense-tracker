package memory

import (
	"sort"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

type WalletRepository struct {
	repo *inMemoryRepositories
}

func (r *WalletRepository) SaveWallet(w domain.Wallet) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	if w.ID == 0 {
		w.ID = r.repo.nextID()
	}
	r.repo.wallets[w.ID] = w
	return nil
}

func (r *WalletRepository) GetWalletByID(id int) (domain.Wallet, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()
	w, ok := r.repo.wallets[id]
	if !ok {
		return domain.Wallet{}, domain.ErrWalletNotFound
	}
	return w, nil
}

func (r *WalletRepository) FindWalletsByUser(userID int) ([]domain.Wallet, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()

	var userWallets []domain.Wallet
	for _, w := range r.repo.wallets {
		if w.UserID == userID {
			userWallets = append(userWallets, w)
		}
	}
	sort.Slice(userWallets, func(i, j int) bool {
		return userWallets[i].ID < userWallets[j].ID
	})
	return userWallets, nil
}

func (r *WalletRepository) DeleteWallet(id int) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	delete(r.repo.wallets, id)
	return nil
}

func (r *WalletRepository) UpdateWallet(w domain.Wallet) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	existingWallet, ok := r.repo.wallets[w.ID]
	if !ok {
		return domain.ErrWalletNotFound
	}
	existingWallet.Name = w.Name
	r.repo.wallets[w.ID] = existingWallet
	return nil
}
