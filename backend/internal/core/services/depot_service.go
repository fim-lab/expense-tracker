package services

import (
	"errors"
	"strings"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type depotService struct {
	repo ports.ExpenseRepository
}

func NewDepotService(repo ports.ExpenseRepository) ports.DepotService {
	return &depotService{repo: repo}
}

func (s *depotService) CreateDepot(userID int, d domain.Depot) error {
	d.UserID = userID

	if strings.TrimSpace(d.Name) == "" {
		return errors.New("depot name is required")
	}

	wallet, err := s.repo.GetWalletByID(d.WalletID)
	if err != nil || wallet.UserID != userID {
		return errors.New("invalid wallet for depot")
	}

	return s.repo.SaveDepot(d)
}

func (s *depotService) GetDepots(userID int) ([]domain.Depot, error) {
	return s.repo.FindDepotsByUser(userID)
}

// TODO: func (s *depotService) UpdateDepots(...

func (s *depotService) DeleteDepot(userID int, id int) error {
	existing, err := s.repo.GetDepotByID(id)
	if err != nil || existing.UserID != userID {
		return domain.ErrUnauthorized
	}
	return s.repo.DeleteDepot(id)
}
