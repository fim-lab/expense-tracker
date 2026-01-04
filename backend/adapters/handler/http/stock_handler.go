package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/fim-lab/expense-tracker/backend/internal/core/domain"
	"github.com/fim-lab/expense-tracker/backend/internal/core/ports"
)

type StockHandler struct {
	service ports.StockService
}

func NewStockHandler(service *ports.StockService) *StockHandler {
	return &StockHandler{service: *service}
}

func (h *StockHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getStocksHandler(w, userID)
	case http.MethodPost:
		h.createStockHandler(w, r, userID)
	case http.MethodDelete:
		h.deleteStockHandler(w, r, userID)
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *StockHandler) getStocksHandler(w http.ResponseWriter, userID int) {
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

func (h *StockHandler) createStockHandler(w http.ResponseWriter, r *http.Request, userID int) {
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

func (h *StockHandler) deleteStockHandler(w http.ResponseWriter, r *http.Request, userID int) {
	stockID := r.URL.Query().Get("id")
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
		log.Printf("Error deleting wallet: %v", err)
		http.Error(w, "Error deleting wallet", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
