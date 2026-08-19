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
	unmatched   map[string]float64
}

func buildPortfolio(trades []domain.Trade) PortfolioSnapshot {
	snapshot := PortfolioSnapshot{unmatched: map[string]float64{}}

	lotsByWKN := map[string][]*domain.Lot{}
	var lots []*domain.Lot

	for _, t := range sortTradesChronologically(trades) {
		switch t.Type {
		case domain.TradeTypeBuy:
			lot := &domain.Lot{
				TradeID:              t.ID,
				DepotID:              t.DepotID,
				WKN:                  t.WKN,
				DateOfPurchase:       t.Timestamp,
				Quantity:             t.Quantity,
				Remaining:            t.Quantity,
				TotalInCents:         t.TotalInCents,
				RemainingCostInCents: t.TotalInCents,
			}
			lots = append(lots, lot)
			lotsByWKN[t.WKN] = append(lotsByWKN[t.WKN], lot)
		case domain.TradeTypeSell:
			snapshot.applySell(t, lotsByWKN[t.WKN])
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
			WKN:                 sell.WKN,
			Quantity:            take,
			CostBasisInCents:    costBasis,
			ProceedsInCents:     proceeds,
			RealizedGainInCents: proceeds - costBasis,
			BuyDate:             lot.DateOfPurchase,
			SellDate:            sell.Timestamp,
		})
	}

	if toSell > 0 {
		s.unmatched[sell.WKN] += toSell
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
	byWKN := map[string]*domain.Position{}
	var order []string

	for _, lot := range s.openLots {
		position, ok := byWKN[lot.WKN]
		if !ok {
			position = &domain.Position{DepotID: depotID, WKN: lot.WKN, Lots: []domain.Lot{}}
			byWKN[lot.WKN] = position
			order = append(order, lot.WKN)
		}
		position.Quantity += lot.Remaining
		position.InvestedInCents += lot.RemainingCostInCents
		position.Lots = append(position.Lots, lot)
	}

	sort.Strings(order)
	positions := make([]domain.Position, 0, len(order))
	for _, wkn := range order {
		position := byWKN[wkn]
		position.Quantity = clampQuantity(position.Quantity)
		if position.Quantity <= 0 {
			continue
		}
		position.AvgPriceInCents = int(math.Round(float64(position.InvestedInCents) / position.Quantity))
		positions = append(positions, *position)
	}
	return positions
}

func (s PortfolioSnapshot) realizedGain() int {
	var total int
	for _, allocation := range s.allocations {
		total += allocation.RealizedGainInCents
	}
	return total
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
			WKN:                 t.WKN,
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
		dtos = append(dtos, dto)
	}
	return dtos
}
