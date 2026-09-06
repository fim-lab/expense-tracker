package httpadapter

import (
	"encoding/json"
	"net/http"

	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type UserHandler struct {
	userService ports.UserService
}

func NewUserHandler(userService *ports.UserService) *UserHandler {
	return &UserHandler{
		userService: *userService,
	}
}

func (h *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "invalid user ID", http.StatusUnauthorized)
		return
	}
	user, err := h.userService.GetUserByID(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}