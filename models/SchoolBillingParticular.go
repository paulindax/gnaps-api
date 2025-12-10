package models

import (
	"gorm.io/datatypes"
	"time"
)

// SchoolBillingParticular model generated from database table 'school_billing_particulars'
type SchoolBillingParticular struct {
	ID        uint      `json:"id" gorm:"primarykey"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	SchoolId  *int64    `json:"school_id" gorm:"column:school_id"`

	ParticularName   *string         `json:"particular_name" gorm:"column:particular_name"`
	Amount           *float64        `json:"amount" gorm:"column:amount"`
	DiscountAmount   *float64        `json:"discount_amount" gorm:"column:discount_amount"`
	AmountPaid       *float64        `json:"amount_paid" gorm:"column:amount_paid"`
	CreditAmount     *float64        `json:"credit_amount" gorm:"column:credit_amount"`
	Priority         *int            `json:"priority" gorm:"column:priority"`
	RecieverType     *string         `json:"reciever_type" gorm:"column:reciever_type"`
	BillParticularId *int64          `json:"bill_particular_id" gorm:"column:bill_particular_id"`
	ZoneId           *int64          `json:"zone_id" gorm:"column:zone_id"`
	BillId           *int64          `json:"bill_id" gorm:"column:bill_id"`
	SchoolBillingId  *int64          `json:"school_billing_id" gorm:"column:school_billing_id"`
	BillingItemId    *int            `json:"billing_item_id" gorm:"column:billing_item_id"`
	FinanceAccountId *int            `json:"finance_account_id" gorm:"column:finance_account_id"`
	IsApproved       *bool           `json:"is_approved" gorm:"column:is_approved"`
	IsDeleted        *bool           `json:"is_deleted" gorm:"column:is_deleted"`
	FeeDetails       *datatypes.JSON `json:"fee_details" gorm:"column:fee_details"`
}

func (SchoolBillingParticular) TableName() string {
	return "school_billing_particulars"
}
