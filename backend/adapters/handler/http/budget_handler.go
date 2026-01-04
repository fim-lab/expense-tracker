package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/fim-lab/expense-tracker/backend/internal/core/domain"
	"github.com/fim-lab/expense-tracker/backend/internal/core/ports"
)

type BudgetHandler struct {
	service ports.BudgetService
}

func NewBudgetHandler(service *ports.BudgetService) *BudgetHandler {
	return &BudgetHandler{service: *service}
}

func (h *BudgetHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized: Invalid user ID session", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getBudgetsHandler(w, r, userID)
	case http.MethodPost:
		h.createBudgetHandler(w, r, userID)
	case http.MethodDelete:
		h.deleteBudgetHandler(w, r, userID)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (h *BudgetHandler) getBudgetsHandler(w http.ResponseWriter, r *http.Request, userID int) {
	budgets, err := h.service.GetBudgets(userID)
	if err != nil {
		log.Printf("Error fetching budgets: %v", err)
		http.Error(w, "Could not fetch budgets", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(budgets)
}

func (h *BudgetHandler) createBudgetHandler(w http.ResponseWriter, r *http.Request, userID int) {
	var budget domain.Budget

	err := json.NewDecoder(r.Body).Decode(&budget)
	if err != nil {
		log.Printf("JSON decode error: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	budget.UserID = userID

	err = h.service.CreateBudget(userID, budget)
	if err != nil {
		log.Printf("Error creating budget: %v", err)
		http.Error(w, "Error creating budget", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(budget)
}

func (h *BudgetHandler) deleteBudgetHandler(w http.ResponseWriter, r *http.Request, userID int) {
	budgetID := r.URL.Query().Get("id")
	if budgetID == "" {
		http.Error(w, "Missing budget ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(budgetID)
	if err != nil {
		http.Error(w, "Id is not valid", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteBudget(userID, id)
	if err != nil {
		log.Printf("Error deleting budget: %v", err)
		http.Error(w, "Error deleting budget", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
