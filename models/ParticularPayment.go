package models

import (
	"time"
	"gorm.io/datatypes"
)

// ParticularPayment model generated from database table 'particular_payments'
type ParticularPayment struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	Amount *float64 `json:"amount" gorm:"column:amount"`

	StudentFeeId *int `json:"student_fee_id" gorm:"column:student_fee_id"`
	FinanceFeeId *int `json:"finance_fee_id" gorm:"column:finance_fee_id"`
	FinanceFeeParticularId *int `json:"finance_fee_particular_id" gorm:"column:finance_fee_particular_id"`
	FinanceTransactionId *int `json:"finance_transaction_id" gorm:"column:finance_transaction_id"`
	IsPaid *bool `json:"is_paid" gorm:"column:is_paid"`
	FinePaid *bool `json:"fine_paid" gorm:"column:fine_paid"`
	SchoolId *int `json:"school_id" gorm:"column:school_id"`
	FeeDetails *datatypes.JSON `json:"fee_details" gorm:"column:fee_details"`
}

func (ParticularPayment) TableName() string {
	return "particular_payments"
}
