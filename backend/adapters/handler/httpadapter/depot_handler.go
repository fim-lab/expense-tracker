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

type DepotHandler struct {
	service ports.DepotService
}

func NewDepotHandler(service *ports.DepotService) *DepotHandler {
	return &DepotHandler{service: *service}
}

func (h *DepotHandler) GetDepots(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized: Invalid user ID session", http.StatusUnauthorized)
		return
	}

	depots, err := h.service.GetDepots(userID)
	if err != nil {
		log.Printf("Error fetching depots: %v", err)
		http.Error(w, "Could not fetch depots", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(depots)
}

func (h *DepotHandler) CreateDepot(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized: Invalid user ID session", http.StatusUnauthorized)
		return
	}

	var depot domain.Depot

	err := json.NewDecoder(r.Body).Decode(&depot)
	if err != nil {
		log.Printf("JSON decode error: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	depot.UserID = userID

	err = h.service.CreateDepot(userID, depot)
	if err != nil {
		log.Printf("Error creating depot: %v", err)
		http.Error(w, "Error creating depot", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(depot)
}

func (h *DepotHandler) DeleteDepot(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized: Invalid user ID session", http.StatusUnauthorized)
		return
	}

	depotID := chi.URLParam(r, "id")
	if depotID == "" {
		http.Error(w, "Missing depot ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(depotID)
	if err != nil {
		http.Error(w, "Id is not valid", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteDepot(userID, id)
	if err != nil {
		log.Printf("Error deleting depot: %v", err)
		http.Error(w, "Error deleting depot", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
