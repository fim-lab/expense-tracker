package postgres

import (
	"database/sql"
	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

type SessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) SaveSession(session domain.Session) error {
	query := `
        INSERT INTO sessions (token, user_id, expiry)
        VALUES ($1, $2, $3)
    `
	_, err := r.db.Exec(query, session.SessionToken, session.UserID, session.Expiry)
	return err
}

func (r *SessionRepository) GetSessionByToken(token string) (domain.Session, error) {
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

func (r *SessionRepository) DeleteSession(token string) error {
	query := `DELETE FROM sessions WHERE token = $1`
	_, err := r.db.Exec(query, token)
	return err
}
