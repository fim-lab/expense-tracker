package services

import (
	"fmt"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type transactionTemplateService struct {
	transactionTemplateRepo ports.TransactionTemplateRepository
	walletRepo              ports.WalletRepository
	budgetRepo              ports.BudgetRepository
}

func NewTransactionTemplateService(
	transactionTemplateRepo ports.TransactionTemplateRepository,
	walletRepo ports.WalletRepository,
	budgetRepo ports.BudgetRepository,
) ports.TransactionTemplateService {
	return &transactionTemplateService{
		transactionTemplateRepo: transactionTemplateRepo,
		walletRepo:              walletRepo,
		budgetRepo:              budgetRepo,
	}
}

func (s *transactionTemplateService) CreateTransactionTemplate(userID int, tt domain.TransactionTemplate) error {
	if tt.UserID != userID {
		return domain.ErrUnauthorized
	}
	if err := tt.Validate(); err != nil {
		return err
	}

	_, err := s.walletRepo.GetWalletByID(tt.WalletID)
	if err != nil {
		return fmt.Errorf("wallet validation failed: %w", err)
	}

	if tt.BudgetID != nil {
		_, err := s.budgetRepo.GetBudgetByID(*tt.BudgetID)
		if err != nil {
			return fmt.Errorf("budget validation failed: %w", err)
		}
	}

	return s.transactionTemplateRepo.SaveTransactionTemplate(tt)
}

func (s *transactionTemplateService) GetTransactionTemplate(userID int, id int) (domain.TransactionTemplate, error) {
	tt, err := s.transactionTemplateRepo.GetTransactionTemplateByID(id)
	if err != nil {
		return domain.TransactionTemplate{}, err
	}
	if tt.UserID != userID {
		return domain.TransactionTemplate{}, domain.ErrUnauthorized
	}
	return tt, nil
}

func (s *transactionTemplateService) GetTransactionTemplates(userID int) ([]domain.TransactionTemplate, error) {
	return s.transactionTemplateRepo.FindTransactionTemplatesByUser(userID)
}

func (s *transactionTemplateService) UpdateTransactionTemplate(userID int, tt domain.TransactionTemplate) error {
	existingTT, err := s.transactionTemplateRepo.GetTransactionTemplateByID(tt.ID)
	if err != nil {
		return err
	}
	if existingTT.UserID != userID {
		return domain.ErrUnauthorized
	}
	if tt.UserID != userID {
		return domain.ErrUnauthorized
	}

	if err := tt.Validate(); err != nil {
		return err
	}

	_, err = s.walletRepo.GetWalletByID(tt.WalletID)
	if err != nil {
		return fmt.Errorf("wallet validation failed: %w", err)
	}

	if tt.BudgetID != nil {
		_, err := s.budgetRepo.GetBudgetByID(*tt.BudgetID)
		if err != nil {
			return fmt.Errorf("budget validation failed: %w", err)
		}
	}

	return s.transactionTemplateRepo.UpdateTransactionTemplate(tt)
}

func (s *transactionTemplateService) DeleteTransactionTemplate(userID int, id int) error {
	existingTT, err := s.transactionTemplateRepo.GetTransactionTemplateByID(id)
	if err != nil {
		return err
	}
	if existingTT.UserID != userID {
		return domain.ErrUnauthorized
	}
	return s.transactionTemplateRepo.DeleteTransactionTemplate(id)
}
