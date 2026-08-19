package services

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type importService struct {
	userRepo                ports.UserRepository
	budgetRepo              ports.BudgetRepository
	walletRepo              ports.WalletRepository
	depotRepo               ports.DepotRepository
	transactionRepo         ports.TransactionRepository
	tradeRepo               ports.TradeRepository
	transactionTemplateRepo ports.TransactionTemplateRepository
}

func NewImportService(
	userRepo ports.UserRepository,
	budgetRepo ports.BudgetRepository,
	walletRepo ports.WalletRepository,
	depotRepo ports.DepotRepository,
	transactionRepo ports.TransactionRepository,
	tradeRepo ports.TradeRepository,
	transactionTemplateRepo ports.TransactionTemplateRepository,
) ports.ImportService {
	return &importService{
		userRepo:                userRepo,
		budgetRepo:              budgetRepo,
		walletRepo:              walletRepo,
		depotRepo:               depotRepo,
		transactionRepo:         transactionRepo,
		tradeRepo:               tradeRepo,
		transactionTemplateRepo: transactionTemplateRepo,
	}
}

func (s *importService) DeleteAllUserData(userID int) error {
	if err := s.transactionRepo.DeleteAllByUser(userID); err != nil {
		return fmt.Errorf("failed to delete transactions: %w", err)
	}

	if err := s.transactionTemplateRepo.DeleteAllByUser(userID); err != nil {
		return fmt.Errorf("failed to delete templates: %w", err)
	}

	if err := s.tradeRepo.DeleteAllByUser(userID); err != nil {
		return fmt.Errorf("failed to delete trades: %w", err)
	}

	if err := s.depotRepo.DeleteAllByUser(userID); err != nil {
		return fmt.Errorf("failed to delete depots: %w", err)
	}

	if err := s.budgetRepo.DeleteAllByUser(userID); err != nil {
		return fmt.Errorf("failed to delete budgets: %w", err)
	}

	if err := s.walletRepo.DeleteAllByUser(userID); err != nil {
		return fmt.Errorf("failed to delete wallets: %w", err)
	}

	if err := s.userRepo.UpdateUserSalary(userID, 0); err != nil {
		return fmt.Errorf("failed to reset salary: %w", err)
	}

	return nil
}

func (s *importService) ImportData(userID int, data domain.FullImportData) error {
	if data.Settings.Gehalt > 0 {
		if err := s.userRepo.UpdateUserSalary(userID, data.Settings.Gehalt); err != nil {
			return fmt.Errorf("failed to update salary: %w", err)
		}
	}

	existingWallets, err := s.walletRepo.FindWalletsByUser(userID)
	if err != nil {
		return fmt.Errorf("failed to fetch existing wallets: %w", err)
	}
	walletMap := make(map[string]int)
	for _, w := range existingWallets {
		walletMap[w.Name] = w.ID
	}

	for _, importWallet := range data.Settings.Wallets {
		if importWallet.IsDepot {
			continue
		}
		if _, exists := walletMap[importWallet.Name]; !exists {
			w := domain.Wallet{
				UserID: userID,
				Name:   importWallet.Name,
			}
			if err := s.walletRepo.SaveWallet(w); err != nil {
				return fmt.Errorf("failed to save wallet %s: %w", importWallet.Name, err)
			}
		}
	}

	// Refetch to get newly created Ids
	existingWallets, err = s.walletRepo.FindWalletsByUser(userID)
	if err != nil {
		return err
	}
	for _, w := range existingWallets {
		walletMap[w.Name] = w.ID
	}

	var firstWalletID int
	if len(existingWallets) > 0 {
		firstWalletID = existingWallets[0].ID
	}

	if firstWalletID != 0 {
		existingDepots, err := s.depotRepo.FindDepotsByUser(userID)
		if err != nil {
			return fmt.Errorf("failed to fetch existing depots: %w", err)
		}
		depotMap := make(map[string]bool)
		for _, d := range existingDepots {
			depotMap[d.Name] = true
		}

		for _, importWallet := range data.Settings.Wallets {
			if !importWallet.IsDepot {
				continue
			}
			if !depotMap[importWallet.Name] {
				d := domain.Depot{
					UserID:   userID,
					Name:     importWallet.Name,
					WalletID: firstWalletID,
				}
				if err := s.depotRepo.SaveDepot(d); err != nil {
					return fmt.Errorf("failed to save depot %s: %w", importWallet.Name, err)
				}
			}
		}
	}

	existingBudgets, err := s.budgetRepo.FindBudgetsByUser(userID)
	if err != nil {
		return fmt.Errorf("failed to fetch existing budgets: %w", err)
	}
	budgetMap := make(map[string]int)
	for _, b := range existingBudgets {
		budgetMap[b.Name] = b.ID
	}

	for _, importBudget := range data.Settings.Budgets {
		if strings.ToLower(importBudget.Account) != "private" {
			continue
		}
		if _, exists := budgetMap[importBudget.Name]; !exists {
			b := domain.Budget{
				UserID:     userID,
				Name:       importBudget.Name,
				LimitCents: importBudget.ValueInCents,
			}
			if err := s.budgetRepo.SaveBudget(b); err != nil {
				return fmt.Errorf("failed to save budget %s: %w", importBudget.Name, err)
			}
		}
	}

	// Refetch budgets to get ids
	existingBudgets, err = s.budgetRepo.FindBudgetsByUser(userID)
	if err != nil {
		return err
	}
	for _, b := range existingBudgets {
		budgetMap[b.Name] = b.ID
	}

	for _, importTx := range data.Transactions {
		walletID, ok := walletMap[importTx.Wallet]
		if !ok {
			continue
		}

		txType := domain.TransactionType(strings.ToUpper(importTx.Type))
		amount := importTx.AmountInCents
		if amount < 0 {
			amount = -amount
		}

		t := domain.Transaction{
			UserID:        userID,
			Date:          importTx.Date,
			Description:   importTx.Description,
			AmountInCents: amount,
			Type:          txType,
			IsPending:     importTx.IsPending,
			WalletID:      walletID,
		}

		if importTx.Budget != "" {
			if budgetId, ok := budgetMap[importTx.Budget]; ok {
				t.BudgetID = &budgetId
			}
		}

		if _, err := s.transactionRepo.SaveTransaction(t); err != nil {
			return fmt.Errorf("failed to save transaction: %w", err)
		}
	}

	return nil
}

func (s *importService) ImportTestData(userID int) error {
	file, err := os.Open("data/testdata.json")
	if err != nil {
		return fmt.Errorf("could not open testdata file: %w", err)
	}
	defer file.Close()

	var data domain.FullImportData
	if err := json.NewDecoder(file).Decode(&data); err != nil {
		return fmt.Errorf("could not decode testdata: %w", err)
	}

	return s.ImportData(userID, data)
}
