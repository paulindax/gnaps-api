package services

import (
	"errors"
	"gnaps-api/models"
	"gnaps-api/repositories"
)

type BillService struct {
	billRepo           *repositories.BillRepository
	billItemRepo       *repositories.BillItemRepository
	billAssignmentRepo *repositories.BillAssignmentRepository
}

func NewBillService(
	billRepo *repositories.BillRepository,
	billItemRepo *repositories.BillItemRepository,
	billAssignmentRepo *repositories.BillAssignmentRepository,
) *BillService {
	return &BillService{
		billRepo:           billRepo,
		billItemRepo:       billItemRepo,
		billAssignmentRepo: billAssignmentRepo,
	}
}

// Bill operations
func (s *BillService) GetBillByID(id uint) (*models.Bill, error) {
	bill, err := s.billRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("bill not found")
	}
	return bill, nil
}

func (s *BillService) ListBills(filters map[string]interface{}, page, limit int) ([]models.Bill, int64, error) {
	return s.billRepo.List(filters, page, limit)
}

func (s *BillService) CreateBill(bill *models.Bill) error {
	// Validate required fields
	if bill.Name == nil || *bill.Name == "" {
		return errors.New("name is required")
	}

	// Set defaults
	bill.IsDeleted = false

	return s.billRepo.Create(bill)
}

func (s *BillService) UpdateBill(id uint, updates map[string]interface{}) error {
	// Verify bill exists
	_, err := s.billRepo.FindByID(id)
	if err != nil {
		return errors.New("bill not found")
	}

	return s.billRepo.Update(id, updates)
}

func (s *BillService) DeleteBill(id uint) error {
	_, err := s.billRepo.FindByID(id)
	if err != nil {
		return errors.New("bill not found")
	}

	return s.billRepo.Delete(id)
}

// Bill Item operations
func (s *BillService) GetBillItemByID(id uint) (*models.BillItem, error) {
	billItem, err := s.billItemRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("bill item not found")
	}
	return billItem, nil
}

func (s *BillService) GetBillItemsByBillID(billId uint) ([]models.BillItem, error) {
	return s.billItemRepo.FindByBillID(billId)
}

func (s *BillService) CreateBillItem(billItem *models.BillItem) error {
	// Validate required fields
	if billItem.BillId == nil {
		return errors.New("bill_id is required")
	}
	if billItem.BillParticularId == nil {
		return errors.New("bill_particular_id is required")
	}
	if billItem.Amount == nil {
		return errors.New("amount is required")
	}

	// Set defaults
	isDeleted := false
	billItem.IsDeleted = &isDeleted

	return s.billItemRepo.Create(billItem)
}

func (s *BillService) UpdateBillItem(id uint, updates map[string]interface{}) error {
	// Verify bill item exists
	_, err := s.billItemRepo.FindByID(id)
	if err != nil {
		return errors.New("bill item not found")
	}

	return s.billItemRepo.Update(id, updates)
}

func (s *BillService) DeleteBillItem(id uint) error {
	_, err := s.billItemRepo.FindByID(id)
	if err != nil {
		return errors.New("bill item not found")
	}

	// Also delete all assignments for this bill item
	_ = s.billAssignmentRepo.DeleteByBillItemID(id)

	return s.billItemRepo.Delete(id)
}

// Bill Assignment operations
func (s *BillService) GetAssignmentsByBillItemID(billItemId uint) ([]models.BillAssignment, error) {
	return s.billAssignmentRepo.FindByBillItemID(billItemId)
}

func (s *BillService) CreateAssignments(assignments []models.BillAssignment) error {
	if len(assignments) == 0 {
		return errors.New("no assignments provided")
	}

	// Validate all assignments
	for _, assignment := range assignments {
		if assignment.BillItemId == nil {
			return errors.New("bill_item_id is required")
		}
		if assignment.EntityType == nil || *assignment.EntityType == "" {
			return errors.New("entity_type is required")
		}
		if assignment.EntityId == nil {
			return errors.New("entity_id is required")
		}
	}

	return s.billAssignmentRepo.BulkCreate(assignments)
}

func (s *BillService) DeleteAssignment(id uint) error {
	_, err := s.billAssignmentRepo.FindByID(id)
	if err != nil {
		return errors.New("assignment not found")
	}

	return s.billAssignmentRepo.Delete(id)
}
