package repositories

import (
	"gnaps-api/models"

	"gorm.io/gorm"
)

type BillParticularRepository struct {
	db *gorm.DB
}

func NewBillParticularRepository(db *gorm.DB) *BillParticularRepository {
	return &BillParticularRepository{db: db}
}

func (r *BillParticularRepository) FindByID(id uint) (*models.BillParticular, error) {
	var particular models.BillParticular
	err := r.db.Where("id = ?", id).
		First(&particular).Error
	if err != nil {
		return nil, err
	}
	// Check if deleted
	if particular.IsDeleted != nil && *particular.IsDeleted {
		return nil, gorm.ErrRecordNotFound
	}
	return &particular, nil
}

func (r *BillParticularRepository) List(filters map[string]interface{}, page, limit int) ([]models.BillParticular, int64, error) {
	var particulars []models.BillParticular
	var total int64

	query := r.db.Model(&models.BillParticular{})

	// Filter out deleted records
	query = query.Where("is_deleted = ? OR is_deleted IS NULL", false)

	// Handle search filter (searches name)
	if search, ok := filters["search"]; ok {
		searchPattern := "%" + search.(string) + "%"
		query = query.Where("name LIKE ?", searchPattern)
		delete(filters, "search")
	}

	// Apply other filters
	for key, value := range filters {
		if key == "name" {
			query = query.Where("name LIKE ?", "%"+value.(string)+"%")
		} else if key == "finance_account_id" {
			query = query.Where("finance_account_id = ?", value)
		} else {
			query = query.Where(key+" = ?", value)
		}
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).
		Order("priority ASC, created_at DESC").
		Find(&particulars).Error

	return particulars, total, err
}

func (r *BillParticularRepository) Create(particular *models.BillParticular) error {
	return r.db.Create(particular).Error
}

func (r *BillParticularRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.BillParticular{}).Where("id = ?", id).Updates(updates).Error
}

func (r *BillParticularRepository) Delete(id uint) error {
	trueVal := true
	return r.db.Model(&models.BillParticular{}).Where("id = ?", id).Update("is_deleted", &trueVal).Error
}

func (r *BillParticularRepository) GetMaxPriority() (int, error) {
	var maxPriority *int
	err := r.db.Model(&models.BillParticular{}).
		Where("is_deleted = ? OR is_deleted IS NULL", false).
		Select("MAX(priority)").
		Scan(&maxPriority).Error

	if maxPriority == nil {
		return 0, err
	}
	return *maxPriority, err
}
