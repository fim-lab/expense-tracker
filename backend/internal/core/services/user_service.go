package services

import (
	"github.com/fim-lab/expense-tracker/backend/internal/core/domain"
	"github.com/fim-lab/expense-tracker/backend/internal/core/ports"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repo ports.ExpenseRepository
}

func NewUserService(repo ports.ExpenseRepository) ports.UserService {
	return &userService{repo: repo}
}

func (s *userService) Authenticate(username, password string) (domain.User, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return domain.User{}, domain.ErrUserNotFound
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return domain.User{}, domain.ErrUnauthorized
	}

	return user, nil
}
