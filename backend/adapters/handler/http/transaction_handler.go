package http

import (
	"encoding/json"
	"net/http"

	"github.com/fim-lab/expense-tracker/backend/internal/core/domain"
	"github.com/fim-lab/expense-tracker/backend/internal/core/ports"
	"github.com/google/uuid"
)

type TransactionHandler struct {
	service ports.TransactionService
}

func NewTransactionHandler(service *ports.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: *service}
}

func (h *TransactionHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)

	switch r.Method {
	case http.MethodGet:
		h.getTransactionsHandler(w, r, userID)
	case http.MethodPost:
		h.createTransactionHandler(w, r, userID)
	case http.MethodDelete:
		h.deleteTransactionHandler(w, r, userID)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (h *TransactionHandler) getTransactionsHandler(w http.ResponseWriter, r *http.Request, userID int) {
	transactions, err := h.service.GetTransactions(userID)
	if err != nil {
		http.Error(w, "Could not fetch transactions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transactions)
}

func (h *TransactionHandler) createTransactionHandler(w http.ResponseWriter, r *http.Request, userID int) {
	var transaction domain.Transaction

	err := json.NewDecoder(r.Body).Decode(&transaction)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	transaction.UserID = userID

	err = h.service.CreateTransaction(userID, transaction)
	if err != nil {
		http.Error(w, "Error creating transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(transaction)
}

func (h *TransactionHandler) deleteTransactionHandler(w http.ResponseWriter, r *http.Request, userID int) {
	transactionID := r.URL.Query().Get("id")
	if transactionID == "" {
		http.Error(w, "Missing transaction ID", http.StatusBadRequest)
		return
	}

	uuid, err := uuid.Parse(transactionID)
	if err != nil {
		http.Error(w, "Id is not a valid UUID", http.StatusInternalServerError)
	}

	err = h.service.DeleteTransaction(userID, uuid)
	if err != nil {
		http.Error(w, "Error deleting transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
