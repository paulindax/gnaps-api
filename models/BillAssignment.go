package models

import (
	"gorm.io/gorm"
	"time"
)

// BillAssignment model for assigning bill items to entities
type BillAssignment struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	BillItemId *uint   `json:"bill_item_id" gorm:"column:bill_item_id"`
	EntityType *string `json:"entity_type" gorm:"column:entity_type"` // region, zone, group, school
	EntityId   *uint   `json:"entity_id" gorm:"column:entity_id"`

	// Relations
	BillItem *BillItem `json:"bill_item,omitempty" gorm:"foreignKey:BillItemId"`
}

func (BillAssignment) TableName() string {
	return "bill_assignments"
}
