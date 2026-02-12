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
	query := `INSERT INTO depots (user_id, wallet_id, name) VALUES ($1, $2, $3)`
	_, err := r.db.Exec(query, d.UserID, d.WalletID, d.Name)
	return err
}

func (r *DepotRepository) GetDepotByID(id int) (domain.Depot, error) {
	var d domain.Depot
	query := `SELECT id, user_id, wallet_id, name FROM depots WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&d.ID, &d.UserID, &d.WalletID, &d.Name)
	return d, err
}

func (r *DepotRepository) FindDepotsByUser(userID int) ([]domain.Depot, error) {
	query := `SELECT id, user_id, wallet_id, name FROM depots WHERE user_id = $1`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var depots []domain.Depot
	for rows.Next() {
		var d domain.Depot
		if err := rows.Scan(&d.ID, &d.UserID, &d.WalletID, &d.Name); err != nil {
			return nil, err
		}
		depots = append(depots, d)
	}
	return depots, nil
}

func (r *DepotRepository) DeleteDepot(id int) error {
	_, err := r.db.Exec("DELETE FROM depots WHERE id = $1", id)
	return err
}
