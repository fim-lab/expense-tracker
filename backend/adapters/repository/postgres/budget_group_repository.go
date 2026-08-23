package postgres

import (
	"database/sql"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

type BudgetGroupRepository struct {
	db *sql.DB
}

func NewBudgetGroupRepository(db *sql.DB) *BudgetGroupRepository {
	return &BudgetGroupRepository{db: db}
}

func (r *BudgetGroupRepository) SaveBudgetGroup(g domain.BudgetGroup) error {
	query := `INSERT INTO budget_groups (user_id, name) VALUES ($1, $2)`
	_, err := r.db.Exec(query, g.UserID, g.Name)
	return err
}

func (r *BudgetGroupRepository) FindBudgetGroupsByUser(userID int) ([]domain.BudgetGroup, error) {
	rows, err := r.db.Query("SELECT id, user_id, name FROM budget_groups WHERE user_id = $1 ORDER BY id ASC", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []domain.BudgetGroup
	for rows.Next() {
		var g domain.BudgetGroup
		if err := rows.Scan(&g.ID, &g.UserID, &g.Name); err != nil {
			return nil, err
		}
		res = append(res, g)
	}
	return res, nil
}

func (r *BudgetGroupRepository) UpdateBudgetGroup(g domain.BudgetGroup) error {
	query := `UPDATE budget_groups SET name = $2 WHERE id = $1`
	_, err := r.db.Exec(query, g.ID, g.Name)
	return err
}

func (r *BudgetGroupRepository) DeleteBudgetGroup(id int) error {
	_, err := r.db.Exec("DELETE FROM budget_groups WHERE id = $1", id)
	return err
}

func (r *BudgetGroupRepository) DeleteAllByUser(userID int) error {
	_, err := r.db.Exec("DELETE FROM budget_groups WHERE user_id = $1", userID)
	return err
}
