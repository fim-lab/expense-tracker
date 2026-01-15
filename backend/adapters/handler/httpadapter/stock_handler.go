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

func NewStockHandler(service *ports.StockService) *StockHandler {
	return &StockHandler{service: *service}
}

func (h *StockHandler) GetStocks(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stocks, err := h.service.GetStocks(userID)
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
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var stock domain.Stock
	if err := json.NewDecoder(r.Body).Decode(&stock); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	stock.UserID = userID

	if err := h.service.CreateStock(userID, stock); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(stock)
}

func (h *StockHandler) DeleteStock(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	stockID := chi.URLParam(r, "id")
	if stockID == "" {
		http.Error(w, "Missing stock ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(stockID)
	if err != nil {
		http.Error(w, "Id is not valid", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteStock(userID, id)
	if err != nil {
		log.Printf("Error deleting stock: %v", err)
		http.Error(w, "Error deleting stock", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
