package services

import (
	"testing"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
)

func TestTradeService_CreateBuyCreatesTradeAndCashTransaction(t *testing.T) {
	f := newStockFixture(t)

	tradeID := f.mustBuy(t, 1, 10, 10000)

	if tradeID == 0 {
		t.Error("expected CreateTrade to return the saved trade with its new id")
	}
	if balance := f.walletBalance(t); balance != -10000 {
		t.Errorf("expected the wallet to be debited by 10000, got %d", balance)
	}

	transaction := f.linkedTransaction(t, tradeID)
	if transaction.Type != domain.Expense || transaction.AmountInCents != 10000 {
		t.Errorf("expected an expense of 10000, got %s of %d", transaction.Type, transaction.AmountInCents)
	}
	if transaction.WalletID != f.walletID {
		t.Errorf("expected the transaction on wallet %d, got %d", f.walletID, transaction.WalletID)
	}

	portfolio := f.mustGetPortfolio(t)
	if len(portfolio.Positions) != 1 {
		t.Fatalf("expected 1 position, got %d", len(portfolio.Positions))
	}
	position := portfolio.Positions[0]
	if position.Quantity != 10 || position.InvestedInCents != 10000 || position.AvgPriceInCents != 1000 {
		t.Errorf("expected 10 shares invested with 10000 at 1000 each, got %+v", position)
	}
	if len(position.Lots) != 1 || position.Lots[0].TradeID != tradeID || position.Lots[0].Remaining != 10 {
		t.Errorf("expected one open lot of the buy trade with 10 shares left, got %+v", position.Lots)
	}
}

func TestTradeService_OversellRejectedWithNoSideEffects(t *testing.T) {
	f := newStockFixture(t)
	f.mustBuy(t, 1, 10, 10000)

	balanceBefore := f.walletBalance(t)
	transactionsBefore := f.transactionCount(t)

	_, err := f.tradeSvc.CreateTrade(f.userID, f.trade(domain.TradeTypeSell, 2, 20, 24000))
	if err != domain.ErrInsufficientShares {
		t.Fatalf("expected ErrInsufficientShares, got %v", err)
	}

	if balance := f.walletBalance(t); balance != balanceBefore {
		t.Errorf("expected the wallet balance to stay at %d, got %d", balanceBefore, balance)
	}
	if count := f.transactionCount(t); count != transactionsBefore {
		t.Errorf("expected no new wallet transaction, got %d instead of %d", count, transactionsBefore)
	}
	if count := f.tradeCount(t); count != 1 {
		t.Errorf("expected the rejected sell not to be stored, got %d trades", count)
	}
}

func TestTradeService_SellDrainsLotsFIFOAndCreatesIncome(t *testing.T) {
	f := newStockFixture(t)
	firstBuy := f.mustBuy(t, 1, 10, 10000)
	secondBuy := f.mustBuy(t, 2, 5, 6000)

	sellID := f.mustSell(t, 3, 12, 15000)

	if balance := f.walletBalance(t); balance != -1000 {
		t.Errorf("expected -10000 - 6000 + 15000 = -1000, got %d", balance)
	}

	transaction := f.linkedTransaction(t, sellID)
	if transaction.Type != domain.Income || transaction.AmountInCents != 15000 {
		t.Errorf("expected an income of 15000, got %s of %d", transaction.Type, transaction.AmountInCents)
	}

	portfolio := f.mustGetPortfolio(t)
	if len(portfolio.Positions) != 1 {
		t.Fatalf("expected 1 remaining position, got %d", len(portfolio.Positions))
	}
	position := portfolio.Positions[0]
	if position.Quantity != 3 || position.InvestedInCents != 3600 {
		t.Errorf("expected 3 shares left at a cost of 3600, got %+v", position)
	}
	if len(position.Lots) != 1 || position.Lots[0].TradeID != secondBuy {
		t.Errorf("expected only the younger lot to be left, got %+v", position.Lots)
	}
	if portfolio.RealizedGainInCents != 2600 {
		t.Errorf("expected a realized gain of 2600, got %d", portfolio.RealizedGainInCents)
	}
	if f.mustGetTrade(t, firstBuy).ID != firstBuy {
		t.Error("expected the drained buy trade to still exist")
	}
}

