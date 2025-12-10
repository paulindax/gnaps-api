package services

import (
	"errors"
	"gnaps-api/models"
	"gnaps-api/repositories"
	"gnaps-api/utils"
)

type BillService struct {
	billRepo     *repositories.BillRepository
	billItemRepo *repositories.BillItemRepository
}

func NewBillService(
	billRepo *repositories.BillRepository,
	billItemRepo *repositories.BillItemRepository,
) *BillService {
	return &BillService{
		billRepo:     billRepo,
		billItemRepo: billItemRepo,
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

func (s *BillService) GetBillItemsByBillIDWithPagination(billId uint, page, limit int, search string) ([]models.BillItem, int64, error) {
	return s.billItemRepo.FindByBillIDWithPagination(billId, page, limit, search)
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

	return s.billItemRepo.Delete(id)
}

// ============================================
// Owner-based methods for data filtering
// ============================================

// GetBillByIDWithOwner retrieves a bill with owner filtering
func (s *BillService) GetBillByIDWithOwner(id uint, ownerCtx *utils.OwnerContext) (*models.Bill, error) {
	return s.billRepo.FindByIDWithOwner(id, ownerCtx)
}

// ListBillsWithOwner retrieves bills with owner filtering
func (s *BillService) ListBillsWithOwner(filters map[string]interface{}, page, limit int, ownerCtx *utils.OwnerContext) ([]models.Bill, int64, error) {
	return s.billRepo.ListWithOwner(filters, page, limit, ownerCtx)
}

// CreateBillWithOwner creates a bill with owner context
func (s *BillService) CreateBillWithOwner(bill *models.Bill, ownerCtx *utils.OwnerContext) error {
	if bill.Name == nil || *bill.Name == "" {
		return errors.New("name is required")
	}

	bill.IsDeleted = false

	return s.billRepo.CreateWithOwner(bill, ownerCtx)
}

// UpdateBillWithOwner updates a bill with owner verification
func (s *BillService) UpdateBillWithOwner(id uint, updates map[string]interface{}, ownerCtx *utils.OwnerContext) error {
	return s.billRepo.UpdateWithOwner(id, updates, ownerCtx)
}

// DeleteBillWithOwner soft deletes a bill with owner verification
func (s *BillService) DeleteBillWithOwner(id uint, ownerCtx *utils.OwnerContext) error {
	return s.billRepo.DeleteWithOwner(id, ownerCtx)
}
