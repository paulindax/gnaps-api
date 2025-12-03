package repositories

import (
	"gnaps-api/models"

	"gorm.io/gorm"
)

type BillItemRepository struct {
	db *gorm.DB
}

func NewBillItemRepository(db *gorm.DB) *BillItemRepository {
	return &BillItemRepository{db: db}
}

func (r *BillItemRepository) FindByID(id uint) (*models.BillItem, error) {
	var billItem models.BillItem
	err := r.db.Preload("BillParticular").
		Where("id = ?", id).
		First(&billItem).Error
	if err != nil {
		return nil, err
	}
	// Check if deleted
	if billItem.IsDeleted != nil && *billItem.IsDeleted {
		return nil, gorm.ErrRecordNotFound
	}
	return &billItem, nil
}

func (r *BillItemRepository) FindByBillID(billId uint) ([]models.BillItem, error) {
	var billItems []models.BillItem
	err := r.db.Preload("BillParticular").
		Where("bill_id = ? AND (is_deleted = ? OR is_deleted IS NULL)", billId, false).
		Order("created_at ASC").
		Find(&billItems).Error
	return billItems, err
}

func (r *BillItemRepository) FindByBillIDWithPagination(billId uint, page, limit int, search string) ([]models.BillItem, int64, error) {
	var billItems []models.BillItem
	var total int64

	query := r.db.Model(&models.BillItem{}).
		Preload("BillParticular").
		Where("bill_items.bill_id = ? AND (bill_items.is_deleted = ? OR bill_items.is_deleted IS NULL)", billId, false)

	// Add search filter if provided
	if search != "" {
		query = query.Joins("JOIN bill_particulars ON bill_particulars.id = bill_items.bill_particular_id").
			Where("bill_particulars.name LIKE ?", "%"+search+"%")
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Fetch records with pagination
	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("bill_items.created_at ASC").Find(&billItems).Error

	return billItems, total, err
}

func (r *BillItemRepository) Create(billItem *models.BillItem) error {
	return r.db.Create(billItem).Error
}

func (r *BillItemRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.BillItem{}).Where("id = ?", id).Updates(updates).Error
}

func (r *BillItemRepository) Delete(id uint) error {
	trueVal := true
	return r.db.Model(&models.BillItem{}).Where("id = ?", id).Update("is_deleted", &trueVal).Error
}
