package services

import (
	"testing"

	"github.com/fim-lab/expense-tracker/adapters/repository/memory"
	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type stockFixture struct {
	repos        ports.Repositories
	depotSvc     ports.DepotService
	txSvc        ports.TransactionService
	tradeSvc     ports.TradeService
	portfolioSvc ports.PortfolioService
	userID       int
	walletID     int
	depotID      int
}

func newStockFixture(t *testing.T) stockFixture {
	t.Helper()

	repos := memory.NewCleanRepositories()
	userID, walletID, depotID := 1, 1, 1

	if err := repos.WalletRepository().SaveWallet(domain.Wallet{ID: walletID, UserID: userID, Name: "Main Wallet"}); err != nil {
		t.Fatalf("could not seed the wallet: %v", err)
	}
	if err := repos.DepotRepository().SaveDepot(domain.Depot{ID: depotID, UserID: userID, Name: "Main Depot", WalletID: walletID}); err != nil {
		t.Fatalf("could not seed the depot: %v", err)
	}

	depotSvc := NewDepotService(repos.DepotRepository(), repos.WalletRepository(), repos.TradeRepository())
	txSvc := NewTransactionService(repos.TransactionRepository(), repos.BudgetRepository(), repos.WalletRepository())

	return stockFixture{
		repos:        repos,
		depotSvc:     depotSvc,
		txSvc:        txSvc,
		tradeSvc:     NewTradeService(repos.TradeRepository(), depotSvc, txSvc),
		portfolioSvc: NewPortfolioService(repos.TradeRepository(), depotSvc),
		userID:       userID,
		walletID:     walletID,
		depotID:      depotID,
	}
}

func (f stockFixture) trade(tradeType domain.TradeType, day int, quantity float64, totalInCents int) domain.Trade {
	return domain.Trade{
		DepotID:      f.depotID,
		WKN:          testWKN,
		Type:         tradeType,
		Quantity:     quantity,
		TotalInCents: totalInCents,
		Timestamp:    tradeDay(day),
	}
}

func (f stockFixture) mustBuy(t *testing.T, day int, quantity float64, totalInCents int) int {
	t.Helper()
	created, err := f.tradeSvc.CreateTrade(f.userID, f.trade(domain.TradeTypeBuy, day, quantity, totalInCents))
	if err != nil {
		t.Fatalf("buying %v shares failed: %v", quantity, err)
	}
	return created.ID
}

func (f stockFixture) mustSell(t *testing.T, day int, quantity float64, totalInCents int) int {
	t.Helper()
	created, err := f.tradeSvc.CreateTrade(f.userID, f.trade(domain.TradeTypeSell, day, quantity, totalInCents))
	if err != nil {
		t.Fatalf("selling %v shares failed: %v", quantity, err)
	}
	return created.ID
}

func (f stockFixture) walletBalance(t *testing.T) int {
	t.Helper()
	wallet, err := f.repos.WalletRepository().GetWalletByID(f.walletID)
	if err != nil {
		t.Fatalf("could not read the wallet: %v", err)
	}
	return wallet.BalanceCents
}

func (f stockFixture) transactionCount(t *testing.T) int {
	t.Helper()
	count, err := f.txSvc.GetTransactionCount(f.userID)
	if err != nil {
		t.Fatalf("could not count transactions: %v", err)
	}
	return count
}

func (f stockFixture) tradeCount(t *testing.T) int {
	t.Helper()
	trades, err := f.repos.TradeRepository().FindTradesByDepot(f.depotID)
	if err != nil {
		t.Fatalf("could not read the trades: %v", err)
	}
	return len(trades)
}

func (f stockFixture) mustGetTrade(t *testing.T, id int) domain.Trade {
	t.Helper()
	trade, err := f.repos.TradeRepository().GetTradeByID(id)
	if err != nil {
		t.Fatalf("could not read trade %d: %v", id, err)
	}
	return trade
}

func (f stockFixture) mustGetPortfolio(t *testing.T) domain.Portfolio {
	t.Helper()
	portfolio, err := f.portfolioSvc.GetPortfolio(f.userID, f.depotID)
	if err != nil {
		t.Fatalf("could not read the portfolio: %v", err)
	}
	return portfolio
}

func (f stockFixture) linkedTransaction(t *testing.T, tradeID int) domain.Transaction {
	t.Helper()
	trade := f.mustGetTrade(t, tradeID)
	if trade.WalletTransactionID == nil {
		t.Fatalf("trade %d has no wallet transaction", tradeID)
	}
	transaction, err := f.txSvc.GetTransactionByID(f.userID, *trade.WalletTransactionID)
	if err != nil {
		t.Fatalf("could not read the wallet transaction of trade %d: %v", tradeID, err)
	}
	return transaction
}
