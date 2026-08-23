package services

import (
	"strings"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type budgetGroupService struct {
	budgetGroupRepo ports.BudgetGroupRepository
}

func NewBudgetGroupService(budgetGroupRepo ports.BudgetGroupRepository) ports.BudgetGroupService {
	return &budgetGroupService{budgetGroupRepo: budgetGroupRepo}
}

func (s *budgetGroupService) CreateBudgetGroup(userID int, g domain.BudgetGroup) error {
	g.UserID = userID

	if strings.TrimSpace(g.Name) == "" {
		return domain.ErrMissingBudgetGroup
	}

	return s.budgetGroupRepo.SaveBudgetGroup(g)
}

func (s *budgetGroupService) GetBudgetGroups(userID int) ([]domain.BudgetGroup, error) {
	return s.budgetGroupRepo.FindBudgetGroupsByUser(userID)
}

func (s *budgetGroupService) UpdateBudgetGroup(userID int, g domain.BudgetGroup) error {
	existing, err := s.findOwned(userID, g.ID)
	if err != nil {
		return err
	}
	_ = existing

	if strings.TrimSpace(g.Name) == "" {
		return domain.ErrMissingBudgetGroup
	}

	g.UserID = userID
	return s.budgetGroupRepo.UpdateBudgetGroup(g)
}

func (s *budgetGroupService) DeleteBudgetGroup(userID int, id int) error {
	if _, err := s.findOwned(userID, id); err != nil {
		return err
	}
	return s.budgetGroupRepo.DeleteBudgetGroup(id)
}

func (s *budgetGroupService) findOwned(userID int, id int) (domain.BudgetGroup, error) {
	groups, err := s.budgetGroupRepo.FindBudgetGroupsByUser(userID)
	if err != nil {
		return domain.BudgetGroup{}, err
	}
	for _, g := range groups {
		if g.ID == id {
			return g, nil
		}
	}
	return domain.BudgetGroup{}, domain.ErrBudgetGroupNotFound
}
