package postgres

import (
	"database/sql"
	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUserByUsername(username string) (domain.User, error) {
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

func (r *UserRepository) SaveUser(u domain.User) error {
	query := `INSERT INTO users (username, password_hash)
	          VALUES ($1, $2)`
	_, err := r.db.Exec(query, u.Username, u.PasswordHash)
	return err
}

func (r *UserRepository) GetUserByID(userID int) (domain.User, error) {
	var u domain.User
	query := `SELECT id, username, password_hash FROM users WHERE id = $1`
	err := r.db.QueryRow(query, userID).Scan(&u.ID, &u.Username, &u.PasswordHash)
	if err != nil {
		if err == sql.ErrNoRows {
			return domain.User{}, domain.ErrUserNotFound
		}
		return domain.User{}, err
	}
	return u, nil
}
