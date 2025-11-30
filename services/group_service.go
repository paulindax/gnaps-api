package services

import (
	"errors"
	"gnaps-api/models"
	"gnaps-api/repositories"
)

type GroupService struct {
	groupRepo *repositories.GroupRepository
}

func NewGroupService(groupRepo *repositories.GroupRepository) *GroupService {
	return &GroupService{groupRepo: groupRepo}
}

func (s *GroupService) GetGroupByID(id uint) (*models.SchoolGroup, error) {
	group, err := s.groupRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("group not found")
	}
	return group, nil
}

func (s *GroupService) ListGroups(filters map[string]interface{}, page, limit int) ([]models.SchoolGroup, int64, error) {
	return s.groupRepo.List(filters, page, limit)
}

func (s *GroupService) CreateGroup(group *models.SchoolGroup) error {
	// Validate required fields
	if group.Name == nil || *group.Name == "" {
		return errors.New("name is required")
	}

	// Verify zone exists if zone_id is provided
	if group.ZoneId != nil {
		exists, err := s.groupRepo.VerifyZoneExists(*group.ZoneId)
		if err != nil || !exists {
			return errors.New("invalid zone ID - Zone does not exist")
		}
	}

	// Set defaults
	group.IsDeleted = false

	return s.groupRepo.Create(group)
}

func (s *GroupService) UpdateGroup(id uint, updates map[string]interface{}) error {
	// Verify group exists
	group, err := s.groupRepo.FindByID(id)
	if err != nil {
		return errors.New("group not found")
	}

	// Verify zone exists if zone_id is being changed
	if zoneId, ok := updates["zone_id"]; ok {
		zoneIdVal := zoneId.(int64)
		if group.ZoneId == nil || zoneIdVal != *group.ZoneId {
			exists, err := s.groupRepo.VerifyZoneExists(zoneIdVal)
			if err != nil || !exists {
				return errors.New("invalid zone ID - Zone does not exist")
			}
		}
	}

	return s.groupRepo.Update(id, updates)
}

func (s *GroupService) DeleteGroup(id uint) error {
	_, err := s.groupRepo.FindByID(id)
	if err != nil {
		return errors.New("group not found")
	}

	return s.groupRepo.Delete(id)
}
