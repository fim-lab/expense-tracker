package services

import (
	"strings"

	"github.com/fim-lab/expense-tracker/internal/core/domain"
	"github.com/fim-lab/expense-tracker/internal/core/ports"
)

type templateGroupService struct {
	templateGroupRepo ports.TemplateGroupRepository
}

func NewTemplateGroupService(templateGroupRepo ports.TemplateGroupRepository) ports.TemplateGroupService {
	return &templateGroupService{templateGroupRepo: templateGroupRepo}
}

func (s *templateGroupService) CreateTemplateGroup(userID int, g domain.TemplateGroup) (domain.TemplateGroup, error) {
	g.UserID = userID

	if strings.TrimSpace(g.Name) == "" {
		return domain.TemplateGroup{}, domain.ErrMissingTemplateGroup
	}

	id, err := s.templateGroupRepo.SaveTemplateGroup(g)
	if err != nil {
		return domain.TemplateGroup{}, err
	}
	g.ID = id
	return g, nil
}

func (s *templateGroupService) GetTemplateGroups(userID int) ([]domain.TemplateGroup, error) {
	return s.templateGroupRepo.FindTemplateGroupsByUser(userID)
}

func (s *templateGroupService) UpdateTemplateGroup(userID int, g domain.TemplateGroup) error {
	existing, err := s.findOwned(userID, g.ID)
	if err != nil {
		return err
	}
	_ = existing

	if strings.TrimSpace(g.Name) == "" {
		return domain.ErrMissingTemplateGroup
	}

	g.UserID = userID
	return s.templateGroupRepo.UpdateTemplateGroup(g)
}

func (s *templateGroupService) DeleteTemplateGroup(userID int, id int) error {
	if _, err := s.findOwned(userID, id); err != nil {
		return err
	}
	return s.templateGroupRepo.DeleteTemplateGroup(id)
}

func (s *templateGroupService) findOwned(userID int, id int) (domain.TemplateGroup, error) {
	groups, err := s.templateGroupRepo.FindTemplateGroupsByUser(userID)
	if err != nil {
		return domain.TemplateGroup{}, err
	}
	for _, g := range groups {
		if g.ID == id {
			return g, nil
		}
	}
	return domain.TemplateGroup{}, domain.ErrTemplateGroupNotFound
}
