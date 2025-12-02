package repositories

import (
	"gnaps-api/models"

	"gorm.io/gorm"
)

type BillRepository struct {
	db *gorm.DB
}

func NewBillRepository(db *gorm.DB) *BillRepository {
	return &BillRepository{db: db}
}

func (r *BillRepository) FindByID(id uint) (*models.Bill, error) {
	var bill models.Bill
	err := r.db.Where("id = ?", id).
		First(&bill).Error
	if err != nil {
		return nil, err
	}
	// Check if deleted
	if bill.IsDeleted {
		return nil, gorm.ErrRecordNotFound
	}
	return &bill, nil
}

func (r *BillRepository) List(filters map[string]interface{}, page, limit int) ([]models.Bill, int64, error) {
	var bills []models.Bill
	var total int64

	query := r.db.Model(&models.Bill{})

	// Filter out deleted records
	query = query.Where("is_deleted = ? OR is_deleted IS NULL", false)

	// Handle search filter
	if search, ok := filters["search"]; ok {
		searchPattern := "%" + search.(string) + "%"
		query = query.Where("name LIKE ? OR description LIKE ?", searchPattern, searchPattern)
		delete(filters, "search")
	}

	// Apply other filters
	for key, value := range filters {
		switch key {
		case "name":
			query = query.Where("name LIKE ?", "%"+value.(string)+"%")
		case "academic_year":
			query = query.Where("academic_year = ?", value)
		case "term":
			query = query.Where("term = ?", value)
		case "status":
			query = query.Where("status = ?", value)
		default:
			query = query.Where(key+" = ?", value)
		}
	}

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).
		Order("created_at DESC").
		Find(&bills).Error

	return bills, total, err
}

func (r *BillRepository) Create(bill *models.Bill) error {
	return r.db.Create(bill).Error
}

func (r *BillRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.Bill{}).Where("id = ?", id).Updates(updates).Error
}

func (r *BillRepository) Delete(id uint) error {
	trueVal := true
	return r.db.Model(&models.Bill{}).Where("id = ?", id).Update("is_deleted", &trueVal).Error
}
