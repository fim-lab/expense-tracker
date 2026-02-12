package httpadapter

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
	"github.com/go-chi/chi/v5"
)

type TransactionTemplateHandler struct {
	transactionTemplateService ports.TransactionTemplateService
}

func NewTransactionTemplateHandler(transactionTemplateService *ports.TransactionTemplateService) *TransactionTemplateHandler {
	return &TransactionTemplateHandler{
		transactionTemplateService: *transactionTemplateService,
	}
}

func (h *TransactionTemplateHandler) CreateTransactionTemplate(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "invalid user ID", http.StatusUnauthorized)
		return
	}

	var tt domain.TransactionTemplate
	if err := json.NewDecoder(r.Body).Decode(&tt); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.transactionTemplateService.CreateTransactionTemplate(userID, tt); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tt)
}

func (h *TransactionTemplateHandler) GetTransactionTemplates(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "invalid user ID", http.StatusUnauthorized)
		return
	}

	templates, err := h.transactionTemplateService.GetTransactionTemplates(userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(templates)
}

func (h *TransactionTemplateHandler) GetTransactionTemplateByID(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "invalid user ID", http.StatusUnauthorized)
		return
	}

	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "invalid transaction template ID", http.StatusBadRequest)
		return
	}

	tt, err := h.transactionTemplateService.GetTransactionTemplate(userID, id)
	if err != nil {
		if err == domain.ErrTransactionTemplateNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tt)
}

func (h *TransactionTemplateHandler) UpdateTransactionTemplate(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "invalid user ID", http.StatusUnauthorized)
		return
	}

	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "invalid transaction template ID", http.StatusBadRequest)
		return
	}

	var tt domain.TransactionTemplate
	if err := json.NewDecoder(r.Body).Decode(&tt); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}
	tt.ID = id

	if err := h.transactionTemplateService.UpdateTransactionTemplate(userID, tt); err != nil {
		if err == domain.ErrTransactionTemplateNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func (h *TransactionTemplateHandler) DeleteTransactionTemplate(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "invalid user ID", http.StatusUnauthorized)
		return
	}

	idParam := chi.URLParam(r, "id")
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "invalid transaction template ID", http.StatusBadRequest)
		return
	}

	if err := h.transactionTemplateService.DeleteTransactionTemplate(userID, id); err != nil {
		if err == domain.ErrTransactionTemplateNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
