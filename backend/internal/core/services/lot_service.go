package services

import (
	"errors"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type lotService struct {
	lotRepo   ports.LotRepository
	depotRepo ports.DepotRepository
}

func NewLotService(lotRepo ports.LotRepository, depotRepo ports.DepotRepository) ports.LotService {
	return &lotService{lotRepo: lotRepo, depotRepo: depotRepo}
}

func (s *lotService) CreateLot(userID int, lot domain.Lot) error {
	if lot.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	if lot.WKN == "" {
		return errors.New("WKN is required")
	}

	depot, err := s.depotRepo.GetDepotByID(lot.DepotID)
	if err != nil || depot.UserID != userID {
		return errors.New("invalid depot for lot")
	}

	lot.Remaining = lot.Amount

	return s.lotRepo.SaveLot(lot)
}

func (s *lotService) GetLots(userID int) ([]domain.Lot, error) {
	depots, err := s.depotRepo.FindDepotsByUser(userID)
	if err != nil {
		return nil, err
	}

	var allLots []domain.Lot
	for _, depot := range depots {
		lots, err := s.lotRepo.FindLotsByDepot(depot.ID)
		if err != nil {
			return nil, err
		}
		allLots = append(allLots, lots...)
	}

	return allLots, nil
}

func (s *lotService) DeleteLot(userID int, id int) error {
	existing, err := s.lotRepo.GetLotByID(id)
	if err != nil {
		return err
	}

	depot, err := s.depotRepo.GetDepotByID(existing.DepotID)
	if err != nil || depot.UserID != userID {
		return domain.ErrUnauthorized
	}

	return s.lotRepo.DeleteLot(id)
}
