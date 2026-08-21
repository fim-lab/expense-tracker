package postgres

import (
	"database/sql"
	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

type DepotRepository struct {
	db *sql.DB
}

func NewDepotRepository(db *sql.DB) *DepotRepository {
	return &DepotRepository{db: db}
}

func (r *DepotRepository) SaveDepot(d domain.Depot) error {
	query := `INSERT INTO depots (user_id, wallet_id, budget_id, name) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(query, d.UserID, d.WalletID, d.BudgetID, d.Name)
	return err
}

func (r *DepotRepository) GetDepotByID(id int) (domain.Depot, error) {
	var d domain.Depot
	query := `SELECT id, user_id, wallet_id, budget_id, name FROM depots WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&d.ID, &d.UserID, &d.WalletID, &d.BudgetID, &d.Name)
	return d, err
}

func (r *DepotRepository) UpdateDepot(d domain.Depot) error {
	query := `UPDATE depots SET wallet_id = $1, budget_id = $2, name = $3 WHERE id = $4`
	_, err := r.db.Exec(query, d.WalletID, d.BudgetID, d.Name, d.ID)
	return err
}

func (r *DepotRepository) FindDepotsByUser(userID int) ([]domain.Depot, error) {
	query := `SELECT id, user_id, wallet_id, budget_id, name FROM depots WHERE user_id = $1`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var depots []domain.Depot
	for rows.Next() {
		var d domain.Depot
		if err := rows.Scan(&d.ID, &d.UserID, &d.WalletID, &d.BudgetID, &d.Name); err != nil {
			return nil, err
		}
		depots = append(depots, d)
	}
	return depots, rows.Err()
}

func (r *DepotRepository) DeleteDepot(id int) error {
	_, err := r.db.Exec("DELETE FROM depots WHERE id = $1", id)
	return err
}

func (r *DepotRepository) DeleteAllByUser(userID int) error {
	_, err := r.db.Exec("DELETE FROM depots WHERE user_id = $1", userID)
	return err
}
