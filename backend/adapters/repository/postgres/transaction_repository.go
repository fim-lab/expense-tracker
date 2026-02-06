package postgres

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) SaveTransaction(t domain.Transaction) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	defer tx.Rollback()
	tags, _ := json.Marshal(t.Tags)

	query := `INSERT INTO transactions (user_id, date, budget_id, wallet_id, description, amount_in_cents, type, is_pending, is_debt, tags)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err = tx.Exec(query, t.UserID, t.Date, t.BudgetID, t.WalletID, t.Description, t.AmountInCents, t.Type, t.IsPending, t.IsDebt, tags)
	if err != nil {
		return fmt.Errorf("failed to insert transaction: %w", err)
	}
	adjustment := t.AmountInCents
	if t.Type == domain.Expense {
		adjustment = -t.AmountInCents
	}

	queryBudget := `
		UPDATE budgets 
		SET balance_cents = balance_cents + $1 
		WHERE id = $2 AND user_id = $3
	`
	_, err = tx.Exec(queryBudget, adjustment, t.BudgetID, t.UserID)
	if err != nil {
		return fmt.Errorf("failed to update budget balance: %w", err)
	}

	queryWallet := `
		UPDATE wallets
		SET balance_cents = balance_cents + $1
		WHERE id = $2 AND user_id = $3
	`
	_, err = tx.Exec(queryWallet, adjustment, t.WalletID, t.UserID)
	if err != nil {
		return fmt.Errorf("failed to update wallet balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *TransactionRepository) UpdateTransaction(t domain.Transaction) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	defer tx.Rollback()

	var oldT domain.Transaction
	var oldNullBudgetID sql.NullInt32
	queryFetch := `SELECT amount_in_cents, type, budget_id, wallet_id FROM transactions WHERE id = $1 AND user_id = $2`
	err = tx.QueryRow(queryFetch, t.ID, t.UserID).Scan(&oldT.AmountInCents, &oldT.Type, &oldNullBudgetID, &oldT.WalletID)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.ErrTransactionNotFound
		}
		return fmt.Errorf("could not find original transaction: %w", err)
	}
	if oldNullBudgetID.Valid {
		oldBudgetID := int(oldNullBudgetID.Int32)
		oldT.BudgetID = &oldBudgetID
	}

	oldAdjustment := oldT.AmountInCents
	if oldT.Type == domain.Income {
		oldAdjustment = -oldT.AmountInCents
	}

	queryRevertBudget := `UPDATE budgets SET balance_cents = balance_cents + $1 WHERE id = $2`
	_, err = tx.Exec(queryRevertBudget, oldAdjustment, oldT.BudgetID)
	if err != nil {
		return fmt.Errorf("failed to revert budget balance: %w", err)
	}

	queryRevertWallet := `UPDATE wallets SET balance_cents = balance_cents + $1 WHERE id = $2`
	_, err = tx.Exec(queryRevertWallet, oldAdjustment, oldT.WalletID)
	if err != nil {
		return fmt.Errorf("failed to revert wallet balance: %w", err)
	}

	tags, _ := json.Marshal(t.Tags)
	query := `
		UPDATE transactions
		SET date = $2, budget_id = $3, wallet_id = $4, description = $5, amount_in_cents = $6, type = $7, is_pending = $8, is_debt = $9, tags = $10
	    WHERE id = $1 AND user_id = $11`
	_, err = tx.Exec(query, t.ID, t.Date, t.BudgetID, t.WalletID, t.Description, t.AmountInCents, t.Type, t.IsPending, t.IsDebt, tags, t.UserID)
	if err != nil {
		return fmt.Errorf("failed to update transaction record: %w", err)
	}

	newAdjustment := t.AmountInCents
	if t.Type == domain.Expense {
		newAdjustment = -t.AmountInCents
	}

	queryApplyBudget := `UPDATE budgets SET balance_cents = balance_cents + $1 WHERE id = $2`
	_, err = tx.Exec(queryApplyBudget, newAdjustment, t.BudgetID)
	if err != nil {
		return fmt.Errorf("failed to apply new budget balance: %w", err)
	}

	queryApplyWallet := `UPDATE wallets SET balance_cents = balance_cents + $1 WHERE id = $2`
	_, err = tx.Exec(queryApplyWallet, newAdjustment, t.WalletID)
	if err != nil {
		return fmt.Errorf("failed to apply new wallet balance: %w", err)
	}

	return tx.Commit()
}

func (r *TransactionRepository) GetTransactionByID(id int) (domain.Transaction, error) {
	var t domain.Transaction
	var tags []byte
	query := `SELECT id, user_id, date, budget_id, wallet_id, description, amount_in_cents, type, is_pending, is_debt, tags 
	          FROM transactions WHERE id = $1`
	var nullBudgetID sql.NullInt32
	err := r.db.QueryRow(query, id).Scan(
		&t.ID, &t.UserID, &t.Date, &nullBudgetID, &t.WalletID, &t.Description, &t.AmountInCents, &t.Type, &t.IsPending, &t.IsDebt, &tags,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Transaction{}, domain.ErrTransactionNotFound
		}
		return domain.Transaction{}, err
	}
	if nullBudgetID.Valid {
		budgetID := int(nullBudgetID.Int32)
		t.BudgetID = &budgetID
	}
	json.Unmarshal(tags, &t.Tags)
	return t, nil
}

func (r *TransactionRepository) GetTransactionCount(userID int) (int, error) {
	query := `SELECT COUNT(*) FROM transactions WHERE user_id = $1`
	var count int
	err := r.db.QueryRow(query, userID).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *TransactionRepository) FindTransactionsByUser(userID int, limit int, offset int) ([]domain.TransactionDTO, error) {
	query := `
		SELECT t.id, t.date, t.description, t.amount_in_cents, t.type, t.is_pending, b.name as budget_name, w.name as wallet_name
		FROM transactions t
		LEFT JOIN budgets b ON t.budget_id = b.id
		LEFT JOIN wallets w ON t.wallet_id = w.id
		WHERE t.user_id = $1
		ORDER BY t.date DESC, t.id DESC
		LIMIT $2 OFFSET $3`
	rows, err := r.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []domain.TransactionDTO
	for rows.Next() {
		var t domain.TransactionDTO
		var nullBudgetName sql.NullString
		err := rows.Scan(&t.ID, &t.Date, &t.Description, &t.AmountInCents, &t.Type, &t.IsPending, &nullBudgetName, &t.WalletName)
		if err != nil {
			return nil, err
		}
		if nullBudgetName.Valid {
			t.BudgetName = nullBudgetName.String
		} else {
			t.BudgetName = ""
		}
		txs = append(txs, t)
	}
	return txs, nil
}

func (r *TransactionRepository) SearchTransactions(userID int, criteria domain.TransactionSearchCriteria) ([]domain.TransactionDTO, error) {
	query := `
		SELECT t.id, t.date, t.description, t.amount_in_cents, t.type, t.is_pending, b.name as budget_name, w.name as wallet_name
		FROM transactions t
		LEFT JOIN budgets b ON t.budget_id = b.id
		LEFT JOIN wallets w ON t.wallet_id = w.id
	`
	whereClause := " WHERE t.user_id = $1"
	args := []interface{}{userID}
	argID := 2

	if criteria.SearchTerm != nil && *criteria.SearchTerm != "" {
		whereClause += fmt.Sprintf(" AND t.description ILIKE $%d", argID)
		args = append(args, "%"+*criteria.SearchTerm+"%")
		argID++
	}
	if criteria.FromDate != nil {
		whereClause += fmt.Sprintf(" AND t.date >= $%d", argID)
		args = append(args, *criteria.FromDate)
		argID++
	}
	if criteria.UntilDate != nil {
		whereClause += fmt.Sprintf(" AND t.date <= $%d", argID)
		args = append(args, *criteria.UntilDate)
		argID++
	}
	if criteria.BudgetID != nil {
		whereClause += fmt.Sprintf(" AND t.budget_id = $%d", argID)
		args = append(args, *criteria.BudgetID)
		argID++
	}
	if criteria.WalletID != nil {
		whereClause += fmt.Sprintf(" AND t.wallet_id = $%d", argID)
		args = append(args, *criteria.WalletID)
		argID++
	}
	if criteria.Type != nil {
		whereClause += fmt.Sprintf(" AND t.type = $%d", argID)
		args = append(args, *criteria.Type)
		argID++
	}

	query += whereClause
	query += " ORDER BY t.date DESC, t.id DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argID, argID+1)
	args = append(args, criteria.PageSize, (criteria.Page-1)*criteria.PageSize)

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var txs []domain.TransactionDTO
	for rows.Next() {
		var t domain.TransactionDTO
		var nullBudgetName sql.NullString
		err := rows.Scan(&t.ID, &t.Date, &t.Description, &t.AmountInCents, &t.Type, &t.IsPending, &nullBudgetName, &t.WalletName)
		if err != nil {
			return nil, err
		}
		if nullBudgetName.Valid {
			t.BudgetName = nullBudgetName.String
		} else {
			t.BudgetName = ""
		}
		txs = append(txs, t)
	}
	return txs, nil
}

func (r *TransactionRepository) CountSearchedTransactions(userID int, criteria domain.TransactionSearchCriteria) (int, error) {
	query := `SELECT COUNT(t.id) FROM transactions t`
	whereClause := " WHERE t.user_id = $1"
	args := []interface{}{userID}
	argID := 2

	if criteria.SearchTerm != nil && *criteria.SearchTerm != "" {
		whereClause += fmt.Sprintf(" AND t.description ILIKE $%d", argID)
		args = append(args, "%"+*criteria.SearchTerm+"%")
		argID++
	}
	if criteria.FromDate != nil {
		whereClause += fmt.Sprintf(" AND t.date >= $%d", argID)
		args = append(args, *criteria.FromDate)
		argID++
	}
	if criteria.UntilDate != nil {
		whereClause += fmt.Sprintf(" AND t.date <= $%d", argID)
		args = append(args, *criteria.UntilDate)
		argID++
	}
	if criteria.BudgetID != nil {
		whereClause += fmt.Sprintf(" AND t.budget_id = $%d", argID)
		args = append(args, *criteria.BudgetID)
		argID++
	}
	if criteria.WalletID != nil {
		whereClause += fmt.Sprintf(" AND t.wallet_id = $%d", argID)
		args = append(args, *criteria.WalletID)
		argID++
	}
	if criteria.Type != nil {
		whereClause += fmt.Sprintf(" AND t.type = $%d", argID)
		args = append(args, *criteria.Type)
		argID++
	}

	query += whereClause
	var count int
	err := r.db.QueryRow(query, args...).Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (r *TransactionRepository) DeleteTransaction(id int) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var amount int
	var tType domain.TransactionType
	var nullBudgetID sql.NullInt32
	var walletID int
	var userID int
	queryFetch := `SELECT amount_in_cents, type, budget_id, wallet_id, user_id FROM transactions WHERE id = $1`
	err = tx.QueryRow(queryFetch, id).Scan(&amount, &tType, &nullBudgetID, &walletID, &userID)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.ErrTransactionNotFound
		}
		return err
	}

	adjustment := -amount
	if tType == domain.Expense {
		adjustment = amount
	}

	if nullBudgetID.Valid {
		queryBudget := `UPDATE budgets SET balance_cents = balance_cents + $1 WHERE id = $2 AND user_id = $3`
		_, err = tx.Exec(queryBudget, adjustment, nullBudgetID.Int32, userID)
		if err != nil {
			return err
		}
	}

	queryWallet := `UPDATE wallets SET balance_cents = balance_cents + $1 WHERE id = $2 AND user_id = $3`
	_, err = tx.Exec(queryWallet, adjustment, walletID, userID)
	if err != nil {
		return err
	}

	result, err := tx.Exec("DELETE FROM transactions WHERE id = $1", id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrTransactionNotFound
	}

	return tx.Commit()
}

func (r *TransactionRepository) CreateTransfer(from, to domain.Transaction) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("could not start transaction: %w", err)
	}
	defer tx.Rollback()

	fromTags, _ := json.Marshal(from.Tags)
	query := `INSERT INTO transactions (user_id, date, budget_id, wallet_id, description, amount_in_cents, type, is_pending, is_debt, tags)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err = tx.Exec(query, from.UserID, from.Date, from.BudgetID, from.WalletID, from.Description, from.AmountInCents, from.Type, from.IsPending, from.IsDebt, fromTags)
	if err != nil {
		return fmt.Errorf("failed to insert from-transaction: %w", err)
	}

	queryWalletFrom := `
		UPDATE wallets
		SET balance_cents = balance_cents - $1
		WHERE id = $2 AND user_id = $3
	`
	_, err = tx.Exec(queryWalletFrom, from.AmountInCents, from.WalletID, from.UserID)
	if err != nil {
		return fmt.Errorf("failed to update from-wallet balance: %w", err)
	}

	toTags, _ := json.Marshal(to.Tags)
	query = `INSERT INTO transactions (user_id, date, budget_id, wallet_id, description, amount_in_cents, type, is_pending, is_debt, tags)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err = tx.Exec(query, to.UserID, to.Date, to.ID, to.WalletID, to.Description, to.AmountInCents, to.Type, to.IsPending, to.IsDebt, toTags)
	if err != nil {
		return fmt.Errorf("failed to insert to-transaction: %w", err)
	}

	queryWalletTo := `
		UPDATE wallets
		SET balance_cents = balance_cents + $1
		WHERE id = $2 AND user_id = $3
	`
	_, err = tx.Exec(queryWalletTo, to.AmountInCents, to.WalletID, to.UserID)
	if err != nil {
		return fmt.Errorf("failed to update to-wallet balance: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *TransactionRepository) CountTransactionsByBudgetID(budgetID int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM transactions WHERE budget_id = $1`
	err := r.db.QueryRow(query, budgetID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count transactions for budget ID %d: %w", budgetID, err)
	}
	return count, nil
}

func (r *TransactionRepository) CountTransactionsByWalletID(walletID int) (int, error) {
	var count int
	query := `SELECT COUNT(*) FROM transactions WHERE wallet_id = $1`
	err := r.db.QueryRow(query, walletID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count transactions for wallet ID %d: %w", walletID, err)
	}
	return count, nil
}
