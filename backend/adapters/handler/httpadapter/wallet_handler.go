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

type WalletHandler struct {
	service ports.WalletService
}

func NewWalletHandler(service *ports.WalletService) *WalletHandler {
	return &WalletHandler{service: *service}
}

func (h *WalletHandler) GetWallets(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized: Invalid user ID session", http.StatusUnauthorized)
		return
	}

	wallets, err := h.service.GetWallets(userID)
	if err != nil {
		log.Printf("Error fetching wallets: %v", err)
		http.Error(w, "Could not fetch wallets", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(wallets)
}

func (h *WalletHandler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized: Invalid user ID session", http.StatusUnauthorized)
		return
	}

	var wallet domain.Wallet

	err := json.NewDecoder(r.Body).Decode(&wallet)
	if err != nil {
		log.Printf("JSON decode error: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	wallet.UserID = userID

	err = h.service.CreateWallet(userID, wallet)
	if err != nil {
		log.Printf("Error creating wallet: %v", err)
		http.Error(w, "Error creating wallet", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(wallet)
}

func (h *WalletHandler) GetWallet(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized: Invalid user ID session", http.StatusUnauthorized)
		return
	}

	walletID := chi.URLParam(r, "id")
	if walletID == "" {
		http.Error(w, "Missing wallet ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(walletID)
	if err != nil {
		http.Error(w, "Id is not valid", http.StatusBadRequest)
		return
	}

	wallet, err := h.service.GetWallet(userID, id)
	if err != nil {
		log.Printf("Error fetching wallet %d for user %d: %v", id, userID, err)
		if err == domain.ErrUnauthorized {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		} else {
			http.Error(w, "Could not fetch wallet", http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(wallet)
}

func (h *WalletHandler) DeleteWallet(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized: Invalid user ID session", http.StatusUnauthorized)
		return
	}

	walletID := chi.URLParam(r, "id")
	if walletID == "" {
		http.Error(w, "Missing wallet ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(walletID)
	if err != nil {
		http.Error(w, "Id is not valid", http.StatusBadRequest)
		return
	}

	err = h.service.DeleteWallet(userID, id)
	if err != nil {
		log.Printf("Error deleting wallet: %v", err)
		http.Error(w, "Error deleting wallet", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *WalletHandler) UpdateWallet(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized: Invalid user ID session", http.StatusUnauthorized)
		return
	}

	walletID := chi.URLParam(r, "id")
	if walletID == "" {
		http.Error(w, "Missing wallet ID", http.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(walletID)
	if err != nil {
		http.Error(w, "Id is not valid", http.StatusBadRequest)
		return
	}

	var wallet domain.Wallet

	err = json.NewDecoder(r.Body).Decode(&wallet)
	if err != nil {
		log.Printf("JSON decode error: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	wallet.ID = id
	wallet.UserID = userID

	err = h.service.UpdateWallet(userID, wallet)
	if err != nil {
		log.Printf("Error updating wallet %d for user %d: %v", id, userID, err)
		switch err {
		case domain.ErrWalletNotFound, domain.ErrMissingWallet:
			http.Error(w, err.Error(), http.StatusBadRequest)
		case domain.ErrUnauthorized:
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
		default:
			http.Error(w, "Could not update wallet", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
