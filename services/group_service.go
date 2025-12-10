package services

import (
	"errors"
	"gnaps-api/models"
	"gnaps-api/repositories"
	"gnaps-api/utils"
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

// ============================================
// Owner-based methods for data filtering
// ============================================

func (s *GroupService) GetGroupByIDWithOwner(id uint, ownerCtx *utils.OwnerContext) (*models.SchoolGroup, error) {
	return s.groupRepo.FindByIDWithOwner(id, ownerCtx)
}

func (s *GroupService) ListGroupsWithOwner(filters map[string]interface{}, page, limit int, ownerCtx *utils.OwnerContext) ([]models.SchoolGroup, int64, error) {
	return s.groupRepo.ListWithOwner(filters, page, limit, ownerCtx)
}

func (s *GroupService) CreateGroupWithOwner(group *models.SchoolGroup, ownerCtx *utils.OwnerContext) error {
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

	return s.groupRepo.CreateWithOwner(group, ownerCtx)
}

func (s *GroupService) UpdateGroupWithOwner(id uint, updates map[string]interface{}, ownerCtx *utils.OwnerContext) error {
	// Verify zone exists if zone_id is being changed
	if zoneId, ok := updates["zone_id"]; ok {
		zoneIdVal := zoneId.(int64)
		exists, err := s.groupRepo.VerifyZoneExists(zoneIdVal)
		if err != nil || !exists {
			return errors.New("invalid zone ID - Zone does not exist")
		}
	}

	return s.groupRepo.UpdateWithOwner(id, updates, ownerCtx)
}

func (s *GroupService) DeleteGroupWithOwner(id uint, ownerCtx *utils.OwnerContext) error {
	return s.groupRepo.DeleteWithOwner(id, ownerCtx)
}
