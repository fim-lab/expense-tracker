package services

import (
	"fmt"
	"log"
	"math"
	"strings"
	"time"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type tradeService struct {
	tradeRepo          ports.TradeRepository
	depotService       ports.DepotService
	transactionService ports.TransactionService
}

func NewTradeService(
	tradeRepo ports.TradeRepository,
	depotService ports.DepotService,
	transactionService ports.TransactionService,
) ports.TradeService {
	return &tradeService{
		tradeRepo:          tradeRepo,
		depotService:       depotService,
		transactionService: transactionService,
	}
}

func (s *tradeService) CreateTrade(userID int, t domain.Trade) (domain.Trade, error) {
	depot, err := s.depotService.GetDepotByID(userID, t.DepotID)
	if err != nil {
		return domain.Trade{}, err
	}

	t, err = normalizeTrade(t)
	if err != nil {
		return domain.Trade{}, err
	}
	t.ID = 0
	t.WalletTransactionID = nil

	existing, err := s.tradeRepo.FindTradesByDepot(t.DepotID)
	if err != nil {
		return domain.Trade{}, err
	}
	if err := validateTradeHistory(append(copyTrades(existing), t)); err != nil {
		return domain.Trade{}, err
	}

	tradeID, err := s.tradeRepo.SaveTrade(t)
	if err != nil {
		return domain.Trade{}, err
	}
	t.ID = tradeID

	if err := s.syncCashTransaction(userID, depot, &t); err != nil {
		if deleteErr := s.tradeRepo.DeleteTrade(tradeID); deleteErr != nil {
			log.Printf("trade %d has no wallet transaction and could not be removed: %v", tradeID, deleteErr)
		}
		return domain.Trade{}, err
	}

	if err := s.tradeRepo.UpdateTrade(t); err != nil {
		log.Printf("trade %d could not be linked to wallet transaction %d: %v", tradeID, *t.WalletTransactionID, err)
		return domain.Trade{}, err
	}

	return t, nil
}

func (s *tradeService) UpdateTrade(userID int, t domain.Trade) error {
	existingTrade, err := s.tradeRepo.GetTradeByID(t.ID)
	if err != nil {
		return err
	}

	depot, err := s.depotService.GetDepotByID(userID, existingTrade.DepotID)
	if err != nil {
		return domain.ErrUnauthorized
	}

	if t.DepotID != 0 && t.DepotID != existingTrade.DepotID {
		return domain.ErrTradeDepotChange
	}
	t.DepotID = existingTrade.DepotID
	t.WalletTransactionID = existingTrade.WalletTransactionID

	t, err = normalizeTrade(t)
	if err != nil {
		return err
	}

	existing, err := s.tradeRepo.FindTradesByDepot(t.DepotID)
	if err != nil {
		return err
	}
	candidate := copyTrades(existing)
	for i := range candidate {
		if candidate[i].ID == t.ID {
			candidate[i] = t
		}
	}
	if err := validateTradeHistory(candidate); err != nil {
		return err
	}

	if err := s.syncCashTransaction(userID, depot, &t); err != nil {
		return err
	}

	return s.tradeRepo.UpdateTrade(t)
}

func (s *tradeService) DeleteTrade(userID int, id int) error {
	existingTrade, err := s.tradeRepo.GetTradeByID(id)
	if err != nil {
		return err
	}

	if _, err := s.depotService.GetDepotByID(userID, existingTrade.DepotID); err != nil {
		return domain.ErrUnauthorized
	}

	existing, err := s.tradeRepo.FindTradesByDepot(existingTrade.DepotID)
	if err != nil {
		return err
	}
	candidate := make([]domain.Trade, 0, len(existing))
	for _, trade := range existing {
		if trade.ID != id {
			candidate = append(candidate, trade)
		}
	}
	if err := validateTradeHistory(candidate); err != nil {
		return err
	}

	if err := s.tradeRepo.DeleteTrade(id); err != nil {
		return err
	}

	if existingTrade.WalletTransactionID != nil {
		if err := s.transactionService.DeleteTransaction(userID, *existingTrade.WalletTransactionID); err != nil {
			log.Printf("trade %d deleted, wallet transaction %d could not be removed: %v", id, *existingTrade.WalletTransactionID, err)
		}
	}

	return nil
}

func (s *tradeService) syncCashTransaction(userID int, depot domain.Depot, t *domain.Trade) error {
	transactionType := domain.Expense
	verb := "Buy"
	if t.Type == domain.TradeTypeSell {
		transactionType = domain.Income
		verb = "Sell"
	}
	description := fmt.Sprintf("%s %g %s", verb, t.Quantity, t.WKN)

	if t.WalletTransactionID != nil {
		existing, err := s.transactionService.GetTransactionByID(userID, *t.WalletTransactionID)
		if err == nil {
			existing.Date = t.Timestamp
			existing.WalletID = depot.WalletID
			existing.Description = description
			existing.AmountInCents = t.CashFlowInCents()
			existing.Type = transactionType
			return s.transactionService.UpdateTransaction(userID, existing)
		}
		log.Printf("wallet transaction %d of trade %d is missing, creating a new one", *t.WalletTransactionID, t.ID)
	}

	transactionID, err := s.transactionService.CreateTransaction(userID, domain.Transaction{
		UserID:        userID,
		Date:          t.Timestamp,
		WalletID:      depot.WalletID,
		BudgetID:      &depot.BudgetID,
		Description:   description,
		AmountInCents: t.CashFlowInCents(),
		Type:          transactionType,
	})
	if err != nil {
		return fmt.Errorf("failed to create wallet transaction: %w", err)
	}
	t.WalletTransactionID = &transactionID
	return nil
}

func normalizeTrade(t domain.Trade) (domain.Trade, error) {
	t.WKN = strings.ToUpper(strings.TrimSpace(t.WKN))
	if t.WKN == "" {
		return t, domain.ErrMissingWKN
	}
	if t.Type != domain.TradeTypeBuy && t.Type != domain.TradeTypeSell {
		return t, domain.ErrInvalidTradeType
	}
	if math.IsNaN(t.Quantity) || math.IsInf(t.Quantity, 0) || t.Quantity <= quantityEpsilon {
		return t, domain.ErrInvalidQuantity
	}
	if t.TotalInCents <= 0 {
		return t, domain.ErrInvalidAmount
	}
	if t.Timestamp.IsZero() {
		t.Timestamp = time.Now()
	}
	t.Timestamp = normalizeTradeTimestamp(t.Timestamp)
	return t, nil
}

func normalizeTradeTimestamp(ts time.Time) time.Time {
	utc := ts.UTC()
	return time.Date(utc.Year(), utc.Month(), utc.Day(), 12, 0, 0, 0, time.UTC)
}

func copyTrades(trades []domain.Trade) []domain.Trade {
	return append(make([]domain.Trade, 0, len(trades)+1), trades...)
}
