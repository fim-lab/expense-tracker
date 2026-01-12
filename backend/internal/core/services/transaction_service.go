package services

import (
	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type transactionService struct {
	repo ports.ExpenseRepository
}

func NewTransactionService(repo ports.ExpenseRepository) ports.TransactionService {
	return &transactionService{repo: repo}
}

func (s *transactionService) CreateTransaction(userID int, t domain.Transaction) error {
	t.UserID = userID

	budget, err := s.repo.GetBudgetByID(t.BudgetID)
	if err != nil || budget.UserID != userID {
		return domain.ErrBudgetNotFound
	}

	if t.AmountInCents <= 0 {
		return domain.ErrInvalidAmount
	}

	return s.repo.SaveTransaction(t)
}

func (s *transactionService) GetTransactions(userID int, limit int, offset int) ([]domain.TransactionDTO, error) {
	txs, err := s.repo.FindTransactionsByUser(userID, limit, offset)
	if err != nil {
		return nil, err
	}

	budgets, _ := s.repo.FindBudgetsByUser(userID)
	wallets, _ := s.repo.FindWalletsByUser(userID)

	bMap := make(map[int]string)
	for _, b := range budgets {
		bMap[b.ID] = b.Name
	}

	wMap := make(map[int]string)
	for _, w := range wallets {
		wMap[w.ID] = w.Name
	}

	dtos := make([]domain.TransactionDTO, 0, len(txs))
	for _, t := range txs {
		dtos = append(dtos, domain.TransactionDTO{
			ID:            t.ID,
			Date:          t.Date,
			Description:   t.Description,
			AmountInCents: t.AmountInCents,
			Type:          t.Type,
			BudgetName:    bMap[t.BudgetID],
			WalletName:    wMap[t.WalletID],
			IsPending:     t.IsPending,
		})
	}

	return dtos, nil
}

func (s *transactionService) GetTransactionCount(userID int) (int, error) {
	return s.repo.GetTransactionCount(userID)
}

// TODO: UpdateTransactions

func (s *transactionService) DeleteTransaction(userID int, id int) error {
	existing, err := s.repo.GetTransactionByID(id)
	if err != nil || existing.UserID != userID {
		return domain.ErrUnauthorized
	}
	return s.repo.DeleteTransaction(id)
}
