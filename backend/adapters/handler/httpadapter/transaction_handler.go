package httpadapter

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
	"github.com/go-chi/chi/v5"
)

type TransactionHandler struct {
	service ports.TransactionService
}

func NewTransactionHandler(service *ports.TransactionService) *TransactionHandler {
	return &TransactionHandler{service: *service}
}

func (h *TransactionHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	limit := 20
	offset := 0

	query := r.URL.Query()

	if l := query.Get("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	if o := query.Get("offset"); o != "" {
		if val, err := strconv.Atoi(o); err == nil && val >= 0 {
			offset = val
		}
	}

	txs, err := h.service.GetTransactions(userID, limit, offset)
	if err != nil {
		http.Error(w, "Fetch failed", http.StatusInternalServerError)
		return
	}

	total, err := h.service.GetTransactionCount(userID)
	if err != nil {
		http.Error(w, "Fetch failed", http.StatusInternalServerError)
		return
	}

	response := struct {
		Data  []domain.TransactionDTO `json:"data"`
		Total int                     `json:"total"`
	}{
		Data:  txs,
		Total: total,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *TransactionHandler) SearchTransactions(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	query := r.URL.Query()

	var criteria domain.TransactionSearchCriteria

	if pageStr := query.Get("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			criteria.Page = page
		}
	}
	if pageSizeStr := query.Get("pageSize"); pageSizeStr != "" {
		if pageSize, err := strconv.Atoi(pageSizeStr); err == nil {
			criteria.PageSize = pageSize
		}
	}

	if searchTerm := query.Get("q"); searchTerm != "" {
		criteria.SearchTerm = &searchTerm
	}

	if fromDateStr := query.Get("from"); fromDateStr != "" {
		if fromDate, err := time.Parse("2006-01-02", fromDateStr); err == nil {
			criteria.FromDate = &fromDate
		}
	}

	if untilDateStr := query.Get("until"); untilDateStr != "" {
		if untilDate, err := time.Parse("2006-01-02", untilDateStr); err == nil {
			criteria.UntilDate = &untilDate
		}
	}

	if budgetIDStr := query.Get("budget_id"); budgetIDStr != "" {
		if budgetID, err := strconv.Atoi(budgetIDStr); err == nil {
			criteria.BudgetID = &budgetID
		}
	}

	if walletIDStr := query.Get("wallet_id"); walletIDStr != "" {
		if walletID, err := strconv.Atoi(walletIDStr); err == nil {
			criteria.WalletID = &walletID
		}
	}

	if typeStr := query.Get("type"); typeStr != "" {
		transactionType := domain.TransactionType(typeStr)
		if transactionType == domain.Income || transactionType == domain.Expense {
			criteria.Type = &transactionType
		}
	}

	result, err := h.service.Search(userID, criteria)
	if err != nil {
		http.Error(w, "Failed to search transactions", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
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

func (h *TransactionHandler) Transfer(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	var req struct {
		FromWalletID int `json:"fromWalletId"`
		ToWalletID   int `json:"toWalletId"`
		Amount       int `json:"amount"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	err := h.service.CreateTransfer(userID, req.FromWalletID, req.ToWalletID, req.Amount)
	if err != nil {
		if err == domain.ErrSameWalletTransfer || err == domain.ErrInvalidAmount {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if err == domain.ErrWalletNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, "Error creating transfer", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TransactionHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	transactionIDStr := chi.URLParam(r, "id")
	if transactionIDStr == "" {
		http.Error(w, "Missing transaction ID", http.StatusBadRequest)
		return
	}

	transactionID, err := strconv.Atoi(transactionIDStr)
	if err != nil {
		http.Error(w, "Id is not valid", http.StatusBadRequest)
		return
	}

	var transaction domain.Transaction
	err = json.NewDecoder(r.Body).Decode(&transaction)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	transaction.ID = transactionID
	transaction.UserID = userID

	err = h.service.UpdateTransaction(userID, transaction)
	if err != nil {
		if err == domain.ErrUnauthorized {
			http.Error(w, "Unauthorized", http.StatusForbidden)
			return
		}
		http.Error(w, "Error updating transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(transaction)
}

func (h *TransactionHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value("userID").(int)
	transactionIDStr := chi.URLParam(r, "id")
	if transactionIDStr == "" {
		http.Error(w, "Missing transaction ID", http.StatusBadRequest)
		return
	}

	transactionID, err := strconv.Atoi(transactionIDStr)
	if err != nil {
		http.Error(w, "Id is not valid", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteTransaction(userID, transactionID)
	if err != nil {
		http.Error(w, "Error deleting transaction", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
