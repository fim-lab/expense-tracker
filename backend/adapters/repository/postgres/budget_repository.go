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
	query := `INSERT INTO budgets (user_id, name, limit_cents, group_id)
	          VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(query, b.UserID, b.Name, b.LimitCents, nullableInt(b.GroupID))
	return err
}

func (r *BudgetRepository) UpdateBudget(b domain.Budget) error {
	query := `
		UPDATE budgets
		SET name = $2, limit_cents = $3, group_id = $4
	    WHERE id = $1`
	_, err := r.db.Exec(query, b.ID, b.Name, b.LimitCents, nullableInt(b.GroupID))
	return err
}

func (r *BudgetRepository) GetBudgetByID(id int) (domain.Budget, error) {
	var b domain.Budget
	var groupID sql.NullInt64
	err := r.db.QueryRow("SELECT id, user_id, name, limit_cents, group_id FROM budgets WHERE id = $1", id).
		Scan(&b.ID, &b.UserID, &b.Name, &b.LimitCents, &groupID)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Budget{}, domain.ErrMissingBudget
		}
		return domain.Budget{}, err
	}
	b.GroupID = intFromNullable(groupID)
	return b, nil
}

func (r *BudgetRepository) FindBudgetsByUser(userID int) ([]domain.Budget, error) {
	rows, err := r.db.Query("SELECT id, user_id, name, limit_cents, balance_cents, group_id FROM budgets WHERE user_id = $1 ORDER BY id ASC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []domain.Budget
	for rows.Next() {
		var b domain.Budget
		var groupID sql.NullInt64
		if err := rows.Scan(&b.ID, &b.UserID, &b.Name, &b.LimitCents, &b.BalanceCents, &groupID); err != nil {
			return nil, err
		}
		b.GroupID = intFromNullable(groupID)
		res = append(res, b)
	}
	return res, nil
}

func (r *BudgetRepository) DeleteBudget(id int) error {
	_, err := r.db.Exec("DELETE FROM budgets WHERE id = $1", id)
	return err
}

func (r *BudgetRepository) DeleteAllByUser(userID int) error {
	_, err := r.db.Exec("DELETE FROM budgets WHERE user_id = $1", userID)
	return err
}

func nullableInt(v *int) sql.NullInt64 {
	if v == nil {
		return sql.NullInt64{}
	}
	return sql.NullInt64{Int64: int64(*v), Valid: true}
}

func intFromNullable(v sql.NullInt64) *int {
	if !v.Valid {
		return nil
	}
	i := int(v.Int64)
	return &i
}
