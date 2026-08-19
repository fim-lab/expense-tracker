package services

import (
	"testing"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

func TestDepotService_DeleteDepotWithTradesIsRejected(t *testing.T) {
	f := newStockFixture(t)
	f.mustBuy(t, 1, 10, 10000)

	if err := f.depotSvc.DeleteDepot(f.userID, f.depotID); err != domain.ErrNotEmpty {
		t.Fatalf("expected ErrNotEmpty, got %v", err)
	}
	if _, err := f.depotSvc.GetDepotByID(f.userID, f.depotID); err != nil {
		t.Errorf("expected the depot to survive, got %v", err)
	}
}

func TestDepotService_DeleteEmptyDepotSucceeds(t *testing.T) {
	f := newStockFixture(t)

	if err := f.depotSvc.DeleteDepot(f.userID, f.depotID); err != nil {
		t.Fatalf("expected deleting an empty depot to succeed, got %v", err)
	}
	if _, err := f.depotSvc.GetDepotByID(f.userID, f.depotID); err != domain.ErrDepotNotFound {
		t.Errorf("expected ErrDepotNotFound after the delete, got %v", err)
	}
}

func TestDepotService_MissingDepotReportsDepotNotFound(t *testing.T) {
	f := newStockFixture(t)

	if _, err := f.depotSvc.GetDepotByID(f.userID, 999); err != domain.ErrDepotNotFound {
		t.Errorf("expected ErrDepotNotFound, got %v", err)
	}
}

func TestDepotService_GetDepotsCarryTheInvestedSum(t *testing.T) {
	f := newStockFixture(t)
	secondDepotID := 2
	if err := f.repos.DepotRepository().SaveDepot(domain.Depot{ID: secondDepotID, UserID: f.userID, Name: "Alpha Depot", WalletID: f.walletID}); err != nil {
		t.Fatalf("could not seed the second depot: %v", err)
	}
	f.mustBuy(t, 1, 10, 100000)
	f.mustBuy(t, 2, 5, 60000)
	f.mustSell(t, 3, 12, 150000)

	depots, err := f.depotSvc.GetDepots(f.userID)
	if err != nil {
		t.Fatalf("could not read the depots: %v", err)
	}
	if len(depots) != 2 {
		t.Fatalf("expected 2 depots, got %d", len(depots))
	}

	if depots[0].Name != "Alpha Depot" {
		t.Errorf("expected the depots sorted by name, got %s first", depots[0].Name)
	}
	if depots[0].InvestedInCents != 0 {
		t.Errorf("expected the empty depot to have nothing invested, got %d", depots[0].InvestedInCents)
	}

	if depots[1].InvestedInCents != 36000 {
		t.Errorf("expected 36000 invested after the partial sell, got %d", depots[1].InvestedInCents)
	}
	if depots[1].WalletID != f.walletID {
		t.Errorf("expected the depot to keep its wallet %d, got %d", f.walletID, depots[1].WalletID)
	}
}
