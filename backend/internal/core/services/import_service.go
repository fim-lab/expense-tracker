package services

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

const importTransferBudget = "Übertrag"

var tradeDescriptionPattern = regexp.MustCompile(`(?i)^(buy|sell)\s+(\d+(?:\.\d+)?)\s*stk\.\s*(\S+)\s*$`)

// parseTradeDescription recognizes descriptions like "Buy 1.86511Stk. A1JX52"
func parseTradeDescription(description string) (tradeType domain.TradeType, quantity float64, wkn string, ok bool) {
	normalized := strings.Map(func(r rune) rune {
		if unicode.IsSpace(r) {
			return ' '
		}
		return r
	}, description)

	match := tradeDescriptionPattern.FindStringSubmatch(normalized)
	if match == nil {
		return "", 0, "", false
	}

	quantity, err := strconv.ParseFloat(match[2], 64)
	if err != nil {
		return "", 0, "", false
	}

	tradeType = domain.TradeTypeBuy
	if strings.EqualFold(match[1], "sell") {
		tradeType = domain.TradeTypeSell
	}

	return tradeType, quantity, strings.ToUpper(match[3]), true
}

type importService struct {
	userRepo                ports.UserRepository
	budgetRepo              ports.BudgetRepository
	budgetGroupRepo         ports.BudgetGroupRepository
	walletRepo              ports.WalletRepository
	depotRepo               ports.DepotRepository
	transactionRepo         ports.TransactionRepository
	tradeRepo               ports.TradeRepository
	transactionTemplateRepo ports.TransactionTemplateRepository
	stockService            ports.StockService
}

func NewImportService(
	userRepo ports.UserRepository,
	budgetRepo ports.BudgetRepository,
	budgetGroupRepo ports.BudgetGroupRepository,
	walletRepo ports.WalletRepository,
	depotRepo ports.DepotRepository,
	transactionRepo ports.TransactionRepository,
	tradeRepo ports.TradeRepository,
	transactionTemplateRepo ports.TransactionTemplateRepository,
	stockService ports.StockService,
) ports.ImportService {
	return &importService{
		userRepo:                userRepo,
		budgetRepo:              budgetRepo,
		budgetGroupRepo:         budgetGroupRepo,
		walletRepo:              walletRepo,
		depotRepo:               depotRepo,
		transactionRepo:         transactionRepo,
		tradeRepo:               tradeRepo,
		transactionTemplateRepo: transactionTemplateRepo,
		stockService:            stockService,
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

	if err := s.budgetGroupRepo.DeleteAllByUser(userID); err != nil {
		return fmt.Errorf("failed to delete budget groups: %w", err)
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

	existingBudgets, err := s.budgetRepo.FindBudgetsByUser(userID)
	if err != nil {
		return fmt.Errorf("failed to fetch existing budgets: %w", err)
	}
	budgetMap := make(map[string]int)
	for _, b := range existingBudgets {
		budgetMap[b.Name] = b.ID
	}

	existingGroups, err := s.budgetGroupRepo.FindBudgetGroupsByUser(userID)
	if err != nil {
		return fmt.Errorf("failed to fetch existing budget groups: %w", err)
	}
	groupMap := make(map[string]int)
	for _, g := range existingGroups {
		groupMap[g.Name] = g.ID
	}

	for _, importBudget := range data.Settings.Budgets {
		account := strings.TrimSpace(importBudget.Account)
		if account == "" || strings.EqualFold(account, "private") {
			continue
		}
		if _, exists := groupMap[account]; !exists {
			g := domain.BudgetGroup{
				UserID: userID,
				Name:   account,
			}
			if err := s.budgetGroupRepo.SaveBudgetGroup(g); err != nil {
				return fmt.Errorf("failed to save budget group %s: %w", account, err)
			}
			groupMap[account] = 0 // placeholder until refetch, guards against re-creating in this loop
		}
	}

	// Refetch groups to get ids
	existingGroups, err = s.budgetGroupRepo.FindBudgetGroupsByUser(userID)
	if err != nil {
		return fmt.Errorf("failed to fetch budget groups: %w", err)
	}
	groupMap = make(map[string]int)
	for _, g := range existingGroups {
		groupMap[g.Name] = g.ID
	}

	for _, importBudget := range data.Settings.Budgets {
		if _, exists := budgetMap[importBudget.Name]; !exists {
			b := domain.Budget{
				UserID:     userID,
				Name:       importBudget.Name,
				LimitCents: importBudget.ValueInCents,
			}
			account := strings.TrimSpace(importBudget.Account)
			if account != "" && !strings.EqualFold(account, "private") {
				if groupID, ok := groupMap[account]; ok {
					b.GroupID = &groupID
				}
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

	var firstBudgetID int
	if len(existingBudgets) > 0 {
		firstBudgetID = existingBudgets[0].ID
	}

	if firstWalletID != 0 && firstBudgetID != 0 {
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
					BudgetID: firstBudgetID,
				}
				if err := s.depotRepo.SaveDepot(d); err != nil {
					return fmt.Errorf("failed to save depot %s: %w", importWallet.Name, err)
				}
			}
		}
	}

	existingDepots, err := s.depotRepo.FindDepotsByUser(userID)
	if err != nil {
		return fmt.Errorf("failed to fetch existing depots: %w", err)
	}
	depotByWallet := make(map[int]int)
	for _, d := range existingDepots {
		depotByWallet[d.WalletID] = d.ID
	}

	for i := len(data.Transactions) - 1; i >= 0; i-- {
		importTx := data.Transactions[i]
		walletID, ok := walletMap[importTx.Wallet]
		if !ok {
			continue
		}

		txType := domain.TransactionType(strings.ToUpper(importTx.Type))
		amount := importTx.AmountInCents
		if amount < 0 {
			amount = -amount
		}

		isDebt := importTx.IsDebt
		isPending := importTx.IsPending
		t := domain.Transaction{
			UserID:        userID,
			Date:          importTx.Date,
			Description:   importTx.Description,
			AmountInCents: amount,
			Type:          txType,
			IsPending:     &isPending,
			IsDebt:        &isDebt,
			WalletID:      walletID,
		}

		if importTx.Budget != "" && importTx.Budget != importTransferBudget {
			if budgetId, ok := budgetMap[importTx.Budget]; ok {
				t.BudgetID = &budgetId
			}
		}

		transactionID, err := s.transactionRepo.SaveTransaction(t)
		if err != nil {
			return fmt.Errorf("failed to save transaction: %w", err)
		}

		tradeType, quantity, wkn, isTrade := parseTradeDescription(importTx.Description)
		if !isTrade {
			continue
		}

		depotID, ok := depotByWallet[walletID]
		if !ok {
			return fmt.Errorf("transaction %q looks like a trade but wallet %q has no depot", importTx.Description, importTx.Wallet)
		}

		fallback := int(math.Round(float64(amount) / quantity))
		stock, err := s.stockService.GetOrCreateByWKN(wkn, fallback)
		if err != nil {
			return fmt.Errorf("failed to resolve stock %q: %w", wkn, err)
		}

		trade := domain.Trade{
			DepotID:             depotID,
			WalletTransactionID: &transactionID,
			WKN:                 wkn,
			StockID:             stock.ID,
			Type:                tradeType,
			Quantity:            quantity,
			TotalInCents:        amount,
			Timestamp:           normalizeTradeTimestamp(t.Date),
		}
		if _, err := s.tradeRepo.SaveTrade(trade); err != nil {
			return fmt.Errorf("failed to save trade for transaction %q: %w", importTx.Description, err)
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
