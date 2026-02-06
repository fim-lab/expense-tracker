package memory

import (
	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

type UserRepository struct {
	repo *inMemoryRepositories
}

func (r *UserRepository) GetUserByUsername(username string) (domain.User, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()
	user, ok := r.repo.users[username]
	if !ok {
		return domain.User{}, domain.ErrUserNotFound
	}
	return user, nil
}

func (r *UserRepository) SaveUser(u domain.User) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	if u.ID == 0 {
		u.ID = r.repo.nextID()
	}

	foundExistingUserByID := false
	var existingUsername string
	for uname, user := range r.repo.users {
		if user.ID == u.ID {
			foundExistingUserByID = true
			existingUsername = uname
			break
		}
	}

	if foundExistingUserByID && existingUsername != u.Username {
		delete(r.repo.users, existingUsername)
	}

	r.repo.users[u.Username] = u
	return nil
}

func (r *UserRepository) GetUserByID(userID int) (domain.User, error) {
	r.repo.mu.RLock()
	defer r.repo.mu.RUnlock()
	for _, user := range r.repo.users {
		if user.ID == userID {
			return user, nil
		}
	}
	return domain.User{}, domain.ErrUserNotFound
}

func (r *UserRepository) UpdateUserSalary(userID int, salary int) error {
	r.repo.mu.Lock()
	defer r.repo.mu.Unlock()
	for username, user := range r.repo.users {
		if user.ID == userID {
			user.SalaryCents = salary
			r.repo.users[username] = user
			return nil
		}
	}
	return domain.ErrUserNotFound
}
