package models

import (
	"gorm.io/gorm"
	"time"
)

// Zone model generated from database table 'zones'
type Zone struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Name *string `json:"name" gorm:"column:name"`
	Code *string `json:"code" gorm:"column:code"`
	RegionId *int64 `json:"region_id" gorm:"column:region_id"`
	IsDeleted *bool `json:"is_deleted" gorm:"column:is_deleted"`
}

func (Zone) TableName() string {
	return "zones"
}