func TestTradeService_DeleteTradeRemovesCashTransactionAndRestoresBalance(t *testing.T) {
	f := newStockFixture(t)
	balanceBefore := f.walletBalance(t)

	tradeID := f.mustBuy(t, 1, 10, 10000)

	if err := f.tradeSvc.DeleteTrade(f.userID, tradeID); err != nil {
		t.Fatalf("deleting the trade failed: %v", err)
	}

	if balance := f.walletBalance(t); balance != balanceBefore {
		t.Errorf("expected the wallet balance back at %d, got %d", balanceBefore, balance)
	}
	if count := f.transactionCount(t); count != 0 {
		t.Errorf("expected the wallet transaction to be gone, got %d", count)
	}
	if count := f.tradeCount(t); count != 0 {
		t.Errorf("expected no trades left, got %d", count)
	}
}

func TestTradeService_DeleteBuyWhoseSharesWereSoldIsRejected(t *testing.T) {
	f := newStockFixture(t)
	buyID := f.mustBuy(t, 1, 10, 10000)
	f.mustSell(t, 2, 10, 12000)

	balanceBefore := f.walletBalance(t)

	if err := f.tradeSvc.DeleteTrade(f.userID, buyID); err != domain.ErrInsufficientShares {
		t.Fatalf("expected ErrInsufficientShares, got %v", err)
	}

	if count := f.tradeCount(t); count != 2 {
		t.Errorf("expected both trades to survive, got %d", count)
	}
	if count := f.transactionCount(t); count != 2 {
		t.Errorf("expected both wallet transactions to survive, got %d", count)
	}
	if balance := f.walletBalance(t); balance != balanceBefore {
		t.Errorf("expected the wallet balance to stay at %d, got %d", balanceBefore, balance)
	}
}

func TestTradeService_UpdateRevalidatesAndSyncsCashTransaction(t *testing.T) {
	f := newStockFixture(t)
	buyID := f.mustBuy(t, 1, 10, 10000)
	f.mustSell(t, 2, 10, 12000)

	t.Run("an edit that would uncover the sell is rejected", func(t *testing.T) {
		shrunk := f.trade(domain.TradeTypeBuy, 1, 5, 5000)
		shrunk.ID = buyID

		if err := f.tradeSvc.UpdateTrade(f.userID, shrunk); err != domain.ErrInsufficientShares {
			t.Fatalf("expected ErrInsufficientShares, got %v", err)
		}
		if transaction := f.linkedTransaction(t, buyID); transaction.AmountInCents != 10000 {
			t.Errorf("expected the wallet transaction to stay at 10000, got %d", transaction.AmountInCents)
		}
		if trade := f.mustGetTrade(t, buyID); trade.Quantity != 10 {
			t.Errorf("expected the trade to stay at 10 shares, got %v", trade.Quantity)
		}
	})

	t.Run("a corrected total moves the wallet by the difference", func(t *testing.T) {
		balanceBefore := f.walletBalance(t)

		corrected := f.trade(domain.TradeTypeBuy, 1, 10, 11000)
		corrected.ID = buyID
		if err := f.tradeSvc.UpdateTrade(f.userID, corrected); err != nil {
			t.Fatalf("updating the trade failed: %v", err)
		}

		if transaction := f.linkedTransaction(t, buyID); transaction.AmountInCents != 11000 {
			t.Errorf("expected the wallet transaction at 11000, got %d", transaction.AmountInCents)
		}
		if balance := f.walletBalance(t); balance != balanceBefore-1000 {
			t.Errorf("expected the balance to drop by 1000 to %d, got %d", balanceBefore-1000, balance)
		}
	})
}

