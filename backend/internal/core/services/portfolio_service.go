package services

import (
	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type portfolioService struct {
	tradeRepo    ports.TradeRepository
	depotService ports.DepotService
}

func NewPortfolioService(tradeRepo ports.TradeRepository, depotService ports.DepotService) ports.PortfolioService {
	return &portfolioService{tradeRepo: tradeRepo, depotService: depotService}
}

func (s *portfolioService) GetPortfolio(userID int, depotID int) (domain.Portfolio, error) {
	trades, err := s.tradesOfDepot(userID, depotID)
	if err != nil {
		return domain.Portfolio{}, err
	}

	snapshot := buildPortfolio(trades)
	portfolio := domain.Portfolio{
		DepotID:             depotID,
		Positions:           snapshot.positions(depotID),
		RealizedGainInCents: snapshot.realizedGain(),
	}
	for _, position := range portfolio.Positions {
		portfolio.InvestedInCents += position.InvestedInCents
	}
	return portfolio, nil
}

func (s *portfolioService) GetTrades(userID int, depotID int) ([]domain.TradeDTO, error) {
	trades, err := s.tradesOfDepot(userID, depotID)
	if err != nil {
		return nil, err
	}
	return buildPortfolio(trades).tradeDTOs(trades), nil
}

func (s *portfolioService) tradesOfDepot(userID int, depotID int) ([]domain.Trade, error) {
	if _, err := s.depotService.GetDepotByID(userID, depotID); err != nil {
		return nil, err
	}
	return s.tradeRepo.FindTradesByDepot(depotID)
}
