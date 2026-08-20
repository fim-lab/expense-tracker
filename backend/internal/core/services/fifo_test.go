package services

import (
	"testing"
	"time"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

const testWKN = "A1JX52"
const testStockID = 1

func tradeDay(day int) time.Time {
	return time.Date(2026, time.March, day, 12, 0, 0, 0, time.UTC)
}

func buyTrade(id int, day int, quantity float64, totalInCents int) domain.Trade {
	return domain.Trade{
		ID:           id,
		DepotID:      1,
		WKN:          testWKN,
		StockID:      testStockID,
		Type:         domain.TradeTypeBuy,
		Quantity:     quantity,
		TotalInCents: totalInCents,
		Timestamp:    tradeDay(day),
	}
}

func sellTrade(id int, day int, quantity float64, totalInCents int) domain.Trade {
	return domain.Trade{
		ID:           id,
		DepotID:      1,
		WKN:          testWKN,
		StockID:      testStockID,
		Type:         domain.TradeTypeSell,
		Quantity:     quantity,
		TotalInCents: totalInCents,
		Timestamp:    tradeDay(day),
	}
}

func TestFIFO_SameDayBuyThenSellIsCovered(t *testing.T) {
	buy := buyTrade(1, 1, 10, 10000)
	sell := sellTrade(0, 1, 10, 12000)

	if err := validateTradeHistory([]domain.Trade{buy, sell}); err != nil {
		t.Errorf("expected a same-day buy and sell to be covered, got %v", err)
	}

	snapshot := buildPortfolio([]domain.Trade{buy, sell})
	if len(snapshot.openLots) != 0 {
		t.Errorf("expected no open lots after selling everything, got %d", len(snapshot.openLots))
	}
}

func TestFIFO_PartialSellAcrossTwoLots(t *testing.T) {
	trades := []domain.Trade{
		buyTrade(1, 1, 10, 10000),
		buyTrade(2, 2, 5, 6000),
		sellTrade(3, 3, 12, 15000),
	}

	snapshot := buildPortfolio(trades)

	if len(snapshot.allocations) != 2 {
		t.Fatalf("expected the sell to consume 2 lots, got %d", len(snapshot.allocations))
	}
	first := snapshot.allocations[0]
	if first.BuyTradeID != 1 || first.Quantity != 10 || first.CostBasisInCents != 10000 {
		t.Errorf("expected the oldest lot to be consumed completely, got %+v", first)
	}
	second := snapshot.allocations[1]
	if second.BuyTradeID != 2 || second.Quantity != 2 || second.CostBasisInCents != 2400 {
		t.Errorf("expected 2 shares of the second lot at 2400 cents, got %+v", second)
	}

	if len(snapshot.openLots) != 1 {
		t.Fatalf("expected 1 open lot, got %d", len(snapshot.openLots))
	}
	openLot := snapshot.openLots[0]
	if openLot.TradeID != 2 || openLot.Remaining != 3 || openLot.RemainingCostInCents != 3600 {
		t.Errorf("expected 3 shares of lot 2 left at 3600 cents, got %+v", openLot)
	}
}

func TestFIFO_ExactLotEmptyingSell(t *testing.T) {
	trades := []domain.Trade{
		buyTrade(1, 1, 10, 10000),
		sellTrade(2, 2, 10, 12000),
	}

	snapshot := buildPortfolio(trades)

	if len(snapshot.openLots) != 0 {
		t.Errorf("expected no open lots, got %d", len(snapshot.openLots))
	}
	if len(snapshot.allocations) != 1 {
		t.Fatalf("expected 1 allocation, got %d", len(snapshot.allocations))
	}
	allocation := snapshot.allocations[0]
	if allocation.CostBasisInCents != 10000 {
		t.Errorf("expected the full cost basis of 10000, got %d", allocation.CostBasisInCents)
	}
	if allocation.ProceedsInCents != 12000 || allocation.RealizedGainInCents != 2000 {
		t.Errorf("expected proceeds 12000 and gain 2000, got %d and %d", allocation.ProceedsInCents, allocation.RealizedGainInCents)
	}
}

func TestFIFO_CostBasisSumsExactly(t *testing.T) {
	trades := []domain.Trade{
		buyTrade(1, 1, 3, 1000),
		sellTrade(2, 2, 1, 400),
		sellTrade(3, 3, 1, 400),
		sellTrade(4, 4, 1, 400),
	}

	snapshot := buildPortfolio(trades)

	if len(snapshot.allocations) != 3 {
		t.Fatalf("expected 3 allocations, got %d", len(snapshot.allocations))
	}
	var totalCostBasis int
	for _, allocation := range snapshot.allocations {
		totalCostBasis += allocation.CostBasisInCents
	}
	if totalCostBasis != 1000 {
		t.Errorf("expected the allocated cost basis to add up to 1000, got %d", totalCostBasis)
	}
	if len(snapshot.openLots) != 0 {
		t.Errorf("expected no open lots, got %d", len(snapshot.openLots))
	}
}

func TestFIFO_PartiallyConsumedLotCostInvariant(t *testing.T) {
	trades := []domain.Trade{
		buyTrade(1, 1, 3, 1000),
		sellTrade(2, 2, 1, 400),
	}

	snapshot := buildPortfolio(trades)

	if len(snapshot.allocations) != 1 || len(snapshot.openLots) != 1 {
		t.Fatalf("expected 1 allocation and 1 open lot, got %d and %d", len(snapshot.allocations), len(snapshot.openLots))
	}
	allocated := snapshot.allocations[0].CostBasisInCents
	remaining := snapshot.openLots[0].RemainingCostInCents
	if allocated+remaining != 1000 {
		t.Errorf("expected allocated (%d) + remaining (%d) to equal the lot total 1000", allocated, remaining)
	}
}

func TestFIFO_FloatResidualClampedToZero(t *testing.T) {
	trades := []domain.Trade{
		buyTrade(1, 1, 0.3, 3000),
		sellTrade(2, 2, 0.1, 1100),
		sellTrade(3, 3, 0.1, 1100),
		sellTrade(4, 4, 0.1, 1100),
	}

	if err := validateTradeHistory(trades); err != nil {
		t.Errorf("expected selling every fractional share to be covered, got %v", err)
	}

	snapshot := buildPortfolio(trades)
	if len(snapshot.openLots) != 0 {
		t.Errorf("expected no dust lot to survive, got %+v", snapshot.openLots)
	}
	if len(snapshot.unmatched) != 0 {
		t.Errorf("expected nothing unmatched, got %+v", snapshot.unmatched)
	}
}

func TestFIFO_SeparateWKNsDoNotCrossFeed(t *testing.T) {
	other := sellTrade(2, 2, 5, 6000)
	other.StockID = 2
	trades := []domain.Trade{buyTrade(1, 1, 10, 10000), other}

	if err := validateTradeHistory(trades); err != domain.ErrInsufficientShares {
		t.Errorf("expected ErrInsufficientShares when selling an instrument that was never bought, got %v", err)
	}

	snapshot := buildPortfolio(trades)
	if len(snapshot.openLots) != 1 || snapshot.openLots[0].Remaining != 10 {
		t.Errorf("expected the untouched lot to keep all 10 shares, got %+v", snapshot.openLots)
	}
}

func TestFIFO_SellBeforeAnyBuyOfWKN(t *testing.T) {
	trades := []domain.Trade{
		sellTrade(1, 1, 5, 6000),
		buyTrade(2, 5, 5, 5000),
	}

	if err := validateTradeHistory(trades); err != domain.ErrInsufficientShares {
		t.Errorf("expected ErrInsufficientShares for a sell that predates every buy, got %v", err)
	}

	snapshot := buildPortfolio(trades)
	if snapshot.unmatched[testStockID] != 5 {
		t.Errorf("expected 5 unmatched shares, got %v", snapshot.unmatched[testStockID])
	}
	if snapshot.realizedGain() != 0 {
		t.Errorf("expected no realized gain from an uncovered sell, got %d", snapshot.realizedGain())
	}
	positions := snapshot.positions(1)
	if len(positions) != 1 || positions[0].Quantity != 5 {
		t.Errorf("expected the later buy to be a position of 5 shares, got %+v", positions)
	}
}

func TestFIFO_UnsortedInputProducesSameResult(t *testing.T) {
	ordered := []domain.Trade{
		buyTrade(1, 1, 10, 10000),
		buyTrade(2, 2, 5, 6000),
		sellTrade(3, 3, 12, 15000),
	}
	shuffled := []domain.Trade{ordered[2], ordered[0], ordered[1]}

	expected := buildPortfolio(ordered)
	actual := buildPortfolio(shuffled)

	if len(expected.allocations) != len(actual.allocations) {
		t.Fatalf("expected %d allocations, got %d", len(expected.allocations), len(actual.allocations))
	}
	for i := range expected.allocations {
		if expected.allocations[i] != actual.allocations[i] {
			t.Errorf("allocation %d differs: %+v vs %+v", i, expected.allocations[i], actual.allocations[i])
		}
	}
	if expected.realizedGain() != actual.realizedGain() {
		t.Errorf("expected realized gain %d, got %d", expected.realizedGain(), actual.realizedGain())
	}
}
