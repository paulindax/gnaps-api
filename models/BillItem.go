package models

import (
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"time"
)

// BillItem model generated from database table 'bill_items'
type BillItem struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Name             *string         `json:"name" gorm:"column:name"`
	Amount           *float64        `json:"amount" gorm:"column:amount"`
	Priority         *int64          `json:"priority" gorm:"column:priority"`
	BillId           *int64          `json:"bill_id" gorm:"column:bill_id"`
	IsDeleted        *bool           `json:"is_deleted" gorm:"column:is_deleted"`
	IsApproved       *bool           `json:"is_approved" gorm:"column:is_approved"`
	BillParticularId *int64          `json:"bill_particular_id" gorm:"column:bill_particular_id"`
	FinanceAccountId *int64          `json:"finance_account_id" gorm:"column:finance_account_id"`
	RegionIds        *datatypes.JSON `json:"region_ids" gorm:"column:region_ids"`
	ZoneIds          *datatypes.JSON `json:"zone_ids" gorm:"column:zone_ids"`
	SchoolGroupIds   *datatypes.JSON `json:"school_group_ids" gorm:"column:school_group_ids"`
	SchoolIds        *datatypes.JSON `json:"school_ids" gorm:"column:school_ids"`
	OwnerType        *string         `json:"owner_type" gorm:"column:owner_type"`
	OwnerId          *int64          `json:"owner_id" gorm:"column:owner_id"`

	// Transient fields (not in database)
	BillParticular *BillParticular `json:"bill_particular,omitempty" gorm:"foreignKey:BillParticularId"`
}

func (BillItem) TableName() string {
	return "bill_items"
}

// SetOwner implements the OwnerFieldSetter interface
func (b *BillItem) SetOwner(ownerType string, ownerID int64) {
	b.OwnerType = &ownerType
	b.OwnerId = &ownerID
}
