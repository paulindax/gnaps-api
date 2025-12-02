package repositories

import (
	"gnaps-api/models"

	"gorm.io/gorm"
)

type BillAssignmentRepository struct {
	db *gorm.DB
}

func NewBillAssignmentRepository(db *gorm.DB) *BillAssignmentRepository {
	return &BillAssignmentRepository{db: db}
}

func (r *BillAssignmentRepository) FindByID(id uint) (*models.BillAssignment, error) {
	var assignment models.BillAssignment
	err := r.db.Where("id = ?", id).
		First(&assignment).Error
	if err != nil {
		return nil, err
	}
	return &assignment, nil
}

func (r *BillAssignmentRepository) FindByBillItemID(billItemId uint) ([]models.BillAssignment, error) {
	var assignments []models.BillAssignment
	err := r.db.Where("bill_item_id = ?", billItemId).
		Find(&assignments).Error
	return assignments, err
}

func (r *BillAssignmentRepository) Create(assignment *models.BillAssignment) error {
	return r.db.Create(assignment).Error
}

func (r *BillAssignmentRepository) BulkCreate(assignments []models.BillAssignment) error {
	if len(assignments) == 0 {
		return nil
	}
	return r.db.Create(&assignments).Error
}

func (r *BillAssignmentRepository) Delete(id uint) error {
	return r.db.Delete(&models.BillAssignment{}, id).Error
}

func (r *BillAssignmentRepository) DeleteByBillItemID(billItemId uint) error {
	return r.db.Where("bill_item_id = ?", billItemId).Delete(&models.BillAssignment{}).Error
}
