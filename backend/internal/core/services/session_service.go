package services

import (
	"time"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type SessionService struct {
	repo ports.SessionRepository
}

func NewSessionService(repo ports.SessionRepository) ports.SessionService {
	return &SessionService{repo: repo}
}

func (s *SessionService) CreateSession(session domain.Session) error {
	return s.repo.SaveSession(session)
}

func (s *SessionService) ValidateSession(token string) (bool, int) {
	session, err := s.repo.GetSessionByToken(token)
	if err != nil {
		return false, 0
	}

	if session.Expiry.Before(time.Now()) {
		return false, 0
	}

	return true, session.UserID
}

func (s *SessionService) DeleteSession(sessionID string) error {
	return s.repo.DeleteSession(sessionID)
}
