package models

import (
	"gorm.io/gorm"
	"time"
	"gorm.io/datatypes"
)

// FinanceTransaction model generated from database table 'finance_transactions'
type FinanceTransaction struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Title *string `json:"title" gorm:"column:title"`
	Description *string `json:"description" gorm:"column:description"`
	Amount *float64 `json:"amount" gorm:"column:amount"`
	FinanceAccountId *int64 `json:"finance_account_id" gorm:"column:finance_account_id"`
	TransactionDate time.Time `json:"transaction_date" gorm:"column:transaction_date"`
	FinanceId *int64 `json:"finance_id" gorm:"column:finance_id"`
	FinanceType *string `json:"finance_type" gorm:"column:finance_type"`
	SchoolId *int64 `json:"school_id" gorm:"column:school_id"`
	ReceiptNo *string `json:"receipt_no" gorm:"column:receipt_no"`
	VoucherNo *string `json:"voucher_no" gorm:"column:voucher_no"`
	PaymentMode *string `json:"payment_mode" gorm:"column:payment_mode"`
	ModeInfo *string `json:"mode_info" gorm:"column:mode_info"`
	PaymentNote *string `json:"payment_note" gorm:"column:payment_note"`
	UserId *int64 `json:"user_id" gorm:"column:user_id"`
	ReferenceNo *string `json:"reference_no" gorm:"column:reference_no"`
	PaymentDetails *datatypes.JSON `json:"payment_details" gorm:"column:payment_details"`
}

func (FinanceTransaction) TableName() string {
	return "finance_transactions"
}