// The linked transaction is editable from the normal transactions view, so
// syncing it must not throw away what the user set there.
func TestTradeService_UpdatePreservesBudgetAndTagsOnLinkedTransaction(t *testing.T) {
	f := newStockFixture(t)
	budgetID := 5
	if err := f.repos.BudgetRepository().SaveBudget(domain.Budget{ID: budgetID, UserID: f.userID, Name: "Investing", LimitCents: 100000}); err != nil {
		t.Fatalf("could not seed the budget: %v", err)
	}

	tradeID := f.mustBuy(t, 1, 10, 10000)

	transaction := f.linkedTransaction(t, tradeID)
	transaction.BudgetID = &budgetID
	transaction.Tags = []string{"etf"}
	transaction.IsPending = true
	if err := f.txSvc.UpdateTransaction(f.userID, transaction); err != nil {
		t.Fatalf("could not annotate the wallet transaction: %v", err)
	}

	corrected := f.trade(domain.TradeTypeBuy, 1, 10, 11000)
	corrected.ID = tradeID
	if err := f.tradeSvc.UpdateTrade(f.userID, corrected); err != nil {
		t.Fatalf("updating the trade failed: %v", err)
	}

	updated := f.linkedTransaction(t, tradeID)
	if updated.BudgetID == nil || *updated.BudgetID != budgetID {
		t.Errorf("expected the budget to survive the update, got %v", updated.BudgetID)
	}
	if len(updated.Tags) != 1 || updated.Tags[0] != "etf" {
		t.Errorf("expected the tags to survive the update, got %v", updated.Tags)
	}
	if !updated.IsPending {
		t.Error("expected the pending flag to survive the update")
	}
	if updated.AmountInCents != 11000 {
		t.Errorf("expected the amount to be synced to 11000, got %d", updated.AmountInCents)
	}
}

func TestTradeService_BackdatedSellIsValidatedAgainstEarlierBuysOnly(t *testing.T) {
	f := newStockFixture(t)
	f.mustBuy(t, 10, 5, 5000)

	_, err := f.tradeSvc.CreateTrade(f.userID, f.trade(domain.TradeTypeSell, 1, 5, 6000))
	if err != domain.ErrInsufficientShares {
		t.Fatalf("expected a sell before the buy to be rejected, got %v", err)
	}

	earlierBuy := f.mustBuy(t, 1, 5, 4000)
	f.mustSell(t, 5, 5, 6000)

	trades, err := f.portfolioSvc.GetTrades(f.userID, f.depotID)
	if err != nil {
		t.Fatalf("could not read the trades: %v", err)
	}
	var sell domain.TradeDTO
	for _, trade := range trades {
		if trade.Type == domain.TradeTypeSell {
			sell = trade
		}
	}
	// The older lot cost 4000, so consuming it first is what FIFO means here.
	if sell.CostBasisInCents != 4000 || sell.RealizedGainInCents != 2000 {
		t.Errorf("expected the sell to consume the oldest lot (cost 4000, gain 2000), got %+v", sell)
	}

	portfolio := f.mustGetPortfolio(t)
	if len(portfolio.Positions) != 1 || portfolio.Positions[0].Lots[0].TradeID == earlierBuy {
		t.Errorf("expected the earlier lot to be the drained one, got %+v", portfolio.Positions)
	}
}

func TestTradeService_DepotChangeRejected(t *testing.T) {
	f := newStockFixture(t)
	otherDepotID := 2
	if err := f.repos.DepotRepository().SaveDepot(domain.Depot{ID: otherDepotID, UserID: f.userID, Name: "Second Depot", WalletID: f.walletID, BudgetID: f.budgetID}); err != nil {
		t.Fatalf("could not seed the second depot: %v", err)
	}
	tradeID := f.mustBuy(t, 1, 10, 10000)

	moved := f.trade(domain.TradeTypeBuy, 1, 10, 10000)
	moved.ID = tradeID
	moved.DepotID = otherDepotID

	if err := f.tradeSvc.UpdateTrade(f.userID, moved); err != domain.ErrTradeDepotChange {
		t.Fatalf("expected ErrTradeDepotChange, got %v", err)
	}
	if trade := f.mustGetTrade(t, tradeID); trade.DepotID != f.depotID {
		t.Errorf("expected the trade to stay in depot %d, got %d", f.depotID, trade.DepotID)
	}
}

