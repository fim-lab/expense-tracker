package services

import (
	"testing"
	"time"

	"github.com/fim-lab/expense-tracker/adapters/repository/memory"
	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

func TestLotService_CreateLot(t *testing.T) {
	repos := memory.NewCleanRepositories()
	walletRepo := repos.WalletRepository()
	depotRepo := repos.DepotRepository()
	lotRepo := repos.LotRepository()
	
	userID := 1
	walletID := 1
	walletRepo.SaveWallet(domain.Wallet{ID: walletID, UserID: userID, Name: "Main Wallet"})
	
	depotID := 1
	depotRepo.SaveDepot(domain.Depot{ID: depotID, UserID: userID, Name: "Main Depot", WalletID: walletID})

	svc := NewLotService(lotRepo, depotRepo)

	lot := domain.Lot{
		DepotID:        depotID,
		WKN:            "A1JX52",
		Amount:         10.5,
		DateOfPurchase: time.Now(),
		PriceInCents:   10000,
	}

	err := svc.CreateLot(userID, lot)
	if err != nil {
		t.Fatalf("CreateLot failed: %v", err)
	}

	lots, err := svc.GetLots(userID)
	if err != nil {
		t.Fatalf("GetLots failed: %v", err)
	}

	if len(lots) != 1 {
		t.Errorf("Expected 1 lot, got %d", len(lots))
	}

	if lots[0].Remaining != lots[0].Amount {
		t.Errorf("Expected remaining %f, got %f", lots[0].Amount, lots[0].Remaining)
	}
}
