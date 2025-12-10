package repositories

import (
	"gnaps-api/models"

	"gorm.io/gorm"
)

type MomoPaymentRepository struct {
	db *gorm.DB
}

func NewMomoPaymentRepository(db *gorm.DB) *MomoPaymentRepository {
	return &MomoPaymentRepository{db: db}
}

func (r *MomoPaymentRepository) Create(payment *models.MomoPayment) error {
	return r.db.Create(payment).Error
}

func (r *MomoPaymentRepository) FindByID(id uint) (*models.MomoPayment, error) {
	var payments []models.MomoPayment
	err := r.db.Where("id = ?", id).Limit(1).Find(&payments).Error
	if err != nil {
		return nil, err
	}
	if len(payments) == 0 {
		return nil, nil // Return nil, nil when not found (not an error condition)
	}
	return &payments[0], nil
}

func (r *MomoPaymentRepository) FindByPayeeAndType(payeeID int64, payeeType string) (*models.MomoPayment, error) {
	var payments []models.MomoPayment
	err := r.db.Where("payee_id = ? AND payee_type = ?", payeeID, payeeType).
		Order("created_at DESC").
		Limit(1).
		Find(&payments).Error
	if err != nil {
		return nil, err
	}
	if len(payments) == 0 {
		return nil, nil // Return nil, nil when not found
	}
	return &payments[0], nil
}

func (r *MomoPaymentRepository) FindPendingByPayeeAndType(schoolID int64, payeeID int64, payeeType string) (*models.MomoPayment, error) {
	var payments []models.MomoPayment
	// Check for both "created" and "pending" status (created = not yet processed, pending = sent to Hubtel)
	err := r.db.Where("school_id = ? AND payee_id = ? AND payee_type = ? AND status IN ?", schoolID, payeeID, payeeType, []string{"created", "pending"}).
		Order("created_at DESC").
		Limit(1).
		Find(&payments).Error
	if err != nil {
		return nil, err
	}
	if len(payments) == 0 {
		return nil, nil // No pending payment found (this is expected, not an error)
	}
	return &payments[0], nil
}

func (r *MomoPaymentRepository) FindByMomoTransactionID(transactionID string) (*models.MomoPayment, error) {
	var payment models.MomoPayment
	err := r.db.Where("momo_transaction_id = ?", transactionID).First(&payment).Error
	if err != nil {
		return nil, err
	}
	return &payment, nil
}

func (r *MomoPaymentRepository) Update(id uint, updates map[string]interface{}) error {
	return r.db.Model(&models.MomoPayment{}).Where("id = ?", id).Updates(updates).Error
}

func (r *MomoPaymentRepository) UpdateStatus(id uint, status, bankStatus string, transStatus *string) error {
	updates := map[string]interface{}{
		"status":      status,
		"bank_status": bankStatus,
	}
	if transStatus != nil {
		updates["trans_status"] = *transStatus
	}
	return r.db.Model(&models.MomoPayment{}).Where("id = ?", id).Updates(updates).Error
}

func (r *MomoPaymentRepository) IncrementRetries(id uint) error {
	return r.db.Model(&models.MomoPayment{}).Where("id = ?", id).
		UpdateColumn("retries", gorm.Expr("COALESCE(retries, 0) + 1")).Error
}

func (r *MomoPaymentRepository) GetPendingPayments() ([]models.MomoPayment, error) {
	var payments []models.MomoPayment
	err := r.db.Where("status = ? AND (retries IS NULL OR retries < ?)", "pending", 75).
		Order("created_at ASC").
		Limit(100).
		Find(&payments).Error
	return payments, err
}

// GetCreatedPayments returns payments with "created" status that need to be processed with Hubtel
func (r *MomoPaymentRepository) GetCreatedPayments() ([]models.MomoPayment, error) {
	var payments []models.MomoPayment
	err := r.db.Where("status = ?", "created").
		Order("created_at ASC").
		Limit(50).
		Find(&payments).Error
	return payments, err
}

func (r *MomoPaymentRepository) ListByPayee(payeeID int64, payeeType string, page, limit int) ([]models.MomoPayment, int64, error) {
	var payments []models.MomoPayment
	var total int64

	query := r.db.Model(&models.MomoPayment{}).
		Where("payee_id = ? AND payee_type = ?", payeeID, payeeType)

	query.Count(&total)

	offset := (page - 1) * limit
	err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&payments).Error

	return payments, total, err
}
