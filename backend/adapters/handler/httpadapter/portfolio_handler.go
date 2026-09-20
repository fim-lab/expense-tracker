package httpadapter

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type PortfolioHandler struct {
	service ports.PortfolioService
}

func NewPortfolioHandler(service ports.PortfolioService) *PortfolioHandler {
	return &PortfolioHandler{service: service}
}

func (h *PortfolioHandler) GetPortfolio(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(w, r)
	if !ok {
		return
	}
	depotID, ok := idFromURL(w, r, "id")
	if !ok {
		return
	}

	portfolio, err := h.service.GetPortfolio(userID, depotID)
	if err != nil {
		log.Printf("Error fetching portfolio of depot %d: %v", depotID, err)
		writeStockError(w, err, "Could not fetch portfolio")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(portfolio)
}

func (h *PortfolioHandler) GetOwnedStocks(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(w, r)
	if !ok {
		return
	}

	stocks, err := h.service.GetOwnedStocks(userID)
	if err != nil {
		log.Printf("Error fetching owned stocks for user %d: %v", userID, err)
		writeStockError(w, err, "Could not fetch stocks")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(stocks)
}

func (h *PortfolioHandler) RefreshStaleStockPrices(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(w, r)
	if !ok {
		return
	}

	if err := h.service.RefreshStaleStockPrices(userID); err != nil {
		log.Printf("Error refreshing stale stock prices for user %d: %v", userID, err)
		http.Error(w, "Could not refresh stock prices", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *PortfolioHandler) GetTrades(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(w, r)
	if !ok {
		return
	}
	depotID, ok := idFromURL(w, r, "id")
	if !ok {
		return
	}

	trades, err := h.service.GetTrades(userID, depotID)
	if err != nil {
		log.Printf("Error fetching trades of depot %d: %v", depotID, err)
		writeStockError(w, err, "Could not fetch trades")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(trades)
}
