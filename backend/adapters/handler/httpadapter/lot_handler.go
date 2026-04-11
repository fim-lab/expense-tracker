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

type LotHandler struct {
	service ports.LotService
}

func NewLotHandler(service *ports.LotService) *LotHandler {
	return &LotHandler{service: *service}
}

func (h *LotHandler) GetLots(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	lots, err := h.service.GetLots(userID)
	if err != nil {
		log.Printf("Error fetching lots: %v", err)
		http.Error(w, "Could not fetch lots", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(lots)
}

func (h *LotHandler) CreateLot(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var lot domain.Lot
	if err := json.NewDecoder(r.Body).Decode(&lot); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := h.service.CreateLot(userID, lot); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(lot)
}

func (h *LotHandler) DeleteLot(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	lotID := chi.URLParam(r, "id")
	if lotID == "" {
		http.Error(w, "Missing lot ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(lotID)
	if err != nil {
		http.Error(w, "Id is not valid", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteLot(userID, id)
	if err != nil {
		log.Printf("Error deleting lot: %v", err)
		http.Error(w, "Error deleting lot", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
