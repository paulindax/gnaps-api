package repositories

import (
	"gnaps-api/models"

	"gorm.io/gorm"
)

type ExecutiveRepository struct {
	db *gorm.DB
}

func NewExecutiveRepository(db *gorm.DB) *ExecutiveRepository {
	return &ExecutiveRepository{db: db}
}

func (r *ExecutiveRepository) FindByID(id uint) (*models.Executive, error) {
	var executive models.Executive
	err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&executive).Error
	if err != nil {
		return nil, err
	}
	return &executive, nil
}

func (r *ExecutiveRepository) List(filters map[string]interface{}, page, limit int) ([]models.Executive, int64, error) {
	var executives []models.Executive
	var total int64

	query := r.db.Where("is_deleted = ?", false)

	// Apply filters
	for key, value := range filters {
		if key == "first_name" || key == "last_name" || key == "executive_no" {
			query = query.Where(key+" LIKE ?", "%"+value.(string)+"%")
		} else if key == "name" {
			query = query.Where("CONCAT(first_name, ' ', IFNULL(middle_name, ''), ' ', last_name) LIKE ?", "%"+value.(string)+"%")
		} else {
			query = query.Where(key+" = ?", value)
		}
	}

	query.Model(&models.Executive{}).Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&executives).Error

	return executives, total, err
}

func (r *ExecutiveRepository) Create(executive *models.Executive) error {
	return r.db.Create(executive).Error
}

func (r *ExecutiveRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.Executive{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ExecutiveRepository) Delete(id uint) error {
	trueVal := true
	return r.db.Model(&models.Executive{}).Where("id = ?", id).Update("is_deleted", &trueVal).Error
}

func (r *ExecutiveRepository) ExecutiveNoExists(executiveNo string, excludeID *uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.Executive{}).Where("executive_no = ? AND is_deleted = ?", executiveNo, false)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

func (r *ExecutiveRepository) EmailExists(email string, excludeID *uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.Executive{}).Where("email = ? AND is_deleted = ?", email, false)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}
