package postgres

import (
	"database/sql"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

type TradeRepository struct {
	db *sql.DB
}

func NewTradeRepository(db *sql.DB) *TradeRepository {
	return &TradeRepository{db: db}
}

const tradeColumns = `id, depot_id, wallet_transaction_id, wkn, type, quantity, total_in_cents, fees_in_cents, taxes_in_cents, timestamp`

func scanTrade(row interface{ Scan(...any) error }) (domain.Trade, error) {
	var t domain.Trade
	err := row.Scan(&t.ID, &t.DepotID, &t.WalletTransactionID, &t.WKN, &t.Type, &t.Quantity, &t.TotalInCents, &t.FeesInCents, &t.TaxesInCents, &t.Timestamp)
	return t, err
}

func (r *TradeRepository) SaveTrade(t domain.Trade) (int, error) {
	query := `INSERT INTO trades (depot_id, wallet_transaction_id, wkn, type, quantity, total_in_cents, fees_in_cents, taxes_in_cents, timestamp)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id`
	var id int
	err := r.db.QueryRow(query, t.DepotID, t.WalletTransactionID, t.WKN, t.Type, t.Quantity, t.TotalInCents, t.FeesInCents, t.TaxesInCents, t.Timestamp).Scan(&id)
	return id, err
}

func (r *TradeRepository) GetTradeByID(id int) (domain.Trade, error) {
	t, err := scanTrade(r.db.QueryRow(`SELECT `+tradeColumns+` FROM trades WHERE id = $1`, id))
	if err == sql.ErrNoRows {
		return domain.Trade{}, domain.ErrTradeNotFound
	}
	return t, err
}

func (r *TradeRepository) FindTradesByDepot(depotID int) ([]domain.Trade, error) {
	rows, err := r.db.Query(`SELECT `+tradeColumns+` FROM trades WHERE depot_id = $1 ORDER BY timestamp, id`, depotID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trades []domain.Trade
	for rows.Next() {
		t, err := scanTrade(rows)
		if err != nil {
			return nil, err
		}
		trades = append(trades, t)
	}
	return trades, rows.Err()
}

func (r *TradeRepository) CountTradesByDepot(depotID int) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM trades WHERE depot_id = $1`, depotID).Scan(&count)
	return count, err
}

func (r *TradeRepository) UpdateTrade(t domain.Trade) error {
	query := `UPDATE trades SET depot_id = $1, wallet_transaction_id = $2, wkn = $3, type = $4, quantity = $5, total_in_cents = $6, fees_in_cents = $7, taxes_in_cents = $8, timestamp = $9
	          WHERE id = $10`
	res, err := r.db.Exec(query, t.DepotID, t.WalletTransactionID, t.WKN, t.Type, t.Quantity, t.TotalInCents, t.FeesInCents, t.TaxesInCents, t.Timestamp, t.ID)
	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return domain.ErrTradeNotFound
	}
	return nil
}

func (r *TradeRepository) DeleteTrade(id int) error {
	_, err := r.db.Exec(`DELETE FROM trades WHERE id = $1`, id)
	return err
}

func (r *TradeRepository) DeleteAllTradesOfDepot(depotID int) error {
	_, err := r.db.Exec(`DELETE FROM trades WHERE depot_id = $1`, depotID)
	return err
}

func (r *TradeRepository) DeleteAllByUser(userID int) error {
	query := `DELETE FROM trades WHERE depot_id IN (SELECT id FROM depots WHERE user_id = $1)`
	_, err := r.db.Exec(query, userID)
	return err
}
