package services

import (
	"errors"
	"gnaps-api/models"
	"gnaps-api/repositories"
	"time"
)

type SchoolBillService struct {
	schoolBillRepo *repositories.SchoolBillRepository
}

func NewSchoolBillService(schoolBillRepo *repositories.SchoolBillRepository) *SchoolBillService {
	return &SchoolBillService{
		schoolBillRepo: schoolBillRepo,
	}
}

// GetSchoolBillByID retrieves a school bill by ID
func (s *SchoolBillService) GetSchoolBillByID(id uint) (*models.SchoolBill, error) {
	schoolBill, err := s.schoolBillRepo.FindByID(id)
	if err != nil {
		return nil, errors.New("school bill not found")
	}
	return schoolBill, nil
}

// GetSchoolBills retrieves all bills for a school
func (s *SchoolBillService) GetSchoolBills(schoolId int64) ([]repositories.SchoolBillWithName, error) {
	return s.schoolBillRepo.FindBySchoolID(schoolId)
}

// GetBalance retrieves the balance for a specific school bill
func (s *SchoolBillService) GetBalance(schoolBillId uint) (float64, string, error) {
	return s.schoolBillRepo.GetBalance(schoolBillId)
}

// PaymentRequest represents a payment request
type PaymentRequest struct {
	SchoolID     int64   `json:"school_id"`
	SchoolName   string  `json:"school_name"`
	SchoolBillID uint    `json:"school_bill_id"`
	Amount       float64 `json:"amount"`
	PaymentMode  string  `json:"payment_mode"` // Cash, MoMo
	PaymentDate  string  `json:"payment_date"`
	PaymentNote  string  `json:"payment_note"`
	MomoNumber   string  `json:"momo_number"`
	MomoNetwork  string  `json:"momo_network"` // MTN, TELECEL, AIRTELTIGO
	UserID       int64   `json:"user_id"`
}

// RecordPayment records a payment against a school bill
func (s *SchoolBillService) RecordPayment(req PaymentRequest) (*models.FinanceTransaction, error) {
	// Validate amount
	if req.Amount <= 0 {
		return nil, errors.New("payment amount must be greater than 0")
	}

	// Get current balance
	balance, billName, err := s.schoolBillRepo.GetBalance(req.SchoolBillID)
	if err != nil {
		return nil, errors.New("school bill not found")
	}

	// Check if amount exceeds balance
	if req.Amount > balance {
		return nil, errors.New("payment amount exceeds outstanding balance")
	}

	// Parse payment date
	paymentDate, err := time.Parse("2006-01-02", req.PaymentDate)
	if err != nil {
		paymentDate = time.Now()
	}

	// Generate receipt number
	receiptNo := s.schoolBillRepo.GenerateReceiptNo()

	// Create finance transaction
	title := "Payment for " + billName
	description := "School payment via " + req.PaymentMode
	schoolBillID := int64(req.SchoolBillID)
	financeType := "SchoolBill"

	transaction := &models.FinanceTransaction{
		Title:           &title,
		Description:     &description,
		Amount:          &req.Amount,
		TransactionDate: paymentDate,
		FinanceId:       &schoolBillID,
		FinanceType:     &financeType,
		SchoolId:        &req.SchoolID,
		ReceiptNo:       &receiptNo,
		PaymentMode:     &req.PaymentMode,
		PaymentNote:     &req.PaymentNote,
		UserId:          &req.UserID,
	}

	// Add MoMo info if applicable
	if req.PaymentMode == "MoMo" && req.MomoNumber != "" {
		transaction.ModeInfo = &req.MomoNumber
	}

	// Create the transaction
	if err := s.schoolBillRepo.CreateFinanceTransaction(transaction); err != nil {
		return nil, errors.New("failed to record payment")
	}

	// Update school bill with new payment
	if err := s.schoolBillRepo.UpdatePayment(req.SchoolBillID, req.Amount); err != nil {
		return nil, errors.New("failed to update school bill balance")
	}

	return transaction, nil
}

// GetSchoolBillingParticulars retrieves the particulars for a school bill
func (s *SchoolBillService) GetSchoolBillingParticulars(schoolBillId uint) ([]models.SchoolBillingParticular, error) {
	return s.schoolBillRepo.GetSchoolBillingParticulars(schoolBillId)
}

// GetPaymentHistory retrieves payment history for a school
func (s *SchoolBillService) GetPaymentHistory(schoolId int64, page, limit int) ([]repositories.PaymentHistoryItem, int64, error) {
	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}
	return s.schoolBillRepo.GetPaymentHistoryBySchoolID(schoolId, page, limit)
}
