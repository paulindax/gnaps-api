package models

import (
	"gorm.io/gorm"
	"time"
)

// SchoolGroup model generated from database table 'school_groups'
type SchoolGroup struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Name *string `json:"name" gorm:"column:name"`
	IsDeleted bool `json:"is_deleted" gorm:"column:is_deleted"`
	ZoneId *int64 `json:"zone_id" gorm:"column:zone_id"`
	Description *string `json:"description" gorm:"column:description"`
}

func (SchoolGroup) TableName() string {
	return "school_groups"
}
