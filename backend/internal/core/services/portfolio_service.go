package services

import (
	"math"
	"sort"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type portfolioService struct {
	tradeRepo    ports.TradeRepository
	depotService ports.DepotService
	stockService ports.StockService
}

func NewPortfolioService(tradeRepo ports.TradeRepository, depotService ports.DepotService, stockService ports.StockService) ports.PortfolioService {
	return &portfolioService{tradeRepo: tradeRepo, depotService: depotService, stockService: stockService}
}

func (s *portfolioService) stocksByID() (map[int]domain.Stock, error) {
	stocks, err := s.stockService.GetStocks()
	if err != nil {
		return nil, err
	}
	byID := make(map[int]domain.Stock, len(stocks))
	for _, stock := range stocks {
		byID[stock.ID] = stock
	}
	return byID, nil
}

func (s *portfolioService) GetPortfolio(userID int, depotID int) (domain.Portfolio, error) {
	trades, err := s.tradesOfDepot(userID, depotID)
	if err != nil {
		return domain.Portfolio{}, err
	}

	stocksByID, err := s.stocksByID()
	if err != nil {
		return domain.Portfolio{}, err
	}

	snapshot := buildPortfolio(trades)
	positions := snapshot.positions(depotID)
	for i := range positions {
		position := &positions[i]
		stock := stocksByID[position.StockID]
		position.WKN = stock.WKN
		position.Ticker = stock.Ticker
		position.CurrentPriceInCents = stock.PriceInCents
		position.CurrentValueInCents = int(math.Round(position.Quantity * float64(stock.PriceInCents)))
		position.UnrealizedGainInCents = position.CurrentValueInCents - position.InvestedInCents
	}
	sort.Slice(positions, func(i, j int) bool { return positions[i].WKN < positions[j].WKN })

	portfolio := domain.Portfolio{
		DepotID:             depotID,
		Positions:           positions,
		RealizedGainInCents: snapshot.realizedGain(),
	}
	for _, position := range portfolio.Positions {
		portfolio.InvestedInCents += position.InvestedInCents
		portfolio.CurrentValueInCents += position.CurrentValueInCents
	}
	portfolio.UnrealizedGainInCents = portfolio.CurrentValueInCents - portfolio.InvestedInCents
	return portfolio, nil
}

func (s *portfolioService) GetTrades(userID int, depotID int) ([]domain.TradeDTO, error) {
	trades, err := s.tradesOfDepot(userID, depotID)
	if err != nil {
		return nil, err
	}

	stocksByID, err := s.stocksByID()
	if err != nil {
		return nil, err
	}

	dtos := buildPortfolio(trades).tradeDTOs(trades)
	for i := range dtos {
		dtos[i].WKN = stocksByID[dtos[i].StockID].WKN
	}
	return dtos, nil
}

func (s *portfolioService) tradesOfDepot(userID int, depotID int) ([]domain.Trade, error) {
	if _, err := s.depotService.GetDepotByID(userID, depotID); err != nil {
		return nil, err
	}
	return s.tradeRepo.FindTradesByDepot(depotID)
}
