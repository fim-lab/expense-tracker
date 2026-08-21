package services

import (
	"strings"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type stockService struct {
	stockRepo ports.StockRepository
	tradeRepo ports.TradeRepository
}

func NewStockService(stockRepo ports.StockRepository, tradeRepo ports.TradeRepository) ports.StockService {
	return &stockService{stockRepo: stockRepo, tradeRepo: tradeRepo}
}

func (s *stockService) GetStocks() ([]domain.Stock, error) {
	return s.stockRepo.FindAllStocks()
}

func (s *stockService) GetOrCreateByWKN(wkn string, fallbackPriceInCents int) (domain.Stock, error) {
	wkn = strings.ToUpper(strings.TrimSpace(wkn))
	if wkn == "" {
		return domain.Stock{}, domain.ErrMissingWKN
	}

	stock, err := s.stockRepo.FindStockByWKN(wkn)
	if err == nil {
		return stock, nil
	}
	if err != domain.ErrStockNotFound {
		return domain.Stock{}, err
	}

	stock = domain.Stock{WKN: wkn, Ticker: "", PriceInCents: fallbackPriceInCents}
	id, err := s.stockRepo.SaveStock(stock)
	if err != nil {
		return domain.Stock{}, err
	}
	stock.ID = id
	return stock, nil
}

func (s *stockService) CreateStock(stock domain.Stock) (domain.Stock, error) {
	stock.WKN = strings.ToUpper(strings.TrimSpace(stock.WKN))
	if stock.WKN == "" {
		return domain.Stock{}, domain.ErrMissingWKN
	}
	stock.ID = 0

	id, err := s.stockRepo.SaveStock(stock)
	if err != nil {
		return domain.Stock{}, err
	}
	stock.ID = id
	return stock, nil
}

func (s *stockService) UpdateStock(stock domain.Stock) (domain.Stock, error) {
	stock.WKN = strings.ToUpper(strings.TrimSpace(stock.WKN))
	if stock.WKN == "" {
		return domain.Stock{}, domain.ErrMissingWKN
	}

	if err := s.stockRepo.UpdateStock(stock); err != nil {
		return domain.Stock{}, err
	}
	return stock, nil
}

func (s *stockService) DeleteStock(id int) error {
	count, err := s.tradeRepo.CountTradesByStock(id)
	if err != nil {
		return err
	}
	if count > 0 {
		return domain.ErrNotEmpty
	}
	return s.stockRepo.DeleteStock(id)
}
