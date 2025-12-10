package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"time"
)

// SchoolBill model generated from database table 'school_bills'
type SchoolBill struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	SchoolId       *int64          `json:"school_id" gorm:"column:school_id"`
	IsPaid         *bool           `json:"is_paid" gorm:"column:is_paid"`
	Amount         *float64        `json:"amount" gorm:"column:amount"`
	Discounts      *float64        `json:"discounts" gorm:"column:discounts"`
	AmountPaid     *float64        `json:"amount_paid" gorm:"column:amount_paid"`
	CreditAmount   *float64        `json:"credit_amount" gorm:"column:credit_amount"`
	Balance        *float64        `json:"balance" gorm:"column:balance"`
	BillId         *int64          `json:"bill_id" gorm:"column:bill_id"`
	ZoneId         *int64          `json:"zone_id" gorm:"column:zone_id"`
	FeeDetails     *datatypes.JSON `json:"fee_details" gorm:"column:fee_details"`
	SchoolGroupIds *datatypes.JSON `json:"school_group_ids" gorm:"column:school_group_ids"`
	OwnerType      *string         `json:"owner_type" gorm:"column:owner_type"`
	OwnerId        *int64          `json:"owner_id" gorm:"column:owner_id"`
}

func (SchoolBill) TableName() string {
	return "school_bills"
}

// SetOwner implements the OwnerFieldSetter interface
func (s *SchoolBill) SetOwner(ownerType string, ownerID int64) {
	s.OwnerType = &ownerType
	s.OwnerId = &ownerID
}
