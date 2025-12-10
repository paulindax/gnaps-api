package repositories

import (
	"gnaps-api/models"

	"gorm.io/gorm"
)

type RegionRepository struct {
	db *gorm.DB
}

func NewRegionRepository(db *gorm.DB) *RegionRepository {
	return &RegionRepository{db: db}
}

func (r *RegionRepository) FindByID(id uint) (*models.Region, error) {
	var region models.Region
	err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&region).Error
	if err != nil {
		return nil, err
	}
	return &region, nil
}

func (r *RegionRepository) List(filters map[string]interface{}, page, limit int) ([]models.Region, int64, error) {
	var regions []models.Region
	var total int64

	query := r.db.Where("is_deleted = ?", false)

	// Handle search filter (searches both name and code)
	if search, ok := filters["search"]; ok {
		searchPattern := "%" + search.(string) + "%"
		query = query.Where("name LIKE ? OR code LIKE ?", searchPattern, searchPattern)
		delete(filters, "search")
	}

	// Apply other filters
	for key, value := range filters {
		if key == "name" {
			query = query.Where("name LIKE ?", "%"+value.(string)+"%")
		} else {
			query = query.Where(key+" = ?", value)
		}
	}

	query.Model(&models.Region{}).Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&regions).Error

	return regions, total, err
}

func (r *RegionRepository) Create(region *models.Region) error {
	return r.db.Create(region).Error
}

func (r *RegionRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.Region{}).Where("id = ?", id).Updates(updates).Error
}

func (r *RegionRepository) Delete(id uint) error {
	trueVal := true
	return r.db.Model(&models.Region{}).Where("id = ?", id).Update("is_deleted", &trueVal).Error
}

func (r *RegionRepository) CodeExists(code string, excludeID *uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.Region{}).Where("code = ? AND is_deleted = ?", code, false)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// ListWithRoleFilter returns regions filtered by role-based access
// - system_admin/national_admin: all regions
// - region_admin: only their region
// - zone_admin: the region containing their zone
func (r *RegionRepository) ListWithRoleFilter(filters map[string]interface{}, page, limit int, regionID, zoneID *int64) ([]models.Region, int64, error) {
	var regions []models.Region
	var total int64

	query := r.db.Where("is_deleted = ?", false)

	// Apply role-based filtering
	if regionID != nil {
		// region_admin can only see their own region
		query = query.Where("id = ?", *regionID)
	} else if zoneID != nil {
		// zone_admin can see the region their zone belongs to
		query = query.Where("id IN (SELECT region_id FROM zones WHERE id = ? AND is_deleted = ?)", *zoneID, false)
	}
	// system_admin and national_admin see all regions (no additional filter)

	// Handle search filter (searches both name and code)
	if search, ok := filters["search"]; ok {
		searchPattern := "%" + search.(string) + "%"
		query = query.Where("name LIKE ? OR code LIKE ?", searchPattern, searchPattern)
		delete(filters, "search")
	}

	// Apply other filters
	for key, value := range filters {
		if key == "name" {
			query = query.Where("name LIKE ?", "%"+value.(string)+"%")
		} else {
			query = query.Where(key+" = ?", value)
		}
	}

	query.Model(&models.Region{}).Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&regions).Error

	return regions, total, err
}

// FindByIDWithRoleFilter returns a region if accessible by the user's role
func (r *RegionRepository) FindByIDWithRoleFilter(id uint, regionID, zoneID *int64) (*models.Region, error) {
	var region models.Region
	query := r.db.Where("id = ? AND is_deleted = ?", id, false)

	// Apply role-based filtering
	if regionID != nil {
		// region_admin can only see their own region
		query = query.Where("id = ?", *regionID)
	} else if zoneID != nil {
		// zone_admin can see the region their zone belongs to
		query = query.Where("id IN (SELECT region_id FROM zones WHERE id = ? AND is_deleted = ?)", *zoneID, false)
	}

	err := query.First(&region).Error
	if err != nil {
		return nil, err
	}
	return &region, nil
}
