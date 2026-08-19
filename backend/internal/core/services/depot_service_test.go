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
