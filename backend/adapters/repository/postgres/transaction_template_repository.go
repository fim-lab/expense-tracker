package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/lib/pq"
)

type TransactionTemplateRepository struct {
	db *sql.DB
}

func NewTransactionTemplateRepository(db *sql.DB) *TransactionTemplateRepository {
	return &TransactionTemplateRepository{db: db}
}

func (r *TransactionTemplateRepository) SaveTransactionTemplate(tt domain.TransactionTemplate) error {
	query := `
		INSERT INTO transaction_templates (user_id, day, budget_id, wallet_id, description, amount_in_cents, type, tags)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`
	var id int
	_, tags := json.Marshal(tt.Tags)
	err := r.db.QueryRow(
		query,
		tt.UserID,
		tt.Day,
		tt.BudgetID,
		tt.WalletID,
		tt.Description,
		tt.AmountInCents,
		tt.Type,
		tags,
	).Scan(&id)

	if err != nil {
		return fmt.Errorf("error saving transaction template: %w", err)
	}
	tt.ID = id
	return nil
}

func (r *TransactionTemplateRepository) GetTransactionTemplateByID(id int) (domain.TransactionTemplate, error) {
	query := `
		SELECT id, user_id, day, budget_id, wallet_id, description, amount_in_cents, type, tags
		FROM transaction_templates
		WHERE id = $1
	`
	var tt domain.TransactionTemplate
	var budgetID sql.NullInt64
	var tags []sql.NullString

	err := r.db.QueryRow(query, id).Scan(
		&tt.ID,
		&tt.UserID,
		&tt.Day,
		&budgetID,
		&tt.WalletID,
		&tt.Description,
		&tt.AmountInCents,
		&tt.Type,
		&tags,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return domain.TransactionTemplate{}, domain.ErrTransactionTemplateNotFound
		}
		return domain.TransactionTemplate{}, fmt.Errorf("error getting transaction template by ID: %w", err)
	}

	if budgetID.Valid {
		bID := int(budgetID.Int64)
		tt.BudgetID = &bID
	}
	tt.Tags = make([]string, len(tags))
	for i, t := range tags {
		if t.Valid {
			tt.Tags[i] = t.String
		}
	}

	return tt, nil
}

func (r *TransactionTemplateRepository) FindTransactionTemplatesByUser(userID int) ([]domain.TransactionTemplate, error) {
	query := `
		SELECT id, user_id, day, budget_id, wallet_id, description, amount_in_cents, type, tags
		FROM transaction_templates
		WHERE user_id = $1
		ORDER BY id
	`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("error finding transaction templates by user: %w", err)
	}
	defer rows.Close()

	var templates []domain.TransactionTemplate
	for rows.Next() {
		var tt domain.TransactionTemplate
		var budgetID sql.NullInt64
		var tags pq.StringArray

		if err := rows.Scan(
			&tt.ID,
			&tt.UserID,
			&tt.Day,
			&budgetID,
			&tt.WalletID,
			&tt.Description,
			&tt.AmountInCents,
			&tt.Type,
			&tags,
		); err != nil {
			return nil, fmt.Errorf("error scanning transaction template row: %w", err)
		}

		if budgetID.Valid {
			bID := int(budgetID.Int64)
			tt.BudgetID = &bID
		}
		tt.Tags = []string(tags)

		templates = append(templates, tt)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error with rows in FindTransactionTemplatesByUser: %w", err)
	}

	return templates, nil
}

func (r *TransactionTemplateRepository) UpdateTransactionTemplate(tt domain.TransactionTemplate) error {
	query := `
		UPDATE transaction_templates
		SET day = $1, budget_id = $2, wallet_id = $3, description = $4, amount_in_cents = $5, type = $6, tags = $7
		WHERE id = $8 AND user_id = $9
	`

	_, tags := json.Marshal(tt.Tags)
	res, err := r.db.Exec(
		query,
		tt.Day,
		tt.BudgetID,
		tt.WalletID,
		tt.Description,
		tt.AmountInCents,
		tt.Type,
		tags,
		tt.ID,
		tt.UserID,
	)
	if err != nil {
		return fmt.Errorf("error updating transaction template: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected for update: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrTransactionTemplateNotFound
	}
	return nil
}

func (r *TransactionTemplateRepository) DeleteTransactionTemplate(id int) error {
	query := `
		DELETE FROM transaction_templates
		WHERE id = $1
	`
	res, err := r.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("error deleting transaction template: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("error checking rows affected for delete: %w", err)
	}
	if rowsAffected == 0 {
		return domain.ErrTransactionTemplateNotFound
	}
	return nil
}
