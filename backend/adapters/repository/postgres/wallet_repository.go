package postgres

import (
	"database/sql"
	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

type WalletRepository struct {
	db *sql.DB
}

func NewWalletRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) SaveWallet(w domain.Wallet) error {
	query := `INSERT INTO wallets (user_id, name) VALUES ($1, $2)`
	_, err := r.db.Exec(query, w.UserID, w.Name)
	return err
}

func (r *WalletRepository) UpdateWallet(w domain.Wallet) error {
	query := `
		UPDATE wallets
		SET name = $2
	    WHERE id = $1`
	_, err := r.db.Exec(query, w.ID, w.Name)
	return err
}
func (r *WalletRepository) GetWalletByID(id int) (domain.Wallet, error) {
	var w domain.Wallet
	query := `SELECT id, user_id, name FROM wallets WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&w.ID, &w.UserID, &w.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Wallet{}, domain.ErrWalletNotFound
		}
		return domain.Wallet{}, err
	}
	return w, nil
}

func (r *WalletRepository) FindWalletsByUser(userID int) ([]domain.Wallet, error) {
	query := `SELECT id, user_id, name, balance_cents FROM wallets WHERE user_id = $1 ORDER BY id ASC`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []domain.Wallet
	for rows.Next() {
		var w domain.Wallet
		if err := rows.Scan(&w.ID, &w.UserID, &w.Name, &w.BalanceCents); err != nil {
			return nil, err
		}
		res = append(res, w)
	}
	return res, nil
}

func (r *WalletRepository) DeleteWallet(id int) error {
	_, err := r.db.Exec("DELETE FROM wallets WHERE id = $1", id)
	return err
}
