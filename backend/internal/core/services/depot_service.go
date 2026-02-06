package services

import (
	"errors"
	"strings"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type depotService struct {
	depotRepo  ports.DepotRepository
	walletRepo ports.WalletRepository
}

func NewDepotService(depotRepo ports.DepotRepository, walletRepo ports.WalletRepository) ports.DepotService {
	return &depotService{depotRepo: depotRepo, walletRepo: walletRepo}
}

func (s *depotService) CreateDepot(userID int, d domain.Depot) error {
	d.UserID = userID

	if strings.TrimSpace(d.Name) == "" {
		return errors.New("depot name is required")
	}

	wallet, err := s.walletRepo.GetWalletByID(d.WalletID)
	if err != nil || wallet.UserID != userID {
		return errors.New("invalid wallet for depot")
	}

	return s.depotRepo.SaveDepot(d)
}

func (s *depotService) GetDepots(userID int) ([]domain.Depot, error) {
	return s.depotRepo.FindDepotsByUser(userID)
}

// TODO: func (s *depotService) UpdateDepots(...

func (s *depotService) DeleteDepot(userID int, id int) error {
	existing, err := s.depotRepo.GetDepotByID(id)
	if err != nil || existing.UserID != userID {
		return domain.ErrUnauthorized
	}
	return s.depotRepo.DeleteDepot(id)
}
