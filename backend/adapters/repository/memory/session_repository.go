package memory

import (
	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

type SessionRepository struct {
	repo *inMemoryRepositories
}

func (r *SessionRepository) SaveSession(session domain.Session) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	r.repo.sessions[session.SessionToken] = session
	return nil
}

func (r *SessionRepository) GetSessionByToken(token string) (domain.Session, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()
	s, ok := r.repo.sessions[token]
	if !ok {
		return domain.Session{}, domain.ErrSessionNotFound
	}
	return s, nil
}

func (r *SessionRepository) DeleteSession(sessionID string) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	delete(r.repo.sessions, sessionID)
	return nil
}
