package postgres

import (
	"database/sql"
	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

type StockRepository struct {
	db *sql.DB
}

func NewStockRepository(db *sql.DB) *StockRepository {
	return &StockRepository{db: db}
}

func (r *StockRepository) SaveStock(s domain.Stock) error {
	query := `INSERT INTO stocks (user_id, depot_id, date_of_purchase, wkn, amount, price_in_cents) 
	          VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(query, s.UserID, s.DepotID, s.DateOfPurchase, s.WKN, s.Amount, s.PriceInCents)
	return err
}

func (r *StockRepository) GetStockByID(id int) (domain.Stock, error) {
	var s domain.Stock
	query := `SELECT id, user_id, depot_id, date_of_purchase, wkn, amount, price_in_cents 
	          FROM stocks WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&s.ID, &s.UserID, &s.DepotID, &s.DateOfPurchase, &s.WKN, &s.Amount, &s.PriceInCents)
	return s, err
}

func (r *StockRepository) FindStocksByUser(userID int) ([]domain.Stock, error) {
	query := `SELECT id, user_id, depot_id, date_of_purchase, wkn, amount, price_in_cents 
	          FROM stocks WHERE user_id = $1`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []domain.Stock
	for rows.Next() {
		var s domain.Stock
		if err := rows.Scan(&s.ID, &s.UserID, &s.DepotID, &s.DateOfPurchase, &s.WKN, &s.Amount, &s.PriceInCents); err != nil {
			return nil, err
		}
		stocks = append(stocks, s)
	}
	return stocks, nil
}

func (r *StockRepository) DeleteStock(id int) error {
	query := `DELETE FROM stocks WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
