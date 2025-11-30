package repositories

import (
	"gnaps-api/models"

	"gorm.io/gorm"
)

type NewsRepository struct {
	db *gorm.DB
}

func NewNewsRepository(db *gorm.DB) *NewsRepository {
	return &NewsRepository{db: db}
}

// Create creates a new news item
func (r *NewsRepository) Create(news *models.New) error {
	return r.db.Create(news).Error
}

// FindByID retrieves a news item by ID
func (r *NewsRepository) FindByID(id uint) (*models.New, error) {
	var news models.New
	err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&news).Error
	if err != nil {
		return nil, err
	}
	return &news, nil
}

// List retrieves all news with filters and pagination
func (r *NewsRepository) List(filters map[string]interface{}, page, limit int) ([]models.New, int64, error) {
	var news []models.New
	var total int64

	query := r.db.Where("is_deleted = ?", false)

	// Apply filters
	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}
	if category, ok := filters["category"]; ok {
		query = query.Where("category = ?", category)
	}
	if featured, ok := filters["featured"]; ok {
		query = query.Where("featured = ?", featured)
	}
	if executiveID, ok := filters["executive_id"]; ok {
		query = query.Where("executive_id = ?", executiveID)
	}
	if title, ok := filters["title"]; ok {
		query = query.Where("title LIKE ?", "%"+title.(string)+"%")
	}
	if content, ok := filters["content"]; ok {
		query = query.Where("content LIKE ?", "%"+content.(string)+"%")
	}

	// Count total before pagination
	query.Model(&models.New{}).Count(&total)

	// Apply pagination
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&news).Error
	if err != nil {
		return nil, 0, err
	}

	return news, total, nil
}

// ListAll retrieves all news without pagination (for role-based filtering)
func (r *NewsRepository) ListAll(filters map[string]interface{}) ([]models.New, error) {
	var news []models.New

	query := r.db.Where("is_deleted = ?", false)

	// Apply filters
	if status, ok := filters["status"]; ok {
		query = query.Where("status = ?", status)
	}
	if category, ok := filters["category"]; ok {
		query = query.Where("category = ?", category)
	}
	if featured, ok := filters["featured"]; ok {
		query = query.Where("featured = ?", featured)
	}
	if executiveID, ok := filters["executive_id"]; ok {
		query = query.Where("executive_id = ?", executiveID)
	}
	if title, ok := filters["title"]; ok {
		query = query.Where("title LIKE ?", "%"+title.(string)+"%")
	}
	if content, ok := filters["content"]; ok {
		query = query.Where("content LIKE ?", "%"+content.(string)+"%")
	}

	err := query.Order("created_at DESC").Find(&news).Error
	if err != nil {
		return nil, err
	}

	return news, nil
}

// Update updates a news item
func (r *NewsRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.New{}).Where("id = ?", id).Updates(updates).Error
}

// Delete soft deletes a news item
func (r *NewsRepository) Delete(id uint) error {
	trueVal := true
	return r.db.Model(&models.New{}).Where("id = ?", id).Update("is_deleted", &trueVal).Error
}

// GetZonesByRegion retrieves all zones under a region
func (r *NewsRepository) GetZonesByRegion(regionID int64) ([]models.Zone, error) {
	var zones []models.Zone
	err := r.db.Where("region_id = ? AND is_deleted = ?", regionID, false).Find(&zones).Error
	return zones, err
}

// GetSchoolsByZone retrieves all schools under a zone
func (r *NewsRepository) GetSchoolsByZone(zoneID int64) ([]models.School, error) {
	var schools []models.School
	err := r.db.Where("zone_id = ? AND is_deleted = ?", zoneID, false).Find(&schools).Error
	return schools, err
}

// GetZoneByID retrieves a zone by ID
func (r *NewsRepository) GetZoneByID(id int64) (*models.Zone, error) {
	var zone models.Zone
	err := r.db.Where("id = ?", id).First(&zone).Error
	if err != nil {
		return nil, err
	}
	return &zone, nil
}

// GetSchoolByID retrieves a school by ID
func (r *NewsRepository) GetSchoolByID(id int64) (*models.School, error) {
	var school models.School
	err := r.db.Where("id = ?", id).First(&school).Error
	if err != nil {
		return nil, err
	}
	return &school, nil
}
