package services

import (
	"errors"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type stockService struct {
	repo ports.ExpenseRepository
}

func NewStockService(repo ports.ExpenseRepository) ports.StockService {
	return &stockService{repo: repo}
}

func (s *stockService) CreateStock(userID int, st domain.Stock) error {
	st.UserID = userID

	if st.Amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	if st.WKN == "" {
		return errors.New("WKN is required")
	}

	depot, err := s.repo.GetDepotByID(st.DepotID)
	if err != nil || depot.UserID != userID {
		return errors.New("invalid depot for stock")
	}

	return s.repo.SaveStock(st)
}

func (s *stockService) GetStocks(userID int) ([]domain.Stock, error) {
	return s.repo.FindStocksByUser(userID)
}

// TODO: func (s *stockService) UpdateStocks(...

func (s *stockService) DeleteStock(userID int, id int) error {
	existing, err := s.repo.GetStockByID(id)
	if err != nil || existing.UserID != userID {
		return domain.ErrUnauthorized
	}
	return s.repo.DeleteStock(id)
}
