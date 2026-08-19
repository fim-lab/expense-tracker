package services

import (
	"testing"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
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
