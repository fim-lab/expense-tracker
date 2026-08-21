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

type StockHandler struct {
	service ports.StockService
}

func NewStockHandler(service ports.StockService) *StockHandler {
	return &StockHandler{service: service}
}

func (h *StockHandler) GetStocks(w http.ResponseWriter, r *http.Request) {
	stocks, err := h.service.GetStocks()
	if err != nil {
		log.Printf("Error fetching stocks: %v", err)
		http.Error(w, "Could not fetch stocks", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stocks)
}

func (h *StockHandler) CreateStock(w http.ResponseWriter, r *http.Request) {
	var stock domain.Stock
	if err := json.NewDecoder(r.Body).Decode(&stock); err != nil {
		log.Printf("JSON decode error: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	created, err := h.service.CreateStock(stock)
	if err != nil {
		log.Printf("Error creating stock: %v", err)
		switch err {
		case domain.ErrMissingWKN:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "Error creating stock", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *StockHandler) UpdateStock(w http.ResponseWriter, r *http.Request) {
	stockID := chi.URLParam(r, "id")
	id, err := strconv.Atoi(stockID)
	if err != nil {
		http.Error(w, "Id is not valid", http.StatusBadRequest)
		return
	}

	var stock domain.Stock
	if err := json.NewDecoder(r.Body).Decode(&stock); err != nil {
		log.Printf("JSON decode error: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	stock.ID = id

	updated, err := h.service.UpdateStock(stock)
	if err != nil {
		log.Printf("Error updating stock %d: %v", id, err)
		switch err {
		case domain.ErrStockNotFound, domain.ErrMissingWKN:
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "Could not update stock", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updated)
}

func (h *StockHandler) DeleteStock(w http.ResponseWriter, r *http.Request) {
	stockID := chi.URLParam(r, "id")
	id, err := strconv.Atoi(stockID)
	if err != nil {
		http.Error(w, "Id is not valid", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteStock(id); err != nil {
		log.Printf("Error deleting stock %d: %v", id, err)
		switch err {
		case domain.ErrNotEmpty:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case domain.ErrStockNotFound:
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, "Error deleting stock", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
