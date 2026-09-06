package services

import (
	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
	"golang.org/x/crypto/bcrypt"
)

type userService struct {
	repo ports.UserRepository
}

// dummyPasswordHash to unify response time
var dummyPasswordHash = generateDummyPasswordHash()

func generateDummyPasswordHash() []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte("no-such-user"), bcrypt.DefaultCost)
	if err != nil {
		panic("could not generate dummy password hash: " + err.Error())
	}
	return hash
}

func NewUserService(repo ports.UserRepository) ports.UserService {
	return &userService{repo: repo}
}

func (s *userService) Authenticate(username, password string) (domain.User, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		_ = bcrypt.CompareHashAndPassword(dummyPasswordHash, []byte(password))
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
