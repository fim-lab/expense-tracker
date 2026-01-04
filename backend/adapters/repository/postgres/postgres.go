package postgres

import (
	"database/sql"
	"encoding/json"

	"github.com/fim-lab/expense-tracker/backend/internal/core/domain"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// --- User Methods ---

func (r *Repository) GetUserByUsername(username string) (domain.User, error) {
	var u domain.User
	query := `SELECT id, username, password_hash FROM users WHERE username = $1`
	err := r.db.QueryRow(query, username).Scan(&u.ID, &u.Username, &u.PasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, err
	}
	return u, nil
}

func (r *Repository) SaveUser(u domain.User) error {
	query := `INSERT INTO users (username, password_hash) 
	          VALUES ($1, $2)`
	_, err := r.db.Exec(query, u.Username, u.PasswordHash)
	return err
}

func (r *Repository) UpdateUser(u domain.User) error {
	query := `
        UPDATE users
        SET username = $2, password_hash = $3
        WHERE id = $1`
	_, err := r.db.Exec(query, u.ID, u.Username, u.PasswordHash)
	return err
}

// --- Budget Methods ---

func (r *Repository) SaveBudget(b domain.Budget) error {
	query := `INSERT INTO budgets (user_id, name, limit_cents) 
	          VALUES ($1, $2, $3)`
	_, err := r.db.Exec(query, b.UserID, b.Name, b.LimitCents)
	return err
}

func (r *Repository) UpdateBudget(b domain.Budget) error {
	query := `
		UPDATE budgets
		SET name = $2, limit_cents = $3 
	    WHERE id = $1`
	_, err := r.db.Exec(query, b.ID, b.Name, b.LimitCents)
	return err
}

func (r *Repository) GetBudgetByID(id int) (domain.Budget, error) {
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

func (r *Repository) FindBudgetsByUser(userID int) ([]domain.Budget, error) {
	rows, err := r.db.Query("SELECT id, user_id, name, limit_cents FROM budgets WHERE user_id = $1", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []domain.Budget
	for rows.Next() {
		var b domain.Budget
		if err := rows.Scan(&b.ID, &b.UserID, &b.Name, &b.LimitCents); err != nil {
			return nil, err
		}
		res = append(res, b)
	}
	return res, nil
}

func (r *Repository) DeleteBudget(id int) error {
	_, err := r.db.Exec("DELETE FROM budgets WHERE id = $1", id)
	return err
}

// --- Wallet Methods ---

func (r *Repository) SaveWallet(w domain.Wallet) error {
	query := `INSERT INTO wallets (user_id, name) VALUES ($1, $2)`
	_, err := r.db.Exec(query, w.UserID, w.Name)
	return err
}

func (r *Repository) UpdateWallet(b domain.Wallet) error {
	query := `
		UPDATE wallets
		SET name = $2 
	    WHERE id = $1`
	_, err := r.db.Exec(query, b.ID, b.Name)
	return err
}

func (r *Repository) GetWalletByID(id int) (domain.Wallet, error) {
	var w domain.Wallet
	query := `SELECT id, user_id, name FROM wallets WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&w.ID, &w.UserID, &w.Name)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Wallet{}, domain.ErrWalletNotFound
		}
		return domain.Wallet{}, err
	}
	return w, nil
}

func (r *Repository) FindWalletsByUser(userID int) ([]domain.Wallet, error) {
	query := `
		SELECT w.id, w.user_id, w.name, 
		COALESCE(SUM(CASE WHEN t.type = 'INCOME' THEN t.amount_in_cents 
		                  WHEN t.type = 'EXPENSE' THEN -t.amount_in_cents 
		                  ELSE 0 END), 0) as balance
		FROM wallets w
		LEFT JOIN transactions t ON w.id = t.wallet_id
		WHERE w.user_id = $1
		GROUP BY w.id, w.user_id, w.name`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []domain.Wallet
	for rows.Next() {
		var w domain.Wallet
		if err := rows.Scan(&w.ID, &w.UserID, &w.Name, &w.Balance); err != nil {
			return nil, err
		}
		res = append(res, w)
	}
	return res, nil
}

func (r *Repository) DeleteWallet(id int) error {
	_, err := r.db.Exec("DELETE FROM wallets WHERE id = $1", id)
	return err
}

// --- Transaction Methods ---

func (r *Repository) SaveTransaction(t domain.Transaction) error {
	tags, _ := json.Marshal(t.Tags)
	query := `INSERT INTO transactions (user_id, date, budget_id, wallet_id, description, amount_in_cents, type, is_pending, is_debt, tags)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`
	_, err := r.db.Exec(query, t.UserID, t.Date, t.BudgetID, t.WalletID, t.Description, t.AmountInCents, t.Type, t.IsPending, t.IsDebt, tags)
	return err
}

func (r *Repository) UpdateTransaction(t domain.Transaction) error {
	tags, _ := json.Marshal(t.Tags)
	query := `
		UPDATE transactions
		SET date = $2, budget_id = $3, wallet_id = $4, description = $5, amount_in_cents = $6, type = $7, is_pending = $8, is_debt = $9, tags = $10
	    WHERE id = $1`
	_, err := r.db.Exec(query, t.ID, t.Date, t.BudgetID, t.WalletID, t.Description, t.AmountInCents, t.Type, t.IsPending, t.IsDebt, tags)
	return err
}

func (r *Repository) GetTransactionByID(id int) (domain.Transaction, error) {
	var t domain.Transaction
	var tags []byte
	query := `SELECT id, user_id, date, budget_id, wallet_id, description, amount_in_cents, type, is_pending, is_debt, tags 
	          FROM transactions WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(
		&t.ID, &t.UserID, &t.Date, &t.BudgetID, &t.WalletID, &t.Description, &t.AmountInCents, &t.Type, &t.IsPending, &t.IsDebt, &tags,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Transaction{}, domain.ErrTransactionNotFound
		}
		return domain.Transaction{}, err
	}
	json.Unmarshal(tags, &t.Tags)
	return t, nil
}

func (r *Repository) FindTransactionsByUser(userID int) ([]domain.Transaction, error) {
	query := `SELECT id, user_id, date, budget_id, wallet_id, description, amount_in_cents, type, is_pending, is_debt, tags 
	          FROM transactions WHERE user_id = $1 ORDER BY date DESC`
	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []domain.Transaction
	for rows.Next() {
		var t domain.Transaction
		var tags []byte
		err := rows.Scan(&t.ID, &t.UserID, &t.Date, &t.BudgetID, &t.WalletID, &t.Description, &t.AmountInCents, &t.Type, &t.IsPending, &t.IsDebt, &tags)
		if err != nil {
			return nil, err
		}
		json.Unmarshal(tags, &t.Tags)
		res = append(res, t)
	}
	return res, nil
}

func (r *Repository) DeleteTransaction(id int) error {
	result, err := r.db.Exec("DELETE FROM transactions WHERE id = $1", id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return domain.ErrTransactionNotFound
	}
	return nil
}

// --- Session Methods ---

func (r *Repository) SaveSession(session domain.Session) error {
	query := `
        INSERT INTO sessions (token, user_id, expiry)
        VALUES ($1, $2, $3)
    `
	_, err := r.db.Exec(query, session.SessionToken, session.UserID, session.Expiry)
	return err
}

func (r *Repository) GetSessionByToken(token string) (domain.Session, error) {
	query := `SELECT token, user_id, expiry FROM sessions WHERE token = $1`
	row := r.db.QueryRow(query, token)

	var session domain.Session
	err := row.Scan(&session.SessionToken, &session.UserID, &session.Expiry)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.Session{}, domain.ErrSessionNotFound
		}
		return domain.Session{}, err
	}
	return session, nil
}

func (r *Repository) DeleteSession(token string) error {
	query := `DELETE FROM sessions WHERE token = $1`
	_, err := r.db.Exec(query, token)
	return err
}
