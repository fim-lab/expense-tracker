package services

import (
	"testing"

	"github.com/fim-lab/expense-tracker/adapters/repository/memory"
	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type fakePriceFetcher struct {
	priceInCents int
	err          error
	calls        int
}

func (f *fakePriceFetcher) FetchPrice(ticker string) (int, error) {
	f.calls++
	if f.err != nil {
		return 0, f.err
	}
	return f.priceInCents, nil
}

func TestRefreshStockPrice(t *testing.T) {
	newStock := func(t *testing.T, repos ports.Repositories, stock domain.Stock) domain.Stock {
		t.Helper()
		id, err := repos.StockRepository().SaveStock(stock)
		if err != nil {
			t.Fatalf("could not seed the stock: %v", err)
		}
		stock.ID = id
		return stock
	}

	t.Run("missing ticker is rejected", func(t *testing.T) {
		repos := memory.NewCleanRepositories()
		fetcher := &fakePriceFetcher{}
		svc := NewStockService(repos.StockRepository(), repos.TradeRepository(), fetcher)

		stock := newStock(t, repos, domain.Stock{WKN: "A1", Ticker: "", PriceInCents: 1000})

		_, err := svc.RefreshStockPrice(stock.ID, nil)
		if err != domain.ErrMissingTicker {
			t.Fatalf("expected ErrMissingTicker, got %v", err)
		}
		if fetcher.calls != 0 {
			t.Errorf("expected the fetcher not to be called, got %d calls", fetcher.calls)
		}
	})

	t.Run("fetch failure changes nothing", func(t *testing.T) {
		repos := memory.NewCleanRepositories()
		fetcher := &fakePriceFetcher{err: domain.ErrPriceFetchFailed}
		svc := NewStockService(repos.StockRepository(), repos.TradeRepository(), fetcher)

		stock := newStock(t, repos, domain.Stock{WKN: "A1", Ticker: "AAPL", PriceInCents: 1000})

		_, err := svc.RefreshStockPrice(stock.ID, nil)
		if err != domain.ErrPriceFetchFailed {
			t.Fatalf("expected ErrPriceFetchFailed, got %v", err)
		}

		saved, _ := repos.StockRepository().GetStockByID(stock.ID)
		if saved.PriceInCents != 1000 || !saved.LastFetched.IsZero() {
			t.Errorf("expected the stock to be unchanged, got %+v", saved)
		}
	})

	t.Run("price change over 10 percent needs confirmation and saves nothing", func(t *testing.T) {
		repos := memory.NewCleanRepositories()
		fetcher := &fakePriceFetcher{priceInCents: 1200} // +20% vs 1000
		svc := NewStockService(repos.StockRepository(), repos.TradeRepository(), fetcher)

		stock := newStock(t, repos, domain.Stock{WKN: "A1", Ticker: "AAPL", PriceInCents: 1000})

		result, err := svc.RefreshStockPrice(stock.ID, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if !result.NeedsConfirmation || result.OldPriceInCents != 1000 || result.NewPriceInCents != 1200 {
			t.Errorf("expected confirmation for 1000 -> 1200, got %+v", result)
		}

		saved, _ := repos.StockRepository().GetStockByID(stock.ID)
		if saved.PriceInCents != 1000 || !saved.LastFetched.IsZero() {
			t.Errorf("expected the stock to be unchanged pending confirmation, got %+v", saved)
		}
	})

	t.Run("price change within 10 percent saves automatically", func(t *testing.T) {
		repos := memory.NewCleanRepositories()
		fetcher := &fakePriceFetcher{priceInCents: 1050} // +5% vs 1000
		svc := NewStockService(repos.StockRepository(), repos.TradeRepository(), fetcher)

		stock := newStock(t, repos, domain.Stock{WKN: "A1", Ticker: "AAPL", PriceInCents: 1000})

		result, err := svc.RefreshStockPrice(stock.ID, nil)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.NeedsConfirmation {
			t.Fatalf("expected no confirmation to be required, got %+v", result)
		}
		if result.Stock.PriceInCents != 1050 || result.Stock.LastFetched.IsZero() {
			t.Errorf("expected the stock to be saved with the new price and a fetch time, got %+v", result.Stock)
		}
	})

	t.Run("confirmed price is saved without calling the fetcher again", func(t *testing.T) {
		repos := memory.NewCleanRepositories()
		fetcher := &fakePriceFetcher{}
		svc := NewStockService(repos.StockRepository(), repos.TradeRepository(), fetcher)

		stock := newStock(t, repos, domain.Stock{WKN: "A1", Ticker: "AAPL", PriceInCents: 1000})

		confirmed := 1200
		result, err := svc.RefreshStockPrice(stock.ID, &confirmed)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if result.NeedsConfirmation || result.Stock.PriceInCents != 1200 {
			t.Errorf("expected the confirmed price to be saved directly, got %+v", result)
		}
		if fetcher.calls != 0 {
			t.Errorf("expected the fetcher not to be called for a confirmed price, got %d calls", fetcher.calls)
		}
	})
}
