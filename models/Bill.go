package models

import (
	"gorm.io/gorm"
	"time"
	"gorm.io/datatypes"
)

// Bill model generated from database table 'bills'
type Bill struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Name *string `json:"name" gorm:"column:name"`
	Description *string `json:"description" gorm:"column:description"`
	IsDeleted bool `json:"is_deleted" gorm:"column:is_deleted"`
	IsApproved *bool `json:"is_approved" gorm:"column:is_approved"`
	ZonesIds *datatypes.JSON `json:"zones_ids" gorm:"column:zones_ids"`
	ForwardArrears *bool `json:"forward_arrears" gorm:"column:forward_arrears"`
	IsGenerating *bool `json:"is_generating" gorm:"column:is_generating"`
	Settings *datatypes.JSON `json:"settings" gorm:"column:settings"`
}

func (Bill) TableName() string {
	return "bills"
}
