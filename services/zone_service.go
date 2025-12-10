package services

import (
	"errors"
	"gnaps-api/models"
	"gnaps-api/repositories"
	"gnaps-api/utils"
)

type ZoneService struct {
	zoneRepo *repositories.ZoneRepository
}

func NewZoneService(zoneRepo *repositories.ZoneRepository) *ZoneService {
	return &ZoneService{zoneRepo: zoneRepo}
}

func (s *ZoneService) GetZoneByID(id uint) (*models.Zone, error) {
	zone, err := s.zoneRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("zone not found")
	}
	return zone, nil
}

func (s *ZoneService) ListZones(filters map[string]interface{}, page, limit int) ([]models.Zone, int64, error) {
	return s.zoneRepo.List(filters, page, limit)
}

func (s *ZoneService) CreateZone(zone *models.Zone) error {
	// Validate required fields
	if zone.Name == nil || *zone.Name == "" {
		return errors.New("name is required")
	}
	if zone.Code == nil || *zone.Code == "" {
		return errors.New("code is required")
	}

	// Verify region exists if region_id is provided
	if zone.RegionId != nil {
		exists, err := s.zoneRepo.VerifyRegionExists(*zone.RegionId)
		if err != nil || !exists {
			return errors.New("invalid region ID - Region does not exist")
		}
	}

	// Check if code already exists
	exists, err := s.zoneRepo.CodeExists(*zone.Code, nil)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("zone with this code already exists")
	}

	// Set defaults
	isDeleted := false
	zone.IsDeleted = &isDeleted

	return s.zoneRepo.Create(zone)
}

func (s *ZoneService) UpdateZone(id uint, updates map[string]interface{}) error {
	// Verify zone exists
	zone, err := s.zoneRepo.FindByID(id)
	if err != nil {
		return errors.New("zone not found")
	}

	// Verify region exists if region_id is being changed
	if regionId, ok := updates["region_id"]; ok {
		regionIdVal := regionId.(int64)
		if zone.RegionId == nil || regionIdVal != *zone.RegionId {
			exists, err := s.zoneRepo.VerifyRegionExists(regionIdVal)
			if err != nil || !exists {
				return errors.New("invalid region ID - Region does not exist")
			}
		}
	}

	// Check if code is being changed and if new code already exists
	if code, ok := updates["code"]; ok {
		codeStr := code.(string)
		if codeStr != "" && (zone.Code == nil || codeStr != *zone.Code) {
			exists, err := s.zoneRepo.CodeExists(codeStr, &id)
			if err != nil {
				return err
			}
			if exists {
				return errors.New("zone with this code already exists")
			}
		}
	}

	return s.zoneRepo.Update(id, updates)
}

func (s *ZoneService) DeleteZone(id uint) error {
	_, err := s.zoneRepo.FindByID(id)
	if err != nil {
		return errors.New("zone not found")
	}

	return s.zoneRepo.Delete(id)
}

// ============================================
// Role-Based Filtering Methods
// ============================================

// ListZonesWithRole returns zones filtered by role-based access
func (s *ZoneService) ListZonesWithRole(filters map[string]interface{}, page, limit int, ownerCtx *utils.OwnerContext) ([]models.Zone, int64, error) {
	regionID := ownerCtx.GetRegionIDFilter()
	zoneID := ownerCtx.GetZoneIDFilter()
	return s.zoneRepo.ListWithRoleFilter(filters, page, limit, regionID, zoneID)
}

// GetZoneByIDWithRole returns a zone if accessible by the user's role
func (s *ZoneService) GetZoneByIDWithRole(id uint, ownerCtx *utils.OwnerContext) (*models.Zone, error) {
	regionID := ownerCtx.GetRegionIDFilter()
	zoneID := ownerCtx.GetZoneIDFilter()
	zone, err := s.zoneRepo.FindByIDWithRoleFilter(id, regionID, zoneID)
	if err != nil {
		return nil, errors.New("zone not found or access denied")
	}
	return zone, nil
}
