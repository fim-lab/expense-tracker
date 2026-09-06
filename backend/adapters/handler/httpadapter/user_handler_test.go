package httpadapter

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fim-lab/expense-tracker/adapters/repository/memory"
	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/services"
)

func TestUserHandler_GetUser(t *testing.T) {
	repos := memory.NewCleanRepositories()
	testUser := domain.User{
		ID:           1,
		Username:     "testuser",
		PasswordHash: "hash",
	}
	repos.UserRepository().SaveUser(testUser)

	userService := services.NewUserService(repos.UserRepository())
	handler := NewUserHandler(&userService)

	req := httptest.NewRequest("GET", "/api/users/me", nil)
	ctx := context.WithValue(req.Context(), "userID", 1)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.GetUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status OK; got %d", rr.Code)
	}
}
