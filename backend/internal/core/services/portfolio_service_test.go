package services

import (
	"testing"
	"time"

	"github.com/fim-lab/expense-tracker/adapters/repository/memory"
	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

func TestPortfolioService_RealizedGainAndPositionAfterPartialSell(t *testing.T) {
	f := newStockFixture(t)
	f.mustBuy(t, 1, 10, 10000)
	f.mustBuy(t, 2, 5, 6000)
	sellID := f.mustSell(t, 3, 12, 15000)

	portfolio := f.mustGetPortfolio(t)
	if portfolio.RealizedGainInCents != 2600 {
		t.Errorf("expected a realized gain of 2600, got %d", portfolio.RealizedGainInCents)
	}
	if portfolio.InvestedInCents != 3600 {
		t.Errorf("expected 3600 still invested, got %d", portfolio.InvestedInCents)
	}
	if len(portfolio.Positions) != 1 {
		t.Fatalf("expected 1 position, got %d", len(portfolio.Positions))
	}
	position := portfolio.Positions[0]
	if position.Quantity != 3 || position.InvestedInCents != 3600 || position.AvgPriceInCents != 1200 {
		t.Errorf("expected 3 shares at 1200 each (3600 invested), got %+v", position)
	}

	trades, err := f.portfolioSvc.GetTrades(f.userID, f.depotID)
	if err != nil {
		t.Fatalf("could not read the trades: %v", err)
	}
	if len(trades) != 3 {
		t.Fatalf("expected 3 trades, got %d", len(trades))
	}
	if trades[0].ID != sellID {
		t.Errorf("expected the newest trade first, got trade %d", trades[0].ID)
	}
	if trades[0].CostBasisInCents != 12400 || trades[0].ProceedsInCents != 15000 || trades[0].RealizedGainInCents != 2600 {
		t.Errorf("expected the sell annotated with 12400/15000/2600, got %+v", trades[0])
	}
	for _, trade := range trades[1:] {
		if trade.RealizedGainInCents != 0 || trade.CostBasisInCents != 0 {
			t.Errorf("expected buys to carry no realized numbers, got %+v", trade)
		}
	}
}

func TestPortfolioService_ClosedPositionIsExcluded(t *testing.T) {
	f := newStockFixture(t)
	f.mustBuy(t, 1, 10, 10000)
	f.mustSell(t, 2, 10, 12000)

	portfolio := f.mustGetPortfolio(t)

	if len(portfolio.Positions) != 0 {
		t.Errorf("expected a fully sold instrument to have no position, got %+v", portfolio.Positions)
	}
	if portfolio.InvestedInCents != 0 {
		t.Errorf("expected nothing invested, got %d", portfolio.InvestedInCents)
	}
	if portfolio.RealizedGainInCents != 2000 {
		t.Errorf("expected the realized gain of 2000 to remain, got %d", portfolio.RealizedGainInCents)
	}
}

func TestPortfolioService_EmptyDepotReturnsEmptySlicesNotNil(t *testing.T) {
	f := newStockFixture(t)

	portfolio := f.mustGetPortfolio(t)
	if portfolio.Positions == nil {
		t.Error("expected an empty positions slice, got nil")
	}

	trades, err := f.portfolioSvc.GetTrades(f.userID, f.depotID)
	if err != nil {
		t.Fatalf("could not read the trades: %v", err)
	}
	if trades == nil {
		t.Error("expected an empty trades slice, got nil")
	}
	if len(trades) != 0 {
		t.Errorf("expected no trades, got %d", len(trades))
	}
}

func TestPortfolioService_OutputIsSortedDeterministically(t *testing.T) {
	f := newStockFixture(t)
	for i, wkn := range []string{"ZZZ999", "AAA111", "MMM555"} {
		trade := f.trade(domain.TradeTypeBuy, i+1, 1, 1000)
		trade.WKN = wkn
		if _, err := f.tradeSvc.CreateTrade(f.userID, trade); err != nil {
			t.Fatalf("buying %s failed: %v", wkn, err)
		}
	}
	// Two more lots of the same instrument, bought out of order.
	older := f.trade(domain.TradeTypeBuy, 4, 1, 1000)
	older.WKN = "AAA111"
	if _, err := f.tradeSvc.CreateTrade(f.userID, older); err != nil {
		t.Fatalf("buying the second AAA111 lot failed: %v", err)
	}

	portfolio := f.mustGetPortfolio(t)

	expectedOrder := []string{"AAA111", "MMM555", "ZZZ999"}
	if len(portfolio.Positions) != len(expectedOrder) {
		t.Fatalf("expected %d positions, got %d", len(expectedOrder), len(portfolio.Positions))
	}
	for i, wkn := range expectedOrder {
		if portfolio.Positions[i].WKN != wkn {
			t.Errorf("expected position %d to be %s, got %s", i, wkn, portfolio.Positions[i].WKN)
		}
	}

	lots := portfolio.Positions[0].Lots
	if len(lots) != 2 {
		t.Fatalf("expected 2 lots for AAA111, got %d", len(lots))
	}
	if lots[0].DateOfPurchase.After(lots[1].DateOfPurchase) {
		t.Errorf("expected the lots in purchase order, got %v before %v", lots[0].DateOfPurchase, lots[1].DateOfPurchase)
	}
}

func TestPortfolioService_TradesReportWhetherTheyCanBeDeleted(t *testing.T) {
	f := newStockFixture(t)
	soldBuyID := f.mustBuy(t, 1, 10, 100000)
	untouchedBuyID := f.mustBuy(t, 2, 5, 60000)
	sellID := f.mustSell(t, 3, 10, 120000)

	trades, err := f.portfolioSvc.GetTrades(f.userID, f.depotID)
	if err != nil {
		t.Fatalf("could not read the trades: %v", err)
	}

	canDelete := map[int]bool{}
	for _, trade := range trades {
		canDelete[trade.ID] = trade.CanDelete
	}

	if canDelete[soldBuyID] {
		t.Error("expected the buy whose shares were sold to be undeletable")
	}
	if !canDelete[untouchedBuyID] {
		t.Error("expected the untouched buy to be deletable")
	}
	if !canDelete[sellID] {
		t.Error("expected a sell to always be deletable")
	}

	// The flag must agree with what DeleteTrade actually does.
	if err := f.tradeSvc.DeleteTrade(f.userID, soldBuyID); err != domain.ErrInsufficientShares {
		t.Errorf("expected deleting the sold buy to fail, got %v", err)
	}
	if err := f.tradeSvc.DeleteTrade(f.userID, untouchedBuyID); err != nil {
		t.Errorf("expected deleting the untouched buy to succeed, got %v", err)
	}
}

type stalePriceFetcher struct {
	prices map[string]int
	errs   map[string]error
	calls  map[string]int
}

func newStalePriceFetcher() *stalePriceFetcher {
	return &stalePriceFetcher{prices: map[string]int{}, errs: map[string]error{}, calls: map[string]int{}}
}

func (f *stalePriceFetcher) FetchPrice(ticker string) (int, error) {
	f.calls[ticker]++
	if err, ok := f.errs[ticker]; ok {
		return 0, err
	}
	return f.prices[ticker], nil
}

type staleFixture struct {
	repos        ports.Repositories
	portfolioSvc ports.PortfolioService
	tradeSvc     ports.TradeService
	fetcher      *stalePriceFetcher
}

func newStaleFixture(t *testing.T) staleFixture {
	t.Helper()

	repos := memory.NewCleanRepositories()
	fetcher := newStalePriceFetcher()
	stockSvc := NewStockService(repos.StockRepository(), repos.TradeRepository(), fetcher)
	depotSvc := NewDepotService(repos.DepotRepository(), repos.WalletRepository(), repos.BudgetRepository(), repos.TradeRepository(), stockSvc)
	txSvc := NewTransactionService(repos.TransactionRepository(), repos.BudgetRepository(), repos.WalletRepository())

	return staleFixture{
		repos:        repos,
		portfolioSvc: NewPortfolioService(repos.TradeRepository(), depotSvc, stockSvc),
		tradeSvc:     NewTradeService(repos.TradeRepository(), depotSvc, txSvc, stockSvc),
		fetcher:      fetcher,
	}
}

func (f staleFixture) seedUser(t *testing.T, userID, walletID, budgetID, depotID int) {
	t.Helper()
	if err := f.repos.WalletRepository().SaveWallet(domain.Wallet{ID: walletID, UserID: userID, Name: "Wallet"}); err != nil {
		t.Fatalf("could not seed the wallet: %v", err)
	}
	if err := f.repos.BudgetRepository().SaveBudget(domain.Budget{ID: budgetID, UserID: userID, Name: "Investments", LimitCents: 1000000}); err != nil {
		t.Fatalf("could not seed the budget: %v", err)
	}
	if err := f.repos.DepotRepository().SaveDepot(domain.Depot{ID: depotID, UserID: userID, Name: "Depot", WalletID: walletID, BudgetID: budgetID}); err != nil {
		t.Fatalf("could not seed the depot: %v", err)
	}
}

func (f staleFixture) seedStock(t *testing.T, stock domain.Stock) domain.Stock {
	t.Helper()
	id, err := f.repos.StockRepository().SaveStock(stock)
	if err != nil {
		t.Fatalf("could not seed the stock: %v", err)
	}
	stock.ID = id
	return stock
}

func (f staleFixture) buyIntoDepot(t *testing.T, userID, depotID int, wkn string, day int) {
	t.Helper()
	trade := domain.Trade{
		DepotID:      depotID,
		WKN:          wkn,
		Type:         domain.TradeTypeBuy,
		Quantity:     1,
		TotalInCents: 1000,
		Timestamp:    tradeDay(day),
	}
	if _, err := f.tradeSvc.CreateTrade(userID, trade); err != nil {
		t.Fatalf("buying %s into depot %d failed: %v", wkn, depotID, err)
	}
}

func (f staleFixture) sellFromDepot(t *testing.T, userID, depotID int, wkn string, day int) {
	t.Helper()
	trade := domain.Trade{
		DepotID:      depotID,
		WKN:          wkn,
		Type:         domain.TradeTypeSell,
		Quantity:     1,
		TotalInCents: 1000,
		Timestamp:    tradeDay(day),
	}
	if _, err := f.tradeSvc.CreateTrade(userID, trade); err != nil {
		t.Fatalf("selling %s from depot %d failed: %v", wkn, depotID, err)
	}
}

func (f staleFixture) mustGetStock(t *testing.T, id int) domain.Stock {
	t.Helper()
	stock, err := f.repos.StockRepository().GetStockByID(id)
	if err != nil {
		t.Fatalf("could not read stock %d: %v", id, err)
	}
	return stock
}

func TestPortfolioService_RefreshStaleStockPrices(t *testing.T) {
	t.Run("stale stock within threshold is refreshed", func(t *testing.T) {
		f := newStaleFixture(t)
		f.seedUser(t, 1, 1, 1, 1)
		stock := f.seedStock(t, domain.Stock{WKN: "A1", Ticker: "AAA", PriceInCents: 1000, LastFetched: time.Now().Add(-48 * time.Hour)})
		f.buyIntoDepot(t, 1, 1, "A1", 1)
		f.fetcher.prices["AAA"] = 1050 // +5%, within the 10% auto-save threshold

		if err := f.portfolioSvc.RefreshStaleStockPrices(1); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		updated := f.mustGetStock(t, stock.ID)
		if updated.PriceInCents != 1050 {
			t.Errorf("expected the price to be refreshed to 1050, got %d", updated.PriceInCents)
		}
		if time.Since(updated.LastFetched) > time.Minute {
			t.Errorf("expected LastFetched to be bumped to now, got %v", updated.LastFetched)
		}
		if f.fetcher.calls["AAA"] != 1 {
			t.Errorf("expected the fetcher to be called once, got %d calls", f.fetcher.calls["AAA"])
		}
	})

	t.Run("fresh stock is left untouched", func(t *testing.T) {
		f := newStaleFixture(t)
		f.seedUser(t, 1, 1, 1, 1)
		lastFetched := time.Now().Add(-1 * time.Hour)
		stock := f.seedStock(t, domain.Stock{WKN: "A1", Ticker: "AAA", PriceInCents: 1000, LastFetched: lastFetched})
		f.buyIntoDepot(t, 1, 1, "A1", 1)
		f.fetcher.prices["AAA"] = 2000

		if err := f.portfolioSvc.RefreshStaleStockPrices(1); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		updated := f.mustGetStock(t, stock.ID)
		if updated.PriceInCents != 1000 || !updated.LastFetched.Equal(lastFetched) {
			t.Errorf("expected the fresh stock to be untouched, got %+v", updated)
		}
		if f.fetcher.calls["AAA"] != 0 {
			t.Errorf("expected the fetcher not to be called, got %d calls", f.fetcher.calls["AAA"])
		}
	})

	t.Run("stock needing confirmation is left stale with no error", func(t *testing.T) {
		f := newStaleFixture(t)
		f.seedUser(t, 1, 1, 1, 1)
		lastFetched := time.Now().Add(-48 * time.Hour)
		stock := f.seedStock(t, domain.Stock{WKN: "A1", Ticker: "AAA", PriceInCents: 1000, LastFetched: lastFetched})
		f.buyIntoDepot(t, 1, 1, "A1", 1)
		f.fetcher.prices["AAA"] = 1300 // +30%, needs confirmation

		if err := f.portfolioSvc.RefreshStaleStockPrices(1); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		updated := f.mustGetStock(t, stock.ID)
		if updated.PriceInCents != 1000 || !updated.LastFetched.Equal(lastFetched) {
			t.Errorf("expected the stock to be left stale pending manual confirmation, got %+v", updated)
		}
	})

	t.Run("another user's stock is never touched", func(t *testing.T) {
		f := newStaleFixture(t)
		f.seedUser(t, 1, 1, 1, 1)
		f.seedUser(t, 2, 2, 2, 2)
		mine := f.seedStock(t, domain.Stock{WKN: "A1", Ticker: "AAA", PriceInCents: 1000, LastFetched: time.Now().Add(-48 * time.Hour)})
		theirs := f.seedStock(t, domain.Stock{WKN: "B1", Ticker: "BBB", PriceInCents: 1000, LastFetched: time.Now().Add(-48 * time.Hour)})
		f.buyIntoDepot(t, 1, 1, "A1", 1)
		f.buyIntoDepot(t, 2, 2, "B1", 1)
		f.fetcher.prices["AAA"] = 1050
		f.fetcher.prices["BBB"] = 1050

		if err := f.portfolioSvc.RefreshStaleStockPrices(1); err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if updated := f.mustGetStock(t, mine.ID); updated.PriceInCents != 1050 {
			t.Errorf("expected the caller's own stock to be refreshed, got %+v", updated)
		}
		if updated := f.mustGetStock(t, theirs.ID); updated.PriceInCents != 1000 {
			t.Errorf("expected the other user's stock to be untouched, got %+v", updated)
		}
		if f.fetcher.calls["BBB"] != 0 {
			t.Errorf("expected the fetcher not to be called for the other user's stock, got %d calls", f.fetcher.calls["BBB"])
		}
	})

	t.Run("one stock's fetch failure does not block the others", func(t *testing.T) {
		f := newStaleFixture(t)
		f.seedUser(t, 1, 1, 1, 1)
		failing := f.seedStock(t, domain.Stock{WKN: "A1", Ticker: "AAA", PriceInCents: 1000, LastFetched: time.Now().Add(-48 * time.Hour)})
		succeeding := f.seedStock(t, domain.Stock{WKN: "B1", Ticker: "BBB", PriceInCents: 1000, LastFetched: time.Now().Add(-48 * time.Hour)})
		f.buyIntoDepot(t, 1, 1, "A1", 1)
		f.buyIntoDepot(t, 1, 1, "B1", 2)
		f.fetcher.errs["AAA"] = domain.ErrPriceFetchFailed
		f.fetcher.prices["BBB"] = 1050

		if err := f.portfolioSvc.RefreshStaleStockPrices(1); err != nil {
			t.Fatalf("expected the pass to succeed despite one fetch failing, got %v", err)
		}

		if updated := f.mustGetStock(t, failing.ID); updated.PriceInCents != 1000 {
			t.Errorf("expected the failing stock to be left unchanged, got %+v", updated)
		}
		if updated := f.mustGetStock(t, succeeding.ID); updated.PriceInCents != 1050 {
			t.Errorf("expected the succeeding stock to still be refreshed, got %+v", updated)
		}
	})
}

func TestPortfolioService_GetOwnedStocks(t *testing.T) {
	t.Run("only stocks with an open position are returned", func(t *testing.T) {
		f := newStaleFixture(t)
		f.seedUser(t, 1, 1, 1, 1)
		held := f.seedStock(t, domain.Stock{WKN: "A1", Ticker: "AAA", PriceInCents: 1000})
		closed := f.seedStock(t, domain.Stock{WKN: "B1", Ticker: "BBB", PriceInCents: 1000})
		f.buyIntoDepot(t, 1, 1, "A1", 1)
		f.buyIntoDepot(t, 1, 1, "B1", 1)
		f.sellFromDepot(t, 1, 1, "B1", 2)

		owned, err := f.portfolioSvc.GetOwnedStocks(1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(owned) != 1 || owned[0].ID != held.ID {
			t.Errorf("expected only the still-held stock %d, got %+v (closed stock was %d)", held.ID, owned, closed.ID)
		}
	})

	t.Run("another user's holdings are never returned", func(t *testing.T) {
		f := newStaleFixture(t)
		f.seedUser(t, 1, 1, 1, 1)
		f.seedUser(t, 2, 2, 2, 2)
		f.seedStock(t, domain.Stock{WKN: "A1", Ticker: "AAA", PriceInCents: 1000})
		f.seedStock(t, domain.Stock{WKN: "B1", Ticker: "BBB", PriceInCents: 1000})
		f.buyIntoDepot(t, 1, 1, "A1", 1)
		f.buyIntoDepot(t, 2, 2, "B1", 1)

		owned, err := f.portfolioSvc.GetOwnedStocks(1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}

		if len(owned) != 1 || owned[0].WKN != "A1" {
			t.Errorf("expected only the caller's own holding, got %+v", owned)
		}
	})

	t.Run("no depots returns an empty slice, not nil", func(t *testing.T) {
		f := newStaleFixture(t)
		f.seedUser(t, 1, 1, 1, 1)

		owned, err := f.portfolioSvc.GetOwnedStocks(1)
		if err != nil {
			t.Fatalf("expected no error, got %v", err)
		}
		if owned == nil {
			t.Error("expected an empty slice, got nil")
		}
		if len(owned) != 0 {
			t.Errorf("expected no owned stocks, got %+v", owned)
		}
	})
}
