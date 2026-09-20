package services

import (
	"log"
	"math"
	"strings"
	"time"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

// priceChangeThreshold is how far a freshly fetched price may differ from
// the stock's current price before RefreshStockPrice requires confirmation.
const priceChangeThreshold = 0.10

type stockService struct {
	stockRepo    ports.StockRepository
	tradeRepo    ports.TradeRepository
	priceFetcher ports.StockPriceFetcher
}

func NewStockService(stockRepo ports.StockRepository, tradeRepo ports.TradeRepository, priceFetcher ports.StockPriceFetcher) ports.StockService {
	return &stockService{stockRepo: stockRepo, tradeRepo: tradeRepo, priceFetcher: priceFetcher}
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

func (s *stockService) RefreshStockPrice(id int, confirmedPriceInCents *int) (domain.StockPriceRefresh, error) {
	stock, err := s.stockRepo.GetStockByID(id)
	if err != nil {
		return domain.StockPriceRefresh{}, err
	}
	if strings.TrimSpace(stock.Ticker) == "" {
		return domain.StockPriceRefresh{}, domain.ErrMissingTicker
	}

	newPriceInCents := 0
	if confirmedPriceInCents != nil {
		newPriceInCents = *confirmedPriceInCents
	} else {
		fetched, err := s.priceFetcher.FetchPrice(stock.Ticker)
		if err != nil {
			log.Printf("could not fetch price for stock %d (%s): %v", id, stock.Ticker, err)
			return domain.StockPriceRefresh{}, domain.ErrPriceFetchFailed
		}

		if priceChangeExceedsThreshold(stock.PriceInCents, fetched) {
			return domain.StockPriceRefresh{
				NeedsConfirmation: true,
				OldPriceInCents:   stock.PriceInCents,
				NewPriceInCents:   fetched,
			}, nil
		}
		newPriceInCents = fetched
	}

	stock.PriceInCents = newPriceInCents
	stock.LastFetched = time.Now().UTC()
	if err := s.stockRepo.UpdateStock(stock); err != nil {
		return domain.StockPriceRefresh{}, err
	}
	return domain.StockPriceRefresh{Stock: stock}, nil
}

func priceChangeExceedsThreshold(oldCents, newCents int) bool {
	if oldCents == 0 {
		return false
	}
	return math.Abs(float64(newCents-oldCents))/float64(oldCents) > priceChangeThreshold
}
