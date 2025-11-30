package repositories

import (
	"gnaps-api/models"

	"gorm.io/gorm"
)

type PositionRepository struct {
	db *gorm.DB
}

func NewPositionRepository(db *gorm.DB) *PositionRepository {
	return &PositionRepository{db: db}
}

func (r *PositionRepository) FindByID(id uint) (*models.Position, error) {
	var position models.Position
	err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&position).Error
	if err != nil {
		return nil, err
	}
	return &position, nil
}

func (r *PositionRepository) List(filters map[string]interface{}, page, limit int) ([]models.Position, int64, error) {
	var positions []models.Position
	var total int64

	query := r.db.Where("is_deleted = ?", false)

	// Handle search filter
	if search, ok := filters["search"]; ok {
		searchPattern := "%" + search.(string) + "%"
		query = query.Where("name LIKE ?", searchPattern)
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

	query.Model(&models.Position{}).Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&positions).Error

	return positions, total, err
}

func (r *PositionRepository) Create(position *models.Position) error {
	return r.db.Create(position).Error
}

func (r *PositionRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.Position{}).Where("id = ?", id).Updates(updates).Error
}

func (r *PositionRepository) Delete(id uint) error {
	trueVal := true
	return r.db.Model(&models.Position{}).Where("id = ?", id).Update("is_deleted", &trueVal).Error
}

func (r *PositionRepository) NameExists(name string, excludeID *uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.Position{}).Where("name = ? AND is_deleted = ?", name, false)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}

func (r *PositionRepository) IsPositionInUse(id uint) (bool, int64, error) {
	var count int64
	err := r.db.Model(&models.Executive{}).Where("position_id = ? AND is_deleted = ?", id, 0).Count(&count).Error
	return count > 0, count, err
}
