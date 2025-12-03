package repositories

import (
	"gnaps-api/models"

	"gorm.io/gorm"
)

type FinanceAccountRepository struct {
	db *gorm.DB
}

func NewFinanceAccountRepository(db *gorm.DB) *FinanceAccountRepository {
	return &FinanceAccountRepository{db: db}
}

func (r *FinanceAccountRepository) FindByID(id uint) (*models.FinanceAccount, error) {
	var account models.FinanceAccount
	err := r.db.Where("id = ? AND is_deleted = ?", id, false).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}

func (r *FinanceAccountRepository) List(filters map[string]interface{}, page, limit int) ([]models.FinanceAccount, int64, error) {
	var accounts []models.FinanceAccount
	var total int64

	query := r.db.Where("is_deleted = ?", false)

	// Handle search filter (searches name, code, and description)
	if search, ok := filters["search"]; ok {
		searchPattern := "%" + search.(string) + "%"
		query = query.Where("name LIKE ? OR code LIKE ? OR description LIKE ?", searchPattern, searchPattern, searchPattern)
		delete(filters, "search")
	}

	// Apply other filters
	for key, value := range filters {
		if key == "name" || key == "description" {
			query = query.Where(key+" LIKE ?", "%"+value.(string)+"%")
		} else {
			query = query.Where(key+" = ?", value)
		}
	}

	// Count total - use Model to specify table for count
	countQuery := query.Model(&models.FinanceAccount{})
	if err := countQuery.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch records with pagination
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&accounts).Error

	return accounts, total, err
}

func (r *FinanceAccountRepository) Create(account *models.FinanceAccount) error {
	return r.db.Create(account).Error
}

func (r *FinanceAccountRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.FinanceAccount{}).Where("id = ?", id).Updates(updates).Error
}

func (r *FinanceAccountRepository) Delete(id uint) error {
	trueVal := true
	return r.db.Model(&models.FinanceAccount{}).Where("id = ?", id).Update("is_deleted", &trueVal).Error
}

func (r *FinanceAccountRepository) CodeExists(code string, excludeID *uint) (bool, error) {
	var count int64
	query := r.db.Model(&models.FinanceAccount{}).Where("code = ? AND is_deleted = ?", code, false)
	if excludeID != nil {
		query = query.Where("id != ?", *excludeID)
	}
	err := query.Count(&count).Error
	return count > 0, err
}
