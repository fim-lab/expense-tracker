package httpadapter

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
	"github.com/go-chi/chi/v5"
)

type TradeHandler struct {
	service ports.TradeService
}

func NewTradeHandler(service ports.TradeService) *TradeHandler {
	return &TradeHandler{service: service}
}

func userIDFromContext(w http.ResponseWriter, r *http.Request) (int, bool) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized: Invalid user ID session", http.StatusUnauthorized)
		return 0, false
	}
	return userID, true
}

func idFromURL(w http.ResponseWriter, r *http.Request, name string) (int, bool) {
	raw := chi.URLParam(r, name)
	if raw == "" {
		http.Error(w, "Missing "+name, http.StatusBadRequest)
		return 0, false
	}
	id, err := strconv.Atoi(raw)
	if err != nil {
		http.Error(w, "Id is not valid", http.StatusBadRequest)
		return 0, false
	}
	return id, true
}

func writeStockError(w http.ResponseWriter, err error, fallback string) {
	switch {
	case errors.Is(err, domain.ErrDepotNotFound), errors.Is(err, domain.ErrTradeNotFound):
		http.Error(w, err.Error(), http.StatusNotFound)
	case errors.Is(err, domain.ErrUnauthorized):
		http.Error(w, err.Error(), http.StatusForbidden)
	case errors.Is(err, domain.ErrInsufficientShares),
		errors.Is(err, domain.ErrInvalidAmount),
		errors.Is(err, domain.ErrInvalidQuantity),
		errors.Is(err, domain.ErrInvalidTradeType),
		errors.Is(err, domain.ErrMissingWKN),
		errors.Is(err, domain.ErrTradeDepotChange),
		errors.Is(err, domain.ErrNotEmpty):
		http.Error(w, err.Error(), http.StatusBadRequest)
	default:
		http.Error(w, fallback, http.StatusInternalServerError)
	}
}

func (h *TradeHandler) CreateTrade(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(w, r)
	if !ok {
		return
	}
	depotID, ok := idFromURL(w, r, "id")
	if !ok {
		return
	}

	var trade domain.Trade
	if err := json.NewDecoder(r.Body).Decode(&trade); err != nil {
		log.Printf("JSON decode error: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	trade.DepotID = depotID

	created, err := h.service.CreateTrade(userID, trade)
	if err != nil {
		log.Printf("Error creating trade in depot %d: %v", depotID, err)
		writeStockError(w, err, "Error creating trade")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(created)
}

func (h *TradeHandler) UpdateTrade(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(w, r)
	if !ok {
		return
	}
	tradeID, ok := idFromURL(w, r, "id")
	if !ok {
		return
	}

	var trade domain.Trade
	if err := json.NewDecoder(r.Body).Decode(&trade); err != nil {
		log.Printf("JSON decode error: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	trade.ID = tradeID

	if err := h.service.UpdateTrade(userID, trade); err != nil {
		log.Printf("Error updating trade %d: %v", tradeID, err)
		writeStockError(w, err, "Error updating trade")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *TradeHandler) DeleteTrade(w http.ResponseWriter, r *http.Request) {
	userID, ok := userIDFromContext(w, r)
	if !ok {
		return
	}
	tradeID, ok := idFromURL(w, r, "id")
	if !ok {
		return
	}

	if err := h.service.DeleteTrade(userID, tradeID); err != nil {
		log.Printf("Error deleting trade %d: %v", tradeID, err)
		writeStockError(w, err, "Error deleting trade")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
