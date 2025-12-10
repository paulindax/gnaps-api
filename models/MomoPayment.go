package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"time"
)

// MomoPayment model generated from database table 'momo_payments'
type MomoPayment struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Amount                *float64        `json:"amount" gorm:"column:amount"`
	Status                *string         `json:"status" gorm:"column:status"`
	PaymentGatewayId      *int64          `json:"payment_gateway_id" gorm:"column:payment_gateway_id"`
	SchoolId              *int64          `json:"school_id" gorm:"column:school_id"`
	PayeeId               *int64          `json:"payee_id" gorm:"column:payee_id"`
	PayeeType             *string         `json:"payee_type" gorm:"column:payee_type"`
	UserId                *int64          `json:"user_id" gorm:"column:user_id"`
	FeeName               *string         `json:"fee_name" gorm:"column:fee_name"`
	TransactionFee        *float64        `json:"transaction_fee" gorm:"column:transaction_fee"`
	MomoNetwork           *string         `json:"momo_network" gorm:"column:momo_network"`
	MomoNumber            *string         `json:"momo_number" gorm:"column:momo_number"`
	TransactionDate       time.Time       `json:"transaction_date" gorm:"column:transaction_date"`
	BankStatus            *string         `json:"bank_status" gorm:"column:bank_status"`
	MomoTransactionId     *string         `json:"momo_transaction_id" gorm:"column:momo_transaction_id"`
	FinanceTransactionIds *datatypes.JSON `json:"finance_transaction_ids" gorm:"column:finance_transaction_ids"`
	VCode                 *string         `json:"v_code" gorm:"column:v_code"`
	GatewayResponse       *datatypes.JSON `json:"gateway_response" gorm:"column:gateway_response"`
	PaymentDetails        *datatypes.JSON `json:"payment_details" gorm:"column:payment_details"`
	MapiMomoPaymentId     *int64          `json:"mapi_momo_payment_id" gorm:"column:mapi_momo_payment_id"`
	TransStatus           *string         `json:"trans_status" gorm:"column:trans_status"`
	IsDeleted             *bool           `json:"is_deleted" gorm:"column:is_deleted"`
	Retries               *int            `json:"retries" gorm:"column:retries"`
	OwnerType             *string         `json:"owner_type" gorm:"column:owner_type"`
	OwnerId               *int64          `json:"owner_id" gorm:"column:owner_id"`
}

func (MomoPayment) TableName() string {
	return "momo_payments"
}

// SetOwner implements the OwnerFieldSetter interface
func (m *MomoPayment) SetOwner(ownerType string, ownerID int64) {
	m.OwnerType = &ownerType
	m.OwnerId = &ownerID
}
