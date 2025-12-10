package repositories

import (
	"gnaps-api/models"
	"time"

	"gorm.io/gorm"
)

type SchoolBillRepository struct {
	db *gorm.DB
}

func NewSchoolBillRepository(db *gorm.DB) *SchoolBillRepository {
	return &SchoolBillRepository{db: db}
}

// SchoolBillWithName extends SchoolBill with bill name
type SchoolBillWithName struct {
	models.SchoolBill
	BillName string `json:"bill_name" gorm:"column:bill_name"`
}

func (r *SchoolBillRepository) FindByID(id uint) (*models.SchoolBill, error) {
	var schoolBill models.SchoolBill
	err := r.db.Where("id = ?", id).First(&schoolBill).Error
	if err != nil {
		return nil, err
	}
	return &schoolBill, nil
}

func (r *SchoolBillRepository) FindBySchoolID(schoolId int64) ([]SchoolBillWithName, error) {
	var schoolBills []SchoolBillWithName

	err := r.db.Table("school_bills").
		Select("school_bills.*, bills.name as bill_name").
		Joins("LEFT JOIN bills ON bills.id = school_bills.bill_id").
		Where("school_bills.school_id = ?", schoolId).
		Order("school_bills.created_at DESC").
		Find(&schoolBills).Error

	return schoolBills, err
}

func (r *SchoolBillRepository) GetBalance(schoolBillId uint) (float64, string, error) {
	var result struct {
		Balance  float64 `gorm:"column:balance"`
		BillName string  `gorm:"column:bill_name"`
	}

	err := r.db.Table("school_bills").
		Select("school_bills.balance, bills.name as bill_name").
		Joins("LEFT JOIN bills ON bills.id = school_bills.bill_id").
		Where("school_bills.id = ?", schoolBillId).
		First(&result).Error

	if err != nil {
		return 0, "", err
	}

	balance := result.Balance
	if balance < 0 {
		balance = 0
	}

	return balance, result.BillName, nil
}

func (r *SchoolBillRepository) UpdatePayment(schoolBillId uint, amountPaid float64) error {
	// Get current school bill
	var schoolBill models.SchoolBill
	if err := r.db.First(&schoolBill, schoolBillId).Error; err != nil {
		return err
	}

	// Calculate new values
	currentAmountPaid := float64(0)
	if schoolBill.AmountPaid != nil {
		currentAmountPaid = *schoolBill.AmountPaid
	}

	newAmountPaid := currentAmountPaid + amountPaid

	// Calculate new balance
	totalAmount := float64(0)
	if schoolBill.Amount != nil {
		totalAmount = *schoolBill.Amount
	}

	discounts := float64(0)
	if schoolBill.Discounts != nil {
		discounts = *schoolBill.Discounts
	}

	newBalance := totalAmount - discounts - newAmountPaid

	// Determine if paid
	isPaid := newBalance <= 0

	// Update school bill
	return r.db.Model(&schoolBill).Updates(map[string]interface{}{
		"amount_paid": newAmountPaid,
		"balance":     newBalance,
		"is_paid":     isPaid,
	}).Error
}

func (r *SchoolBillRepository) GetSchoolBillingParticulars(schoolBillId uint) ([]models.SchoolBillingParticular, error) {
	var particulars []models.SchoolBillingParticular

	err := r.db.Where("school_billing_id = ?", schoolBillId).
		Order("priority ASC").
		Find(&particulars).Error

	return particulars, err
}

// CreateFinanceTransaction creates a finance transaction record
func (r *SchoolBillRepository) CreateFinanceTransaction(transaction *models.FinanceTransaction) error {
	return r.db.Create(transaction).Error
}

// GenerateReceiptNo generates a unique receipt number
func (r *SchoolBillRepository) GenerateReceiptNo() string {
	// Format: RCP-YYYYMMDD-XXXXX (where XXXXX is a sequential number)
	today := time.Now().Format("20060102")

	var count int64
	r.db.Model(&models.FinanceTransaction{}).
		Where("receipt_no LIKE ?", "RCP-"+today+"-%").
		Count(&count)

	return "RCP-" + today + "-" + padLeft(count+1, 5)
}

// PaymentHistoryItem represents a payment transaction with additional details
type PaymentHistoryItem struct {
	ID              uint    `json:"id"`
	Title           string  `json:"title"`
	Description     string  `json:"description"`
	Amount          float64 `json:"amount"`
	TransactionDate string  `json:"transaction_date"`
	PaymentMode     string  `json:"payment_mode"`
	ModeInfo        string  `json:"mode_info"`
	ReceiptNo       string  `json:"receipt_no"`
	ReferenceNo     string  `json:"reference_no"`
	FinanceType     string  `json:"finance_type"`
	CreatedAt       string  `json:"created_at"`
}

// GetPaymentHistoryBySchoolID retrieves all finance transactions for a school
func (r *SchoolBillRepository) GetPaymentHistoryBySchoolID(schoolId int64, page, limit int) ([]PaymentHistoryItem, int64, error) {
	var payments []PaymentHistoryItem
	var total int64

	// Count total records
	r.db.Table("finance_transactions").
		Where("school_id = ?", schoolId).
		Count(&total)

	// Get paginated results
	offset := (page - 1) * limit
	err := r.db.Table("finance_transactions").
		Select(`
			id,
			COALESCE(title, '') as title,
			COALESCE(description, '') as description,
			COALESCE(amount, 0) as amount,
			DATE_FORMAT(transaction_date, '%Y-%m-%d') as transaction_date,
			COALESCE(payment_mode, '') as payment_mode,
			COALESCE(mode_info, '') as mode_info,
			COALESCE(receipt_no, '') as receipt_no,
			COALESCE(reference_no, '') as reference_no,
			COALESCE(finance_type, '') as finance_type,
			DATE_FORMAT(created_at, '%Y-%m-%d %H:%i:%s') as created_at
		`).
		Where("school_id = ?", schoolId).
		Order("transaction_date DESC, created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&payments).Error

	return payments, total, err
}

// Helper function to pad number with leading zeros
func padLeft(n int64, width int) string {
	result := ""
	for i := 0; i < width; i++ {
		result = "0" + result
	}
	numStr := ""
	for n > 0 {
		numStr = string('0'+n%10) + numStr
		n /= 10
	}
	if numStr == "" {
		numStr = "0"
	}
	if len(numStr) >= width {
		return numStr
	}
	return result[:width-len(numStr)] + numStr
}
