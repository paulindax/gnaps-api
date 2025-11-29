package models

import (
	"gorm.io/gorm"
	"time"
)

// Region model generated from database table 'regions'
type Region struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Name *string `json:"name" gorm:"column:name"`
	Code *string `json:"code" gorm:"column:code"`
	IsDeleted *bool `json:"is_deleted" gorm:"column:is_deleted"`
}

func (Region) TableName() string {
	return "regions"
}
