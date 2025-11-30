package repositories

import (
	"gnaps-api/models"

	"gorm.io/gorm"
)

type GroupRepository struct {
	db *gorm.DB
}

func NewGroupRepository(db *gorm.DB) *GroupRepository {
	return &GroupRepository{db: db}
}

func (r *GroupRepository) FindByID(id uint) (*models.SchoolGroup, error) {
	var group models.SchoolGroup
	err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&group).Error
	if err != nil {
		return nil, err
	}
	return &group, nil
}

func (r *GroupRepository) List(filters map[string]interface{}, page, limit int) ([]models.SchoolGroup, int64, error) {
	var groups []models.SchoolGroup
	var total int64

	query := r.db.Where("is_deleted = ?", false)

	// Handle search filter (searches name and description)
	if search, ok := filters["search"]; ok {
		searchPattern := "%" + search.(string) + "%"
		query = query.Where("name LIKE ? OR (description IS NOT NULL AND description LIKE ?)", searchPattern, searchPattern)
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

	query.Model(&models.SchoolGroup{}).Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&groups).Error

	return groups, total, err
}

func (r *GroupRepository) Create(group *models.SchoolGroup) error {
	return r.db.Create(group).Error
}

func (r *GroupRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.SchoolGroup{}).Where("id = ?", id).Updates(updates).Error
}

func (r *GroupRepository) Delete(id uint) error {
	trueVal := true
	return r.db.Model(&models.SchoolGroup{}).Where("id = ?", id).Update("is_deleted", &trueVal).Error
}

func (r *GroupRepository) VerifyZoneExists(zoneID int64) (bool, error) {
	var count int64
	err := r.db.Model(&models.Zone{}).Where("id = ? AND is_deleted = ?", zoneID, false).Count(&count).Error
	return count > 0, err
}
