package http

import (
	"encoding/json"
	"net/http"
	"github.com/fim-lab/expense-tracker/backend/internal/core/domain"
	"github.com/fim-lab/expense-tracker/backend/internal/core/ports"
)

type Adapter struct {
	service ports.ExpenseService
}

func NewAdapter(service ports.ExpenseService) *Adapter {
	return &Adapter{
		service: service,
	}
}

func (a *Adapter) CreateTransactionHandler(w http.ResponseWriter, r *http.Request) {
	var tx domain.Transaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	if err := a.service.CreateTransaction(tx); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
}

func (a *Adapter) GetTransactionsHandler(w http.ResponseWriter, r *http.Request) {
	txs, err := a.service.GetTransactions()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(txs)
}

func (a *Adapter) DeleteTransactionHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	if err := a.service.DeleteTransaction(id); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}