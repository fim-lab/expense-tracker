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

const stockColumns = `id, wkn, ticker, price_in_cents, last_fetched`

func scanStock(row interface{ Scan(...any) error }) (domain.Stock, error) {
	var s domain.Stock
	err := row.Scan(&s.ID, &s.WKN, &s.Ticker, &s.PriceInCents, &s.LastFetched)
	return s, err
}

func (r *StockRepository) FindAllStocks() ([]domain.Stock, error) {
	rows, err := r.db.Query(`SELECT ` + stockColumns + ` FROM stocks ORDER BY wkn`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []domain.Stock
	for rows.Next() {
		s, err := scanStock(rows)
		if err != nil {
			return nil, err
		}
		stocks = append(stocks, s)
	}
	return stocks, rows.Err()
}

func (r *StockRepository) GetStockByID(id int) (domain.Stock, error) {
	s, err := scanStock(r.db.QueryRow(`SELECT `+stockColumns+` FROM stocks WHERE id = $1`, id))
	if err == sql.ErrNoRows {
		return domain.Stock{}, domain.ErrStockNotFound
	}
	return s, err
}

func (r *StockRepository) FindStockByWKN(wkn string) (domain.Stock, error) {
	s, err := scanStock(r.db.QueryRow(`SELECT `+stockColumns+` FROM stocks WHERE wkn = $1`, wkn))
	if err == sql.ErrNoRows {
		return domain.Stock{}, domain.ErrStockNotFound
	}
	return s, err
}

func (r *StockRepository) SaveStock(s domain.Stock) (int, error) {
	query := `INSERT INTO stocks (wkn, ticker, price_in_cents, last_fetched)
	          VALUES ($1, $2, $3, $4) RETURNING id`
	var id int
	err := r.db.QueryRow(query, s.WKN, s.Ticker, s.PriceInCents, s.LastFetched).Scan(&id)
	return id, err
}

func (r *StockRepository) UpdateStock(s domain.Stock) error {
	query := `UPDATE stocks SET wkn = $1, ticker = $2, price_in_cents = $3, last_fetched = $4
	          WHERE id = $5`
	res, err := r.db.Exec(query, s.WKN, s.Ticker, s.PriceInCents, s.LastFetched, s.ID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrStockNotFound
	}
	return nil
}

func (r *StockRepository) DeleteStock(id int) error {
	_, err := r.db.Exec(`DELETE FROM stocks WHERE id = $1`, id)
	return err
}
