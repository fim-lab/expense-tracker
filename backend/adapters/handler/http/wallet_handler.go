package http

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type WalletHandler struct {
	service ports.WalletService
}

func NewWalletHandler(service *ports.WalletService) *WalletHandler {
	return &WalletHandler{service: *service}
}

func (h *WalletHandler) Handle(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value("userID").(int)
	if !ok {
		http.Error(w, "Unauthorized: Invalid user ID session", http.StatusUnauthorized)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.getWalletsHandler(w, userID)
	case http.MethodPost:
		h.createWalletHandler(w, r, userID)
	case http.MethodDelete:
		h.deleteWalletHandler(w, r, userID)
	default:
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
	}
}

func (h *WalletHandler) getWalletsHandler(w http.ResponseWriter, userID int) {
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

func (h *WalletHandler) createWalletHandler(w http.ResponseWriter, r *http.Request, userID int) {
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

func (h *WalletHandler) deleteWalletHandler(w http.ResponseWriter, r *http.Request, userID int) {
	walletID := r.URL.Query().Get("id")
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
