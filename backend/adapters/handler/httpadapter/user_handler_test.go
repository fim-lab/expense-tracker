package httpadapter

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fim-lab/expense-tracker/adapters/repository/memory"
	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/services"
)

func TestUserHandler_GetSalary(t *testing.T) {
	repos := memory.NewCleanRepositories()
	testUser := domain.User{
		ID:           1,
		Username:     "testuser",
		PasswordHash: "hash",
		SalaryCents:  50000,
	}
	repos.UserRepository().SaveUser(testUser)

	userService := services.NewUserService(repos.UserRepository())
	handler := NewUserHandler(&userService)

	req := httptest.NewRequest("GET", "/api/user/salary", nil)
	ctx := context.WithValue(req.Context(), "userID", 1)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.GetUser(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status OK; got %d", rr.Code)
	}
}

func TestUserHandler_UpdateSalary(t *testing.T) {
	repos := memory.NewCleanRepositories()
	testUser := domain.User{
		ID:           1,
		Username:     "testuser",
		PasswordHash: "hash",
		SalaryCents:  0,
	}
	repos.UserRepository().SaveUser(testUser)

	userService := services.NewUserService(repos.UserRepository())
	handler := NewUserHandler(&userService)

	payload := map[string]int{"salaryCents": 50000}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("PUT", "/api/user/salary", bytes.NewBuffer(body))
	ctx := context.WithValue(req.Context(), "userID", 1)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.UpdateSalary(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status OK; got %d", rr.Code)
	}

	user, _ := repos.UserRepository().GetUserByID(1)
	if user.SalaryCents != 50000 {
		t.Errorf("expected salary to be 50000, but got %d", user.SalaryCents)
	}
}
