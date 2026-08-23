package httpadapter

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
	"github.com/go-chi/chi/v5"
)

type BudgetGroupHandler struct {
	service ports.BudgetGroupService
}

func NewBudgetGroupHandler(service *ports.BudgetGroupService) *BudgetGroupHandler {
	return &BudgetGroupHandler{service: *service}
}

func (h *BudgetGroupHandler) GetBudgetGroups(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized: Invalid user ID session", http.StatusUnauthorized)
		return
	}

	groups, err := h.service.GetBudgetGroups(userID)
	if err != nil {
		log.Printf("Error fetching budget groups: %v", err)
		http.Error(w, "Could not fetch budget groups", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(groups)
}

func (h *BudgetGroupHandler) CreateBudgetGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized: Invalid user ID session", http.StatusUnauthorized)
		return
	}

	var group domain.BudgetGroup
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		log.Printf("JSON decode error: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	group.UserID = userID

	if err := h.service.CreateBudgetGroup(userID, group); err != nil {
		log.Printf("Error creating budget group: %v", err)
		switch err {
		case domain.ErrMissingBudgetGroup:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "Error creating budget group", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(group)
}

func (h *BudgetGroupHandler) UpdateBudgetGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized: Invalid user ID session", http.StatusUnauthorized)
		return
	}

	groupIDStr := chi.URLParam(r, "id")
	if groupIDStr == "" {
		http.Error(w, "Missing budget group ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(groupIDStr)
	if err != nil {
		http.Error(w, "Id is not valid", http.StatusBadRequest)
		return
	}

	var group domain.BudgetGroup
	if err := json.NewDecoder(r.Body).Decode(&group); err != nil {
		log.Printf("JSON decode error: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	group.ID = id
	group.UserID = userID

	if err := h.service.UpdateBudgetGroup(userID, group); err != nil {
		log.Printf("Error updating budget group %d for user %d: %v", id, userID, err)
		switch err {
		case domain.ErrBudgetGroupNotFound, domain.ErrMissingBudgetGroup:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case domain.ErrUnauthorized:
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		default:
			http.Error(w, "Could not update budget group", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *BudgetGroupHandler) DeleteBudgetGroup(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized: Invalid user ID session", http.StatusUnauthorized)
		return
	}

	groupIDStr := chi.URLParam(r, "id")
	if groupIDStr == "" {
		http.Error(w, "Missing budget group ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(groupIDStr)
	if err != nil {
		http.Error(w, "Id is not valid", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteBudgetGroup(userID, id); err != nil {
		log.Printf("Error deleting budget group: %v", err)
		switch err {
		case domain.ErrBudgetGroupNotFound:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case domain.ErrUnauthorized:
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		default:
			http.Error(w, "Error deleting budget group", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
