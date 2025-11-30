package services

import (
	"errors"
	"fmt"
	"gnaps-api/models"
	"gnaps-api/repositories"
)

type PositionService struct {
	positionRepo *repositories.PositionRepository
}

func NewPositionService(positionRepo *repositories.PositionRepository) *PositionService {
	return &PositionService{positionRepo: positionRepo}
}

func (s *PositionService) GetPositionByID(id uint) (*models.Position, error) {
	position, err := s.positionRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("position not found")
	}
	return position, nil
}

func (s *PositionService) ListPositions(filters map[string]interface{}, page, limit int) ([]models.Position, int64, error) {
	return s.positionRepo.List(filters, page, limit)
}

func (s *PositionService) CreatePosition(position *models.Position) error {
	// Validate required fields
	if position.Name == nil || *position.Name == "" {
		return errors.New("name is required")
	}

	// Check if name already exists
	exists, err := s.positionRepo.NameExists(*position.Name, nil)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("position with this name already exists")
	}

	// Set defaults
	falseVal := false
	position.IsDeleted = &falseVal

	return s.positionRepo.Create(position)
}

func (s *PositionService) UpdatePosition(id uint, updates map[string]interface{}) error {
	// Verify position exists
	position, err := s.positionRepo.FindByID(id)
	if err != nil {
		return errors.New("position not found")
	}

	// Check if name is being changed and if new name already exists
	if name, ok := updates["name"]; ok {
		nameStr := name.(*string)
		if nameStr != nil && *nameStr != "" && (position.Name == nil || *nameStr != *position.Name) {
			exists, err := s.positionRepo.NameExists(*nameStr, &id)
			if err != nil {
				return err
			}
			if exists {
				return errors.New("position with this name already exists")
			}
		}
	}

	return s.positionRepo.Update(id, updates)
}

func (s *PositionService) DeletePosition(id uint) error {
	// Verify position exists
	_, err := s.positionRepo.FindByID(id)
	if err != nil {
		return errors.New("position not found")
	}

	// Check if position is in use
	inUse, count, err := s.positionRepo.IsPositionInUse(id)
	if err != nil {
		return err
	}
	if inUse {
		return fmt.Errorf("cannot delete position - currently assigned to %d executive(s)", count)
	}

	return s.positionRepo.Delete(id)
}
