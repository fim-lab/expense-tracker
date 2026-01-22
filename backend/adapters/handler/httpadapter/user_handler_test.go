package httpadapter

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fim-lab/expense-tracker/adapters/repository/memory"
	"github.com/fim-lab/expense-tracker/internal/core/services"
)

func TestUserHandler_GetSalary(t *testing.T) {
	repo := memory.NewCleanRepository()
	userService := services.NewUserService(repo)
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
	repo := memory.NewCleanRepository()
	userService := services.NewUserService(repo)
	handler := NewUserHandler(&userService)

	payload := map[string]int{"salary": 50000}
	body, _ := json.Marshal(payload)
	req := httptest.NewRequest("PUT", "/api/user/salary", bytes.NewBuffer(body))
	ctx := context.WithValue(req.Context(), "userID", 1)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	handler.UpdateSalary(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected status OK; got %d", rr.Code)
	}
}
