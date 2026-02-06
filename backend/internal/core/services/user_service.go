package services

import (
	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repo ports.UserRepository
}

func NewUserService(repo ports.UserRepository) ports.UserService {
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

func (s *userService) GetUserByID(userID int) (domain.User, error) {
	return s.repo.GetUserByID(userID)
}

func (s *userService) UpdateSalary(userID int, salary int) error {
	return s.repo.UpdateUserSalary(userID, salary)
}
