package postgres

import (
	"database/sql"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

type LotRepository struct {
	db *sql.DB
}

func NewLotRepository(db *sql.DB) *LotRepository {
	return &LotRepository{db: db}
}

func (r *LotRepository) SaveLot(l domain.Lot) error {
	query := `INSERT INTO lots (depot_id, date_of_purchase, wkn, amount, remaining, price_in_cents) 
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(query, l.DepotID, l.DateOfPurchase, l.WKN, l.Amount, l.Remaining, l.PriceInCents)
	return err
}

func (r *LotRepository) GetLotByID(id int) (domain.Lot, error) {
	var l domain.Lot
	query := `SELECT id, depot_id, date_of_purchase, wkn, amount, remaining, price_in_cents 
	          FROM lots WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&l.ID, &l.DepotID, &l.DateOfPurchase, &l.WKN, &l.Amount, &l.Remaining, &l.PriceInCents)
	return l, err
}

func (r *LotRepository) FindLotsByDepot(depotID int) ([]domain.Lot, error) {
	query := `SELECT id, depot_id, date_of_purchase, wkn, amount, remaining, price_in_cents 
	          FROM lots WHERE depot_id = $1`
	rows, err := r.db.Query(query, depotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var lots []domain.Lot
	for rows.Next() {
		var l domain.Lot
		if err := rows.Scan(&l.ID, &l.DepotID, &l.DateOfPurchase, &l.WKN, &l.Amount, &l.Remaining, &l.PriceInCents); err != nil {
			return nil, err
		}
		lots = append(lots, l)
	}
	return lots, nil
}

func (r *LotRepository) DeleteLot(id int) error {
	query := `DELETE FROM lots WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *LotRepository) DeleteAllLotsOfDepot(depotID int) error {
	_, err := r.db.Exec("DELETE FROM lots WHERE depot_id = $1", depotID)
	return err
}

func (r *LotRepository) DeleteAllByUser(userID int) error {
	query := `DELETE FROM lots WHERE depot_id IN (SELECT id FROM depots WHERE user_id = $1)`
	_, err := r.db.Exec(query, userID)
	return err
}
