package models

import (
	"gorm.io/gorm"
	"time"
)

// Position model generated from database table 'positions'
type Position struct {
	ID        uint           `json:"id" gorm:"primarykey"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	Name      *string `json:"name" gorm:"column:name"`
	IsDeleted *bool   `json:"is_deleted" gorm:"column:is_deleted"`
}

func (Position) TableName() string {
	return "positions"
}
