package services

import (
	"errors"
	"gnaps-api/models"
	"gnaps-api/repositories"
	"gnaps-api/utils"
)

type RegionService struct {
	regionRepo *repositories.RegionRepository
}

func NewRegionService(regionRepo *repositories.RegionRepository) *RegionService {
	return &RegionService{regionRepo: regionRepo}
}

func (s *RegionService) GetRegionByID(id uint) (*models.Region, error) {
	region, err := s.regionRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("region not found")
	}
	return region, nil
}

func (s *RegionService) ListRegions(filters map[string]interface{}, page, limit int) ([]models.Region, int64, error) {
	return s.regionRepo.List(filters, page, limit)
}

func (s *RegionService) CreateRegion(region *models.Region) error {
	// Validate required fields
	if region.Name == nil || *region.Name == "" {
		return errors.New("name is required")
	}
	if region.Code == nil || *region.Code == "" {
		return errors.New("code is required")
	}

	// Check if code already exists
	exists, err := s.regionRepo.CodeExists(*region.Code, nil)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("region with this code already exists")
	}

	// Set defaults
	isDeleted := false
	region.IsDeleted = &isDeleted

	return s.regionRepo.Create(region)
}

func (s *RegionService) UpdateRegion(id uint, updates map[string]interface{}) error {
	// Verify region exists
	region, err := s.regionRepo.FindByID(id)
	if err != nil {
		return errors.New("region not found")
	}

	// Check if code is being changed and if new code already exists
	if code, ok := updates["code"]; ok {
		codeStr := code.(string)
		if codeStr != "" && (region.Code == nil || codeStr != *region.Code) {
			exists, err := s.regionRepo.CodeExists(codeStr, &id)
			if err != nil {
				return err
			}
			if exists {
				return errors.New("region with this code already exists")
			}
		}
	}

	return s.regionRepo.Update(id, updates)
}

func (s *RegionService) DeleteRegion(id uint) error {
	_, err := s.regionRepo.FindByID(id)
	if err != nil {
		return errors.New("region not found")
	}

	return s.regionRepo.Delete(id)
}

// ============================================
// Role-Based Filtering Methods
// ============================================

// ListRegionsWithRole returns regions filtered by role-based access
func (s *RegionService) ListRegionsWithRole(filters map[string]interface{}, page, limit int, ownerCtx *utils.OwnerContext) ([]models.Region, int64, error) {
	regionID := ownerCtx.GetRegionIDFilter()
	zoneID := ownerCtx.GetZoneIDFilter()
	return s.regionRepo.ListWithRoleFilter(filters, page, limit, regionID, zoneID)
}

// GetRegionByIDWithRole returns a region if accessible by the user's role
func (s *RegionService) GetRegionByIDWithRole(id uint, ownerCtx *utils.OwnerContext) (*models.Region, error) {
	regionID := ownerCtx.GetRegionIDFilter()
	zoneID := ownerCtx.GetZoneIDFilter()
	region, err := s.regionRepo.FindByIDWithRoleFilter(id, regionID, zoneID)
	if err != nil {
		return nil, errors.New("region not found or access denied")
	}
	return region, nil
}
