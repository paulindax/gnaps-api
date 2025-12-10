package repositories

import (
	"gnaps-api/models"

	"gorm.io/gorm"
)

type ZoneRepository struct {
	db *gorm.DB
}

func NewZoneRepository(db *gorm.DB) *ZoneRepository {
	return &ZoneRepository{db: db}
}

func (r *ZoneRepository) FindByID(id uint) (*models.Zone, error) {
	var zone models.Zone
	err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&zone).Error
	if err != nil {
		return nil, err
	}
	return &zone, nil
}

func (r *ZoneRepository) List(filters map[string]interface{}, page, limit int) ([]models.Zone, int64, error) {
	var zones []models.Zone
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

	query.Model(&models.Zone{}).Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&zones).Error

	return zones, total, err
}

func (r *ZoneRepository) Create(zone *models.Zone) error {
	return r.db.Create(zone).Error
}

func (r *ZoneRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.Zone{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ZoneRepository) Delete(id uint) error {
	trueVal := true
	return r.db.Model(&models.Zone{}).Where("id = ?", id).Update("is_deleted", &trueVal).Error
}

func (r *ZoneRepository) CodeExists(code string, excludeID *uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.Zone{}).Where("code = ? AND is_deleted = ?", code, false)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

func (r *ZoneRepository) VerifyRegionExists(regionID int64) (bool, error) {
	var count int64
	err := r.db.Model(&models.Region{}).Where("id = ? AND is_deleted = ?", regionID, false).Count(&count).Error
	return count > 0, err
}

// ListWithRoleFilter returns zones filtered by role-based access
// - system_admin/national_admin: all zones
// - region_admin: zones in their region
// - zone_admin: only their zone
func (r *ZoneRepository) ListWithRoleFilter(filters map[string]interface{}, page, limit int, regionID, zoneID *int64) ([]models.Zone, int64, error) {
	var zones []models.Zone
	var total int64

	query := r.db.Where("is_deleted = ?", false)

	// Apply role-based filtering
	if zoneID != nil {
		// zone_admin can only see their own zone
		query = query.Where("id = ?", *zoneID)
	} else if regionID != nil {
		// region_admin can see all zones in their region
		query = query.Where("region_id = ?", *regionID)
	}
	// system_admin and national_admin see all zones (no additional filter)

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

	query.Model(&models.Zone{}).Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&zones).Error

	return zones, total, err
}

// FindByIDWithRoleFilter returns a zone if accessible by the user's role
func (r *ZoneRepository) FindByIDWithRoleFilter(id uint, regionID, zoneID *int64) (*models.Zone, error) {
	var zone models.Zone
	query := r.db.Where("id = ? AND is_deleted = ?", id, false)

	// Apply role-based filtering
	if zoneID != nil {
		// zone_admin can only see their own zone
		query = query.Where("id = ?", *zoneID)
	} else if regionID != nil {
		// region_admin can see zones in their region
		query = query.Where("region_id = ?", *regionID)
	}

	err := query.First(&zone).Error
	if err != nil {
		return nil, err
	}
	return &zone, nil
}
