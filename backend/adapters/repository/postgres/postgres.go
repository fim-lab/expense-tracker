package postgres

import (
	"database/sql"
	"encoding/json"
	"github.com/fim-lab/expense-tracker/backend/internal/core/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Save(t domain.Transaction) error {
	tagsJSON, _ := json.Marshal(t.Tags)
	query := `
		INSERT INTO transactions (id, date, budget, description, amount_in_cents, wallet, type, is_pending, is_debt, tags)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		ON CONFLICT (id) DO UPDATE SET
		date=$2, budget=$3, description=$4, amount_in_cents=$5, wallet=$6, type=$7, is_pending=$8, is_debt=$9, tags=$10`
	
	_, err := r.db.Exec(query, t.ID, t.Date, t.Budget, t.Description, t.AmountInCents, t.Wallet, t.Type, t.IsPending, t.IsDebt, tagsJSON)
	return err
}

func (r *Repository) FindAll() ([]domain.Transaction, error) {
	rows, err := r.db.Query("SELECT id, date, budget, description, amount_in_cents, wallet, type, is_pending, is_debt, tags FROM transactions ORDER BY date DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []domain.Transaction
	for rows.Next() {
		var t domain.Transaction
		var tagsData []byte
		err := rows.Scan(&t.ID, &t.Date, &t.Budget, &t.Description, &t.AmountInCents, &t.Wallet, &t.Type, &t.IsPending, &t.IsDebt, &tagsData)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(tagsData, &t.Tags)
		txs = append(txs, t)
	}
	return txs, nil
}

func (r *Repository) Delete(id string) error {
	_, err := r.db.Exec("DELETE FROM transactions WHERE id = $1", id)
	return err
}