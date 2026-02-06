package postgres

import (
	"database/sql"
	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

type BudgetRepository struct {
	db *sql.DB
}

func NewBudgetRepository(db *sql.DB) *BudgetRepository {
	return &BudgetRepository{db: db}
}

func (r *BudgetRepository) SaveBudget(b domain.Budget) error {
	query := `INSERT INTO budgets (user_id, name, limit_cents) 
	          VALUES ($1, $2, $3)`
	_, err := r.db.Exec(query, b.UserID, b.Name, b.LimitCents)
	return err
}

func (r *BudgetRepository) UpdateBudget(b domain.Budget) error {
	query := `
		UPDATE budgets
		SET name = $2, limit_cents = $3 
	    WHERE id = $1`
	_, err := r.db.Exec(query, b.ID, b.Name, b.LimitCents)
	return err
}

func (r *BudgetRepository) GetBudgetByID(id int) (domain.Budget, error) {
	var b domain.Budget
	err := r.db.QueryRow("SELECT id, user_id, name, limit_cents FROM budgets WHERE id = $1", id).
		Scan(&b.ID, &b.UserID, &b.Name, &b.LimitCents)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Budget{}, domain.ErrMissingBudget
		}
		return domain.Budget{}, err
	}
	return b, nil
}

func (r *BudgetRepository) FindBudgetsByUser(userID int) ([]domain.Budget, error) {
	rows, err := r.db.Query("SELECT id, user_id, name, limit_cents, balance_cents FROM budgets WHERE user_id = $1 ORDER BY id ASC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []domain.Budget
	for rows.Next() {
		var b domain.Budget
		if err := rows.Scan(&b.ID, &b.UserID, &b.Name, &b.LimitCents, &b.BalanceCents); err != nil {
			return nil, err
		}
		res = append(res, b)
	}
	return res, nil
}

func (r *BudgetRepository) DeleteBudget(id int) error {
	_, err := r.db.Exec("DELETE FROM budgets WHERE id = $1", id)
	return err
}
