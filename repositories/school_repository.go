package repositories

import (
	"gnaps-api/models"

	"gorm.io/gorm"
)

type SchoolRepository struct {
	db *gorm.DB
}

func NewSchoolRepository(db *gorm.DB) *SchoolRepository {
	return &SchoolRepository{db: db}
}

func (r *SchoolRepository) Search(keyword string, limit int) ([]models.School, error) {
	var schools []models.School
	query := r.db.Model(&models.School{}).Where("is_deleted = ?", false)

	if keyword != "" {
		query = query.Where("name LIKE ?", "%"+keyword+"%")
	}

	err := query.Limit(limit).Order("name ASC").Find(&schools).Error
	return schools, err
}

func (r *SchoolRepository) FindByID(id uint) (*models.School, error) {
	var school models.School
	err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&school).Error
	if err != nil {
		return nil, err
	}
	return &school, nil
}

func (r *SchoolRepository) List(filters map[string]interface{}, page, limit int) ([]models.School, int64, error) {
	var schools []models.School
	var total int64

	query := r.db.Model(&models.School{}).Where("is_deleted = ?", false)

	// Apply filters
	for key, value := range filters {
		// Handle LIKE queries for name and member_no
		if key == "name" || key == "member_no" {
			query = query.Where(key+" LIKE ?", "%"+value.(string)+"%")
		} else {
			query = query.Where(key+" = ?", value)
		}
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&schools).Error

	return schools, total, err
}

func (r *SchoolRepository) Create(school *models.School) error {
	return r.db.Create(school).Error
}

func (r *SchoolRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.School{}).Where("id = ?", id).Updates(updates).Error
}

func (r *SchoolRepository) Delete(id uint) error {
	return r.db.Model(&models.School{}).Where("id = ?", id).Update("is_deleted", true).Error
}

// MemberNoExists checks if a member number already exists
func (r *SchoolRepository) MemberNoExists(memberNo string, excludeID *uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.School{}).Where("member_no = ? AND is_deleted = ?", memberNo, false)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// EmailExists checks if an email already exists
func (r *SchoolRepository) EmailExists(email string, excludeID *uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.School{}).Where("email = ? AND is_deleted = ?", email, false)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

// VerifyZoneExists checks if a zone exists
func (r *SchoolRepository) VerifyZoneExists(zoneID int64) (bool, error) {
	var count int64
	err := r.db.Model(&models.Zone{}).Where("id = ? AND is_deleted = ?", zoneID, false).Count(&count).Error
	return count > 0, err
}
