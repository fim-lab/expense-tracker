package services

import (
	"math"
	"sort"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

const quantityEpsilon = 1e-9

func epsilonFor(values ...float64) float64 {
	largest := 1.0
	for _, v := range values {
		if math.Abs(v) > largest {
			largest = math.Abs(v)
		}
	}
	return quantityEpsilon * largest
}

func clampQuantity(q float64) float64 {
	if math.Abs(q) < epsilonFor(q) {
		return 0
	}
	return q
}

type PortfolioSnapshot struct {
	openLots    []domain.Lot
	allocations []domain.SellAllocation
	unmatched   map[int]float64
}

func buildPortfolio(trades []domain.Trade) PortfolioSnapshot {
	snapshot := PortfolioSnapshot{unmatched: map[int]float64{}}

	lotsByStockID := map[int][]*domain.Lot{}
	var lots []*domain.Lot

	for _, t := range sortTradesChronologically(trades) {
		switch t.Type {
		case domain.TradeTypeBuy:
			lot := &domain.Lot{
				TradeID:              t.ID,
				DepotID:              t.DepotID,
				StockID:              t.StockID,
				DateOfPurchase:       t.Timestamp,
				Quantity:             t.Quantity,
				Remaining:            t.Quantity,
				TotalInCents:         t.TotalInCents,
				RemainingCostInCents: t.TotalInCents,
			}
			lots = append(lots, lot)
			lotsByStockID[t.StockID] = append(lotsByStockID[t.StockID], lot)
		case domain.TradeTypeSell:
			snapshot.applySell(t, lotsByStockID[t.StockID])
		}
	}

	for _, lot := range lots {
		if lot.Remaining > epsilonFor(lot.Quantity) {
			snapshot.openLots = append(snapshot.openLots, *lot)
		}
	}
	sort.SliceStable(snapshot.openLots, func(i, j int) bool {
		a, b := snapshot.openLots[i], snapshot.openLots[j]
		if !a.DateOfPurchase.Equal(b.DateOfPurchase) {
			return a.DateOfPurchase.Before(b.DateOfPurchase)
		}
		return a.TradeID < b.TradeID
	})

	return snapshot
}

func (s *PortfolioSnapshot) applySell(sell domain.Trade, lots []*domain.Lot) {
	toSell := sell.Quantity
	remainingProceeds := sell.TotalInCents

	for _, lot := range lots {
		if toSell <= 0 {
			break
		}
		if lot.Remaining <= 0 {
			continue
		}

		take := math.Min(lot.Remaining, toSell)
		emptiesLot := take == lot.Remaining
		completesSell := take == toSell

		costBasis := lot.RemainingCostInCents
		if !emptiesLot {
			costBasis = int(math.Round(float64(lot.TotalInCents) * take / lot.Quantity))
			if costBasis > lot.RemainingCostInCents {
				costBasis = lot.RemainingCostInCents
			}
		}

		proceeds := remainingProceeds
		if !completesSell {
			proceeds = int(math.Round(float64(sell.TotalInCents) * take / sell.Quantity))
			if proceeds > remainingProceeds {
				proceeds = remainingProceeds
			}
		}

		lot.Remaining = clampQuantity(lot.Remaining - take)
		lot.RemainingCostInCents -= costBasis
		toSell = clampQuantity(toSell - take)
		remainingProceeds -= proceeds

		s.allocations = append(s.allocations, domain.SellAllocation{
			SellTradeID:         sell.ID,
			BuyTradeID:          lot.TradeID,
			StockID:             sell.StockID,
			Quantity:            take,
			CostBasisInCents:    costBasis,
			ProceedsInCents:     proceeds,
			RealizedGainInCents: proceeds - costBasis,
			BuyDate:             lot.DateOfPurchase,
			SellDate:            sell.Timestamp,
		})
	}

	if toSell > 0 {
		s.unmatched[sell.StockID] += toSell
	}
}

func sortTradesChronologically(trades []domain.Trade) []domain.Trade {
	sorted := append([]domain.Trade(nil), trades...)
	sort.SliceStable(sorted, func(i, j int) bool {
		a, b := sorted[i], sorted[j]
		if !a.Timestamp.Equal(b.Timestamp) {
			return a.Timestamp.Before(b.Timestamp)
		}
		if tradeTypeRank(a.Type) != tradeTypeRank(b.Type) {
			return tradeTypeRank(a.Type) < tradeTypeRank(b.Type)
		}
		return tradeOrderID(a.ID) < tradeOrderID(b.ID)
	})
	return sorted
}

func tradeTypeRank(t domain.TradeType) int {
	if t == domain.TradeTypeSell {
		return 1
	}
	return 0
}

func tradeOrderID(id int) int {
	if id == 0 {
		return math.MaxInt
	}
	return id
}

func validateTradeHistory(trades []domain.Trade) error {
	snapshot := buildPortfolio(trades)
	for _, quantity := range snapshot.unmatched {
		if quantity > quantityEpsilon {
			return domain.ErrInsufficientShares
		}
	}
	return nil
}

func (s PortfolioSnapshot) positions(depotID int) []domain.Position {
	byStockID := map[int]*domain.Position{}
	var order []int

	for _, lot := range s.openLots {
		position, ok := byStockID[lot.StockID]
		if !ok {
			position = &domain.Position{DepotID: depotID, StockID: lot.StockID, Lots: []domain.Lot{}}
			byStockID[lot.StockID] = position
			order = append(order, lot.StockID)
		}
		position.Quantity += lot.Remaining
		position.InvestedInCents += lot.RemainingCostInCents
		position.Lots = append(position.Lots, lot)
	}

	sort.Ints(order)
	positions := make([]domain.Position, 0, len(order))
	for _, stockID := range order {
		position := byStockID[stockID]
		position.Quantity = clampQuantity(position.Quantity)
		if position.Quantity <= 0 {
			continue
		}
		position.AvgPriceInCents = int(math.Round(float64(position.InvestedInCents) / position.Quantity))
		positions = append(positions, *position)
	}
	return positions
}

func currentValueInCents(positions []domain.Position, priceByStockID map[int]int) int {
	var total int
	for _, position := range positions {
		total += int(math.Round(position.Quantity * float64(priceByStockID[position.StockID])))
	}
	return total
}

func investedInCents(trades []domain.Trade) int {
	var total int
	for _, lot := range buildPortfolio(trades).openLots {
		total += lot.RemainingCostInCents
	}
	return total
}

func (s PortfolioSnapshot) realizedGain() int {
	var total int
	for _, allocation := range s.allocations {
		total += allocation.RealizedGainInCents
	}
	return total
}

func canDeleteTrade(trades []domain.Trade, t domain.Trade) bool {
	if t.Type != domain.TradeTypeBuy {
		return true
	}
	remaining := make([]domain.Trade, 0, len(trades))
	for _, other := range trades {
		if other.ID != t.ID {
			remaining = append(remaining, other)
		}
	}
	return validateTradeHistory(remaining) == nil
}

func (s PortfolioSnapshot) tradeDTOs(trades []domain.Trade) []domain.TradeDTO {
	costByTrade := map[int]int{}
	proceedsByTrade := map[int]int{}
	for _, allocation := range s.allocations {
		costByTrade[allocation.SellTradeID] += allocation.CostBasisInCents
		proceedsByTrade[allocation.SellTradeID] += allocation.ProceedsInCents
	}

	sorted := sortTradesChronologically(trades)
	dtos := make([]domain.TradeDTO, 0, len(sorted))
	for i := len(sorted) - 1; i >= 0; i-- {
		t := sorted[i]
		dto := domain.TradeDTO{
			ID:                  t.ID,
			DepotID:             t.DepotID,
			WalletTransactionID: t.WalletTransactionID,
			StockID:             t.StockID,
			Type:                t.Type,
			Quantity:            t.Quantity,
			TotalInCents:        t.TotalInCents,
			FeesInCents:         t.FeesInCents,
			TaxesInCents:        t.TaxesInCents,
			Timestamp:           t.Timestamp,
		}
		if t.Type == domain.TradeTypeSell {
			dto.CostBasisInCents = costByTrade[t.ID]
			dto.ProceedsInCents = proceedsByTrade[t.ID]
			dto.RealizedGainInCents = dto.ProceedsInCents - dto.CostBasisInCents
		}
		dto.CanDelete = canDeleteTrade(sorted, t)
		dtos = append(dtos, dto)
	}
	return dtos
}