func TestTradeService_CrossUserIsolation(t *testing.T) {
	f := newStockFixture(t)
	tradeID := f.mustBuy(t, 1, 10, 10000)
	otherUserID := 2

	if _, err := f.tradeSvc.CreateTrade(otherUserID, f.trade(domain.TradeTypeBuy, 1, 1, 1000)); err != domain.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized when trading in a foreign depot, got %v", err)
	}

	foreignEdit := f.trade(domain.TradeTypeBuy, 1, 1, 1000)
	foreignEdit.ID = tradeID
	if err := f.tradeSvc.UpdateTrade(otherUserID, foreignEdit); err != domain.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized when editing a foreign trade, got %v", err)
	}
	if err := f.tradeSvc.DeleteTrade(otherUserID, tradeID); err != domain.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized when deleting a foreign trade, got %v", err)
	}
	if _, err := f.portfolioSvc.GetPortfolio(otherUserID, f.depotID); err != domain.ErrUnauthorized {
		t.Errorf("expected ErrUnauthorized when reading a foreign portfolio, got %v", err)
	}

	if trade := f.mustGetTrade(t, tradeID); trade.Quantity != 10 {
		t.Errorf("expected the trade to be untouched, got %+v", trade)
	}
	if _, err := f.tradeSvc.CreateTrade(f.userID, f.trade(domain.TradeTypeBuy, 1, 1, 1000)); err != nil {
		t.Errorf("expected the owner to still be able to trade, got %v", err)
	}
}

func TestTradeService_InvalidInputRejected(t *testing.T) {
	cases := map[string]struct {
		mutate      func(*domain.Trade)
		expectedErr error
	}{
		"quantity zero": {func(tr *domain.Trade) { tr.Quantity = 0 }, domain.ErrInvalidQuantity},
		"quantity dust": {func(tr *domain.Trade) { tr.Quantity = 1e-12 }, domain.ErrInvalidQuantity},
		"total zero":    {func(tr *domain.Trade) { tr.TotalInCents = 0 }, domain.ErrInvalidAmount},
		"missing wkn":   {func(tr *domain.Trade) { tr.WKN = "   " }, domain.ErrMissingWKN},
		"unknown type":  {func(tr *domain.Trade) { tr.Type = "TRANSFER_IN" }, domain.ErrInvalidTradeType},
		"unknown depot": {func(tr *domain.Trade) { tr.DepotID = 999 }, domain.ErrDepotNotFound},
	}

	for name, testCase := range cases {
		t.Run(name, func(t *testing.T) {
			f := newStockFixture(t)
			trade := f.trade(domain.TradeTypeBuy, 1, 10, 10000)
			testCase.mutate(&trade)

			if _, err := f.tradeSvc.CreateTrade(f.userID, trade); err != testCase.expectedErr {
				t.Fatalf("expected %v, got %v", testCase.expectedErr, err)
			}
			if count := f.tradeCount(t); count != 0 {
				t.Errorf("expected no trade to be stored, got %d", count)
			}
			if count := f.transactionCount(t); count != 0 {
				t.Errorf("expected no wallet transaction, got %d", count)
			}
		})
	}
}

func TestTradeService_TolerateMissingLinkedTransaction(t *testing.T) {
	f := newStockFixture(t)
	tradeID := f.mustBuy(t, 1, 10, 10000)
	transaction := f.linkedTransaction(t, tradeID)

	if err := f.txSvc.DeleteTransaction(f.userID, transaction.ID); err != nil {
		t.Fatalf("could not delete the wallet transaction: %v", err)
	}

	if err := f.tradeSvc.DeleteTrade(f.userID, tradeID); err != nil {
		t.Fatalf("expected deleting the trade to succeed, got %v", err)
	}
	if count := f.tradeCount(t); count != 0 {
		t.Errorf("expected no trades left, got %d", count)
	}
}

func TestTradeService_WKNIsNormalized(t *testing.T) {
	f := newStockFixture(t)

	sloppy := f.trade(domain.TradeTypeBuy, 1, 10, 10000)
	sloppy.WKN = "  a1jx52 "
	if _, err := f.tradeSvc.CreateTrade(f.userID, sloppy); err != nil {
		t.Fatalf("buying with a sloppy WKN failed: %v", err)
	}
	f.mustBuy(t, 2, 5, 6000)

	portfolio := f.mustGetPortfolio(t)
	if len(portfolio.Positions) != 1 {
		t.Fatalf("expected both buys in one position, got %d positions", len(portfolio.Positions))
	}
	if portfolio.Positions[0].WKN != testWKN || portfolio.Positions[0].Quantity != 15 {
		t.Errorf("expected 15 shares of %s, got %+v", testWKN, portfolio.Positions[0])
	}
}
