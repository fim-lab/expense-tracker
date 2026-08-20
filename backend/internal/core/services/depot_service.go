package services

import (
	"errors"
	"sort"
	"strings"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type depotService struct {
	depotRepo    ports.DepotRepository
	walletRepo   ports.WalletRepository
	budgetRepo   ports.BudgetRepository
	tradeRepo    ports.TradeRepository
	stockService ports.StockService
}

func NewDepotService(depotRepo ports.DepotRepository, walletRepo ports.WalletRepository, budgetRepo ports.BudgetRepository, tradeRepo ports.TradeRepository, stockService ports.StockService) ports.DepotService {
	return &depotService{depotRepo: depotRepo, walletRepo: walletRepo, budgetRepo: budgetRepo, tradeRepo: tradeRepo, stockService: stockService}
}

func (s *depotService) CreateDepot(userID int, d domain.Depot) error {
	d.UserID = userID

	if strings.TrimSpace(d.Name) == "" {
		return errors.New("depot name is required")
	}

	wallet, err := s.walletRepo.GetWalletByID(d.WalletID)
	if err != nil || wallet.UserID != userID {
		return errors.New("invalid wallet for depot")
	}

	budget, err := s.budgetRepo.GetBudgetByID(d.BudgetID)
	if err != nil || budget.UserID != userID {
		return domain.ErrBudgetNotFound
	}

	return s.depotRepo.SaveDepot(d)
}

func (s *depotService) GetDepots(userID int) ([]domain.DepotDTO, error) {
	depots, err := s.depotRepo.FindDepotsByUser(userID)
	if err != nil {
		return nil, err
	}

	stocks, err := s.stockService.GetStocks()
	if err != nil {
		return nil, err
	}
	priceByStockID := make(map[int]int, len(stocks))
	for _, stock := range stocks {
		priceByStockID[stock.ID] = stock.PriceInCents
	}

	dtos := make([]domain.DepotDTO, 0, len(depots))
	for _, depot := range depots {
		trades, err := s.tradeRepo.FindTradesByDepot(depot.ID)
		if err != nil {
			return nil, err
		}
		positions := buildPortfolio(trades).positions(depot.ID)
		dtos = append(dtos, domain.DepotDTO{
			ID:                  depot.ID,
			Name:                depot.Name,
			WalletID:            depot.WalletID,
			BudgetID:            depot.BudgetID,
			InvestedInCents:     investedInCents(trades),
			CurrentValueInCents: currentValueInCents(positions, priceByStockID),
		})
	}

	sort.Slice(dtos, func(i, j int) bool { return dtos[i].Name < dtos[j].Name })

	return dtos, nil
}

func (s *depotService) GetDepotByID(userID int, id int) (domain.Depot, error) {
	depot, err := s.depotRepo.GetDepotByID(id)
	if err != nil {
		return domain.Depot{}, domain.ErrDepotNotFound
	}
	if depot.UserID != userID {
		return domain.Depot{}, domain.ErrUnauthorized
	}
	return depot, nil
}

func (s *depotService) UpdateDepot(userID int, d domain.Depot) error {
	existing, err := s.depotRepo.GetDepotByID(d.ID)
	if err != nil {
		return err
	}
	if existing.UserID != userID {
		return domain.ErrUnauthorized
	}

	if strings.TrimSpace(d.Name) == "" {
		return errors.New("depot name is required")
	}

	wallet, err := s.walletRepo.GetWalletByID(d.WalletID)
	if err != nil || wallet.UserID != userID {
		return errors.New("invalid wallet for depot")
	}

	budget, err := s.budgetRepo.GetBudgetByID(d.BudgetID)
	if err != nil || budget.UserID != userID {
		return domain.ErrBudgetNotFound
	}

	d.UserID = userID
	return s.depotRepo.UpdateDepot(d)
}

func (s *depotService) DeleteDepot(userID int, id int) error {
	if _, err := s.GetDepotByID(userID, id); err != nil {
		return err
	}

	tradeCount, err := s.tradeRepo.CountTradesByDepot(id)
	if err != nil {
		return err
	}
	if tradeCount > 0 {
		return domain.ErrNotEmpty
	}

	return s.depotRepo.DeleteDepot(id)
}
